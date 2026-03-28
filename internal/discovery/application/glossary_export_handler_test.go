package application_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/application"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// Mock: StoryReader
// ---------------------------------------------------------------------------

type mockStoryReader struct {
	stories map[string]*discoverydomain.DomainStory
	err     error
}

func (m *mockStoryReader) Read(_ context.Context, path string) (*discoverydomain.DomainStory, error) {
	if m.err != nil {
		return nil, m.err
	}

	story, ok := m.stories[path]
	if !ok {
		return nil, fmt.Errorf("story not found: %s", path)
	}

	return story, nil
}

// ---------------------------------------------------------------------------
// Mock: GlossaryWriter
// ---------------------------------------------------------------------------

type mockGlossaryWriter struct {
	writtenEntries []vo.UbiquitousLanguageEntry
	writtenPath    string
	err            error
}

func (m *mockGlossaryWriter) Write(_ context.Context, path string, entries []vo.UbiquitousLanguageEntry) error {
	if m.err != nil {
		return m.err
	}

	m.writtenPath = path
	m.writtenEntries = entries

	return nil
}

// ---------------------------------------------------------------------------
// Helper: build a minimal valid story for handler tests
// ---------------------------------------------------------------------------

func buildHandlerTestStory(t *testing.T, title string) *discoverydomain.DomainStory {
	t.Helper()

	story, err := discoverydomain.NewDomainStory(
		title,
		discoverydomain.StoryTypeCoarseGrained,
		discoverydomain.TimeTypeAsIs,
		discoverydomain.PurityTypePure,
		"test trigger",
	)
	require.NoError(t, err)

	actor, err := discoverydomain.NewStoryActor("Customer", discoverydomain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddActor(actor))

	wo, err := discoverydomain.NewWorkObject("Order", discoverydomain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddWorkObject(wo))

	sentence, err := discoverydomain.NewStorySentence(1, "Customer", "creates", "Order", vo.UserStated, "")
	require.NoError(t, err)
	require.NoError(t, story.AddSentence(sentence))

	return story
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestGlossaryExportHandler_Export_HappyPath(t *testing.T) {
	t.Parallel()

	story := buildHandlerTestStory(t, "Order Flow")
	reader := &mockStoryReader{stories: map[string]*discoverydomain.DomainStory{
		"story1.yaml": story,
	}}
	writer := &mockGlossaryWriter{}

	handler := application.NewGlossaryExportHandler(reader, writer)
	err := handler.Export(context.TODO(), []string{"story1.yaml"}, nil, "output.yaml")

	require.NoError(t, err)
	assert.Equal(t, "output.yaml", writer.writtenPath)
	assert.NotEmpty(t, writer.writtenEntries)
}

func TestGlossaryExportHandler_Export_StoryReadError(t *testing.T) {
	t.Parallel()

	reader := &mockStoryReader{err: fmt.Errorf("disk failure")}
	writer := &mockGlossaryWriter{}

	handler := application.NewGlossaryExportHandler(reader, writer)
	err := handler.Export(context.TODO(), []string{"missing.yaml"}, nil, "output.yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "disk failure")
}

func TestGlossaryExportHandler_Export_GlossaryWriteError(t *testing.T) {
	t.Parallel()

	story := buildHandlerTestStory(t, "Order Flow")
	reader := &mockStoryReader{stories: map[string]*discoverydomain.DomainStory{
		"story1.yaml": story,
	}}
	writer := &mockGlossaryWriter{err: fmt.Errorf("write failure")}

	handler := application.NewGlossaryExportHandler(reader, writer)
	err := handler.Export(context.TODO(), []string{"story1.yaml"}, nil, "output.yaml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "write failure")
}

func TestGlossaryExportHandler_Export_EmptyStoryPaths(t *testing.T) {
	t.Parallel()

	reader := &mockStoryReader{stories: map[string]*discoverydomain.DomainStory{}}
	writer := &mockGlossaryWriter{}

	handler := application.NewGlossaryExportHandler(reader, writer)
	err := handler.Export(context.TODO(), []string{}, nil, "output.yaml")

	require.NoError(t, err)
	assert.Empty(t, writer.writtenEntries)
}

func TestGlossaryExportHandler_Export_NilContextMap(t *testing.T) {
	t.Parallel()

	story := buildHandlerTestStory(t, "Order Flow")
	reader := &mockStoryReader{stories: map[string]*discoverydomain.DomainStory{
		"story1.yaml": story,
	}}
	writer := &mockGlossaryWriter{}

	handler := application.NewGlossaryExportHandler(reader, writer)
	err := handler.Export(context.TODO(), []string{"story1.yaml"}, nil, "output.yaml")

	// nil contextMap is valid — entries get "General" context.
	require.NoError(t, err)
	assert.NotEmpty(t, writer.writtenEntries)
}
