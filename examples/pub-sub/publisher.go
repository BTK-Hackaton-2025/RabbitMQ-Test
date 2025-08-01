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
	err := godotenv.Load("../../.env")
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

	// Declare a FANOUT exchange
	err = ch.ExchangeDeclare(
		"news_broadcast", // name
		"fanout",         // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	fmt.Println("=== News Broadcaster (Pub/Sub Pattern) ===")
	fmt.Println("Real-world examples:")
	fmt.Println("📰 News updates to all subscribers")
	fmt.Println("📧 Email notifications to all users")
	fmt.Println("🔔 Push notifications to all mobile apps")
	fmt.Println("📊 Real-time dashboard updates")
	fmt.Println()

	for {
		var news string
		fmt.Print("📢 Enter breaking news (or 'quit' to exit): ")
		fmt.Scanln(&news)
		
		if news == "quit" {
			break
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		
		// Publish to exchange (not directly to queue)
		err = ch.PublishWithContext(ctx,
			"news_broadcast", // exchange
			"",               // routing key (ignored for fanout)
			false,            // mandatory
			false,            // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(news),
				Timestamp:   time.Now(),
			})
		
		cancel()
		failOnError(err, "Failed to publish news")
		
		log.Printf("📰 [x] Broadcasted: %s", news)
	}
}
