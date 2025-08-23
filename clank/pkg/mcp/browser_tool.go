package mcp

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	mcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/playwright-community/playwright-go"
)

type BrowserConfig struct {
	Browser       string
	Headless      bool
	UserAgent     string
	WindowSize    string
	ImplicitWait  int
	PageLoadWait  int
	DisableImages bool
	DisableCSS    bool
	ProxyURL      string
}

type PlaywrightSession struct {
	playwright *playwright.Playwright
	browser    playwright.Browser
	page       playwright.Page
}

// Enhanced BrowserAutomationHandler with Playwright
func BrowserAutomationHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	url := req.GetString("url", "")
	if url == "" {
		return mcp.NewToolResultError("url parameter is required"), nil
	}

	// Parse configuration
	config := BrowserConfig{
		Browser:       req.GetString("browser", "firefox"),
		Headless:      req.GetBool("headless", true),
		UserAgent:     req.GetString("user_agent", ""),
		WindowSize:    req.GetString("window_size", "1920,1080"),
		ImplicitWait:  req.GetInt("implicit_wait", 10),
		PageLoadWait:  req.GetInt("page_load_wait", 2),
		DisableImages: req.GetBool("disable_images", false),
		DisableCSS:    req.GetBool("disable_css", false),
		ProxyURL:      req.GetString("proxy_url", ""),
	}

	// Actions configuration
	extractArticle := req.GetBool("extract_article", false)
	executeJS := req.GetString("execute_js", "")
	waitForElement := req.GetString("wait_for_element", "")
	clickElement := req.GetString("click_element", "")
	typeText := req.GetString("type_text", "")
	typeSelector := req.GetString("type_selector", "")
	scrollTo := req.GetString("scroll_to", "")
	savePath := req.GetString("save_path", "")

	// Cookie management
	setCookies := req.GetString("set_cookies", "")
	getCookies := req.GetBool("get_cookies", false)

	// Form handling
	fillForm := req.GetString("fill_form", "")

	// Initialize Playwright session
	session, err := initPlaywrightSession(ctx, config)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to initialize Playwright: %v", err)), nil
	}
	defer cleanupPlaywrightSession(session)

	// Set default timeout
	session.page.SetDefaultTimeout(float64(config.ImplicitWait * 1000))

	// Set cookies if provided (before navigating to URL)
	if setCookies != "" {
		if err := handleSetCookiesPlaywright(session.page, setCookies, url); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to set cookies: %v", err)), nil
		}
	}

	// Navigate to URL with enhanced options
	navigateOptions := playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
		Timeout:   playwright.Float(float64(config.PageLoadWait * 1000)),
	}

	if _, err := session.page.Goto(url, navigateOptions); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to load URL: %v", err)), nil
	}

	// Additional wait for page load
	time.Sleep(time.Duration(config.PageLoadWait) * time.Second)

	// Wait for specific element if requested
	if waitForElement != "" {
		if err := waitForElementPlaywright(session.page, waitForElement, time.Duration(config.ImplicitWait)*time.Second); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Element not found within timeout: %v", err)), nil
		}
	}

	// Handle form filling
	if fillForm != "" {
		if err := handleFormFillingPlaywright(session.page, fillForm); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fill form: %v", err)), nil
		}
	}

	// Handle element clicking
	if clickElement != "" {
		if err := handleElementClickPlaywright(session.page, clickElement); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to click element: %v", err)), nil
		}
	}

	// Handle text typing
	if typeText != "" && typeSelector != "" {
		if err := handleTextTypingPlaywright(session.page, typeSelector, typeText); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to type text: %v", err)), nil
		}
	}

	// Handle scrolling
	if scrollTo != "" {
		if err := handleScrollingPlaywright(session.page, scrollTo); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to scroll: %v", err)), nil
		}
	}

	// Execute custom JavaScript
	var jsResult interface{}
	if executeJS != "" {
		result, err := session.page.Evaluate(executeJS)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to execute JavaScript: %v", err)), nil
		}
		jsResult = result
	}

	// Collect results
	results := make(map[string]interface{})

	// Extract content based on request
	if extractArticle {
		articleText, err := extractArticleTextPlaywright(session.page)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to extract article: %v", err)), nil
		}
		results["article_text"] = articleText
	}

	// Get page HTML if not extracting article
	if !extractArticle {
		html, err := session.page.Content()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get page source: %v", err)), nil
		}
		results["html"] = html
	}

	// Take screenshot if requested
	// if takeScreenshot {
	// 	screenshot, err := session.page.Screenshot(playwright.PageScreenshotOptions{
	// 		FullPage: playwright.Bool(true),
	// 		Type:     playwright.ScreenshotTypePNG,
	// 	})
	// 	if err != nil {
	// 		return mcp.NewToolResultError(fmt.Sprintf("Failed to take screenshot: %v", err)), nil
	// 	}
	// 	results["screenshot_base64"] = base64.StdEncoding.EncodeToString(screenshot)
	// }

	// Get cookies if requested
	if getCookies {
		cookies, err := session.page.Context().Cookies()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get cookies: %v", err)), nil
		}
		results["cookies"] = cookies
	}

	// Add JavaScript result if executed
	if jsResult != nil {
		results["js_result"] = jsResult
	}

	// Add page metadata
	title, _ := session.page.Title()
	currentURL := session.page.URL()
	results["page_title"] = title
	results["current_url"] = currentURL

	// Save output if requested
	if savePath != "" {
		var outputToSave string
		if extractArticle {
			outputToSave = fmt.Sprintf("%v", results["article_text"])
		} else {
			outputToSave = fmt.Sprintf("%v", results["html"])
		}

		if err := os.WriteFile(savePath, []byte(outputToSave), 0644); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to save output to file: %v", err)), nil
		}
		results["saved_to"] = savePath
	}

	// Format output
	output := formatResults(results)
	return mcp.NewToolResultText(output), nil
}

func initPlaywrightSession(ctx context.Context, config BrowserConfig) (*PlaywrightSession, error) {
	// Install browsers if needed (this should be done once during setup)
	runOptions := &playwright.RunOptions{
		SkipInstallBrowsers: true, // Set to false on first run
	}

	pw, err := playwright.Run(runOptions)
	if err != nil {
		return nil, fmt.Errorf("could not start Playwright: %v", err)
	}

	// Browser launch options
	launchOptions := playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(config.Headless),
	}

	// Parse window size
	var width, height int = 1920, 1080
	if config.WindowSize != "" {
		parts := strings.Split(config.WindowSize, ",")
		if len(parts) == 2 {
			if w, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
				width = w
			}
			if h, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
				height = h
			}
		}
	}

	// Browser-specific configuration
	var browser playwright.Browser
	switch strings.ToLower(config.Browser) {
	case "firefox":
		if config.ProxyURL != "" {
			launchOptions.Proxy = &playwright.Proxy{
				Server: config.ProxyURL,
			}
		}
		browser, err = pw.Firefox.Launch(launchOptions)
	case "chrome", "chromium":
		args := []string{}
		if config.DisableImages {
			args = append(args, "--blink-settings=imagesEnabled=false")
		}
		if len(args) > 0 {
			launchOptions.Args = args
		}
		if config.ProxyURL != "" {
			launchOptions.Proxy = &playwright.Proxy{
				Server: config.ProxyURL,
			}
		}
		browser, err = pw.Chromium.Launch(launchOptions)
	case "webkit", "safari":
		browser, err = pw.WebKit.Launch(launchOptions)
	default:
		return nil, fmt.Errorf("unsupported browser: %s", config.Browser)
	}

	if err != nil {
		pw.Stop()
		return nil, fmt.Errorf("could not launch browser: %v", err)
	}

	// Create browser context with enhanced options
	contextOptions := playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{
			Width:  width,
			Height: height,
		},
		IgnoreHttpsErrors: playwright.Bool(true),
	}

	// Set user agent if provided
	if config.UserAgent != "" {
		contextOptions.UserAgent = playwright.String(config.UserAgent)
	}

	context, err := browser.NewContext(contextOptions)
	if err != nil {
		browser.Close()
		pw.Stop()
		return nil, fmt.Errorf("could not create browser context: %v", err)
	}

	// Block resources if requested
	if config.DisableImages || config.DisableCSS {
		err = context.Route("**/*", func(route playwright.Route) {
			request := route.Request()
			resourceType := request.ResourceType()

			if config.DisableImages && resourceType == "image" {
				route.Abort()
				return
			}
			if config.DisableCSS && resourceType == "stylesheet" {
				route.Abort()
				return
			}
			route.Continue()
		})
		if err != nil {
			context.Close()
			browser.Close()
			pw.Stop()
			return nil, fmt.Errorf("could not set up resource blocking: %v", err)
		}
	}

	// Create new page
	page, err := context.NewPage()
	if err != nil {
		context.Close()
		browser.Close()
		pw.Stop()
		return nil, fmt.Errorf("could not create page: %v", err)
	}

	return &PlaywrightSession{
		playwright: pw,
		browser:    browser,
		page:       page,
	}, nil
}

func cleanupPlaywrightSession(session *PlaywrightSession) {
	if session != nil {
		if session.page != nil {
			session.page.Close()
		}
		if session.browser != nil {
			session.browser.Close()
		}
		if session.playwright != nil {
			session.playwright.Stop()
		}
	}
}

func extractArticleTextPlaywright(page playwright.Page) (string, error) {
	// Enhanced article extraction using multiple strategies with Playwright

	// Strategy 1: Semantic HTML5 elements
	semanticSelectors := []string{
		"article",
		"[role='article']",
		"main article",
		"[role='main'] article",
		".article-content",
		".post-content",
		".entry-content",
		".content-body",
	}

	for _, selector := range semanticSelectors {
		if text := tryExtractWithSelectorPlaywright(page, selector); text != "" {
			return cleanArticleText(text), nil
		}
	}

	// Strategy 2: Common article container patterns
	containerSelectors := []string{
		"main",
		"[role='main']",
		".main-content",
		".article-body",
		".story-body",
		".post-body",
		".content-wrapper",
		"#article-content",
		"#main-content",
	}

	for _, selector := range containerSelectors {
		if text := tryExtractWithSelectorPlaywright(page, selector); text != "" {
			return cleanArticleText(text), nil
		}
	}

	// Strategy 3: Readability algorithm (simplified)
	return extractWithReadabilityAlgorithmPlaywright(page)
}

func tryExtractWithSelectorPlaywright(page playwright.Page, selector string) string {
	// Check if element exists
	count, err := page.Locator(selector).Count()
	if err != nil || count == 0 {
		return ""
	}

	// Use JavaScript to extract clean text
	script := `
		(selector) => {
			const element = document.querySelector(selector);
			if (!element) return '';
			
			function extractText(node) {
				if (node.nodeType === Node.TEXT_NODE) {
					return node.textContent.trim();
				}
				
				let result = '';
				const children = node.childNodes;
				
				for (let i = 0; i < children.length; i++) {
					const child = children[i];
					if (child.nodeType === Node.ELEMENT_NODE) {
						const tagName = child.tagName.toLowerCase();
						if (['script', 'style', 'nav', 'header', 'footer', 'aside'].includes(tagName)) {
							continue;
						}
						if (['p', 'div', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'li'].includes(tagName)) {
							result += extractText(child) + '\n\n';
						} else {
							result += extractText(child);
						}
					} else if (child.nodeType === Node.TEXT_NODE) {
						result += child.textContent.trim() + ' ';
					}
				}
				
				return result;
			}
			
			return extractText(element);
		}
	`

	result, err := page.Evaluate(script, selector)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(fmt.Sprintf("%v", result))
}

func extractWithReadabilityAlgorithmPlaywright(page playwright.Page) (string, error) {
	// Simplified readability algorithm using Playwright
	script := `
		() => {
			function scoreElement(element) {
				let score = 0;
				const text = element.textContent || '';
				
				// Score based on text length
				score += Math.min(text.length / 100, 10);
				
				// Score based on paragraph count
				const paragraphs = element.querySelectorAll('p');
				score += paragraphs.length * 2;
				
				// Penalize elements with many links
				const links = element.querySelectorAll('a');
				if (links.length > paragraphs.length) {
					score -= links.length - paragraphs.length;
				}
				
				// Bonus for article-like class names
				const className = element.className || '';
				const id = element.id || '';
				const combined = (className + ' ' + id).toLowerCase();
				
				if (/article|content|main|post|story|text|body/.test(combined)) {
					score += 5;
				}
				if (/sidebar|nav|menu|comment|footer|header|ad/.test(combined)) {
					score -= 3;
				}
				
				return score;
			}
			
			const candidates = document.querySelectorAll('div, article, section, main');
			let bestElement = null;
			let bestScore = 0;
			
			for (let i = 0; i < candidates.length; i++) {
				const candidate = candidates[i];
				const score = scoreElement(candidate);
				
				if (score > bestScore) {
					bestScore = score;
					bestElement = candidate;
				}
			}
			
			if (bestElement) {
				let text = '';
				const paragraphs = bestElement.querySelectorAll('p, h1, h2, h3, h4, h5, h6');
				
				for (let i = 0; i < paragraphs.length; i++) {
					const p = paragraphs[i];
					const pText = p.textContent.trim();
					if (pText.length > 20) {
						text += pText + '\n\n';
					}
				}
				
				return text;
			}
			
			return '';
		}
	`

	result, err := page.Evaluate(script)
	if err != nil {
		return "", err
	}

	text := fmt.Sprintf("%v", result)
	if text == "" {
		// Fallback to all paragraphs
		paragraphs := page.Locator("p")
		count, _ := paragraphs.Count()
		var texts []string

		for i := 0; i < count; i++ {
			t, err := paragraphs.Nth(i).TextContent()
			if err == nil && len(strings.TrimSpace(t)) > 20 {
				texts = append(texts, t)
			}
		}
		text = strings.Join(texts, "\n\n")
	}

	return cleanArticleText(text), nil
}

func waitForElementPlaywright(page playwright.Page, selector string, timeout time.Duration) error {
	_, err := page.WaitForSelector(selector, playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(float64(timeout.Milliseconds())),
	})
	return err
}

func handleSetCookiesPlaywright(page playwright.Page, cookiesData, targetURL string) error {
	// Parse URL to get domain
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return err
	}

	// Simple cookie parsing (expecting format: "name1=value1; name2=value2")
	var cookies []playwright.OptionalCookie
	cookiePairs := strings.Split(cookiesData, ";")

	for _, pair := range cookiePairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 {
			cookie := playwright.OptionalCookie{
				Name:   strings.TrimSpace(parts[0]),
				Value:  strings.TrimSpace(parts[1]),
				Domain: playwright.String(parsedURL.Host),
			}
			cookies = append(cookies, cookie)
		}
	}

	return page.Context().AddCookies(cookies)
}

func handleFormFillingPlaywright(page playwright.Page, formData string) error {
	// Simple form filling (expecting format: "selector1=value1; selector2=value2")
	formPairs := strings.Split(formData, ";")
	for _, pair := range formPairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 {
			selector := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			locator := page.Locator(selector)
			if count, _ := locator.Count(); count > 0 {
				locator.Fill(value)
			}
		}
	}

	return nil
}

func handleElementClickPlaywright(page playwright.Page, selector string) error {
	return page.Locator(selector).Click()
}

func handleTextTypingPlaywright(page playwright.Page, selector, text string) error {
	locator := page.Locator(selector)
	if err := locator.Clear(); err != nil {
		return err
	}
	return locator.Type(text)
}

func handleScrollingPlaywright(page playwright.Page, scrollTo string) error {
	switch scrollTo {
	case "top":
		_, err := page.Evaluate("() => window.scrollTo(0, 0)")
		return err
	case "bottom":
		_, err := page.Evaluate("() => window.scrollTo(0, document.body.scrollHeight)")
		return err
	default:
		// Try as CSS selector
		locator := page.Locator(scrollTo)
		if count, _ := locator.Count(); count > 0 {
			return locator.ScrollIntoViewIfNeeded()
		}
		return fmt.Errorf("element not found for scrolling: %s", scrollTo)
	}
}

func cleanArticleText(text string) string {
	// Remove excessive whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	// Remove multiple newlines
	text = regexp.MustCompile(`\n\s*\n\s*\n`).ReplaceAllString(text, "\n\n")
	// Trim whitespace
	text = strings.TrimSpace(text)

	return text
}

func formatResults(results map[string]interface{}) string {
	var output strings.Builder

	// Page information
	if title, ok := results["page_title"]; ok {
		output.WriteString(fmt.Sprintf("Page Title: %s\n", title))
	}
	if url, ok := results["current_url"]; ok {
		output.WriteString(fmt.Sprintf("Current URL: %s\n\n", url))
	}

	// Main content
	if articleText, ok := results["article_text"]; ok {
		output.WriteString("=== ARTICLE TEXT ===\n")
		output.WriteString(fmt.Sprintf("%s\n\n", articleText))
	}

	if html, ok := results["html"]; ok && results["article_text"] == nil {
		output.WriteString("=== PAGE HTML ===\n")
		htmlStr := fmt.Sprintf("%s", html)
		if len(htmlStr) > 5000 {
			output.WriteString(htmlStr[:5000] + "\n... (truncated, full HTML available)\n\n")
		} else {
			output.WriteString(htmlStr + "\n\n")
		}
	}

	// JavaScript result
	if jsResult, ok := results["js_result"]; ok {
		output.WriteString("=== JAVASCRIPT RESULT ===\n")
		output.WriteString(fmt.Sprintf("%v\n\n", jsResult))
	}

	// Screenshot info
	if screenshot, ok := results["screenshot_base64"]; ok {
		screenshotStr := fmt.Sprintf("%s", screenshot)
		output.WriteString("=== SCREENSHOT ===\n")
		output.WriteString(fmt.Sprintf("Screenshot captured (%d bytes, base64 encoded)\n", len(screenshotStr)))
		output.WriteString("Base64 data: " + screenshotStr[:100] + "...\n\n")
	}

	// Cookies
	if cookies, ok := results["cookies"]; ok {
		output.WriteString("=== COOKIES ===\n")
		output.WriteString(fmt.Sprintf("%v\n\n", cookies))
	}

	// File save info
	if savedTo, ok := results["saved_to"]; ok {
		output.WriteString(fmt.Sprintf("Content saved to: %s\n", savedTo))
	}

	return output.String()
}

// Helper function to safely get int from request (keeping for compatibility)
func GetInt(req mcp.CallToolRequest, key string, defaultValue int) int {
	// Assert that Arguments is a map
	argsMap, ok := req.Params.Arguments.(map[string]interface{})
	if !ok || argsMap == nil {
		return defaultValue
	}

	val, exists := argsMap[key]
	if !exists || val == nil {
		return defaultValue
	}

	switch v := val.(type) {
	case float64:
		return int(v)
	case float32:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case string:
		if intVal, err := strconv.Atoi(v); err == nil {
			return intVal
		}
	}

	return defaultValue
}
