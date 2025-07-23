package openai

import (
	"context"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/responses"
	"github.com/openai/openai-go/shared"
)

// Config holds API key and model names for the client.
type Config struct {
	APIKey      string
	ModelDigest string
	ModelSearch string
	BaseURL     string // optional, used in tests
}

// Client wraps the OpenAI SDK client.
type Client struct {
	raw    *openai.Client
	config Config
}

// New creates a new Client with the given configuration.
func New(cfg Config) *Client {
	opts := []option.RequestOption{
		option.WithAPIKey(cfg.APIKey),
	}
	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}
	cli := openai.NewClient(opts...)
	return &Client{raw: &cli, config: cfg}
}

// Search performs a web search using the Responses API and returns the text.
func (c *Client) Search(ctx context.Context, query string) (string, error) {
	resp, err := c.raw.Responses.New(ctx, responses.ResponseNewParams{
		Model: shared.ResponsesModel(c.config.ModelSearch),
		Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(query)},
		Tools: []responses.ToolUnionParam{{
			OfWebSearchPreview: &responses.WebSearchToolParam{Type: responses.WebSearchToolType("web_search")},
		}},
		Temperature: openai.Float(0.0),
	})
	if err != nil {
		return "", err
	}
	return resp.OutputText(), nil
}

// Raw exposes the underlying SDK client (used in tests).
func (c *Client) Raw() *openai.Client { return c.raw }
