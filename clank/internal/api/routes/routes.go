package routes

import (
	"fmt"
	"net/http"

	"clank/config"
	"clank/internal/api/handlers"
)

// SetupRoutes configures all API routes
func SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	// Create handlers
	extractionHandler := handlers.NewExtractionHandler(cfg)

	// Register routes
	mux.HandleFunc("/api/v1/extract", extractionHandler.HandleURLExtraction)
	mux.HandleFunc("/api/v1/health", handlers.HandleHealthCheck)

	return mux
}
