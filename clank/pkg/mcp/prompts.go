package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"clank/pkg/models"

	mcp "github.com/mark3labs/mcp-go/mcp"          // MCP prompt functions
	mcpserver "github.com/mark3labs/mcp-go/server" // alias MCPServer
)

type PromptManager struct {
	prompts map[string]models.Prompt
	mu      sync.RWMutex
	server  *mcpserver.MCPServer
}

// NewPromptManager creates a new PromptManager
func NewPromptManager(server *mcpserver.MCPServer) *PromptManager {
	return &PromptManager{
		prompts: make(map[string]models.Prompt),
		server:  server,
	}
}

// LoadPrompts loads all JSON prompt files from a directory
func (pm *PromptManager) LoadPrompts(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Failed to read prompt file %s: %v", filePath, err)
			continue
		}

		var prompt models.Prompt
		if err := json.Unmarshal(data, &prompt); err != nil {
			log.Printf("Failed to unmarshal prompt from %s: %v", filePath, err)
			continue
		}

		if prompt.Name == "" {
			prompt.Name = entry.Name()[:len(entry.Name())-len(filepath.Ext(entry.Name()))]
		}

		pm.RegisterPrompt(prompt.Name, prompt.Text)
		log.Printf("Loaded prompt: %s", prompt.Name)
	}

	return nil
}

// RegisterPrompt adds a prompt to the MCP server
func (pm *PromptManager) RegisterPrompt(name, text string) {
	pm.mu.Lock()
	pm.prompts[name] = models.Prompt{Name: name, Text: text}
	pm.mu.Unlock()

	prompt := mcp.NewPrompt(name,
		mcp.WithPromptDescription(fmt.Sprintf("Prompt: %s", name)),
		mcp.WithArgument("input", mcp.ArgumentDescription("Input text to process with this prompt")),
	)

	handler := func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		input := request.Params.Arguments["input"]
		promptText := text
		if input != "" {
			promptText = fmt.Sprintf("%s\n\nInput: %s", text, input)
		}

		return mcp.NewGetPromptResult(
			fmt.Sprintf("Prompt: %s", name),
			[]mcp.PromptMessage{
				mcp.NewPromptMessage(
					mcp.RoleUser,
					mcp.NewTextContent(promptText),
				),
			},
		), nil
	}

	pm.server.AddPrompt(prompt, handler)
}

// GetPrompts returns all registered prompts
func (pm *PromptManager) GetPrompts() []models.Prompt {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make([]models.Prompt, 0, len(pm.prompts))
	for _, p := range pm.prompts {
		result = append(result, p)
	}
	return result
}
