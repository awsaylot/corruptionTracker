package mcp

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	mcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/tebeka/selenium"
)

const (
	seleniumPort   = 4444
	firefoxBinPath = "C:/Program Files/Mozilla Firefox/firefox.exe"
	chromeBinPath  = "C:/Program Files/Google/Chrome/Application/chrome.exe"
	seleniumJar    = "C:/selenium/selenium-server-4.35.0.jar"
	checkInterval  = 500 * time.Millisecond
	timeout        = 15 * time.Second
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

// Enhanced BrowserAutomationHandler with comprehensive features
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
	takeScreenshot := req.GetBool("take_screenshot", false)
	executeJS := req.GetString("execute_js", "")
	waitForElement := req.GetString("wait_for_element", "")
	clickElement := req.GetString("click_element", "")
	typeText := req.GetString("type_text", "")
	typeSelector := req.GetString("type_selector", "")
	scrollTo := req.GetString("scroll_to", "")
	savePath := req.GetString("save_path", "")

	// Cookie management
	setCookies := req.GetString("set_cookies", "") // JSON string of cookies
	getCookies := req.GetBool("get_cookies", false)

	// Form handling
	fillForm := req.GetString("fill_form", "") // JSON string of form data

	if err := ensureSeleniumServer(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to start Selenium server: %v", err)), nil
	}

	// Create WebDriver with enhanced configuration
	wd, err := createWebDriver(config)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create WebDriver: %v", err)), nil
	}
	defer wd.Quit()

	// Set implicit wait
	wd.SetImplicitWaitTimeout(time.Duration(config.ImplicitWait) * time.Second)

	// Set cookies if provided (before navigating to URL)
	if setCookies != "" {
		if err := handleSetCookies(wd, setCookies, url); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to set cookies: %v", err)), nil
		}
	}

	// Navigate to URL
	if err := wd.Get(url); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to load URL: %v", err)), nil
	}

	// Wait for page load
	time.Sleep(time.Duration(config.PageLoadWait) * time.Second)

	// Wait for specific element if requested
	if waitForElement != "" {
		if err := waitForElementPresent(wd, waitForElement, 10*time.Second); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Element not found within timeout: %v", err)), nil
		}
	}

	// Handle form filling
	if fillForm != "" {
		if err := handleFormFilling(wd, fillForm); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to fill form: %v", err)), nil
		}
	}

	// Handle element clicking
	if clickElement != "" {
		if err := handleElementClick(wd, clickElement); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to click element: %v", err)), nil
		}
	}

	// Handle text typing
	if typeText != "" && typeSelector != "" {
		if err := handleTextTyping(wd, typeSelector, typeText); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to type text: %v", err)), nil
		}
	}

	// Handle scrolling
	if scrollTo != "" {
		if err := handleScrolling(wd, scrollTo); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to scroll: %v", err)), nil
		}
	}

	// Execute custom JavaScript
	var jsResult interface{}
	if executeJS != "" {
		result, err := wd.ExecuteScript(executeJS, nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to execute JavaScript: %v", err)), nil
		}
		jsResult = result
	}

	// Collect results
	results := make(map[string]interface{})

	// Extract content based on request
	if extractArticle {
		articleText, err := extractArticleTextAdvanced(wd)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to extract article: %v", err)), nil
		}
		results["article_text"] = articleText
	}

	// Get page HTML if not extracting article
	if !extractArticle {
		html, err := wd.PageSource()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get page source: %v", err)), nil
		}
		results["html"] = html
	}

	// Take screenshot if requested
	if takeScreenshot {
		screenshot, err := wd.Screenshot()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to take screenshot: %v", err)), nil
		}
		results["screenshot_base64"] = base64.StdEncoding.EncodeToString(screenshot)
	}

	// Get cookies if requested
	if getCookies {
		cookies, err := wd.GetCookies()
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
	title, _ := wd.Title()
	currentURL, _ := wd.CurrentURL()
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

func createWebDriver(config BrowserConfig) (selenium.WebDriver, error) {
	var caps selenium.Capabilities

	switch strings.ToLower(config.Browser) {
	case "firefox":
		caps = selenium.Capabilities{"browserName": "firefox"}
		firefoxOptions := map[string]interface{}{
			"binary": firefoxBinPath,
		}

		args := []string{}
		if config.Headless {
			args = append(args, "--headless")
		}
		if config.DisableImages {
			args = append(args, "--disable-images")
		}

		firefoxOptions["args"] = args

		// Firefox preferences
		prefs := make(map[string]interface{})
		if config.UserAgent != "" {
			prefs["general.useragent.override"] = config.UserAgent
		}
		if config.DisableCSS {
			prefs["permissions.default.stylesheet"] = 2
		}
		if len(prefs) > 0 {
			firefoxOptions["prefs"] = prefs
		}

		caps["moz:firefoxOptions"] = firefoxOptions

	case "chrome":
		caps = selenium.Capabilities{"browserName": "chrome"}
		chromeOptions := map[string]interface{}{
			"binary": chromeBinPath,
		}

		args := []string{}
		if config.Headless {
			args = append(args, "--headless", "--no-sandbox", "--disable-dev-shm-usage")
		}
		if config.UserAgent != "" {
			args = append(args, "--user-agent="+config.UserAgent)
		}
		if config.DisableImages {
			args = append(args, "--blink-settings=imagesEnabled=false")
		}
		if config.WindowSize != "" {
			args = append(args, "--window-size="+config.WindowSize)
		}
		if config.ProxyURL != "" {
			args = append(args, "--proxy-server="+config.ProxyURL)
		}

		chromeOptions["args"] = args
		caps["goog:chromeOptions"] = chromeOptions

	default:
		return nil, fmt.Errorf("unsupported browser: %s", config.Browser)
	}

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", seleniumPort))
	if err != nil {
		return nil, err
	}

	// Set window size
	// if config.WindowSize != "" && !config.Headless {
	// 	parts := strings.Split(config.WindowSize, ",")
	// 	if len(parts) == 2 {
	// 		width, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	// 		height, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
	// 		if width > 0 && height > 0 {
	// 			wd.SetWindowSize(width, height)
	// 		}
	// 	}
	// }

	return wd, nil
}

func extractArticleTextAdvanced(wd selenium.WebDriver) (string, error) {
	// Enhanced article extraction using multiple strategies

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
		if text := tryExtractWithSelector(wd, selector); text != "" {
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
		if text := tryExtractWithSelector(wd, selector); text != "" {
			return cleanArticleText(text), nil
		}
	}

	// Strategy 3: Readability algorithm (simplified)
	return extractWithReadabilityAlgorithm(wd)
}

func tryExtractWithSelector(wd selenium.WebDriver, selector string) string {
	element, err := wd.FindElement(selenium.ByCSSSelector, selector)
	if err != nil {
		return ""
	}

	// Use JavaScript to extract clean text
	script := `
		var element = arguments[0];
		var text = '';
		
		function extractText(node) {
			if (node.nodeType === Node.TEXT_NODE) {
				return node.textContent.trim();
			}
			
			var result = '';
			var children = node.childNodes;
			
			for (var i = 0; i < children.length; i++) {
				var child = children[i];
				if (child.nodeType === Node.ELEMENT_NODE) {
					var tagName = child.tagName.toLowerCase();
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
	`

	result, err := wd.ExecuteScript(script, []interface{}{element})
	if err != nil {
		return ""
	}

	return strings.TrimSpace(fmt.Sprintf("%v", result))
}

func extractWithReadabilityAlgorithm(wd selenium.WebDriver) (string, error) {
	// Simplified readability algorithm
	script := `
		function scoreElement(element) {
			var score = 0;
			var text = element.textContent || '';
			
			// Score based on text length
			score += Math.min(text.length / 100, 10);
			
			// Score based on paragraph count
			var paragraphs = element.querySelectorAll('p');
			score += paragraphs.length * 2;
			
			// Penalize elements with many links
			var links = element.querySelectorAll('a');
			if (links.length > paragraphs.length) {
				score -= links.length - paragraphs.length;
			}
			
			// Bonus for article-like class names
			var className = element.className || '';
			var id = element.id || '';
			var combined = (className + ' ' + id).toLowerCase();
			
			if (/article|content|main|post|story|text|body/.test(combined)) {
				score += 5;
			}
			if (/sidebar|nav|menu|comment|footer|header|ad/.test(combined)) {
				score -= 3;
			}
			
			return score;
		}
		
		var candidates = document.querySelectorAll('div, article, section, main');
		var bestElement = null;
		var bestScore = 0;
		
		for (var i = 0; i < candidates.length; i++) {
			var candidate = candidates[i];
			var score = scoreElement(candidate);
			
			if (score > bestScore) {
				bestScore = score;
				bestElement = candidate;
			}
		}
		
		if (bestElement) {
			var text = '';
			var paragraphs = bestElement.querySelectorAll('p, h1, h2, h3, h4, h5, h6');
			
			for (var i = 0; i < paragraphs.length; i++) {
				var p = paragraphs[i];
				var pText = p.textContent.trim();
				if (pText.length > 20) {
					text += pText + '\n\n';
				}
			}
			
			return text;
		}
		
		return '';
	`

	result, err := wd.ExecuteScript(script, nil)
	if err != nil {
		return "", err
	}

	text := fmt.Sprintf("%v", result)
	if text == "" {
		// Fallback to all paragraphs
		paragraphs, _ := wd.FindElements(selenium.ByTagName, "p")
		var texts []string
		for _, p := range paragraphs {
			t, _ := p.Text()
			if len(strings.TrimSpace(t)) > 20 {
				texts = append(texts, t)
			}
		}
		text = strings.Join(texts, "\n\n")
	}

	return cleanArticleText(text), nil
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

func waitForElementPresent(wd selenium.WebDriver, selector string, timeout time.Duration) error {
	start := time.Now()
	for time.Since(start) < timeout {
		_, err := wd.FindElement(selenium.ByCSSSelector, selector)
		if err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("element '%s' not found within %v", selector, timeout)
}

func handleSetCookies(wd selenium.WebDriver, cookiesJSON, targetURL string) error {
	// Parse URL to get domain
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return err
	}

	// Navigate to domain first (required for setting cookies)
	wd.Get(fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host))

	// Simple cookie parsing (expecting format: "name1=value1; name2=value2")
	cookiePairs := strings.Split(cookiesJSON, ";")
	for _, pair := range cookiePairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 {
			cookie := &selenium.Cookie{
				Name:   strings.TrimSpace(parts[0]),
				Value:  strings.TrimSpace(parts[1]),
				Domain: parsedURL.Host,
			}
			wd.AddCookie(cookie)
		}
	}

	return nil
}

func handleFormFilling(wd selenium.WebDriver, formJSON string) error {
	// Simple form filling (expecting format: "selector1=value1; selector2=value2")
	formPairs := strings.Split(formJSON, ";")
	for _, pair := range formPairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 {
			selector := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			element, err := wd.FindElement(selenium.ByCSSSelector, selector)
			if err != nil {
				continue
			}

			element.Clear()
			element.SendKeys(value)
		}
	}

	return nil
}

func handleElementClick(wd selenium.WebDriver, selector string) error {
	element, err := wd.FindElement(selenium.ByCSSSelector, selector)
	if err != nil {
		return err
	}

	return element.Click()
}

func handleTextTyping(wd selenium.WebDriver, selector, text string) error {
	element, err := wd.FindElement(selenium.ByCSSSelector, selector)
	if err != nil {
		return err
	}

	element.Clear()
	return element.SendKeys(text)
}

func handleScrolling(wd selenium.WebDriver, scrollTo string) error {
	var script string

	switch scrollTo {
	case "top":
		script = "window.scrollTo(0, 0);"
	case "bottom":
		script = "window.scrollTo(0, document.body.scrollHeight);"
	default:
		// Try as CSS selector
		script = fmt.Sprintf(`
			var element = document.querySelector('%s');
			if (element) {
				element.scrollIntoView({behavior: 'smooth', block: 'center'});
			}
		`, scrollTo)
	}

	_, err := wd.ExecuteScript(script, nil)
	return err
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

func ensureSeleniumServer() error {
	if isPortOpen("localhost", seleniumPort) {
		return nil
	}

	cmd := exec.Command("java", "-jar", seleniumJar, "standalone", "--port", fmt.Sprint(seleniumPort))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not start Selenium server: %w", err)
	}

	start := time.Now()
	for {
		if isPortOpen("localhost", seleniumPort) {
			return nil
		}
		if time.Since(start) > timeout {
			return fmt.Errorf("timeout waiting for Selenium server to start")
		}
		time.Sleep(checkInterval)
	}
}

func isPortOpen(host string, port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 500*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// Helper function to safely get int from request
// func (req mcp.CallToolRequest) GetInt(key string, defaultValue int) int {
// 	if val, exists := req.Params.Arguments[key]; exists {
// 		if intVal, ok := val.(float64); ok {
// 			return int(intVal)
// 		}
// 		if strVal, ok := val.(string); ok {
// 			if intVal, err := strconv.Atoi(strVal); err == nil {
// 				return intVal
// 			}
// 		}
// 	}
// 	return defaultValue
// }
