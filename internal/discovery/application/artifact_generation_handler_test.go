package application_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/discovery/application"
	discoverydomain "github.com/alty-cli/alty/internal/discovery/domain"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
)

// ---------------------------------------------------------------------------
// Fakes
// ---------------------------------------------------------------------------

type fakeArtifactRenderer struct {
	prdContent  string
	dddContent  string
	archContent string
	callCount   map[string]int
	calledModel *ddd.DomainModel
}

func newFakeRenderer(prd, ddd, arch string) *fakeArtifactRenderer {
	return &fakeArtifactRenderer{
		prdContent:  prd,
		dddContent:  ddd,
		archContent: arch,
		callCount:   make(map[string]int),
	}
}

func (f *fakeArtifactRenderer) RenderPRD(_ context.Context, model *ddd.DomainModel) (string, error) {
	f.callCount["RenderPRD"]++
	f.calledModel = model
	return f.prdContent, nil
}

func (f *fakeArtifactRenderer) RenderDDD(_ context.Context, model *ddd.DomainModel) (string, error) {
	f.callCount["RenderDDD"]++
	f.calledModel = model
	return f.dddContent, nil
}

func (f *fakeArtifactRenderer) RenderArchitecture(_ context.Context, model *ddd.DomainModel) (string, error) {
	f.callCount["RenderArchitecture"]++
	f.calledModel = model
	return f.archContent, nil
}

type fakeFileWriterA struct {
	written map[string]string
}

func newFakeFileWriterA() *fakeFileWriterA {
	return &fakeFileWriterA{written: make(map[string]string)}
}

func (f *fakeFileWriterA) WriteFile(_ context.Context, path, content string) error {
	f.written[path] = content
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeStandardEvent() discoverydomain.DiscoveryCompletedEvent {
	answers := []discoverydomain.Answer{
		discoverydomain.NewAnswer("Q1", "Customer, Admin"),
		discoverydomain.NewAnswer("Q2", "Order, Product"),
		discoverydomain.NewAnswer("Q3", "Customer reviews order, System processes payment"),
		discoverydomain.NewAnswer("Q4", "Payment must not be negative, Order must have items"),
		discoverydomain.NewAnswer("Q5", "Admin manages Product catalog"),
		discoverydomain.NewAnswer("Q6", "OrderPlaced, PaymentProcessed"),
		discoverydomain.NewAnswer("Q7", "When OrderPlaced, send confirmation email"),
		discoverydomain.NewAnswer("Q8", "Order history, Sales dashboard"),
		discoverydomain.NewAnswer("Q9", "Sales, Inventory"),
		discoverydomain.NewAnswer("Q10", "Sales is core competitive advantage, Inventory is supporting necessary"),
	}
	playbacks := []discoverydomain.Playback{
		discoverydomain.NewPlayback("Playback 1", true, ""),
	}
	return discoverydomain.NewDiscoveryCompletedEvent(
		"session-1",
		discoverydomain.PersonaDeveloper,
		discoverydomain.RegisterTechnical,
		answers,
		playbacks,
		nil,
	)
}

func makeEventWithAnswers(answers []discoverydomain.Answer) discoverydomain.DiscoveryCompletedEvent {
	return discoverydomain.NewDiscoveryCompletedEvent(
		"session-1",
		discoverydomain.PersonaDeveloper,
		discoverydomain.RegisterTechnical,
		answers,
		[]discoverydomain.Playback{discoverydomain.NewPlayback("Playback 1", true, "")},
		nil,
	)
}

// ---------------------------------------------------------------------------
// Tests — Build Preview
// ---------------------------------------------------------------------------

func TestArtifactGenerationHandler_BuildPreview(t *testing.T) {
	t.Parallel()

	t.Run("returns preview without writing", func(t *testing.T) {
		t.Parallel()
		renderer := newFakeRenderer("# PRD", "# DDD", "# ARCH")
		writer := newFakeFileWriterA()
		handler := application.NewArtifactGenerationHandler(renderer, writer)
		event := makeStandardEvent()

		preview, err := handler.BuildPreview(context.Background(), event)

		require.NoError(t, err)
		assert.Equal(t, "# PRD", preview.PRDContent)
		assert.Equal(t, "# DDD", preview.DDDContent)
		assert.Equal(t, "# ARCH", preview.ArchitectureContent)
		assert.Empty(t, writer.written)
	})

	t.Run("model is finalized", func(t *testing.T) {
		t.Parallel()
		renderer := newFakeRenderer("", "", "")
		writer := newFakeFileWriterA()
		handler := application.NewArtifactGenerationHandler(renderer, writer)
		event := makeStandardEvent()

		preview, err := handler.BuildPreview(context.Background(), event)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(preview.Model.Events()), 1)
	})

	t.Run("empty answers raises", func(t *testing.T) {
		t.Parallel()
		renderer := newFakeRenderer("", "", "")
		writer := newFakeFileWriterA()
		handler := application.NewArtifactGenerationHandler(renderer, writer)
		event := makeEventWithAnswers(nil)

		_, err := handler.BuildPreview(context.Background(), event)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no substantive answers")
	})

	t.Run("renderer called with model", func(t *testing.T) {
		t.Parallel()
		renderer := newFakeRenderer("", "", "")
		writer := newFakeFileWriterA()
		handler := application.NewArtifactGenerationHandler(renderer, writer)
		event := makeStandardEvent()

		handler.BuildPreview(context.Background(), event)

		assert.Equal(t, 1, renderer.callCount["RenderPRD"])
		assert.Equal(t, 1, renderer.callCount["RenderDDD"])
		assert.Equal(t, 1, renderer.callCount["RenderArchitecture"])
	})

	t.Run("generates two bounded contexts", func(t *testing.T) {
		t.Parallel()
		renderer := newFakeRenderer("", "", "")
		writer := newFakeFileWriterA()
		handler := application.NewArtifactGenerationHandler(renderer, writer)
		event := makeStandardEvent()

		preview, err := handler.BuildPreview(context.Background(), event)

		require.NoError(t, err)
		assert.Len(t, preview.Model.BoundedContexts(), 2)
	})

	t.Run("generates domain stories", func(t *testing.T) {
		t.Parallel()
		renderer := newFakeRenderer("", "", "")
		writer := newFakeFileWriterA()
		handler := application.NewArtifactGenerationHandler(renderer, writer)
		event := makeStandardEvent()

		preview, err := handler.BuildPreview(context.Background(), event)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(preview.Model.DomainStories()), 1)
	})
}

// ---------------------------------------------------------------------------
// Tests — Write Artifacts
// ---------------------------------------------------------------------------

func TestArtifactGenerationHandler_WriteArtifacts(t *testing.T) {
	t.Parallel()

	t.Run("writes three files", func(t *testing.T) {
		t.Parallel()
		renderer := newFakeRenderer("# PRD", "# DDD", "# ARCH")
		writer := newFakeFileWriterA()
		handler := application.NewArtifactGenerationHandler(renderer, writer)
		event := makeStandardEvent()
		preview, _ := handler.BuildPreview(context.Background(), event)

		err := handler.WriteArtifacts(context.Background(), preview, "/tmp/test")

		require.NoError(t, err)
		assert.Len(t, writer.written, 3)
		hasPRD := false
		hasDDD := false
		hasArch := false
		for p := range writer.written {
			if strings.Contains(p, "PRD.md") {
				hasPRD = true
			}
			if strings.Contains(p, "DDD.md") {
				hasDDD = true
			}
			if strings.Contains(p, "ARCHITECTURE.md") {
				hasArch = true
			}
		}
		assert.True(t, hasPRD)
		assert.True(t, hasDDD)
		assert.True(t, hasArch)
	})

	t.Run("writes preview content", func(t *testing.T) {
		t.Parallel()
		renderer := newFakeRenderer("PRD body", "DDD body", "ARCH body")
		writer := newFakeFileWriterA()
		handler := application.NewArtifactGenerationHandler(renderer, writer)
		event := makeStandardEvent()
		preview, _ := handler.BuildPreview(context.Background(), event)

		handler.WriteArtifacts(context.Background(), preview, "/tmp/out")

		for p, content := range writer.written {
			if strings.Contains(p, "PRD.md") {
				assert.Equal(t, "PRD body", content)
			}
			if strings.Contains(p, "DDD.md") {
				assert.Contains(t, content, "DDD body")
			}
			if strings.Contains(p, "ARCHITECTURE.md") {
				assert.Equal(t, "ARCH body", content)
			}
		}
	})

	t.Run("write does not re-render", func(t *testing.T) {
		t.Parallel()
		renderer := newFakeRenderer("", "", "")
		writer := newFakeFileWriterA()
		handler := application.NewArtifactGenerationHandler(renderer, writer)
		event := makeStandardEvent()
		preview, _ := handler.BuildPreview(context.Background(), event)

		// Reset counts after preview
		renderer.callCount = make(map[string]int)
		handler.WriteArtifacts(context.Background(), preview, "/tmp/out")

		assert.Equal(t, 0, renderer.callCount["RenderPRD"])
		assert.Equal(t, 0, renderer.callCount["RenderDDD"])
		assert.Equal(t, 0, renderer.callCount["RenderArchitecture"])
	})
}

// ---------------------------------------------------------------------------
// Tests — Generate (convenience)
// ---------------------------------------------------------------------------

func TestArtifactGenerationHandler_Generate(t *testing.T) {
	t.Parallel()

	t.Run("generates and writes", func(t *testing.T) {
		t.Parallel()
		renderer := newFakeRenderer("# PRD", "# DDD", "# ARCH")
		writer := newFakeFileWriterA()
		handler := application.NewArtifactGenerationHandler(renderer, writer)
		event := makeStandardEvent()

		model, err := handler.Generate(context.Background(), event, "/tmp/test")

		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(model.Events()), 1)
		assert.Len(t, writer.written, 3)
	})
}

// ---------------------------------------------------------------------------
// Tests — No Default Context
// ---------------------------------------------------------------------------

func TestArtifactGenerationHandler_NoDefaultContext(t *testing.T) {
	t.Parallel()

	t.Run("terms use real context name", func(t *testing.T) {
		t.Parallel()
		renderer := newFakeRenderer("", "", "")
		writer := newFakeFileWriterA()
		handler := application.NewArtifactGenerationHandler(renderer, writer)
		event := makeStandardEvent()

		preview, err := handler.BuildPreview(context.Background(), event)

		require.NoError(t, err)
		for _, term := range preview.Model.UbiquitousLanguage().Terms() {
			assert.NotEqual(t, "Default", term.ContextName(),
				"Term '%s' should not be assigned to 'Default' context", term.Term())
		}
	})
}

// ---------------------------------------------------------------------------
// Tests — No Artificial Relationships
// ---------------------------------------------------------------------------

func TestArtifactGenerationHandler_NoArtificialRelationships(t *testing.T) {
	t.Parallel()

	t.Run("no manufactured relationships", func(t *testing.T) {
		t.Parallel()
		renderer := newFakeRenderer("", "", "")
		writer := newFakeFileWriterA()
		handler := application.NewArtifactGenerationHandler(renderer, writer)
		event := makeStandardEvent()

		preview, err := handler.BuildPreview(context.Background(), event)

		require.NoError(t, err)
		assert.Empty(t, preview.Model.ContextRelationships())
	})
}

// ---------------------------------------------------------------------------
// Tests — No Silent SUPPORTING Default
// ---------------------------------------------------------------------------

func TestArtifactGenerationHandler_NoSilentSupportingDefault(t *testing.T) {
	t.Parallel()

	t.Run("empty Q10 leaves contexts unclassified", func(t *testing.T) {
		t.Parallel()
		renderer := newFakeRenderer("", "", "")
		writer := newFakeFileWriterA()
		handler := application.NewArtifactGenerationHandler(renderer, writer)
		answers := []discoverydomain.Answer{
			discoverydomain.NewAnswer("Q1", "Customer"),
			discoverydomain.NewAnswer("Q3", "Customer places order"),
			discoverydomain.NewAnswer("Q4", "Order must have at least one item"),
			discoverydomain.NewAnswer("Q9", "Sales, Inventory"),
			discoverydomain.NewAnswer("Q10", ""),
		}
		event := makeEventWithAnswers(answers)

		_, err := handler.BuildPreview(context.Background(), event)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "has no classification")
	})

	t.Run("missing Q10 leaves contexts unclassified", func(t *testing.T) {
		t.Parallel()
		renderer := newFakeRenderer("", "", "")
		writer := newFakeFileWriterA()
		handler := application.NewArtifactGenerationHandler(renderer, writer)
		answers := []discoverydomain.Answer{
			discoverydomain.NewAnswer("Q1", "Customer"),
			discoverydomain.NewAnswer("Q3", "Customer places order"),
			discoverydomain.NewAnswer("Q4", "Order must have at least one item"),
			discoverydomain.NewAnswer("Q9", "Sales"),
		}
		event := makeEventWithAnswers(answers)

		_, err := handler.BuildPreview(context.Background(), event)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "has no classification")
	})

	t.Run("resolved context gets classified", func(t *testing.T) {
		t.Parallel()
		renderer := newFakeRenderer("", "", "")
		writer := newFakeFileWriterA()
		handler := application.NewArtifactGenerationHandler(renderer, writer)
		answers := []discoverydomain.Answer{
			discoverydomain.NewAnswer("Q1", "Customer"),
			discoverydomain.NewAnswer("Q3", "Customer places order"),
			discoverydomain.NewAnswer("Q4", "Order must have at least one item"),
			discoverydomain.NewAnswer("Q9", "Sales"),
			discoverydomain.NewAnswer("Q10", "Sales is core competitive advantage"),
		}
		event := makeEventWithAnswers(answers)

		preview, err := handler.BuildPreview(context.Background(), event)

		require.NoError(t, err)
		bcs := preview.Model.BoundedContexts()
		require.Len(t, bcs, 1)
		assert.NotNil(t, bcs[0].Classification())
	})
}

// ---------------------------------------------------------------------------
// Tests — MVP Questions Only
// ---------------------------------------------------------------------------

func TestArtifactGenerationHandler_MVPQuestions(t *testing.T) {
	t.Parallel()

	t.Run("mvp questions only", func(t *testing.T) {
		t.Parallel()
		renderer := newFakeRenderer("", "", "")
		writer := newFakeFileWriterA()
		handler := application.NewArtifactGenerationHandler(renderer, writer)
		answers := []discoverydomain.Answer{
			discoverydomain.NewAnswer("Q1", "Customer"),
			discoverydomain.NewAnswer("Q3", "Customer places order, System confirms"),
			discoverydomain.NewAnswer("Q4", "Order must have at least one item"),
			discoverydomain.NewAnswer("Q9", "Orders"),
			discoverydomain.NewAnswer("Q10", "Orders is core competitive advantage"),
		}
		event := makeEventWithAnswers(answers)

		preview, err := handler.BuildPreview(context.Background(), event)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(preview.Model.Events()), 1)
		assert.Len(t, preview.Model.BoundedContexts(), 1)
	})
}

// ---------------------------------------------------------------------------
// Tests — SplitAnswer
// ---------------------------------------------------------------------------

func TestSplitAnswer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"comma separated", "Order, Product, Customer", []string{"Order", "Product", "Customer"}},
		{"newline separated", "1. Order\n2. Product", []string{"Order", "Product"}},
		{"single item", "Order", []string{"Order"}},
		{"empty string", "", nil},
		{"whitespace only", "   ", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := application.SplitAnswer(tt.input)
			if tt.expected == nil {
				assert.Empty(t, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
