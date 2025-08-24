package main

import (
	"log"

	"clankClient/internal/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		log.Fatalf("Error: %v", err)
	}
	
	log.Println("✅ Operation completed successfully!")
}
