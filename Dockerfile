# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Use vendor mode - no network download needed inside container
ENV GOFLAGS=-mod=vendor

# Copy go mod files and vendor directory
COPY go.mod go.sum ./
COPY vendor/ ./vendor/

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk add --no-cache ca-certificates tzdata

# Copy binary
COPY --from=builder /app/server .

# Create uploads directory
RUN mkdir -p /app/uploads/firmwares

# Expose port
EXPOSE 8080

# Run
CMD ["./server"]
