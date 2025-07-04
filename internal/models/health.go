package models

import "time"

type HealthMetric struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // "heart_rate", "sleep", "steps", "weight", etc.
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"` // "oura", "fitbit", "manual", etc.
	CreatedAt time.Time `json:"created_at"`
}

type SleepData struct {
	ID             string    `json:"id"`
	Date           time.Time `json:"date"`
	Bedtime        time.Time `json:"bedtime"`
	WakeTime       time.Time `json:"wake_time"`
	TotalSleep     int       `json:"total_sleep"`     // minutes
	DeepSleep      int       `json:"deep_sleep"`      // minutes
	REMSleep       int       `json:"rem_sleep"`       // minutes
	LightSleep     int       `json:"light_sleep"`     // minutes
	SleepScore     int       `json:"sleep_score"`     // 0-100
	Source         string    `json:"source"`
	CreatedAt      time.Time `json:"created_at"`
}

type HeartRateData struct {
	ID           string    `json:"id"`
	RestingHR    int       `json:"resting_hr"`
	MaxHR        int       `json:"max_hr,omitempty"`
	HRVariability float64  `json:"hr_variability,omitempty"`
	Date         time.Time `json:"date"`
	Source       string    `json:"source"`
	CreatedAt    time.Time `json:"created_at"`
}