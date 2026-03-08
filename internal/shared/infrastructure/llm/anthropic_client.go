package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	anthropicAPIURL     = "https://api.anthropic.com/v1/messages"
	anthropicAPIVersion = "2023-06-01"
	defaultModel        = "claude-sonnet-4-20250514"
	defaultMaxTokens    = 4096
)

// AnthropicClient implements Client using the Anthropic Messages API via HTTP.
type AnthropicClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// Compile-time interface check.
var _ Client = (*AnthropicClient)(nil)

// NewAnthropicClient creates an AnthropicClient.
func NewAnthropicClient(apiKey, model string, timeout float64) *AnthropicClient {
	if model == "" {
		model = defaultModel
	}
	if timeout <= 0 {
		timeout = 30.0
	}
	return &AnthropicClient{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout * float64(time.Second)),
		},
	}
}

// anthropicRequest is the Anthropic Messages API request body.
type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicResponse is the Anthropic Messages API response body.
type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Model string `json:"model"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// StructuredOutput sends a prompt with a JSON schema instruction and returns the response.
func (a *AnthropicClient) StructuredOutput(ctx context.Context, prompt string, schema map[string]any) (Response, error) {
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return Response{}, fmt.Errorf("marshaling schema: %w", err)
	}
	systemPrompt := "Respond with valid JSON matching this schema:\n" + string(schemaJSON)
	return a.doRequest(ctx, prompt, systemPrompt)
}

// TextCompletion sends a prompt and returns a text completion.
func (a *AnthropicClient) TextCompletion(ctx context.Context, prompt string) (Response, error) {
	return a.doRequest(ctx, prompt, "")
}

func (a *AnthropicClient) doRequest(ctx context.Context, prompt, system string) (Response, error) {
	reqBody := anthropicRequest{
		Model:     a.model,
		MaxTokens: defaultMaxTokens,
		System:    system,
		Messages: []anthropicMessage{
			{Role: "user", Content: prompt},
		},
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return Response{}, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPIURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return Response{}, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", anthropicAPIVersion)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("%w: %s", ErrLLMUnavailable, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("%w: reading response: %s", ErrLLMUnavailable, err)
	}

	if resp.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("%w: API returned status %d: %s",
			ErrLLMUnavailable, resp.StatusCode, string(respBody))
	}

	var apiResp anthropicResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return Response{}, fmt.Errorf("%w: parsing response: %s", ErrLLMUnavailable, err)
	}

	if len(apiResp.Content) == 0 {
		return Response{}, fmt.Errorf("%w: empty content in response", ErrLLMUnavailable)
	}

	return NewResponse(
		apiResp.Content[0].Text,
		apiResp.Model,
		apiResp.Usage.InputTokens+apiResp.Usage.OutputTokens,
	), nil
}
