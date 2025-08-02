package bot

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	// Maximum lengths for user input to prevent DoS attacks
	MaxPayloadLength = 2000 // Maximum length for command payloads
	MaxQueryLength   = 1000 // Maximum length for search queries
	MaxChatLength    = 4000 // Maximum length for chat messages
)

// validateUserInput validates user input for basic security
func validateUserInput(input string, maxLength int, allowEmpty bool) error {
	if input == "" && !allowEmpty {
		return fmt.Errorf("input cannot be empty")
	}

	if !utf8.ValidString(input) {
		return fmt.Errorf("input contains invalid UTF-8 characters")
	}

	if utf8.RuneCountInString(input) > maxLength {
		return fmt.Errorf("input too long (max: %d characters)", maxLength)
	}

	return nil
}

// validatePayload validates command payload
func validatePayload(payload string) error {
	return validateUserInput(payload, MaxPayloadLength, true)
}

// validateQuery validates search query
func validateQuery(query string) error {
	if err := validateUserInput(query, MaxQueryLength, false); err != nil {
		return err
	}

	// Additional validation for search queries
	trimmed := strings.TrimSpace(query)
	if len(trimmed) < 2 {
		return fmt.Errorf("search query too short (minimum 2 characters)")
	}

	return nil
}

// validateChatMessage validates chat message
func validateChatMessage(message string) error {
	return validateUserInput(message, MaxChatLength, false)
}

// sanitizeInput removes potentially dangerous characters
func sanitizeInput(input string) string {
	// Remove null bytes and other control characters except newline and tab
	result := strings.Map(func(r rune) rune {
		if r == 0 || (r < 32 && r != '\n' && r != '\t') {
			return -1 // Remove character
		}
		return r
	}, input)

	return strings.TrimSpace(result)
}
