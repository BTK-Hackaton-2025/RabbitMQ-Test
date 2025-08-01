package models

import "time"

// Product represents a product being processed through the platform
type Product struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Currency    string    `json:"currency"`
	Category    string    `json:"category"`
	Images      []Image   `json:"images"`
	SEO         SEOData   `json:"seo"`
	Status      string    `json:"status"` // processing, enhanced, listed, error
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Image represents an image in the system
type Image struct {
	ID           string `json:"id"`
	OriginalURL  string `json:"original_url"`
	EnhancedURL  string `json:"enhanced_url,omitempty"`
	S3Key        string `json:"s3_key"`
	Size         int64  `json:"size"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Format       string `json:"format"`
	IsProcessed  bool   `json:"is_processed"`
	ProcessingAt time.Time `json:"processing_at,omitempty"`
}

// SEOData contains SEO-optimized content
type SEOData struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
	MetaTags    map[string]string `json:"meta_tags"`
	GeneratedBy string   `json:"generated_by"` // ai, manual
	Score       float64  `json:"score"`        // SEO optimization score
}

// MarketplaceListing represents a product listing on a marketplace
type MarketplaceListing struct {
	ID           string    `json:"id"`
	ProductID    string    `json:"product_id"`
	Marketplace  string    `json:"marketplace"` // amazon, trendyol, hepsiburada
	ListingID    string    `json:"listing_id"`  // External marketplace ID
	Status       string    `json:"status"`      // pending, active, rejected, paused
	Price        float64   `json:"price"`
	Stock        int       `json:"stock"`
	URL          string    `json:"url"`
	LastSyncAt   time.Time `json:"last_sync_at"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// Order represents an order from any marketplace
type Order struct {
	ID           string     `json:"id"`
	Marketplace  string     `json:"marketplace"`
	OrderID      string     `json:"order_id"` // External order ID
	ProductID    string     `json:"product_id"`
	UserID       string     `json:"user_id"`
	Quantity     int        `json:"quantity"`
	Price        float64    `json:"price"`
	Status       string     `json:"status"` // new, processing, shipped, delivered, cancelled
	CustomerInfo Customer   `json:"customer_info"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Customer represents customer information
type Customer struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address Address `json:"address"`
}

// Address represents shipping address
type Address struct {
	Street   string `json:"street"`
	City     string `json:"city"`
	State    string `json:"state"`
	Country  string `json:"country"`
	ZipCode  string `json:"zip_code"`
}

// InventoryUpdate represents inventory synchronization data
type InventoryUpdate struct {
	ProductID   string    `json:"product_id"`
	Marketplace string    `json:"marketplace"`
	Stock       int       `json:"stock"`
	Price       float64   `json:"price"`
	UpdateType  string    `json:"update_type"` // stock, price, both
	Timestamp   time.Time `json:"timestamp"`
}

// ProcessingEvent represents events in the processing pipeline
type ProcessingEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // image_uploaded, ai_enhanced, seo_generated, listed
	ProductID string                 `json:"product_id"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"` // service name
}

// ServiceResponse represents a standard service response
type ServiceResponse struct {
	Success   bool                   `json:"success"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}
