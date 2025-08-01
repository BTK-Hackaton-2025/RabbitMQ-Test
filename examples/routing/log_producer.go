package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
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

	// Declare a DIRECT exchange
	err = ch.ExchangeDeclare(
		"logs_direct", // name
		"direct",      // type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	fmt.Println("=== Log Router (Direct Routing Pattern) ===")
	fmt.Println("Real-world examples:")
	fmt.Println("🔴 ERROR logs → alerts team")
	fmt.Println("🟡 WARNING logs → monitoring system")
	fmt.Println("🟢 INFO logs → log aggregator")
	fmt.Println("🔵 DEBUG logs → development team")
	fmt.Println()
	fmt.Println("Available log levels: info, warning, error, debug")
	fmt.Println()

	for {
		var input string
		fmt.Print("Enter log (format: level:message) or 'quit': ")
		fmt.Scanln(&input)
		
		if input == "quit" {
			break
		}

		parts := strings.SplitN(input, ":", 2)
		if len(parts) != 2 {
			fmt.Println("❌ Invalid format! Use: level:message")
			continue
		}

		level := strings.TrimSpace(parts[0])
		message := strings.TrimSpace(parts[1])

		// Validate log level
		validLevels := []string{"info", "warning", "error", "debug"}
		valid := false
		for _, validLevel := range validLevels {
			if level == validLevel {
				valid = true
				break
			}
		}

		if !valid {
			fmt.Printf("❌ Invalid level '%s'. Use: %s\n", level, strings.Join(validLevels, ", "))
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		
		// Publish with routing key = log level
		err = ch.PublishWithContext(ctx,
			"logs_direct", // exchange
			level,         // routing key (this determines which queues receive it)
			false,         // mandatory
			false,         // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(message),
				Timestamp:   time.Now(),
			})
		
		cancel()
		failOnError(err, "Failed to publish log")
		
		log.Printf("📝 [x] Sent [%s] %s", level, message)
	}
}
