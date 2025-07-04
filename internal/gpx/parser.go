package gpx

import (
	"encoding/xml"
	"math"
	"time"

	"health-hub/internal/models"
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

					// Calculate elevation gain
					if point.Elevation > prevPoint.Elevation {
						totalElevation += point.Elevation - prevPoint.Elevation
					}

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