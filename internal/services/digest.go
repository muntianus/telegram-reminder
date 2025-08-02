package services

import (
	"context"
	"time"

	"telegram-reminder/internal/domain"
	"telegram-reminder/internal/logger"
)

// AIClient defines the interface for AI completion services
type AIClient interface {
	EnhancedSystemCompletion(ctx context.Context, prompt, model string) (string, error)
}

// DigestService handles digest generation and management
type DigestService struct {
	aiClient AIClient
	timeout  time.Duration
}

// NewDigestService creates a new digest service
func NewDigestService(aiClient AIClient, timeout time.Duration) *DigestService {
	return &DigestService{
		aiClient: aiClient,
		timeout:  timeout,
	}
}

// DigestRequest contains parameters for digest generation
type DigestRequest struct {
	Type         domain.DigestType
	Model        string
	ChatID       int64
	TemplateName string
}

// DigestResponse contains the generated digest
type DigestResponse struct {
	Content string
	Type    domain.DigestType
	Error   error
}

// GenerateDigest generates a digest of the specified type
func (s *DigestService) GenerateDigest(ctx context.Context, req DigestRequest) (*DigestResponse, error) {
	logger.L.Debug("generating digest", "type", req.Type, "chat", req.ChatID)

	configs := domain.GetDigestConfigs()
	config, exists := configs[req.Type]
	if !exists {
		logger.L.Error("unknown digest type", "type", req.Type)
		return nil, ErrUnknownDigestType
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Apply template to prompt
	prompt := s.applyTemplate(config.Prompt, req.Model, req.TemplateName)

	// Generate completion
	content, err := s.generateCompletion(ctx, prompt, req.Model)
	if err != nil {
		logger.L.Error("digest generation failed", "type", req.Type, "model", req.Model, "error", err)
		return &DigestResponse{
			Type:  req.Type,
			Error: err,
		}, err
	}

	logger.L.Debug("digest generated successfully", "type", req.Type, "length", len(content))
	
	return &DigestResponse{
		Content: content,
		Type:    req.Type,
	}, nil
}

// GetAvailableDigests returns all available digest types
func (s *DigestService) GetAvailableDigests() map[domain.DigestType]domain.DigestConfig {
	return domain.GetDigestConfigs()
}

// generateCompletion generates AI completion for the given prompt
func (s *DigestService) generateCompletion(ctx context.Context, prompt, model string) (string, error) {
	return s.aiClient.EnhancedSystemCompletion(ctx, prompt, model)
}

// applyTemplate applies template variables to the prompt
func (s *DigestService) applyTemplate(prompt, model, templateName string) string {
	// Apply template substitutions
	vars := map[string]string{
		"date":  time.Now().Format("2006-01-02"),
		"model": model,
	}
	
	result := prompt
	for key, value := range vars {
		// Simple string replacement - in production, use a proper template engine
		result = replaceAll(result, "{"+key+"}", value)
	}
	
	return result
}

// Simple string replacement function
func replaceAll(s, old, new string) string {
	for {
		replaced := replace(s, old, new)
		if replaced == s {
			break
		}
		s = replaced
	}
	return s
}

func replace(s, old, new string) string {
	// Simple implementation - in production use strings.ReplaceAll
	if len(old) == 0 {
		return s
	}
	
	for i := 0; i <= len(s)-len(old); i++ {
		if s[i:i+len(old)] == old {
			return s[:i] + new + s[i+len(old):]
		}
	}
	return s
}

// Errors
var (
	ErrUnknownDigestType = NewServiceError("unknown digest type")
	ErrNotImplemented    = NewServiceError("not implemented")
)

// ServiceError represents a service-level error
type ServiceError struct {
	Message string
}

func (e ServiceError) Error() string {
	return e.Message
}

// NewServiceError creates a new service error
func NewServiceError(message string) error {
	return ServiceError{Message: message}
}