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
	log.Println("🛍️ Starting Trendyol Marketplace Service...")

	// Load configuration
	cfg := config.LoadConfig()
	cfg.ServiceName = "trendyol-service"

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
		{"trendyol_listings", "stox.listings", ""},                    // Fanout - receives all listings
		{"trendyol_orders", "stox.orders", "order.trendyol.*"},       // Topic - Trendyol orders
		{"trendyol_sync", "stox.sync", "trendyol_sync"},              // Direct - Trendyol sync
	}

	for _, q := range queues {
		err = client.DeclareQueue(q.name, q.exchange, q.routing)
		if err != nil {
			log.Fatalf("Failed to declare queue %s: %v", q.name, err)
		}
	}

	log.Println("✅ Trendyol Service initialized successfully")

	// Start consuming listings
	go func() {
		err := client.ConsumeMessages("trendyol_listings", handleTrendyolListing)
		if err != nil {
			log.Printf("Trendyol listings consumer error: %v", err)
		}
	}()

	// Start consuming orders
	go func() {
		err := client.ConsumeMessages("trendyol_orders", handleTrendyolOrder)
		if err != nil {
			log.Printf("Trendyol orders consumer error: %v", err)
		}
	}()

	// Start consuming sync operations
	go func() {
		err := client.ConsumeMessages("trendyol_sync", handleTrendyolSync)
		if err != nil {
			log.Printf("Trendyol sync consumer error: %v", err)
		}
	}()

	// Simulate periodic orders
	go simulateTrendyolOrders(client)

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("🛍️ Trendyol Service shutting down...")
}

// handleTrendyolListing processes product listings for Trendyol
func handleTrendyolListing(data []byte) error {
	var product models.Product
	err := json.Unmarshal(data, &product)
	if err != nil {
		return fmt.Errorf("failed to unmarshal product: %w", err)
	}

	log.Printf("🇹🇷 Trendyol: Processing listing for product %s", product.ID)

	// Mock Trendyol API integration
	time.Sleep(1500 * time.Millisecond) // Simulate API call

	// Convert price to Turkish Lira (mock exchange rate)
	priceInTL := product.Price * 27.5 // ~27.5 TL per USD

	// Create Trendyol listing
	listing := models.MarketplaceListing{
		ID:          fmt.Sprintf("tdy_%s_%d", product.ID, time.Now().Unix()),
		ProductID:   product.ID,
		Marketplace: "trendyol",
		ListingID:   fmt.Sprintf("TY%d", time.Now().Unix()%10000000), // Mock Trendyol ID
		Status:      "active",
		Price:       priceInTL * 1.08, // 8% markup for Trendyol
		Stock:       150,               // Mock initial stock
		URL:         fmt.Sprintf("https://trendyol.com/product/ty%d", time.Now().Unix()%10000000),
		LastSyncAt:  time.Now(),
	}

	log.Printf("  ✅ Listed on Trendyol:")
	log.Printf("    Product ID: %s", listing.ListingID)
	log.Printf("    Price: ₺%.2f", listing.Price)
	log.Printf("    URL: %s", listing.URL)

	// Send listing confirmation
	client, _ := rabbitmq.NewClient(rabbitmq.Config{
		URL: "amqp://stox:stoxpass123@localhost:5672/",
	})
	defer client.Close()

	// Publish listing event
	event := models.ProcessingEvent{
		ID:        fmt.Sprintf("evt_tdy_%d", time.Now().Unix()),
		Type:      "marketplace_listed",
		ProductID: product.ID,
		Data: map[string]interface{}{
			"marketplace": "trendyol",
			"listing_id":  listing.ListingID,
			"price":       listing.Price,
			"currency":    "TL",
			"url":         listing.URL,
		},
		Timestamp: time.Now(),
		Source:    "trendyol-service",
	}

	err = client.PublishMessage("stox.listings", "event.listed", event)
	if err != nil {
		log.Printf("Warning: Failed to publish listing event: %v", err)
	}

	return nil
}

// handleTrendyolOrder processes incoming Trendyol orders
func handleTrendyolOrder(data []byte) error {
	var order models.Order
	err := json.Unmarshal(data, &order)
	if err != nil {
		return fmt.Errorf("failed to unmarshal order: %w", err)
	}

	log.Printf("📦 Trendyol: Processing order %s", order.OrderID)

	// Mock order processing
	order.Status = "processing"
	order.UpdatedAt = time.Now()

	log.Printf("  ✅ Order processed:")
	log.Printf("    Product: %s", order.ProductID)
	log.Printf("    Quantity: %d", order.Quantity)
	log.Printf("    Customer: %s", order.CustomerInfo.Name)
	log.Printf("    Price: ₺%.2f", order.Price)

	return nil
}

// handleTrendyolSync processes sync operations for Trendyol
func handleTrendyolSync(data []byte) error {
	var update models.InventoryUpdate
	err := json.Unmarshal(data, &update)
	if err != nil {
		return fmt.Errorf("failed to unmarshal sync update: %w", err)
	}

	if update.Marketplace != "trendyol" && update.Marketplace != "all" {
		return nil // Skip if not for Trendyol
	}

	log.Printf("🔄 Trendyol: Syncing %s for product %s", update.UpdateType, update.ProductID)

	// Mock Trendyol API sync
	time.Sleep(800 * time.Millisecond)

	if update.UpdateType == "stock" || update.UpdateType == "both" {
		log.Printf("  📊 Updated stock to: %d", update.Stock)
	}
	if update.UpdateType == "price" || update.UpdateType == "both" {
		// Convert to Turkish Lira
		priceInTL := update.Price * 27.5
		log.Printf("  💰 Updated price to: ₺%.2f", priceInTL)
	}

	return nil
}

// simulateTrendyolOrders creates demo orders for testing
func simulateTrendyolOrders(client *rabbitmq.Client) {
	time.Sleep(18 * time.Second) // Wait for listings to be processed

	orders := []models.Order{
		{
			ID:          "tdy_order_001",
			Marketplace: "trendyol",
			OrderID:     "TDY-987654321",
			ProductID:   "prod_002",
			UserID:      "user_456",
			Quantity:    2,
			Price:       8799.00, // Price in Turkish Lira
			Status:      "new",
			CustomerInfo: models.Customer{
				Name:  "Ahmet Yılmaz",
				Email: "ahmet.yilmaz@email.com",
				Phone: "+90-555-0123",
				Address: models.Address{
					Street:  "Bağdat Caddesi 123",
					City:    "İstanbul",
					State:   "İstanbul",
					Country: "Turkey",
					ZipCode: "34740",
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for i, order := range orders {
		time.Sleep(time.Duration(8+i*4) * time.Second)

		log.Printf("🎬 Demo: Simulating Trendyol order %s", order.OrderID)
		
		err := client.PublishMessage("stox.orders", "order.trendyol.tr", order)
		if err != nil {
			log.Printf("Failed to publish demo order: %v", err)
		}
	}
}
