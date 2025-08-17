package handlers

import (
	"net/http"

	"clank/internal/models"
	"clank/internal/prompts"

	"github.com/gin-gonic/gin"
)

// Global prompt service instance
var promptServiceInstance *prompts.Service

// InitPromptService initializes the global prompt service
func InitPromptService(promptsDir string) error {
	service, err := prompts.NewService(promptsDir)
	if err != nil {
		return err
	}
	promptServiceInstance = service
	return nil
}

// ListPrompts lists all loaded prompts
func ListPrompts(c *gin.Context) {
	if promptServiceInstance == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Prompt service not initialized"})
		return
	}

	allPrompts := promptServiceInstance.ListAvailablePrompts()
	result := make(map[string]interface{})
	for name, prompt := range allPrompts {
		result[name] = map[string]interface{}{
			"name":        prompt.Name,
			"description": prompt.Description,
			"arguments":   prompt.Arguments,
			"metadata":    prompt.Metadata,
			"loaded_at":   prompt.LoadedAt,
			"compiled_at": prompt.CompiledAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"prompts": result, "count": len(result)})
}

// GetPrompt returns details about a specific prompt
func GetPrompt(c *gin.Context) {
	if promptServiceInstance == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Prompt service not initialized"})
		return
	}

	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prompt name is required"})
		return
	}

	prompt, err := promptServiceInstance.GetPromptInfo(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prompt not found", "name": name})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"prompt": map[string]interface{}{
			"name":        prompt.Name,
			"description": prompt.Description,
			"arguments":   prompt.Arguments,
			"template":    prompt.Template,
			"metadata":    prompt.Metadata,
			"file_path":   prompt.FilePath,
			"loaded_at":   prompt.LoadedAt,
			"compiled_at": prompt.CompiledAt,
		},
	})
}

// RenderPromptRequest defines the input for rendering a prompt
type RenderPromptRequest struct {
	Arguments map[string]interface{} `json:"arguments"`
	UserID    string                 `json:"user_id,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	NodeData  map[string]interface{} `json:"node_data,omitempty"`
}

// RenderPrompt renders a prompt with provided arguments
func RenderPrompt(c *gin.Context) {
	if promptServiceInstance == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Prompt service not initialized"})
		return
	}

	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prompt name is required"})
		return
	}

	var req RenderPromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	// Validate arguments
	if err := promptServiceInstance.ValidatePromptArguments(name, req.Arguments); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid arguments", "details": err.Error()})
		return
	}

	// Render the prompt
	result, err := promptServiceInstance.RenderPrompt(c.Request.Context(), name, req.Arguments)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render prompt", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"prompt_name":   name,
		"rendered_text": result,
		"arguments":     req.Arguments,
	})
}

// ReloadPrompts reloads all prompts from disk
func ReloadPrompts(c *gin.Context) {
	if promptServiceInstance == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Prompt service not initialized"})
		return
	}

	if err := promptServiceInstance.ReloadPrompts(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload prompts", "details": err.Error()})
		return
	}

	allPrompts := promptServiceInstance.ListAvailablePrompts()
	c.JSON(http.StatusOK, gin.H{"message": "Prompts reloaded successfully", "count": len(allPrompts)})
}

// ValidatePrompt validates arguments for a prompt without rendering
func ValidatePrompt(c *gin.Context) {
	if promptServiceInstance == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Prompt service not initialized"})
		return
	}

	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Prompt name is required"})
		return
	}

	var req struct {
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	if err := promptServiceInstance.ValidatePromptArguments(name, req.Arguments); err != nil {
		if ve, ok := err.(*models.ValidationError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"valid": false, "error": ve.Message, "field": ve.Field, "value": ve.Value})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"valid": false, "error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": true, "message": "Arguments are valid"})
}
