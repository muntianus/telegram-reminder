package main

import (
	"path/filepath"
	"testing"

	botpkg "telegram-reminder/internal/bot"
)

func TestWhitelistAddRemove(t *testing.T) {
	dir := t.TempDir()
	botpkg.WhitelistFile = filepath.Join(dir, "wl.json")

	if err := botpkg.AddIDToWhitelist(10); err != nil {
		t.Fatalf("add id: %v", err)
	}
	if err := botpkg.AddIDToWhitelist(10); err != nil {
		t.Fatalf("add duplicate: %v", err)
	}
	if err := botpkg.AddIDToWhitelist(20); err != nil {
		t.Fatalf("add second: %v", err)
	}

	ids, err := botpkg.LoadWhitelist()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	want := []int64{10, 20}
	if len(ids) != len(want) || ids[0] != want[0] || ids[1] != want[1] {
		t.Fatalf("unexpected ids: %v", ids)
	}

	if err := botpkg.RemoveIDFromWhitelist(10); err != nil {
		t.Fatalf("remove: %v", err)
	}

	ids, err = botpkg.LoadWhitelist()
	if err != nil {
		t.Fatalf("load after remove: %v", err)
	}
	if len(ids) != 1 || ids[0] != 20 {
		t.Fatalf("unexpected ids after remove: %v", ids)
	}
}

func TestFormatWhitelist(t *testing.T) {
	got := botpkg.FormatWhitelist([]int64{5, 7})
	if got != "5\n7" {
		t.Errorf("unexpected format: %q", got)
	}
	if botpkg.FormatWhitelist(nil) != "" {
		t.Errorf("expected empty string")
	}
}
