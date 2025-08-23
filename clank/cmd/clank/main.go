package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"clank/pkg/mcp" // Adjust this import path to match your module
)

func main() {
	// CRITICAL: Set up logging to stderr so it appears in Claude Desktop logs
	log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("[INFO] Starting Clank MCP Server...")

	// Toggle here: set true for HTTP mode, false for stdio mode
	setModeHTTP := true

	// Honor the toggle by setting env vars accordingly
	if setModeHTTP {
		os.Setenv("MCP_USE_STDIO", "false")
		if os.Getenv("MCP_HTTP_ADDR") == "" {
			os.Setenv("MCP_HTTP_ADDR", ":8080")
		}
	} else {
		os.Setenv("MCP_USE_STDIO", "true")
	}

	log.Printf("[INFO] MCP_USE_STDIO: %s", os.Getenv("MCP_USE_STDIO"))
	log.Printf("[INFO] MCP_HTTP_ADDR: %s", os.Getenv("MCP_HTTP_ADDR"))

	// Initialize the MCP server
	server := mcp.NewServer("clank", "1.0.0")

	// Optionally, load additional resources manually via getter
	if err := server.ResourceManager().LoadJSON("sample", "sample.json"); err != nil {
		log.Printf("Warning: failed to load sample.json: %v", err)
	}

	// Optionally, list all loaded prompts
	for _, prompt := range server.PromptManager().GetPrompts() {
		log.Printf("Registered prompt: %s", prompt.Name)
	}

	// Optionally, list registered tools
	for _, name := range server.ToolManager().List() {
		log.Printf("Registered tool: %s", name)
	}

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Printf("[INFO] Received signal: %v", sig)
		cancel()
	}()

	// Start the MCP server
	log.Println("[INFO] MCP server starting...")
	if err := server.Serve(ctx); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}

	log.Println("[INFO] MCP server shut down cleanly")
}
