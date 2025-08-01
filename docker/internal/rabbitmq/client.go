package rabbitmq

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"ecommerce-rabbitmq/internal/types"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewClient(amqpURL string) (*Client, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{
		conn: conn,
		ch:   ch,
	}, nil
}

func (c *Client) Close() error {
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) SetupExchangesAndQueues() error {
	// Work queue for order processing
	_, err := c.ch.QueueDeclare("order_processing", true, false, false, false, nil)
	if err != nil {
		return err
	}

	// Fanout exchange for notifications
	err = c.ch.ExchangeDeclare("order_notifications", "fanout", true, false, false, false, nil)
	if err != nil {
		return err
	}

	// Direct exchange for regional fulfillment
	err = c.ch.ExchangeDeclare("regional_fulfillment", "direct", true, false, false, false, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) PublishOrder(order *types.Order) error {
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Send to work queue
	err = c.ch.PublishWithContext(ctx, "", "order_processing", false, false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         orderJSON,
		})
	if err != nil {
		return err
	}
	log.Printf("📋 [WORK QUEUE] Order sent: %s", order.ID)

	// 2. Send to fanout for notifications
	err = c.ch.PublishWithContext(ctx, "order_notifications", "", false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        orderJSON,
		})
	if err != nil {
		return err
	}
	log.Printf("📡 [PUB/SUB] Order broadcasted: %s", order.ID)

	// 3. Send to direct exchange for regional routing
	err = c.ch.PublishWithContext(ctx, "regional_fulfillment", order.Region, false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        orderJSON,
		})
	if err != nil {
		return err
	}
	log.Printf("🎯 [ROUTING] Order routed to %s: %s", order.Region, order.ID)

	return nil
}

func (c *Client) ConsumeOrders(workerType string, handler func(*types.Order) error) error {
	var msgs <-chan amqp.Delivery
	var err error

	switch workerType {
	case "processor":
		c.ch.Qos(1, 0, false) // Fair dispatch
		msgs, err = c.ch.Consume("order_processing", "", false, false, false, false, nil)
		if err != nil {
			return err
		}

	case "inventory", "email", "analytics":
		q, err := c.ch.QueueDeclare("", false, false, true, false, nil)
		if err != nil {
			return err
		}
		err = c.ch.QueueBind(q.Name, "", "order_notifications", false, nil)
		if err != nil {
			return err
		}
		msgs, err = c.ch.Consume(q.Name, "", true, false, false, false, nil)
		if err != nil {
			return err
		}

	default:
		if len(workerType) > 12 && workerType[:12] == "fulfillment_" {
			region := workerType[12:]
			q, err := c.ch.QueueDeclare("fulfillment_"+region, false, false, false, false, nil)
			if err != nil {
				return err
			}
			err = c.ch.QueueBind(q.Name, region, "regional_fulfillment", false, nil)
			if err != nil {
				return err
			}
			msgs, err = c.ch.Consume(q.Name, "", true, false, false, false, nil)
			if err != nil {
				return err
			}
		}
	}

	go func() {
		for d := range msgs {
			var order types.Order
			if err := json.Unmarshal(d.Body, &order); err != nil {
				log.Printf("Error unmarshaling order: %v", err)
				continue
			}

			if err := handler(&order); err != nil {
				log.Printf("Error handling order: %v", err)
				if workerType == "processor" {
					d.Nack(false, true) // Requeue on error
				}
				continue
			}

			if workerType == "processor" {
				d.Ack(false)
			}
		}
	}()

	return nil
}
