package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
)

// BrowserAutomation handles browser automation tasks using Playwright
type BrowserAutomation struct {
	pw      *playwright.Playwright
	browser playwright.Browser
	context playwright.BrowserContext
	page    playwright.Page
}

// NewBrowserAutomation creates a new browser automation instance
func NewBrowserAutomation() (*BrowserAutomation, error) {
	return &BrowserAutomation{}, nil
}

// Initialize sets up the Playwright environment and launches the browser
func (ba *BrowserAutomation) Initialize(ctx context.Context) error {
	var err error

	// Install browsers if needed
	err = playwright.Install()
	if err != nil {
		return fmt.Errorf("could not install playwright dependencies: %v", err)
	}

	// Initialize playwright
	ba.pw, err = playwright.Run()
	if err != nil {
		return fmt.Errorf("could not start playwright: %v", err)
	}

	// Launch the browser in non-headless mode with dev tools
	ba.browser, err = ba.pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
		Devtools: playwright.Bool(true),
		SlowMo:   playwright.Float(100), // Add a 100ms delay between actions for visibility
	})
	if err != nil {
		return fmt.Errorf("could not launch browser: %v", err)
	}

	// Create a new context
	ba.context, err = ba.browser.NewContext()
	if err != nil {
		return fmt.Errorf("could not create browser context: %v", err)
	}

	// Create a new page
	ba.page, err = ba.context.NewPage()
	if err != nil {
		return fmt.Errorf("could not create page: %v", err)
	}

	return nil
}

// NavigateTo navigates to the specified URL
func (ba *BrowserAutomation) NavigateTo(ctx context.Context, url string) error {
	// Set a longer timeout for navigation (60 seconds)
	if _, err := ba.page.Goto(url, playwright.PageGotoOptions{
		Timeout:   playwright.Float(60000),
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		return fmt.Errorf("could not navigate to %s: %v", url, err)
	}
	return nil
}

// Click clicks on an element matching the specified selector
func (ba *BrowserAutomation) Click(ctx context.Context, selector string) error {
	if err := ba.page.Click(selector); err != nil {
		return fmt.Errorf("could not click on %s: %v", selector, err)
	}
	return nil
}

// Type types text into an element matching the specified selector
func (ba *BrowserAutomation) Type(ctx context.Context, selector, text string) error {
	if err := ba.page.Fill(selector, text); err != nil {
		return fmt.Errorf("could not type into %s: %v", selector, err)
	}
	return nil
}

// Screenshot takes a screenshot of the current page
func (ba *BrowserAutomation) Screenshot(ctx context.Context, path string) error {
	_, err := ba.page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(path),
	})
	if err != nil {
		return fmt.Errorf("could not take screenshot: %v", err)
	}
	return nil
}

// WaitForSelector waits for an element matching the specified selector to be present
func (ba *BrowserAutomation) WaitForSelector(ctx context.Context, selector string) error {
	if _, err := ba.page.WaitForSelector(selector, playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(60000), // 60 second timeout
	}); err != nil {
		return fmt.Errorf("could not wait for selector %s: %v", selector, err)
	}

	// Add a small delay to ensure the element is fully interactive
	time.Sleep(1 * time.Second)
	return nil
}

// GetText gets the text content of an element matching the specified selector
func (ba *BrowserAutomation) GetText(ctx context.Context, selector string) (string, error) {
	element, err := ba.page.QuerySelector(selector)
	if err != nil {
		return "", fmt.Errorf("could not find element %s: %v", selector, err)
	}

	text, err := element.TextContent()
	if err != nil {
		return "", fmt.Errorf("could not get text content: %v", err)
	}

	return text, nil
}

// Close closes the browser and cleans up resources
func (ba *BrowserAutomation) Close() error {
	if err := ba.browser.Close(); err != nil {
		return fmt.Errorf("could not close browser: %v", err)
	}
	if err := ba.pw.Stop(); err != nil {
		return fmt.Errorf("could not stop playwright: %v", err)
	}
	return nil
}
