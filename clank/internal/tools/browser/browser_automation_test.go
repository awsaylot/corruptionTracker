package browser

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

	// Test getting text from Google's search input
	t.Log("Checking for search input...")
	_, err = browser.GetElementTextBySelector(ctx, "textarea[name=q]") // Google sometimes uses textarea instead of input
	if err != nil {
		t.Fatalf("Failed to get search input: %v", err)
	}
	t.Log("Found search input")
}
