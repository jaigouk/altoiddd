package application_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/fitness/application"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// Fake file writer
// ---------------------------------------------------------------------------

type fakeFileWriterF struct {
	written map[string]string
}

func newFakeFileWriterF() *fakeFileWriterF {
	return &fakeFileWriterF{written: make(map[string]string)}
}

func (f *fakeFileWriterF) WriteFile(_ context.Context, path, content string) error {
	f.written[path] = content
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeModelWithContexts(names []string) *ddd.DomainModel {
	model := ddd.NewDomainModel("test-model")

	// Add a story mentioning all contexts
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

func TestFitnessGenerationHandler_BuildPreview(t *testing.T) {
	t.Parallel()

	t.Run("returns preview", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer)
		model := makeModelWithContexts([]string{"Orders"})

		preview, err := handler.BuildPreview(model, "myapp", nil)

		require.NoError(t, err)
		require.NotNil(t, preview)
		assert.NotEmpty(t, preview.TomlContent)
		assert.NotEmpty(t, preview.TestContent)
		assert.NotEmpty(t, preview.Summary)
	})

	t.Run("does not write", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer)
		model := makeModelWithContexts([]string{"Orders"})

		handler.BuildPreview(model, "myapp", nil)

		assert.Equal(t, 0, len(writer.written))
	})

	t.Run("contains all bc names", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer)
		model := makeModelWithContexts([]string{"Orders", "Notifications"})

		preview, err := handler.BuildPreview(model, "myapp", nil)

		require.NoError(t, err)
		require.NotNil(t, preview)
		assert.Contains(t, preview.Summary, "Orders")
		assert.Contains(t, preview.Summary, "Notifications")
	})

	t.Run("no contexts raises", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer)
		model := ddd.NewDomainModel("empty")

		_, err := handler.BuildPreview(model, "myapp", nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "No bounded contexts")
	})

	t.Run("returns nil when fitness not available", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer)
		model := makeModelWithContexts([]string{"Orders"})
		profile := vo.GenericProfile{} // FitnessAvailable() = false

		preview, err := handler.BuildPreview(model, "myapp", profile)

		require.NoError(t, err)
		assert.Nil(t, preview)
	})
}

// ---------------------------------------------------------------------------
// Tests — Write Files
// ---------------------------------------------------------------------------

func TestFitnessGenerationHandler_WriteFiles(t *testing.T) {
	t.Parallel()

	t.Run("creates toml and test files", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer)
		model := makeModelWithContexts([]string{"Orders"})

		preview, _ := handler.BuildPreview(model, "myapp", nil)
		err := handler.WriteFiles(context.Background(), preview, "/project")

		require.NoError(t, err)
		assert.True(t, len(writer.written) >= 2)

		hasToml := false
		hasTest := false
		for p := range writer.written {
			if strings.Contains(p, "importlinter") {
				hasToml = true
			}
			if strings.Contains(p, "test_") {
				hasTest = true
			}
		}
		assert.True(t, hasToml)
		assert.True(t, hasTest)
	})

	t.Run("content matches preview", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer)
		model := makeModelWithContexts([]string{"Orders"})

		preview, _ := handler.BuildPreview(model, "myapp", nil)
		handler.WriteFiles(context.Background(), preview, "/project")

		for p, content := range writer.written {
			if strings.Contains(p, "importlinter") {
				assert.Equal(t, preview.TomlContent, content)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// Tests — Approve and Write
// ---------------------------------------------------------------------------

func TestFitnessGenerationHandler_ApproveAndWrite(t *testing.T) {
	t.Parallel()

	t.Run("calls approve and writes files", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer)
		model := makeModelWithContexts([]string{"Orders"})

		preview, _ := handler.BuildPreview(model, "myapp", nil)
		err := handler.ApproveAndWrite(context.Background(), preview, "/project")

		require.NoError(t, err)
		assert.True(t, len(preview.Suite.Events()) > 0)
		assert.True(t, len(writer.written) >= 2)
	})

	t.Run("approve twice raises", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer)
		model := makeModelWithContexts([]string{"Orders"})

		preview, _ := handler.BuildPreview(model, "myapp", nil)
		handler.ApproveAndWrite(context.Background(), preview, "/project")
		err := handler.ApproveAndWrite(context.Background(), preview, "/project")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already approved")
	})

	t.Run("no generate convenience method", func(t *testing.T) {
		t.Parallel()
		// Verify FitnessGenerationHandler doesn't have a Generate method
		// This is enforced by not defining one; the test is implicit.
	})
}
