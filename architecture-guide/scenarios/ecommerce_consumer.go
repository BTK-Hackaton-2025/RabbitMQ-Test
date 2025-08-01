package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

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
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run ecommerce_consumer.go [worker_type]\nTypes: processor, inventory, email, analytics, fulfillment_US, fulfillment_EU")
	}

	workerType := os.Args[1]

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

	var msgs <-chan amqp.Delivery

	switch workerType {
	case "processor":
		// Work queue consumer - competes with other processors
		q, err := ch.QueueDeclare("order_processing", true, false, false, false, nil)
		failOnError(err, "Failed to declare queue")

		ch.Qos(1, 0, false) // Fair dispatch

		msgs, err = ch.Consume(q.Name, "", false, false, false, false, nil)
		failOnError(err, "Failed to register consumer")

		log.Printf("📋 [ORDER PROCESSOR] Worker started. Competing with other processors...")

	case "inventory", "email", "analytics":
		// Pub/Sub consumers - all get the same messages
		err = ch.ExchangeDeclare("order_notifications", "fanout", true, false, false, false, nil)
		failOnError(err, "Failed to declare exchange")

		q, err := ch.QueueDeclare("", false, false, true, false, nil) // Exclusive queue
		failOnError(err, "Failed to declare queue")

		err = ch.QueueBind(q.Name, "", "order_notifications", false, nil)
		failOnError(err, "Failed to bind queue")

		msgs, err = ch.Consume(q.Name, "", true, false, false, false, nil)
		failOnError(err, "Failed to register consumer")

		log.Printf("📡 [%s SERVICE] Listening for order notifications...", workerType)

	case "fulfillment_US", "fulfillment_EU", "fulfillment_ASIA":
		// Direct routing consumers - only get messages for their region
		region := workerType[12:] // Extract region from fulfillment_XX

		err = ch.ExchangeDeclare("regional_fulfillment", "direct", true, false, false, false, nil)
		failOnError(err, "Failed to declare exchange")

		q, err := ch.QueueDeclare("fulfillment_"+region, false, false, false, false, nil)
		failOnError(err, "Failed to declare queue")

		err = ch.QueueBind(q.Name, region, "regional_fulfillment", false, nil)
		failOnError(err, "Failed to bind queue")

		msgs, err = ch.Consume(q.Name, "", true, false, false, false, nil)
		failOnError(err, "Failed to register consumer")

		log.Printf("🎯 [%s FULFILLMENT] Listening for %s orders...", region, region)

	default:
		log.Fatalf("Unknown worker type: %s", workerType)
	}

	forever := make(chan struct{})

	go func() {
		for d := range msgs {
			var order Order
			json.Unmarshal(d.Body, &order)

			switch workerType {
			case "processor":
				log.Printf("🔄 Processing order %s (Product: %s, Amount: $%.2f)", 
					order.ID, order.Product, order.Amount)
				// Simulate processing time
				// time.Sleep(2 * time.Second)
				log.Printf("✅ Order %s processed successfully", order.ID)
				d.Ack(false) // Manual ack for work queue

			case "inventory":
				log.Printf("📦 INVENTORY: Reserving stock for %s (Product: %s)", 
					order.ID, order.Product)

			case "email":
				log.Printf("📧 EMAIL: Sending confirmation to user %s for order %s", 
					order.UserID, order.ID)

			case "analytics":
				log.Printf("📊 ANALYTICS: Recording sale - Product: %s, Amount: $%.2f, Region: %s", 
					order.Product, order.Amount, order.Region)

			default: // fulfillment centers
				region := workerType[12:]
				log.Printf("🏭 FULFILLMENT [%s]: Preparing shipment for order %s", 
					region, order.ID)
			}
		}
	}()

	log.Printf("🎯 [%s] Ready. To exit press CTRL+C", workerType)
	<-forever
}
