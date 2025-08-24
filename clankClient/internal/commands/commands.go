package commands

import (
	"fmt"
	"log"
	"os"
	"time"

	"clankClient/internal/client"
	"clankClient/internal/config"
	"clankClient/internal/server"
)

func ShowUsage() {
	fmt.Printf("Clank MCP Client v%s\n", config.ClientVersion)
	fmt.Println("\nUsage:")
	fmt.Println("  clankClient <command> [args...]")
	fmt.Println("\nCommands:")
	fmt.Println("  serve [port]   - Start HTTP server daemon (default port: 8081)")
	fmt.Println("  info           - Show server information and capabilities")
	fmt.Println("  tools          - List available tools")
	fmt.Println("  resources      - List available resources") 
	fmt.Println("  prompts        - List available prompts")
	fmt.Println("  call <tool>    - Call a specific tool (interactive)")
	fmt.Println("  get <resource> - Get a specific resource")
	fmt.Println("  prompt <n>     - Get a specific prompt")
	fmt.Println("  demo           - Run a comprehensive demonstration")
}
func Execute() error {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	// Parse command line arguments
	if len(os.Args) < 2 {
		ShowUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Handle serve command separately since it doesn't need the traditional client setup
	if command == "serve" {
		return executeServe()
	}

	// Create client for other commands
	clankClient, err := client.New(config.DefaultServerURL)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Connect to server
	if err := clankClient.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer clankClient.Disconnect()

	// Execute command
	switch command {
	case "info":
		return executeInfo()
		
	case "tools":
		return clankClient.ListTools()
		
	case "resources":
		return clankClient.ListResources()
		
	case "prompts":
		return clankClient.ListPrompts()
		
	case "call":
		return executeCall(clankClient)
		
	case "get":
		return executeGet(clankClient)
		
	case "prompt":
		return executePrompt(clankClient)
		
	case "demo":
		return executeDemo(clankClient)
		
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func executeServe() error {
	port := "8081" // default port
	if len(os.Args) >= 3 {
		port = os.Args[2]
	}
	
	srv := server.New(port)
	return srv.Start()
}
		return executeDemo(clankClient)
		
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func executeInfo() error {
	fmt.Println("✅ Connected successfully! See logs above for server details.")
	return nil
}

func executeCall(c *client.ClankClient) error {
	if len(os.Args) < 3 {
		return fmt.Errorf("please specify a tool name")
	}
	toolName := os.Args[2]
	
	// For demonstration, call with minimal arguments
	args := make(map[string]interface{})
	if toolName == "clank:system_info" {
		args["type"] = "cwd"
	} else if toolName == "clank:file_operations" {
		args["operation"] = "list"
		args["path"] = "."
	}
	
	return c.CallTool(toolName, args)
}

func executeGet(c *client.ClankClient) error {
	if len(os.Args) < 3 {
		return fmt.Errorf("please specify a resource URI")
	}
	resourceURI := os.Args[2]
	return c.GetResource(resourceURI)
}

func executePrompt(c *client.ClankClient) error {
	if len(os.Args) < 3 {
		return fmt.Errorf("please specify a prompt name")
	}
	promptName := os.Args[2]
	return c.GetPrompt(promptName, nil)
}

func executeDemo(c *client.ClankClient) error {
	fmt.Println("🚀 Running comprehensive demo...")
	time.Sleep(1 * time.Second)
	
	// List all capabilities
	if err := c.ListTools(); err != nil {
		log.Printf("Error listing tools: %v", err)
	}
	
	time.Sleep(500 * time.Millisecond)
	
	if err := c.ListResources(); err != nil {
		log.Printf("Error listing resources: %v", err)
	}
	
	time.Sleep(500 * time.Millisecond)
	
	if err := c.ListPrompts(); err != nil {
		log.Printf("Error listing prompts: %v", err)
	}
	
	time.Sleep(500 * time.Millisecond)
	
	// Try calling system info tool
	fmt.Println("\n🔧 Testing system info tool...")
	if err := c.CallTool("clank:system_info", map[string]interface{}{
		"type": "cwd",
	}); err != nil {
		log.Printf("Error calling system_info tool: %v", err)
	}
	
	time.Sleep(500 * time.Millisecond)
	
	// Try calling file operations tool
	fmt.Println("\n📁 Testing file operations tool...")
	if err := c.CallTool("clank:file_operations", map[string]interface{}{
		"operation": "list",
		"path":      ".",
	}); err != nil {
		log.Printf("Error calling file_operations tool: %v", err)
	}
	
	fmt.Println("\n✅ Demo completed!")
	return nil
}
