# 🛍️ Stox E-Commerce Automation Platform - RabbitMQ Implementation

## 🎯 Project Overview

This is a **RabbitMQ-focused implementation** for an e-commerce automation platform that:

- Takes product photos and enhances them with AI
- Generates SEO-optimized content
- Lists products on multiple marketplaces (Amazon, Trendyol, Hepsiburada)
- Synchronizes inventory, pricing, and orders across platforms

## 🐰 RabbitMQ Architecture

### Message Flow Pipeline

```
📸 Image Upload → 🤖 AI Enhancement → 📝 SEO Generation → 🏪 Multi-Platform Listing → 📊 Sync Management
```

### Exchange Design

- **`stox.images`** (Topic) - Image processing pipeline
- **`stox.listings`** (Fanout) - Broadcast to all marketplaces
- **`stox.sync`** (Direct) - Real-time synchronization
- **`stox.orders`** (Topic) - Order routing by marketplace

### Services (All Mocked)

1. **Image Service** - Handles image uploads and validation
2. **AI Service** - Mock AI enhancement (background removal, quality improvement)
3. **SEO Service** - Mock content generation with RAG
4. **Amazon Service** - Mock Amazon marketplace integration
5. **Trendyol Service** - Mock Trendyol marketplace integration
6. **Hepsiburada Service** - Mock Hepsiburada marketplace integration
7. **Sync Service** - Mock inventory/price synchronization
8. **Order Service** - Mock order management

## 🚀 Quick Start

1. **Start RabbitMQ:**

   ```bash
   docker run -d --name rabbitmq \
     -p 5672:5672 -p 15672:15672 \
     -e RABBITMQ_DEFAULT_USER=stox \
     -e RABBITMQ_DEFAULT_PASS=stoxpass123 \
     rabbitmq:3.12-management-alpine
   ```

2. **Install Dependencies:**

   ```bash
   go mod init stox-rabbitmq
   go mod tidy
   ```

3. **Run Services:**

   ```bash
   # Terminal 1: Image Service
   go run cmd/image-service/main.go

   # Terminal 2: AI Service
   go run cmd/ai-service/main.go

   # Terminal 3: SEO Service
   go run cmd/seo-service/main.go

   # Terminal 4: Marketplace Services
   go run cmd/amazon-service/main.go
   go run cmd/trendyol-service/main.go
   go run cmd/hepsiburada-service/main.go

   # Terminal 5: Sync Service
   go run cmd/sync-service/main.go

   # Terminal 6: Demo Producer
   go run cmd/demo/main.go
   ```

## 📋 RabbitMQ Patterns Used

### 1. Work Queue (Image Processing)

- **Queue:** `image_processing`
- **Pattern:** Load balancing for AI enhancement
- **Use Case:** Multiple AI workers process images in parallel

### 2. Pub/Sub (Marketplace Broadcasting)

- **Exchange:** `stox.listings` (Fanout)
- **Pattern:** Broadcast product listings to all marketplaces
- **Use Case:** Single product → Multiple platform listings

### 3. Topic Routing (Order Management)

- **Exchange:** `stox.orders` (Topic)
- **Pattern:** Route orders by marketplace and region
- **Routes:** `order.amazon.us`, `order.trendyol.tr`, `order.hepsiburada.tr`

### 4. Direct Routing (Sync Operations)

- **Exchange:** `stox.sync` (Direct)
- **Pattern:** Direct inventory/price updates
- **Routes:** `inventory_sync`, `price_sync`, `stock_sync`

## 🔍 Monitoring

- **RabbitMQ Management UI:** http://localhost:15672
- **Username:** stox
- **Password:** stoxpass123

## 📁 Project Structure

```
stox-rabbitmq/
├── cmd/                    # Service entry points
│   ├── image-service/
│   ├── ai-service/
│   ├── seo-service/
│   ├── amazon-service/
│   ├── trendyol-service/
│   ├── hepsiburada-service/
│   ├── sync-service/
│   └── demo/
├── internal/               # Internal packages
│   ├── rabbitmq/          # RabbitMQ client wrapper
│   ├── models/            # Data models
│   └── config/            # Configuration
├── pkg/                   # Public packages
└── docker-compose.yml     # Container orchestration
```

## 🎮 Demo Scenarios

The demo will show:

1. **Image Upload** → AI Enhancement workflow
2. **Product Listing** → Multi-marketplace broadcasting
3. **Order Processing** → Cross-platform order routing
4. **Inventory Sync** → Real-time updates across platforms

All services are **mocked** but demonstrate real RabbitMQ message patterns you'll use in production.
