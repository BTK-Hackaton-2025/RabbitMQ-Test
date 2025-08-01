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
		log.Fatal("Usage: go run subscriber.go [subscriber_name]")
	}
	
	subscriberName := os.Args[1]
	
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
		"news_broadcast", // name
		"fanout",         // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	// Declare exclusive queue (unique per subscriber)
	q, err := ch.QueueDeclare(
		"",    // name (empty = auto-generated)
		false, // durable
		false, // delete when unused
		true,  // exclusive (only this connection)
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	// Bind queue to exchange
	err = ch.QueueBind(
		q.Name,           // queue name
		"",               // routing key (ignored for fanout)
		"news_broadcast", // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue to an exchange")

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
			log.Printf("🔔 [%s] Breaking News: %s", subscriberName, d.Body)
		}
	}()

	log.Printf("📺 [%s] Waiting for news broadcasts. To exit press CTRL+C", subscriberName)
	<-forever
}
