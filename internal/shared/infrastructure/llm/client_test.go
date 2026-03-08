package llm_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alty-cli/alty/internal/shared/infrastructure/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// LLMProvider enum
// ---------------------------------------------------------------------------

func TestAllProviders(t *testing.T) {
	t.Parallel()
	providers := llm.AllProviders()
	assert.Len(t, providers, 4)
}

func TestProviderValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		provider llm.LLMProvider
		want     string
	}{
		{"anthropic", llm.ProviderAnthropic, "anthropic"},
		{"ollama", llm.ProviderOllama, "ollama"},
		{"vertexai", llm.ProviderVertexAI, "vertexai"},
		{"none", llm.ProviderNone, "none"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, string(tt.provider))
		})
	}
}

// ---------------------------------------------------------------------------
// Config
// ---------------------------------------------------------------------------

func TestConfigDefaults(t *testing.T) {
	t.Parallel()
	cfg := llm.DefaultConfig()
	assert.Equal(t, llm.ProviderNone, cfg.Provider())
	assert.Equal(t, "", cfg.Model())
	assert.Equal(t, "", cfg.APIKey())
	assert.Equal(t, 30.0, cfg.Timeout())
}

func TestConfigCustomValues(t *testing.T) {
	t.Parallel()
	cfg := llm.NewConfig(llm.ProviderAnthropic, "claude-sonnet-4-20250514", "sk-test", 60.0)
	assert.Equal(t, llm.ProviderAnthropic, cfg.Provider())
	assert.Equal(t, "claude-sonnet-4-20250514", cfg.Model())
	assert.Equal(t, "sk-test", cfg.APIKey())
	assert.Equal(t, 60.0, cfg.Timeout())
}

func TestConfigStringMasksAPIKey(t *testing.T) {
	t.Parallel()
	cfg := llm.NewConfig(llm.ProviderAnthropic, "claude-sonnet-4-20250514", "sk-secret-key-12345", 30.0)
	s := cfg.String()
	assert.NotContains(t, s, "sk-secret-key-12345")
	assert.Contains(t, s, "***")
}

func TestConfigStringShowsEmptyWhenNoAPIKey(t *testing.T) {
	t.Parallel()
	cfg := llm.NewConfig(llm.ProviderNone, "", "", 30.0)
	s := cfg.String()
	assert.NotContains(t, s, "***")
}

func TestConfigStringIncludesProviderAndModel(t *testing.T) {
	t.Parallel()
	cfg := llm.NewConfig(llm.ProviderAnthropic, "claude-sonnet-4-20250514", "sk-test", 30.0)
	s := cfg.String()
	assert.Contains(t, s, "anthropic")
	assert.Contains(t, s, "claude-sonnet-4-20250514")
}

// ---------------------------------------------------------------------------
// Response
// ---------------------------------------------------------------------------

func TestResponseFields(t *testing.T) {
	t.Parallel()
	resp := llm.NewResponse("hello", "claude", 42)
	assert.Equal(t, "hello", resp.Content())
	assert.Equal(t, "claude", resp.ModelUsed())
	assert.Equal(t, 42, resp.UsageTokens())
}

// ---------------------------------------------------------------------------
// NoopClient
// ---------------------------------------------------------------------------

func TestNoopClientImplementsClient(t *testing.T) {
	t.Parallel()
	var _ llm.Client = (*llm.NoopClient)(nil)
}

func TestNoopStructuredOutputReturnsLLMUnavailable(t *testing.T) {
	t.Parallel()
	client := &llm.NoopClient{}
	_, err := client.StructuredOutput(context.Background(), "test", map[string]any{"type": "object"})
	require.Error(t, err)
	assert.True(t, errors.Is(err, llm.ErrLLMUnavailable))
	assert.Contains(t, err.Error(), "not configured")
}

func TestNoopTextCompletionReturnsLLMUnavailable(t *testing.T) {
	t.Parallel()
	client := &llm.NoopClient{}
	_, err := client.TextCompletion(context.Background(), "test")
	require.Error(t, err)
	assert.True(t, errors.Is(err, llm.ErrLLMUnavailable))
	assert.Contains(t, err.Error(), "not configured")
}

// ---------------------------------------------------------------------------
// Factory
// ---------------------------------------------------------------------------

func TestFactoryCreatesAnthropicWhenConfigured(t *testing.T) {
	t.Parallel()
	cfg := llm.NewConfig(llm.ProviderAnthropic, "claude-sonnet-4-20250514", "sk-test-key", 30.0)
	factory := llm.Factory{}
	client := factory.Create(cfg)
	_, ok := client.(*llm.AnthropicClient)
	assert.True(t, ok, "expected AnthropicClient")
}

func TestFactoryCreatesNoopForNoneProvider(t *testing.T) {
	t.Parallel()
	cfg := llm.NewConfig(llm.ProviderNone, "", "", 30.0)
	factory := llm.Factory{}
	client := factory.Create(cfg)
	_, ok := client.(*llm.NoopClient)
	assert.True(t, ok, "expected NoopClient")
}

func TestFactoryCreatesNoopWhenNoAPIKey(t *testing.T) {
	t.Parallel()
	cfg := llm.NewConfig(llm.ProviderAnthropic, "claude-sonnet-4-20250514", "", 30.0)
	factory := llm.Factory{}
	client := factory.Create(cfg)
	_, ok := client.(*llm.NoopClient)
	assert.True(t, ok, "expected NoopClient for empty API key")
}

func TestFactoryDefaultsToNoop(t *testing.T) {
	t.Parallel()
	cfg := llm.DefaultConfig()
	factory := llm.Factory{}
	client := factory.Create(cfg)
	_, ok := client.(*llm.NoopClient)
	assert.True(t, ok, "expected NoopClient for default config")
}

func TestFactoryCreatesNoopForUnsupportedProvider(t *testing.T) {
	t.Parallel()
	cfg := llm.NewConfig(llm.ProviderOllama, "", "key", 30.0)
	factory := llm.Factory{}
	client := factory.Create(cfg)
	_, ok := client.(*llm.NoopClient)
	assert.True(t, ok, "expected NoopClient for unsupported provider")
}

// ---------------------------------------------------------------------------
// AnthropicClient compile-time check
// ---------------------------------------------------------------------------

func TestAnthropicClientImplementsClient(t *testing.T) {
	t.Parallel()
	var _ llm.Client = (*llm.AnthropicClient)(nil)
}
