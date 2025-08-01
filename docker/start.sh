#!/bin/bash

echo "🐳 Building and starting E-commerce RabbitMQ system..."

# Build the images
echo "📦 Building Docker images..."
docker-compose build

# Start the services
echo "🚀 Starting services..."
docker-compose up -d

# Show status
echo "📊 Service status:"
docker-compose ps

echo ""
echo "✅ System is ready!"
echo "🌐 RabbitMQ Management: http://localhost:15672"
echo "👤 Username: ecommerce_user"
echo "🔑 Password: ecommerce_pass_2024_secure!"
echo ""
echo "📋 Available commands:"
echo "  docker-compose logs -f                    # View logs"
echo "  docker-compose exec producer /app         # Run producer"
echo "  docker-compose up -d --scale processor=5  # Scale workers"
echo "  docker-compose down                       # Stop all"
