package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"health-hub/internal/gpx"
	"health-hub/internal/models"
	"health-hub/internal/storage"
	"health-hub/internal/templates"
)

type Handlers struct {
	storage   storage.Storage
	templates *templates.Templates
}

func NewHandlers(s storage.Storage) *Handlers {
	tmpl := templates.NewTemplates()
	if err := tmpl.LoadTemplates(); err != nil {
		fmt.Printf("ERROR: Failed to load templates: %v\n", err)
		panic(fmt.Sprintf("Failed to load templates: %v", err))
	}
	fmt.Println("INFO: Templates loaded successfully")
	return &Handlers{
		storage:   s,
		templates: tmpl,
	}
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := struct {
		Title string
	}{
		Title: "Home",
	}

	tmpl := h.templates.GetTemplate("home")
	if tmpl == nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		fmt.Printf("Template execution error: %v\n", err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}

func (h *Handlers) Upload(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) GetActivities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	activities, err := h.storage.GetActivities()
	if err != nil {
		fmt.Printf("ERROR: Failed to get activities: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activities)
}

func (h *Handlers) GetHealthMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics, err := h.storage.GetHealthMetrics()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (h *Handlers) UploadGPX(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("gpx")
	if err != nil {
		fmt.Printf("ERROR: Failed to read GPX file: %v\n", err)
		http.Error(w, "Error reading file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file data", http.StatusInternalServerError)
		return
	}

	// Save the raw GPX file
	filename := fmt.Sprintf("gpx_%d_%s", time.Now().UnixNano(), header.Filename)
	if err := h.storage.SaveFile(filename, data); err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	// Parse GPX and create activity record
	track, activity, err := gpx.ParseGPX(string(data))
	if err != nil {
		fmt.Printf("ERROR: Failed to parse GPX file %s: %v\n", header.Filename, err)
		http.Error(w, "Error parsing GPX file", http.StatusBadRequest)
		return
	}

	// Set additional activity details
	if activity.Name == "" {
		activity.Name = strings.TrimSuffix(header.Filename, ".gpx")
	}
	activity.GPXFile = filename

	// Save activity first to get the generated ID
	if err := h.storage.SaveActivity(activity); err != nil {
		http.Error(w, "Error saving activity", http.StatusInternalServerError)
		return
	}

	// Use the same ID for the GPX track so we can link them
	track.ID = activity.ID
	if err := h.storage.SaveGPXTrack(track); err != nil {
		http.Error(w, "Error saving GPX track", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<div class="p-3 bg-green-100 border border-green-400 text-green-700 rounded">✓ GPX uploaded successfully!</div>`))
}

func (h *Handlers) UploadHealthData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("health")
	if err != nil {
		http.Error(w, "Error reading file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file data", http.StatusInternalServerError)
		return
	}

	// Try to parse as array of health metrics
	var metrics []models.HealthMetric
	if err := json.Unmarshal(data, &metrics); err != nil {
		// Try to parse as single metric
		var metric models.HealthMetric
		if err := json.Unmarshal(data, &metric); err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
		metrics = []models.HealthMetric{metric}
	}

	// Save all metrics
	for _, metric := range metrics {
		if err := h.storage.SaveHealthMetric(&metric); err != nil {
			http.Error(w, "Error saving health metric", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf(`<div class="p-3 bg-green-100 border border-green-400 text-green-700 rounded">✓ Uploaded %d health metrics!</div>`, len(metrics))))
}

func (h *Handlers) Activities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	activities, err := h.storage.GetActivities()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get unit preference from cookie or default to metric
	useImperial := false
	if cookie, err := r.Cookie("units"); err == nil && cookie.Value == "imperial" {
		useImperial = true
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Activities - Health Hub</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://cdn.jsdelivr.net/npm/sortable-tablesort@2.0.0/sortable.min.js"></script>
</head>
<body class="bg-gray-100 min-h-screen">
    <div class="container mx-auto px-4 py-8 max-w-7xl">
        <!-- Header -->
        <div class="flex justify-between items-center mb-8">
            <div>
                <h1 class="text-4xl font-bold text-gray-900 mb-2">Activities</h1>
                <p class="text-gray-600">Your fitness journey</p>
            </div>
            <div class="flex items-center space-x-4">
                <!-- Unit Toggle -->
                <div class="flex items-center space-x-2">
                    <span class="text-sm text-gray-600">Units:</span>
                    <button id="unit-toggle" class="{{if .UseImperial}}bg-orange-500{{else}}bg-blue-500{{end}} text-white px-3 py-1 rounded text-sm font-medium hover:opacity-80 transition-opacity">
                        {{if .UseImperial}}Imperial{{else}}Metric{{end}}
                    </button>
                </div>
                <a href="/" class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded transition duration-200">
                    Back to Home
                </a>
            </div>
        </div>

        {{if .Activities}}
        <!-- Search and Filters -->
        <div class="bg-white rounded-lg shadow-md p-6 mb-6">
            <div class="flex flex-col md:flex-row md:items-center md:justify-between space-y-4 md:space-y-0">
                <div class="flex-1 max-w-md">
                    <input type="text" id="search-input" placeholder="Search activities..." 
                           class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent">
                </div>
                <div class="flex items-center space-x-4">
                    <select id="type-filter" class="px-3 py-2 border border-gray-300 rounded-lg">
                        <option value="">All Types</option>
                        <option value="running">Running</option>
                        <option value="cycling">Cycling</option>
                        <option value="walking">Walking</option>
                        <option value="hiking">Hiking</option>
                    </select>
                    <span class="text-sm text-gray-600">{{len .Activities}} activities</span>
                </div>
            </div>
        </div>

        <!-- Activities Table -->
        <div class="bg-white rounded-lg shadow-md overflow-hidden">
            <div class="overflow-x-auto">
                <table id="activities-table" class="min-w-full divide-y divide-gray-200 sortable">
                    <thead class="bg-gray-50">
                        <tr>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100">
                                Activity
                            </th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100">
                                Type
                            </th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100">
                                Date
                            </th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100">
                                Distance
                            </th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100">
                                Duration
                            </th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100">
                                Avg Speed
                            </th>
                            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                Actions
                            </th>
                        </tr>
                    </thead>
                    <tbody class="bg-white divide-y divide-gray-200">
                        {{range .Activities}}
                        <tr class="hover:bg-gray-50 cursor-pointer activity-row" data-activity-id="{{.ID}}" data-type="{{.Type}}">
                            <td class="px-6 py-4 whitespace-nowrap">
                                <div class="flex items-center">
                                    <div class="flex-shrink-0 h-10 w-10">
                                        <div class="h-10 w-10 rounded-full bg-{{getTypeColor .Type}}-100 flex items-center justify-center">
                                            <span class="text-{{getTypeColor .Type}}-600 font-medium text-sm">{{getTypeIcon .Type}}</span>
                                        </div>
                                    </div>
                                    <div class="ml-4">
                                        <div class="text-sm font-medium text-gray-900">{{.Name}}</div>
                                        {{if .GPXFile}}
                                        <div class="text-sm text-gray-500">GPS Track Available</div>
                                        {{end}}
                                    </div>
                                </div>
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap">
                                <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-{{getTypeColor .Type}}-100 text-{{getTypeColor .Type}}-800">
                                    {{.Type}}
                                </span>
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900" data-sort="{{.StartTime.Unix}}">
                                {{.StartTime.Format "Jan 2, 2006"}}
                                <div class="text-xs text-gray-500">{{.StartTime.Format "3:04 PM"}}</div>
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900" data-sort="{{.Distance}}">
                                <span class="font-medium">{{if $.UseImperial}}{{printf "%.1f" (metersToMiles .Distance)}} mi{{else}}{{printf "%.1f" (metersToKm .Distance)}} km{{end}}</span>
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900" data-sort="{{.Duration}}">
                                {{formatDuration .Duration}}
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900" data-sort="{{.AvgSpeed}}">
                                {{if $.UseImperial}}{{printf "%.1f" (kmhToMph .AvgSpeed)}} mph{{else}}{{printf "%.1f" .AvgSpeed}} km/h{{end}}
                            </td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                                <a href="/activity/{{.ID}}" class="text-blue-600 hover:text-blue-900">View Details</a>
                            </td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            </div>
        </div>
        {{else}}
        <div class="bg-white rounded-lg shadow-md">
            <div class="px-6 py-12 text-center">
                <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"></path>
                </svg>
                <h3 class="mt-2 text-sm font-medium text-gray-900">No activities yet</h3>
                <p class="mt-1 text-sm text-gray-500">Get started by uploading a GPX file.</p>
                <div class="mt-6 space-x-4">
                    <a href="/" class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded transition duration-200">
                        Single Upload
                    </a>
                    <a href="/bulk-upload" class="bg-orange-500 hover:bg-orange-700 text-white font-bold py-2 px-4 rounded transition duration-200">
                        Bulk Upload
                    </a>
                </div>
            </div>
        </div>
        {{end}}
    </div>

    <script>
        // Unit toggle functionality
        document.getElementById('unit-toggle').addEventListener('click', function() {
            const currentUnit = this.textContent.trim();
            const newUnit = currentUnit === 'Metric' ? 'imperial' : 'metric';
            
            // Set cookie
            document.cookie = 'units=' + newUnit + '; path=/; max-age=' + (365 * 24 * 60 * 60);
            
            // Reload page to apply new units
            window.location.reload();
        });

        // Search functionality
        const searchInput = document.getElementById('search-input');
        const typeFilter = document.getElementById('type-filter');
        const tableRows = document.querySelectorAll('.activity-row');

        function filterTable() {
            const searchTerm = searchInput.value.toLowerCase();
            const selectedType = typeFilter.value.toLowerCase();

            tableRows.forEach(row => {
                const activityName = row.querySelector('td:first-child .text-sm.font-medium').textContent.toLowerCase();
                const activityType = row.dataset.type.toLowerCase();
                
                const matchesSearch = activityName.includes(searchTerm);
                const matchesType = !selectedType || activityType === selectedType;
                
                row.style.display = matchesSearch && matchesType ? '' : 'none';
            });
        }

        searchInput.addEventListener('input', filterTable);
        typeFilter.addEventListener('change', filterTable);

        // Row click to view activity
        tableRows.forEach(row => {
            row.addEventListener('click', function(e) {
                if (e.target.tagName !== 'A') {
                    const activityId = this.dataset.activityId;
                    window.location.href = '/activity/' + activityId;
                }
            });
        });

        // Initialize sortable table
        if (typeof Sortable !== 'undefined') {
            Sortable.initTable(document.getElementById('activities-table'));
        }
    </script>
</body>
</html>`

	funcMap := template.FuncMap{
		"metersToKm": func(meters float64) float64 {
			return meters / 1000
		},
		"metersToMiles": func(meters float64) float64 {
			return meters * 0.000621371
		},
		"kmhToMph": func(kmh float64) float64 {
			return kmh * 0.621371
		},
		"formatDuration": func(seconds int) string {
			hours := seconds / 3600
			minutes := (seconds % 3600) / 60
			if hours > 0 {
				return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds%60)
			}
			return fmt.Sprintf("%d:%02d", minutes, seconds%60)
		},
		"getTypeColor": func(activityType string) string {
			switch strings.ToLower(activityType) {
			case "running":
				return "red"
			case "cycling":
				return "blue"
			case "walking":
				return "green"
			case "hiking":
				return "yellow"
			default:
				return "gray"
			}
		},
		"getTypeIcon": func(activityType string) string {
			switch strings.ToLower(activityType) {
			case "running":
				return "🏃"
			case "cycling":
				return "🚴"
			case "walking":
				return "🚶"
			case "hiking":
				return "🥾"
			default:
				return "🏃"
			}
		},
	}

	data := struct {
		Activities  []*models.Activity
		UseImperial bool
	}{
		Activities:  activities,
		UseImperial: useImperial,
	}

	t, err := template.New("activities").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(w, data)
}

func (h *Handlers) StatsActivities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	activities, err := h.storage.GetActivities()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html := fmt.Sprintf(`
	<div class="bg-gradient-to-r from-blue-50 to-blue-100 p-6 rounded-lg">
		<h3 class="text-lg font-semibold text-gray-900 mb-2">Activities</h3>
		<p class="text-2xl font-bold text-blue-600 mb-2">%d</p>
		<p class="text-sm text-gray-600 mb-3">total activities</p>
		<a href="/activities" class="inline-block bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded text-sm transition duration-200">
			View All Activities
		</a>
	</div>`, len(activities))

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (h *Handlers) StatsHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics, err := h.storage.GetHealthMetrics()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html := fmt.Sprintf(`
	<div class="bg-gradient-to-r from-green-50 to-green-100 p-6 rounded-lg">
		<h3 class="text-lg font-semibold text-gray-900 mb-2">Health Metrics</h3>
		<p class="text-2xl font-bold text-green-600 mb-2">%d</p>
		<p class="text-sm text-gray-600">total metrics</p>
	</div>`, len(metrics))

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (h *Handlers) Stats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	activities, err := h.storage.GetActivities()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get unit preference from cookie
	useImperial := false
	if cookie, err := r.Cookie("units"); err == nil && cookie.Value == "imperial" {
		useImperial = true
	}

	// Calculate stats for different time periods
	now := time.Now()
	sevenDaysAgo := now.AddDate(0, 0, -7)
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	var totalDistance, totalDuration float64
	var totalActivities int
	var last7Days, last30Days []ActivityStat
	var weeklyStats []WeekStat

	// Initialize weekly stats for last 4 weeks
	for i := 0; i < 4; i++ {
		weekStart := now.AddDate(0, 0, -7*(i+1))
		weekEnd := now.AddDate(0, 0, -7*i)
		weeklyStats = append(weeklyStats, WeekStat{
			Week:       fmt.Sprintf("Week %d", i+1),
			StartDate:  weekStart,
			EndDate:    weekEnd,
			Distance:   0,
			Activities: 0,
			Duration:   0,
		})
	}

	// Initialize daily stats for last 7 days
	for i := 0; i < 7; i++ {
		day := now.AddDate(0, 0, -i)
		last7Days = append(last7Days, ActivityStat{
			Date:       day,
			Distance:   0,
			Activities: 0,
			Duration:   0,
		})
	}

	// Initialize daily stats for last 30 days
	for i := 0; i < 30; i++ {
		day := now.AddDate(0, 0, -i)
		last30Days = append(last30Days, ActivityStat{
			Date:       day,
			Distance:   0,
			Activities: 0,
			Duration:   0,
		})
	}

	// Process activities
	for _, activity := range activities {
		totalDistance += activity.Distance
		totalDuration += float64(activity.Duration)
		totalActivities++

		// Check if activity is within last 7 days
		if activity.StartTime.After(sevenDaysAgo) {
			for i := range last7Days {
				if isSameDay(activity.StartTime, last7Days[i].Date) {
					last7Days[i].Distance += activity.Distance
					last7Days[i].Activities++
					last7Days[i].Duration += float64(activity.Duration)
					break
				}
			}
		}

		// Check if activity is within last 30 days
		if activity.StartTime.After(thirtyDaysAgo) {
			for i := range last30Days {
				if isSameDay(activity.StartTime, last30Days[i].Date) {
					last30Days[i].Distance += activity.Distance
					last30Days[i].Activities++
					last30Days[i].Duration += float64(activity.Duration)
					break
				}
			}

			// Add to weekly stats
			for i := range weeklyStats {
				if activity.StartTime.After(weeklyStats[i].StartDate) && activity.StartTime.Before(weeklyStats[i].EndDate) {
					weeklyStats[i].Distance += activity.Distance
					weeklyStats[i].Activities++
					weeklyStats[i].Duration += float64(activity.Duration)
					break
				}
			}
		}
	}

	// Reverse arrays to show oldest to newest
	for i := 0; i < len(last7Days)/2; i++ {
		last7Days[i], last7Days[len(last7Days)-1-i] = last7Days[len(last7Days)-1-i], last7Days[i]
	}
	for i := 0; i < len(last30Days)/2; i++ {
		last30Days[i], last30Days[len(last30Days)-1-i] = last30Days[len(last30Days)-1-i], last30Days[i]
	}
	for i := 0; i < len(weeklyStats)/2; i++ {
		weeklyStats[i], weeklyStats[len(weeklyStats)-1-i] = weeklyStats[len(weeklyStats)-1-i], weeklyStats[i]
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Stats - Health Hub</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body class="bg-gray-100 min-h-screen">
    <div class="container mx-auto px-4 py-8 max-w-6xl">
        <!-- Header -->
        <div class="flex justify-between items-center mb-8">
            <div>
                <h1 class="text-4xl font-bold text-gray-900 mb-2">Activity Stats</h1>
                <p class="text-gray-600">Your fitness journey overview</p>
            </div>
            <div class="flex items-center space-x-4">
                <!-- Unit Toggle -->
                <div class="flex items-center space-x-2">
                    <span class="text-sm text-gray-600">Units:</span>
                    <button id="unit-toggle" class="{{if .UseImperial}}bg-orange-500{{else}}bg-blue-500{{end}} text-white px-3 py-1 rounded text-sm font-medium hover:opacity-80 transition-opacity">
                        {{if .UseImperial}}Imperial{{else}}Metric{{end}}
                    </button>
                </div>
                <a href="/" class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded transition duration-200">
                    Back to Home
                </a>
            </div>
        </div>

        <!-- Overall Stats -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
            <div class="bg-white rounded-lg shadow-md p-6">
                <h3 class="text-lg font-semibold text-gray-900 mb-2">Total Distance</h3>
                <p class="text-3xl font-bold text-blue-600">
                    {{if .UseImperial}}{{printf "%.1f" (metersToMiles .TotalDistance)}} mi{{else}}{{printf "%.1f" (metersToKm .TotalDistance)}} km{{end}}
                </p>
                <p class="text-sm text-gray-600">All time</p>
            </div>
            <div class="bg-white rounded-lg shadow-md p-6">
                <h3 class="text-lg font-semibold text-gray-900 mb-2">Total Activities</h3>
                <p class="text-3xl font-bold text-green-600">{{.TotalActivities}}</p>
                <p class="text-sm text-gray-600">Completed</p>
            </div>
            <div class="bg-white rounded-lg shadow-md p-6">
                <h3 class="text-lg font-semibold text-gray-900 mb-2">Total Time</h3>
                <p class="text-3xl font-bold text-purple-600">{{printf "%.1f" (div .TotalDuration 3600)}} hrs</p>
                <p class="text-sm text-gray-600">Moving time</p>
            </div>
        </div>

        <!-- Charts Section -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
            <!-- Last 7 Days Chart -->
            <div class="bg-white rounded-lg shadow-md p-6">
                <h3 class="text-xl font-semibold text-gray-900 mb-4">Last 7 Days</h3>
                <canvas id="last7DaysChart" width="400" height="200"></canvas>
            </div>

            <!-- Weekly Trends Chart -->
            <div class="bg-white rounded-lg shadow-md p-6">
                <h3 class="text-xl font-semibold text-gray-900 mb-4">Weekly Trends (Last 4 Weeks)</h3>
                <canvas id="weeklyChart" width="400" height="200"></canvas>
            </div>
        </div>

        <!-- Monthly Overview -->
        <div class="bg-white rounded-lg shadow-md p-6">
            <h3 class="text-xl font-semibold text-gray-900 mb-4">Last 30 Days Overview</h3>
            <canvas id="monthlyChart" width="800" height="300"></canvas>
        </div>
    </div>

    <script>
        // Chart.js configuration
        Chart.defaults.font.family = 'Arial, sans-serif';
        Chart.defaults.color = '#374151';

        // Last 7 Days Chart
        const ctx7Days = document.getElementById('last7DaysChart').getContext('2d');
        const last7DaysData = {
            labels: [{{range .Last7Days}}'{{.Date.Format "Jan 2"}}',{{end}}],
            datasets: [{
                label: 'Distance (km)',
                data: [{{range .Last7Days}}{{printf "%.1f" (div .Distance 1000)}},{{end}}],
                backgroundColor: 'rgba(59, 130, 246, 0.1)',
                borderColor: 'rgba(59, 130, 246, 1)',
                borderWidth: 2,
                fill: true,
                tension: 0.4
            }]
        };
        new Chart(ctx7Days, {
            type: 'line',
            data: last7DaysData,
            options: {
                responsive: true,
                scales: {
                    y: {
                        beginAtZero: true,
                        title: {
                            display: true,
                            text: 'Distance (km)'
                        }
                    }
                }
            }
        });

        // Weekly Trends Chart
        const ctxWeekly = document.getElementById('weeklyChart').getContext('2d');
        const weeklyData = {
            labels: [{{range .WeeklyStats}}'{{.Week}}',{{end}}],
            datasets: [{
                label: 'Distance (km)',
                data: [{{range .WeeklyStats}}{{printf "%.1f" (div .Distance 1000)}},{{end}}],
                backgroundColor: 'rgba(16, 185, 129, 0.6)',
                borderColor: 'rgba(16, 185, 129, 1)',
                borderWidth: 2
            }, {
                label: 'Activities',
                data: [{{range .WeeklyStats}}{{.Activities}},{{end}}],
                backgroundColor: 'rgba(245, 158, 11, 0.6)',
                borderColor: 'rgba(245, 158, 11, 1)',
                borderWidth: 2,
                yAxisID: 'y1'
            }]
        };
        new Chart(ctxWeekly, {
            type: 'bar',
            data: weeklyData,
            options: {
                responsive: true,
                scales: {
                    y: {
                        beginAtZero: true,
                        title: {
                            display: true,
                            text: 'Distance (km)'
                        }
                    },
                    y1: {
                        type: 'linear',
                        display: true,
                        position: 'right',
                        title: {
                            display: true,
                            text: 'Activities'
                        },
                        grid: {
                            drawOnChartArea: false,
                        }
                    }
                }
            }
        });

        // Monthly Overview Chart
        const ctxMonthly = document.getElementById('monthlyChart').getContext('2d');
        const monthlyData = {
            labels: [{{range .Last30Days}}'{{.Date.Format "Jan 2"}}',{{end}}],
            datasets: [{
                label: 'Distance (km)',
                data: [{{range .Last30Days}}{{printf "%.1f" (div .Distance 1000)}},{{end}}],
                backgroundColor: 'rgba(139, 92, 246, 0.1)',
                borderColor: 'rgba(139, 92, 246, 1)',
                borderWidth: 2,
                fill: true,
                tension: 0.4
            }]
        };
        new Chart(ctxMonthly, {
            type: 'line',
            data: monthlyData,
            options: {
                responsive: true,
                scales: {
                    y: {
                        beginAtZero: true,
                        title: {
                            display: true,
                            text: 'Distance (km)'
                        }
                    }
                }
            }
        });

        // Unit toggle functionality
        document.getElementById('unit-toggle').addEventListener('click', function() {
            const currentUnit = this.textContent.trim();
            const newUnit = currentUnit === 'Metric' ? 'imperial' : 'metric';
            
            // Set cookie
            document.cookie = 'units=' + newUnit + '; path=/; max-age=' + (365 * 24 * 60 * 60);
            
            // Reload page to apply new units
            window.location.reload();
        });
    </script>
</body>
</html>`

	data := StatsData{
		TotalDistance:   totalDistance,
		TotalActivities: totalActivities,
		TotalDuration:   totalDuration,
		Last7Days:       last7Days,
		Last30Days:      last30Days,
		WeeklyStats:     weeklyStats,
		UseImperial:     useImperial,
	}

	funcMap := template.FuncMap{
		"div": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"metersToKm": func(meters float64) float64 {
			return meters / 1000
		},
		"metersToMiles": func(meters float64) float64 {
			return meters * 0.000621371
		},
	}

	t, err := template.New("stats").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(w, data)
}

// Helper types for stats calculations
type ActivityStat struct {
	Date       time.Time
	Distance   float64
	Activities int
	Duration   float64
}

type WeekStat struct {
	Week       string
	StartDate  time.Time
	EndDate    time.Time
	Distance   float64
	Activities int
	Duration   float64
}

type StatsData struct {
	TotalDistance   float64
	TotalActivities int
	TotalDuration   float64
	Last7Days       []ActivityStat
	Last30Days      []ActivityStat
	WeeklyStats     []WeekStat
	UseImperial     bool
}

func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func (h *Handlers) BulkUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Bulk Upload - Health Hub</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100 min-h-screen">
    <div class="container mx-auto px-4 py-8 max-w-4xl">
        <!-- Header -->
        <div class="flex justify-between items-center mb-8">
            <div>
                <h1 class="text-4xl font-bold text-gray-900 mb-2">Bulk Upload</h1>
                <p class="text-gray-600">Upload multiple GPX files at once</p>
            </div>
            <a href="/" class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded transition duration-200">
                Back to Home
            </a>
        </div>

        <!-- Upload Section -->
        <div class="bg-white rounded-lg shadow-md p-8 mb-8">
            <div class="text-center mb-8">
                <div class="mx-auto w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mb-4">
                    <svg class="w-8 h-8 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"></path>
                    </svg>
                </div>
                <h2 class="text-2xl font-bold text-gray-900 mb-2">Upload Your GPX Files</h2>
                <p class="text-gray-600">Select multiple GPX files or drag and drop them here</p>
            </div>

            <!-- Drag and Drop Area -->
            <div id="drop-zone" class="border-2 border-dashed border-blue-300 rounded-lg p-8 text-center hover:border-blue-400 transition-colors cursor-pointer">
                <form id="bulk-upload-form" hx-post="/api/upload/bulk-gpx" hx-encoding="multipart/form-data" 
                      hx-target="#upload-results" hx-swap="innerHTML" hx-indicator="#upload-progress">
                    <input type="file" id="file-input" name="gpx-files" accept=".gpx" multiple required 
                           class="hidden">
                    <div id="file-list" class="mb-4 hidden">
                        <h3 class="text-lg font-semibold text-gray-900 mb-2">Selected Files:</h3>
                        <div id="selected-files" class="space-y-2"></div>
                    </div>
                    <div id="drop-text" class="mb-6">
                        <p class="text-xl text-gray-600 mb-2">Drop GPX files here or</p>
                        <button type="button" onclick="document.getElementById('file-input').click()" 
                                class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-3 px-6 rounded-lg transition duration-200">
                            Select Files
                        </button>
                    </div>
                    <button type="submit" id="upload-btn" class="hidden bg-green-500 hover:bg-green-700 text-white font-bold py-3 px-8 rounded-lg transition duration-200">
                        Upload All Files
                    </button>
                </form>
            </div>

            <!-- Progress Indicator -->
            <div id="upload-progress" class="htmx-indicator mt-6">
                <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
                    <div class="flex items-center">
                        <div class="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600 mr-3"></div>
                        <span class="text-blue-800 font-medium">Processing files...</span>
                    </div>
                    <div class="mt-2 bg-white rounded-full h-2">
                        <div class="bg-blue-600 h-2 rounded-full animate-pulse" style="width: 45%"></div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Upload Results -->
        <div id="upload-results" class="space-y-4"></div>

        <!-- Instructions -->
        <div class="bg-blue-50 border border-blue-200 rounded-lg p-6">
            <h3 class="text-lg font-semibold text-blue-900 mb-3">📋 Upload Instructions</h3>
            <ul class="text-blue-800 space-y-2">
                <li>• Select multiple GPX files (you can Ctrl+click or Cmd+click to select multiple files)</li>
                <li>• Drag and drop files directly onto the upload area</li>
                <li>• Each file will be processed individually with detailed progress</li>
                <li>• Invalid files will be skipped with error messages</li>
                <li>• Successfully uploaded activities will appear in your activity log</li>
            </ul>
        </div>
    </div>

    <script>
        const dropZone = document.getElementById('drop-zone');
        const fileInput = document.getElementById('file-input');
        const fileList = document.getElementById('file-list');
        const selectedFiles = document.getElementById('selected-files');
        const uploadBtn = document.getElementById('upload-btn');
        const dropText = document.getElementById('drop-text');

        // Drag and drop functionality
        dropZone.addEventListener('dragover', (e) => {
            e.preventDefault();
            dropZone.classList.add('border-blue-500', 'bg-blue-50');
        });

        dropZone.addEventListener('dragleave', (e) => {
            e.preventDefault();
            dropZone.classList.remove('border-blue-500', 'bg-blue-50');
        });

        dropZone.addEventListener('drop', (e) => {
            e.preventDefault();
            dropZone.classList.remove('border-blue-500', 'bg-blue-50');
            
            const files = Array.from(e.dataTransfer.files).filter(file => 
                file.name.toLowerCase().endsWith('.gpx')
            );
            
            if (files.length > 0) {
                // Create a new FileList-like object
                const dt = new DataTransfer();
                files.forEach(file => dt.items.add(file));
                fileInput.files = dt.files;
                updateFileList(files);
            }
        });

        // File input change handler
        fileInput.addEventListener('change', (e) => {
            const files = Array.from(e.target.files);
            updateFileList(files);
        });

        function updateFileList(files) {
            if (files.length === 0) {
                fileList.classList.add('hidden');
                uploadBtn.classList.add('hidden');
                dropText.classList.remove('hidden');
                return;
            }

            selectedFiles.innerHTML = '';
            files.forEach((file, index) => {
                const fileItem = document.createElement('div');
                fileItem.className = 'flex items-center justify-between bg-gray-50 p-3 rounded';
                fileItem.innerHTML = 
                    '<div class="flex items-center">' +
                        '<svg class="w-5 h-5 text-green-600 mr-2" fill="currentColor" viewBox="0 0 20 20">' +
                            '<path fill-rule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm3.293-7.707a1 1 0 011.414 0L9 10.586V3a1 1 0 112 0v7.586l1.293-1.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z" clip-rule="evenodd"></path>' +
                        '</svg>' +
                        '<span class="text-gray-700">' + file.name + '</span>' +
                        '<span class="text-gray-500 text-sm ml-2">(' + (file.size / 1024).toFixed(1) + ' KB)</span>' +
                    '</div>';
                selectedFiles.appendChild(fileItem);
            });

            fileList.classList.remove('hidden');
            uploadBtn.classList.remove('hidden');
            dropText.classList.add('hidden');
        }

        // Click to select files
        dropZone.addEventListener('click', (e) => {
            if (e.target === dropZone || e.target.closest('#drop-text')) {
                fileInput.click();
            }
        });
    </script>
</body>
</html>`

	t, _ := template.New("bulk-upload").Parse(tmpl)
	t.Execute(w, nil)
}

func (h *Handlers) BulkUploadGPX(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form with larger memory limit for multiple files
	err := r.ParseMultipartForm(100 << 20) // 100MB limit
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["gpx-files"]
	if len(files) == 0 {
		http.Error(w, "No files selected", http.StatusBadRequest)
		return
	}

	var results []BulkUploadResult
	successCount := 0
	errorCount := 0

	for i, fileHeader := range files {
		result := BulkUploadResult{
			FileName: fileHeader.Filename,
			Index:    i + 1,
			Total:    len(files),
		}

		// Open the file
		file, err := fileHeader.Open()
		if err != nil {
			result.Status = "error"
			result.Error = "Failed to open file"
			errorCount++
			results = append(results, result)
			continue
		}

		// Read file content
		data, err := ioutil.ReadAll(file)
		file.Close()
		if err != nil {
			result.Status = "error"
			result.Error = "Failed to read file content"
			errorCount++
			results = append(results, result)
			continue
		}

		// Save the raw GPX file
		filename := fmt.Sprintf("gpx_%d_%s", time.Now().UnixNano(), fileHeader.Filename)
		if err := h.storage.SaveFile(filename, data); err != nil {
			result.Status = "error"
			result.Error = "Failed to save file"
			errorCount++
			results = append(results, result)
			continue
		}

		// Parse GPX and create activity record
		track, activity, err := gpx.ParseGPX(string(data))
		if err != nil {
			result.Status = "error"
			result.Error = fmt.Sprintf("Invalid GPX format: %v", err)
			errorCount++
			results = append(results, result)
			continue
		}

		// Set additional activity details
		if activity.Name == "" {
			activity.Name = strings.TrimSuffix(fileHeader.Filename, ".gpx")
		}
		activity.GPXFile = filename

		// Save activity first to get the ID, then set track ID to match
		if err := h.storage.SaveActivity(activity); err != nil {
			result.Status = "error"
			result.Error = "Failed to save activity"
			errorCount++
			results = append(results, result)
			continue
		}

		// Link track to activity by using the same ID
		track.ID = activity.ID
		if err := h.storage.SaveGPXTrack(track); err != nil {
			result.Status = "error"
			result.Error = "Failed to save GPS track"
			errorCount++
			results = append(results, result)
			continue
		}

		result.Status = "success"
		result.ActivityName = activity.Name
		result.Distance = activity.Distance / 1000 // Convert to km
		result.Duration = activity.Duration
		successCount++
		results = append(results, result)
	}

	// Generate HTML response
	html := fmt.Sprintf(`
	<div class="bg-white rounded-lg shadow-md p-6">
		<div class="flex items-center justify-between mb-6">
			<h3 class="text-xl font-semibold text-gray-900">Upload Results</h3>
			<div class="flex space-x-4">
				<span class="bg-green-100 text-green-800 px-3 py-1 rounded-full text-sm font-medium">
					✓ %d Successful
				</span>
				<span class="bg-red-100 text-red-800 px-3 py-1 rounded-full text-sm font-medium">
					✗ %d Failed
				</span>
			</div>
		</div>
		<div class="space-y-3 max-h-96 overflow-y-auto">`, successCount, errorCount)

	for _, result := range results {
		if result.Status == "success" {
			html += fmt.Sprintf(`
			<div class="flex items-center justify-between bg-green-50 border border-green-200 rounded-lg p-4">
				<div class="flex items-center">
					<svg class="w-5 h-5 text-green-600 mr-3" fill="currentColor" viewBox="0 0 20 20">
						<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"></path>
					</svg>
					<div>
						<p class="font-medium text-green-900">%s</p>
						<p class="text-sm text-green-700">Activity: %s • %.1f km • %d:%02d duration</p>
					</div>
				</div>
				<span class="text-green-600 text-sm font-medium">%d/%d</span>
			</div>`, result.FileName, result.ActivityName, result.Distance, result.Duration/3600, (result.Duration%3600)/60, result.Index, result.Total)
		} else {
			html += fmt.Sprintf(`
			<div class="flex items-center justify-between bg-red-50 border border-red-200 rounded-lg p-4">
				<div class="flex items-center">
					<svg class="w-5 h-5 text-red-600 mr-3" fill="currentColor" viewBox="0 0 20 20">
						<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd"></path>
					</svg>
					<div>
						<p class="font-medium text-red-900">%s</p>
						<p class="text-sm text-red-700">%s</p>
					</div>
				</div>
				<span class="text-red-600 text-sm font-medium">%d/%d</span>
			</div>`, result.FileName, result.Error, result.Index, result.Total)
		}
	}

	html += `
		</div>
		<div class="mt-6 flex justify-between items-center">
			<button onclick="location.reload()" class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded transition duration-200">
				Upload More Files
			</button>
			<a href="/activities" class="bg-green-500 hover:bg-green-700 text-white font-bold py-2 px-4 rounded transition duration-200">
				View Activities
			</a>
		</div>
	</div>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (h *Handlers) ActivityDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract activity ID from URL path
	path := r.URL.Path
	activityID := strings.TrimPrefix(path, "/activity/")
	
	if activityID == "" {
		http.Error(w, "Activity ID required", http.StatusBadRequest)
		return
	}

	// Get all activities and find the specific one
	activities, err := h.storage.GetActivities()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var activity *models.Activity
	for _, a := range activities {
		if a.ID == activityID {
			activity = a
			break
		}
	}

	if activity == nil {
		http.Error(w, "Activity not found", http.StatusNotFound)
		return
	}

	// Get unit preference from cookie
	useImperial := false
	if cookie, err := r.Cookie("units"); err == nil && cookie.Value == "imperial" {
		useImperial = true
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Activity.Name}} - Health Hub</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100 min-h-screen">
    <div class="container mx-auto px-4 py-8 max-w-6xl">
        <!-- Header -->
        <div class="flex justify-between items-center mb-8">
            <div>
                <h1 class="text-4xl font-bold text-gray-900 mb-2">{{.Activity.Name}}</h1>
                <p class="text-gray-600">Activity Details</p>
            </div>
            <div class="flex items-center space-x-4">
                <!-- Unit Toggle -->
                <div class="flex items-center space-x-2">
                    <span class="text-sm text-gray-600">Units:</span>
                    <button id="unit-toggle" class="{{if .UseImperial}}bg-orange-500{{else}}bg-blue-500{{end}} text-white px-3 py-1 rounded text-sm font-medium hover:opacity-80 transition-opacity">
                        {{if .UseImperial}}Imperial{{else}}Metric{{end}}
                    </button>
                </div>
                <a href="/activities" class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded transition duration-200">
                    Back to Activities
                </a>
            </div>
        </div>

        <!-- Activity Overview -->
        <div class="bg-white rounded-lg shadow-md p-8 mb-8">
            <div class="flex items-center mb-6">
                <div class="flex-shrink-0 h-16 w-16 mr-6">
                    <div class="h-16 w-16 rounded-full bg-{{getTypeColor .Activity.Type}}-100 flex items-center justify-center">
                        <span class="text-{{getTypeColor .Activity.Type}}-600 font-medium text-2xl">{{getTypeIcon .Activity.Type}}</span>
                    </div>
                </div>
                <div>
                    <h2 class="text-2xl font-bold text-gray-900">{{.Activity.Name}}</h2>
                    <div class="flex items-center space-x-4 mt-2">
                        <span class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-{{getTypeColor .Activity.Type}}-100 text-{{getTypeColor .Activity.Type}}-800">
                            {{.Activity.Type}}
                        </span>
                        {{if .Activity.GPXFile}}
                        <span class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-green-100 text-green-800">
                            📍 GPS Track
                        </span>
                        {{end}}
                        <span class="text-gray-500 text-sm">{{.Activity.StartTime.Format "Monday, January 2, 2006 at 3:04 PM"}}</span>
                    </div>
                </div>
            </div>

            <!-- Main Stats Grid -->
            <div class="grid grid-cols-2 md:grid-cols-4 gap-6 mb-8">
                <div class="text-center">
                    <div class="text-3xl font-bold text-blue-600">
                        {{if .UseImperial}}{{printf "%.1f" (metersToMiles .Activity.Distance)}}{{else}}{{printf "%.1f" (metersToKm .Activity.Distance)}}{{end}}
                    </div>
                    <div class="text-sm text-gray-500 uppercase tracking-wide">
                        {{if .UseImperial}}Miles{{else}}Kilometers{{end}}
                    </div>
                </div>
                <div class="text-center">
                    <div class="text-3xl font-bold text-green-600">{{formatDuration .Activity.Duration}}</div>
                    <div class="text-sm text-gray-500 uppercase tracking-wide">Duration</div>
                </div>
                <div class="text-center">
                    <div class="text-3xl font-bold text-purple-600">
                        {{if .UseImperial}}{{printf "%.1f" (kmhToMph .Activity.AvgSpeed)}}{{else}}{{printf "%.1f" .Activity.AvgSpeed}}{{end}}
                    </div>
                    <div class="text-sm text-gray-500 uppercase tracking-wide">
                        Avg {{if .UseImperial}}mph{{else}}km/h{{end}}
                    </div>
                </div>
                <div class="text-center">
                    <div class="text-3xl font-bold text-orange-600">
                        {{if .UseImperial}}{{printf "%.0f" (metersToFeet .Activity.TotalElevation)}}{{else}}{{printf "%.0f" .Activity.TotalElevation}}{{end}}
                    </div>
                    <div class="text-sm text-gray-500 uppercase tracking-wide">
                        {{if .UseImperial}}Feet{{else}}Meters{{end}} Elevation
                    </div>
                </div>
            </div>
        </div>

        <!-- Detailed Stats -->
        <div class="grid md:grid-cols-2 gap-8">
            <!-- Performance Stats -->
            <div class="bg-white rounded-lg shadow-md p-6">
                <h3 class="text-xl font-bold text-gray-900 mb-6">Performance</h3>
                <div class="space-y-4">
                    <div class="flex justify-between items-center py-3 border-b border-gray-100">
                        <span class="text-gray-600">Average Speed</span>
                        <span class="font-semibold">
                            {{if .UseImperial}}{{printf "%.1f" (kmhToMph .Activity.AvgSpeed)}} mph{{else}}{{printf "%.1f" .Activity.AvgSpeed}} km/h{{end}}
                        </span>
                    </div>
                    <div class="flex justify-between items-center py-3 border-b border-gray-100">
                        <span class="text-gray-600">Max Speed</span>
                        <span class="font-semibold">
                            {{if .UseImperial}}{{printf "%.1f" (kmhToMph .Activity.MaxSpeed)}} mph{{else}}{{printf "%.1f" .Activity.MaxSpeed}} km/h{{end}}
                        </span>
                    </div>
                    <div class="flex justify-between items-center py-3 border-b border-gray-100">
                        <span class="text-gray-600">Average Pace</span>
                        <span class="font-semibold">{{calculatePace .Activity.Duration .Activity.Distance .UseImperial}}</span>
                    </div>
                    {{if .Activity.Calories}}
                    <div class="flex justify-between items-center py-3 border-b border-gray-100">
                        <span class="text-gray-600">Calories</span>
                        <span class="font-semibold">{{.Activity.Calories}} cal</span>
                    </div>
                    {{end}}
                </div>
            </div>

            <!-- Activity Details -->
            <div class="bg-white rounded-lg shadow-md p-6">
                <h3 class="text-xl font-bold text-gray-900 mb-6">Details</h3>
                <div class="space-y-4">
                    <div class="flex justify-between items-center py-3 border-b border-gray-100">
                        <span class="text-gray-600">Start Time</span>
                        <span class="font-semibold">{{.Activity.StartTime.Format "3:04 PM"}}</span>
                    </div>
                    <div class="flex justify-between items-center py-3 border-b border-gray-100">
                        <span class="text-gray-600">End Time</span>
                        <span class="font-semibold">{{.Activity.EndTime.Format "3:04 PM"}}</span>
                    </div>
                    <div class="flex justify-between items-center py-3 border-b border-gray-100">
                        <span class="text-gray-600">Total Distance</span>
                        <span class="font-semibold">
                            {{if .UseImperial}}{{printf "%.2f" (metersToMiles .Activity.Distance)}} miles{{else}}{{printf "%.2f" (metersToKm .Activity.Distance)}} km{{end}}
                        </span>
                    </div>
                    <div class="flex justify-between items-center py-3 border-b border-gray-100">
                        <span class="text-gray-600">Elevation Gain</span>
                        <span class="font-semibold">
                            {{if .UseImperial}}{{printf "%.0f" (metersToFeet .Activity.TotalElevation)}} ft{{else}}{{printf "%.0f" .Activity.TotalElevation}} m{{end}}
                        </span>
                    </div>
                    {{if .Activity.TotalPoints}}
                    <div class="flex justify-between items-center py-3 border-b border-gray-100">
                        <span class="text-gray-600">GPS Points</span>
                        <span class="font-semibold">{{.Activity.TotalPoints}}</span>
                    </div>
                    {{end}}
                </div>
            </div>
        </div>

        <!-- Activity Actions -->
        <div class="bg-white rounded-lg shadow-md p-6 mt-8">
            <h3 class="text-xl font-bold text-gray-900 mb-4">Actions</h3>
            <div class="flex space-x-4">
                <a href="/activities" class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded transition duration-200">
                    ← Back to Activities
                </a>
                {{if .Activity.GPXFile}}
                <a href="/gps-track/{{.Activity.ID}}" class="bg-green-500 hover:bg-green-700 text-white font-bold py-2 px-4 rounded transition duration-200">
                    📍 View GPS Track
                </a>
                {{end}}
                <a href="/stats" class="bg-purple-500 hover:bg-purple-700 text-white font-bold py-2 px-4 rounded transition duration-200">
                    📊 View Stats
                </a>
            </div>
        </div>
    </div>

    <script>
        // Unit toggle functionality
        document.getElementById('unit-toggle').addEventListener('click', function() {
            const currentUnit = this.textContent.trim();
            const newUnit = currentUnit === 'Metric' ? 'imperial' : 'metric';
            
            // Set cookie
            document.cookie = 'units=' + newUnit + '; path=/; max-age=' + (365 * 24 * 60 * 60);
            
            // Reload page to apply new units
            window.location.reload();
        });
    </script>
</body>
</html>`

	funcMap := template.FuncMap{
		"metersToKm": func(meters float64) float64 {
			return meters / 1000
		},
		"metersToMiles": func(meters float64) float64 {
			return meters * 0.000621371
		},
		"metersToFeet": func(meters float64) float64 {
			return meters * 3.28084
		},
		"kmhToMph": func(kmh float64) float64 {
			return kmh * 0.621371
		},
		"formatDuration": func(seconds int) string {
			hours := seconds / 3600
			minutes := (seconds % 3600) / 60
			if hours > 0 {
				return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds%60)
			}
			return fmt.Sprintf("%d:%02d", minutes, seconds%60)
		},
		"calculatePace": func(durationSeconds int, distanceMeters float64, useImperial bool) string {
			if distanceMeters == 0 {
				return "N/A"
			}
			
			var distance float64
			var unit string
			
			if useImperial {
				distance = distanceMeters * 0.000621371 // to miles
				unit = "/mi"
			} else {
				distance = distanceMeters / 1000 // to km
				unit = "/km"
			}
			
			paceSeconds := float64(durationSeconds) / distance
			paceMinutes := int(paceSeconds) / 60
			paceSecondsRemainder := int(paceSeconds) % 60
			
			return fmt.Sprintf("%d:%02d%s", paceMinutes, paceSecondsRemainder, unit)
		},
		"getTypeColor": func(activityType string) string {
			switch strings.ToLower(activityType) {
			case "running":
				return "red"
			case "cycling":
				return "blue"
			case "walking":
				return "green"
			case "hiking":
				return "yellow"
			default:
				return "gray"
			}
		},
		"getTypeIcon": func(activityType string) string {
			switch strings.ToLower(activityType) {
			case "running":
				return "🏃"
			case "cycling":
				return "🚴"
			case "walking":
				return "🚶"
			case "hiking":
				return "🥾"
			default:
				return "🏃"
			}
		},
	}

	data := struct {
		Activity    *models.Activity
		UseImperial bool
	}{
		Activity:    activity,
		UseImperial: useImperial,
	}

	t, err := template.New("activity-detail").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(w, data)
}

func (h *Handlers) GPSTrack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract activity ID from URL path
	path := r.URL.Path
	activityID := strings.TrimPrefix(path, "/gps-track/")
	
	if activityID == "" {
		http.Error(w, "Activity ID required", http.StatusBadRequest)
		return
	}

	// Get activity
	activities, err := h.storage.GetActivities()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var activity *models.Activity
	for _, a := range activities {
		if a.ID == activityID {
			activity = a
			break
		}
	}

	if activity == nil {
		http.Error(w, "Activity not found", http.StatusNotFound)
		return
	}

	if activity.GPXFile == "" {
		http.Error(w, "No GPS track available for this activity", http.StatusNotFound)
		return
	}

	// Get GPX track data
	gpxTracks, err := h.storage.GetGPXTracks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var track *models.GPXTrack
	for _, t := range gpxTracks {
		if t.ID == activity.ID { // GPX tracks use the same ID as activities
			track = t
			break
		}
	}

	if track == nil {
		http.Error(w, "GPS track data not found", http.StatusNotFound)
		return
	}

	// Get unit preference
	useImperial := false
	if cookie, err := r.Cookie("units"); err == nil && cookie.Value == "imperial" {
		useImperial = true
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GPS Track - {{.Activity.Name}} - Health Hub</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" />
    <script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"></script>
</head>
<body class="bg-gray-100 min-h-screen">
    <div class="container mx-auto px-4 py-8 max-w-7xl">
        <!-- Header -->
        <div class="flex justify-between items-center mb-8">
            <div>
                <h1 class="text-4xl font-bold text-gray-900 mb-2">GPS Track</h1>
                <p class="text-gray-600">{{.Activity.Name}} - {{.Activity.StartTime.Format "Jan 2, 2006"}}</p>
            </div>
            <div class="flex items-center space-x-4">
                <!-- Unit Toggle -->
                <div class="flex items-center space-x-2">
                    <span class="text-sm text-gray-600">Units:</span>
                    <button id="unit-toggle" class="{{if .UseImperial}}bg-orange-500{{else}}bg-blue-500{{end}} text-white px-3 py-1 rounded text-sm font-medium hover:opacity-80 transition-opacity">
                        {{if .UseImperial}}Imperial{{else}}Metric{{end}}
                    </button>
                </div>
                <a href="/activity/{{.Activity.ID}}" class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded transition duration-200">
                    Back to Activity
                </a>
            </div>
        </div>

        <!-- Quick Stats -->
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <div class="bg-white rounded-lg shadow-md p-4 text-center">
                <div class="text-2xl font-bold text-blue-600">
                    {{if .UseImperial}}{{printf "%.1f" (metersToMiles .Activity.Distance)}}{{else}}{{printf "%.1f" (metersToKm .Activity.Distance)}}{{end}}
                </div>
                <div class="text-sm text-gray-500">{{if .UseImperial}}Miles{{else}}Kilometers{{end}}</div>
            </div>
            <div class="bg-white rounded-lg shadow-md p-4 text-center">
                <div class="text-2xl font-bold text-green-600">{{formatDuration .Activity.Duration}}</div>
                <div class="text-sm text-gray-500">Duration</div>
            </div>
            <div class="bg-white rounded-lg shadow-md p-4 text-center">
                <div class="text-2xl font-bold text-purple-600">
                    {{if .UseImperial}}{{printf "%.1f" (kmhToMph .Activity.AvgSpeed)}}{{else}}{{printf "%.1f" .Activity.AvgSpeed}}{{end}}
                </div>
                <div class="text-sm text-gray-500">Avg {{if .UseImperial}}mph{{else}}km/h{{end}}</div>
            </div>
            <div class="bg-white rounded-lg shadow-md p-4 text-center">
                <div class="text-2xl font-bold text-orange-600">{{.Track.TotalPoints}}</div>
                <div class="text-sm text-gray-500">GPS Points</div>
            </div>
        </div>

        <!-- Map Container -->
        <div class="bg-white rounded-lg shadow-md overflow-hidden mb-6">
            <div class="p-4 border-b border-gray-200">
                <h3 class="text-lg font-semibold text-gray-900">GPS Track Map</h3>
                <p class="text-sm text-gray-600">Interactive map showing your route</p>
            </div>
            <div id="map" style="height: 500px; width: 100%;"></div>
        </div>

        <!-- Elevation Profile -->
        <div class="bg-white rounded-lg shadow-md p-6 mb-6">
            <h3 class="text-lg font-semibold text-gray-900 mb-4">Elevation Profile</h3>
            <canvas id="elevationChart" width="800" height="200"></canvas>
        </div>

        <!-- Track Statistics -->
        <div class="grid md:grid-cols-2 gap-6">
            <div class="bg-white rounded-lg shadow-md p-6">
                <h3 class="text-lg font-semibold text-gray-900 mb-4">Track Details</h3>
                <div class="space-y-3">
                    <div class="flex justify-between">
                        <span class="text-gray-600">Total Points:</span>
                        <span class="font-semibold">{{.Track.TotalPoints}}</span>
                    </div>
                    <div class="flex justify-between">
                        <span class="text-gray-600">Start Point:</span>
                        <span class="font-semibold">{{printf "%.6f, %.6f" .Track.StartLat .Track.StartLon}}</span>
                    </div>
                    <div class="flex justify-between">
                        <span class="text-gray-600">End Point:</span>
                        <span class="font-semibold">{{printf "%.6f, %.6f" .Track.EndLat .Track.EndLon}}</span>
                    </div>
                    <div class="flex justify-between">
                        <span class="text-gray-600">Min Elevation:</span>
                        <span class="font-semibold">
                            {{if .UseImperial}}{{printf "%.0f ft" (metersToFeet .Track.MinElevation)}}{{else}}{{printf "%.0f m" .Track.MinElevation}}{{end}}
                        </span>
                    </div>
                    <div class="flex justify-between">
                        <span class="text-gray-600">Max Elevation:</span>
                        <span class="font-semibold">
                            {{if .UseImperial}}{{printf "%.0f ft" (metersToFeet .Track.MaxElevation)}}{{else}}{{printf "%.0f m" .Track.MaxElevation}}{{end}}
                        </span>
                    </div>
                </div>
            </div>

            <div class="bg-white rounded-lg shadow-md p-6">
                <h3 class="text-lg font-semibold text-gray-900 mb-4">Activity Info</h3>
                <div class="space-y-3">
                    <div class="flex justify-between">
                        <span class="text-gray-600">Activity Type:</span>
                        <span class="font-semibold">{{.Activity.Type}}</span>
                    </div>
                    <div class="flex justify-between">
                        <span class="text-gray-600">Start Time:</span>
                        <span class="font-semibold">{{.Activity.StartTime.Format "3:04 PM"}}</span>
                    </div>
                    <div class="flex justify-between">
                        <span class="text-gray-600">End Time:</span>
                        <span class="font-semibold">{{.Activity.EndTime.Format "3:04 PM"}}</span>
                    </div>
                    <div class="flex justify-between">
                        <span class="text-gray-600">Max Speed:</span>
                        <span class="font-semibold">
                            {{if .UseImperial}}{{printf "%.1f mph" (kmhToMph .Activity.MaxSpeed)}}{{else}}{{printf "%.1f km/h" .Activity.MaxSpeed}}{{end}}
                        </span>
                    </div>
                    <div class="flex justify-between">
                        <span class="text-gray-600">Total Elevation:</span>
                        <span class="font-semibold">
                            {{if .UseImperial}}{{printf "%.0f ft" (metersToFeet .Activity.TotalElevation)}}{{else}}{{printf "%.0f m" .Activity.TotalElevation}}{{end}}
                        </span>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script>
        // Unit toggle functionality
        document.getElementById('unit-toggle').addEventListener('click', function() {
            const currentUnit = this.textContent.trim();
            const newUnit = currentUnit === 'Metric' ? 'imperial' : 'metric';
            
            // Set cookie
            document.cookie = 'units=' + newUnit + '; path=/; max-age=' + (365 * 24 * 60 * 60);
            
            // Reload page to apply new units
            window.location.reload();
        });

        // Initialize map
        const map = L.map('map');

        // Add OpenStreetMap tiles
        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            attribution: '© OpenStreetMap contributors'
        }).addTo(map);

        // GPS track data
        const trackPoints = [
            {{range .Track.Points}}
            [{{.Lat}}, {{.Lon}}, {{.Elevation}}],
            {{end}}
        ];

        if (trackPoints.length > 0) {
            // Create polyline for the track
            const track = L.polyline(trackPoints.map(p => [p[0], p[1]]), {
                color: '#3B82F6',
                weight: 4,
                opacity: 0.8
            }).addTo(map);

            // Add start marker
            L.marker([trackPoints[0][0], trackPoints[0][1]], {
                icon: L.divIcon({
                    className: 'custom-div-icon',
                    html: '<div style="background-color: #10B981; color: white; border-radius: 50%; width: 20px; height: 20px; display: flex; align-items: center; justify-content: center; font-weight: bold; font-size: 12px;">S</div>',
                    iconSize: [20, 20],
                    iconAnchor: [10, 10]
                })
            }).addTo(map).bindPopup('Start');

            // Add end marker
            const lastPoint = trackPoints[trackPoints.length - 1];
            L.marker([lastPoint[0], lastPoint[1]], {
                icon: L.divIcon({
                    className: 'custom-div-icon',
                    html: '<div style="background-color: #EF4444; color: white; border-radius: 50%; width: 20px; height: 20px; display: flex; align-items: center; justify-content: center; font-weight: bold; font-size: 12px;">F</div>',
                    iconSize: [20, 20],
                    iconAnchor: [10, 10]
                })
            }).addTo(map).bindPopup('Finish');

            // Fit map to track bounds
            map.fitBounds(track.getBounds(), { padding: [20, 20] });

            // Create elevation profile
            const elevationData = trackPoints.map((point, index) => ({
                x: index,
                y: {{if .UseImperial}}point[2] * 3.28084{{else}}point[2]{{end}} // Convert to feet if imperial
            }));

            // Simple elevation chart using canvas
            const canvas = document.getElementById('elevationChart');
            const ctx = canvas.getContext('2d');
            const width = canvas.width;
            const height = canvas.height;

            if (elevationData.length > 1) {
                const minElevation = Math.min(...elevationData.map(d => d.y));
                const maxElevation = Math.max(...elevationData.map(d => d.y));
                const elevationRange = maxElevation - minElevation;

                // Clear canvas
                ctx.clearRect(0, 0, width, height);

                // Draw grid
                ctx.strokeStyle = '#E5E7EB';
                ctx.lineWidth = 1;
                for (let i = 0; i <= 10; i++) {
                    const y = (height - 40) * (i / 10) + 20;
                    ctx.beginPath();
                    ctx.moveTo(40, y);
                    ctx.lineTo(width - 20, y);
                    ctx.stroke();
                }

                // Draw elevation line
                ctx.strokeStyle = '#8B5CF6';
                ctx.lineWidth = 2;
                ctx.beginPath();

                elevationData.forEach((point, index) => {
                    const x = 40 + (width - 60) * (index / (elevationData.length - 1));
                    const y = height - 20 - ((point.y - minElevation) / elevationRange) * (height - 40);
                    
                    if (index === 0) {
                        ctx.moveTo(x, y);
                    } else {
                        ctx.lineTo(x, y);
                    }
                });
                ctx.stroke();

                // Draw labels
                ctx.fillStyle = '#374151';
                ctx.font = '12px Arial';
                ctx.fillText('{{if .UseImperial}}Elevation (ft){{else}}Elevation (m){{end}}', 10, 15);
                ctx.fillText(minElevation.toFixed(0), 45, height - 5);
                ctx.fillText(maxElevation.toFixed(0), 45, 25);
            } else {
                ctx.fillStyle = '#9CA3AF';
                ctx.font = '14px Arial';
                ctx.fillText('No elevation data available', width / 2 - 80, height / 2);
            }
        } else {
            // No track data
            map.setView([0, 0], 2);
            L.marker([0, 0]).addTo(map).bindPopup('No GPS data available');
        }
    </script>
</body>
</html>`

	// Enhance track data with statistics
	enhancedTrack := *track
	if len(track.Points) > 0 {
		enhancedTrack.StartLat = track.Points[0].Lat
		enhancedTrack.StartLon = track.Points[0].Lon
		enhancedTrack.EndLat = track.Points[len(track.Points)-1].Lat
		enhancedTrack.EndLon = track.Points[len(track.Points)-1].Lon
		enhancedTrack.TotalPoints = len(track.Points)
		
		// Calculate min/max elevation
		if len(track.Points) > 0 {
			minElev, maxElev := track.Points[0].Elevation, track.Points[0].Elevation
			for _, point := range track.Points {
				if point.Elevation < minElev {
					minElev = point.Elevation
				}
				if point.Elevation > maxElev {
					maxElev = point.Elevation
				}
			}
			enhancedTrack.MinElevation = minElev
			enhancedTrack.MaxElevation = maxElev
		}
	}

	funcMap := template.FuncMap{
		"metersToKm": func(meters float64) float64 {
			return meters / 1000
		},
		"metersToMiles": func(meters float64) float64 {
			return meters * 0.000621371
		},
		"metersToFeet": func(meters float64) float64 {
			return meters * 3.28084
		},
		"kmhToMph": func(kmh float64) float64 {
			return kmh * 0.621371
		},
		"formatDuration": func(seconds int) string {
			hours := seconds / 3600
			minutes := (seconds % 3600) / 60
			if hours > 0 {
				return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds%60)
			}
			return fmt.Sprintf("%d:%02d", minutes, seconds%60)
		},
	}

	data := struct {
		Activity    *models.Activity
		Track       *models.GPXTrack
		UseImperial bool
	}{
		Activity:    activity,
		Track:       &enhancedTrack,
		UseImperial: useImperial,
	}

	t, err := template.New("gps-track").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(w, data)
}

// Helper type for bulk upload results
type BulkUploadResult struct {
	FileName     string
	Index        int
	Total        int
	Status       string // "success" or "error"
	Error        string
	ActivityName string
	Distance     float64 // in km
	Duration     int     // in seconds
}

func (h *Handlers) RecalculateElevation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all activities with GPX files
	activities, err := h.storage.GetActivities()
	if err != nil {
		http.Error(w, "Error getting activities", http.StatusInternalServerError)
		return
	}

	recalculated := 0
	errors := 0

	for _, activity := range activities {
		if activity.GPXFile == "" {
			continue // Skip activities without GPX files
		}

		// Read the raw GPX file from uploads directory
		gpxPath := fmt.Sprintf("data/uploads/%s", activity.GPXFile)
		gpxData, err := ioutil.ReadFile(gpxPath)
		if err != nil {
			fmt.Printf("Warning: Could not read GPX file %s: %v\n", gpxPath, err)
			errors++
			continue
		}

		// Reparse the GPX with current algorithm
		track, newActivity, err := gpx.ParseGPX(string(gpxData))
		if err != nil {
			fmt.Printf("Warning: Could not parse GPX file %s: %v\n", gpxPath, err)
			errors++
			continue
		}

		// Update the activity with new elevation data but preserve original metadata
		activity.TotalElevation = newActivity.TotalElevation
		activity.Distance = newActivity.Distance
		activity.Duration = newActivity.Duration
		activity.AvgSpeed = newActivity.AvgSpeed
		activity.MaxSpeed = newActivity.MaxSpeed

		// Save updated activity
		if err := h.storage.SaveActivity(activity); err != nil {
			fmt.Printf("Warning: Could not save updated activity %s: %v\n", activity.ID, err)
			errors++
			continue
		}

		// Update the track data as well
		track.ID = activity.ID
		if err := h.storage.SaveGPXTrack(track); err != nil {
			fmt.Printf("Warning: Could not save updated track %s: %v\n", activity.ID, err)
			errors++
			continue
		}

		recalculated++
		fmt.Printf("INFO: Recalculated elevation for activity %s: %.2fm\n", activity.ID, activity.TotalElevation)
	}

	w.Header().Set("Content-Type", "text/html")
	if errors == 0 {
		w.Write([]byte(fmt.Sprintf(`<div class="p-3 bg-green-100 border border-green-400 text-green-700 rounded">✓ Recalculated elevation for %d activities</div>`, recalculated)))
	} else {
		w.Write([]byte(fmt.Sprintf(`<div class="p-3 bg-yellow-100 border border-yellow-400 text-yellow-700 rounded">⚠ Recalculated %d activities, %d errors</div>`, recalculated, errors)))
	}
}