package gpx

import (
	"encoding/xml"
	"math"
	"time"

	"health-hub/internal/models"
	"health-hub/internal/config"
)

// GPX XML structure
type GPX struct {
	XMLName xml.Name `xml:"gpx"`
	Tracks  []Track  `xml:"trk"`
}

type Track struct {
	Name     string    `xml:"name"`
	Segments []Segment `xml:"trkseg"`
}

type Segment struct {
	Points []TrackPoint `xml:"trkpt"`
}

type TrackPoint struct {
	Lat       float64 `xml:"lat,attr"`
	Lon       float64 `xml:"lon,attr"`
	Elevation float64 `xml:"ele,omitempty"`
	Time      string  `xml:"time,omitempty"`
}

func ParseGPX(content string) (*models.GPXTrack, *models.Activity, error) {
	cfg := config.Load()
	return ParseGPXWithConfig(content, cfg)
}

func ParseGPXWithConfig(content string, cfg *config.Config) (*models.GPXTrack, *models.Activity, error) {
	var gpx GPX
	err := xml.Unmarshal([]byte(content), &gpx)
	if err != nil {
		return nil, nil, err
	}

	track := &models.GPXTrack{
		Points: []models.GPXPoint{},
	}

	activity := &models.Activity{
		Type: "activity",
	}

	var totalDistance float64
	var totalElevation float64
	var maxSpeed float64
	var speeds []float64
	var startTime, endTime time.Time
	var prevPoint *models.GPXPoint

	for _, trk := range gpx.Tracks {
		if track.Name == "" {
			track.Name = trk.Name
			activity.Name = trk.Name
		}

		for _, seg := range trk.Segments {
			for _, pt := range seg.Points {
				point := models.GPXPoint{
					Lat:       pt.Lat,
					Lon:       pt.Lon,
					Elevation: pt.Elevation,
				}

				// Parse time
				if pt.Time != "" {
					if t, err := time.Parse(time.RFC3339, pt.Time); err == nil {
						point.Time = t
						if startTime.IsZero() || t.Before(startTime) {
							startTime = t
						}
						if endTime.IsZero() || t.After(endTime) {
							endTime = t
						}
					}
				}

				// Calculate distance and speed if we have a previous point
				if prevPoint != nil {
					dist := haversineDistance(prevPoint.Lat, prevPoint.Lon, point.Lat, point.Lon)
					totalDistance += dist

					// Note: Elevation calculation moved to after all points are collected

					// Calculate speed if we have time data
					if !prevPoint.Time.IsZero() && !point.Time.IsZero() {
						timeDiff := point.Time.Sub(prevPoint.Time).Seconds()
						if timeDiff > 0 {
							speed := (dist / timeDiff) * 3.6 // Convert m/s to km/h
							speeds = append(speeds, speed)
							if speed > maxSpeed {
								maxSpeed = speed
							}
						}
					}
				}

				track.Points = append(track.Points, point)
				prevPoint = &point
			}
		}
	}

	// Calculate elevation gain using smoothing algorithm
	totalElevation = calculateSmoothedElevation(track.Points, cfg)

	// Calculate average speed
	var avgSpeed float64
	if len(speeds) > 0 {
		var totalSpeed float64
		for _, speed := range speeds {
			totalSpeed += speed
		}
		avgSpeed = totalSpeed / float64(len(speeds))
	}

	// Set activity stats
	activity.StartTime = startTime
	activity.EndTime = endTime
	if !startTime.IsZero() && !endTime.IsZero() {
		activity.Duration = int(endTime.Sub(startTime).Seconds())
	}
	activity.Distance = totalDistance
	activity.TotalElevation = totalElevation
	activity.MaxSpeed = maxSpeed
	activity.AvgSpeed = avgSpeed
	activity.TotalPoints = len(track.Points)

	return track, activity, nil
}

// haversineDistance calculates the distance between two points on Earth using the Haversine formula
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // Earth's radius in meters

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// calculateSmoothedElevation calculates elevation gain using a moving average smoothing algorithm
// to filter out GPS noise, inspired by proven GPX smoothing techniques
func calculateSmoothedElevation(points []models.GPXPoint, cfg *config.Config) float64 {
	if !cfg.ElevationSmoothingEnabled || len(points) < 2 {
		return calculateSimpleElevation(points)
	}

	windowSize := cfg.ElevationSmoothingWindow
	minGain := cfg.ElevationMinGain
	
	// First, smooth the elevation data using moving average
	smoothedElevations := make([]float64, len(points))
	
	for i := 0; i < len(points); i++ {
		// Calculate symmetric window around point i
		start := i - windowSize/2
		end := i + windowSize/2
		
		// Clamp to array bounds
		if start < 0 {
			start = 0
		}
		if end >= len(points) {
			end = len(points) - 1
		}
		
		// Calculate moving average for this window
		sum := 0.0
		count := 0
		for j := start; j <= end; j++ {
			sum += points[j].Elevation
			count++
		}
		
		smoothedElevations[i] = sum / float64(count)
	}
	
	// Now calculate elevation gain from smoothed data
	totalElevation := 0.0
	
	for i := 1; i < len(smoothedElevations); i++ {
		elevationDiff := smoothedElevations[i] - smoothedElevations[i-1]
		
		// Only count elevation gains above minimum threshold
		if elevationDiff > minGain {
			totalElevation += elevationDiff
		}
	}
	
	return totalElevation
}

// calculateSimpleElevation is the fallback method for simple elevation calculation
func calculateSimpleElevation(points []models.GPXPoint) float64 {
	totalElevation := 0.0
	for i := 1; i < len(points); i++ {
		if points[i].Elevation > points[i-1].Elevation {
			totalElevation += points[i].Elevation - points[i-1].Elevation
		}
	}
	return totalElevation
}

// calculateMedian calculates the median of a slice of float64 values
func calculateMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	if len(values) == 1 {
		return values[0]
	}
	
	// Simple bubble sort for small arrays (typically 5-15 elements)
	sorted := make([]float64, len(values))
	copy(sorted, values)
	
	for i := 0; i < len(sorted); i++ {
		for j := 0; j < len(sorted)-1-i; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}
	
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}