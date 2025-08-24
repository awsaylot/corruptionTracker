package main

import (
	"fmt"
	"log"
	
	"clankClient/internal/client"
)

// Simple test to verify our client can be instantiated
func main() {
	fmt.Println("Testing clankClient instantiation...")
	
	// Try to create a client
	c, err := client.New("http://localhost:8080")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	
	fmt.Println("Client created successfully!")
	
	// Don't actually connect since the server might not be running
	fmt.Println("Test completed successfully!")
}
