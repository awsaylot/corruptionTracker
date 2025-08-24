#!/bin/bash

# Build script for clankClient
set -e

echo "Building clankClient..."

# Clean previous builds
rm -rf bin/

# Create bin directory
mkdir -p bin

# Build for current platform
echo "Building for current platform..."
go build -o bin/clankClient ./cmd/clankClient

# Build for multiple platforms if requested
if [ "$1" = "all" ]; then
    echo "Building for multiple platforms..."
    
    # Windows
    GOOS=windows GOARCH=amd64 go build -o bin/clankClient-windows-amd64.exe ./cmd/clankClient
    
    # macOS
    GOOS=darwin GOARCH=amd64 go build -o bin/clankClient-darwin-amd64 ./cmd/clankClient
    GOOS=darwin GOARCH=arm64 go build -o bin/clankClient-darwin-arm64 ./cmd/clankClient
    
    # Linux
    GOOS=linux GOARCH=amd64 go build -o bin/clankClient-linux-amd64 ./cmd/clankClient
    GOOS=linux GOARCH=arm64 go build -o bin/clankClient-linux-arm64 ./cmd/clankClient
    
    echo "Cross-platform builds completed!"
fi

echo "Build completed! Binaries are in the 'bin' directory."
