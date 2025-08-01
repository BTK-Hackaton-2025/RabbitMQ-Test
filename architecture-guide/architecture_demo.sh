#!/bin/bash

echo "🏗️ RabbitMQ Architecture Patterns Demo"
echo "======================================"
echo
echo "This demo shows when and how to use different RabbitMQ patterns"
echo

# Function to wait for user input
wait_for_user() {
    echo "Press Enter to continue..."
    read
}

echo "🎯 SCENARIO: You're building an e-commerce platform"
echo "Let's see how different components use different patterns:"
echo
wait_for_user

echo "📋 PATTERN 1: WORK QUEUE"
echo "Problem: Multiple order processing workers need to share the workload"
echo "Solution: Work Queue - distributes orders among available workers"
echo
echo "Demo:"
echo "Terminal 1: cd architecture-guide/scenarios && go run ecommerce_orders.go"
echo "Terminal 2: cd architecture-guide/scenarios && go run ecommerce_consumer.go processor"
echo "Terminal 3: cd architecture-guide/scenarios && go run ecommerce_consumer.go processor"
echo
echo "Try placing orders and see how they're distributed between workers!"
wait_for_user

echo "📡 PATTERN 2: PUB/SUB (Fanout)"
echo "Problem: When an order is placed, multiple services need to know"
echo "Solution: Fanout Exchange - broadcasts to inventory, email, analytics"
echo
echo "Demo (additional terminals):"
echo "Terminal 4: cd architecture-guide/scenarios && go run ecommerce_consumer.go inventory"
echo "Terminal 5: cd architecture-guide/scenarios && go run ecommerce_consumer.go email"
echo "Terminal 6: cd architecture-guide/scenarios && go run ecommerce_consumer.go analytics"
echo
echo "Each service gets the SAME order notification!"
wait_for_user

echo "🎯 PATTERN 3: ROUTING (Direct)"
echo "Problem: Orders need to go to different fulfillment centers by region"
echo "Solution: Direct Exchange - routes by region (US, EU, ASIA)"
echo
echo "Demo (additional terminals):"
echo "Terminal 7: cd architecture-guide/scenarios && go run ecommerce_consumer.go fulfillment_US"
echo "Terminal 8: cd architecture-guide/scenarios && go run ecommerce_consumer.go fulfillment_EU"
echo "Terminal 9: cd architecture-guide/scenarios && go run ecommerce_consumer.go fulfillment_ASIA"
echo
echo "Each fulfillment center only gets orders for their region!"
wait_for_user

echo "🧪 TEST SCENARIOS:"
echo
echo "1. Place a US order: user123:laptop:999.99:US:express"
echo "   Watch: Processor workers compete, all services notified, only US fulfillment"
echo
echo "2. Place an EU order: user456:phone:599.99:EU:standard"
echo "   Watch: Different distribution pattern"
echo
echo "3. Start multiple processors and see load balancing"
echo
echo "4. Stop a service and see fault tolerance"
wait_for_user

echo "🔍 ARCHITECTURE ANALYSIS:"
echo
echo "✅ Work Queue Benefits:"
echo "   - Load balancing across processors"
echo "   - Fault tolerance (if one worker dies, others continue)"
echo "   - Scalability (add more workers as needed)"
echo
echo "✅ Pub/Sub Benefits:"
echo "   - Loose coupling (services don't know about each other)"
echo "   - Easy to add new services (just subscribe)"
echo "   - Guaranteed delivery to all interested parties"
echo
echo "✅ Routing Benefits:"
echo "   - Geographic/logical separation"
echo "   - Selective processing"
echo "   - Resource optimization"
wait_for_user

echo "🚨 COMMON MISTAKES TO AVOID:"
echo
echo "❌ Using Pub/Sub for load balancing:"
echo "   DON'T: fanout → worker1, worker2 (both get same task)"
echo "   DO: queue → worker1 OR worker2 (load balanced)"
echo
echo "❌ Using Work Queue for notifications:"
echo "   DON'T: queue → service1 OR service2 (only one notified)"
echo "   DO: fanout → service1 AND service2 (both notified)"
echo
echo "❌ Over-engineering routing:"
echo "   DON'T: complex topic patterns when simple routing works"
echo "   DO: start simple, add complexity when needed"
wait_for_user

echo "🎓 WHEN TO USE WHAT - QUICK REFERENCE:"
echo
echo "📨 Need 1-to-1 communication? → Simple Queue"
echo "📋 Need load balancing? → Work Queue"
echo "📡 Need broadcasting? → Pub/Sub (Fanout)"
echo "🎯 Need selective routing? → Direct Exchange"
echo "🔀 Need pattern matching? → Topic Exchange"
echo
echo "💡 Pro Tip: Most real applications use a COMBINATION of patterns!"
echo "   Like our e-commerce example using all three patterns."

echo
echo "🎉 Demo complete! Now you understand when to use each pattern."
echo "Try the examples and see the patterns in action!"
