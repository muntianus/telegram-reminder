package services

import (
	"context"
	"testing"
	"time"

	"telegram-reminder/internal/domain"
)

// MockAIClient implements AIClient for testing
type MockAIClient struct {
	response string
	err      error
}

func (m *MockAIClient) EnhancedSystemCompletion(ctx context.Context, prompt, model string) (string, error) {
	return m.response, m.err
}

func TestDigestService_GenerateDigest(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   string
		mockError      error
		request        DigestRequest
		expectedError  bool
		expectedLength int
	}{
		{
			name:         "successful crypto digest generation",
			mockResponse: "Test crypto digest content",
			mockError:    nil,
			request: DigestRequest{
				Type:   domain.CryptoDigest,
				Model:  "gpt-4.1",
				ChatID: 12345,
			},
			expectedError:  false,
			expectedLength: 26,
		},
		{
			name:         "unknown digest type",
			mockResponse: "",
			mockError:    nil,
			request: DigestRequest{
				Type:   domain.DigestType("unknown"),
				Model:  "gpt-4.1",
				ChatID: 12345,
			},
			expectedError: true,
		},
		{
			name:         "AI client error",
			mockResponse: "",
			mockError:    NewServiceError("AI error"),
			request: DigestRequest{
				Type:   domain.TechDigest,
				Model:  "gpt-4.1",
				ChatID: 12345,
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock AI client
			mockClient := &MockAIClient{
				response: tt.mockResponse,
				err:      tt.mockError,
			}

			// Create digest service
			service := NewDigestService(mockClient, 5*time.Second)

			// Generate digest
			resp, err := service.GenerateDigest(context.Background(), tt.request)

			// Check error expectation
			if tt.expectedError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check response
			if !tt.expectedError && resp != nil {
				if len(resp.Content) != tt.expectedLength {
					t.Errorf("expected content length %d, got %d", tt.expectedLength, len(resp.Content))
				}
				if resp.Type != tt.request.Type {
					t.Errorf("expected type %v, got %v", tt.request.Type, resp.Type)
				}
			}
		})
	}
}

func TestDigestService_GetAvailableDigests(t *testing.T) {
	mockClient := &MockAIClient{}
	service := NewDigestService(mockClient, 5*time.Second)

	digests := service.GetAvailableDigests()

	// Check that all expected digest types are available
	expectedTypes := []domain.DigestType{
		domain.CryptoDigest,
		domain.TechDigest,
		domain.RealEstateDigest,
		domain.BusinessDigest,
		domain.InvestmentDigest,
		domain.StartupDigest,
		domain.GlobalDigest,
	}

	for _, expectedType := range expectedTypes {
		if _, exists := digests[expectedType]; !exists {
			t.Errorf("expected digest type %v not found", expectedType)
		}
	}

	// Check that each digest has required fields
	for digestType, config := range digests {
		if config.Type != digestType {
			t.Errorf("digest type mismatch: expected %v, got %v", digestType, config.Type)
		}
		if config.Name == "" {
			t.Errorf("digest %v has empty name", digestType)
		}
		if config.CommandName == "" {
			t.Errorf("digest %v has empty command name", digestType)
		}
		if config.Prompt == "" {
			t.Errorf("digest %v has empty prompt", digestType)
		}
	}
}