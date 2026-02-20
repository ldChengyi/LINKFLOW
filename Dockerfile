# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Go proxy for China
ENV GOPROXY=https://goproxy.cn,direct

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

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
