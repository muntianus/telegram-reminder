package bot

import (
	"context"
	"testing"
)

func TestOpenAISearchFormatting(t *testing.T) {
	origAPI := searchAPIFunc
	searchAPIFunc = func(ctx context.Context, q string) (string, error) {
		return "**hi** [site](https://example.com)", nil
	}
	defer func() { searchAPIFunc = origAPI }()

	searchMu.Lock()
	searchCache = map[string]searchEntry{}
	searchMu.Unlock()

	out, err := OpenAISearch("test")
	if err != nil {
		t.Fatalf("search error: %v", err)
	}
	expected := "<b>hi</b> <a href=\"https://example.com\">site</a>"
	if out != expected {
		t.Fatalf("unexpected output: %q", out)
	}
}
