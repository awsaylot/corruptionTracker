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

// GetPrompt returns a specific prompt by name
func (pm *PromptManager) GetPrompt(name string) (models.Prompt, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	prompt, exists := pm.prompts[name]
	return prompt, exists
}

// GetPromptNames returns a list of all prompt names
func (pm *PromptManager) GetPromptNames() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	names := make([]string, 0, len(pm.prompts))
	for name := range pm.prompts {
		names = append(names, name)
	}
	return names
}

// ExecutePrompt formats a prompt with input data
func (pm *PromptManager) ExecutePrompt(name, input string) (string, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	prompt, exists := pm.prompts[name]
	if !exists {
		return "", fmt.Errorf("prompt '%s' not found", name)
	}
	
	if input == "" {
		return prompt.Text, nil
	}
	
	return fmt.Sprintf("%s\n\nInput: %s", prompt.Text, input), nil
}

// --- Tool Handlers ---

// HandleGetPromptsRequest handles the get_prompts tool request
func (pm *PromptManager) HandleGetPromptsRequest(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Check if a specific prompt name is requested
	promptName := request.GetString("name", "")
	
	if promptName != "" {
		// Return specific prompt
		prompt, exists := pm.GetPrompt(promptName)
		if !exists {
			return mcp.NewToolResultError(fmt.Sprintf("prompt '%s' not found", promptName)), nil
		}
		
		data, err := json.MarshalIndent(prompt, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("error marshaling prompt: %v", err)), nil
		}
		
		return mcp.NewToolResultText(string(data)), nil
	}
	
	// Return all prompts
	prompts := pm.GetPrompts()
	data, err := json.MarshalIndent(prompts, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error marshaling prompts: %v", err)), nil
	}
	
	return mcp.NewToolResultText(string(data)), nil
}

// HandleExecutePromptRequest handles the execute_prompt tool request
func (pm *PromptManager) HandleExecutePromptRequest(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	promptName, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError("prompt name is required"), nil
	}
	
	input := request.GetString("input", "")
	
	executedPrompt, err := pm.ExecutePrompt(promptName, input)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	
	// Return as formatted text
	result := fmt.Sprintf("=== Prompt: %s ===\n\n%s", promptName, executedPrompt)
	return mcp.NewToolResultText(result), nil
}
