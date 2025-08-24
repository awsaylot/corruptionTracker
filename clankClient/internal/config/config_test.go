package config

import (
	"testing"
	
	"clankClient/internal/version"
)

func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"DefaultServerURL", DefaultServerURL, "http://localhost:8080"},
		{"ClientName", ClientName, version.Name},
		{"ClientVersion", ClientVersion, version.Version},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.value, tt.expected)
			}
		})
	}
}
