package routes

import (
	"clank/config"
	"clank/internal/api/handlers"
	"clank/internal/api/handlers/graph"
	"clank/internal/api/middleware"
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	cfg := config.LoadConfig()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		if strings.HasPrefix(c.Request.URL.Path, "/extractor") {
			// Set longer timeout for extractor endpoints
			ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Minute)
			defer cancel()
			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	})

	// Core endpoints
	r.GET("/health", handlers.HealthHandler)

	// WebSocket for real-time chat with llama.cpp
	r.GET("/ws", handlers.WebSocketHandler(cfg))

	// Article extraction endpoints
	r.POST("/extractor", handlers.ExtractorHandler(cfg))
	r.POST("/extractor/integrate", handlers.ExtractAndIntegrateHandler(cfg))
	r.POST("/extractor/stream", handlers.ExtractorStreamHandler(cfg))

	// Graph integration endpoint (for extracted data)
	r.POST("/graph/integrate", handlers.GraphIntegrationHandler)

	// LLM endpoints (direct HTTP API to llama.cpp)
	llm := r.Group("/llm")
	{
		// Direct chat with llama.cpp
		llm.POST("/chat", handlers.LLMHandler)
		llm.POST("/chat/stream", handlers.LLMStreamHandler(cfg))

		// MCP-enhanced chat (with tools and context)
		llm.POST("/chat/mcp", handlers.MCPChatHandler(cfg))
		llm.GET("/mcp/sse", handlers.MCPHandlerSSE)
	}

	// API routes (your existing graph/database operations)
	api := r.Group("/api")
	api.Use(middleware.RequireDatabase())
	{
		// Graph operations
		api.GET("/nodes", graph.GetAllNodes)
		api.POST("/node", graph.CreateNode)
		api.GET("/node/:id", graph.GetNode)
		api.PUT("/node/:id", graph.UpdateNode)
		api.DELETE("/node/:id", graph.DeleteNode)
		api.GET("/search", graph.SearchNodes)
		api.GET("/network", graph.GetNetwork)

		// Batch operations
		api.POST("/nodes/batch", graph.BatchCreateNodes)
		api.DELETE("/nodes/batch", graph.BatchDeleteNodes)

		// Graph operations
		api.GET("/path", graph.GetShortestPath)
		api.GET("/subgraph/:nodeId", graph.GetSubgraph)

		// Relationship operations
		api.POST("/relationship", graph.CreateRelationship)
		api.GET("/relationship/:id", graph.GetRelationship)
		api.PUT("/relationship/:id", graph.UpdateRelationship)
		api.DELETE("/relationship/:id", graph.DeleteRelationship)

		// Analytics endpoints
		analytics := api.Group("/analytics")
		{
			analytics.GET("/corruption-score/:nodeId", graph.GetCorruptionScoreHandler)
			analytics.GET("/entity-connections/:nodeId", graph.GetEntityConnectionsHandler)
			analytics.GET("/timeline", graph.GetTimelineHandler)
			analytics.GET("/network-stats", graph.GetNetworkStatsHandler)
		}

		// Tools endpoint (enhanced with graph context and LLM)
		api.POST("/run-tool", handlers.ToolHandler)
	}

	return r
}
