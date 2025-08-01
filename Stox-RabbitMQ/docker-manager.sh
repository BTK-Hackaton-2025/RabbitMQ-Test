#!/bin/bash

# Stox RabbitMQ Docker Management Script
# This script manages the entire Stox e-commerce automation platform

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.yml"
PROJECT_NAME="stox"

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

# Function to check if Docker is running
check_docker() {
    if ! docker info >/dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker first."
        exit 1
    fi
    print_success "Docker is running"
}

# Function to check if docker-compose is available
check_compose() {
    if ! command -v docker-compose >/dev/null 2>&1; then
        if ! docker compose version >/dev/null 2>&1; then
            print_error "Neither docker-compose nor 'docker compose' is available"
            exit 1
        else
            COMPOSE_CMD="docker compose"
        fi
    else
        COMPOSE_CMD="docker-compose"
    fi
    print_success "Docker Compose is available: $COMPOSE_CMD"
}

# Function to build all services
build_services() {
    print_status "Building all Stox microservices..."
    $COMPOSE_CMD -p $PROJECT_NAME build --parallel
    print_success "All services built successfully"
}

# Function to start the platform
start_platform() {
    print_status "Starting Stox E-Commerce Automation Platform..."
    
    # Start RabbitMQ first
    print_status "Starting RabbitMQ message broker..."
    $COMPOSE_CMD -p $PROJECT_NAME up -d rabbitmq
    
    # Wait for RabbitMQ to be healthy
    print_status "Waiting for RabbitMQ to be ready..."
    timeout=60
    while [ $timeout -gt 0 ]; do
        if docker exec stox-rabbitmq rabbitmq-diagnostics ping >/dev/null 2>&1; then
            print_success "RabbitMQ is ready"
            break
        fi
        sleep 2
        timeout=$((timeout - 2))
    done
    
    if [ $timeout -le 0 ]; then
        print_error "RabbitMQ failed to start within 60 seconds"
        exit 1
    fi
    
    # Start all microservices
    print_status "Starting all microservices..."
    $COMPOSE_CMD -p $PROJECT_NAME up -d
    
    print_success "Platform started successfully!"
    print_status "Services running:"
    $COMPOSE_CMD -p $PROJECT_NAME ps
}

# Function to stop the platform
stop_platform() {
    print_status "Stopping Stox platform..."
    $COMPOSE_CMD -p $PROJECT_NAME down
    print_success "Platform stopped"
}

# Function to restart the platform
restart_platform() {
    print_status "Restarting Stox platform..."
    stop_platform
    start_platform
}

# Function to show logs
show_logs() {
    service=${1:-""}
    if [ -n "$service" ]; then
        print_status "Showing logs for $service..."
        $COMPOSE_CMD -p $PROJECT_NAME logs -f $service
    else
        print_status "Showing logs for all services..."
        $COMPOSE_CMD -p $PROJECT_NAME logs -f
    fi
}

# Function to show status
show_status() {
    print_status "Stox Platform Status:"
    echo
    $COMPOSE_CMD -p $PROJECT_NAME ps
    echo
    
    # Check RabbitMQ management UI
    if curl -s http://localhost:15672 >/dev/null 2>&1; then
        print_success "RabbitMQ Management UI: http://localhost:15672 (stox/stoxpass123)"
    else
        print_warning "RabbitMQ Management UI not accessible"
    fi
    
    # Show queue statistics
    print_status "Queue Statistics:"
    if command -v rabbitmqctl >/dev/null 2>&1; then
        docker exec stox-rabbitmq rabbitmqctl list_queues name messages consumers
    fi
}

# Function to run demo
run_demo() {
    print_status "Running demonstration..."
    $COMPOSE_CMD -p $PROJECT_NAME --profile demo up demo-service
    print_success "Demo completed"
}

# Function to scale services
scale_service() {
    service=$1
    replicas=$2
    
    if [ -z "$service" ] || [ -z "$replicas" ]; then
        print_error "Usage: scale <service> <replicas>"
        exit 1
    fi
    
    print_status "Scaling $service to $replicas replicas..."
    $COMPOSE_CMD -p $PROJECT_NAME up -d --scale $service=$replicas $service
    print_success "$service scaled to $replicas replicas"
}

# Function to monitor system
monitor() {
    print_status "Starting monitoring mode..."
    $COMPOSE_CMD -p $PROJECT_NAME --profile monitoring up -d rabbitmq-exporter
    print_success "Monitoring started. Metrics available at http://localhost:9419/metrics"
}

# Function to cleanup
cleanup() {
    print_status "Cleaning up Stox platform..."
    $COMPOSE_CMD -p $PROJECT_NAME down -v --remove-orphans
    docker system prune -f
    print_success "Cleanup completed"
}

# Function to backup RabbitMQ data
backup() {
    backup_dir="./backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    print_status "Creating backup in $backup_dir..."
    
    # Export definitions
    docker exec stox-rabbitmq rabbitmqctl export_definitions /tmp/definitions.json
    docker cp stox-rabbitmq:/tmp/definitions.json "$backup_dir/"
    
    # Backup persistent data
    docker run --rm -v stox_rabbitmq_data:/data -v "$(pwd)/$backup_dir":/backup alpine tar czf /backup/rabbitmq_data.tar.gz -C /data .
    
    print_success "Backup created: $backup_dir"
}

# Function to show help
show_help() {
    echo "Stox RabbitMQ Docker Management"
    echo "Usage: $0 <command> [options]"
    echo
    echo "Commands:"
    echo "  build              Build all microservices"
    echo "  start              Start the entire platform"
    echo "  stop               Stop the platform"
    echo "  restart            Restart the platform"
    echo "  status             Show platform status"
    echo "  logs [service]     Show logs (all services or specific service)"
    echo "  demo               Run demonstration"
    echo "  scale <svc> <n>    Scale service to n replicas"
    echo "  monitor            Start monitoring"
    echo "  backup             Backup RabbitMQ data"
    echo "  cleanup            Stop and remove all containers and volumes"
    echo "  help               Show this help"
    echo
    echo "Examples:"
    echo "  $0 start                    # Start all services"
    echo "  $0 logs ai-service          # Show AI service logs"
    echo "  $0 scale ai-service 5       # Scale AI service to 5 replicas"
    echo "  $0 status                   # Show current status"
}

# Main script logic
main() {
    check_docker
    check_compose
    
    case "${1:-help}" in
        build)
            build_services
            ;;
        start)
            build_services
            start_platform
            ;;
        stop)
            stop_platform
            ;;
        restart)
            restart_platform
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs "$2"
            ;;
        demo)
            run_demo
            ;;
        scale)
            scale_service "$2" "$3"
            ;;
        monitor)
            monitor
            ;;
        backup)
            backup
            ;;
        cleanup)
            cleanup
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
