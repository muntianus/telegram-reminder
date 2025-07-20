package main

import (
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

func TestAddIDToWhitelistDuplicate(t *testing.T) {
	botpkg.ResetWhitelist()
	if err := botpkg.AddIDToWhitelist(123); err != nil {
		t.Fatalf("failed to add first ID: %v", err)
	}
	if err := botpkg.AddIDToWhitelist(123); err != nil {
		t.Fatalf("failed to add duplicate ID: %v", err)
	}
	ids, err := botpkg.LoadWhitelist()
	if err != nil {
		t.Fatalf("failed to load whitelist: %v", err)
	}
	if len(ids) != 1 || ids[0] != 123 {
		t.Errorf("expected [123], got %v", ids)
	}
}

func TestRemoveIDFromWhitelistNonExistent(t *testing.T) {
	botpkg.ResetWhitelist()
	if err := botpkg.AddIDToWhitelist(123); err != nil {
		t.Fatalf("failed to add ID: %v", err)
	}
	if err := botpkg.RemoveIDFromWhitelist(456); err != nil {
		t.Fatalf("failed to remove non-existent ID: %v", err)
	}
	ids, err := botpkg.LoadWhitelist()
	if err != nil {
		t.Fatalf("failed to load whitelist: %v", err)
	}
	if len(ids) != 1 || ids[0] != 123 {
		t.Errorf("expected [123], got %v", ids)
	}
}

func TestWhitelistConcurrentAccess(t *testing.T) {
	botpkg.ResetWhitelist()
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

	for i := 0; i < 10; i++ {
		<-done
	}

	close(errors)
	for err := range errors {
		t.Errorf("goroutine error: %v", err)
	}

	ids, err := botpkg.LoadWhitelist()
	if err != nil {
		t.Fatalf("failed to load whitelist: %v", err)
	}
	if len(ids) == 0 {
		t.Error("expected at least some IDs to be added")
	}
	for _, id := range ids {
		if id < 0 || id >= 10 {
			t.Errorf("unexpected ID: %d", id)
		}
	}
}
