package commands

import (
	"testing"
)

func TestShowUsage(t *testing.T) {
	// This test just ensures ShowUsage doesn't panic
	// In a real scenario, you might want to capture output
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ShowUsage() panicked: %v", r)
		}
	}()
	
	ShowUsage()
}
