package main

import (
	"clank/config"
	"clank/internal/api/routes"
	"clank/internal/db"
	"log"
)

func main() {
	cfg := config.LoadConfig()

	// Initialize Neo4j connection
	if err := db.InitDB(cfg.Neo4j.URI, cfg.Neo4j.Username, cfg.Neo4j.Password); err != nil {
		log.Fatalf("Failed to initialize Neo4j: %v", err)
	}
	defer db.CloseDB()

	r := routes.SetupRouter()
	log.Printf("Server running on %s", cfg.Server.Address)
	r.Run(cfg.Server.Address)
}
