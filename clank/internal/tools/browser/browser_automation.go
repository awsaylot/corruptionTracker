package browser

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
	if ba.page == nil {
		return fmt.Errorf("browser not initialized")
	}

	_, err := ba.page.Goto(url)
	if err != nil {
		return fmt.Errorf("failed to navigate to %s: %v", url, err)
	}

	return nil
}

// GetElementTextBySelector gets the text content of an element by selector
func (ba *BrowserAutomation) GetElementTextBySelector(ctx context.Context, selector string) (string, error) {
	if ba.page == nil {
		return "", fmt.Errorf("browser not initialized")
	}

	element, err := ba.page.QuerySelector(selector)
	if err != nil {
		return "", fmt.Errorf("failed to query selector %s: %v", selector, err)
	}
	if element == nil {
		return "", fmt.Errorf("no element found for selector %s", selector)
	}

	text, err := element.TextContent()
	if err != nil {
		return "", fmt.Errorf("failed to get text content: %v", err)
	}

	return text, nil
}

// GetElementAttributeBySelector gets the value of an attribute from an element by selector
func (ba *BrowserAutomation) GetElementAttributeBySelector(ctx context.Context, selector, attribute string) (string, error) {
	if ba.page == nil {
		return "", fmt.Errorf("browser not initialized")
	}

	element, err := ba.page.QuerySelector(selector)
	if err != nil {
		return "", fmt.Errorf("failed to query selector %s: %v", selector, err)
	}
	if element == nil {
		return "", fmt.Errorf("no element found for selector %s", selector)
	}

	value, err := element.GetAttribute(attribute)
	if err != nil {
		return "", fmt.Errorf("failed to get attribute %s: %v", attribute, err)
	}

	return value, nil
}

// Close cleans up browser resources
func (ba *BrowserAutomation) Close() error {
	if ba.page != nil {
		err := ba.page.Close()
		if err != nil {
			return fmt.Errorf("failed to close page: %v", err)
		}
	}

	if ba.context != nil {
		err := ba.context.Close()
		if err != nil {
			return fmt.Errorf("failed to close context: %v", err)
		}
	}

	if ba.browser != nil {
		err := ba.browser.Close()
		if err != nil {
			return fmt.Errorf("failed to close browser: %v", err)
		}
	}

	if ba.pw != nil {
		err := ba.pw.Stop()
		if err != nil {
			return fmt.Errorf("failed to stop playwright: %v", err)
		}
	}

	return nil
}
