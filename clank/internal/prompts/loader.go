package prompts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"

	"clank/internal/models"
)

// PromptLoader manages loading and rendering of prompt templates
type PromptLoader struct {
	mu           sync.RWMutex
	prompts      map[string]*models.Prompt
	templates    map[string]*template.Template
	promptsDir   string
	watchEnabled bool
	lastModified map[string]time.Time
}

// NewPromptLoader creates a new prompt loader
func NewPromptLoader(promptsDir string) *PromptLoader {
	return &PromptLoader{
		prompts:      make(map[string]*models.Prompt),
		templates:    make(map[string]*template.Template),
		promptsDir:   promptsDir,
		lastModified: make(map[string]time.Time),
	}
}

// LoadPrompts loads all prompt files from the prompts directory
func (pl *PromptLoader) LoadPrompts() error {
	if _, err := os.Stat(pl.promptsDir); os.IsNotExist(err) {
		return fmt.Errorf("prompts directory does not exist: %s", pl.promptsDir)
	}

	return filepath.WalkDir(pl.promptsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		return pl.LoadPromptFile(path)
	})
}

// LoadPromptFile loads a single prompt file
func (pl *PromptLoader) LoadPromptFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read prompt file %s: %w", filePath, err)
	}

	var prompt models.Prompt
	if err := json.Unmarshal(data, &prompt); err != nil {
		return fmt.Errorf("failed to parse prompt file %s: %w", filePath, err)
	}

	// Validate prompt
	if err := pl.validatePrompt(&prompt); err != nil {
		return fmt.Errorf("invalid prompt in file %s: %w", filePath, err)
	}

	// Set runtime fields
	prompt.FilePath = filePath
	prompt.LoadedAt = time.Now()

	// Compile template
	tmpl, err := pl.compileTemplate(&prompt)
	if err != nil {
		return fmt.Errorf("failed to compile template for %s: %w", prompt.Name, err)
	}

	prompt.CompiledAt = time.Now()

	// Get file modification time
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info for %s: %w", filePath, err)
	}

	pl.mu.Lock()
	defer pl.mu.Unlock()

	pl.prompts[prompt.Name] = &prompt
	pl.templates[prompt.Name] = tmpl
	pl.lastModified[filePath] = fileInfo.ModTime()

	fmt.Printf("Loaded prompt: %s from %s\n", prompt.Name, filePath)
	return nil
}

// validatePrompt validates a prompt structure
func (pl *PromptLoader) validatePrompt(prompt *models.Prompt) error {
	if prompt.Name == "" {
		return &models.ValidationError{Field: "name", Message: "prompt name is required"}
	}

	if prompt.Template == "" {
		return &models.ValidationError{Field: "template", Message: "prompt template is required"}
	}

	// Validate arguments
	for i, arg := range prompt.Arguments {
		if arg.Name == "" {
			return &models.ValidationError{
				Field:   fmt.Sprintf("arguments[%d].name", i),
				Message: "argument name is required",
			}
		}
	}

	return nil
}

// compileTemplate compiles a Handlebars-style template
func (pl *PromptLoader) compileTemplate(prompt *models.Prompt) (*template.Template, error) {
	// Create custom template functions
	funcMap := template.FuncMap{
		"json": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
		"indent": func(spaces int, text string) string {
			indent := strings.Repeat(" ", spaces)
			lines := strings.Split(text, "\n")
			for i, line := range lines {
				if line != "" {
					lines[i] = indent + line
				}
			}
			return strings.Join(lines, "\n")
		},
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": strings.Title,

		// 🔑 context function so you can call {{context "UserID"}}
		"context": func(key string, ctx map[string]interface{}) interface{} {
			if v, ok := ctx[key]; ok {
				return v
			}
			return ""
		},
	}

	// Convert Handlebars syntax to Go template syntax
	templateText := pl.convertHandlebarsToGo(prompt.Template)

	tmpl, err := template.New(prompt.Name).Funcs(funcMap).Parse(templateText)
	if err != nil {
		return nil, fmt.Errorf("template parsing error: %w", err)
	}

	return tmpl, nil
}

// convertHandlebarsToGo converts Handlebars syntax to Go template syntax
func (pl *PromptLoader) convertHandlebarsToGo(text string) string {
	// Convert {{#if condition}} to {{if condition}}
	text = strings.ReplaceAll(text, "{{#if ", "{{if ")

	// Convert {{/if}} to {{end}}
	text = strings.ReplaceAll(text, "{{/if}}", "{{end}}")

	// Convert {{{variable}}} (unescaped) to {{.variable}}
	text = strings.ReplaceAll(text, "{{{", "{{")
	text = strings.ReplaceAll(text, "}}}", "}}")

	// For now, leave variable resolution as-is
	return text
}

// GetPrompt returns a prompt by name
func (pl *PromptLoader) GetPrompt(name string) (*models.Prompt, error) {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	prompt, exists := pl.prompts[name]
	if !exists {
		return nil, fmt.Errorf("prompt not found: %s", name)
	}

	return prompt, nil
}

// ListPrompts returns all loaded prompts
func (pl *PromptLoader) ListPrompts() map[string]*models.Prompt {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	result := make(map[string]*models.Prompt)
	for k, v := range pl.prompts {
		result[k] = v
	}
	return result
}

// RenderPrompt renders a prompt with the given context
func (pl *PromptLoader) RenderPrompt(name string, context *models.PromptContext) (*models.PromptResult, error) {
	startTime := time.Now()

	pl.mu.RLock()
	prompt, exists := pl.prompts[name]
	tmpl, tmplExists := pl.templates[name]
	pl.mu.RUnlock()

	if !exists || !tmplExists {
		return nil, fmt.Errorf("prompt not found: %s", name)
	}

	// Validate required arguments
	if err := pl.validateContext(prompt, context); err != nil {
		return nil, fmt.Errorf("context validation failed: %w", err)
	}

	// Flatten context for {{context}} helper
	flatCtx := map[string]interface{}{
		"Arguments": context.Arguments,
		"Metadata":  context.Metadata,
		"NodeData":  context.NodeData,
		"UserID":    context.UserID,
		"SessionID": context.SessionID,
		"Timestamp": context.Timestamp,
	}

	// Prepare template data
	data := map[string]interface{}{
		"Arguments": context.Arguments,
		"Metadata":  context.Metadata,
		"NodeData":  context.NodeData,
		"UserID":    context.UserID,
		"SessionID": context.SessionID,
		"Timestamp": context.Timestamp,
		"Context":   flatCtx, // 🔑 needed for {{context "UserID" .Context}}
	}

	// Render template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("template execution failed: %w", err)
	}

	return &models.PromptResult{
		PromptName:   name,
		RenderedText: buf.String(),
		Arguments:    context.Arguments,
		Metadata:     context.Metadata,
		RenderTime:   time.Since(startTime),
		Timestamp:    time.Now(),
	}, nil
}

// validateContext validates that required arguments are provided
func (pl *PromptLoader) validateContext(prompt *models.Prompt, context *models.PromptContext) error {
	for _, arg := range prompt.Arguments {
		if arg.Required {
			if _, exists := context.Arguments[arg.Name]; !exists {
				return &models.ValidationError{
					Field:   arg.Name,
					Message: fmt.Sprintf("required argument '%s' is missing", arg.Name),
				}
			}
		}
	}
	return nil
}

// ReloadPrompt reloads a specific prompt file
func (pl *PromptLoader) ReloadPrompt(name string) error {
	pl.mu.RLock()
	prompt, exists := pl.prompts[name]
	pl.mu.RUnlock()

	if !exists {
		return fmt.Errorf("prompt not found: %s", name)
	}

	return pl.LoadPromptFile(prompt.FilePath)
}

// ReloadIfChanged checks if prompt files have changed and reloads them
func (pl *PromptLoader) ReloadIfChanged() error {
	pl.mu.RLock()
	filesToCheck := make(map[string]time.Time)
	for path, modTime := range pl.lastModified {
		filesToCheck[path] = modTime
	}
	pl.mu.RUnlock()

	for filePath, lastMod := range filesToCheck {
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			// File might have been deleted
			continue
		}

		if fileInfo.ModTime().After(lastMod) {
			fmt.Printf("Reloading changed prompt file: %s\n", filePath)
			if err := pl.LoadPromptFile(filePath); err != nil {
				fmt.Printf("Error reloading prompt file %s: %v\n", filePath, err)
			}
		}
	}

	return nil
}
