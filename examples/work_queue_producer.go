package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	// Load environment variables
	err := godotenv.Load("../.env")
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

	// Declare a DURABLE queue (survives broker restart)
	q, err := ch.QueueDeclare(
		"task_queue", // name
		true,         // durable (important!)
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare a queue")

	fmt.Println("=== Work Queue Producer ===")
	fmt.Println("This simulates a web server sending tasks to workers")
	fmt.Println("Real-world use: Image processing, email sending, data processing")
	fmt.Println()

	for {
		var task string
		fmt.Print("Enter a task (or 'quit' to exit): ")
		fmt.Scanln(&task)
		
		if task == "quit" {
			break
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		
		// Publish with persistence
		err = ch.PublishWithContext(ctx,
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				DeliveryMode: amqp.Persistent, // Make message persistent
				ContentType:  "text/plain",
				Body:         []byte(task),
				Timestamp:    time.Now(),
			})
		
		cancel()
		failOnError(err, "Failed to publish a task")
		
		log.Printf(" [x] Sent task: %s", task)
	}
}
