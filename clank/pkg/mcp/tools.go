package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	mcp "github.com/mark3labs/mcp-go/mcp"
)

// ToolManager manages tool registration and execution
type ToolManager struct {
	server *Server
	tools  map[string]struct{}
	mu     sync.RWMutex
}

// NewToolManager creates a new ToolManager
func NewToolManager(server *Server) *ToolManager {
	return &ToolManager{
		server: server,
		tools:  make(map[string]struct{}),
	}
}

// DynamicToolDef represents a tool definition loaded from JSON
type DynamicToolDef struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Arguments   map[string]ToolArgDef `json:"arguments"`
	Handler     string                `json:"handler,omitempty"`
}

// ToolArgDef represents an argument definition
type ToolArgDef struct {
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

// AddTool registers a tool with the MCP server
func (tm *ToolManager) AddTool(name string, tool mcp.Tool, handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)) {
	log.Printf("[DEBUG] ToolManager: Adding tool '%s'", name)
	tm.server.mcpServer.AddTool(tool, handler)

	tm.mu.Lock()
	tm.tools[name] = struct{}{}
	tm.mu.Unlock()

	log.Printf("[DEBUG] Tool '%s' registered successfully", name)
}

// LoadToolFromJSON loads a tool definition from a JSON file
func (tm *ToolManager) LoadToolFromJSON(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read tool file %s: %w", filePath, err)
	}

	var toolDef DynamicToolDef
	if err := json.Unmarshal(data, &toolDef); err != nil {
		return fmt.Errorf("failed to unmarshal tool definition from %s: %w", filePath, err)
	}

	if toolDef.Name == "" {
		return fmt.Errorf("tool name is required in %s", filePath)
	}

	// Build MCP tool from definition
	toolOptions := []mcp.ToolOption{mcp.WithDescription(toolDef.Description)}

	for argName, argDef := range toolDef.Arguments {
		var propOptions []mcp.PropertyOption
		if argDef.Description != "" {
			propOptions = append(propOptions, mcp.Description(argDef.Description))
		}
		if len(argDef.Enum) > 0 {
			propOptions = append(propOptions, mcp.Enum(argDef.Enum...))
		}
		if argDef.Required {
			propOptions = append(propOptions, mcp.Required())
		}

		switch argDef.Type {
		case "string":
			toolOptions = append(toolOptions, mcp.WithString(argName, propOptions...))
		case "number":
			toolOptions = append(toolOptions, mcp.WithNumber(argName, propOptions...))
		case "boolean":
			toolOptions = append(toolOptions, mcp.WithBoolean(argName, propOptions...))
		default:
			log.Printf("[WARN] Unknown argument type '%s' for argument '%s' in tool '%s'", argDef.Type, argName, toolDef.Name)
			toolOptions = append(toolOptions, mcp.WithString(argName, propOptions...))
		}
	}

	tool := mcp.NewTool(toolDef.Name, toolOptions...)

	handler := tm.createDynamicHandler(toolDef)
	tm.AddTool(toolDef.Name, tool, handler)

	log.Printf("[INFO] Successfully loaded dynamic tool: %s", toolDef.Name)
	return nil
}

// createDynamicHandler creates a handler function for dynamic tools
func (tm *ToolManager) createDynamicHandler(toolDef DynamicToolDef) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		log.Printf("[DEBUG] Executing dynamic tool: %s", toolDef.Name)

		switch toolDef.Handler {
		case "browser_automation":
			return BrowserAutomationHandler(ctx, request)
		case "file_operations":
			return FileOperationsHandler(ctx, request)
		default:
			// Generic fallback: echo parameters
			params := make(map[string]interface{})
			if request.Params.Arguments != nil {
				if args, ok := request.Params.Arguments.(map[string]any); ok {
					for argName := range toolDef.Arguments {
						if value, exists := args[argName]; exists {
							params[argName] = value
						}
					}
				}
			}
			result := fmt.Sprintf("Dynamic tool '%s' executed with parameters: %+v", toolDef.Name, params)
			return mcp.NewToolResultText(result), nil
		}
	}
}

// List returns the list of registered tool names
func (tm *ToolManager) List() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	out := make([]string, 0, len(tm.tools))
	for name := range tm.tools {
		out = append(out, name)
	}
	log.Printf("[DEBUG] ToolManager: Listing tools (%d total)", len(out))
	return out
}

// Has checks whether a tool with the given name is registered
func (tm *ToolManager) Has(name string) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	_, ok := tm.tools[name]
	log.Printf("[DEBUG] ToolManager: Has('%s') = %v", name, ok)
	return ok
}
