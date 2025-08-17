package tools

import (
	"context"
	"testing"
	"time"
)

func TestBrowserAutomation(t *testing.T) {
	ctx := context.Background()

	browser, err := NewBrowserAutomation()
	if err != nil {
		t.Fatalf("Failed to create browser automation: %v", err)
	}

	err = browser.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize browser: %v", err)
	}
	defer browser.Close()

	t.Log("Starting navigation test...")

	// Test navigation to a simple local test page
	err = browser.NavigateTo(ctx, "about:blank")
	if err != nil {
		t.Fatalf("Failed to navigate to blank page: %v", err)
	}
	t.Log("Navigation to blank page successful")

	// Now try Google
	err = browser.NavigateTo(ctx, "https://www.google.com")
	if err != nil {
		t.Fatalf("Failed to navigate to Google: %v", err)
	}
	t.Log("Navigation to Google successful")

	// Add a longer wait to see what's happening
	time.Sleep(5 * time.Second)
	t.Log("Waited 5 seconds...")

	// Test waiting for and getting text from Google's search input
	t.Log("Waiting for search input...")
	err = browser.WaitForSelector(ctx, "textarea[name=q]") // Google sometimes uses textarea instead of input
	if err != nil {
		t.Fatalf("Failed to wait for selector: %v", err)
	}
	t.Log("Found search input") // Test screenshot
	err = browser.Screenshot(ctx, "example_screenshot.png")
	if err != nil {
		t.Fatalf("Failed to take screenshot: %v", err)
	}
}
