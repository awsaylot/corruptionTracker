package db

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

var (
	driver    neo4j.Driver
	available bool
	mu        sync.RWMutex
)

// IsAvailable returns true if the Neo4j database is available
func IsAvailable() bool {
	mu.RLock()
	defer mu.RUnlock()
	return available && driver != nil
}

// SetAvailable sets the availability status of the Neo4j database
func setAvailable(status bool) {
	mu.Lock()
	defer mu.Unlock()
	available = status
	if !status {
		log.Printf("Neo4j database is now unavailable")
	}
}

// InitDB initializes a Neo4j database connection with retry logic
func InitDB(uri, username, password string) error {
	var err error
	maxRetries := 3
	retryDelay := 5 * time.Second

	for i := 0; i < maxRetries; i++ {
		log.Printf("Attempting to connect to Neo4j (attempt %d/%d)...", i+1, maxRetries)

		driver, err = neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, ""),
			func(config *neo4j.Config) {
				config.MaxConnectionLifetime = 30 * time.Minute
				config.MaxConnectionPoolSize = 50
				config.Log = neo4j.ConsoleLogger(neo4j.INFO)
				config.ConnectionAcquisitionTimeout = 5 * time.Second
			})

		if err != nil {
			log.Printf("Failed to create Neo4j driver: %v", err)
			if i < maxRetries-1 {
				time.Sleep(retryDelay)
				continue
			}
			setAvailable(false)
			return fmt.Errorf("failed to create Neo4j driver after %d attempts: %v", maxRetries, err)
		}

		// Verify connection
		err = driver.VerifyConnectivity()
		if err != nil {
			log.Printf("Failed to connect to Neo4j: %v", err)
			if i < maxRetries-1 {
				time.Sleep(retryDelay)
				continue
			}
			setAvailable(false)
			return fmt.Errorf("failed to connect to Neo4j after %d attempts: %v", maxRetries, err)
		}

		setAvailable(true)
		log.Printf("Successfully connected to Neo4j database")
		return nil
	}

	return fmt.Errorf("failed to initialize Neo4j connection")
}

// GetDriver returns the Neo4j driver instance
func GetDriver() neo4j.Driver {
	mu.RLock()
	defer mu.RUnlock()
	return driver
}

// CloseDB closes the Neo4j database connection
func CloseDB() error {
	mu.Lock()
	defer mu.Unlock()

	if driver != nil {
		err := driver.Close()
		if err != nil {
			return fmt.Errorf("failed to close Neo4j connection: %v", err)
		}
		available = false
		driver = nil
	}
	return nil
}

// withDatabase executes a function with database error handling
func withDatabase(f func() (interface{}, error)) (interface{}, error) {
	if !IsAvailable() {
		return nil, fmt.Errorf("database is not available - ensure Neo4j is running and properly configured")
	}

	result, err := f()
	if err != nil {
		// Check if error is due to connection issues
		if neo4j.IsConnectivityError(err) {
			setAvailable(false)
			return nil, fmt.Errorf("database connection lost: %v", err)
		}
		return nil, err
	}

	return result, nil
}

// ExecuteRead executes a read transaction with the given work function
func ExecuteRead(work func(tx neo4j.Transaction) (interface{}, error)) (interface{}, error) {
	return withDatabase(func() (interface{}, error) {
		session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
		defer session.Close()

		result, err := session.ReadTransaction(work)
		if err != nil {
			return nil, fmt.Errorf("read transaction failed: %v", err)
		}

		return result, nil
	})
}

// ExecuteWrite executes a write transaction with the given work function
func ExecuteWrite(work func(tx neo4j.Transaction) (interface{}, error)) (interface{}, error) {
	return withDatabase(func() (interface{}, error) {
		session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
		defer session.Close()

		result, err := session.WriteTransaction(work)
		if err != nil {
			return nil, fmt.Errorf("write transaction failed: %v", err)
		}

		return result, nil
	})
}

// TryReconnect attempts to reconnect to the Neo4j database
func TryReconnect(uri, username, password string) error {
	if IsAvailable() {
		return nil
	}

	log.Printf("Attempting to reconnect to Neo4j database...")
	return InitDB(uri, username, password)
}
