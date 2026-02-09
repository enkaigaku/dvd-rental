# Build stage
FROM golang:1.25.7-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/payment-service ./cmd/payment-service

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install certificates for SSL/TLS connections
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/payment-service .

# Create a non-root user
RUN adduser -D -g '' appuser
USER appuser

# Expose gRPC port
EXPOSE 50055

# Run the binary
CMD ["./payment-service"]
