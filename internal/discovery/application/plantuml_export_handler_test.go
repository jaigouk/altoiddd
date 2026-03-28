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
// Mock: PlantUMLRenderer
// ---------------------------------------------------------------------------

type mockPlantUMLRenderer struct {
	result string
	err    error
}

func (m *mockPlantUMLRenderer) Render(_ context.Context, _ *discoverydomain.DomainStory) (string, error) {
	return m.result, m.err
}

// ---------------------------------------------------------------------------
// Mock: FileWriter (for PlantUML handler tests)
// ---------------------------------------------------------------------------

type mockPumlFileWriter struct {
	writtenPath    string
	writtenContent string
	err            error
}

func (m *mockPumlFileWriter) WriteFile(_ context.Context, path string, content string) error {
	if m.err != nil {
		return m.err
	}

	m.writtenPath = path
	m.writtenContent = content

	return nil
}

// ---------------------------------------------------------------------------
// Helper: build a minimal story for PlantUML handler tests
// ---------------------------------------------------------------------------

func buildPumlHandlerTestStory(t *testing.T) *discoverydomain.DomainStory {
	t.Helper()

	story, err := discoverydomain.NewDomainStory(
		"Handler Test Story",
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

func TestPlantUMLExportHandler_Export_HappyPath(t *testing.T) {
	t.Parallel()

	story := buildPumlHandlerTestStory(t)
	reader := &mockStoryReader{stories: map[string]*discoverydomain.DomainStory{
		"story.yaml": story,
	}}
	renderer := &mockPlantUMLRenderer{result: "@startuml\n@enduml"}
	writer := &mockPumlFileWriter{}

	handler := application.NewPlantUMLExportHandler(reader, renderer, writer)
	err := handler.Export(context.TODO(), "story.yaml", "output.puml")

	require.NoError(t, err)
	assert.Equal(t, "output.puml", writer.writtenPath)
	assert.Equal(t, "@startuml\n@enduml", writer.writtenContent)
}

func TestPlantUMLExportHandler_Export_StoryReadError(t *testing.T) {
	t.Parallel()

	reader := &mockStoryReader{err: fmt.Errorf("disk failure")}
	renderer := &mockPlantUMLRenderer{result: "unused"}
	writer := &mockPumlFileWriter{}

	handler := application.NewPlantUMLExportHandler(reader, renderer, writer)
	err := handler.Export(context.TODO(), "missing.yaml", "output.puml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading story")
}

func TestPlantUMLExportHandler_Export_RenderError(t *testing.T) {
	t.Parallel()

	story := buildPumlHandlerTestStory(t)
	reader := &mockStoryReader{stories: map[string]*discoverydomain.DomainStory{
		"story.yaml": story,
	}}
	renderer := &mockPlantUMLRenderer{err: fmt.Errorf("render failure")}
	writer := &mockPumlFileWriter{}

	handler := application.NewPlantUMLExportHandler(reader, renderer, writer)
	err := handler.Export(context.TODO(), "story.yaml", "output.puml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "rendering plantuml")
}

func TestPlantUMLExportHandler_Export_WriteError(t *testing.T) {
	t.Parallel()

	story := buildPumlHandlerTestStory(t)
	reader := &mockStoryReader{stories: map[string]*discoverydomain.DomainStory{
		"story.yaml": story,
	}}
	renderer := &mockPlantUMLRenderer{result: "@startuml\n@enduml"}
	writer := &mockPumlFileWriter{err: fmt.Errorf("write failure")}

	handler := application.NewPlantUMLExportHandler(reader, renderer, writer)
	err := handler.Export(context.TODO(), "story.yaml", "output.puml")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "writing plantuml")
}
