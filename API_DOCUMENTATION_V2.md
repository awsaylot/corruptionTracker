# Corruption Tracker API Documentation v2

## Overview
The Corruption Tracker API provides a comprehensive set of endpoints for analyzing and tracking corruption-related information from news articles and other sources. The API includes capabilities for article extraction, graph operations, analytics, and LLM-powered analysis.

## Base URL
All API endpoints are prefixed with `/api` except for WebSocket, health check, and direct LLM endpoints.

## Authentication
Currently supports CORS with the following headers:
- Origin: *
- Methods: GET, POST, PUT, DELETE, OPTIONS
- Headers: Origin, Content-Type, Authorization

## Core Endpoints

### Health Check
```http
GET /health
```
Checks the API service health status.

## Article Analysis

### URL Extraction
```http
POST /api/extraction/url
```
Extracts and analyzes articles from URLs with sequential analysis capabilities.

**Request Body:**
```json
{
  "url": "string",
  "depth": number (2-10, optional)
}
```

**Response:**
```json
{
  "sessionId": "string",
  "article": {
    "id": "string",
    "url": "string",
    "title": "string",
    "content": "string",
    "source": "string",
    "author": "string",
    "publishDate": "datetime",
    "extractedAt": "datetime",
    "metadata": {}
  },
  "session": {
    // Analysis session details
  }
}
```

## Graph Operations

### Nodes

#### Get All Nodes
```http
GET /api/nodes
```

#### Create Node
```http
POST /api/node
```

#### Get Node
```http
GET /api/node/:id
```

#### Update Node
```http
PUT /api/node/:id
```

#### Delete Node
```http
DELETE /api/node/:id
```

#### Search Nodes
```http
GET /api/search
```

#### Get Network
```http
GET /api/network
```

### Batch Operations

#### Batch Create Nodes
```http
POST /api/nodes/batch
```

#### Batch Delete Nodes
```http
DELETE /api/nodes/batch
```

### Graph Analysis

#### Get Shortest Path
```http
GET /api/path
```

#### Get Subgraph
```http
GET /api/subgraph/:nodeId
```

### Relationships

#### Create Relationship
```http
POST /api/relationship
```

#### Get Relationship
```http
GET /api/relationship/:id
```

#### Update Relationship
```http
PUT /api/relationship/:id
```

#### Delete Relationship
```http
DELETE /api/relationship/:id
```

## Analytics

### Corruption Score
```http
GET /api/analytics/corruption-score/:nodeId
```
Calculates and returns corruption score for a specific entity.

### Entity Connections
```http
GET /api/analytics/entity-connections/:nodeId
```
Returns detailed connection analysis for an entity.

### Timeline
```http
GET /api/analytics/timeline
```
Provides temporal analysis of corruption-related events.

### Network Statistics
```http
GET /api/analytics/network-stats
```
Returns statistical analysis of the corruption network.

## LLM Integration

### WebSocket Chat
```http
GET /ws
```
Real-time chat interface with LLM capabilities.

### Direct LLM Endpoints

#### Chat
```http
POST /llm/chat
```
Direct chat interface with llama.cpp.

#### Streaming Chat
```http
POST /llm/chat/stream
```
Streaming chat interface with llama.cpp.

### MCP Enhanced Chat

#### MCP Chat
```http
POST /llm/chat/mcp
```
Enhanced chat with tools and context awareness.

#### MCP SSE
```http
GET /llm/mcp/sse
```
Server-Sent Events for MCP interactions.

## Prompt Management

### List Prompts
```http
GET /api/prompts/
```
Lists all available prompts.

### Get Prompt
```http
GET /api/prompts/:name
```
Retrieves details for a specific prompt.

### Render Prompt
```http
POST /api/prompts/:name/render
```
Renders a prompt with provided arguments.

### Validate Prompt
```http
POST /api/prompts/:name/validate
```
Validates arguments for a prompt.

### Reload Prompts
```http
POST /api/prompts/reload
```
Hot-reloads all prompts.

## Tools

### Run Tool
```http
POST /api/run-tool
```
Executes tools with graph context and LLM integration.

## Data Models

### Article
```json
{
  "id": "string",
  "url": "string",
  "title": "string",
  "content": "string",
  "source": "string",
  "author": "string",
  "publishDate": "datetime",
  "extractedAt": "datetime",
  "metadata": {}
}
```

### ExtractedEntity
```json
{
  "id": "string",
  "type": "string",
  "name": "string",
  "properties": {},
  "confidence": "number",
  "mentions": [
    {
      "text": "string",
      "context": "string",
      "position": {
        "start": "number",
        "end": "number"
      }
    }
  ],
  "articleId": "string",
  "extractedAt": "datetime"
}
```

### ExtractedRelationship
```json
{
  "id": "string",
  "type": "string",
  "fromId": "string",
  "toId": "string",
  "properties": {},
  "confidence": "number",
  "context": "string"
}
```

## Error Handling
All endpoints return appropriate HTTP status codes:
- 200: Success
- 400: Bad Request
- 401: Unauthorized
- 404: Not Found
- 500: Internal Server Error

Error responses include a message field with details about the error.

## Rate Limiting
Currently implemented through middleware with configurable limits.

## Notes
- The API uses Neo4j as its primary database
- Sequential analysis system supports configurable depth (2-10 levels)
- LLM integration provides both synchronous and streaming responses
- WebSocket support for real-time updates
- Hot-reloading support for prompt templates
