package main

import (
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
	log.Println("🎬 Starting Stox Demo - E-Commerce Automation Pipeline")
	log.Println("===============================================")

	// Load configuration
	cfg := config.LoadConfig()
	cfg.ServiceName = "demo-service"

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

	log.Println("✅ Demo service initialized")
	log.Println()
	log.Println("🚀 This demo will show the complete Stox platform workflow:")
	log.Println("   1. Image Upload → AI Enhancement")
	log.Println("   2. SEO Content Generation")  
	log.Println("   3. Multi-Marketplace Broadcasting")
	log.Println("   4. Order Processing & Inventory Sync")
	log.Println()
	log.Println("📊 Monitor RabbitMQ Management UI: http://localhost:15672")
	log.Println("   Username: stox")
	log.Println("   Password: stoxpass123")
	log.Println()

	// Wait for other services to start
	log.Println("⏳ Waiting for all services to start...")
	time.Sleep(5 * time.Second)

	// Start the demo pipeline
	go runDemo(client)

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("🎬 Demo service shutting down...")
}

// runDemo orchestrates the entire demo workflow
func runDemo(client *rabbitmq.Client) {
	// Demo products to process
	products := []models.Product{
		{
			ID:          "demo_prod_001",
			UserID:      "demo_user_123",
			Title:       "Premium Wireless Earbuds",
			Description: "High-fidelity wireless earbuds with active noise cancellation",
			Price:       149.99,
			Currency:    "USD",
			Category:    "Electronics",
			Status:      "uploaded",
			Images: []models.Image{
				{
					ID:          "demo_img_001",
					OriginalURL: "https://example.com/earbuds1.jpg",
					Size:        1200000,
					Width:       2000,
					Height:      2000,
					Format:      "JPEG",
				},
				{
					ID:          "demo_img_002", 
					OriginalURL: "https://example.com/earbuds2.jpg",
					Size:        980000,
					Width:       1800,
					Height:      1800,
					Format:      "JPEG",
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "demo_prod_002",
			UserID:      "demo_user_456",
			Title:       "Smart Home Security Camera",
			Description: "4K wireless security camera with night vision and motion detection",
			Price:       89.99,
			Currency:    "USD",
			Category:    "Electronics",
			Status:      "uploaded",
			Images: []models.Image{
				{
					ID:          "demo_img_003",
					OriginalURL: "https://example.com/camera1.jpg",
					Size:        1500000,
					Width:       2400,
					Height:      1600,
					Format:      "JPEG",
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	log.Println("🎬 DEMO PHASE 1: Image Upload & Processing")
	log.Println("==========================================")

	for i, product := range products {
		time.Sleep(time.Duration(3+i*2) * time.Second)

		log.Printf("📸 Uploading product: %s", product.Title)
		log.Printf("   Product ID: %s", product.ID)
		log.Printf("   Images: %d", len(product.Images))
		log.Printf("   Category: %s", product.Category)
		
		// Send to image processing pipeline
		err := client.PublishMessage("", "image_uploads", product)
		if err != nil {
			log.Printf("❌ Failed to upload product %s: %v", product.ID, err)
			continue
		}
		
		log.Printf("✅ Product %s sent to image processing pipeline", product.ID)
		log.Println()
	}

	// Wait for pipeline to process
	log.Println("⏳ Processing pipeline in progress...")
	log.Println("   🔄 AI Enhancement in progress...")
	log.Println("   📝 SEO Content Generation...")
	log.Println("   🏪 Multi-Marketplace Broadcasting...")
	log.Println()

	time.Sleep(12 * time.Second)

	log.Println("🎬 DEMO PHASE 2: Inventory & Price Management")
	log.Println("============================================")

	// Simulate inventory changes
	inventoryUpdates := []models.InventoryUpdate{
		{
			ProductID:   "demo_prod_001",
			Marketplace: "all",
			Stock:       50,
			UpdateType:  "stock",
			Timestamp:   time.Now(),
		},
		{
			ProductID:   "demo_prod_002",
			Marketplace: "amazon",
			Price:       79.99,
			UpdateType:  "price",
			Timestamp:   time.Now(),
		},
	}

	for i, update := range inventoryUpdates {
		time.Sleep(time.Duration(5+i*3) * time.Second)

		log.Printf("🔄 Inventory Update: %s", update.UpdateType)
		log.Printf("   Product: %s", update.ProductID)
		log.Printf("   Marketplace: %s", update.Marketplace)
		if update.UpdateType == "stock" || update.UpdateType == "both" {
			log.Printf("   New Stock: %d", update.Stock)
		}
		if update.UpdateType == "price" || update.UpdateType == "both" {
			log.Printf("   New Price: $%.2f", update.Price)
		}

		err := client.PublishMessage("", "inventory_updates", update)
		if err != nil {
			log.Printf("❌ Failed to send inventory update: %v", err)
			continue
		}

		log.Printf("✅ Inventory update sent to sync service")
		log.Println()
	}

	time.Sleep(8 * time.Second)

	log.Println("🎬 DEMO PHASE 3: Order Processing Simulation")
	log.Println("===========================================")

	// Simulate incoming orders
	orders := []models.Order{
		{
			ID:          "demo_order_001",
			Marketplace: "amazon",
			OrderID:     "AMZ-DEMO-123",
			ProductID:   "demo_prod_001",
			UserID:      "demo_user_123",
			Quantity:    2,
			Price:       164.99,
			Status:      "new",
			CustomerInfo: models.Customer{
				Name:  "Alice Johnson",
				Email: "alice.johnson@email.com",
				Phone: "+1-555-0199",
				Address: models.Address{
					Street:  "456 Oak Street",
					City:    "San Francisco",
					State:   "CA",
					Country: "USA",
					ZipCode: "94102",
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "demo_order_002",
			Marketplace: "trendyol",
			OrderID:     "TDY-DEMO-456",
			ProductID:   "demo_prod_002",
			UserID:      "demo_user_456",
			Quantity:    1,
			Price:       2469.75, // Price in Turkish Lira
			Status:      "new",
			CustomerInfo: models.Customer{
				Name:  "Mehmet Özkan",
				Email: "mehmet.ozkan@email.com",
				Phone: "+90-555-0234",
				Address: models.Address{
					Street:  "Atatürk Bulvarı 789",
					City:    "İzmir",
					State:   "İzmir",
					Country: "Turkey",
					ZipCode: "35220",
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for i, order := range orders {
		time.Sleep(time.Duration(4+i*3) * time.Second)

		log.Printf("📦 New Order Received:")
		log.Printf("   Order ID: %s", order.OrderID)
		log.Printf("   Marketplace: %s", order.Marketplace)
		log.Printf("   Product: %s", order.ProductID)
		log.Printf("   Quantity: %d", order.Quantity)
		log.Printf("   Customer: %s", order.CustomerInfo.Name)

		// Route order to appropriate marketplace service
		routingKey := fmt.Sprintf("order.%s.%s", order.Marketplace, getRegionCode(order.CustomerInfo.Address.Country))
		
		err := client.PublishMessage("stox.orders", routingKey, order)
		if err != nil {
			log.Printf("❌ Failed to route order: %v", err)
			continue
		}

		log.Printf("✅ Order routed to %s service", order.Marketplace)
		log.Println()
	}

	log.Println("🎉 DEMO COMPLETE!")
	log.Println("=================")
	log.Println("✅ All phases completed successfully:")
	log.Println("   📸 Image processing pipeline")
	log.Println("   🤖 AI enhancement workflow")
	log.Println("   📝 SEO content generation")
	log.Println("   🏪 Multi-marketplace listing")
	log.Println("   🔄 Inventory synchronization")
	log.Println("   📦 Order processing")
	log.Println()
	log.Println("📊 Check RabbitMQ Management UI to see message flows")
	log.Println("🔍 Monitor service logs to see processing details")
}

// getRegionCode returns region code based on country
func getRegionCode(country string) string {
	switch country {
	case "USA", "Canada":
		return "us"
	case "Turkey":
		return "tr"
	case "Germany", "France", "Italy", "Spain":
		return "eu"
	default:
		return "intl"
	}
}
