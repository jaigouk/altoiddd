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
// Mock: EgnRenderer
// ---------------------------------------------------------------------------

type mockEgnRenderer struct {
	result string
	err    error
}

func (m *mockEgnRenderer) Render(_ context.Context, _ *discoverydomain.DomainStory) (string, error) {
	return m.result, m.err
}

// ---------------------------------------------------------------------------
// Mock: FileWriter (for Egon handler tests)
// ---------------------------------------------------------------------------

type mockEgnFileWriter struct {
	writtenPath    string
	writtenContent string
	err            error
}

func (m *mockEgnFileWriter) WriteFile(_ context.Context, path string, content string) error {
	if m.err != nil {
		return m.err
	}

	m.writtenPath = path
	m.writtenContent = content

	return nil
}

// ---------------------------------------------------------------------------
// Helper: build a minimal story for Egon handler tests
// ---------------------------------------------------------------------------

func buildEgnHandlerTestStory(t *testing.T) *discoverydomain.DomainStory {
	t.Helper()

	story, err := discoverydomain.NewDomainStory(
		"Egon Handler Test",
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

func TestEgnExportHandler_Export_HappyPath(t *testing.T) {
	t.Parallel()

	story := buildEgnHandlerTestStory(t)
	reader := &mockStoryReader{stories: map[string]*discoverydomain.DomainStory{
		"story.yaml": story,
	}}
	renderer := &mockEgnRenderer{result: `{"domain":{},"dst":[]}`}
	writer := &mockEgnFileWriter{}

	handler := application.NewEgnExportHandler(reader, renderer, writer)
	err := handler.Export(context.TODO(), "story.yaml", "output.egn")

	require.NoError(t, err)
	assert.Equal(t, "output.egn", writer.writtenPath)
	assert.JSONEq(t, `{"domain":{},"dst":[]}`, writer.writtenContent)
}

func TestEgnExportHandler_Export_StoryReadError(t *testing.T) {
	t.Parallel()

	reader := &mockStoryReader{err: fmt.Errorf("disk failure")}
	renderer := &mockEgnRenderer{result: "unused"}
	writer := &mockEgnFileWriter{}

	handler := application.NewEgnExportHandler(reader, renderer, writer)
	err := handler.Export(context.TODO(), "missing.yaml", "output.egn")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading story")
}

func TestEgnExportHandler_Export_RenderError(t *testing.T) {
	t.Parallel()

	story := buildEgnHandlerTestStory(t)
	reader := &mockStoryReader{stories: map[string]*discoverydomain.DomainStory{
		"story.yaml": story,
	}}
	renderer := &mockEgnRenderer{err: fmt.Errorf("render failure")}
	writer := &mockEgnFileWriter{}

	handler := application.NewEgnExportHandler(reader, renderer, writer)
	err := handler.Export(context.TODO(), "story.yaml", "output.egn")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "rendering egn")
}

func TestEgnExportHandler_Export_WriteError(t *testing.T) {
	t.Parallel()

	story := buildEgnHandlerTestStory(t)
	reader := &mockStoryReader{stories: map[string]*discoverydomain.DomainStory{
		"story.yaml": story,
	}}
	renderer := &mockEgnRenderer{result: `{"domain":{},"dst":[]}`}
	writer := &mockEgnFileWriter{err: fmt.Errorf("write failure")}

	handler := application.NewEgnExportHandler(reader, renderer, writer)
	err := handler.Export(context.TODO(), "story.yaml", "output.egn")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "writing egn")
}
