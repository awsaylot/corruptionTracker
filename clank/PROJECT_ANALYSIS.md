# Clank Project Analysis

## Overview

Clank is an MCP (Model Context Protocol) server implementation written in Go that provides a bridge between AI assistants (like Claude) and system tools. The project appears to be part of a larger "corruptionTracker" system with multiple components.

## Project Structure

### Core Components

#### 1. **MCP Server (`pkg/mcp/server.go`)**
- Main server implementation that manages tools, prompts, and resources
- Supports both HTTP and stdio transport modes
- Includes timeout mechanisms for resource loading (5 seconds for each component)
- Built-in graceful shutdown handling

#### 2. **Tool Management (`pkg/mcp/tools.go`)**
- Dynamic tool loading from JSON configurations
- Built-in tools: file operations, system info, resource management
- Support for custom handlers (currently supports browser automation)
- Thread-safe tool registration and management

#### 3. **Browser Automation (`pkg/mcp/browser_tool.go`)**
- Selenium WebDriver integration for Firefox automation
- Intelligent article extraction using CSS selectors
- Automatic Selenium server management
- Content saving capabilities

#### 4. **Resource Management (`pkg/mcp/resources.go`)**
- JSON resource loading and caching
- Thread-safe resource access
- Support for empty file handling

#### 5. **Prompt Management (`pkg/mcp/prompts.go`)**
- Dynamic prompt loading from JSON files
- Integration with MCP prompt system
- Template-based prompt processing

## Key Features

### Built-in Tools
1. **File Operations**
   - Read, write, and list files/directories
   - Cross-platform file system access

2. **System Information**
   - Current working directory
   - System time
   - Environment variable access

3. **Resource Management**
   - JSON resource fetching by name
   - Cached resource access

4. **Browser Automation**
   - Firefox browser control via Selenium
   - Intelligent article text extraction
   - Full HTML page source extraction
   - Content saving to files

### Configuration
- HTTP server mode on port 8080 by default
- Selenium integration with automatic server startup
- Firefox binary path: `C:/Program Files/Mozilla Firefox/firefox.exe`
- Selenium JAR path: `C:/selenium/selenium-server-4.35.0.jar`

### Directory Structure
```
clank/
├── cmd/clank/main.go          # Entry point
├── pkg/
│   ├── mcp/                   # MCP server implementation
│   │   ├── server.go         # Main server
│   │   ├── tools.go          # Tool management
│   │   ├── browser_tool.go   # Browser automation
│   │   ├── resources.go      # Resource management
│   │   └── prompts.go        # Prompt management
│   └── models/
│       └── prompt.go         # Prompt data model
├── tools/
│   └── browser_automation.json # Tool configuration
├── prompts/                   # (Empty) Prompt definitions
├── resources/
│   ├── sample.json           # (Empty) Sample resource
│   └── templates/            # Resource templates
├── go.mod                    # Go module definition
├── go.sum                    # Dependency checksums
└── main.exe                  # Compiled binary
```

## Dependencies

### Core Dependencies
- `github.com/mark3labs/mcp-go v0.38.0` - MCP protocol implementation
- `github.com/tebeka/selenium v0.9.9` - Selenium WebDriver bindings

### Transitive Dependencies
- UUID generation, JSON schema validation, YAML parsing
- Generic data structures and utilities

## Parent Project Context

The Clank component is part of a larger "corruptionTracker" project that includes:

### Planned Components (commented out in docker-compose.yml)
1. **Frontend** - Next.js application on port 3000
2. **Neo4j Database** - Graph database for data storage
3. **Main Clank Service** - The current MCP server

### Active Components
1. **LLM Service** - Llama.cpp server with CUDA support
   - Uses CodeLlama-7b-Instruct model
   - GPU-accelerated inference
   - Serves on port 8090

## Current State

### Implemented Features
- ✅ MCP server with HTTP/stdio modes
- ✅ File system operations
- ✅ System information access
- ✅ Browser automation with Firefox
- ✅ Dynamic tool loading
- ✅ Resource and prompt management
- ✅ Graceful shutdown
- ✅ Timeout handling for initialization

### Missing/Incomplete Features
- 🔄 Empty prompts directory (no custom prompts loaded)
- 🔄 Empty sample.json resource
- 🔄 Frontend integration (commented out)
- 🔄 Database integration (commented out)
- ⚠️ Hardcoded Windows paths for Firefox and Selenium

## Technical Architecture

### Design Patterns
1. **Manager Pattern** - Separate managers for tools, prompts, and resources
2. **Factory Pattern** - Dynamic tool creation from JSON configurations
3. **Observer Pattern** - Graceful shutdown handling with context cancellation

### Concurrency
- Extensive use of mutexes for thread safety
- Context-based cancellation for graceful shutdown
- Timeout mechanisms to prevent hanging during initialization

### Error Handling
- Comprehensive error wrapping and logging
- Non-fatal warnings for missing resources/prompts
- Graceful degradation when components fail to load

## Potential Use Cases

Based on the project structure and name "corruptionTracker", this appears to be:

1. **Investigation/Research Tool** - Using browser automation to gather information
2. **Content Analysis System** - Processing web content and documents
3. **AI-Assisted Research Platform** - Integrating LLM capabilities with web scraping
4. **Data Collection Framework** - Systematic information gathering and storage

## Recommendations

### Immediate Improvements
1. **Cross-platform Support** - Make Firefox/Selenium paths configurable
2. **Configuration Management** - Add config files instead of hardcoded values
3. **Example Content** - Add sample prompts and resources for demonstration
4. **Documentation** - Add README and API documentation

### Architecture Enhancements
1. **Plugin System** - More extensible tool loading mechanism
2. **Database Integration** - Implement the Neo4j component
3. **Frontend Development** - Build the web interface
4. **Authentication** - Add security for HTTP mode

### Operational Improvements
1. **Docker Integration** - Complete the containerization
2. **Health Checks** - Add monitoring endpoints
3. **Metrics** - Add performance and usage metrics
4. **Testing** - Add comprehensive test suite

## Security Considerations

- Browser automation could be used maliciously
- File system access requires careful permission management
- HTTP mode needs authentication in production
- Resource loading should validate file paths to prevent directory traversal

## Conclusion

Clank is a well-structured MCP server that successfully bridges AI assistants with system tools and browser automation. The code quality is high with proper error handling, concurrency safety, and modular design. The project shows potential for being part of a larger research or investigation platform, though several components remain incomplete or commented out in the current state.
