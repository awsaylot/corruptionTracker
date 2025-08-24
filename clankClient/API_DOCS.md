# ClankClient HTTP Server API Documentation

The ClankClient now runs as a persistent HTTP server that your main API can communicate with to access Clank MCP functionality.

## Server Setup

### Starting the Server
```bash
# Build and run
go build -o clankClient ./cmd/clankClient
./clankClient serve [port]

# Or use the convenience script
./start_server.bat  # Windows
```

Default port: `8081`

### Configuration
- The server connects to the Clank MCP Server at `http://localhost:8080` (configurable in `internal/config/config.go`)
- The HTTP server provides REST endpoints to interact with the MCP server
- Includes CORS headers for web client compatibility

## API Endpoints

All responses follow this format:
```json
{
  "success": true,
  "data": { ... },
  "error": ""
}
```

### Health & Info Endpoints

#### `GET /health`
Health check endpoint
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "service": "clank-client-daemon", 
    "version": "1.0.0"
  }
}
```

#### `GET /api/info`
Server connection information
```json
{
  "success": true,
  "data": {
    "message": "Connected to Clank MCP Server",
    "server": "http://localhost:8080",
    "version": "1.0.0"
  }
}
```

### Discovery Endpoints

#### `GET /api/tools`
List all available tools from the MCP server
```json
{
  "success": true,
  "data": {
    "tools": [
      {
        "name": "clank:system_info",
        "description": "Get system information",
        "inputSchema": { ... }
      },
      // ... more tools
    ]
  }
}
```

#### `GET /api/resources`
List all available resources
```json
{
  "success": true,
  "data": {
    "resources": [
      {
        "uri": "resource://example",
        "name": "Example Resource",
        "description": "An example resource",
        "mimeType": "text/plain"
      }
      // ... more resources
    ]
  }
}
```

#### `GET /api/prompts`
List all available prompts
```json
{
  "success": true,
  "data": {
    "prompts": [
      {
        "name": "example_prompt",
        "description": "An example prompt",
        "arguments": []
      }
      // ... more prompts
    ]
  }
}
```

### Action Endpoints

#### `POST /api/tools/call`
Call a specific tool
```json
// Request
{
  "tool_name": "clank:system_info",
  "arguments": {
    "type": "cwd"
  }
}

// Response
{
  "success": true,
  "data": {
    "content": [
      {
        "type": "text",
        "text": "/path/to/current/directory"
      }
    ],
    "isError": false
  }
}
```

#### `POST /api/resources/get`
Get a specific resource
```json
// Request
{
  "resource_uri": "resource://example"
}

// Response
{
  "success": true,
  "data": {
    "contents": [
      {
        "uri": "resource://example",
        "mimeType": "text/plain",
        "text": "Resource content here"
      }
    ]
  }
}
```

#### `POST /api/prompts/get`
Get a specific prompt
```json
// Request
{
  "prompt_name": "example_prompt",
  "arguments": {
    "param1": "value1"
  }
}

// Response
{
  "success": true,
  "data": {
    "description": "Example prompt description",
    "messages": [
      {
        "role": "user",
        "content": {
          "type": "text",
          "text": "Prompt content with param1: value1"
        }
      }
    ]
  }
}
```

## Example Usage from Your Main API

### Node.js/Express Example
```javascript
const axios = require('axios');

const CLANK_CLIENT_URL = 'http://localhost:8081';

// Call a tool
async function callClankTool(toolName, arguments) {
  try {
    const response = await axios.post(`${CLANK_CLIENT_URL}/api/tools/call`, {
      tool_name: toolName,
      arguments: arguments
    });
    
    if (response.data.success) {
      return response.data.data;
    } else {
      throw new Error(response.data.error);
    }
  } catch (error) {
    console.error('Error calling Clank tool:', error);
    throw error;
  }
}

// Usage
const result = await callClankTool('clank:file_operations', {
  operation: 'list',
  path: './documents'
});
```

### Python Example
```python
import requests
import json

CLANK_CLIENT_URL = 'http://localhost:8081'

def call_clank_tool(tool_name, arguments=None):
    url = f'{CLANK_CLIENT_URL}/api/tools/call'
    payload = {
        'tool_name': tool_name,
        'arguments': arguments or {}
    }
    
    response = requests.post(url, json=payload)
    data = response.json()
    
    if data['success']:
        return data['data']
    else:
        raise Exception(data['error'])

# Usage
result = call_clank_tool('clank:system_info', {'type': 'cwd'})
```

## Error Handling

All endpoints return appropriate HTTP status codes:
- `200 OK` - Success
- `500 Internal Server Error` - Error occurred

Error responses include details in the `error` field:
```json
{
  "success": false,
  "data": null,
  "error": "Failed to call tool: tool not found"
}
```

## Testing

Use the provided test scripts to verify the server is working:
```bash
# Windows
./test_server.bat

# Linux/Mac
./test_server.sh
```

Or test individual endpoints with curl:
```bash
curl http://localhost:8081/health
curl http://localhost:8081/api/tools
curl -X POST http://localhost:8081/api/tools/call \
  -H "Content-Type: application/json" \
  -d '{"tool_name": "clank:system_info", "arguments": {"type": "cwd"}}'
```
