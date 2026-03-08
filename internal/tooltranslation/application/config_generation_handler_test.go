package application_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
	"github.com/alty-cli/alty/internal/tooltranslation/application"
	ttdomain "github.com/alty-cli/alty/internal/tooltranslation/domain"
)

// ---------------------------------------------------------------------------
// Fakes
// ---------------------------------------------------------------------------

type fakePublisherC struct {
	published []any
}

func (f *fakePublisherC) Publish(_ context.Context, event any) error {
	f.published = append(f.published, event)
	return nil
}

type fakeFileWriterC struct {
	written map[string]string
}

func newFakeFileWriterC() *fakeFileWriterC {
	return &fakeFileWriterC{written: make(map[string]string)}
}

func (f *fakeFileWriterC) WriteFile(_ context.Context, path, content string) error {
	f.written[path] = content
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeConfigModel(names []string) *ddd.DomainModel {
	model := ddd.NewDomainModel("config-test")
	steps := make([]string, len(names))
	for i, name := range names {
		steps[i] = "User manages " + name
	}
	story := vo.NewDomainStory("Test flow", []string{"User"}, "User starts", steps, nil)
	model.AddDomainStory(story)

	for _, name := range names {
		model.AddTerm(name, name+" domain", name, nil)
		bc := vo.NewDomainBoundedContext(name, "Manages "+name, nil, nil, "")
		model.AddBoundedContext(bc)
		core := vo.SubdomainCore
		model.ClassifySubdomain(name, core, "test")
		agg := vo.NewAggregateDesign(name+"Root", name, name+"Root", nil, []string{"must be valid"}, nil, nil)
		model.DesignAggregate(agg)
	}
	model.Finalize()
	return model
}

// ---------------------------------------------------------------------------
// Tests — Build Preview
// ---------------------------------------------------------------------------

func TestConfigGenerationHandler_BuildPreview(t *testing.T) {
	t.Parallel()

	t.Run("returns config preview", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterC()
		handler := application.NewConfigGenerationHandler(writer, &fakePublisherC{})
		model := makeConfigModel([]string{"Orders"})

		preview, err := handler.BuildPreview(model, []ttdomain.SupportedTool{ttdomain.ToolClaudeCode}, nil)

		require.NoError(t, err)
		require.NotNil(t, preview)
		assert.Len(t, preview.Configs, 1)
		assert.NotEmpty(t, preview.Summary)
	})

	t.Run("does not write", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterC()
		handler := application.NewConfigGenerationHandler(writer, &fakePublisherC{})
		model := makeConfigModel([]string{"Orders"})

		handler.BuildPreview(model, []ttdomain.SupportedTool{ttdomain.ToolClaudeCode}, nil)

		assert.Empty(t, writer.written)
	})

	t.Run("preview for multiple tools", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterC()
		handler := application.NewConfigGenerationHandler(writer, &fakePublisherC{})
		model := makeConfigModel([]string{"Orders"})

		preview, err := handler.BuildPreview(model,
			[]ttdomain.SupportedTool{ttdomain.ToolClaudeCode, ttdomain.ToolCursor}, nil)

		require.NoError(t, err)
		assert.Len(t, preview.Configs, 2)
		assert.Contains(t, preview.Summary, "claude-code")
		assert.Contains(t, preview.Summary, "cursor")
	})

	t.Run("empty tools raises", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterC()
		handler := application.NewConfigGenerationHandler(writer, &fakePublisherC{})
		model := makeConfigModel([]string{"Orders"})

		_, err := handler.BuildPreview(model, []ttdomain.SupportedTool{}, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no tools")
	})
}

// ---------------------------------------------------------------------------
// Tests — Approve and Write
// ---------------------------------------------------------------------------

func TestConfigGenerationHandler_ApproveAndWrite(t *testing.T) {
	t.Parallel()

	t.Run("writes files", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterC()
		handler := application.NewConfigGenerationHandler(writer, &fakePublisherC{})
		model := makeConfigModel([]string{"Orders"})
		preview, _ := handler.BuildPreview(model,
			[]ttdomain.SupportedTool{ttdomain.ToolClaudeCode}, nil)

		err := handler.ApproveAndWrite(context.Background(), preview, "/project")

		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(writer.written), 1)
		hasClaude := false
		for p := range writer.written {
			if strings.Contains(p, "CLAUDE.md") {
				hasClaude = true
			}
		}
		assert.True(t, hasClaude)
	})

	t.Run("emits events", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterC()
		handler := application.NewConfigGenerationHandler(writer, &fakePublisherC{})
		model := makeConfigModel([]string{"Orders"})
		preview, _ := handler.BuildPreview(model,
			[]ttdomain.SupportedTool{ttdomain.ToolClaudeCode}, nil)

		handler.ApproveAndWrite(context.Background(), preview, "/project")

		for _, config := range preview.Configs {
			assert.Len(t, config.Events(), 1)
		}
	})

	t.Run("approve twice raises", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterC()
		handler := application.NewConfigGenerationHandler(writer, &fakePublisherC{})
		model := makeConfigModel([]string{"Orders"})
		preview, _ := handler.BuildPreview(model,
			[]ttdomain.SupportedTool{ttdomain.ToolClaudeCode}, nil)

		handler.ApproveAndWrite(context.Background(), preview, "/project")
		err := handler.ApproveAndWrite(context.Background(), preview, "/project")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "already approved")
	})

	t.Run("multiple tools writes all files", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterC()
		handler := application.NewConfigGenerationHandler(writer, &fakePublisherC{})
		model := makeConfigModel([]string{"Orders"})
		preview, _ := handler.BuildPreview(model,
			[]ttdomain.SupportedTool{ttdomain.ToolClaudeCode, ttdomain.ToolCursor}, nil)

		handler.ApproveAndWrite(context.Background(), preview, "/project")

		paths := make([]string, 0, len(writer.written))
		for p := range writer.written {
			paths = append(paths, p)
		}
		assert.GreaterOrEqual(t, len(paths), 3) // Claude: 2 files, Cursor: 2 files
		hasClaude := false
		hasCursor := false
		for _, p := range paths {
			if strings.Contains(p, "CLAUDE.md") {
				hasClaude = true
			}
			if strings.Contains(p, ".cursor/rules") {
				hasCursor = true
			}
		}
		assert.True(t, hasClaude)
		assert.True(t, hasCursor)
	})

	t.Run("all four tools", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterC()
		handler := application.NewConfigGenerationHandler(writer, &fakePublisherC{})
		model := makeConfigModel([]string{"Orders"})
		allTools := []ttdomain.SupportedTool{
			ttdomain.ToolClaudeCode, ttdomain.ToolCursor,
			ttdomain.ToolRooCode, ttdomain.ToolOpenCode,
		}
		preview, err := handler.BuildPreview(model, allTools, nil)
		require.NoError(t, err)
		assert.Len(t, preview.Configs, 4)

		handler.ApproveAndWrite(context.Background(), preview, "/project")
		assert.GreaterOrEqual(t, len(writer.written), 7)
	})

	t.Run("written content not empty", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterC()
		handler := application.NewConfigGenerationHandler(writer, &fakePublisherC{})
		model := makeConfigModel([]string{"Orders"})
		preview, _ := handler.BuildPreview(model,
			[]ttdomain.SupportedTool{ttdomain.ToolClaudeCode}, nil)

		handler.ApproveAndWrite(context.Background(), preview, "/project")

		for _, content := range writer.written {
			assert.NotEmpty(t, content)
		}
	})
}

func TestConfigGenerationHandler_ApproveAndWrite_PublishesEvent(t *testing.T) {
	t.Parallel()

	pub := &fakePublisherC{}
	writer := newFakeFileWriterC()
	handler := application.NewConfigGenerationHandler(writer, pub)
	model := makeConfigModel([]string{"Orders"})
	preview, err := handler.BuildPreview(model,
		[]ttdomain.SupportedTool{ttdomain.ToolClaudeCode}, nil)
	require.NoError(t, err)

	err = handler.ApproveAndWrite(context.Background(), preview, "/project")
	require.NoError(t, err)

	require.Len(t, pub.published, 1)
	_, ok := pub.published[0].(ttdomain.ConfigsGeneratedEvent)
	assert.True(t, ok, "expected ConfigsGeneratedEvent, got %T", pub.published[0])
}
