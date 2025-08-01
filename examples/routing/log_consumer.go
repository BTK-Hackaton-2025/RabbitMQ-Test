package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run log_consumer.go [log_levels...]\nExample: go run log_consumer.go error warning")
	}
	
	// Get log levels from command line
	levels := os.Args[1:]
	
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

	// Declare the same exchange
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

	// Declare exclusive queue
	q, err := ch.QueueDeclare(
		"",    // name (auto-generated)
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	// Bind queue to exchange for each log level
	for _, level := range levels {
		err = ch.QueueBind(
			q.Name,        // queue name
			level,         // routing key
			"logs_direct", // exchange
			false,
			nil,
		)
		failOnError(err, "Failed to bind a queue to an exchange")
		log.Printf("🔗 Bound to log level: %s", level)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan struct{})

	go func() {
		for d := range msgs {
			log.Printf("📨 [%s] %s", d.RoutingKey, d.Body)
		}
	}()

	log.Printf("📋 Waiting for logs with levels: %v. To exit press CTRL+C", levels)
	<-forever
}
