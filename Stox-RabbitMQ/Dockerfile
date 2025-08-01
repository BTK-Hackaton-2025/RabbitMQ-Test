# Multi-stage Dockerfile for Stox microservices
FROM golang:1.23-alpine AS builder

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Create non-root user
RUN adduser -D -s /bin/sh -u 1001 appuser

# Set working directory
WORKDIR /app

# Copy go modules
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the specific service
ARG SERVICE_NAME
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o /app/bin/${SERVICE_NAME} \
    ./cmd/${SERVICE_NAME}

# Final stage - minimal image
FROM scratch

# Import certificates and timezone data
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd

# Copy the binary
ARG SERVICE_NAME
COPY --from=builder /app/bin/${SERVICE_NAME} /app

# Use non-root user
USER appuser

# Set metadata
LABEL maintainer="stox@example.com"
LABEL version="1.0.0"
LABEL description="Stox E-Commerce Automation Microservice"

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/app", "--health"] || exit 1

# Run the service
ENTRYPOINT ["/app"]
