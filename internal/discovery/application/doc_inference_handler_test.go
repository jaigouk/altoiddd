package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	discoveryapp "github.com/alto-cli/alto/internal/discovery/application"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	"github.com/alto-cli/alto/internal/shared/infrastructure/llm"
)

// --- Mocks ---

type mockDocReader struct {
	docs map[string]string
	err  error
}

func (m *mockDocReader) ReadDocs(_ context.Context, _ string) (map[string]string, error) {
	return m.docs, m.err
}

type mockLLMDocReader struct {
	result *discoverydomain.InferenceResult
	err    error
	called bool
	docs   map[string]string
}

func (m *mockLLMDocReader) InferModel(_ context.Context, docs map[string]string) (*discoverydomain.InferenceResult, error) {
	m.called = true
	m.docs = docs
	return m.result, m.err
}

type mockRegexImporter struct {
	model  *ddd.DomainModel
	err    error
	called bool
}

func (m *mockRegexImporter) Import(_ context.Context, _ string) (*ddd.DomainModel, error) {
	m.called = true
	return m.model, m.err
}

// --- Tests ---

func TestDocInferenceHandler_InferFromDocs_WhenLLMAvailable_ReturnsInferredModel(t *testing.T) {
	t.Parallel()

	// Given: doc reader returns files, LLM reader returns a result
	model := ddd.NewDomainModel("inferred")
	inferenceResult, err := discoverydomain.NewInferenceResult(model, "high", []string{"README.md"})
	require.NoError(t, err)

	docReader := &mockDocReader{docs: map[string]string{"README.md": "# My Project"}}
	reader := &mockLLMDocReader{result: inferenceResult}
	regex := &mockRegexImporter{}
	handler := discoveryapp.NewDocInferenceHandler(docReader, reader, regex)

	// When
	result, err := handler.InferFromDocs(context.Background(), "/any/dir")

	// Then
	require.NoError(t, err)
	assert.Equal(t, "high", result.Confidence())
	assert.True(t, reader.called)
	assert.False(t, regex.called)
	assert.Contains(t, reader.docs, "README.md")
}

func TestDocInferenceHandler_InferFromDocs_WhenLLMUnavailable_FallsBackToRegex(t *testing.T) {
	t.Parallel()

	// Given: doc reader returns files, LLM returns ErrLLMUnavailable
	model := ddd.NewDomainModel("regex-parsed")
	docReader := &mockDocReader{docs: map[string]string{"DDD.md": "# DDD"}}
	reader := &mockLLMDocReader{err: llm.ErrLLMUnavailable}
	regex := &mockRegexImporter{model: model}
	handler := discoveryapp.NewDocInferenceHandler(docReader, reader, regex)

	// When
	result, err := handler.InferFromDocs(context.Background(), "/any/dir")

	// Then
	require.NoError(t, err)
	assert.Equal(t, "low", result.Confidence())
	assert.True(t, reader.called)
	assert.True(t, regex.called)
}

func TestDocInferenceHandler_InferFromDocs_WhenNoDocs_ReturnsError(t *testing.T) {
	t.Parallel()

	// Given: doc reader returns empty map
	docReader := &mockDocReader{docs: map[string]string{}}
	reader := &mockLLMDocReader{}
	regex := &mockRegexImporter{}
	handler := discoveryapp.NewDocInferenceHandler(docReader, reader, regex)

	// When
	result, err := handler.InferFromDocs(context.Background(), "/any/dir")

	// Then
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no documentation found")
}

func TestDocInferenceHandler_InferFromDocs_WhenDocReaderFails_ReturnsError(t *testing.T) {
	t.Parallel()

	// Given: doc reader returns error
	docReader := &mockDocReader{err: errors.New("directory not found")}
	reader := &mockLLMDocReader{}
	regex := &mockRegexImporter{}
	handler := discoveryapp.NewDocInferenceHandler(docReader, reader, regex)

	// When
	result, err := handler.InferFromDocs(context.Background(), "/nonexistent/path")

	// Then
	require.Error(t, err)
	assert.Nil(t, result)
}

func TestDocInferenceHandler_InferFromDocs_WhenLLMFailsNonUnavailable_ReturnsError(t *testing.T) {
	t.Parallel()

	// Given: LLM fails with a non-unavailable error (should not fallback)
	docReader := &mockDocReader{docs: map[string]string{"README.md": "# Test"}}
	reader := &mockLLMDocReader{err: errors.New("malformed response")}
	regex := &mockRegexImporter{}
	handler := discoveryapp.NewDocInferenceHandler(docReader, reader, regex)

	// When
	result, err := handler.InferFromDocs(context.Background(), "/any/dir")

	// Then
	require.Error(t, err)
	assert.Nil(t, result)
	assert.False(t, regex.called, "should not fallback on non-unavailable errors")
}

func TestDocInferenceHandler_InferFromDocs_WhenRegexFallbackFails_ReturnsError(t *testing.T) {
	t.Parallel()

	// Given: LLM unavailable AND regex fallback also fails
	docReader := &mockDocReader{docs: map[string]string{"README.md": "# Test"}}
	reader := &mockLLMDocReader{err: llm.ErrLLMUnavailable}
	regex := &mockRegexImporter{err: errors.New("no bounded contexts found")}
	handler := discoveryapp.NewDocInferenceHandler(docReader, reader, regex)

	// When
	result, err := handler.InferFromDocs(context.Background(), "/any/dir")

	// Then
	require.Error(t, err)
	assert.Nil(t, result)
}
