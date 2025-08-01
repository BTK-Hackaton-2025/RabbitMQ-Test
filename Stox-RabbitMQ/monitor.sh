#!/bin/bash

# Stox RabbitMQ Process Monitoring Script
# Real-time monitoring of microservices and message queues

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="stox"
RABBITMQ_CONTAINER="stox-rabbitmq"
REFRESH_INTERVAL=2

# Function to clear screen and show header
show_header() {
    clear
    echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║${NC}                    ${CYAN}🎛️  STOX RABBITMQ MONITORING DASHBOARD${NC}                    ${BLUE}║${NC}"
    echo -e "${BLUE}║${NC}                         ${PURPLE}Real-time Process Management${NC}                          ${BLUE}║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════════════════╝${NC}"
    echo
    echo -e "${YELLOW}📅 $(date '+%Y-%m-%d %H:%M:%S')${NC} | ${GREEN}🔄 Auto-refresh: ${REFRESH_INTERVAL}s${NC} | ${BLUE}Press Ctrl+C to exit${NC}"
    echo
}

# Function to show service status
show_services() {
    echo -e "${CYAN}🚀 MICROSERVICES STATUS${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # Get container status
    if ! docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" --filter "name=stox-" > /tmp/stox_containers.txt 2>/dev/null; then
        echo -e "${RED}❌ Error: Unable to fetch container status${NC}"
        return
    fi
    
    # Check if any containers are running
    if [ $(wc -l < /tmp/stox_containers.txt) -le 1 ]; then
        echo -e "${YELLOW}⚠️  No Stox containers are currently running${NC}"
        echo -e "${BLUE}💡 Run './docker-manager.sh start' to start the platform${NC}"
        return
    fi
    
    # Parse and display services
    tail -n +2 /tmp/stox_containers.txt | while IFS=$'\t' read -r name status ports; do
        service_name=$(echo "$name" | sed 's/stox-//g')
        
        if echo "$status" | grep -q "Up"; then
            status_icon="🟢"
            status_color="${GREEN}"
            uptime=$(echo "$status" | grep -o "Up [^,]*" | sed 's/Up //')
        else
            status_icon="🔴"
            status_color="${RED}"
            uptime="Stopped"
        fi
        
        printf "%-20s %s %s%-15s%s %s\n" \
            "$service_name" \
            "$status_icon" \
            "$status_color" \
            "$uptime" \
            "${NC}" \
            "$ports"
    done
    
    echo
}

# Function to show queue information
show_queues() {
    echo -e "${CYAN}📬 MESSAGE QUEUES${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    if ! docker exec $RABBITMQ_CONTAINER rabbitmqctl list_queues name messages consumers memory 2>/dev/null | tail -n +2 > /tmp/stox_queues.txt; then
        echo -e "${RED}❌ Error: Unable to fetch queue information${NC}"
        echo -e "${YELLOW}💡 Make sure RabbitMQ container is running${NC}"
        return
    fi
    
    total_messages=0
    total_consumers=0
    
    printf "%-25s %s %s %s\n" "QUEUE NAME" "MESSAGES" "CONSUMERS" "MEMORY"
    echo "─────────────────────────────────────────────────────────────────────────"
    
    while read -r line; do
        if [ -n "$line" ]; then
            queue_name=$(echo "$line" | awk '{print $1}')
            messages=$(echo "$line" | awk '{print $2}')
            consumers=$(echo "$line" | awk '{print $3}')
            memory=$(echo "$line" | awk '{print $4}')
            
            # Color code based on message count
            if [ "$messages" -gt 100 ]; then
                msg_color="${RED}"
            elif [ "$messages" -gt 10 ]; then
                msg_color="${YELLOW}"
            else
                msg_color="${GREEN}"
            fi
            
            # Consumer status
            if [ "$consumers" -gt 0 ]; then
                consumer_color="${GREEN}"
            else
                consumer_color="${RED}"
            fi
            
            printf "%-25s %s%8s%s %s%9s%s %8s\n" \
                "$queue_name" \
                "$msg_color" "$messages" "${NC}" \
                "$consumer_color" "$consumers" "${NC}" \
                "$memory"
            
            total_messages=$((total_messages + messages))
            total_consumers=$((total_consumers + consumers))
        fi
    done < /tmp/stox_queues.txt
    
    echo "─────────────────────────────────────────────────────────────────────────"
    printf "%-25s %s%8d%s %s%9d%s\n" \
        "TOTAL" \
        "${CYAN}" "$total_messages" "${NC}" \
        "${CYAN}" "$total_consumers" "${NC}"
    
    echo
}

# Function to show system metrics
show_metrics() {
    echo -e "${CYAN}📊 SYSTEM METRICS${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # RabbitMQ status
    if docker exec $RABBITMQ_CONTAINER rabbitmq-diagnostics ping &>/dev/null; then
        rabbitmq_status="${GREEN}🟢 Connected${NC}"
    else
        rabbitmq_status="${RED}🔴 Disconnected${NC}"
    fi
    
    # Docker stats
    running_containers=$(docker ps --filter "name=stox-" --format "{{.Names}}" | wc -l)
    total_containers=$(docker ps -a --filter "name=stox-" --format "{{.Names}}" | wc -l)
    
    # System load
    if command -v uptime &> /dev/null; then
        load_avg=$(uptime | awk -F'load average:' '{ print $2 }' | sed 's/^[ \t]*//')
    else
        load_avg="N/A"
    fi
    
    printf "%-25s %s\n" "RabbitMQ Status:" "$rabbitmq_status"
    printf "%-25s %s\n" "Running Containers:" "${running_containers}/${total_containers}"
    printf "%-25s %s\n" "System Load:" "$load_avg"
    printf "%-25s %s\n" "Management UI:" "${BLUE}http://localhost:15672${NC}"
    
    echo
}

# Function to show recent logs
show_recent_activity() {
    echo -e "${CYAN}📝 RECENT ACTIVITY${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # Show last 5 log entries from each service
    for service in image-service ai-service seo-service; do
        if docker ps --filter "name=stox-$service" --format "{{.Names}}" | grep -q "stox-$service"; then
            echo -e "${YELLOW}📋 $service:${NC}"
            docker logs --tail=2 "stox-$service" 2>/dev/null | sed 's/^/  /' || echo "  No recent logs"
        fi
    done
    
    echo
}

# Function to show control options
show_controls() {
    echo -e "${CYAN}🎮 CONTROL OPTIONS${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${BLUE}Commands available:${NC}"
    echo "  ./docker-manager.sh start     - Start all services"
    echo "  ./docker-manager.sh stop      - Stop all services"
    echo "  ./docker-manager.sh restart   - Restart all services"
    echo "  ./docker-manager.sh status    - Show detailed status"
    echo "  ./docker-manager.sh logs      - Show all logs"
    echo "  ./docker-manager.sh scale ai-service 5  - Scale AI workers"
    echo
}

# Function for interactive mode
interactive_mode() {
    while true; do
        show_header
        show_services
        show_queues
        show_metrics
        show_recent_activity
        show_controls
        
        echo -e "${GREEN}Press any key to refresh, or Ctrl+C to exit...${NC}"
        sleep $REFRESH_INTERVAL
    done
}

# Function for one-time status
status_mode() {
    show_header
    show_services
    show_queues
    show_metrics
    show_recent_activity
}

# Main script logic
main() {
    case "${1:-interactive}" in
        interactive|monitor)
            interactive_mode
            ;;
        status)
            status_mode
            ;;
        services)
            show_services
            ;;
        queues)
            show_queues
            ;;
        metrics)
            show_metrics
            ;;
        help|--help|-h)
            echo "Stox RabbitMQ Process Monitor"
            echo "Usage: $0 [interactive|status|services|queues|metrics|help]"
            echo
            echo "  interactive (default) - Real-time monitoring dashboard"
            echo "  status              - One-time status report"
            echo "  services            - Show only service status"
            echo "  queues              - Show only queue information"
            echo "  metrics             - Show only system metrics"
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Cleanup function
cleanup() {
    echo
    echo -e "${YELLOW}🛑 Monitoring stopped${NC}"
    rm -f /tmp/stox_containers.txt /tmp/stox_queues.txt
    exit 0
}

# Set up signal handling
trap cleanup INT TERM

# Run main function
main "$@"
