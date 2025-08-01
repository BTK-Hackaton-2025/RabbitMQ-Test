package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"stox-rabbitmq/internal/config"
	"stox-rabbitmq/internal/models"
	"stox-rabbitmq/internal/rabbitmq"
)

func main() {
	log.Println("🏪 Starting Amazon Marketplace Service...")

	// Load configuration
	cfg := config.LoadConfig()
	cfg.ServiceName = "amazon-service"

	// Create RabbitMQ client
	client, err := rabbitmq.NewClient(rabbitmq.Config{
		URL: cfg.GetRabbitMQURL(),
	})
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ client: %v", err)
	}
	defer client.Close()

	// Setup exchanges
	err = client.SetupExchanges()
	if err != nil {
		log.Fatalf("Failed to setup exchanges: %v", err)
	}

	// Declare queues
	queues := []struct {
		name     string
		exchange string
		routing  string
	}{
		{"amazon_listings", "stox.listings", ""},                    // Fanout - receives all listings
		{"amazon_orders", "stox.orders", "order.amazon.*"},         // Topic - Amazon orders
		{"amazon_sync", "stox.sync", "amazon_sync"},                // Direct - Amazon sync operations
	}

	for _, q := range queues {
		err = client.DeclareQueue(q.name, q.exchange, q.routing)
		if err != nil {
			log.Fatalf("Failed to declare queue %s: %v", q.name, err)
		}
	}

	log.Println("✅ Amazon Service initialized successfully")

	// Start consuming listings
	go func() {
		err := client.ConsumeMessages("amazon_listings", handleAmazonListing)
		if err != nil {
			log.Printf("Amazon listings consumer error: %v", err)
		}
	}()

	// Start consuming orders
	go func() {
		err := client.ConsumeMessages("amazon_orders", handleAmazonOrder)
		if err != nil {
			log.Printf("Amazon orders consumer error: %v", err)
		}
	}()

	// Start consuming sync operations
	go func() {
		err := client.ConsumeMessages("amazon_sync", handleAmazonSync)
		if err != nil {
			log.Printf("Amazon sync consumer error: %v", err)
		}
	}()

	// Simulate periodic orders
	go simulateAmazonOrders(client)

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("🏪 Amazon Service shutting down...")
}

// handleAmazonListing processes product listings for Amazon
func handleAmazonListing(data []byte) error {
	var product models.Product
	err := json.Unmarshal(data, &product)
	if err != nil {
		return fmt.Errorf("failed to unmarshal product: %w", err)
	}

	log.Printf("🛒 Amazon: Processing listing for product %s", product.ID)

	// Mock Amazon API integration
	time.Sleep(2 * time.Second) // Simulate API call

	// Create Amazon listing
	listing := models.MarketplaceListing{
		ID:          fmt.Sprintf("amz_%s_%d", product.ID, time.Now().Unix()),
		ProductID:   product.ID,
		Marketplace: "amazon",
		ListingID:   fmt.Sprintf("B0%d", time.Now().Unix()%1000000), // Mock ASIN
		Status:      "active",
		Price:       product.Price * 1.1, // 10% markup for Amazon
		Stock:       100,                  // Mock initial stock
		URL:         fmt.Sprintf("https://amazon.com/dp/B0%d", time.Now().Unix()%1000000),
		LastSyncAt:  time.Now(),
	}

	log.Printf("  ✅ Listed on Amazon:")
	log.Printf("    ASIN: %s", listing.ListingID)
	log.Printf("    Price: $%.2f", listing.Price)
	log.Printf("    URL: %s", listing.URL)

	// Send listing confirmation
	client, _ := rabbitmq.NewClient(rabbitmq.Config{
		URL: "amqp://stox:stoxpass123@localhost:5672/",
	})
	defer client.Close()

	// Publish listing event
	event := models.ProcessingEvent{
		ID:        fmt.Sprintf("evt_amz_%d", time.Now().Unix()),
		Type:      "marketplace_listed",
		ProductID: product.ID,
		Data: map[string]interface{}{
			"marketplace": "amazon",
			"listing_id":  listing.ListingID,
			"price":       listing.Price,
			"url":         listing.URL,
		},
		Timestamp: time.Now(),
		Source:    "amazon-service",
	}

	err = client.PublishMessage("stox.listings", "event.listed", event)
	if err != nil {
		log.Printf("Warning: Failed to publish listing event: %v", err)
	}

	return nil
}

// handleAmazonOrder processes incoming Amazon orders
func handleAmazonOrder(data []byte) error {
	var order models.Order
	err := json.Unmarshal(data, &order)
	if err != nil {
		return fmt.Errorf("failed to unmarshal order: %w", err)
	}

	log.Printf("📦 Amazon: Processing order %s", order.OrderID)

	// Mock order processing
	order.Status = "processing"
	order.UpdatedAt = time.Now()

	log.Printf("  ✅ Order processed:")
	log.Printf("    Product: %s", order.ProductID)
	log.Printf("    Quantity: %d", order.Quantity)
	log.Printf("    Customer: %s", order.CustomerInfo.Name)

	return nil
}

// handleAmazonSync processes sync operations for Amazon
func handleAmazonSync(data []byte) error {
	var update models.InventoryUpdate
	err := json.Unmarshal(data, &update)
	if err != nil {
		return fmt.Errorf("failed to unmarshal sync update: %w", err)
	}

	if update.Marketplace != "amazon" && update.Marketplace != "all" {
		return nil // Skip if not for Amazon
	}

	log.Printf("🔄 Amazon: Syncing %s for product %s", update.UpdateType, update.ProductID)

	// Mock Amazon API sync
	time.Sleep(1 * time.Second)

	if update.UpdateType == "stock" || update.UpdateType == "both" {
		log.Printf("  📊 Updated stock to: %d", update.Stock)
	}
	if update.UpdateType == "price" || update.UpdateType == "both" {
		log.Printf("  💰 Updated price to: $%.2f", update.Price)
	}

	return nil
}

// simulateAmazonOrders creates demo orders for testing
func simulateAmazonOrders(client *rabbitmq.Client) {
	time.Sleep(15 * time.Second) // Wait for listings to be processed

	orders := []models.Order{
		{
			ID:          "amz_order_001",
			Marketplace: "amazon",
			OrderID:     "AMZ-123456789",
			ProductID:   "prod_001",
			UserID:      "user_123",
			Quantity:    1,
			Price:       219.99, // Amazon price with markup
			Status:      "new",
			CustomerInfo: models.Customer{
				Name:  "John Smith",
				Email: "john.smith@email.com",
				Phone: "+1-555-0123",
				Address: models.Address{
					Street:  "123 Main St",
					City:    "Seattle",
					State:   "WA",
					Country: "USA",
					ZipCode: "98101",
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for i, order := range orders {
		time.Sleep(time.Duration(10+i*5) * time.Second)

		log.Printf("🎬 Demo: Simulating Amazon order %s", order.OrderID)
		
		err := client.PublishMessage("stox.orders", "order.amazon.us", order)
		if err != nil {
			log.Printf("Failed to publish demo order: %v", err)
		}
	}
}
