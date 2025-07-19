package bot

import (
	"os"
	"testing"
)

func TestLoadWhitelist_Empty(t *testing.T) {
	dir := t.TempDir()
	file := dir + "/wl.json"
	os.WriteFile(file, []byte(""), 0644)
	old := WhitelistFile
	WhitelistFile = file
	defer func() { WhitelistFile = old }()
	ids, err := LoadWhitelist()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("expected empty list, got %v", ids)
	}
}

func TestAddIDToWhitelist(t *testing.T) {
	dir := t.TempDir()
	file := dir + "/wl.json"
	os.WriteFile(file, []byte("[1,2]"), 0644)
	old := WhitelistFile
	WhitelistFile = file
	defer func() { WhitelistFile = old }()
	if err := AddIDToWhitelist(3); err != nil {
		t.Fatalf("add id: %v", err)
	}
	ids, _ := LoadWhitelist()
	if len(ids) != 3 || ids[2] != 3 {
		t.Errorf("unexpected ids: %v", ids)
	}
	// Добавление существующего не должно дублировать
	if err := AddIDToWhitelist(3); err != nil {
		t.Fatalf("add existing: %v", err)
	}
	ids, _ = LoadWhitelist()
	if len(ids) != 3 {
		t.Errorf("should not duplicate id: %v", ids)
	}
}

func TestRemoveIDFromWhitelist(t *testing.T) {
	dir := t.TempDir()
	file := dir + "/wl.json"
	os.WriteFile(file, []byte("[1,2,3]"), 0644)
	old := WhitelistFile
	WhitelistFile = file
	defer func() { WhitelistFile = old }()
	if err := RemoveIDFromWhitelist(2); err != nil {
		t.Fatalf("remove id: %v", err)
	}
	ids, _ := LoadWhitelist()
	if len(ids) != 2 || ids[0] != 1 || ids[1] != 3 {
		t.Errorf("unexpected ids after remove: %v", ids)
	}
	// Удаление несуществующего не должно ломать
	if err := RemoveIDFromWhitelist(99); err != nil {
		t.Fatalf("remove non-existent: %v", err)
	}
	ids, _ = LoadWhitelist()
	if len(ids) != 2 {
		t.Errorf("should not change: %v", ids)
	}
}

func TestFormatWhitelist(t *testing.T) {
	ids := []int64{5, 7}
	out := FormatWhitelist(ids)
	if out != "5\n7" {
		t.Errorf("unexpected format: %q", out)
	}
	if FormatWhitelist(nil) != "" {
		t.Errorf("expected empty string")
	}
}
