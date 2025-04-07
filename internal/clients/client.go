package clients

import (
	"context"
	"net/http"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// Client represents the LLM client.
type Client struct {
	llm llms.Model
}

// NewClient initializes and returns a new LLM client.
func NewClient(host, model string, headers map[string]string) *Client {
	httpClient := &http.Client{
		Transport: &customTransport{
			headers: headers,
			rt:      http.DefaultTransport,
		},
	}
	llm, err := openai.New(
		openai.WithBaseURL(host),
		openai.WithModel(model),
		openai.WithHTTPClient(httpClient),
	)
	if err != nil {
		panic(err)
	}
	return &Client{llm: llm}
}

// GenerateResponse generates a response from the LLM based on the given prompt.
func (c *Client) GenerateResponse(ctx context.Context, prompt string) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, c.llm, prompt)
}

type customTransport struct {
	headers map[string]string
	rt      http.RoundTripper
}

func (c *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Scan/0.1")

	return c.rt.RoundTrip(req)
}
