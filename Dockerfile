# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Set Go proxy for China
ENV GOPROXY=https://goproxy.cn,direct

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Use China mirror for faster package download
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# Install ca-certificates for HTTPS and Python for TTS
RUN apk add --no-cache ca-certificates tzdata python3 py3-pip

# Copy binary and migrations
COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations

# Install edge-tts for TTS
RUN pip3 install --no-cache-dir edge-tts --break-system-packages

# Create uploads directory
RUN mkdir -p /app/uploads/firmwares /app/uploads/tts

# Expose port
EXPOSE 8080

# Run
CMD ["./server"]
