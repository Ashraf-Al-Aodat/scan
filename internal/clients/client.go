package clients

import (
	"context"
	"fmt"
	"net/http"
	"scan/internal/prompts"

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
	messageHistory := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}

	resp, err := c.llm.GenerateContent(ctx, messageHistory, prompts.Tools())
	if err != nil {
		return "", err
	}

	messageHistory = prompts.UpdateMessageHistory(messageHistory, resp)

	// Execute tool calls requested by the model
	messageHistory = prompts.ExecuteToolCalls(ctx, c.llm, messageHistory, resp)
	// Assert part is a ToolCallResponse
	if toolResp, ok := messageHistory[len(messageHistory)-1].Parts[0].(llms.ToolCallResponse); ok {
		return fmt.Sprint(toolResp.Content), nil
	}
	return "", err
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

	// Fix: remove default OpenAI "application/x-www-form-urlencoded"
	//req.Header.Del("Content-Type")

	// Fix: override Host header if required (optional)
	// req.Host = c.host

	return c.rt.RoundTrip(req)
}
