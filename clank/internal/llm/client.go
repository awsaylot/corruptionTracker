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
	// Use a longer timeout for LLM requests, or fall back to 2 minutes if not configured
	timeout := cfg.LLM.Timeout
	if timeout <= 6000*time.Second {
		timeout = 10 * time.Minute
		log.Printf("[LLM] Using extended timeout of %v for LLM requests", timeout)
	}

	return &Client{
		url:     cfg.LLM.URL,
		model:   cfg.LLM.Model,
		timeout: timeout,
		http: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:          10,
				IdleConnTimeout:       6000 * time.Second,
				DisableCompression:    false,
				ResponseHeaderTimeout: 6000 * time.Second,
			},
		},
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
	// Apply chat template
	formattedMessages, err := c.applyChatTemplate(messages)
	if err != nil {
		return fmt.Errorf("failed to apply chat template: %w", err)
	}

	log.Printf("[GenerateStream] Sending request to llama.cpp at %s", c.url)

	llmReq := GenerateRequest{
		Model:    c.model,
		Messages: formattedMessages,
		Stream:   true,
	}

	jsonBody, err := json.Marshal(llmReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.url+"/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("llama.cpp returned status %d: %s", resp.StatusCode, string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	var buffer strings.Builder
	const flushInterval = 100 // chars

	// Helper function to safely send to channel
	safeSend := func(content string) error {
		select {
		case responseChan <- content:
			return nil
		case <-ctx.Done():
			log.Printf("[GenerateStream] context canceled during send: %v", ctx.Err())
			return ctx.Err()
		}
	}

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			log.Printf("[GenerateStream] context canceled: %v", ctx.Err())
			return ctx.Err()
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			if data == "[DONE]" {
				log.Printf("[GenerateStream] stream ended: [DONE]")
				break
			}

			var streamResp SSEResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				// Handle non-JSON data (sometimes llama.cpp sends raw text)
				if data != "" {
					buffer.WriteString(data)
					if buffer.Len() >= flushInterval {
						if err := safeSend(buffer.String()); err != nil {
							return err
						}
						buffer.Reset()
					}
				}
				continue
			}

			if len(streamResp.Choices) > 0 {
				choice := streamResp.Choices[0]
				if choice.Delta.Content != "" {
					buffer.WriteString(choice.Delta.Content)
					if buffer.Len() >= flushInterval {
						if err := safeSend(buffer.String()); err != nil {
							return err
						}
						buffer.Reset()
					}
				}

				if choice.FinishReason != "" && choice.FinishReason != "null" {
					log.Printf("[GenerateStream] stream finished: %s", choice.FinishReason)
					break
				}
			}
		}
	}

	// Flush remaining content
	if buffer.Len() > 0 {
		if err := safeSend(buffer.String()); err != nil {
			return err
		}
		buffer.Reset()
	}

	if err := scanner.Err(); err != nil {
		// Don't treat context cancellation as an error
		if ctx.Err() != nil {
			log.Printf("[GenerateStream] context canceled, scanner stopped: %v", ctx.Err())
			return ctx.Err()
		}
		log.Printf("[GenerateStream] scanner error: %v", err)
		return fmt.Errorf("error reading from llama.cpp stream: %w", err)
	}

	log.Printf("[GenerateStream] successfully processed stream")
	return nil
}

// applyChatTemplate applies the main chat template to messages
func (c *Client) applyChatTemplate(messages []Message) ([]Message, error) {
	// Insert system message handling / role alternation checks here
	// For simplicity, this could just return messages as-is if system message is already handled
	return messages, nil
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
