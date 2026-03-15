// Package llm provides LLM client infrastructure: interface, factory, and adapters.
package llm

import (
	"context"
	"fmt"
)

// Provider enumerates supported LLM provider backends.
type Provider string

// LLM provider constants.
const (
	ProviderAnthropic        Provider = "anthropic"
	ProviderOllama           Provider = "ollama"
	ProviderVertexAI         Provider = "vertexai"
	ProviderOpenAI           Provider = "openai"
	ProviderOpenAICompatible Provider = "openai_compatible"
	ProviderNone             Provider = "none"
)

// AllProviders returns all valid Provider values.
func AllProviders() []Provider {
	return []Provider{
		ProviderAnthropic, ProviderOllama, ProviderVertexAI,
		ProviderOpenAI, ProviderOpenAICompatible, ProviderNone,
	}
}

// Config holds configuration for LLM client creation.
type Config struct {
	provider Provider
	model    string
	apiKey   string
	baseURL  string
	timeout  float64
}

// NewConfig creates a Config with the given values.
func NewConfig(provider Provider, model, apiKey, baseURL string, timeout float64) Config {
	return Config{provider: provider, model: model, apiKey: apiKey, baseURL: baseURL, timeout: timeout}
}

// DefaultConfig returns a Config with no provider (graceful degradation).
func DefaultConfig() Config {
	return Config{provider: ProviderNone, timeout: 30.0}
}

// Provider returns the LLM provider.
func (c Config) Provider() Provider { return c.provider }

// Model returns the model name.
func (c Config) Model() string { return c.model }

// APIKey returns the API key.
func (c Config) APIKey() string { return c.apiKey }

// BaseURL returns the base URL for the provider API.
func (c Config) BaseURL() string { return c.baseURL }

// Timeout returns the timeout in seconds.
func (c Config) Timeout() float64 { return c.timeout }

// String returns a safe representation that masks the API key.
func (c Config) String() string {
	masked := ""
	if c.apiKey != "" {
		masked = "***"
	}
	s := fmt.Sprintf("Config(provider=%s, model=%q, api_key=%q, timeout=%.1f",
		c.provider, c.model, masked, c.timeout)
	if c.baseURL != "" {
		s += fmt.Sprintf(", base_url=%q", c.baseURL)
	}
	s += ")"
	return s
}

// Response is the result of an LLM call.
type Response struct {
	content     string
	modelUsed   string
	usageTokens int
}

// NewResponse creates a Response value object.
func NewResponse(content, modelUsed string, usageTokens int) Response {
	return Response{content: content, modelUsed: modelUsed, usageTokens: usageTokens}
}

// Content returns the response content.
func (r Response) Content() string { return r.content }

// ModelUsed returns the model that generated the response.
func (r Response) ModelUsed() string { return r.modelUsed }

// UsageTokens returns the total token usage.
func (r Response) UsageTokens() int { return r.usageTokens }

// Client is a provider-agnostic LLM client interface.
type Client interface {
	// StructuredOutput sends a prompt with a JSON schema and returns structured output.
	StructuredOutput(ctx context.Context, prompt string, schema map[string]any) (Response, error)

	// TextCompletion sends a prompt and returns a text completion.
	TextCompletion(ctx context.Context, prompt string) (Response, error)
}
