package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"clankClient/internal/config"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type ClankClient struct {
	client *client.Client
	ctx    context.Context
}

func New(serverURL string) (*ClankClient, error) {
	if serverURL == "" {
		serverURL = config.DefaultServerURL
	}

	// Create streamable HTTP client (the recommended HTTP transport)
	mcpClient, err := client.NewStreamableHttpClient(serverURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	return &ClankClient{
		client: mcpClient,
		ctx:    context.Background(),
	}, nil
}

func (c *ClankClient) Connect() error {
	log.Printf("Connecting to Clank MCP server at %s...", config.DefaultServerURL)
	
	// Start the connection
	if err := c.client.Start(c.ctx); err != nil {
		return fmt.Errorf("failed to start connection: %w", err)
	}

	// Initialize the connection - start with minimal structure
	initResult, err := c.client.Initialize(c.ctx, mcp.InitializeRequest{})
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	log.Printf("Connected successfully! Server: %s v%s", 
		initResult.ServerInfo.Name, 
		initResult.ServerInfo.Version)
	
	// Log server info structure to understand what's available
	log.Printf("Server info: %+v", initResult.ServerInfo)

	return nil
}

func (c *ClankClient) ListTools() error {
	log.Println("Fetching available tools...")
	
	result, err := c.client.ListTools(c.ctx, mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	fmt.Printf("\n=== Available Tools (%d) ===\n", len(result.Tools))
	for _, tool := range result.Tools {
		fmt.Printf("• %s\n", tool.Name)
		if tool.Description != "" {
			fmt.Printf("  Description: %s\n", tool.Description)
		}
		
		// Pretty print the input schema
		if schemaBytes, err := json.MarshalIndent(tool.InputSchema, "  ", "  "); err == nil {
			fmt.Printf("  Schema: %s\n", string(schemaBytes))
		}
		fmt.Println()
	}

	return nil
}

func (c *ClankClient) ListResources() error {
	log.Println("Fetching available resources...")
	
	result, err := c.client.ListResources(c.ctx, mcp.ListResourcesRequest{})
	if err != nil {
		return fmt.Errorf("failed to list resources: %w", err)
	}

	fmt.Printf("\n=== Available Resources (%d) ===\n", len(result.Resources))
	for _, resource := range result.Resources {
		fmt.Printf("• %s\n", resource.URI)
		if resource.Name != "" {
			fmt.Printf("  Name: %s\n", resource.Name)
		}
		if resource.Description != "" {
			fmt.Printf("  Description: %s\n", resource.Description)
		}
		if resource.MIMEType != "" {
			fmt.Printf("  MIME Type: %s\n", resource.MIMEType)
		}
		fmt.Println()
	}

	return nil
}

func (c *ClankClient) ListPrompts() error {
	log.Println("Fetching available prompts...")
	
	result, err := c.client.ListPrompts(c.ctx, mcp.ListPromptsRequest{})
	if err != nil {
		return fmt.Errorf("failed to list prompts: %w", err)
	}

	fmt.Printf("\n=== Available Prompts (%d) ===\n", len(result.Prompts))
	for _, prompt := range result.Prompts {
		fmt.Printf("• %s\n", prompt.Name)
		if prompt.Description != "" {
			fmt.Printf("  Description: %s\n", prompt.Description)
		}
		if len(prompt.Arguments) > 0 {
			fmt.Printf("  Arguments: %v\n", prompt.Arguments)
		}
		fmt.Println()
	}

	return nil
}

func (c *ClankClient) CallTool(toolName string, arguments map[string]interface{}) error {
	log.Printf("Calling tool: %s", toolName)
	
	result, err := c.client.CallTool(c.ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: arguments,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to call tool %s: %w", toolName, err)
	}

	fmt.Printf("\n=== Tool Result: %s ===\n", toolName)
	for i, content := range result.Content {
		fmt.Printf("Content %d:\n", i+1)
		// Handle content based on its actual structure
		if contentBytes, err := json.MarshalIndent(content, "  ", "  "); err == nil {
			fmt.Printf("  Data: %s\n", string(contentBytes))
		}
	}
	
	if result.IsError {
		fmt.Printf("⚠️  Tool returned an error\n")
	}

	return nil
}

func (c *ClankClient) GetResource(resourceURI string) error {
	log.Printf("Reading resource: %s", resourceURI)
	
	result, err := c.client.ReadResource(c.ctx, mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: resourceURI,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to read resource %s: %w", resourceURI, err)
	}

	fmt.Printf("\n=== Resource Content: %s ===\n", resourceURI)
	for i, content := range result.Contents {
		fmt.Printf("Content %d:\n", i+1)
		// Handle content based on its actual structure
		if contentBytes, err := json.MarshalIndent(content, "  ", "  "); err == nil {
			fmt.Printf("  Data: %s\n", string(contentBytes))
		}
	}

	return nil
}

func (c *ClankClient) GetPrompt(promptName string, arguments map[string]string) error {
	log.Printf("Getting prompt: %s", promptName)
	
	result, err := c.client.GetPrompt(c.ctx, mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name:      promptName,
			Arguments: arguments,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to get prompt %s: %w", promptName, err)
	}

	fmt.Printf("\n=== Prompt: %s ===\n", promptName)
	if result.Description != "" {
		fmt.Printf("Description: %s\n", result.Description)
	}
	
	for i, message := range result.Messages {
		fmt.Printf("Message %d:\n", i+1)
		fmt.Printf("  Role: %s\n", message.Role)
		
		// Handle message content based on its actual structure
		if contentBytes, err := json.MarshalIndent(message.Content, "    ", "  "); err == nil {
			fmt.Printf("    Content: %s\n", string(contentBytes))
		}
	}

	return nil
}

func (c *ClankClient) Disconnect() error {
	log.Println("Disconnecting from Clank MCP server...")
	return c.client.Close()
}
