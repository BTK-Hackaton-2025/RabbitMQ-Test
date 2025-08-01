# 🐳 Dockerized E-commerce RabbitMQ System

## 🏗️ Architecture Overview

This project demonstrates a **production-ready, containerized e-commerce order processing system** using RabbitMQ with Docker best practices.

### 🎯 What This System Demonstrates

- **Multi-stage Docker builds** for optimal image size
- **Docker secrets** for secure credential management
- **Service scaling** and load balancing
- **Health checks** and monitoring
- **Network isolation** and security
- **Persistent data storage**
- **Production-ready configuration**

## 🏢 Business Architecture

```
Order Placement →
├─ 📋 Work Queue: Order processing (load balanced across workers)
├─ 📡 Pub/Sub: Notify all services (inventory, email, analytics)
└─ 🎯 Routing: Regional fulfillment (US, EU, ASIA)
```

## 🐳 Docker Best Practices Implemented

### 🔒 Security Best Practices

- ✅ **Docker secrets** instead of environment variables
- ✅ **Non-root user** in containers
- ✅ **Minimal base images** (scratch for Go apps)
- ✅ **Network isolation** with custom networks
- ✅ **Resource limits** to prevent DoS
- ✅ **Health checks** for service monitoring

### 🚀 Performance Best Practices

- ✅ **Multi-stage builds** (smaller final images)
- ✅ **Layer caching** optimization
- ✅ **Static binary compilation** for Go
- ✅ **Persistent volumes** for data
- ✅ **Service scaling** capabilities

### 🛠️ Operational Best Practices

- ✅ **Graceful shutdown** handling
- ✅ **Restart policies** for resilience
- ✅ **Centralized logging**
- ✅ **Configuration management**
- ✅ **Service dependencies** with health checks

## 🚀 Quick Start

### Prerequisites

- Docker 20.10+
- Docker Compose 2.0+

### 1. Deploy the System

```bash
cd docker/
./deploy.sh
```

### 2. Choose Option 1 to Deploy All Services

The script will:

- Build optimized Docker images
- Start all services with proper dependencies
- Set up networking and volumes
- Configure secrets securely

### 3. Access the System

- **RabbitMQ Management**: http://localhost:15672
- **Username**: `ecommerce_user`
- **Password**: `ecommerce_pass_2024_secure!`

## 📋 Services Overview

| Service              | Purpose          | Scaling   | Pattern        |
| -------------------- | ---------------- | --------- | -------------- |
| **rabbitmq**         | Message broker   | Single    | N/A            |
| **producer**         | Order placement  | Manual    | Interactive    |
| **processor**        | Order processing | Auto (3x) | Work Queue     |
| **inventory**        | Stock management | Single    | Pub/Sub        |
| **email**            | Notifications    | Single    | Pub/Sub        |
| **analytics**        | Data collection  | Single    | Pub/Sub        |
| **fulfillment-us**   | US shipping      | Single    | Direct Routing |
| **fulfillment-eu**   | EU shipping      | Single    | Direct Routing |
| **fulfillment-asia** | Asia shipping    | Single    | Direct Routing |

## 🧪 Testing the System

### 1. Place Orders

```bash
./deploy.sh
# Choose option 5 (Run interactive producer)

# Then place orders:
user123:laptop:999.99:US:express
user456:phone:599.99:EU:standard
user789:tablet:399.99:ASIA:express
```

### 2. Monitor Processing

```bash
# Watch logs in real-time
docker-compose logs -f

# Or use the script
./deploy.sh
# Choose option 3 (Show logs)
```

### 3. Scale Workers

```bash
# Scale processor workers to handle more load
./deploy.sh
# Choose option 4 (Scale processor workers)
# Enter desired number of replicas
```

## 🔒 Security Features

### Docker Secrets Management

```yaml
secrets:
  rabbitmq_user:
    file: ./secrets/rabbitmq_user.txt
  rabbitmq_password:
    file: ./secrets/rabbitmq_password.txt
```

**Why this is secure:**

- ✅ Credentials never appear in process lists
- ✅ Not stored in environment variables
- ✅ Mounted as in-memory files
- ✅ Can be rotated without rebuilding images

### Network Isolation

```yaml
networks:
  ecommerce_network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

**Benefits:**

- ✅ Services isolated from host network
- ✅ Only exposed ports are accessible
- ✅ Internal service communication secured

## 📊 Performance Monitoring

### Resource Limits

```yaml
deploy:
  resources:
    limits:
      memory: 128M
      cpus: "0.5"
    reservations:
      memory: 64M
      cpus: "0.25"
```

### Health Checks

```yaml
healthcheck:
  test: ["CMD", "rabbitmq-diagnostics", "ping"]
  interval: 30s
  timeout: 10s
  retries: 3
```

## 🏭 Production Deployment

### Environment Variables

```bash
# Production settings
ENVIRONMENT=production
LOG_LEVEL=warn
RABBITMQ_URL=amqp://user:pass@rabbitmq-cluster:5672/

# Scaling
PROCESSOR_REPLICAS=5
```

### High Availability Setup

```yaml
# For production, add:
deploy:
  replicas: 3
  update_config:
    parallelism: 1
    delay: 10s
  restart_policy:
    condition: on-failure
```

## 🔧 Advanced Configuration

### Custom RabbitMQ Config

Edit `config/rabbitmq.conf`:

```ini
# Performance tuning
vm_memory_high_watermark.relative = 0.8
disk_free_limit.absolute = 2GB

# Clustering
cluster_formation.peer_discovery_backend = classic_config
```

### Scaling Strategies

```bash
# Scale specific services
docker-compose up -d --scale processor=5
docker-compose up -d --scale fulfillment-us=2

# Auto-scaling (requires orchestrator like Kubernetes)
kubectl autoscale deployment processor --cpu-percent=70 --min=2 --max=10
```

## 🐛 Troubleshooting

### Common Issues

**1. RabbitMQ won't start**

```bash
# Check logs
docker-compose logs rabbitmq

# Verify secrets
ls -la secrets/
cat secrets/rabbitmq_user.txt
```

**2. Services can't connect to RabbitMQ**

```bash
# Check network
docker network ls
docker network inspect docker_ecommerce_network

# Test connectivity
docker-compose exec processor ping rabbitmq
```

**3. Out of memory errors**

```bash
# Check resource usage
docker stats

# Increase limits in docker-compose.yml
memory: 512M  # Increase from 128M
```

## 📈 Monitoring and Logging

### Built-in Monitoring

- **RabbitMQ Management UI**: Queue lengths, message rates
- **Docker stats**: Resource usage
- **Health checks**: Service availability

### Production Monitoring (Optional)

```bash
# Enable monitoring stack
docker-compose --profile monitoring up -d

# Access Prometheus
http://localhost:9090
```

## 🎓 Learning Outcomes

After running this system, you'll understand:

- ✅ **Docker secrets** vs environment variables
- ✅ **Multi-stage builds** for production images
- ✅ **Service networking** and isolation
- ✅ **Container orchestration** with dependencies
- ✅ **Horizontal scaling** of microservices
- ✅ **Health checks** and monitoring
- ✅ **Persistent storage** for stateful services
- ✅ **Security hardening** for containers

## 🚀 Next Steps

1. **Add database persistence** (PostgreSQL/MongoDB)
2. **Implement API Gateway** (Nginx/Traefik)
3. **Add monitoring stack** (Prometheus/Grafana)
4. **Set up CI/CD pipeline** (GitHub Actions)
5. **Deploy to Kubernetes** for production

---

**🎉 Congratulations!** You now have a production-ready, containerized RabbitMQ system that demonstrates industry best practices for Docker deployment.
