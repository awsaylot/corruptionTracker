# clankClient

A Go-based MCP (Model Context Protocol) client for interacting with the Clank server.

## Project Structure

This project follows the [Go Standard Project Layout](https://github.com/golang-standards/project-layout):

```
clankClient/
├── cmd/
│   └── clankClient/          # Main application entry point
├── internal/                 # Private application code
│   ├── client/              # MCP client implementation
│   ├── commands/            # CLI command handlers
│   └── config/              # Configuration constants
├── pkg/                     # Public library code (empty for now)
├── bin/                     # Built binaries (created during build)
├── build.sh                 # Unix build script
├── build.bat                # Windows build script
├── Makefile                 # Make commands for development
├── go.mod                   # Go module definition
└── README.md               # This file
```

## Features

- Connect to Clank MCP server over HTTP transport
- List available tools, resources, and prompts
- Call tools with arguments
- Read resources from the server
- Execute prompts with parameters
- Comprehensive demo mode to test all capabilities

## Building

### Using Make (Recommended)

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Clean build artifacts
make clean

# Run tests
make test

# Install dependencies
make install
```

### Using Build Scripts

**Unix/Linux/macOS:**
```bash
# Build for current platform
./build.sh

# Build for all platforms
./build.sh all
```

**Windows:**
```cmd
# Build for current platform
build.bat

# Build for all platforms
build.bat all
```

### Using Go Directly

```bash
# Build for current platform
go build -o bin/clankClient ./cmd/clankClient

# Run without building
go run ./cmd/clankClient
```

## Usage

```bash
# Show help
./bin/clankClient

# Show server information
./bin/clankClient info

# List available tools
./bin/clankClient tools

# List available resources
./bin/clankClient resources

# List available prompts
./bin/clankClient prompts

# Call a specific tool
./bin/clankClient call clank:system_info

# Get a specific resource
./bin/clankClient get resource://example

# Execute a specific prompt
./bin/clankClient prompt example-prompt

# Run comprehensive demo
./bin/clankClient demo
```

## Configuration

The client connects to `http://localhost:8080` by default. This can be configured by modifying the `DefaultServerURL` constant in `internal/config/config.go`.

## Dependencies

- [mcp-go](https://github.com/mark3labs/mcp-go) - MCP implementation for Go
- Go 1.24.3 or later

## Architecture

The application is structured into several packages:

- **cmd/clankClient**: Main application entry point
- **internal/client**: Core MCP client functionality
- **internal/commands**: CLI command parsing and execution
- **internal/config**: Configuration constants and settings

This structure provides clean separation of concerns and makes the codebase easier to maintain and extend.
