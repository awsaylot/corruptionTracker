package browser

import (
	"context"
)

// Browser defines the interface for browser automation
type Browser interface {
	// Navigation
	Navigate(ctx context.Context, url string) error
	WaitForSelector(ctx context.Context, selector string) error

	// Content Extraction
	GetContent(ctx context.Context, selector string) (string, error)
	GetAttribute(ctx context.Context, selector, attribute string) (string, error)
	GetHTML(ctx context.Context, selector string) (string, error)

	// Interaction
	Click(ctx context.Context, selector string) error
	Type(ctx context.Context, selector, text string) error

	// Session Management
	Close() error
	NewPage(ctx context.Context) error
}

// PageState represents the state of a browser page
type PageState struct {
	URL        string
	Title      string
	ReadyState string
	Error      error
}

// Config represents browser automation configuration
type Config struct {
	Headless    bool
	Timeout     int
	UserAgent   string
	Proxy       string
	Credentials *Credentials
}

// Credentials represents authentication credentials
type Credentials struct {
	Username string
	Password string
	Token    string
}
