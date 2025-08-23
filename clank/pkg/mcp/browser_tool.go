package mcp

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	mcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/tebeka/selenium"
)

const (
	seleniumPort   = 4444
	firefoxBinPath = "C:/Program Files/Mozilla Firefox/firefox.exe"
	seleniumJar    = "C:/selenium/selenium-server-4.35.0.jar"
	checkInterval  = 500 * time.Millisecond
	timeout        = 15 * time.Second
)

// BrowserAutomationHandler navigates to a URL and extracts either the article text or raw HTML.
func BrowserAutomationHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	url := req.GetString("url", "")
	if url == "" {
		return mcp.NewToolResultError("url parameter is required"), nil
	}

	savePath := req.GetString("save_path", "")
	extractArticle := req.GetBool("extract_article", false)

	if err := ensureSeleniumServer(); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to start Selenium server: %v", err)), nil
	}

	caps := selenium.Capabilities{"browserName": "firefox"}
	caps["moz:firefoxOptions"] = map[string]interface{}{
		"binary": firefoxBinPath,
		"args":   []string{"--headless"}, // headless for automation
	}

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", seleniumPort))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to connect to WebDriver: %v", err)), nil
	}
	defer wd.Quit()

	if err := wd.Get(url); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to load URL: %v", err)), nil
	}

	// Allow JavaScript to load content
	time.Sleep(2 * time.Second)

	var output string
	if extractArticle {
		// Try accessibility and semantic selectors first
		articleSelectors := []string{
			"article",
			"[role='article']",
			"main",
			"[role='main']",
			"div[data-key='article']",
			".story-body",
			".Article",
		}

		var articleText string
		for _, sel := range articleSelectors {
			elem, err := wd.FindElement(selenium.ByCSSSelector, sel)
			if err == nil && elem != nil {
				// Use JS to get all visible text
				script := `
					var root = arguments[0];
					var walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT, {
						acceptNode: function(node) {
							if(!node.parentElement || node.parentElement.hidden || node.textContent.trim() === '') return NodeFilter.FILTER_REJECT;
							return NodeFilter.FILTER_ACCEPT;
						}
					});
					var text = [];
					while(walker.nextNode()) text.push(walker.currentNode.textContent.trim());
					return text.join("\n\n");
				`
				val, err := wd.ExecuteScript(script, []interface{}{elem})
				if err == nil {
					articleText = strings.TrimSpace(fmt.Sprintf("%v", val))
				}
				break
			}
		}

		// Fallback: all <p> tags
		if articleText == "" {
			paragraphs, _ := wd.FindElements(selenium.ByTagName, "p")
			var texts []string
			for _, p := range paragraphs {
				t, _ := p.Text()
				if strings.TrimSpace(t) != "" {
					texts = append(texts, t)
				}
			}
			articleText = strings.Join(texts, "\n\n")
		}

		output = articleText
	} else {
		html, err := wd.PageSource()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get page source: %v", err)), nil
		}
		output = html
	}

	if savePath != "" {
		if err := os.WriteFile(savePath, []byte(output), 0644); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to save output to file: %v", err)), nil
		}
	}

	return mcp.NewToolResultText(output), nil
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
