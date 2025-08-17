package config

import (
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Neo4jConfig struct {
	URI      string `yaml:"uri"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Config struct {
	Server struct {
		Address string `yaml:"address"`
	} `yaml:"server"`
	MCP struct {
		ListenPath string `yaml:"listen_path"`
	} `yaml:"mcp"`
	LLM struct {
		URL     string        `yaml:"url"`
		Model   string        `yaml:"model"`
		Timeout time.Duration `yaml:"timeout"`
	} `yaml:"llm"`
	Neo4j Neo4jConfig `yaml:"neo4j"`
}

// LoadConfig loads config from config/config.yaml
func LoadConfig() *Config {
	f, err := os.Open("config/config.yaml")
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
	}
	defer f.Close()

	cfg := &Config{}

	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		log.Fatalf("Error decoding config file: %v", err)
	}
	return cfg
}
