# Build backend from monorepo root to access local packages
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files first for better caching
COPY go.sum go.sum . 

# Download dependencies
RUN go mod download

# Copy source code - but avoid context issues with replace directives
COPY . .

# Build the server binary with CGO enabled for SQLite
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Expose the port
EXPOSE 18765

# Health check
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:18765/health || exit 1

# Run the server
CMD ["./server"]
