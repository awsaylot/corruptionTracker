package tools

import (
	"context"
	"fmt"

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

	// Launch the browser in headless mode for production
	ba.browser, err = ba.pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true), // Headless for production
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
	if _, err := ba.page.Goto(url, playwright.PageGotoOptions{
		Timeout:   playwright.Float(60000),
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		return fmt.Errorf("could not navigate to %s: %v", url, err)
	}
	return nil
}

// WaitForSelector waits for an element matching the specified selector to be present
func (ba *BrowserAutomation) WaitForSelector(ctx context.Context, selector string) error {
	if _, err := ba.page.WaitForSelector(selector, playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(30000),
	}); err != nil {
		return fmt.Errorf("could not wait for selector %s: %v", selector, err)
	}
	return nil
}

// Close closes the browser and cleans up resources
func (ba *BrowserAutomation) Close() error {
	if ba.browser != nil {
		if err := ba.browser.Close(); err != nil {
			return fmt.Errorf("could not close browser: %v", err)
		}
	}
	if ba.pw != nil {
		if err := ba.pw.Stop(); err != nil {
			return fmt.Errorf("could not stop playwright: %v", err)
		}
	}
	return nil
}
