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
}

func Load() *Config {
	return &Config{
		Port:        getEnvOrDefault("PORT", "8088"),
		DataPath:    getEnvOrDefault("DATA_PATH", "./data"),
		UseS3:       getBoolEnvOrDefault("USE_S3", false),
		S3Bucket:    getEnvOrDefault("S3_BUCKET", ""),
		AWSRegion:   getEnvOrDefault("AWS_REGION", "us-east-1"),
		Environment: getEnvOrDefault("ENVIRONMENT", "development"),
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