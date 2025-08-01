package config

import (
	"fmt"
	"os"
)

// Config holds all configuration for the Stox platform
type Config struct {
	RabbitMQ    RabbitMQConfig
	ServiceName string
	LogLevel    string
}

// RabbitMQConfig holds RabbitMQ connection details
type RabbitMQConfig struct {
	URL      string
	Username string
	Password string
	Host     string
	Port     string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		RabbitMQ: RabbitMQConfig{
			URL:      getEnv("RABBITMQ_URL", "amqp://stox:stoxpass123@localhost:5672/"),
			Username: getEnv("RABBITMQ_USERNAME", "stox"),
			Password: getEnv("RABBITMQ_PASSWORD", "stoxpass123"),
			Host:     getEnv("RABBITMQ_HOST", "localhost"),
			Port:     getEnv("RABBITMQ_PORT", "5672"),
		},
		ServiceName: getEnv("SERVICE_NAME", "stox-service"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

// GetRabbitMQURL constructs the full RabbitMQ URL
func (c *Config) GetRabbitMQURL() string {
	if c.RabbitMQ.URL != "" {
		return c.RabbitMQ.URL
	}
	return fmt.Sprintf("amqp://%s:%s@%s:%s/",
		c.RabbitMQ.Username,
		c.RabbitMQ.Password,
		c.RabbitMQ.Host,
		c.RabbitMQ.Port,
	)
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
