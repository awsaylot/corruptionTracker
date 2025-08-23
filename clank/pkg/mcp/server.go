package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	mcp "github.com/mark3labs/mcp-go/mcp"          // for Tool/Prompt types
	mcpserver "github.com/mark3labs/mcp-go/server" // for MCPServer type and transports
)

type Server struct {
	mcpServer     *mcpserver.MCPServer
	resources     *ResourceManager
	promptManager *PromptManager
	toolManager   *ToolManager
	mu            sync.RWMutex
}

// NewServer initializes the MCP server with resources, prompts, and tools
func NewServer(name, version string) *Server {
	log.Printf("[DEBUG] Creating new MCP server instance: %s v%s", name, version)

	// Create the MCP server with proper capabilities
	mcpSrv := mcpserver.NewMCPServer(name, version)

	s := &Server{
		mcpServer: mcpSrv,
		resources: NewResourceManager("./"),
	}

	// Initialize managers
	s.promptManager = NewPromptManager(s.mcpServer)
	s.toolManager = NewToolManager(s)

	// Register built-in tools first (these should be fast and always available)
	log.Println("[DEBUG] Registering built-in tools...")
	s.registerBuiltinTools()

	// Load resources synchronously but with timeout
	log.Println("[DEBUG] Loading resources...")
	if err := s.loadResourcesWithTimeout(5 * time.Second); err != nil {
		log.Printf("[WARN] Failed to load resources: %v", err)
	} else {
		log.Println("[DEBUG] Resource loading complete")
	}

	// Load prompts synchronously but with timeout
	log.Println("[DEBUG] Loading prompts...")
	if err := s.loadPromptsWithTimeout(5 * time.Second); err != nil {
		log.Printf("[WARN] Failed to load prompts: %v", err)
	} else {
		log.Println("[DEBUG] Prompts loaded successfully")
	}

	// Load dynamic tools synchronously but with timeout
	log.Println("[DEBUG] Loading dynamic tools...")
	if err := s.loadDynamicToolsWithTimeout(5 * time.Second); err != nil {
		log.Printf("[WARN] Failed to load dynamic tools: %v", err)
	} else {
		log.Println("[DEBUG] Dynamic tools loaded successfully")
	}

	log.Printf("[DEBUG] MCP server '%s' initialized", name)
	return s
}

// ResourceManager returns the server's ResourceManager
func (s *Server) ResourceManager() *ResourceManager {
	return s.resources
}

// PromptManager returns the server's PromptManager
func (s *Server) PromptManager() *PromptManager {
	return s.promptManager
}

// ToolManager exposes tool-related operations
func (s *Server) ToolManager() *ToolManager {
	return s.toolManager
}

// loadResourcesWithTimeout loads resources with a timeout
func (s *Server) loadResourcesWithTimeout(timeout time.Duration) error {
	done := make(chan error, 1)

	go func() {
		done <- s.loadResources()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		log.Printf("[WARN] Resource loading timed out after %v", timeout)
		return nil // Don't fail initialization, just log warning
	}
}

// loadPromptsWithTimeout loads prompts with a timeout
func (s *Server) loadPromptsWithTimeout(timeout time.Duration) error {
	done := make(chan error, 1)

	go func() {
		done <- s.promptManager.LoadPrompts("./prompts")
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		log.Printf("[WARN] Prompt loading timed out after %v", timeout)
		return nil // Don't fail initialization, just log warning
	}
}

// loadDynamicToolsWithTimeout loads dynamic tools with a timeout
func (s *Server) loadDynamicToolsWithTimeout(timeout time.Duration) error {
	done := make(chan error, 1)

	go func() {
		done <- s.loadDynamicTools("./tools")
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		log.Printf("[WARN] Dynamic tool loading timed out after %v", timeout)
		return nil // Don't fail initialization, just log warning
	}
}

// loadResources walks through the resources directory and loads all JSON files
func (s *Server) loadResources() error {
	resourcesDir := "./resources"
	if _, err := os.Stat(resourcesDir); os.IsNotExist(err) {
		log.Printf("[DEBUG] Resources directory does not exist, creating: %s", resourcesDir)
		if err := os.MkdirAll(resourcesDir, 0755); err != nil {
			log.Printf("[WARN] Failed to create resources directory: %v", err)
		}
		return nil
	}

	log.Printf("[DEBUG] Walking resources directory: %s", resourcesDir)
	return filepath.Walk(resourcesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("[WARN] Error accessing path '%s': %v", path, err)
			return nil
		}
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		rel, _ := filepath.Rel(resourcesDir, path)
		name := rel[:len(rel)-len(filepath.Ext(rel))]
		log.Printf("[DEBUG] Loading resource '%s' from file '%s'", name, path)
		if err := s.resources.LoadJSON(name, rel); err != nil {
			log.Printf("[WARN] Failed to load JSON resource '%s': %v", name, err)
		}
		return nil
	})
}

// loadDynamicTools loads tool definitions from JSON files
func (s *Server) loadDynamicTools(toolsDir string) error {
	if _, err := os.Stat(toolsDir); os.IsNotExist(err) {
		log.Printf("[DEBUG] Tools directory does not exist, creating: %s", toolsDir)
		if err := os.MkdirAll(toolsDir, 0755); err != nil {
			log.Printf("[WARN] Failed to create tools directory: %v", err)
		}
		return nil
	}

	return filepath.Walk(toolsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("[WARN] Error accessing tool path '%s': %v", path, err)
			return nil
		}
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		log.Printf("[DEBUG] Loading dynamic tool from: %s", path)
		if err := s.toolManager.LoadToolFromJSON(path); err != nil {
			log.Printf("[WARN] Failed to load tool from '%s': %v", path, err)
		}
		return nil
	})
}

// Serve starts the MCP server
func (s *Server) Serve(ctx context.Context) error {
	httpAddr := os.Getenv("MCP_HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8080"
	}
	useStdio := os.Getenv("MCP_USE_STDIO") == "true"

	log.Printf("[DEBUG] Starting Clank MCP server (HTTP %s/mcp), MCP_USE_STDIO=%v", httpAddr, useStdio)

	if useStdio {
		log.Println("[DEBUG] Starting stdio transport")
		return mcpserver.ServeStdio(s.mcpServer)
	}

	// HTTP transport
	httpSrv := mcpserver.NewStreamableHTTPServer(s.mcpServer)

	errCh := make(chan error, 1)
	go func() {
		log.Printf("[DEBUG] HTTP transport starting on %s", httpAddr)
		if err := httpSrv.Start(httpAddr); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("[DEBUG] Context cancelled, shutting down...")
	case err := <-errCh:
		if err != nil {
			log.Printf("[ERROR] HTTP transport error: %v", err)
			return err
		}
	}

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[WARN] Error shutting down HTTP transport: %v", err)
	} else {
		log.Println("[DEBUG] HTTP transport shutdown complete")
	}

	log.Println("[DEBUG] MCP server shutdown complete")
	return nil
}

// --- System info handlers ---

func (s *Server) handleGetCWD() (*mcp.CallToolResult, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error getting current directory: %v", err)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Current directory: %s", filepath.Clean(cwd))), nil
}

func (s *Server) handleGetTime() (*mcp.CallToolResult, error) {
	now := time.Now()
	return mcp.NewToolResultText(fmt.Sprintf("Current time: %s", now.Format(time.RFC3339))), nil
}

func (s *Server) handleGetEnvVar(varName string) (*mcp.CallToolResult, error) {
	if varName == "" {
		return mcp.NewToolResultError("environment variable name required"), nil
	}
	value := os.Getenv(varName)
	if value == "" {
		return mcp.NewToolResultText(fmt.Sprintf("Environment variable %s not set", varName)), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("%s=%s", varName, value)), nil
}

// --- Tool registration ---

func (s *Server) registerBuiltinTools() {
	s.registerSystemInfoTool()
	s.registerResourceTool()
}

func (s *Server) registerSystemInfoTool() {
	tool := mcp.NewTool("system_info",
		mcp.WithDescription("Get system information and environment details"),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("Type of system info: cwd, time, or env_var"),
		),
		mcp.WithString("var_name",
			mcp.Description("Environment variable name (for env_var type)"),
		),
	)

	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		infoType, err := request.RequireString("type")
		if err != nil {
			return mcp.NewToolResultError("type must be a string"), nil
		}

		switch infoType {
		case "cwd":
			return s.handleGetCWD()
		case "time":
			return s.handleGetTime()
		case "env_var":
			varName := request.GetString("var_name", "")
			return s.handleGetEnvVar(varName)
		default:
			return mcp.NewToolResultError("invalid type. Use: cwd, time, or env_var"), nil
		}
	}

	s.toolManager.AddTool("system_info", tool, handler)
}

func (s *Server) registerResourceTool() {
	tool := mcp.NewTool("resource",
		mcp.WithDescription("Fetch a JSON resource by name"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Resource name to fetch"),
		),
	)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, _ := req.RequireString("name")
		res, ok := s.resources.Get(name)
		if !ok {
			return mcp.NewToolResultError(fmt.Sprintf("resource %s not found", name)), nil
		}
		bytes, _ := json.MarshalIndent(res, "", "  ")
		return mcp.NewToolResultText(string(bytes)), nil
	}

	s.toolManager.AddTool("resource", tool, handler)
}
