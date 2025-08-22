package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"clank/config"
	"clank/internal/llm"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/html"
)

// ExtractorRequest represents the request payload for article extraction
type ExtractorRequest struct {
	URL         string            `json:"url" binding:"required"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	ExtraPrompt string            `json:"extra_prompt,omitempty"`
}

// ExtractorResponse represents the response from the extraction process
type ExtractorResponse struct {
	URL            string                  `json:"url"`
	Title          string                  `json:"title,omitempty"`
	Content        string                  `json:"content,omitempty"`
	Entities       []ExtractedEntity       `json:"entities"`
	Relationships  []ExtractedRelationship `json:"relationships"`
	Events         []ExtractedEvent        `json:"events"`
	Metadata       map[string]interface{}  `json:"metadata,omitempty"`
	ProcessingTime time.Duration           `json:"processing_time"`
	Debug          *DebugInfo              `json:"debug,omitempty"`
}

// ExtractedEntity represents an entity found in the article
type ExtractedEntity struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"` // PERSON, ORGANIZATION, LOCATION, etc.
	Properties map[string]interface{} `json:"properties"`
	Confidence float64                `json:"confidence,omitempty"`
	Mentions   []string               `json:"mentions,omitempty"`
}

// ExtractedRelationship represents a relationship between entities
type ExtractedRelationship struct {
	FromEntity string                 `json:"from_entity"`
	ToEntity   string                 `json:"to_entity"`
	Type       string                 `json:"type"` // DONATED_TO, INVESTIGATED_FOR, etc.
	Properties map[string]interface{} `json:"properties"`
	Confidence float64                `json:"confidence,omitempty"`
	Evidence   string                 `json:"evidence,omitempty"` // Supporting text from article
}

// ExtractedEvent represents a time-based event
type ExtractedEvent struct {
	Description string                 `json:"description"`
	Date        string                 `json:"date,omitempty"`
	Entities    []string               `json:"entities"`
	Type        string                 `json:"type"` // CORRUPTION_CHARGE, DONATION, INVESTIGATION, etc.
	Properties  map[string]interface{} `json:"properties"`
	Evidence    string                 `json:"evidence,omitempty"`
}

// DebugInfo contains debugging information about the extraction process
type DebugInfo struct {
	ScrapeTime    time.Duration `json:"scrape_time"`
	LLMTime       time.Duration `json:"llm_time"`
	ContentLength int           `json:"content_length"`
	WordCount     int           `json:"word_count"`
	HTTPStatus    int           `json:"http_status"`
	ContentType   string        `json:"content_type"`
	LLMTokens     int           `json:"llm_tokens,omitempty"`
	Errors        []string      `json:"errors,omitempty"`
	Warnings      []string      `json:"warnings,omitempty"`
}

// ArticleScraper handles web scraping functionality
type ArticleScraper struct {
	client  *http.Client
	timeout time.Duration
}

// NewArticleScraper creates a new article scraper
func NewArticleScraper() *ArticleScraper {
	return &ArticleScraper{
		client: &http.Client{
			Timeout: 6000 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:    10,
				IdleConnTimeout: 6000 * time.Second,
			},
		},
		timeout: 6000 * time.Second,
	}
}

// ScrapeArticle extracts text content from a web page
func (s *ArticleScraper) ScrapeArticle(articleURL string) (title, content string, statusCode int, contentType string, err error) {
	log.Printf("[EXTRACTOR] Starting scrape of URL: %s", articleURL)

	// Validate URL
	parsedURL, err := url.Parse(articleURL)
	if err != nil {
		return "", "", 0, "", fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", "", 0, "", fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}

	// Create request with headers
	req, err := http.NewRequest("GET", articleURL, nil)
	if err != nil {
		return "", "", 0, "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent to avoid bot blocking
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	// Make request
	resp, err := s.client.Do(req)
	if err != nil {
		return "", "", 0, "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[EXTRACTOR] HTTP response: %d %s, Content-Type: %s",
		resp.StatusCode, resp.Status, resp.Header.Get("Content-Type"))

	if resp.StatusCode != http.StatusOK {
		return "", "", resp.StatusCode, resp.Header.Get("Content-Type"),
			fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", resp.StatusCode, resp.Header.Get("Content-Type"),
			fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse HTML and extract text
	title, content = s.extractTextFromHTML(string(body))

	log.Printf("[EXTRACTOR] Extracted content - Title: %d chars, Content: %d chars",
		len(title), len(content))

	return title, content, resp.StatusCode, resp.Header.Get("Content-Type"), nil
}

// extractTextFromHTML parses HTML and extracts readable text content
func (s *ArticleScraper) extractTextFromHTML(htmlContent string) (title, content string) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		log.Printf("[EXTRACTOR] HTML parsing error: %v", err)
		// Fallback: try to extract text with regex
		return s.extractTextWithRegex(htmlContent)
	}

	// Extract title
	title = s.extractTitle(doc)

	// Extract main content
	content = s.extractMainContent(doc)

	// Clean up the content
	content = s.cleanText(content)

	return title, content
}

// extractTitle finds the page title
func (s *ArticleScraper) extractTitle(doc *html.Node) string {
	var title string
	var findTitle func(*html.Node)
	findTitle = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" {
			if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
				title = strings.TrimSpace(n.FirstChild.Data)
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findTitle(c)
			if title != "" {
				return
			}
		}
	}
	findTitle(doc)
	return title
}

// hasClass checks if a node has a specific CSS class
func hasClass(n *html.Node, className string) bool {
	if n.Type != html.ElementNode {
		return false
	}
	for _, attr := range n.Attr {
		if attr.Key == "class" {
			classes := strings.Fields(attr.Val)
			for _, class := range classes {
				if class == className {
					return true
				}
			}
		}
	}
	return false
}

// hasAttribute checks if a node has a specific attribute with a value
func hasAttribute(n *html.Node, attrName, attrValue string) bool {
	if n.Type != html.ElementNode {
		return false
	}
	for _, attr := range n.Attr {
		if attr.Key == attrName && strings.Contains(attr.Val, attrValue) {
			return true
		}
	}
	return false
}

// isArticleContent determines if a node likely contains article content
func (s *ArticleScraper) isArticleContent(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}

	// AP News specific selectors
	if hasClass(n, "RichTextStoryBody") ||
		hasClass(n, "Article") ||
		hasClass(n, "story-body") ||
		hasAttribute(n, "data-module", "ArticleBody") {
		return true
	}

	// Generic article content selectors
	if n.Data == "article" ||
		(n.Data == "div" && (hasClass(n, "article-body") ||
			hasClass(n, "entry-content") ||
			hasClass(n, "post-content") ||
			hasClass(n, "article-content") ||
			hasClass(n, "story-content"))) ||
		(n.Data == "main" && hasAttribute(n, "role", "main")) ||
		(n.Data == "section" && hasClass(n, "article")) {
		return true
	}

	return false
}

// shouldSkipNode determines if a node should be completely skipped
func (s *ArticleScraper) shouldSkipNode(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}

	// Always skip these tags
	skipTags := map[string]bool{
		"script": true, "style": true, "noscript": true,
	}
	if skipTags[n.Data] {
		return true
	}

	// Skip common non-content areas
	skipClasses := []string{
		"navigation", "nav", "menu", "header", "footer", "sidebar",
		"advertisement", "ad", "ads", "social", "share", "related",
		"comments", "comment", "author", "byline", "tags", "tag",
		"breadcrumb", "newsletter", "subscription", "promo",
		"widget", "aside", "trending", "popular", "recommended",
		"banner", "toolbar", "search", "filter", "pagination",
	}

	// Skip nodes with these classes or IDs
	for _, attr := range n.Attr {
		if attr.Key == "class" || attr.Key == "id" {
			attrLower := strings.ToLower(attr.Val)
			for _, skipClass := range skipClasses {
				if strings.Contains(attrLower, skipClass) {
					return true
				}
			}
		}
	}

	// Skip nav, header, footer, aside elements
	if n.Data == "nav" || n.Data == "header" || n.Data == "footer" ||
		n.Data == "aside" || n.Data == "form" {
		return true
	}

	return false
}

// extractMainContent extracts text from article content areas with better targeting
func (s *ArticleScraper) extractMainContent(doc *html.Node) string {
	var buffer bytes.Buffer
	var foundArticleContent bool

	// First pass: look for specific article content containers
	var findArticleContent func(*html.Node) bool
	findArticleContent = func(n *html.Node) bool {
		if s.shouldSkipNode(n) {
			return false
		}

		if s.isArticleContent(n) {
			log.Printf("[EXTRACTOR] Found article content container: <%s> with classes/id", n.Data)
			s.extractTextFromNode(n, &buffer, true)
			return true
		}

		// Recurse into children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if findArticleContent(c) {
				return true
			}
		}
		return false
	}

	foundArticleContent = findArticleContent(doc)

	// Fallback: if no specific article content found, use broader extraction
	if !foundArticleContent || buffer.Len() < 200 {
		log.Printf("[EXTRACTOR] No specific article content found or content too short, using fallback extraction")
		buffer.Reset()
		s.extractTextFromNode(doc, &buffer, false)
	}

	return buffer.String()
}

// extractTextFromNode extracts text from a node and its children
func (s *ArticleScraper) extractTextFromNode(n *html.Node, buffer *bytes.Buffer, strict bool) {
	if s.shouldSkipNode(n) {
		return
	}

	if n.Type == html.ElementNode {
		// In strict mode, be more selective about what we include
		if strict {
			// Only extract from content-bearing elements
			contentTags := map[string]bool{
				"p": true, "div": true, "span": true, "article": true,
				"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
				"li": true, "ul": true, "ol": true, "blockquote": true,
				"section": true, "main": true,
			}

			if !contentTags[n.Data] {
				// Still recurse into children, but don't add spacing
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					s.extractTextFromNode(c, buffer, strict)
				}
				return
			}
		}

		// Add spacing for block elements
		blockElements := map[string]bool{
			"p": true, "div": true, "article": true, "section": true,
			"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
			"li": true, "blockquote": true, "br": true,
		}
		if blockElements[n.Data] {
			buffer.WriteString(" ")
		}
	}

	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" && len(text) > 2 { // Ignore very short text nodes
			buffer.WriteString(text)
			buffer.WriteString(" ")
		}
	}

	// Recurse into children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		s.extractTextFromNode(c, buffer, strict)
	}
}

// extractTextWithRegex is a fallback method using regex
func (s *ArticleScraper) extractTextWithRegex(htmlContent string) (title, content string) {
	// Extract title
	titleRegex := regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)
	if matches := titleRegex.FindStringSubmatch(htmlContent); len(matches) > 1 {
		title = strings.TrimSpace(matches[1])
	}

	// Try to find article content with common patterns
	articlePatterns := []string{
		`<div[^>]*class="[^"]*(?:story-body|article-body|post-content|entry-content|RichTextStoryBody)[^"]*"[^>]*>(.*?)</div>`,
		`<article[^>]*>(.*?)</article>`,
		`<main[^>]*>(.*?)</main>`,
	}

	for _, pattern := range articlePatterns {
		regex := regexp.MustCompile(`(?s)` + pattern)
		if matches := regex.FindStringSubmatch(htmlContent); len(matches) > 1 {
			content = matches[1]
			break
		}
	}

	// If no specific article content found, fall back to full cleanup
	if content == "" {
		content = htmlContent
	}

	// Remove script and style tags
	scriptRegex := regexp.MustCompile(`(?s)<(script|style)[^>]*>.*?</\1>`)
	content = scriptRegex.ReplaceAllString(content, "")

	// Remove HTML tags
	tagRegex := regexp.MustCompile(`<[^>]+>`)
	content = tagRegex.ReplaceAllString(content, " ")

	// Clean up whitespace
	content = s.cleanText(content)

	return title, content
}

// cleanText removes extra whitespace and normalizes text
func (s *ArticleScraper) cleanText(text string) string {
	// Replace multiple whitespace with single space
	spaceRegex := regexp.MustCompile(`\s+`)
	text = spaceRegex.ReplaceAllString(text, " ")

	// Remove common noise patterns
	noisePatterns := []string{
		`(?i)\s*advertisement\s*`,
		`(?i)\s*subscribe\s*`,
		`(?i)\s*newsletter\s*`,
		`(?i)\s*click here\s*`,
		`(?i)\s*read more\s*`,
		`(?i)\s*share\s*`,
		`(?i)\s*tweet\s*`,
		`(?i)\s*facebook\s*`,
		`(?i)\s*twitter\s*`,
		`(?i)\s*follow us\s*`,
	}

	for _, pattern := range noisePatterns {
		regex := regexp.MustCompile(pattern)
		text = regex.ReplaceAllString(text, " ")
	}

	// Trim and return
	return strings.TrimSpace(text)
}

// generateExtractionPrompt creates the prompt for the LLM
// generateExtractionPrompt creates the prompt for the LLM with concise output
func generateExtractionPrompt(title, content, extraPrompt string) string {
	basePrompt := `You are an expert at extracting structured information from news articles for a corruption tracking system. 

Analyze the following article and extract:
1. ENTITIES: People, organizations, locations, and other relevant entities
2. RELATIONSHIPS: Connections between entities (donations, investigations, employment, etc.)
3. EVENTS: Time-based events with dates when possible

Focus on corruption-related information including:
- Political figures and their relationships
- Financial transactions and donations
- Legal investigations and charges
- Corporate connections and lobbying
- Conflicts of interest

Article Title: %s

Article Content: %s

Return the results in this exact JSON format:
{
  "entities": [
    {
      "name": "Entity Name",
      "type": "PERSON|ORGANIZATION|LOCATION|GOVERNMENT_AGENCY|CORPORATION",
      "properties": {
        "title": "job title if person",
        "industry": "if organization",
        "party": "if politician",
        "jurisdiction": "if location/agency"
      },
      "confidence": 0.0-1.0,
      "mentions": ["context where mentioned"]
    }
  ],
  "relationships": [
    {
      "from_entity": "Entity 1 Name",
      "to_entity": "Entity 2 Name", 
      "type": "DONATED_TO|INVESTIGATED_FOR|CONVICTED_OF|EMPLOYED_BY|LOBBIED_FOR|CONTRACTS_WITH",
      "properties": {
        "amount": "if monetary",
        "date": "YYYY-MM-DD if available",
        "description": "brief, one-sentence description"
      },
      "confidence": 0.0-1.0,
      "evidence": "concise, relevant supporting text (max 20 words)"
    }
  ],
  "events": [
    {
      "description": "Brief, one-sentence event description",
      "date": "YYYY-MM-DD if available",
      "entities": ["Entity names involved"],
      "type": "CORRUPTION_CHARGE|INVESTIGATION|DONATION|LOBBYING|CONVICTION|INDICTMENT",
      "properties": {
        "amount": "if monetary",
        "charge": "if legal",
        "outcome": "if resolved"
      },
      "evidence": "concise, relevant supporting text (max 20 words)"
    }
  ]
}

Additional Instructions: %s

- Keep all descriptions and evidence short and focused; do not include full paragraphs.
- Use a single sentence or phrase for evidence and description fields.
- Only include information explicitly stated or strongly implied in the article.
- Ensure all entity names are consistent across entities, relationships, and events.
- Assign confidence scores based on how clearly the information is stated.`

	return fmt.Sprintf(basePrompt, title, content, extraPrompt)
}

// ExtractorHandler handles the article extraction endpoint
func ExtractorHandler(cfg *config.Config) gin.HandlerFunc {
	scraper := NewArticleScraper()
	llmClient := llm.NewClient(cfg)

	return func(c *gin.Context) {
		startTime := time.Now()
		debug := &DebugInfo{
			Errors:   make([]string, 0),
			Warnings: make([]string, 0),
		}

		log.Printf("[EXTRACTOR] Starting extraction request from %s", c.ClientIP())

		// Parse request
		var req ExtractorRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			debug.Errors = append(debug.Errors, fmt.Sprintf("Invalid request format: %v", err))
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request format",
				"debug": debug,
			})
			return
		}

		log.Printf("[EXTRACTOR] Processing URL: %s", req.URL)

		// Scrape article
		scrapeStart := time.Now()
		title, content, statusCode, contentType, err := scraper.ScrapeArticle(req.URL)
		debug.ScrapeTime = time.Since(scrapeStart)
		debug.HTTPStatus = statusCode
		debug.ContentType = contentType
		debug.ContentLength = len(content)
		debug.WordCount = len(strings.Fields(content))

		if err != nil {
			debug.Errors = append(debug.Errors, fmt.Sprintf("Scraping failed: %v", err))
			log.Printf("[EXTRACTOR] Scraping error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Failed to scrape article: %v", err),
				"debug": debug,
			})
			return
		}

		// Validate content length
		if len(content) < 100 {
			debug.Warnings = append(debug.Warnings, "Article content is very short, results may be limited")
		}
		if len(content) > 50000 {
			debug.Warnings = append(debug.Warnings, "Article content is very long, truncating for LLM processing")
			content = content[:50000] + "... [TRUNCATED]"
		}

		log.Printf("[EXTRACTOR] Content extracted - Title: %d chars, Content: %d chars",
			len(title), len(content))

		// Generate extraction prompt
		prompt := generateExtractionPrompt(title, content, req.ExtraPrompt)

		// Create messages for LLM
		messages := []llm.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		}

		// Call LLM for extraction
		llmStart := time.Now()
		result, err := llmClient.Generate(c.Request.Context(), messages)
		debug.LLMTime = time.Since(llmStart)

		if err != nil {
			debug.Errors = append(debug.Errors, fmt.Sprintf("LLM processing failed: %v", err))
			log.Printf("[EXTRACTOR] LLM error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to process article with LLM: %v", err),
				"debug": debug,
			})
			return
		}

		if len(result.Choices) == 0 {
			debug.Errors = append(debug.Errors, "LLM returned no results")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "LLM returned no results",
				"debug": debug,
			})
			return
		}

		llmResponse := result.Choices[0].Message.Content
		log.Printf("[EXTRACTOR] LLM response length: %d chars", len(llmResponse))

		// Parse LLM JSON response
		var extractionResult struct {
			Entities      []ExtractedEntity       `json:"entities"`
			Relationships []ExtractedRelationship `json:"relationships"`
			Events        []ExtractedEvent        `json:"events"`
		}

		if err := json.Unmarshal([]byte(llmResponse), &extractionResult); err != nil {
			debug.Errors = append(debug.Errors, fmt.Sprintf("Failed to parse LLM response as JSON: %v", err))
			debug.Warnings = append(debug.Warnings, "Raw LLM response included in debug output")

			log.Printf("[EXTRACTOR] JSON parsing error: %v", err)
			log.Printf("[EXTRACTOR] Raw LLM response: %s", llmResponse)

			// Try to extract JSON from response if it's wrapped in other text
			jsonStart := strings.Index(llmResponse, "{")
			jsonEnd := strings.LastIndex(llmResponse, "}")
			if jsonStart >= 0 && jsonEnd > jsonStart {
				jsonOnly := llmResponse[jsonStart : jsonEnd+1]
				if err := json.Unmarshal([]byte(jsonOnly), &extractionResult); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":        "Failed to parse LLM response",
						"llm_response": llmResponse,
						"debug":        debug,
					})
					return
				}
				debug.Warnings = append(debug.Warnings, "Successfully extracted JSON from wrapped response")
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":        "Failed to parse LLM response",
					"llm_response": llmResponse,
					"debug":        debug,
				})
				return
			}
		}

		// Build response
		response := ExtractorResponse{
			URL:            req.URL,
			Title:          title,
			Content:        content,
			Entities:       extractionResult.Entities,
			Relationships:  extractionResult.Relationships,
			Events:         extractionResult.Events,
			ProcessingTime: time.Since(startTime),
			Debug:          debug,
		}

		// Add metadata if provided
		if req.Metadata != nil {
			response.Metadata = make(map[string]interface{})
			for k, v := range req.Metadata {
				response.Metadata[k] = v
			}
		}

		log.Printf("[EXTRACTOR] Extraction completed successfully - Entities: %d, Relationships: %d, Events: %d, Processing time: %v",
			len(response.Entities), len(response.Relationships), len(response.Events), response.ProcessingTime)

		c.JSON(http.StatusOK, response)
	}
}
