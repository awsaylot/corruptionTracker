package handlers

import (
	"context"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"clank/config"
	"clank/internal/llm"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
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

	return func(c *gin.Context) {
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: %v", err)
			return
		}
		defer func() {
			log.Printf("Closing WebSocket connection")
			ws.Close()
		}()

		ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		ws.SetPongHandler(func(string) error {
			ws.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		ticker := time.NewTicker(54 * time.Second)
		defer ticker.Stop()
		done := make(chan bool)
		defer close(done)

		go func() {
			for {
				select {
				case <-ticker.C:
					if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
						return
					}
				case <-done:
					return
				}
			}
		}()

		for {
			var wsMsg WebSocketMessage
			if err := ws.ReadJSON(&wsMsg); err != nil {
				log.Printf("Error reading message: %v", err)
				break
			}

			switch wsMsg.Type {
			case "chat":
				if err := handleChatMessage(ws, client, &wsMsg, c.Request.Context()); err != nil {
					log.Printf("Error handling chat message: %v", err)
					sendError(ws, "Error processing chat message")
				}
			case "ping":
				response := WebSocketResponse{Type: "pong"}
				ws.WriteJSON(response)
			default:
				log.Printf("Unknown message type: %s", wsMsg.Type)
			}
		}
	}
}

func handleChatMessage(ws *websocket.Conn, client *llm.Client, wsMsg *WebSocketMessage, parentCtx context.Context) error {
	messages := extractMessages(wsMsg)
	if len(messages) == 0 {
		return sendError(ws, "No message content provided")
	}

	responseChan := make(chan string, 100)
	var closeOnce sync.Once
	closeResponseChan := func() { closeOnce.Do(func() { close(responseChan) }) }
	defer closeResponseChan()

	msgCtx, cancel := context.WithTimeout(parentCtx, 5*time.Minute)
	defer cancel()

	go func(msgs []llm.Message) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic in generation goroutine: %v", r)
			}
		}()
		if err := client.GenerateStream(msgCtx, msgs, responseChan); err != nil {
			log.Printf("Error generating: %v", err)
			select {
			case responseChan <- "ERROR: " + err.Error():
			case <-msgCtx.Done():
			}
		}
	}(messages)

	for chunk := range responseChan {
		if strings.HasPrefix(chunk, "ERROR: ") {
			return sendError(ws, strings.TrimPrefix(chunk, "ERROR: "))
		}

		if err := ws.WriteJSON(WebSocketResponse{
			Type:    "chat_chunk",
			Content: chunk,
			Done:    false,
		}); err != nil {
			return err
		}
	}

	return ws.WriteJSON(WebSocketResponse{
		Type: "chat_chunk",
		Done: true,
	})
}

func extractMessages(wsMsg *WebSocketMessage) []llm.Message {
	if len(wsMsg.Messages) > 0 {
		return wsMsg.Messages
	}
	if wsMsg.Content != "" {
		return []llm.Message{{Role: "user", Content: wsMsg.Content}}
	}
	return nil
}

func sendError(ws *websocket.Conn, errorMsg string) error {
	response := WebSocketResponse{
		Type:  "error",
		Error: errorMsg,
	}
	return ws.WriteJSON(response)
}
