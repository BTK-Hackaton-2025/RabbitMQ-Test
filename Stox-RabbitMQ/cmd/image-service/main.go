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
	log.Println("🖼️  Starting Stox Image Service...")

	// Load configuration
	cfg := config.LoadConfig()
	cfg.ServiceName = "image-service"

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
		{"image_uploads", "", ""},                           // Direct queue for uploads
		{"image_processing", "stox.images", "image.upload"}, // Topic routing
	}

	for _, q := range queues {
		err = client.DeclareQueue(q.name, q.exchange, q.routing)
		if err != nil {
			log.Fatalf("Failed to declare queue %s: %v", q.name, err)
		}
	}

	log.Println("✅ Image Service initialized successfully")

	// Start consuming image uploads
	go func() {
		err := client.ConsumeMessages("image_uploads", handleImageUpload)
		if err != nil {
			log.Printf("Error consuming image uploads: %v", err)
		}
	}()

	// Simulate periodic image uploads for demo
	go simulateImageUploads(client)

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("🖼️  Image Service shutting down...")
}

// handleImageUpload processes incoming image upload messages
func handleImageUpload(data []byte) error {
	var product models.Product
	err := json.Unmarshal(data, &product)
	if err != nil {
		return fmt.Errorf("failed to unmarshal product: %w", err)
	}

	log.Printf("📸 Processing image upload for product: %s", product.ID)

	// Mock image validation and S3 upload
	for i := range product.Images {
		product.Images[i].S3Key = fmt.Sprintf("products/%s/image_%d.jpg", product.ID, i)
		product.Images[i].IsProcessed = false
		log.Printf("  📁 Uploaded image to S3: %s", product.Images[i].S3Key)
	}

	// Update product status
	product.Status = "images_uploaded"
	product.UpdatedAt = time.Now()

	// Create processing event
	event := models.ProcessingEvent{
		ID:        fmt.Sprintf("evt_%d", time.Now().Unix()),
		Type:      "image_uploaded",
		ProductID: product.ID,
		Data: map[string]interface{}{
			"image_count": len(product.Images),
			"total_size":  calculateTotalSize(product.Images),
		},
		Timestamp: time.Now(),
		Source:    "image-service",
	}

	// Send to AI processing pipeline
	client, _ := rabbitmq.NewClient(rabbitmq.Config{
		URL: "amqp://stox:stoxpass123@localhost:5672/",
	})
	defer client.Close()

	// Route to AI service with topic routing
	err = client.PublishMessage("stox.images", "image.process", product)
	if err != nil {
		return fmt.Errorf("failed to send to AI processing: %w", err)
	}

	// Also publish processing event
	err = client.PublishMessage("stox.images", "event.image_uploaded", event)
	if err != nil {
		log.Printf("Warning: Failed to publish event: %v", err)
	}

	log.Printf("✅ Image upload processed and sent to AI enhancement")
	return nil
}

// simulateImageUploads creates demo image upload events
func simulateImageUploads(client *rabbitmq.Client) {
	time.Sleep(3 * time.Second) // Wait for services to start

	products := []models.Product{
		{
			ID:          "prod_001",
			UserID:      "user_123",
			Title:       "Wireless Bluetooth Headphones",
			Description: "High-quality wireless headphones with noise cancellation",
			Price:       199.99,
			Currency:    "USD",
			Category:    "Electronics",
			Status:      "image_uploaded",
			Images: []models.Image{
				{
					ID:          "img_001",
					OriginalURL: "https://example.com/headphones1.jpg",
					Size:        1024000,
					Width:       1920,
					Height:      1080,
					Format:      "JPEG",
				},
				{
					ID:          "img_002",
					OriginalURL: "https://example.com/headphones2.jpg",
					Size:        890000,
					Width:       1920,
					Height:      1080,
					Format:      "JPEG",
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "prod_002",
			UserID:      "user_456",
			Title:       "Smart Fitness Watch",
			Description: "Advanced fitness tracking with heart rate monitor",
			Price:       299.99,
			Currency:    "USD",
			Category:    "Wearables",
			Status:      "image_uploaded",
			Images: []models.Image{
				{
					ID:          "img_003",
					OriginalURL: "https://example.com/watch1.jpg",
					Size:        756000,
					Width:       1500,
					Height:      1500,
					Format:      "JPEG",
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for i, product := range products {
		time.Sleep(time.Duration(5+i*3) * time.Second) // Staggered uploads

		log.Printf("🎬 Demo: Simulating image upload for product %s", product.ID)
		
		err := client.PublishMessage("", "image_uploads", product)
		if err != nil {
			log.Printf("Failed to publish demo product: %v", err)
		}
	}
}

// calculateTotalSize calculates total size of all images
func calculateTotalSize(images []models.Image) int64 {
	var total int64
	for _, img := range images {
		total += img.Size
	}
	return total
}
