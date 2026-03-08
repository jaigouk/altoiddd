package llm

import (
	"context"
	"fmt"
)

// NoopClient is an LLM client that always returns ErrLLMUnavailable.
// Used for graceful degradation when no provider is configured.
type NoopClient struct{}

// Compile-time interface check.
var _ Client = (*NoopClient)(nil)

// StructuredOutput always returns ErrLLMUnavailable.
func (n *NoopClient) StructuredOutput(_ context.Context, _ string, _ map[string]any) (Response, error) {
	return Response{}, fmt.Errorf("LLM service not configured: %w", ErrLLMUnavailable)
}

// TextCompletion always returns ErrLLMUnavailable.
func (n *NoopClient) TextCompletion(_ context.Context, _ string) (Response, error) {
	return Response{}, fmt.Errorf("LLM service not configured: %w", ErrLLMUnavailable)
}
