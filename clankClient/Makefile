# Makefile for clankClient

.PHONY: build clean test run install help

# Build the application
build:
	@echo "Building clankClient..."
	@mkdir -p bin
	@go build -o bin/clankClient ./cmd/clankClient

# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p bin
	@GOOS=windows GOARCH=amd64 go build -o bin/clankClient-windows-amd64.exe ./cmd/clankClient
	@GOOS=darwin GOARCH=amd64 go build -o bin/clankClient-darwin-amd64 ./cmd/clankClient
	@GOOS=darwin GOARCH=arm64 go build -o bin/clankClient-darwin-arm64 ./cmd/clankClient
	@GOOS=linux GOARCH=amd64 go build -o bin/clankClient-linux-amd64 ./cmd/clankClient
	@GOOS=linux GOARCH=arm64 go build -o bin/clankClient-linux-arm64 ./cmd/clankClient
	@echo "Cross-platform builds completed!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run the application
run:
	@go run ./cmd/clankClient

# Install dependencies
install:
	@echo "Installing dependencies..."
	@go mod tidy

# Show help
help:
	@echo "Available commands:"
	@echo "  build      - Build the application"
	@echo "  build-all  - Build for all platforms"
	@echo "  clean      - Clean build artifacts"
	@echo "  test       - Run tests"
	@echo "  run        - Run the application"
	@echo "  install    - Install dependencies"
	@echo "  help       - Show this help message"
