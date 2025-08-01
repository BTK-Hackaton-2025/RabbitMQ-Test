package types

import "time"

// Order represents an e-commerce order
type Order struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	Product  string    `json:"product"`
	Amount   float64   `json:"amount"`
	Region   string    `json:"region"`
	Priority string    `json:"priority"`
	Created  time.Time `json:"created"`
}

// Config holds application configuration
type Config struct {
	AMQPURL      string
	ServiceName  string
	LogLevel     string
	Region       string
	WorkerType   string
}
