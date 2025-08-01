#!/bin/bash

# 🐳 E-commerce RabbitMQ Docker Setup Script
# This script demonstrates Docker best practices for production deployment

set -e

echo "🐳 E-commerce RabbitMQ Docker Deployment"
echo "========================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker and Docker Compose are installed
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Validate secrets
validate_secrets() {
    print_status "Validating secrets..."
    
    if [[ ! -f "secrets/rabbitmq_user.txt" ]]; then
        print_error "RabbitMQ user secret not found"
        exit 1
    fi
    
    if [[ ! -f "secrets/rabbitmq_password.txt" ]]; then
        print_error "RabbitMQ password secret not found"
        exit 1
    fi
    
    if [[ ! -f "secrets/rabbitmq_erlang_cookie.txt" ]]; then
        print_error "RabbitMQ Erlang cookie secret not found"
        exit 1
    fi
    
    print_success "Secrets validation passed"
}

# Build and deploy
deploy() {
    print_status "Building and deploying services..."
    
    # Build images
    print_status "Building Docker images..."
    docker-compose build --no-cache
    
    # Start services
    print_status "Starting services..."
    docker-compose up -d
    
    # Wait for RabbitMQ to be ready
    print_status "Waiting for RabbitMQ to be ready..."
    timeout 60 bash -c 'until docker-compose exec rabbitmq rabbitmq-diagnostics ping; do sleep 2; done'
    
    print_success "All services are running!"
}

# Show service status
show_status() {
    print_status "Service Status:"
    docker-compose ps
    
    echo ""
    print_status "RabbitMQ Management UI: http://localhost:15672"
    print_status "Username: ecommerce_user"
    print_status "Password: ecommerce_pass_2024_secure!"
}

# Show logs
show_logs() {
    print_status "Recent logs from all services:"
    docker-compose logs --tail=50
}

# Scale processors
scale_processors() {
    local replicas=${1:-3}
    print_status "Scaling processor workers to $replicas replicas..."
    docker-compose up -d --scale processor=$replicas
    print_success "Processor workers scaled to $replicas"
}

# Cleanup
cleanup() {
    print_warning "Stopping and removing all services..."
    docker-compose down -v
    docker system prune -f
    print_success "Cleanup completed"
}

# Interactive producer
run_producer() {
    print_status "Starting interactive order producer..."
    print_status "You can now place orders in the format: user_id:product:amount:region:priority"
    print_status "Example: user123:laptop:999.99:US:express"
    
    docker-compose exec producer /app
}

# Menu
show_menu() {
    echo ""
    echo "🐳 Docker Management Options:"
    echo "1) Deploy all services"
    echo "2) Show service status"
    echo "3) Show logs"
    echo "4) Scale processor workers"
    echo "5) Run interactive producer"
    echo "6) Cleanup (stop and remove all)"
    echo "7) Exit"
    echo ""
}

# Main execution
main() {
    check_prerequisites
    validate_secrets
    
    while true; do
        show_menu
        read -p "Choose an option (1-7): " choice
        
        case $choice in
            1)
                deploy
                show_status
                ;;
            2)
                show_status
                ;;
            3)
                show_logs
                ;;
            4)
                read -p "Number of processor replicas (default 3): " replicas
                scale_processors ${replicas:-3}
                ;;
            5)
                run_producer
                ;;
            6)
                cleanup
                ;;
            7)
                print_success "Goodbye!"
                exit 0
                ;;
            *)
                print_error "Invalid option. Please choose 1-7."
                ;;
        esac
    done
}

# Run main function
main "$@"
