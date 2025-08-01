package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"stox-rabbitmq/internal/config"
	"stox-rabbitmq/internal/models"
	"stox-rabbitmq/internal/rabbitmq"
)

func main() {
	log.Println("📝 Starting Stox SEO Content Generation Service...")

	// Load configuration
	cfg := config.LoadConfig()
	cfg.ServiceName = "seo-service"

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

	// Declare queue for SEO processing
	err = client.DeclareQueue("seo_processing", "stox.images", "image.enhanced")
	if err != nil {
		log.Fatalf("Failed to declare SEO queue: %v", err)
	}

	log.Println("✅ SEO Service initialized successfully")

	// Start consuming enhanced images for SEO generation
	go func() {
		err := client.ConsumeMessages("seo_processing", handleSEOGeneration)
		if err != nil {
			log.Printf("SEO service error: %v", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("📝 SEO Service shutting down...")
}

// handleSEOGeneration generates SEO-optimized content using mock RAG
func handleSEOGeneration(data []byte) error {
	var product models.Product
	err := json.Unmarshal(data, &product)
	if err != nil {
		return fmt.Errorf("failed to unmarshal product: %w", err)
	}

	log.Printf("🔍 Generating SEO content for product: %s", product.ID)

	// Mock RAG processing time
	log.Printf("  🧠 Analyzing product images and existing description...")
	log.Printf("  📚 Consulting RAG database for similar products...")
	log.Printf("  🎯 Optimizing for marketplace SEO algorithms...")
	
	time.Sleep(3 * time.Second) // Simulate AI processing

	// Mock SEO content generation based on product category and images
	seoData := generateSEOContent(product)
	product.SEO = seoData

	// Update product status
	product.Status = "seo_generated"
	product.UpdatedAt = time.Now()

	log.Printf("  ✅ Generated SEO title: %s", seoData.Title)
	log.Printf("  ✅ Generated description (%d chars)", len(seoData.Description))
	log.Printf("  ✅ Generated %d keywords", len(seoData.Keywords))
	log.Printf("  ✅ SEO Score: %.1f/10", seoData.Score)

	// Create SEO generation event
	event := models.ProcessingEvent{
		ID:        fmt.Sprintf("evt_seo_%d", time.Now().Unix()),
		Type:      "seo_generated",
		ProductID: product.ID,
		Data: map[string]interface{}{
			"seo_score":     seoData.Score,
			"title_length":  len(seoData.Title),
			"desc_length":   len(seoData.Description),
			"keyword_count": len(seoData.Keywords),
			"generated_by":  seoData.GeneratedBy,
		},
		Timestamp: time.Now(),
		Source:    "seo-service",
	}

	// Send to marketplace listing (Fanout exchange - broadcast to all marketplaces)
	client, _ := rabbitmq.NewClient(rabbitmq.Config{
		URL: "amqp://stox:stoxpass123@localhost:5672/",
	})
	defer client.Close()

	// Broadcast to all marketplaces using fanout exchange
	err = client.PublishMessage("stox.listings", "", product)
	if err != nil {
		return fmt.Errorf("failed to broadcast to marketplaces: %w", err)
	}

	// Publish SEO event
	err = client.PublishMessage("stox.images", "event.seo_generated", event)
	if err != nil {
		log.Printf("Warning: Failed to publish SEO event: %v", err)
	}

	log.Printf("✅ SEO content generated and broadcasted to all marketplaces")
	return nil
}

// generateSEOContent creates optimized content based on product data
func generateSEOContent(product models.Product) models.SEOData {
	// Mock advanced SEO generation with RAG
	category := strings.ToLower(product.Category)
	
	// Generate SEO-optimized title
	title := product.Title
	if category == "electronics" {
		title = fmt.Sprintf("%s - Premium Quality, Fast Shipping | Best Price Guaranteed", product.Title)
	} else if category == "wearables" {
		title = fmt.Sprintf("%s - Advanced Fitness Tracking | Free Shipping", product.Title)
	}

	// Generate SEO description
	description := fmt.Sprintf(
		"%s. %s. Free shipping, 30-day return policy, and 2-year warranty included. " +
		"Trusted by thousands of customers worldwide. Order now for fast delivery!",
		product.Title, product.Description)

	// Generate keywords based on category and product features
	keywords := []string{
		strings.ToLower(product.Title),
		category,
		"free shipping",
		"best price",
		"warranty",
		"premium quality",
	}

	if category == "electronics" {
		keywords = append(keywords, "wireless", "bluetooth", "high-quality", "noise cancellation")
	} else if category == "wearables" {
		keywords = append(keywords, "fitness", "health", "tracking", "smart", "heart rate")
	}

	// Generate meta tags
	metaTags := map[string]string{
		"og:title":       title,
		"og:description": description,
		"og:type":        "product",
		"product:price":  fmt.Sprintf("%.2f %s", product.Price, product.Currency),
		"product:category": product.Category,
	}

	// Calculate mock SEO score
	score := calculateSEOScore(title, description, keywords)

	return models.SEOData{
		Title:       title,
		Description: description,
		Keywords:    keywords,
		MetaTags:    metaTags,
		GeneratedBy: "ai",
		Score:       score,
	}
}

// calculateSEOScore calculates a mock SEO optimization score
func calculateSEOScore(title, description string, keywords []string) float64 {
	score := 5.0 // Base score

	// Title optimization
	if len(title) >= 50 && len(title) <= 60 {
		score += 1.0
	}

	// Description optimization
	if len(description) >= 150 && len(description) <= 160 {
		score += 1.0
	}

	// Keyword optimization
	if len(keywords) >= 5 {
		score += 1.0
	}

	// Content quality (mock analysis)
	if strings.Contains(strings.ToLower(description), "free shipping") {
		score += 0.5
	}
	if strings.Contains(strings.ToLower(description), "warranty") {
		score += 0.5
	}

	// Cap at 10.0
	if score > 10.0 {
		score = 10.0
	}

	return score
}
