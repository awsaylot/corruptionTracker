# Build stage
FROM golang:1.24.3-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o /clank ./cmd/server

# Final stage
FROM alpine:3.18

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy the binary from builder
COPY --from=builder /clank .
# Copy config directory
COPY --from=builder /app/config ./config

# Create sandbox directory
RUN mkdir -p /sandbox

# Set environment variables
ENV GIN_MODE=release

EXPOSE 8080

# Run the application
CMD ["./clank"]
