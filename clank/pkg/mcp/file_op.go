package mcp

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/ledongthuc/pdf"
	mcp "github.com/mark3labs/mcp-go/mcp"
)

// FileOperationsHandler handles file operations dynamically with enhanced features for LLM development
func FileOperationsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	operation, err := request.RequireString("operation")
	if err != nil {
		return mcp.NewToolResultError("operation must be a string"), nil
	}

	path, err := request.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path must be a string"), nil
	}

	// Clean and resolve path
	path = filepath.Clean(path)

	switch operation {
	case "read":
		return handleRead(path)

	case "read_lines":
		options := request.GetString("options", "")
		return handleReadLines(path, options)

	case "write":
		content := request.GetString("content", "")
		return handleWrite(path, content)

	case "edit":
		content := request.GetString("content", "")
		options := request.GetString("options", "")
		return handleEdit(path, content, options)

	case "list":
		options := request.GetString("options", "{}")
		return handleList(path, options)

	case "search":
		pattern := request.GetString("pattern", "")
		if pattern == "" {
			return mcp.NewToolResultError("pattern is required for search operation"), nil
		}
		options := request.GetString("options", "{}")
		return handleSearch(path, pattern, options)

	case "diff":
		path2 := request.GetString("path2", "")
		if path2 == "" {
			return mcp.NewToolResultError("path2 is required for diff operation"), nil
		}
		return handleDiff(path, path2)

	case "mkdir":
		return handleMkdir(path)

	default:
		return mcp.NewToolResultError("invalid operation. Use: read, read_lines, write, edit, list, search, diff, or mkdir"), nil
	}
}

// isPDF checks if the file is a PDF based on file extension
func isPDF(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".pdf"
}

// extractTextFromPDF extracts text content from PDF file
func extractTextFromPDF(path string) (string, error) {
	file, reader, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %v", err)
	}
	defer file.Close()

	var textBuilder strings.Builder
	totalPages := reader.NumPage()

	textBuilder.WriteString(fmt.Sprintf("=== PDF Content (%d pages) ===\n\n", totalPages))

	// Create an empty font map for GetPlainText
	fontMap := make(map[string]*pdf.Font)

	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		page := reader.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		textBuilder.WriteString(fmt.Sprintf("--- Page %d ---\n", pageNum))

		// Extract text from page with font map
		pageText, err := page.GetPlainText(fontMap)
		if err != nil {
			textBuilder.WriteString(fmt.Sprintf("Error extracting text from page %d: %v\n", pageNum, err))
			continue
		}

		// Clean up the extracted text
		cleanedText := strings.TrimSpace(pageText)
		if cleanedText == "" {
			textBuilder.WriteString("(No text content found on this page)\n")
		} else {
			textBuilder.WriteString(cleanedText)
			textBuilder.WriteString("\n")
		}
		textBuilder.WriteString("\n")
	}

	return textBuilder.String(), nil
}

// parseLineRange parses line range from options like "10-20" or "10:20"
func parseLineRange(options string) (int, int, error) {
	options = strings.TrimSpace(options)
	if options == "" {
		return 0, 0, fmt.Errorf("line range required")
	}

	// Handle formats: "10-20", "10:20", "10,20"
	separators := []string{"-", ":", ","}
	var parts []string
	for _, sep := range separators {
		if strings.Contains(options, sep) {
			parts = strings.Split(options, sep)
			break
		}
	}

	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range format, use 'start-end' (e.g., '10-20')")
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start line number: %v", err)
	}

	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end line number: %v", err)
	}

	if start < 1 {
		start = 1
	}
	if end < start {
		return 0, 0, fmt.Errorf("end line must be >= start line")
	}

	return start, end, nil
}

// handleReadLines reads specific line ranges from a file (including PDF support)
func handleReadLines(path, options string) (*mcp.CallToolResult, error) {
	start, end, err := parseLineRange(options)
	if err != nil {
		return mcp.NewToolResultError("line range error: " + err.Error()), nil
	}

	var content string

	if isPDF(path) {
		// For PDF files, first extract all text, then work with lines
		pdfText, err := extractTextFromPDF(path)
		if err != nil {
			return mcp.NewToolResultError("failed to read PDF: " + err.Error()), nil
		}
		content = pdfText
	} else {
		// For regular files, read normally
		fileContent, err := os.ReadFile(path)
		if err != nil {
			return mcp.NewToolResultError("failed to open file: " + err.Error()), nil
		}
		content = string(fileContent)
	}

	lines := strings.Split(content, "\n")
	totalLines := len(lines)

	if start > totalLines {
		return mcp.NewToolResultError(fmt.Sprintf("start line %d exceeds file length %d", start, totalLines)), nil
	}
	if end > totalLines {
		end = totalLines
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Lines %d-%d from %s:\n\n", start, end, path))

	linesRead := 0
	for lineNum := start; lineNum <= end; lineNum++ {
		result.WriteString(fmt.Sprintf("%4d: %s\n", lineNum, lines[lineNum-1]))
		linesRead++
	}

	result.WriteString(fmt.Sprintf("\nRead %d lines (range %d-%d)", linesRead, start, end))

	return mcp.NewToolResultText(result.String()), nil
}

// handleEdit performs in-place editing of specific lines (not supported for PDF)
func handleEdit(path, content, options string) (*mcp.CallToolResult, error) {
	if isPDF(path) {
		return mcp.NewToolResultError("editing PDF files is not supported"), nil
	}

	start, end, err := parseLineRange(options)
	if err != nil {
		return mcp.NewToolResultError("line range error: " + err.Error()), nil
	}

	// Read the entire file
	originalContent, err := os.ReadFile(path)
	if err != nil {
		return mcp.NewToolResultError("failed to read file: " + err.Error()), nil
	}

	lines := strings.Split(string(originalContent), "\n")
	totalLines := len(lines)

	// Validate range
	if start > totalLines {
		return mcp.NewToolResultError(fmt.Sprintf("start line %d exceeds file length %d", start, totalLines)), nil
	}
	if end > totalLines {
		end = totalLines
	}

	// Prepare new content lines
	newContentLines := strings.Split(content, "\n")

	// Build the new file content
	var newLines []string

	// Add lines before the edit range
	if start > 1 {
		newLines = append(newLines, lines[:start-1]...)
	}

	// Add the new content
	newLines = append(newLines, newContentLines...)

	// Add lines after the edit range
	if end < totalLines {
		newLines = append(newLines, lines[end:]...)
	}

	// Join and write atomically
	newFileContent := strings.Join(newLines, "\n")

	// Write atomically using temp file
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, []byte(newFileContent), 0644); err != nil {
		return mcp.NewToolResultError("failed to write temp file: " + err.Error()), nil
	}

	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath) // cleanup
		return mcp.NewToolResultError("failed to rename temp file: " + err.Error()), nil
	}

	linesReplaced := end - start + 1
	newLinesCount := len(newContentLines)
	netChange := newLinesCount - linesReplaced

	var changeDesc string
	if netChange > 0 {
		changeDesc = fmt.Sprintf(" (+%d lines)", netChange)
	} else if netChange < 0 {
		changeDesc = fmt.Sprintf(" (%d lines)", netChange)
	}

	result := fmt.Sprintf("Edit successful: %s\nReplaced lines %d-%d (%d lines) with %d lines%s\nNew file length: %d lines",
		path, start, end, linesReplaced, newLinesCount, changeDesc, len(newLines))

	return mcp.NewToolResultText(result), nil
}

// handleRead reads file with metadata (enhanced with PDF support)
func handleRead(path string) (*mcp.CallToolResult, error) {
	info, err := os.Stat(path)
	if err != nil {
		return mcp.NewToolResultError("failed to stat file: " + err.Error()), nil
	}

	var content string
	var lineCount int
	var fileType string

	if isPDF(path) {
		fileType = "PDF"
		pdfContent, err := extractTextFromPDF(path)
		if err != nil {
			return mcp.NewToolResultError("failed to read PDF: " + err.Error()), nil
		}
		content = pdfContent
		lineCount = strings.Count(content, "\n") + 1
		if len(content) == 0 {
			lineCount = 0
		}
	} else {
		fileType = "Text"
		fileContent, err := os.ReadFile(path)
		if err != nil {
			return mcp.NewToolResultError("failed to read file: " + err.Error()), nil
		}
		content = string(fileContent)
		// Count lines efficiently
		lineCount = strings.Count(content, "\n") + 1
		if len(fileContent) == 0 {
			lineCount = 0
		}
	}

	result := fmt.Sprintf("File: %s\nType: %s\nSize: %d bytes\nLines: %d\nModified: %s\n\n%s",
		path, fileType, info.Size(), lineCount, info.ModTime().Format("2006-01-02 15:04:05"), content)

	return mcp.NewToolResultText(result), nil
}

// handleWrite writes file efficiently with backup option (not for PDF)
func handleWrite(path, content string) (*mcp.CallToolResult, error) {
	if isPDF(path) {
		return mcp.NewToolResultError("writing PDF files is not supported"), nil
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return mcp.NewToolResultError("failed to create directory: " + err.Error()), nil
	}

	// Write file atomically
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, []byte(content), 0644); err != nil {
		return mcp.NewToolResultError("failed to write temp file: " + err.Error()), nil
	}

	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath) // cleanup
		return mcp.NewToolResultError("failed to rename temp file: " + err.Error()), nil
	}

	lines := strings.Count(content, "\n") + 1
	if len(content) == 0 {
		lines = 0
	}

	result := fmt.Sprintf("File written successfully: %s (%d bytes, %d lines)", path, len(content), lines)
	return mcp.NewToolResultText(result), nil
}

// handleList lists directory with filtering options
func handleList(path, options string) (*mcp.CallToolResult, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return mcp.NewToolResultError("failed to list directory: " + err.Error()), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Directory: %s (%d items)\n\n", path, len(entries)))

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		prefix := "📄"
		if entry.IsDir() {
			prefix = "📁"
		} else if isPDF(entry.Name()) {
			prefix = "📕"
		}

		result.WriteString(fmt.Sprintf("%s %s (%d bytes) - %s\n",
			prefix, entry.Name(), info.Size(), info.ModTime().Format("2006-01-02 15:04")))
	}

	return mcp.NewToolResultText(result.String()), nil
}

// handleSearch searches for pattern in files (including PDF support)
func handleSearch(basePath, pattern, options string) (*mcp.CallToolResult, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return mcp.NewToolResultError("invalid regex pattern: " + err.Error()), nil
	}

	var results []string
	maxResults := 50 // Limit results for performance

	err = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}

		if info.IsDir() || len(results) >= maxResults {
			return nil
		}

		// Skip very large files
		if info.Size() > 10*1024*1024 { // 10MB limit
			return nil
		}

		var content string

		if isPDF(path) {
			// Search in PDF content
			pdfText, err := extractTextFromPDF(path)
			if err != nil {
				return nil // Skip PDFs that can't be read
			}
			content = pdfText
		} else {
			// Only search text files for non-PDF files
			ext := strings.ToLower(filepath.Ext(path))
			textExts := map[string]bool{
				".go": true, ".js": true, ".ts": true, ".py": true, ".java": true,
				".c": true, ".cpp": true, ".h": true, ".hpp": true, ".rs": true,
				".txt": true, ".md": true, ".json": true, ".yaml": true, ".yml": true,
				".xml": true, ".html": true, ".css": true, ".sql": true,
			}

			if !textExts[ext] && ext != "" {
				return nil
			}

			fileContent, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			content = string(fileContent)
		}

		scanner := bufio.NewScanner(strings.NewReader(content))
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			if regex.MatchString(line) {
				relPath, _ := filepath.Rel(basePath, path)
				results = append(results, fmt.Sprintf("%s:%d: %s", relPath, lineNum, strings.TrimSpace(line)))
				if len(results) >= maxResults {
					break
				}
			}
		}

		return nil
	})

	if err != nil {
		return mcp.NewToolResultError("search failed: " + err.Error()), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No matches found for pattern: %s", pattern)), nil
	}

	result := fmt.Sprintf("Found %d matches for pattern '%s':\n\n%s",
		len(results), pattern, strings.Join(results, "\n"))

	if len(results) >= maxResults {
		result += fmt.Sprintf("\n\n(Limited to %d results)", maxResults)
	}

	return mcp.NewToolResultText(result), nil
}

// handleDiff compares two files (PDF support limited)
func handleDiff(path1, path2 string) (*mcp.CallToolResult, error) {
	var content1, content2 string

	// Read first file
	if isPDF(path1) {
		pdfText, err := extractTextFromPDF(path1)
		if err != nil {
			return mcp.NewToolResultError("failed to read first PDF file: " + err.Error()), nil
		}
		content1 = pdfText
	} else {
		fileContent, err := os.ReadFile(path1)
		if err != nil {
			return mcp.NewToolResultError("failed to read first file: " + err.Error()), nil
		}
		content1 = string(fileContent)
	}

	// Read second file
	if isPDF(path2) {
		pdfText, err := extractTextFromPDF(path2)
		if err != nil {
			return mcp.NewToolResultError("failed to read second PDF file: " + err.Error()), nil
		}
		content2 = pdfText
	} else {
		fileContent, err := os.ReadFile(path2)
		if err != nil {
			return mcp.NewToolResultError("failed to read second file: " + err.Error()), nil
		}
		content2 = string(fileContent)
	}

	lines1 := strings.Split(content1, "\n")
	lines2 := strings.Split(content2, "\n")

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Comparing %s vs %s\n\n", path1, path2))

	maxLen := len(lines1)
	if len(lines2) > maxLen {
		maxLen = len(lines2)
	}

	differences := 0
	for i := 0; i < maxLen && differences < 20; i++ { // Limit diff output
		line1 := ""
		line2 := ""

		if i < len(lines1) {
			line1 = lines1[i]
		}
		if i < len(lines2) {
			line2 = lines2[i]
		}

		if line1 != line2 {
			differences++
			result.WriteString(fmt.Sprintf("Line %d:\n", i+1))
			if line1 != "" {
				result.WriteString(fmt.Sprintf("- %s\n", line1))
			}
			if line2 != "" {
				result.WriteString(fmt.Sprintf("+ %s\n", line2))
			}
			result.WriteString("\n")
		}
	}

	if differences == 0 {
		result.WriteString("Files are identical")
	} else if differences >= 20 {
		result.WriteString("(Too many differences, showing first 20)")
	}

	return mcp.NewToolResultText(result.String()), nil
}

// handleMkdir creates directory
func handleMkdir(path string) (*mcp.CallToolResult, error) {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return mcp.NewToolResultError("failed to create directory: " + err.Error()), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Directory created successfully: %s", path)), nil
}
