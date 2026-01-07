package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port          string
	AWSRegion     string
	UploadTimeout time.Duration
	Env           string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		AWSRegion:     getEnv("AWS_REGION", "us-east-1"),
		UploadTimeout: time.Duration(getEnvAsInt("UPLOAD_TIMEOUT_SECONDS", 30)) * time.Second,
		Env:           getEnv("APP_ENV", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
