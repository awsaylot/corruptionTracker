package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"clank/config"
)

// Client represents an LLM client
type Client struct {
	url     string
	model   string
	timeout time.Duration
	http    *http.Client
}

// NewClient creates a new LLM client with the given configuration
func NewClient(cfg *config.Config) *Client {
	return &Client{
		url:     cfg.LLM.URL,
		model:   cfg.LLM.Model,
		timeout: cfg.LLM.Timeout,
		http:    &http.Client{Timeout: cfg.LLM.Timeout},
	}
}

// GenerateRequest represents a request to the LLM API
type GenerateRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GenerateResponse represents a response from the LLM API
type GenerateResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

// SSEResponse represents a streaming response chunk from the LLM API
type SSEResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// GenerateStream sends a request to llama.cpp and streams the response chunks through the provided channel
func (c *Client) GenerateStream(ctx context.Context, messages []Message, responseChan chan<- string) error {
	defer close(responseChan)

	log.Printf("Sending request to llama.cpp at %s", c.url)

	// Create HTTP request for llama.cpp's OpenAI-compatible API
	llmReq := GenerateRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   true,
	}

	jsonBody, err := json.Marshal(llmReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("Request payload: %s", string(jsonBody))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.url+"/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request to llama.cpp: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("llama.cpp response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		// Read error response body for debugging
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("llama.cpp returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read the streaming response from llama.cpp
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		log.Printf("Received line from llama.cpp: %s", line)

		// Handle SSE format
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// Check for end of stream
			if data == "[DONE]" {
				log.Printf("llama.cpp stream completed")
				break
			}

			// Try to parse as JSON (OpenAI format)
			var streamResp SSEResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				log.Printf("Failed to parse JSON from llama.cpp, treating as plain text: %v", err)
				// If not JSON, treat as plain text (fallback)
				if data != "" {
					select {
					case responseChan <- data:
					case <-ctx.Done():
						return ctx.Err()
					}
				}
				continue
			}

			// Extract content from OpenAI-compatible response
			if len(streamResp.Choices) > 0 {
				choice := streamResp.Choices[0]
				if choice.Delta.Content != "" {
					log.Printf("Sending chunk to frontend: %s", choice.Delta.Content)
					select {
					case responseChan <- choice.Delta.Content:
					case <-ctx.Done():
						return ctx.Err()
					}
				}

				// Check if stream is finished
				if choice.FinishReason != "" {
					log.Printf("llama.cpp stream finished: %s", choice.FinishReason)
					break
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading from llama.cpp stream: %w", err)
	}

	log.Printf("Successfully processed llama.cpp stream")
	return nil
}

// Generate generates a non-streaming response from llama.cpp
func (c *Client) Generate(ctx context.Context, messages []Message) (*GenerateResponse, error) {
	reqBody := GenerateRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	log.Printf("Sending non-streaming request to llama.cpp: %s", string(jsonBody))

	req, err := http.NewRequestWithContext(ctx, "POST", c.url+"/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to llama.cpp: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read error response body
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("llama.cpp returned status %d: %s", resp.StatusCode, string(body))
	}

	var result GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding llama.cpp response: %w", err)
	}

	log.Printf("Successfully got response from llama.cpp")
	return &result, nil
}
