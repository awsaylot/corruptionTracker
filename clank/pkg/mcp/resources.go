package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type ResourceManager struct {
	basePath string
	cache    map[string]interface{}
	mu       sync.RWMutex
}

func NewResourceManager(basePath string) *ResourceManager {
	return &ResourceManager{
		basePath: basePath,
		cache:    make(map[string]interface{}),
	}
}

// LoadJSON loads a JSON file into the cache
func (r *ResourceManager) LoadJSON(name, filePath string) error {
	fullPath := filepath.Join(r.basePath, filePath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return err
	}

	// Handle empty files
	if len(data) == 0 {
		r.mu.Lock()
		r.cache[name] = make(map[string]interface{})
		r.mu.Unlock()
		return nil
	}

	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}

	r.mu.Lock()
	r.cache[name] = obj
	r.mu.Unlock()

	return nil
}

// Get retrieves a resource by name
func (r *ResourceManager) Get(name string) (interface{}, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	val, ok := r.cache[name]
	return val, ok
}

// GetJSON retrieves a resource as a map
func (r *ResourceManager) GetJSON(name string) (map[string]interface{}, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if val, ok := r.cache[name]; ok {
		if m, ok := val.(map[string]interface{}); ok {
			return m, true
		}
	}
	return nil, false
}

// List returns all cached resource names
func (r *ResourceManager) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.cache))
	for name := range r.cache {
		names = append(names, name)
	}
	return names
}
