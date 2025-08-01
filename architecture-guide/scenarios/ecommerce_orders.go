package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Order represents an e-commerce order
type Order struct {
	ID       string  `json:"id"`
	UserID   string  `json:"user_id"`
	Product  string  `json:"product"`
	Amount   float64 `json:"amount"`
	Region   string  `json:"region"`
	Priority string  `json:"priority"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	// Load environment variables
	err := godotenv.Load("./../../.env")
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	amqpURL := os.Getenv("AMQP_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(amqpURL)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// 1. Work Queue for order processing
	orderQueue, err := ch.QueueDeclare("order_processing", true, false, false, false, nil)
	failOnError(err, "Failed to declare order queue")

	// 2. Fanout exchange for order notifications
	err = ch.ExchangeDeclare("order_notifications", "fanout", true, false, false, false, nil)
	failOnError(err, "Failed to declare notification exchange")

	// 3. Direct exchange for regional fulfillment
	err = ch.ExchangeDeclare("regional_fulfillment", "direct", true, false, false, false, nil)
	failOnError(err, "Failed to declare regional exchange")

	fmt.Println("🛒 E-commerce Order System")
	fmt.Println("========================")
	fmt.Println("This demonstrates a real e-commerce architecture:")
	fmt.Println("📋 Work Queue: Distributes order processing among workers")
	fmt.Println("📡 Pub/Sub: Notifies all systems (inventory, email, analytics)")
	fmt.Println("🎯 Routing: Routes to regional fulfillment centers")
	fmt.Println()
	fmt.Println("Regions: US, EU, ASIA")
	fmt.Println("Priorities: standard, express")
	fmt.Println()

	for {
		var input string
		fmt.Print("Place order (user_id:product:amount:region:priority) or 'quit': ")
		fmt.Scanln(&input)

		if input == "quit" {
			break
		}

		// Parse input (simplified)
		parts := parseOrderInput(input)
		if len(parts) != 5 {
			fmt.Println("❌ Format: user123:laptop:999.99:US:express")
			continue
		}

		order := Order{
			ID:       fmt.Sprintf("order_%d", time.Now().Unix()),
			UserID:   parts[0],
			Product:  parts[1],
			Amount:   parseFloat(parts[2]),
			Region:   parts[3],
			Priority: parts[4],
		}

		orderJSON, _ := json.Marshal(order)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		// 1. Send to WORK QUEUE for processing
		err = ch.PublishWithContext(ctx, "", orderQueue.Name, false, false,
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "application/json",
				Body:         orderJSON,
			})
		failOnError(err, "Failed to publish to work queue")
		log.Printf("📋 [WORK QUEUE] Order sent for processing: %s", order.ID)

		// 2. Send to FANOUT for notifications (inventory, email, analytics)
		err = ch.PublishWithContext(ctx, "order_notifications", "", false, false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        orderJSON,
			})
		failOnError(err, "Failed to publish notifications")
		log.Printf("📡 [PUB/SUB] Order broadcasted to all services: %s", order.ID)

		// 3. Send to DIRECT exchange for regional routing
		err = ch.PublishWithContext(ctx, "regional_fulfillment", order.Region, false, false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        orderJSON,
			})
		failOnError(err, "Failed to publish to regional fulfillment")
		log.Printf("🎯 [ROUTING] Order routed to %s fulfillment center: %s", order.Region, order.ID)

		cancel()
		fmt.Printf("✅ Order %s processed through all channels!\n\n", order.ID)
	}
}

func parseOrderInput(input string) []string {
	// Simple string split - in real world, you'd use proper parsing
	result := []string{}
	current := ""
	for _, char := range input {
		if char == ':' {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func parseFloat(s string) float64 {
	// Simplified - use strconv.ParseFloat in real code
	if s == "" {
		return 0
	}
	// Just return a dummy value for demo
	return 99.99
}
