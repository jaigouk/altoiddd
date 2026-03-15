package application_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/fitness/application"
	fitnessdomain "github.com/alty-cli/alty/internal/fitness/domain"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// Fake file writer
// ---------------------------------------------------------------------------

type fakePublisherF struct {
	published []any
}

func (f *fakePublisherF) Publish(_ context.Context, event any) error {
	f.published = append(f.published, event)
	return nil
}

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
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithContexts([]string{"Orders"})

		preview, err := handler.BuildPreview(model, "myapp", nil, nil)

		require.NoError(t, err)
		require.NotNil(t, preview)
		assert.NotEmpty(t, preview.TomlContent)
		assert.NotEmpty(t, preview.TestContent)
		assert.NotEmpty(t, preview.Summary)
	})

	t.Run("does not write", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithContexts([]string{"Orders"})

		handler.BuildPreview(model, "myapp", nil, nil)

		assert.Empty(t, writer.written)
	})

	t.Run("contains all bc names", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithContexts([]string{"Orders", "Notifications"})

		preview, err := handler.BuildPreview(model, "myapp", nil, nil)

		require.NoError(t, err)
		require.NotNil(t, preview)
		assert.Contains(t, preview.Summary, "Orders")
		assert.Contains(t, preview.Summary, "Notifications")
	})

	t.Run("truly empty model raises", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := ddd.NewDomainModel("empty")

		_, err := handler.BuildPreview(model, "myapp", nil, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "model is empty")
	})

	t.Run("no contexts returns preview with warnings", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		// Model with a story but no contexts — not truly empty
		model := ddd.NewDomainModel("partial")
		story := vo.NewDomainStory("Partial flow", []string{"User"}, "User starts", []string{"User acts"}, nil)
		model.AddDomainStory(story)

		preview, err := handler.BuildPreview(model, "myapp", nil, nil)

		require.NoError(t, err)
		require.NotNil(t, preview)
		assert.NotEmpty(t, preview.Warnings())
		assert.Contains(t, preview.Warnings()[0], "no bounded contexts")
	})

	t.Run("returns nil when fitness not available", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithContexts([]string{"Orders"})
		profile := vo.GenericProfile{} // FitnessAvailable() = false

		preview, err := handler.BuildPreview(model, "myapp", profile, nil)

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
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithContexts([]string{"Orders"})

		preview, _ := handler.BuildPreview(model, "myapp", nil, nil)
		err := handler.WriteFiles(context.Background(), preview, "/project")

		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(writer.written), 2)

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
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithContexts([]string{"Orders"})

		preview, _ := handler.BuildPreview(model, "myapp", nil, nil)
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
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithContexts([]string{"Orders"})

		preview, _ := handler.BuildPreview(model, "myapp", nil, nil)
		err := handler.ApproveAndWrite(context.Background(), preview, "/project")

		require.NoError(t, err)
		assert.NotEmpty(t, preview.Suite.Events())
		assert.GreaterOrEqual(t, len(writer.written), 2)
	})

	t.Run("approve twice raises", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithContexts([]string{"Orders"})

		preview, _ := handler.BuildPreview(model, "myapp", nil, nil)
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

func TestFitnessGenerationHandler_ApproveAndWrite_PublishesEvent(t *testing.T) {
	t.Parallel()

	pub := &fakePublisherF{}
	writer := newFakeFileWriterF()
	handler := application.NewFitnessGenerationHandler(writer, pub)
	model := makeModelWithContexts([]string{"Orders"})

	preview, err := handler.BuildPreview(model, "myapp", nil, nil)
	require.NoError(t, err)

	err = handler.ApproveAndWrite(context.Background(), preview, "/project")
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(pub.published), 1)
	_, ok := pub.published[0].(fitnessdomain.FitnessTestsGenerated)
	assert.True(t, ok, "expected FitnessTestsGenerated, got %T", pub.published[0])
}

// ---------------------------------------------------------------------------
// Tests — Go Stack Support (arch-go)
// ---------------------------------------------------------------------------

func TestFitnessGenerationHandler_BuildPreview_GoStack(t *testing.T) {
	t.Parallel()

	t.Run("returns YAMLContent for GoModProfile", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithContexts([]string{"Orders"})
		profile := vo.GoModProfile{}

		preview, err := handler.BuildPreview(model, "github.com/org/myapp", profile, nil)

		require.NoError(t, err)
		require.NotNil(t, preview)
		assert.NotEmpty(t, preview.YAMLContent)
		assert.Contains(t, preview.YAMLContent, "version: 1")
		assert.Contains(t, preview.YAMLContent, "dependenciesRules:")
		assert.Equal(t, "go-mod", preview.StackID)
	})

	t.Run("Go preview has empty TomlContent and TestContent", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithContexts([]string{"Orders"})
		profile := vo.GoModProfile{}

		preview, err := handler.BuildPreview(model, "github.com/org/myapp", profile, nil)

		require.NoError(t, err)
		require.NotNil(t, preview)
		assert.Empty(t, preview.TomlContent)
		assert.Empty(t, preview.TestContent)
	})

	t.Run("Go preview contains context names in YAML", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithContexts([]string{"Orders", "Shipping"})
		profile := vo.GoModProfile{}

		preview, err := handler.BuildPreview(model, "github.com/org/myapp", profile, nil)

		require.NoError(t, err)
		assert.Contains(t, preview.YAMLContent, "orders")
		assert.Contains(t, preview.YAMLContent, "shipping")
	})

	t.Run("Go preview uses snake_case module paths", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithContexts([]string{"OrderManagement"})
		profile := vo.GoModProfile{}

		preview, err := handler.BuildPreview(model, "github.com/org/myapp", profile, nil)

		require.NoError(t, err)
		// Module paths use snake_case in full package paths
		assert.Contains(t, preview.YAMLContent, "github.com/org/myapp/internal/order_management/domain")
		assert.Contains(t, preview.YAMLContent, "github.com/org/myapp/internal/order_management/application")
		// Comments can still use the readable name - that's fine
	})

	t.Run("Go preview has threshold 100 for greenfield", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithContexts([]string{"Orders"})
		profile := vo.GoModProfile{}

		preview, err := handler.BuildPreview(model, "github.com/org/myapp", profile, nil)

		require.NoError(t, err)
		assert.Contains(t, preview.YAMLContent, "compliance: 100")
		assert.Contains(t, preview.YAMLContent, "coverage: 100")
	})

	t.Run("Go preview has threshold 80 for brownfield", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithContexts([]string{"Orders"})
		profile := vo.GoModProfile{}
		opts := &application.BuildPreviewOptions{Threshold: 80}

		preview, err := handler.BuildPreview(model, "github.com/org/myapp", profile, opts)

		require.NoError(t, err)
		assert.Contains(t, preview.YAMLContent, "compliance: 80")
		assert.Contains(t, preview.YAMLContent, "coverage: 80")
	})
}

func TestFitnessGenerationHandler_BuildPreview_PythonStack(t *testing.T) {
	t.Parallel()

	t.Run("returns TomlContent for PythonUvProfile", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithContexts([]string{"Orders"})
		profile := vo.PythonUvProfile{}

		preview, err := handler.BuildPreview(model, "myapp", profile, nil)

		require.NoError(t, err)
		require.NotNil(t, preview)
		assert.NotEmpty(t, preview.TomlContent)
		assert.NotEmpty(t, preview.TestContent)
		assert.Equal(t, "python-uv", preview.StackID)
	})

	t.Run("Python preview has empty YAMLContent", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithContexts([]string{"Orders"})
		profile := vo.PythonUvProfile{}

		preview, err := handler.BuildPreview(model, "myapp", profile, nil)

		require.NoError(t, err)
		require.NotNil(t, preview)
		assert.Empty(t, preview.YAMLContent)
	})
}

func TestFitnessGenerationHandler_WriteFiles_GoStack(t *testing.T) {
	t.Parallel()

	t.Run("writes arch-go.yml for Go stack", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithContexts([]string{"Orders"})
		profile := vo.GoModProfile{}

		preview, _ := handler.BuildPreview(model, "github.com/org/myapp", profile, nil)
		err := handler.WriteFiles(context.Background(), preview, "/project")

		require.NoError(t, err)
		assert.Len(t, writer.written, 1)

		hasArchGo := false
		for p := range writer.written {
			if strings.Contains(p, "arch-go.yml") {
				hasArchGo = true
			}
		}
		assert.True(t, hasArchGo, "expected arch-go.yml to be written")
	})

	t.Run("arch-go.yml content matches preview", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithContexts([]string{"Orders"})
		profile := vo.GoModProfile{}

		preview, _ := handler.BuildPreview(model, "github.com/org/myapp", profile, nil)
		handler.WriteFiles(context.Background(), preview, "/project")

		for p, content := range writer.written {
			if strings.Contains(p, "arch-go.yml") {
				assert.Equal(t, preview.YAMLContent, content)
			}
		}
	})

	t.Run("Go stack does not write Python files", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithContexts([]string{"Orders"})
		profile := vo.GoModProfile{}

		preview, _ := handler.BuildPreview(model, "github.com/org/myapp", profile, nil)
		handler.WriteFiles(context.Background(), preview, "/project")

		for p := range writer.written {
			assert.NotContains(t, p, "importlinter")
			assert.NotContains(t, p, "test_fitness.py")
		}
	})
}

func TestFitnessGenerationHandler_WriteFiles_PythonStack(t *testing.T) {
	t.Parallel()

	t.Run("writes importlinter and test file for Python stack", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithContexts([]string{"Orders"})
		profile := vo.PythonUvProfile{}

		preview, _ := handler.BuildPreview(model, "myapp", profile, nil)
		err := handler.WriteFiles(context.Background(), preview, "/project")

		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(writer.written), 2)

		hasToml := false
		hasTest := false
		for p := range writer.written {
			if strings.Contains(p, "importlinter") {
				hasToml = true
			}
			if strings.Contains(p, "test_fitness.py") {
				hasTest = true
			}
		}
		assert.True(t, hasToml)
		assert.True(t, hasTest)
	})

	t.Run("Python stack does not write arch-go.yml", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithContexts([]string{"Orders"})
		profile := vo.PythonUvProfile{}

		preview, _ := handler.BuildPreview(model, "myapp", profile, nil)
		handler.WriteFiles(context.Background(), preview, "/project")

		for p := range writer.written {
			assert.NotContains(t, p, "arch-go.yml")
		}
	})
}

// ---------------------------------------------------------------------------
// Tests — Model to BoundedContextMap conversion
// ---------------------------------------------------------------------------

func makeModelWithRelationships() *ddd.DomainModel {
	model := ddd.NewDomainModel("test-model")

	// Add story
	story := vo.NewDomainStory(
		"Test flow",
		[]string{"User"},
		"User starts",
		[]string{"User manages Bootstrap", "User manages Discovery"},
		nil,
	)
	model.AddDomainStory(story)

	// Add contexts
	model.AddTerm("Bootstrap", "Bootstrap domain", "Bootstrap", nil)
	model.AddTerm("Discovery", "Discovery domain", "Discovery", nil)

	bc1 := vo.NewDomainBoundedContext("Bootstrap", "Manages bootstrap", nil, nil, "")
	bc2 := vo.NewDomainBoundedContext("Discovery", "Manages discovery", nil, nil, "")
	model.AddBoundedContext(bc1)
	model.AddBoundedContext(bc2)

	supporting := vo.SubdomainSupporting
	core := vo.SubdomainCore
	model.ClassifySubdomain("Bootstrap", supporting, "test")
	model.ClassifySubdomain("Discovery", core, "test")

	// Add relationship: Bootstrap upstream of Discovery
	rel := vo.NewContextRelationship("Bootstrap", "Discovery", "Domain Events")
	model.AddContextRelationship(rel)

	// Add aggregates
	agg1 := vo.NewAggregateDesign("BootstrapRoot", "Bootstrap", "BootstrapRoot", nil, []string{"must be valid"}, nil, nil)
	agg2 := vo.NewAggregateDesign("DiscoveryRoot", "Discovery", "DiscoveryRoot", nil, []string{"must be valid"}, nil, nil)
	model.DesignAggregate(agg1)
	model.DesignAggregate(agg2)

	model.Finalize()
	return model
}

func TestFitnessGenerationHandler_BuildPreview_GoStack_WithRelationships(t *testing.T) {
	t.Parallel()

	t.Run("relationships are reflected in YAML", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithRelationships()
		profile := vo.GoModProfile{}

		preview, err := handler.BuildPreview(model, "github.com/org/myapp", profile, nil)

		require.NoError(t, err)
		require.NotNil(t, preview)
		// YAML should contain both contexts
		assert.Contains(t, preview.YAMLContent, "bootstrap")
		assert.Contains(t, preview.YAMLContent, "discovery")
	})
}

func TestFitnessGenerationHandler_BuildPreview_GoStack_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("single context produces valid YAML", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithContexts([]string{"Standalone"})
		profile := vo.GoModProfile{}

		preview, err := handler.BuildPreview(model, "github.com/org/myapp", profile, nil)

		require.NoError(t, err)
		assert.Contains(t, preview.YAMLContent, "standalone")
		assert.Contains(t, preview.YAMLContent, "shouldOnlyDependsOn")
	})

	t.Run("many contexts produce cross-isolation rules", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterF()
		handler := application.NewFitnessGenerationHandler(writer, &fakePublisherF{})
		model := makeModelWithContexts([]string{"A", "B", "C"})
		profile := vo.GoModProfile{}

		preview, err := handler.BuildPreview(model, "github.com/org/myapp", profile, nil)

		require.NoError(t, err)
		// Should have isolation rules between unrelated contexts
		assert.Contains(t, preview.YAMLContent, "must not depend on")
	})
}
