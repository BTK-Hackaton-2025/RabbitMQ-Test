package config

import (
	"os"
	"strings"

	"ecommerce-rabbitmq/internal/types"
)

// LoadConfig loads configuration from environment variables
func LoadConfig() *types.Config {
	return &types.Config{
		AMQPURL:     getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
		ServiceName: getEnv("SERVICE_NAME", "ecommerce-service"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		Region:      getEnv("REGION", "US"),
		WorkerType:  getEnv("WORKER_TYPE", "processor"),
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return strings.TrimSpace(value)
	}
	return defaultValue
}
