package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        string
	DataPath    string
	UseS3       bool
	S3Bucket    string
	AWSRegion   string
	Environment string
	
	// Elevation smoothing parameters
	ElevationSmoothingWindow    int     // Number of points to consider for smoothing
	ElevationMinGain           float64  // Minimum elevation gain to count (meters)
	ElevationSmoothingEnabled  bool     // Enable elevation smoothing
}

func Load() *Config {
	return &Config{
		Port:        getEnvOrDefault("PORT", "8088"),
		DataPath:    getEnvOrDefault("DATA_PATH", "./data"),
		UseS3:       getBoolEnvOrDefault("USE_S3", false),
		S3Bucket:    getEnvOrDefault("S3_BUCKET", ""),
		AWSRegion:   getEnvOrDefault("AWS_REGION", "us-east-1"),
		Environment: getEnvOrDefault("ENVIRONMENT", "development"),
		
		// Elevation smoothing defaults (Strava-inspired threshold approach)
		ElevationSmoothingWindow:   getIntEnvOrDefault("ELEVATION_SMOOTHING_WINDOW", 5),
		ElevationMinGain:          getFloatEnvOrDefault("ELEVATION_MIN_GAIN", 1.0),
		ElevationSmoothingEnabled: getBoolEnvOrDefault("ELEVATION_SMOOTHING_ENABLED", true),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnvOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getIntEnvOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getFloatEnvOrDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}