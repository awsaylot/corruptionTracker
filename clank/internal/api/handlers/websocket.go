package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"clank/config"
	"clank/internal/llm"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// WebSocketMessage represents a message from the frontend
type WebSocketMessage struct {
	Type     string        `json:"type"`
	Messages []llm.Message `json:"messages,omitempty"`
	Content  string        `json:"content,omitempty"`
}

// WebSocketResponse represents a response to the frontend
type WebSocketResponse struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Done    bool   `json:"done"`
	Error   string `json:"error,omitempty"`
}

func WebSocketHandler(cfg *config.Config) gin.HandlerFunc {
	client := llm.NewClient(cfg)
	// You could also use MCP service here if you want MCP integration in WebSocket
	// mcpService := NewMCPService(cfg)

	return func(c *gin.Context) {
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: %v", err)
			return
		}
		defer ws.Close()

		// Set up ping/pong to keep connection alive
		ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		ws.SetPongHandler(func(string) error {
			ws.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		// Start ping ticker
		ticker := time.NewTicker(54 * time.Second)
		defer ticker.Stop()

		go func() {
			for range ticker.C {
				if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}()

		for {
			var wsMsg WebSocketMessage
			err := ws.ReadJSON(&wsMsg)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				break
			}

			switch wsMsg.Type {
			case "chat":
				handleChatMessage(ws, client, &wsMsg, c.Request.Context())
			case "ping":
				// Respond to ping
				response := WebSocketResponse{Type: "pong"}
				if err := ws.WriteJSON(response); err != nil {
					log.Printf("Error sending pong: %v", err)
					break
				}
			default:
				log.Printf("Unknown message type: %s", wsMsg.Type)
			}
		}
	}
}

func handleChatMessage(ws *websocket.Conn, client *llm.Client, wsMsg *WebSocketMessage, ctx context.Context) {
	var messages []llm.Message

	// Handle different input formats
	if len(wsMsg.Messages) > 0 {
		messages = wsMsg.Messages
	} else if wsMsg.Content != "" {
		messages = []llm.Message{
			{
				Role:    "user",
				Content: wsMsg.Content,
			},
		}
	} else {
		sendError(ws, "No message content provided")
		return
	}

	log.Printf("Processing chat message with %d messages", len(messages))

	// Create response channel for LLM chunks
	responseChan := make(chan string, 100)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Start generating in background
	go func() {
		defer close(responseChan)
		if err := client.GenerateStream(ctx, messages, responseChan); err != nil {
			log.Printf("Error generating: %v", err)
			sendError(ws, "Error generating response: "+err.Error())
			return
		}
	}()

	// Forward chunks to websocket
	for {
		select {
		case chunk, ok := <-responseChan:
			if !ok {
				// Channel closed, send final message
				response := WebSocketResponse{
					Type: "chat_chunk",
					Done: true,
				}
				if err := ws.WriteJSON(response); err != nil {
					log.Printf("Error writing final chunk: %v", err)
				}
				return
			}

			response := WebSocketResponse{
				Type:    "chat_chunk",
				Content: chunk,
				Done:    false,
			}

			if err := ws.WriteJSON(response); err != nil {
				log.Printf("Error writing chunk: %v", err)
				return
			}

		case <-ctx.Done():
			sendError(ws, "Request timeout")
			return
		}
	}
}

func sendError(ws *websocket.Conn, errorMsg string) {
	response := WebSocketResponse{
		Type:  "error",
		Error: errorMsg,
	}
	if err := ws.WriteJSON(response); err != nil {
		log.Printf("Error sending error message: %v", err)
	}
}
