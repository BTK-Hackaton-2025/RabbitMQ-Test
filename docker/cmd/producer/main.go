package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"ecommerce-rabbitmq/internal/config"
	"ecommerce-rabbitmq/internal/rabbitmq"
	"ecommerce-rabbitmq/internal/types"
)

func main() {
	cfg := config.LoadConfig()
	
	log.Printf("🛒 E-commerce Order Producer starting...")
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

	fmt.Println("🛒 E-commerce Order System (Dockerized)")
	fmt.Println("======================================")
	fmt.Println("This demonstrates containerized RabbitMQ architecture:")
	fmt.Println("📋 Work Queue: Distributes order processing among workers")
	fmt.Println("📡 Pub/Sub: Notifies all systems (inventory, email, analytics)")
	fmt.Println("🎯 Routing: Routes to regional fulfillment centers")
	fmt.Println()
	fmt.Println("Format: user_id:product:amount:region:priority")
	fmt.Println("Example: user123:laptop:999.99:US:express")
	fmt.Println("Regions: US, EU, ASIA")
	fmt.Println("Priorities: standard, express")
	fmt.Println()

	for {
		var input string
		fmt.Print("Place order or 'quit': ")
		fmt.Scanln(&input)

		if input == "quit" {
			break
		}

		order, err := parseOrder(input)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			fmt.Println("Format: user_id:product:amount:region:priority")
			continue
		}

		err = client.PublishOrder(order)
		if err != nil {
			log.Printf("Failed to publish order: %v", err)
			continue
		}

		fmt.Printf("✅ Order %s processed through all channels!\n\n", order.ID)
	}
}

func parseOrder(input string) (*types.Order, error) {
	parts := strings.Split(input, ":")
	if len(parts) != 5 {
		return nil, fmt.Errorf("invalid format, expected 5 parts separated by ':'")
	}

	amount, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %v", err)
	}

	// Validate region
	validRegions := []string{"US", "EU", "ASIA"}
	region := strings.ToUpper(parts[3])
	valid := false
	for _, validRegion := range validRegions {
		if region == validRegion {
			valid = true
			break
		}
	}
	if !valid {
		return nil, fmt.Errorf("invalid region '%s', use: %s", region, strings.Join(validRegions, ", "))
	}

	return &types.Order{
		ID:       fmt.Sprintf("order_%d", time.Now().Unix()),
		UserID:   parts[0],
		Product:  parts[1],
		Amount:   amount,
		Region:   region,
		Priority: parts[4],
		Created:  time.Now(),
	}, nil
}
