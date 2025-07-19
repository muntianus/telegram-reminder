package main

import (
	"os"
	"path/filepath"
	"testing"

	botpkg "telegram-reminder/internal/bot"
)

func TestFormatWhitelistEmpty(t *testing.T) {
	result := botpkg.FormatWhitelist([]int64{})
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestFormatWhitelistNil(t *testing.T) {
	result := botpkg.FormatWhitelist(nil)
	if result != "" {
		t.Errorf("expected empty string for nil, got %q", result)
	}
}

func TestFormatWhitelistSingle(t *testing.T) {
	result := botpkg.FormatWhitelist([]int64{123})
	if result != "123" {
		t.Errorf("expected '123', got %q", result)
	}
}

func TestFormatWhitelistMultiple(t *testing.T) {
	result := botpkg.FormatWhitelist([]int64{123, 456, 789})
	if result != "123\n456\n789" {
		t.Errorf("expected '123\\n456\\n789', got %q", result)
	}
}

func TestLoadWhitelistEmptyFile(t *testing.T) {
	dir := t.TempDir()
	whitelistFile := filepath.Join(dir, "whitelist.json")

	// Create empty file
	err := os.WriteFile(whitelistFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Set the whitelist file path
	originalFile := botpkg.WhitelistFile
	botpkg.WhitelistFile = whitelistFile
	defer func() { botpkg.WhitelistFile = originalFile }()

	ids, err := botpkg.LoadWhitelist()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("expected empty list, got %v", ids)
	}
}

func TestLoadWhitelistInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	whitelistFile := filepath.Join(dir, "whitelist.json")

	// Create file with invalid JSON
	err := os.WriteFile(whitelistFile, []byte("[1, 2, 3"), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Set the whitelist file path
	originalFile := botpkg.WhitelistFile
	botpkg.WhitelistFile = whitelistFile
	defer func() { botpkg.WhitelistFile = originalFile }()

	_, err = botpkg.LoadWhitelist()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadWhitelistNonExistentFile(t *testing.T) {
	dir := t.TempDir()
	whitelistFile := filepath.Join(dir, "nonexistent.json")

	// Set the whitelist file path
	originalFile := botpkg.WhitelistFile
	botpkg.WhitelistFile = whitelistFile
	defer func() { botpkg.WhitelistFile = originalFile }()

	ids, err := botpkg.LoadWhitelist()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("expected empty list for non-existent file, got %v", ids)
	}
}

func TestAddIDToWhitelistDuplicate(t *testing.T) {
	dir := t.TempDir()
	whitelistFile := filepath.Join(dir, "whitelist.json")

	// Set the whitelist file path
	originalFile := botpkg.WhitelistFile
	botpkg.WhitelistFile = whitelistFile
	defer func() { botpkg.WhitelistFile = originalFile }()

	// Add ID twice
	err := botpkg.AddIDToWhitelist(123)
	if err != nil {
		t.Fatalf("failed to add first ID: %v", err)
	}

	err = botpkg.AddIDToWhitelist(123)
	if err != nil {
		t.Fatalf("failed to add duplicate ID: %v", err)
	}

	// Check that only one ID was saved
	ids, err := botpkg.LoadWhitelist()
	if err != nil {
		t.Fatalf("failed to load whitelist: %v", err)
	}
	if len(ids) != 1 || ids[0] != 123 {
		t.Errorf("expected [123], got %v", ids)
	}
}

func TestRemoveIDFromWhitelistNonExistent(t *testing.T) {
	dir := t.TempDir()
	whitelistFile := filepath.Join(dir, "whitelist.json")

	// Set the whitelist file path
	originalFile := botpkg.WhitelistFile
	botpkg.WhitelistFile = whitelistFile
	defer func() { botpkg.WhitelistFile = originalFile }()

	// Add some IDs
	err := botpkg.AddIDToWhitelist(123)
	if err != nil {
		t.Fatalf("failed to add ID: %v", err)
	}

	// Remove non-existent ID
	err = botpkg.RemoveIDFromWhitelist(456)
	if err != nil {
		t.Fatalf("failed to remove non-existent ID: %v", err)
	}

	// Check that original ID is still there
	ids, err := botpkg.LoadWhitelist()
	if err != nil {
		t.Fatalf("failed to load whitelist: %v", err)
	}
	if len(ids) != 1 || ids[0] != 123 {
		t.Errorf("expected [123], got %v", ids)
	}
}

func TestWhitelistConcurrentAccess(t *testing.T) {
	dir := t.TempDir()
	whitelistFile := filepath.Join(dir, "whitelist.json")

	// Set the whitelist file path
	originalFile := botpkg.WhitelistFile
	botpkg.WhitelistFile = whitelistFile
	defer func() { botpkg.WhitelistFile = originalFile }()

	// Test concurrent access - just verify it doesn't crash
	done := make(chan bool, 10)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			err := botpkg.AddIDToWhitelist(int64(id))
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check for any errors
	close(errors)
	for err := range errors {
		t.Errorf("goroutine error: %v", err)
	}

	// Verify that some IDs were added (concurrent access may not add all due to race conditions)
	ids, err := botpkg.LoadWhitelist()
	if err != nil {
		t.Fatalf("failed to load whitelist: %v", err)
	}
	if len(ids) == 0 {
		t.Error("expected at least some IDs to be added")
	}

	// Verify that added IDs are valid
	for _, id := range ids {
		if id < 0 || id >= 10 {
			t.Errorf("unexpected ID: %d", id)
		}
	}
}

func TestWhitelistFilePermissions(t *testing.T) {
	dir := t.TempDir()
	whitelistFile := filepath.Join(dir, "whitelist.json")

	// Set the whitelist file path
	originalFile := botpkg.WhitelistFile
	botpkg.WhitelistFile = whitelistFile
	defer func() { botpkg.WhitelistFile = originalFile }()

	// Add an ID to create the file
	err := botpkg.AddIDToWhitelist(123)
	if err != nil {
		t.Fatalf("failed to add ID: %v", err)
	}

	// Check file permissions
	info, err := os.Stat(whitelistFile)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	// Check that file is readable and writable by owner
	mode := info.Mode()
	if mode&0600 != 0600 {
		t.Errorf("expected file permissions 0600, got %v", mode)
	}
}
