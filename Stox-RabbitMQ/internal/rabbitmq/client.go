package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

// Client wraps RabbitMQ connection and provides high-level operations
type Client struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	config  Config
}

type Config struct {
	URL      string
	Exchange string
	Queue    string
}

// NewClient creates a new RabbitMQ client
func NewClient(config Config) (*Client, error) {
	conn, err := amqp091.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	client := &Client{
		conn:    conn,
		channel: ch,
		config:  config,
	}

	return client, nil
}

// SetupExchanges declares all the exchanges for the Stox platform
func (c *Client) SetupExchanges() error {
	exchanges := []struct {
		name string
		kind string
	}{
		{"stox.images", "topic"},
		{"stox.listings", "fanout"},
		{"stox.sync", "direct"},
		{"stox.orders", "topic"},
	}

	for _, exchange := range exchanges {
		err := c.channel.ExchangeDeclare(
			exchange.name, // name
			exchange.kind, // type
			true,         // durable
			false,        // auto-deleted
			false,        // internal
			false,        // no-wait
			nil,          // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", exchange.name, err)
		}
	}

	log.Println("✅ All exchanges declared successfully")
	return nil
}

// DeclareQueue declares a queue and binds it to an exchange
func (c *Client) DeclareQueue(queueName, exchangeName, routingKey string) error {
	_, err := c.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
	}

	if exchangeName != "" {
		err = c.channel.QueueBind(
			queueName,    // queue name
			routingKey,   // routing key
			exchangeName, // exchange
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s to exchange %s: %w", queueName, exchangeName, err)
		}
	}

	return nil
}

// PublishMessage publishes a message to an exchange
func (c *Client) PublishMessage(exchange, routingKey string, message interface{}) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = c.channel.Publish(
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp091.Persistent, // persistent
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// ConsumeMessages consumes messages from a queue
func (c *Client) ConsumeMessages(queueName string, handler func([]byte) error) error {
	msgs, err := c.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack (we'll handle manually)
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("📨 Received message from queue %s", queueName)
			
			err := handler(d.Body)
			if err != nil {
				log.Printf("❌ Error processing message: %v", err)
				d.Nack(false, false) // Negative acknowledgment, don't requeue
			} else {
				log.Printf("✅ Message processed successfully")
				d.Ack(false) // Acknowledge message
			}
		}
	}()

	log.Printf("🎧 Waiting for messages from queue: %s. To exit press CTRL+C", queueName)
	<-forever

	return nil
}

// Close closes the RabbitMQ connection
func (c *Client) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// HealthCheck checks if the connection is alive
func (c *Client) HealthCheck() error {
	if c.conn == nil || c.conn.IsClosed() {
		return fmt.Errorf("RabbitMQ connection is closed")
	}
	return nil
}
