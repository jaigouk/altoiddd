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
// Mock: GlossaryReader (for report handler tests)
// ---------------------------------------------------------------------------

type mockReportGlossaryReader struct {
	entries []vo.UbiquitousLanguageEntry
	err     error
}

func (m *mockReportGlossaryReader) Read(_ context.Context, _ string) ([]vo.UbiquitousLanguageEntry, error) {
	return m.entries, m.err
}

// ---------------------------------------------------------------------------
// Mock: ContextMapReader (for report handler tests)
// ---------------------------------------------------------------------------

type mockReportContextMapReader struct {
	contextMap *discoverydomain.ContextMap
	err        error
}

func (m *mockReportContextMapReader) Read(_ context.Context, _ string) (*discoverydomain.ContextMap, error) {
	return m.contextMap, m.err
}

// ---------------------------------------------------------------------------
// Mock: FileWriter (for report handler tests)
// ---------------------------------------------------------------------------

type mockReportFileWriter struct {
	writtenPath    string
	writtenContent string
	err            error
}

func (m *mockReportFileWriter) WriteFile(_ context.Context, path string, content string) error {
	if m.err != nil {
		return m.err
	}

	m.writtenPath = path
	m.writtenContent = content

	return nil
}

// ---------------------------------------------------------------------------
// Helper: build a minimal story for report handler tests
// ---------------------------------------------------------------------------

func buildReportHandlerTestStory(t *testing.T, title string) *discoverydomain.DomainStory {
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

// buildReportTestGlossaryEntry creates a valid UbiquitousLanguageEntry for tests.
func buildReportTestGlossaryEntry(t *testing.T, term string) vo.UbiquitousLanguageEntry {
	t.Helper()

	entry, err := vo.NewUbiquitousLanguageEntry(term, "A test definition", "General", vo.UserStated, "Story1")
	require.NoError(t, err)

	return entry
}

// buildReportTestContextMap creates a valid ContextMap for tests.
func buildReportTestContextMap(t *testing.T) *discoverydomain.ContextMap {
	t.Helper()

	sketch, err := discoverydomain.NewBoundedContextSketch(
		"Ordering",
		vo.SubdomainCore,
		0.8,
		[]string{"Customer"},
		[]string{"Order"},
		[]string{"Order Flow"},
		nil,
		vo.AIInferred,
	)
	require.NoError(t, err)

	cm, err := discoverydomain.NewContextMap("TestProject", []discoverydomain.BoundedContextSketch{sketch}, nil)
	require.NoError(t, err)

	return cm
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestDiscoveryReportHandler_GenerateReport_HappyPath(t *testing.T) {
	t.Parallel()

	story := buildReportHandlerTestStory(t, "Order Flow")
	storyReader := &mockStoryReader{stories: map[string]*discoverydomain.DomainStory{
		"order.yaml": story,
	}}
	glossaryReader := &mockReportGlossaryReader{
		entries: []vo.UbiquitousLanguageEntry{
			buildReportTestGlossaryEntry(t, "Customer"),
			buildReportTestGlossaryEntry(t, "Order"),
		},
	}
	contextMapReader := &mockReportContextMapReader{contextMap: buildReportTestContextMap(t)}
	writer := &mockReportFileWriter{}

	handler := application.NewDiscoveryReportHandler(storyReader, glossaryReader, contextMapReader, writer)
	err := handler.GenerateReport(context.TODO(), []string{"order.yaml"}, "glossary.yaml", "contextmap.yaml", "/tmp/project")

	require.NoError(t, err)
	assert.Equal(t, "/tmp/project/.alto/discovery-report.md", writer.writtenPath)
	assert.Contains(t, writer.writtenContent, "# Discovery Report")
	assert.Contains(t, writer.writtenContent, "## Stories")
	assert.Contains(t, writer.writtenContent, "## Trust Distribution")
	assert.Contains(t, writer.writtenContent, "Order Flow")
}

func TestDiscoveryReportHandler_GenerateReport_EmptyGlossary(t *testing.T) {
	t.Parallel()

	story := buildReportHandlerTestStory(t, "Order Flow")
	storyReader := &mockStoryReader{stories: map[string]*discoverydomain.DomainStory{
		"order.yaml": story,
	}}
	glossaryReader := &mockReportGlossaryReader{entries: []vo.UbiquitousLanguageEntry{}}
	contextMapReader := &mockReportContextMapReader{contextMap: buildReportTestContextMap(t)}
	writer := &mockReportFileWriter{}

	handler := application.NewDiscoveryReportHandler(storyReader, glossaryReader, contextMapReader, writer)
	err := handler.GenerateReport(context.TODO(), []string{"order.yaml"}, "glossary.yaml", "contextmap.yaml", "/tmp/project")

	require.NoError(t, err)
	assert.Contains(t, writer.writtenContent, "Terms defined: 0")
}

func TestDiscoveryReportHandler_GenerateReport_StoryReadError_WrapsError(t *testing.T) {
	t.Parallel()

	storyReader := &mockStoryReader{err: fmt.Errorf("disk failure")}
	glossaryReader := &mockReportGlossaryReader{entries: []vo.UbiquitousLanguageEntry{}}
	contextMapReader := &mockReportContextMapReader{contextMap: buildReportTestContextMap(t)}
	writer := &mockReportFileWriter{}

	handler := application.NewDiscoveryReportHandler(storyReader, glossaryReader, contextMapReader, writer)
	err := handler.GenerateReport(context.TODO(), []string{"bad.yaml"}, "glossary.yaml", "contextmap.yaml", "/tmp/project")

	require.Error(t, err)
	assert.Contains(t, err.Error(), `reading story "bad.yaml"`)
}

func TestDiscoveryReportHandler_GenerateReport_ContextMapReadError_WrapsError(t *testing.T) {
	t.Parallel()

	storyReader := &mockStoryReader{stories: map[string]*discoverydomain.DomainStory{}}
	glossaryReader := &mockReportGlossaryReader{entries: []vo.UbiquitousLanguageEntry{}}
	contextMapReader := &mockReportContextMapReader{err: fmt.Errorf("corrupt file")}
	writer := &mockReportFileWriter{}

	handler := application.NewDiscoveryReportHandler(storyReader, glossaryReader, contextMapReader, writer)
	err := handler.GenerateReport(context.TODO(), []string{}, "glossary.yaml", "contextmap.yaml", "/tmp/project")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading context map")
}

func TestDiscoveryReportHandler_GenerateReport_WriteError_Propagates(t *testing.T) {
	t.Parallel()

	story := buildReportHandlerTestStory(t, "Order Flow")
	storyReader := &mockStoryReader{stories: map[string]*discoverydomain.DomainStory{
		"order.yaml": story,
	}}
	glossaryReader := &mockReportGlossaryReader{entries: []vo.UbiquitousLanguageEntry{}}
	contextMapReader := &mockReportContextMapReader{contextMap: buildReportTestContextMap(t)}
	writer := &mockReportFileWriter{err: fmt.Errorf("permission denied")}

	handler := application.NewDiscoveryReportHandler(storyReader, glossaryReader, contextMapReader, writer)
	err := handler.GenerateReport(context.TODO(), []string{"order.yaml"}, "glossary.yaml", "contextmap.yaml", "/tmp/project")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "writing report")
}

func TestDiscoveryReportHandler_GenerateReport_ZeroStories_WritesReport(t *testing.T) {
	t.Parallel()

	storyReader := &mockStoryReader{stories: map[string]*discoverydomain.DomainStory{}}
	glossaryReader := &mockReportGlossaryReader{
		entries: []vo.UbiquitousLanguageEntry{
			buildReportTestGlossaryEntry(t, "Customer"),
		},
	}
	contextMapReader := &mockReportContextMapReader{contextMap: buildReportTestContextMap(t)}
	writer := &mockReportFileWriter{}

	handler := application.NewDiscoveryReportHandler(storyReader, glossaryReader, contextMapReader, writer)
	err := handler.GenerateReport(context.TODO(), []string{}, "glossary.yaml", "contextmap.yaml", "/tmp/project")

	require.NoError(t, err)
	assert.NotEmpty(t, writer.writtenContent, "report should be written even with zero stories")
}
