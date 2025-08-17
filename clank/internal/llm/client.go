package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

func NewClient(cfg *config.Config) *Client {
	return &Client{
		url:     cfg.LLM.URL,
		model:   cfg.LLM.Model,
		timeout: cfg.LLM.Timeout,
		// No Timeout here so streaming isn't cut off; rely on ctx for cancellation.
		http: &http.Client{},
	}
}

type GenerateRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GenerateResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

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

// GenerateStream sends a request to llama.cpp and streams chunks into responseChan.
// IMPORTANT: this function **does not** close responseChan. The caller owns closing it.
func (c *Client) GenerateStream(ctx context.Context, messages []Message, responseChan chan<- string) error {
	llmReq := GenerateRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   true,
	}

	jsonBody, err := json.Marshal(llmReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.url+"/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to llama.cpp: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("llama.cpp returned status %d: %s", resp.StatusCode, string(body))
	}

	// Safe send helper: avoids panic if caller closed the channel.
	trySend := func(s string) error {
		defer func() {
			_ = recover() // swallow "send on closed channel"
		}()
		select {
		case responseChan <- s:
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil
	}

	scanner := bufio.NewScanner(resp.Body)
	// Increase max token size to handle larger SSE lines safely.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// End of stream marker
		if data == "[DONE]" {
			break
		}

		// Parse OpenAI-compatible SSE chunk
		var sse SSEResponse
		if err := json.Unmarshal([]byte(data), &sse); err != nil {
			// Fallback: forward raw data
			if data != "" {
				if err := trySend(data); err != nil {
					return err
				}
			}
			continue
		}

		if len(sse.Choices) > 0 {
			ch := sse.Choices[0]
			if ch.Delta.Content != "" {
				if err := trySend(ch.Delta.Content); err != nil {
					return err
				}
			}
			// Some servers set FinishReason in a mid-stream record—ignore and rely on [DONE]
		}
	}

	if err := scanner.Err(); err != nil {
		// If ctx ended, prefer ctx error for clarity
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("error reading from llama.cpp stream: %w", err)
	}

	return nil
}

// Generate performs a standard (non-streaming) completion.
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
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("llama.cpp returned status %d: %s", resp.StatusCode, string(body))
	}

	var result GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding llama.cpp response: %w", err)
	}

	return &result, nil
}
