package bot

import (
	"os"
	"testing"
)

func TestWhitelist_BadJSON(t *testing.T) {
	dir := t.TempDir()
	badFile := dir + "/bad.json"
	os.WriteFile(badFile, []byte("not valid json"), 0644)
	old := WhitelistFile
	WhitelistFile = badFile
	defer func() { WhitelistFile = old }()
	_, err := LoadWhitelist()
	if err == nil {
		t.Fatal("expected error for bad json")
	}
}

func TestWhitelist_NoPerm(t *testing.T) {
	dir := t.TempDir()
	file := dir + "/wl.json"
	os.WriteFile(file, []byte("[1,2,3]"), 0000)
	old := WhitelistFile
	WhitelistFile = file
	defer func() { WhitelistFile = old }()
	_, err := LoadWhitelist()
	if err == nil {
		t.Fatal("expected error for no permissions")
	}
}

func TestWhitelist_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	emptyFile := dir + "/empty.json"
	os.WriteFile(emptyFile, []byte(""), 0644)
	old := WhitelistFile
	WhitelistFile = emptyFile
	defer func() { WhitelistFile = old }()
	ids, err := LoadWhitelist()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 0 {
		t.Fatalf("expected empty list, got %v", ids)
	}
}
