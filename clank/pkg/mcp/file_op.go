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

// handleReadLines reads specific line ranges from a file
func handleReadLines(path, options string) (*mcp.CallToolResult, error) {
	start, end, err := parseLineRange(options)
	if err != nil {
		return mcp.NewToolResultError("line range error: " + err.Error()), nil
	}

	file, err := os.Open(path)
	if err != nil {
		return mcp.NewToolResultError("failed to open file: " + err.Error()), nil
	}
	defer file.Close()

	var result strings.Builder
	scanner := bufio.NewScanner(file)
	lineNum := 0
	linesRead := 0

	result.WriteString(fmt.Sprintf("Lines %d-%d from %s:\n\n", start, end, path))

	for scanner.Scan() {
		lineNum++
		if lineNum >= start && lineNum <= end {
			result.WriteString(fmt.Sprintf("%4d: %s\n", lineNum, scanner.Text()))
			linesRead++
		}
		if lineNum > end {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return mcp.NewToolResultError("error reading file: " + err.Error()), nil
	}

	if linesRead == 0 {
		result.WriteString("(No lines found in specified range)")
	}

	result.WriteString(fmt.Sprintf("\nRead %d lines (range %d-%d)", linesRead, start, end))

	return mcp.NewToolResultText(result.String()), nil
}

// handleEdit performs in-place editing of specific lines
func handleEdit(path, content, options string) (*mcp.CallToolResult, error) {
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

// handleRead reads file with metadata
func handleRead(path string) (*mcp.CallToolResult, error) {
	info, err := os.Stat(path)
	if err != nil {
		return mcp.NewToolResultError("failed to stat file: " + err.Error()), nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return mcp.NewToolResultError("failed to read file: " + err.Error()), nil
	}

	// Count lines efficiently
	lineCount := strings.Count(string(content), "\n") + 1
	if len(content) == 0 {
		lineCount = 0
	}

	result := fmt.Sprintf("File: %s\nSize: %d bytes\nLines: %d\nModified: %s\n\n%s",
		path, info.Size(), lineCount, info.ModTime().Format("2006-01-02 15:04:05"), string(content))

	return mcp.NewToolResultText(result), nil
}

// handleWrite writes file efficiently with backup option
func handleWrite(path, content string) (*mcp.CallToolResult, error) {
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
		}

		result.WriteString(fmt.Sprintf("%s %s (%d bytes) - %s\n",
			prefix, entry.Name(), info.Size(), info.ModTime().Format("2006-01-02 15:04")))
	}

	return mcp.NewToolResultText(result.String()), nil
}

// handleSearch searches for pattern in files
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

		// Skip binary files and large files
		if info.Size() > 1024*1024 { // 1MB limit
			return nil
		}

		// Only search text files
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

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		scanner := bufio.NewScanner(strings.NewReader(string(content)))
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

// handleDiff compares two files
func handleDiff(path1, path2 string) (*mcp.CallToolResult, error) {
	content1, err := os.ReadFile(path1)
	if err != nil {
		return mcp.NewToolResultError("failed to read first file: " + err.Error()), nil
	}

	content2, err := os.ReadFile(path2)
	if err != nil {
		return mcp.NewToolResultError("failed to read second file: " + err.Error()), nil
	}

	lines1 := strings.Split(string(content1), "\n")
	lines2 := strings.Split(string(content2), "\n")

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
