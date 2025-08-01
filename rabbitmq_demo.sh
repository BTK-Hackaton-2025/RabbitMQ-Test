#!/bin/bash

echo "🐰 RabbitMQ Learning Demo"
echo "========================="
echo
echo "This script demonstrates different RabbitMQ patterns"
echo "Make sure RabbitMQ is running before proceeding!"
echo

# Function to wait for user input
wait_for_user() {
    echo "Press Enter to continue..."
    read
}

# 1. Simple Messaging (Your current setup)
echo "📝 1. SIMPLE MESSAGING PATTERN"
echo "Your current setup in cmd/send and cmd/receive"
echo "- Direct queue communication"
echo "- One producer → one consumer"
echo "- Interactive message sending"
wait_for_user

# 2. Work Queue Demo
echo "📋 2. WORK QUEUE PATTERN"
echo "Demonstrates task distribution among workers"
echo
echo "In Terminal 1, run: cd examples/work-queue && go run producer.go"
echo "In Terminal 2, run: cd examples/work-queue && go run worker.go"
echo "In Terminal 3, run: cd examples/work-queue && go run worker.go"
echo
echo "Try sending tasks like:"
echo "  - quicktask"
echo "  - slow.task....."
echo "  - heavy.processing.task........."
echo
echo "Notice how tasks are distributed among workers!"
wait_for_user

# 3. Pub/Sub Demo
echo "📡 3. PUBLISH/SUBSCRIBE PATTERN"
echo "Broadcasts messages to all subscribers"
echo
echo "In Terminal 1, run: cd examples/pub-sub && go run publisher.go"
echo "In Terminal 2, run: cd examples/pub-sub && go run subscriber.go BBC"
echo "In Terminal 3, run: cd examples/pub-sub && go run subscriber.go CNN"
echo "In Terminal 4, run: cd examples/pub-sub && go run subscriber.go Reuters"
echo
echo "All subscribers will receive the same news!"
wait_for_user

# 4. Routing Demo
echo "🎯 4. ROUTING PATTERN"
echo "Messages routed based on severity/type"
echo
echo "In Terminal 1, run: cd examples/routing && go run log_producer.go"
echo "In Terminal 2, run: cd examples/routing && go run log_consumer.go error"
echo "In Terminal 3, run: cd examples/routing && go run log_consumer.go info warning"
echo "In Terminal 4, run: cd examples/routing && go run log_consumer.go error warning info debug"
echo
echo "Try sending different log levels:"
echo "  - error:Database connection failed"
echo "  - warning:High memory usage detected"
echo "  - info:User logged in successfully"
echo "  - debug:Processing request payload"
wait_for_user

echo "🎉 Demo complete! Choose any pattern to explore further."
