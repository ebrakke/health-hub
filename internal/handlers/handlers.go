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
)

type Handlers struct {
	storage storage.Storage
}

func NewHandlers(s storage.Storage) *Handlers {
	return &Handlers{storage: s}
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Health Hub</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100 min-h-screen">
    <div class="container mx-auto px-4 py-8 max-w-4xl">
        <div class="text-center mb-8">
            <h1 class="text-4xl font-bold text-gray-900 mb-2">Health Hub</h1>
            <p class="text-gray-600">Upload and manage your health and fitness data</p>
        </div>
        
        <div class="bg-white rounded-lg shadow-md p-6 mb-8">
            <h2 class="text-2xl font-bold text-gray-900 mb-6">Upload Data</h2>
            
            <div class="grid md:grid-cols-2 gap-6">
                <div>
                    <h3 class="text-lg font-semibold text-gray-900 mb-3">GPX Files</h3>
                    <form hx-post="/api/upload/gpx" hx-encoding="multipart/form-data" 
                          hx-target="#gpx-status" hx-swap="innerHTML">
                        <input type="file" name="gpx" accept=".gpx" required 
                               class="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100 mb-3">
                        <button type="submit" class="w-full bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded transition duration-200">
                            Upload GPX
                        </button>
                    </form>
                    <div id="gpx-status" class="mt-2"></div>
                </div>
                
                <div>
                    <h3 class="text-lg font-semibold text-gray-900 mb-3">Health Data (JSON)</h3>
                    <form hx-post="/api/upload/health" hx-encoding="multipart/form-data" 
                          hx-target="#health-status" hx-swap="innerHTML">
                        <input type="file" name="health" accept=".json" required 
                               class="block w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-green-50 file:text-green-700 hover:file:bg-green-100 mb-3">
                        <button type="submit" class="w-full bg-green-500 hover:bg-green-700 text-white font-bold py-2 px-4 rounded transition duration-200">
                            Upload Health Data
                        </button>
                    </form>
                    <div id="health-status" class="mt-2"></div>
                </div>
            </div>
        </div>
        
        <div class="bg-white rounded-lg shadow-md p-6 mb-8">
            <h2 class="text-2xl font-bold text-gray-900 mb-6">Your Data</h2>
            <div class="grid md:grid-cols-2 gap-6">
                <div hx-get="/api/stats/activities" hx-trigger="load, every 30s" 
                     hx-target="this" hx-swap="innerHTML"
                     class="bg-gradient-to-r from-blue-50 to-blue-100 p-6 rounded-lg">
                    <h3 class="text-lg font-semibold text-gray-900 mb-2">Activities</h3>
                    <p class="text-gray-600">Loading...</p>
                </div>
                
                <div hx-get="/api/stats/health" hx-trigger="load, every 30s" 
                     hx-target="this" hx-swap="innerHTML"
                     class="bg-gradient-to-r from-green-50 to-green-100 p-6 rounded-lg">
                    <h3 class="text-lg font-semibold text-gray-900 mb-2">Health Metrics</h3>
                    <p class="text-gray-600">Loading...</p>
                </div>
            </div>
        </div>

        <div class="bg-white rounded-lg shadow-md p-6">
            <h2 class="text-2xl font-bold text-gray-900 mb-6">Explore Your Data</h2>
            <div class="grid md:grid-cols-2 gap-6">
                <a href="/activities" class="block bg-gradient-to-r from-blue-500 to-blue-600 hover:from-blue-600 hover:to-blue-700 text-white p-6 rounded-lg transition duration-200 transform hover:scale-105">
                    <h3 class="text-xl font-semibold mb-2">📊 Activity Log</h3>
                    <p class="text-blue-100">View all your activities with detailed information and GPX data</p>
                </a>
                
                <a href="/stats" class="block bg-gradient-to-r from-purple-500 to-purple-600 hover:from-purple-600 hover:to-purple-700 text-white p-6 rounded-lg transition duration-200 transform hover:scale-105">
                    <h3 class="text-xl font-semibold mb-2">📈 Stats & Trends</h3>
                    <p class="text-purple-100">Analyze your fitness progress with charts and weekly trends</p>
                </a>
            </div>
        </div>
    </div>
</body>
</html>`

	t, _ := template.New("home").Parse(tmpl)
	t.Execute(w, nil)
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
		http.Error(w, "Error parsing GPX file", http.StatusBadRequest)
		return
	}

	// Set additional activity details
	if activity.Name == "" {
		activity.Name = strings.TrimSuffix(header.Filename, ".gpx")
	}
	activity.GPXFile = filename

	if err := h.storage.SaveGPXTrack(track); err != nil {
		http.Error(w, "Error saving GPX track", http.StatusInternalServerError)
		return
	}

	if err := h.storage.SaveActivity(activity); err != nil {
		http.Error(w, "Error saving activity", http.StatusInternalServerError)
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

	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Activities - Health Hub</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-50">
    <div class="max-w-6xl mx-auto p-6">
        <div class="mb-6">
            <h1 class="text-3xl font-bold text-gray-900 mb-2">Activities</h1>
            <p class="text-gray-600">Your uploaded GPX activities and their statistics</p>
            <a href="/" class="inline-block mt-4 bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded">
                Back to Home
            </a>
        </div>

        <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {{range .Activities}}
            <div class="bg-white rounded-lg shadow-md p-6">
                <h3 class="text-lg font-semibold text-gray-900 mb-2">{{.Name}}</h3>
                <div class="space-y-2 text-sm text-gray-600">
                    <div class="flex justify-between">
                        <span>Type:</span>
                        <span class="font-medium">{{.Type}}</span>
                    </div>
                    <div class="flex justify-between">
                        <span>Distance:</span>
                        <span class="font-medium">{{printf "%.2f km" (div .Distance 1000)}}</span>
                    </div>
                    <div class="flex justify-between">
                        <span>Duration:</span>
                        <span class="font-medium">{{formatDuration .Duration}}</span>
                    </div>
                    <div class="flex justify-between">
                        <span>Avg Speed:</span>
                        <span class="font-medium">{{printf "%.1f km/h" .AvgSpeed}}</span>
                    </div>
                    <div class="flex justify-between">
                        <span>Max Speed:</span>
                        <span class="font-medium">{{printf "%.1f km/h" .MaxSpeed}}</span>
                    </div>
                    <div class="flex justify-between">
                        <span>Elevation:</span>
                        <span class="font-medium">{{printf "%.0f m" .TotalElevation}}</span>
                    </div>
                    <div class="flex justify-between">
                        <span>GPS Points:</span>
                        <span class="font-medium">{{.TotalPoints}}</span>
                    </div>
                    <div class="flex justify-between">
                        <span>Date:</span>
                        <span class="font-medium">{{.StartTime.Format "2006-01-02"}}</span>
                    </div>
                </div>
            </div>
            {{end}}
        </div>

        {{if eq (len .Activities) 0}}
        <div class="text-center py-12">
            <div class="text-gray-500">
                <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <h3 class="mt-2 text-sm font-medium text-gray-900">No activities</h3>
                <p class="mt-1 text-sm text-gray-500">Get started by uploading your first GPX file.</p>
                <div class="mt-6">
                    <a href="/" class="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700">
                        Upload GPX File
                    </a>
                </div>
            </div>
        </div>
        {{end}}
    </div>
</body>
</html>`

	funcMap := template.FuncMap{
		"div": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"formatDuration": func(seconds int) string {
			duration := time.Duration(seconds) * time.Second
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60
			if hours > 0 {
				return fmt.Sprintf("%dh %dm", hours, minutes)
			}
			return fmt.Sprintf("%dm", minutes)
		},
	}

	t, err := template.New("activities").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Activities []*models.Activity
	}{
		Activities: activities,
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
            <a href="/" class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded transition duration-200">
                Back to Home
            </a>
        </div>

        <!-- Overall Stats -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
            <div class="bg-white rounded-lg shadow-md p-6">
                <h3 class="text-lg font-semibold text-gray-900 mb-2">Total Distance</h3>
                <p class="text-3xl font-bold text-blue-600">{{printf "%.1f" (div .TotalDistance 1000)}} km</p>
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
	}

	funcMap := template.FuncMap{
		"div": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
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
}

func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}