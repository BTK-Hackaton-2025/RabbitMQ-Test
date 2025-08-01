# 🛍️ Stox E-Commerce Automation Platform - RabbitMQ Implementation

## 🎯 **What We Built**

A complete **mock e-commerce automation platform** that demonstrates **real RabbitMQ patterns** for your production system. This shows exactly how to integrate RabbitMQ into microservices for:

- **Image Upload & AI Enhancement**
- **SEO Content Generation**
- **Multi-Marketplace Broadcasting** (Amazon, Trendyol, Hepsiburada)
- **Real-time Inventory Synchronization**
- **Cross-platform Order Management**

## 🐰 **RabbitMQ Patterns Implemented**

### 1. **Work Queue Pattern**

- **Queue:** `image_processing`
- **Use Case:** Load balancing AI enhancement across multiple workers
- **Files:** `cmd/ai-service/main.go` (3 concurrent AI workers)

### 2. **Pub/Sub Pattern (Fanout Exchange)**

- **Exchange:** `stox.listings`
- **Use Case:** Broadcast product listings to ALL marketplaces simultaneously
- **Files:** `cmd/seo-service/main.go` → broadcasts to all marketplace services

### 3. **Topic Routing Pattern**

- **Exchange:** `stox.orders`
- **Routing Keys:** `order.amazon.us`, `order.trendyol.tr`, `order.hepsiburada.tr`
- **Use Case:** Route orders by marketplace and region
- **Files:** All marketplace services consume their specific order patterns

### 4. **Direct Routing Pattern**

- **Exchange:** `stox.sync`
- **Routing Keys:** `amazon_sync`, `trendyol_sync`, `hepsiburada_sync`
- **Use Case:** Direct inventory/price updates to specific marketplaces
- **Files:** `cmd/sync-service/main.go` → direct sync operations

## 📁 **Project Structure**

```
stox-rabbitmq/
├── cmd/                           # 🎯 Service Entry Points
│   ├── image-service/main.go      # 📸 Image upload & validation
│   ├── ai-service/main.go         # 🤖 AI enhancement (3 workers)
│   ├── seo-service/main.go        # 📝 SEO content generation
│   ├── amazon-service/main.go     # 🛒 Amazon marketplace
│   ├── trendyol-service/main.go   # 🇹🇷 Trendyol marketplace
│   ├── hepsiburada-service/main.go# 🟠 Hepsiburada marketplace
│   ├── sync-service/main.go       # 🔄 Inventory synchronization
│   └── demo/main.go               # 🎬 Complete demo pipeline
├── internal/                      # 🔧 Internal Packages
│   ├── rabbitmq/client.go         # 🐰 RabbitMQ wrapper client
│   ├── models/models.go           # 📊 Data structures
│   └── config/config.go           # ⚙️ Configuration management
└── start-demo.sh                  # 🚀 One-click demo script
```

## 🚀 **How to Run**

### **Option 1: Quick Demo** ⚡

```bash
cd /Users/altugmalkan/Desktop/go-rabbitmq/Stox-RabbitMQ
./start-demo.sh
```

### **Option 2: Manual Step-by-Step** 🔧

1. **Start RabbitMQ:**

```bash
docker run -d --name stox-rabbitmq \
  -p 5672:5672 -p 15672:15672 \
  -e RABBITMQ_DEFAULT_USER=stox \
  -e RABBITMQ_DEFAULT_PASS=stoxpass123 \
  rabbitmq:3.12-management-alpine
```

2. **Start Services** (in separate terminals):

```bash
# Terminal 1: Image Service
go run cmd/image-service/main.go

# Terminal 2: AI Service
go run cmd/ai-service/main.go

# Terminal 3: SEO Service
go run cmd/seo-service/main.go

# Terminal 4-6: Marketplace Services
go run cmd/amazon-service/main.go
go run cmd/trendyol-service/main.go
go run cmd/hepsiburada-service/main.go

# Terminal 7: Sync Service
go run cmd/sync-service/main.go

# Terminal 8: Demo Pipeline
go run cmd/demo/main.go
```

## 🎮 **Demo Flow**

### **Phase 1: Image Processing Pipeline**

```
📸 Product Upload → 🤖 AI Enhancement → 📝 SEO Generation → 🏪 Marketplace Broadcasting
```

1. **Image Service** receives product with photos
2. **AI Service** (3 workers) enhances images in parallel
3. **SEO Service** generates optimized content using mock RAG
4. **Fanout Exchange** broadcasts to ALL marketplaces simultaneously

### **Phase 2: Multi-Marketplace Listing**

```
🛒 Amazon + 🇹🇷 Trendyol + 🟠 Hepsiburada = Simultaneous Listings
```

- Amazon: USD pricing with 10% markup
- Trendyol: TL pricing with 8% markup
- Hepsiburada: TL pricing with 12% markup

### **Phase 3: Order Processing & Sync**

```
📦 Orders → 🎯 Topic Routing → 🔄 Inventory Sync
```

- Orders routed by marketplace and region
- Real-time inventory synchronization across platforms
- Cross-platform stock management

## 🔍 **Monitoring & Debugging**

### **RabbitMQ Management UI**

- **URL:** http://localhost:15672
- **Username:** `stox`
- **Password:** `stoxpass123`

### **Key Exchanges to Monitor:**

- `stox.images` - Topic exchange for image processing
- `stox.listings` - Fanout exchange for marketplace broadcasting
- `stox.sync` - Direct exchange for inventory sync
- `stox.orders` - Topic exchange for order routing

### **Key Queues to Monitor:**

- `image_processing` - AI enhancement work queue
- `amazon_listings`, `trendyol_listings`, `hepsiburada_listings` - Marketplace queues
- `inventory_updates` - Sync operations

## 💡 **Key Learning Points**

### **1. Exchange Types Usage:**

- **Fanout:** When you need to send to ALL subscribers (marketplace broadcasting)
- **Topic:** When you need pattern-based routing (orders by marketplace+region)
- **Direct:** When you need exact routing (specific marketplace sync)

### **2. Work Queue Benefits:**

- Load balancing across multiple AI workers
- Automatic message distribution
- Built-in failure handling with manual acknowledgments

### **3. Message Flow Patterns:**

- **Pipeline:** Image → AI → SEO → Marketplaces
- **Broadcasting:** SEO → All Marketplaces
- **Routing:** Orders → Specific Marketplace Services
- **Synchronization:** Inventory Changes → All/Specific Marketplaces

### **4. Production-Ready Features:**

- **Manual Acknowledgments:** Ensures message processing
- **Durable Queues:** Messages persist across restarts
- **Error Handling:** Proper error logging and recovery
- **Health Checks:** Connection monitoring

## 🔧 **Code Architecture Highlights**

### **RabbitMQ Client Wrapper** (`internal/rabbitmq/client.go`)

```go
// High-level operations for your services
client.SetupExchanges()          // Declares all exchanges
client.DeclareQueue(name, exchange, routing)  // Queue management
client.PublishMessage(exchange, routing, data) // Send messages
client.ConsumeMessages(queue, handler)         // Receive messages
```

### **Service Pattern** (All cmd/ services)

```go
// Standard pattern for all services:
1. Load configuration
2. Create RabbitMQ client
3. Setup exchanges and queues
4. Start message consumers
5. Handle graceful shutdown
```

### **Message Types** (`internal/models/models.go`)

```go
Product          // Core product data
Image            // Image processing info
SEOData          // Generated content
MarketplaceListing // Platform-specific listings
Order            // Cross-platform orders
InventoryUpdate  // Sync operations
ProcessingEvent  // Pipeline events
```

## 🚀 **Next Steps for Production**

### **1. Add Real Integrations:**

- Replace mocks with actual marketplace APIs
- Integrate real AI services (OpenAI, AWS Rekognition)
- Add PostgreSQL with pgvector for embeddings
- Implement S3 for image storage

### **2. Add gRPC & API Gateway:**

- Convert inter-service communication to gRPC
- Add REST API gateway for external clients
- Implement authentication & rate limiting

### **3. Add Container Orchestration:**

- Docker Compose for local development
- Kubernetes for production deployment
- Service mesh for advanced networking

### **4. Add Monitoring & Observability:**

- Prometheus + Grafana for metrics
- Distributed tracing with Jaeger
- Centralized logging with ELK stack

### **5. Add Resilience Patterns:**

- Circuit breakers for external APIs
- Retry logic with exponential backoff
- Dead letter queues for failed messages
- Health checks and auto-recovery

---

## 🎉 **Success!**

You now have a **complete RabbitMQ implementation** showing **exactly** how to integrate message queues into your e-commerce automation platform!

**All patterns are production-ready** - just replace the mocks with real implementations and you're ready to scale! 🚀
