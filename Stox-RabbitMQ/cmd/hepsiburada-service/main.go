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
	log.Println("🟠 Starting Hepsiburada Marketplace Service...")

	// Load configuration
	cfg := config.LoadConfig()
	cfg.ServiceName = "hepsiburada-service"

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
		{"hepsiburada_listings", "stox.listings", ""},                    // Fanout - receives all listings
		{"hepsiburada_orders", "stox.orders", "order.hepsiburada.*"},    // Topic - Hepsiburada orders
		{"hepsiburada_sync", "stox.sync", "hepsiburada_sync"},           // Direct - Hepsiburada sync
	}

	for _, q := range queues {
		err = client.DeclareQueue(q.name, q.exchange, q.routing)
		if err != nil {
			log.Fatalf("Failed to declare queue %s: %v", q.name, err)
		}
	}

	log.Println("✅ Hepsiburada Service initialized successfully")

	// Start consuming listings
	go func() {
		err := client.ConsumeMessages("hepsiburada_listings", handleHepsiburadaListing)
		if err != nil {
			log.Printf("Hepsiburada listings consumer error: %v", err)
		}
	}()

	// Start consuming orders
	go func() {
		err := client.ConsumeMessages("hepsiburada_orders", handleHepsiburadaOrder)
		if err != nil {
			log.Printf("Hepsiburada orders consumer error: %v", err)
		}
	}()

	// Start consuming sync operations
	go func() {
		err := client.ConsumeMessages("hepsiburada_sync", handleHepsiburadaSync)
		if err != nil {
			log.Printf("Hepsiburada sync consumer error: %v", err)
		}
	}()

	// Simulate periodic orders
	go simulateHepsiburadaOrders(client)

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("🟠 Hepsiburada Service shutting down...")
}

// handleHepsiburadaListing processes product listings for Hepsiburada
func handleHepsiburadaListing(data []byte) error {
	var product models.Product
	err := json.Unmarshal(data, &product)
	if err != nil {
		return fmt.Errorf("failed to unmarshal product: %w", err)
	}

	log.Printf("🟠 Hepsiburada: Processing listing for product %s", product.ID)

	// Mock Hepsiburada API integration
	time.Sleep(1800 * time.Millisecond) // Simulate API call

	// Convert price to Turkish Lira (mock exchange rate)
	priceInTL := product.Price * 27.5 // ~27.5 TL per USD

	// Create Hepsiburada listing
	listing := models.MarketplaceListing{
		ID:          fmt.Sprintf("hb_%s_%d", product.ID, time.Now().Unix()),
		ProductID:   product.ID,
		Marketplace: "hepsiburada",
		ListingID:   fmt.Sprintf("HB%d", time.Now().Unix()%10000000), // Mock Hepsiburada ID
		Status:      "active",
		Price:       priceInTL * 1.12, // 12% markup for Hepsiburada
		Stock:       200,               // Mock initial stock
		URL:         fmt.Sprintf("https://hepsiburada.com/product/hb%d", time.Now().Unix()%10000000),
		LastSyncAt:  time.Now(),
	}

	log.Printf("  ✅ Listed on Hepsiburada:")
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
		ID:        fmt.Sprintf("evt_hb_%d", time.Now().Unix()),
		Type:      "marketplace_listed",
		ProductID: product.ID,
		Data: map[string]interface{}{
			"marketplace": "hepsiburada",
			"listing_id":  listing.ListingID,
			"price":       listing.Price,
			"currency":    "TL",
			"url":         listing.URL,
		},
		Timestamp: time.Now(),
		Source:    "hepsiburada-service",
	}

	err = client.PublishMessage("stox.listings", "event.listed", event)
	if err != nil {
		log.Printf("Warning: Failed to publish listing event: %v", err)
	}

	return nil
}

// handleHepsiburadaOrder processes incoming Hepsiburada orders
func handleHepsiburadaOrder(data []byte) error {
	var order models.Order
	err := json.Unmarshal(data, &order)
	if err != nil {
		return fmt.Errorf("failed to unmarshal order: %w", err)
	}

	log.Printf("📦 Hepsiburada: Processing order %s", order.OrderID)

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

// handleHepsiburadaSync processes sync operations for Hepsiburada
func handleHepsiburadaSync(data []byte) error {
	var update models.InventoryUpdate
	err := json.Unmarshal(data, &update)
	if err != nil {
		return fmt.Errorf("failed to unmarshal sync update: %w", err)
	}

	if update.Marketplace != "hepsiburada" && update.Marketplace != "all" {
		return nil // Skip if not for Hepsiburada
	}

	log.Printf("🔄 Hepsiburada: Syncing %s for product %s", update.UpdateType, update.ProductID)

	// Mock Hepsiburada API sync
	time.Sleep(1000 * time.Millisecond)

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

// simulateHepsiburadaOrders creates demo orders for testing
func simulateHepsiburadaOrders(client *rabbitmq.Client) {
	time.Sleep(21 * time.Second) // Wait for listings to be processed

	orders := []models.Order{
		{
			ID:          "hb_order_001",
			Marketplace: "hepsiburada",
			OrderID:     "HB-456789123",
			ProductID:   "prod_001",
			UserID:      "user_789",
			Quantity:    1,
			Price:       6149.00, // Price in Turkish Lira
			Status:      "new",
			CustomerInfo: models.Customer{
				Name:  "Fatma Demir",
				Email: "fatma.demir@email.com",
				Phone: "+90-555-0456",
				Address: models.Address{
					Street:  "Halaskargazi Caddesi 456",
					City:    "Ankara",
					State:   "Ankara",
					Country: "Turkey",
					ZipCode: "06230",
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for i, order := range orders {
		time.Sleep(time.Duration(6+i*3) * time.Second)

		log.Printf("🎬 Demo: Simulating Hepsiburada order %s", order.OrderID)
		
		err := client.PublishMessage("stox.orders", "order.hepsiburada.tr", order)
		if err != nil {
			log.Printf("Failed to publish demo order: %v", err)
		}
	}
}
