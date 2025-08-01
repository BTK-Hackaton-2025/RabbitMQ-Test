package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ecommerce-rabbitmq/internal/config"
	"ecommerce-rabbitmq/internal/rabbitmq"
	"ecommerce-rabbitmq/internal/types"
)

func main() {
	cfg := config.LoadConfig()
	
	log.Printf("🔄 E-commerce Worker starting...")
	log.Printf("Worker Type: %s", cfg.WorkerType)
	log.Printf("Service: %s", cfg.ServiceName)

	client, err := rabbitmq.NewClient(cfg.AMQPURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer client.Close()

	err = client.SetupExchangesAndQueues()
	if err != nil {
		log.Fatalf("Failed to setup exchanges and queues: %v", err)
	}

	// Create handler based on worker type
	handler := createHandler(cfg.WorkerType)
	
	err = client.ConsumeOrders(cfg.WorkerType, handler)
	if err != nil {
		log.Fatalf("Failed to start consuming: %v", err)
	}

	log.Printf("🎯 [%s] Worker ready. Waiting for orders...", cfg.WorkerType)

	// Wait for interrupt signal to gracefully shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	log.Println("🔄 Shutting down worker...")
}

func createHandler(workerType string) func(*types.Order) error {
	return func(order *types.Order) error {
		switch workerType {
		case "processor":
			log.Printf("🔄 Processing order %s (Product: %s, Amount: $%.2f)", 
				order.ID, order.Product, order.Amount)
			// Simulate processing time
			time.Sleep(2 * time.Second)
			log.Printf("✅ Order %s processed successfully", order.ID)

		case "inventory":
			log.Printf("📦 INVENTORY: Reserving stock for %s (Product: %s)", 
				order.ID, order.Product)

		case "email":
			log.Printf("📧 EMAIL: Sending confirmation to user %s for order %s", 
				order.UserID, order.ID)

		case "analytics":
			log.Printf("📊 ANALYTICS: Recording sale - Product: %s, Amount: $%.2f, Region: %s", 
				order.Product, order.Amount, order.Region)

		default:
			// Fulfillment centers
			if len(workerType) > 12 && workerType[:12] == "fulfillment_" {
				region := workerType[12:]
				log.Printf("🏭 FULFILLMENT [%s]: Preparing shipment for order %s", 
					region, order.ID)
			} else {
				log.Printf("❓ Unknown worker type: %s, processing order %s", workerType, order.ID)
			}
		}

		return nil
	}
}
