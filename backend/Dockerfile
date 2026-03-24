# Multi-stage build for Agent Orchestrator Backend
FROM golang:1.24-alpine AS builder

# Install build dependencies for SQLite (CGO)
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the server binary with CGO enabled for SQLite
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/server .

# Copy example env file (will be overridden by mounted .env)
COPY --from=builder /app/.env.example .env

# Create data directory for SQLite
RUN mkdir -p /data

# Expose the port
EXPOSE 18765

# Health check
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:18765/health || exit 1

# Run the server
CMD ["./server"]
