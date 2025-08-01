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
	log.Println("🤖 Starting Stox AI Enhancement Service...")

	// Load configuration
	cfg := config.LoadConfig()
	cfg.ServiceName = "ai-service"

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

	// Declare queues for AI processing
	queues := []struct {
		name     string
		exchange string
		routing  string
	}{
		{"ai_processing", "stox.images", "image.process"},     // Receive from image service
		{"ai_enhancement", "", ""},                           // Work queue for AI workers
	}

	for _, q := range queues {
		err = client.DeclareQueue(q.name, q.exchange, q.routing)
		if err != nil {
			log.Fatalf("Failed to declare queue %s: %v", q.name, err)
		}
	}

	log.Println("✅ AI Service initialized successfully")

	// Start consuming images for processing (multiple workers)
	for i := 0; i < 3; i++ { // 3 AI workers
		go func(workerID int) {
			log.Printf("🔧 Starting AI worker #%d", workerID)
			err := client.ConsumeMessages("ai_processing", func(data []byte) error {
				return handleAIProcessing(data, workerID)
			})
			if err != nil {
				log.Printf("AI worker #%d error: %v", workerID, err)
			}
		}(i + 1)
	}

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("🤖 AI Service shutting down...")
}

// handleAIProcessing processes images with mock AI enhancement
func handleAIProcessing(data []byte, workerID int) error {
	var product models.Product
	err := json.Unmarshal(data, &product)
	if err != nil {
		return fmt.Errorf("failed to unmarshal product: %w", err)
	}

	log.Printf("🎨 AI Worker #%d: Enhancing images for product: %s", workerID, product.ID)

	// Mock AI processing time (simulating actual AI work)
	processingTime := time.Duration(2+len(product.Images)) * time.Second
	log.Printf("  ⏳ Processing %d images (estimated %v)...", len(product.Images), processingTime)
	
	time.Sleep(processingTime)

	// Mock AI enhancement results
	for i := range product.Images {
		product.Images[i].IsProcessed = true
		product.Images[i].ProcessingAt = time.Now()
		product.Images[i].EnhancedURL = fmt.Sprintf("https://cdn.stox.com/enhanced/%s/image_%d_enhanced.jpg", 
			product.ID, i)
		
		log.Printf("  ✨ Enhanced image %d: Background removed, quality improved", i+1)
	}

	// Update product status
	product.Status = "ai_enhanced"
	product.UpdatedAt = time.Now()

	// Create AI processing event
	event := models.ProcessingEvent{
		ID:        fmt.Sprintf("evt_ai_%d", time.Now().Unix()),
		Type:      "ai_enhanced",
		ProductID: product.ID,
		Data: map[string]interface{}{
			"worker_id":        workerID,
			"processing_time":  processingTime.Seconds(),
			"images_enhanced":  len(product.Images),
			"enhancements": []string{
				"background_removal",
				"color_enhancement", 
				"noise_reduction",
				"resolution_upscale",
			},
		},
		Timestamp: time.Now(),
		Source:    "ai-service",
	}

	// Send to SEO service for content generation
	client, _ := rabbitmq.NewClient(rabbitmq.Config{
		URL: "amqp://stox:stoxpass123@localhost:5672/",
	})
	defer client.Close()

	// Route to SEO service
	err = client.PublishMessage("stox.images", "image.enhanced", product)
	if err != nil {
		return fmt.Errorf("failed to send to SEO service: %w", err)
	}

	// Publish AI enhancement event
	err = client.PublishMessage("stox.images", "event.ai_enhanced", event)
	if err != nil {
		log.Printf("Warning: Failed to publish AI event: %v", err)
	}

	log.Printf("✅ AI Worker #%d: Product %s enhanced and sent to SEO generation", workerID, product.ID)
	return nil
}
