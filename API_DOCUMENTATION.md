# Corruption Tracker API Documentation

## Base URL
```
http://localhost:8080/api
```

## Core Endpoints

### Health Check
- **GET** `/health`
  - Description: Check if the API is running
  - Response: 200 OK with "Hello" message

### WebSocket
- **GET** `/ws`
  - Description: WebSocket endpoint for LLM interactions
  - Response: WebSocket connection

### MCP Stream
- **GET** `/mcp/sse`
  - Description: Server-Sent Events endpoint for Model Context Protocol
  - Response: SSE stream

## Node Operations

### Get All Nodes
- **GET** `/nodes`
  - Description: Retrieve all nodes in the database
  - Response: 200 OK
  ```json
  [
    {
      "id": "string",
      "type": "string",
      "props": {
        "name": "string",
        "created_at": "timestamp",
        "updated_at": "timestamp",
        ...
      }
    }
  ]
  ```

### Create Node
- **POST** `/node`
  - Description: Create a new node
  - Request Body:
    ```json
    {
      "type": "string",
      "props": {
        "name": "string",
        ...
      }
    }
    ```
  - Response: 201 Created
  ```json
  {
    "id": "string",
    "type": "string",
    "props": {
      "name": "string",
      "created_at": "timestamp",
      "updated_at": "timestamp",
      ...
    }
  }
  ```

### Get Node
- **GET** `/node/:id`
  - Description: Get a specific node by ID
  - Parameters:
    - `id`: Node ID
  - Response: 200 OK
  ```json
  {
    "id": "string",
    "type": "string",
    "props": {
      "name": "string",
      "created_at": "timestamp",
      "updated_at": "timestamp",
      ...
    }
  }
  ```

### Update Node
- **PUT** `/node/:id`
  - Description: Update a node by ID
  - Parameters:
    - `id`: Node ID
  - Request Body:
    ```json
    {
      "props": {
        "name": "string",
        ...
      }
    }
    ```
  - Response: 200 OK

### Delete Node
- **DELETE** `/node/:id`
  - Description: Delete a node by ID
  - Parameters:
    - `id`: Node ID
  - Response: 204 No Content

### Search Nodes
- **GET** `/search`
  - Description: Search nodes based on properties
  - Query Parameters:
    - `q`: Search query string
    - `type`: (Optional) Node type filter
  - Response: 200 OK
  ```json
  [
    {
      "id": "string",
      "type": "string",
      "props": {...}
    }
  ]
  ```

### Get Network
- **GET** `/network`
  - Description: Get entire graph network with connections
  - Response: 200 OK
  ```json
  [
    {
      "id": "string",
      "type": "string",
      "properties": {...},
      "connections": [
        {
          "id": "string",
          "type": "string",
          "properties": {...},
          "relationship": {
            "type": "string",
            "properties": {...},
            "direction": "in|out"
          }
        }
      ]
    }
  ]
  ```

## Batch Operations

### Batch Create Nodes
- **POST** `/nodes/batch`
  - Description: Create multiple nodes in a single transaction
  - Request Body:
    ```json
    [
      {
        "type": "string",
        "props": {...}
      }
    ]
    ```
  - Response: 201 Created

### Batch Delete Nodes
- **DELETE** `/nodes/batch`
  - Description: Delete multiple nodes in a single transaction
  - Request Body:
    ```json
    ["id1", "id2", ...]
    ```
  - Response: 204 No Content

## Relationship Operations

### Create Relationship
- **POST** `/relationship`
  - Description: Create a new relationship between nodes
  - Request Body:
    ```json
    {
      "fromId": "string",
      "toId": "string",
      "type": "string",
      "props": {...}
    }
    ```
  - Response: 201 Created

### Get Relationship
- **GET** `/relationship/:id`
  - Description: Get a specific relationship by ID
  - Parameters:
    - `id`: Relationship ID
  - Response: 200 OK

### Update Relationship
- **PUT** `/relationship/:id`
  - Description: Update a relationship by ID
  - Parameters:
    - `id`: Relationship ID
  - Request Body:
    ```json
    {
      "props": {...}
    }
    ```
  - Response: 200 OK

### Delete Relationship
- **DELETE** `/relationship/:id`
  - Description: Delete a relationship by ID
  - Parameters:
    - `id`: Relationship ID
  - Response: 204 No Content

## Analytics Endpoints

### Get Corruption Score
- **GET** `/analytics/corruption-score/:nodeId`
  - Description: Calculate corruption score for a node
  - Parameters:
    - `nodeId`: Node ID
  - Response: 200 OK
  ```json
  {
    "name": "string",
    "relationships": [
      {
        "type": "string",
        "count": number
      }
    ],
    "corruptionScore": number
  }
  ```

### Get Entity Connections
- **GET** `/analytics/entity-connections/:nodeId`
  - Description: Analyze connections between entities
  - Parameters:
    - `nodeId`: Node ID
    - `depth` (query, optional): Search depth (default: 3)
  - Response: 200 OK
  ```json
  [
    {
      "name": "string",
      "type": "string",
      "relationTypes": ["string"],
      "connectedTypes": ["string"],
      "connectionCount": number
    }
  ]
  ```

### Get Timeline
- **GET** `/analytics/timeline`
  - Description: Generate timeline of events
  - Response: 200 OK
  ```json
  [
    {
      "date": "timestamp",
      "eventType": "string",
      "source": "string",
      "target": "string",
      "amount": number
    }
  ]
  ```

### Get Network Statistics
- **GET** `/analytics/network-stats`
  - Description: Get network statistics
  - Response: 200 OK
  ```json
  {
    "totalNodes": number,
    "totalRelationships": number,
    "nodeTypes": ["string"],
    "relationshipTypes": ["string"]
  }
  ```

## Graph Operations

### Get Shortest Path
- **GET** `/path`
  - Description: Find shortest path between nodes
  - Query Parameters:
    - `from`: Source node ID
    - `to`: Target node ID
  - Response: 200 OK

### Get Subgraph
- **GET** `/subgraph/:nodeId`
  - Description: Get subgraph centered on a node
  - Parameters:
    - `nodeId`: Node ID
  - Response: 200 OK

## Error Responses

All endpoints may return the following error responses:

- **400 Bad Request**
  ```json
  {
    "error": "Error message describing the issue"
  }
  ```

- **404 Not Found**
  ```json
  {
    "error": "Resource not found"
  }
  ```

- **500 Internal Server Error**
  ```json
  {
    "error": "Internal server error message"
  }
  ```
