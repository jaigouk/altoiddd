package infrastructure_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	discoveryapp "github.com/alty-cli/alty/internal/discovery/application"
	discoveryinfra "github.com/alty-cli/alty/internal/discovery/infrastructure"
	"github.com/alty-cli/alty/internal/shared/infrastructure/llm"
)

// Compile-time interface check.
var _ discoveryapp.LLMDocReader = (*discoveryinfra.LLMDocReaderAdapter)(nil)

// --- Stub LLM Client ---

type stubLLMClient struct {
	response llm.Response
	err      error
	called   bool
	prompt   string
	schema   map[string]any
}

func (s *stubLLMClient) StructuredOutput(_ context.Context, prompt string, schema map[string]any) (llm.Response, error) {
	s.called = true
	s.prompt = prompt
	s.schema = schema
	return s.response, s.err
}

func (s *stubLLMClient) TextCompletion(_ context.Context, _ string) (llm.Response, error) {
	return llm.Response{}, nil
}

// --- Tests ---

func TestLLMDocReaderAdapter_InferModel_WhenValidDocs_CallsStructuredOutput(t *testing.T) {
	t.Parallel()

	// Given: stub client returns valid JSON with bounded contexts
	responseJSON := validInferenceJSON()
	client := &stubLLMClient{
		response: llm.NewResponse(responseJSON, "claude-test", 100),
	}
	adapter := discoveryinfra.NewLLMDocReaderAdapter(client)

	docs := map[string]string{
		"README.md": "# My Project\nA task management app.",
		"DDD.md":    "## Bounded Contexts\n### Tasks",
	}

	// When
	result, err := adapter.InferModel(context.Background(), docs)

	// Then
	require.NoError(t, err)
	assert.True(t, client.called)
	assert.Contains(t, client.prompt, "My Project")
	assert.Contains(t, client.prompt, "Bounded Contexts")
	assert.NotNil(t, result)
	assert.Equal(t, "high", result.Confidence())
	assert.ElementsMatch(t, []string{"README.md", "DDD.md"}, result.SourceDocs())
}

func TestLLMDocReaderAdapter_InferModel_ParsesJSONIntoModel(t *testing.T) {
	t.Parallel()

	// Given: LLM returns JSON with bounded contexts and domain stories
	responseJSON := validInferenceJSON()
	client := &stubLLMClient{
		response: llm.NewResponse(responseJSON, "claude-test", 100),
	}
	adapter := discoveryinfra.NewLLMDocReaderAdapter(client)

	docs := map[string]string{"README.md": "# Test"}

	// When
	result, err := adapter.InferModel(context.Background(), docs)

	// Then
	require.NoError(t, err)
	model := result.Model()
	assert.NotNil(t, model)

	contexts := model.BoundedContexts()
	assert.Len(t, contexts, 1)
	assert.Equal(t, "TaskManagement", contexts[0].Name())
}

func TestLLMDocReaderAdapter_InferModel_WhenLLMFails_ReturnsError(t *testing.T) {
	t.Parallel()

	client := &stubLLMClient{err: llm.ErrLLMUnavailable}
	adapter := discoveryinfra.NewLLMDocReaderAdapter(client)

	docs := map[string]string{"README.md": "# Test"}

	// When
	result, err := adapter.InferModel(context.Background(), docs)

	// Then
	require.ErrorIs(t, err, llm.ErrLLMUnavailable)
	assert.Nil(t, result)
}

func TestLLMDocReaderAdapter_InferModel_WhenMalformedResponse_ReturnsError(t *testing.T) {
	t.Parallel()

	client := &stubLLMClient{
		response: llm.NewResponse("not valid json {{{", "claude-test", 50),
	}
	adapter := discoveryinfra.NewLLMDocReaderAdapter(client)

	docs := map[string]string{"README.md": "# Test"}

	// When
	result, err := adapter.InferModel(context.Background(), docs)

	// Then
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "parsing LLM response")
}

func TestLLMDocReaderAdapter_InferModel_WhenEmptyContexts_ReturnsEmptyModel(t *testing.T) {
	t.Parallel()

	emptyJSON := `{"bounded_contexts": [], "domain_stories": [], "domain_events": [], "actors": [], "entities": []}`
	client := &stubLLMClient{
		response: llm.NewResponse(emptyJSON, "claude-test", 50),
	}
	adapter := discoveryinfra.NewLLMDocReaderAdapter(client)

	docs := map[string]string{"README.md": "# Empty"}

	// When
	result, err := adapter.InferModel(context.Background(), docs)

	// Then
	require.NoError(t, err)
	assert.True(t, result.Model().IsEmpty())
}

// validInferenceJSON returns a well-formed LLM response JSON.
func validInferenceJSON() string {
	resp := map[string]any{
		"bounded_contexts": []map[string]string{
			{
				"name":           "TaskManagement",
				"responsibility": "Manages task lifecycle",
				"classification": "core",
			},
		},
		"domain_stories": []map[string]any{
			{
				"title":   "Create Task",
				"actors":  []string{"User"},
				"trigger": "User creates a new task",
				"steps":   []string{"User fills form", "System validates", "Task created"},
			},
		},
		"domain_events": []string{"TaskCreated", "TaskCompleted"},
		"actors":        []string{"User", "Admin"},
		"entities": []map[string]string{
			{"name": "Task", "type": "aggregate"},
		},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal test JSON: %v", err))
	}
	return string(data)
}
