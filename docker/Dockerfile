# Multi-stage Dockerfile for Go applications
# Stage 1: Build stage
FROM golang:1.21-alpine AS builder

# Install necessary packages
RUN apk add --no-cache git ca-certificates tzdata

# Create non-root user for security
RUN adduser -D -s /bin/sh -u 1001 appuser

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
ARG APP_NAME
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o /app/bin/${APP_NAME} \
    ./cmd/${APP_NAME}

# Stage 2: Final stage
FROM scratch

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd

# Copy binary from builder stage
ARG APP_NAME
COPY --from=builder /app/bin/${APP_NAME} /app

# Use non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app", "--health-check"] || exit 1

# Set metadata
LABEL maintainer="your-email@example.com"
LABEL version="1.0.0"
LABEL description="E-commerce RabbitMQ microservice"

# Expose port (if needed)
EXPOSE 8080

# Run the application
ENTRYPOINT ["/app"]
