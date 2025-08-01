#!/bin/bash

echo "🚀 Starting Stox E-Commerce Automation Platform"
echo "=============================================="

# Check if RabbitMQ is running
if ! docker ps | grep -q stox-rabbitmq; then
    echo "❌ RabbitMQ not running. Starting RabbitMQ..."
    docker run -d --name stox-rabbitmq \
        -p 5672:5672 -p 15672:15672 \
        -e RABBITMQ_DEFAULT_USER=stox \
        -e RABBITMQ_DEFAULT_PASS=stoxpass123 \
        rabbitmq:3.12-management-alpine
    
    echo "⏳ Waiting for RabbitMQ to start..."
    sleep 15
fi

echo "✅ RabbitMQ is running"
echo "📊 Management UI: http://localhost:15672 (stox / stoxpass123)"
echo ""

# Function to run service in background
run_service() {
    local service_name=$1
    echo "🔧 Starting $service_name..."
    go run cmd/$service_name/main.go &
    sleep 2
}

# Start all services
echo "🏗️  Starting all microservices..."
echo ""

run_service "image-service"
run_service "ai-service" 
run_service "seo-service"
run_service "amazon-service"
run_service "trendyol-service"
run_service "hepsiburada-service"
run_service "sync-service"

echo "⏳ Waiting for all services to initialize..."
sleep 5

echo ""
echo "🎬 Starting demo pipeline..."
echo ""

# Run the demo
go run cmd/demo/main.go

echo ""
echo "🛑 Demo completed. Press Ctrl+C to stop all services."

# Wait for user interrupt
wait
