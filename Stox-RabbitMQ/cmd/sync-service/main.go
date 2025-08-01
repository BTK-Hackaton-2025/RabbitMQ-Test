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
	log.Println("🔄 Starting Stox Inventory Sync Service...")

	// Load configuration
	cfg := config.LoadConfig()
	cfg.ServiceName = "sync-service"

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

	// Declare queues for sync operations
	queues := []struct {
		name     string
		exchange string
		routing  string
	}{
		{"inventory_updates", "", ""},                              // Direct queue for inventory updates
		{"price_updates", "", ""},                                 // Direct queue for price updates
		{"listing_events", "stox.listings", "event.listed"},      // Topic - listing confirmations
	}

	for _, q := range queues {
		err = client.DeclareQueue(q.name, q.exchange, q.routing)
		if err != nil {
			log.Fatalf("Failed to declare queue %s: %v", q.name, err)
		}
	}

	log.Println("✅ Sync Service initialized successfully")

	// Start consuming listing events to track marketplace status
	go func() {
		err := client.ConsumeMessages("listing_events", handleListingEvent)
		if err != nil {
			log.Printf("Listing events consumer error: %v", err)
		}
	}()

	// Start consuming inventory updates
	go func() {
		err := client.ConsumeMessages("inventory_updates", handleInventoryUpdate)
		if err != nil {
			log.Printf("Inventory updates consumer error: %v", err)
		}
	}()

	// Start consuming price updates
	go func() {
		err := client.ConsumeMessages("price_updates", handlePriceUpdate)
		if err != nil {
			log.Printf("Price updates consumer error: %v", err)
		}
	}()

	// Start periodic sync operations
	go periodicSync(client)

	// Simulate inventory changes for demo
	go simulateInventoryChanges(client)

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("🔄 Sync Service shutting down...")
}

// handleListingEvent processes marketplace listing confirmations
func handleListingEvent(data []byte) error {
	var event models.ProcessingEvent
	err := json.Unmarshal(data, &event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	if event.Type != "marketplace_listed" {
		return nil // Only handle listing events
	}

	marketplace := event.Data["marketplace"].(string)
	listingID := event.Data["listing_id"].(string)

	log.Printf("📊 Tracking new listing: %s on %s (ID: %s)", event.ProductID, marketplace, listingID)

	// Store in mock database for sync tracking
	// In real implementation, this would update PostgreSQL
	return nil
}

// handleInventoryUpdate processes inventory synchronization requests
func handleInventoryUpdate(data []byte) error {
	var update models.InventoryUpdate
	err := json.Unmarshal(data, &update)
	if err != nil {
		return fmt.Errorf("failed to unmarshal inventory update: %w", err)
	}

	log.Printf("📦 Processing inventory update for product %s", update.ProductID)
	log.Printf("  Type: %s", update.UpdateType)
	if update.UpdateType == "stock" || update.UpdateType == "both" {
		log.Printf("  New Stock: %d", update.Stock)
	}
	if update.UpdateType == "price" || update.UpdateType == "both" {
		log.Printf("  New Price: $%.2f", update.Price)
	}

	// Send sync updates to all marketplaces using Direct routing
	client, _ := rabbitmq.NewClient(rabbitmq.Config{
		URL: "amqp://stox:stoxpass123@localhost:5672/",
	})
	defer client.Close()

	marketplaces := []string{"amazon", "trendyol", "hepsiburada"}
	
	for _, marketplace := range marketplaces {
		if update.Marketplace == "all" || update.Marketplace == marketplace {
			routingKey := fmt.Sprintf("%s_sync", marketplace)
			
			err := client.PublishMessage("stox.sync", routingKey, update)
			if err != nil {
				log.Printf("Failed to sync with %s: %v", marketplace, err)
				continue
			}
			
			log.Printf("  ✅ Synced with %s", marketplace)
		}
	}

	return nil
}

// handlePriceUpdate processes price synchronization requests
func handlePriceUpdate(data []byte) error {
	var update models.InventoryUpdate
	err := json.Unmarshal(data, &update)
	if err != nil {
		return fmt.Errorf("failed to unmarshal price update: %w", err)
	}

	log.Printf("💰 Processing price update for product %s: $%.2f", update.ProductID, update.Price)

	// Similar to inventory update but specifically for prices
	update.UpdateType = "price"
	
	// Delegate to inventory update handler for unified processing
	return handleInventoryUpdate(data)
}

// periodicSync performs regular synchronization checks
func periodicSync(client *rabbitmq.Client) {
	ticker := time.NewTicker(30 * time.Second) // Sync every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("🔄 Performing periodic sync check...")
			
			// Mock: Check for inventory discrepancies
			// In real implementation, this would query PostgreSQL and marketplace APIs
			
			products := []string{"prod_001", "prod_002"}
			
			for _, productID := range products {
				// Mock inventory drift detection
				if time.Now().Unix()%60 < 10 { // Random condition for demo
					log.Printf("  ⚠️  Detected inventory drift for product %s", productID)
					
					// Trigger sync
					update := models.InventoryUpdate{
						ProductID:   productID,
						Marketplace: "all",
						Stock:       int(time.Now().Unix() % 100) + 50, // Mock stock level
						UpdateType:  "stock",
						Timestamp:   time.Now(),
					}
					
					err := client.PublishMessage("", "inventory_updates", update)
					if err != nil {
						log.Printf("Failed to trigger sync for %s: %v", productID, err)
					}
				}
			}
		}
	}
}

// simulateInventoryChanges creates demo inventory/price changes
func simulateInventoryChanges(client *rabbitmq.Client) {
	time.Sleep(25 * time.Second) // Wait for all services to be ready

	changes := []models.InventoryUpdate{
		{
			ProductID:   "prod_001",
			Marketplace: "all",
			Stock:       75,
			UpdateType:  "stock",
			Timestamp:   time.Now(),
		},
		{
			ProductID:   "prod_002",
			Marketplace: "amazon",
			Price:       279.99,
			UpdateType:  "price",
			Timestamp:   time.Now(),
		},
		{
			ProductID:   "prod_001",
			Marketplace: "trendyol",
			Stock:       120,
			Price:       189.99,
			UpdateType:  "both",
			Timestamp:   time.Now(),
		},
	}

	for i, change := range changes {
		time.Sleep(time.Duration(15+i*8) * time.Second)

		log.Printf("🎬 Demo: Simulating %s update for product %s", change.UpdateType, change.ProductID)
		
		var queueName string
		if change.UpdateType == "price" {
			queueName = "price_updates"
		} else {
			queueName = "inventory_updates"
		}
		
		err := client.PublishMessage("", queueName, change)
		if err != nil {
			log.Printf("Failed to publish demo change: %v", err)
		}
	}
}
