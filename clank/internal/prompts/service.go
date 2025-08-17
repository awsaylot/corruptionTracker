package prompts

import (
	"context"
	"fmt"
	"time"

	"clank/internal/models"
)

// Service provides high-level prompt operations
type Service struct {
	loader *PromptLoader
}

// NewService creates a new prompt service
func NewService(promptsDir string) (*Service, error) {
	loader := NewPromptLoader(promptsDir)

	// Load all prompts at startup
	if err := loader.LoadPrompts(); err != nil {
		return nil, fmt.Errorf("failed to load prompts: %w", err)
	}

	return &Service{
		loader: loader,
	}, nil
}

// GetSystemPrompt renders the system prompt with context
func (s *Service) GetSystemPrompt(ctx context.Context, graphContext string, userPreferences string) (string, error) {
	promptCtx := &models.PromptContext{
		Arguments: map[string]any{
			"context":          graphContext,
			"user_preferences": userPreferences,
		},
		Timestamp: time.Now(),
	}

	result, err := s.loader.RenderPrompt("system", promptCtx)
	if err != nil {
		return "", fmt.Errorf("failed to render system prompt: %w", err)
	}

	return result.RenderedText, nil
}

// GetAnalysisPrompt renders a network analysis prompt
func (s *Service) GetAnalysisPrompt(ctx context.Context, entityData string, analysisType string, depth int) (string, error) {
	promptCtx := &models.PromptContext{
		Arguments: map[string]any{
			"entity_data":   entityData,
			"analysis_type": analysisType,
			"depth":         depth,
		},
		Timestamp: time.Now(),
	}

	result, err := s.loader.RenderPrompt("network_analysis", promptCtx)
	if err != nil {
		return "", fmt.Errorf("failed to render analysis prompt: %w", err)
	}

	return result.RenderedText, nil
}

// GetCorruptionAnalysisPrompt renders a corruption analysis prompt
func (s *Service) GetCorruptionAnalysisPrompt(ctx context.Context, corruptionData string, entityName string, thresholdInfo string) (string, error) {
	promptCtx := &models.PromptContext{
		Arguments: map[string]any{
			"corruption_data": corruptionData,
			"entity_name":     entityName,
			"threshold_info":  thresholdInfo,
		},
		Timestamp: time.Now(),
	}

	result, err := s.loader.RenderPrompt("corruption_analysis", promptCtx)
	if err != nil {
		return "", fmt.Errorf("failed to render corruption analysis prompt: %w", err)
	}

	return result.RenderedText, nil
}

// RenderPrompt provides a generic way to render any prompt
func (s *Service) RenderPrompt(ctx context.Context, name string, arguments map[string]any) (string, error) {
	promptCtx := &models.PromptContext{
		Arguments: arguments,
		Timestamp: time.Now(),
	}

	result, err := s.loader.RenderPrompt(name, promptCtx)
	if err != nil {
		return "", fmt.Errorf("failed to render prompt %s: %w", name, err)
	}

	return result.RenderedText, nil
}

// ListAvailablePrompts returns information about all loaded prompts
func (s *Service) ListAvailablePrompts() map[string]*models.Prompt {
	return s.loader.ListPrompts()
}

// GetPromptInfo returns detailed information about a specific prompt
func (s *Service) GetPromptInfo(name string) (*models.Prompt, error) {
	return s.loader.GetPrompt(name)
}

// ReloadPrompts reloads all prompts if they have changed
func (s *Service) ReloadPrompts() error {
	return s.loader.ReloadIfChanged()
}

// ReloadSpecificPrompt reloads a specific prompt by name
func (s *Service) ReloadSpecificPrompt(name string) error {
	return s.loader.ReloadPrompt(name)
}

// ValidatePromptArguments checks if the provided arguments are valid for a prompt
func (s *Service) ValidatePromptArguments(name string, arguments map[string]any) error {
	prompt, err := s.loader.GetPrompt(name)
	if err != nil {
		return err
	}

	// Check required arguments
	for _, arg := range prompt.Arguments {
		if arg.Required {
			if _, exists := arguments[arg.Name]; !exists {
				return &models.ValidationError{
					Field:   arg.Name,
					Message: fmt.Sprintf("required argument '%s' is missing", arg.Name),
				}
			}
		}
	}

	return nil
}

// CreatePromptContext creates a prompt context with additional metadata
func (s *Service) CreatePromptContext(arguments map[string]any, userID, sessionID string, nodeData map[string]any) *models.PromptContext {
	return &models.PromptContext{
		Arguments: arguments,
		NodeData:  nodeData,
		UserID:    userID,
		SessionID: sessionID,
		Timestamp: time.Now(),
		Metadata: map[string]any{
			"service_version": "1.0.0",
			"context_created": time.Now(),
		},
	}
}
