package models

import "time"

type Activity struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"` // "running", "cycling", "walking", etc.
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Duration      int       `json:"duration"`       // seconds
	Distance      float64   `json:"distance"`       // meters
	Calories      int       `json:"calories"`
	GPXFile       string    `json:"gpx_file,omitempty"`
	TotalElevation float64  `json:"total_elevation"` // meters
	MaxSpeed      float64   `json:"max_speed"`       // km/h
	AvgSpeed      float64   `json:"avg_speed"`       // km/h
	TotalPoints   int       `json:"total_points"`    // number of GPS points
	CreatedAt     time.Time `json:"created_at"`
}

type GPXTrack struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Points    []GPXPoint  `json:"points"`
	CreatedAt time.Time   `json:"created_at"`
}

type GPXPoint struct {
	Lat       float64   `json:"lat"`
	Lon       float64   `json:"lon"`
	Elevation float64   `json:"elevation,omitempty"`
	Time      time.Time `json:"time,omitempty"`
}