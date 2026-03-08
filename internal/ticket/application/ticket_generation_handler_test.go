package application_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
	"github.com/alty-cli/alty/internal/ticket/application"
)

// ---------------------------------------------------------------------------
// Fakes
// ---------------------------------------------------------------------------

type fakeFileWriterT struct {
	written map[string]string
}

func newFakeFileWriterT() *fakeFileWriterT {
	return &fakeFileWriterT{written: make(map[string]string)}
}

func (f *fakeFileWriterT) WriteFile(_ context.Context, path, content string) error {
	f.written[path] = content
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeTicketModel(contexts []struct {
	Name           string
	Classification vo.SubdomainClassification
}) *ddd.DomainModel {
	model := ddd.NewDomainModel("ticket-test")

	names := make([]string, len(contexts))
	for i, c := range contexts {
		names[i] = c.Name
	}
	steps := make([]string, len(names))
	for i, name := range names {
		steps[i] = "User manages " + name
	}
	story := vo.NewDomainStory("Test flow", []string{"User"}, "User starts", steps, nil)
	model.AddDomainStory(story)

	for _, c := range contexts {
		model.AddTerm(c.Name, c.Name+" domain", c.Name, nil)
		bc := vo.NewDomainBoundedContext(c.Name, "Manages "+c.Name, nil, nil, "")
		model.AddBoundedContext(bc)
		model.ClassifySubdomain(c.Name, c.Classification, "test")

		if c.Classification == vo.SubdomainCore {
			agg := vo.NewAggregateDesign(c.Name+"Root", c.Name, c.Name+"Root",
				nil, []string{"must be valid"}, nil, nil)
			model.DesignAggregate(agg)
		}
	}
	model.Finalize()
	return model
}

// ---------------------------------------------------------------------------
// Tests — Build Preview
// ---------------------------------------------------------------------------

func TestTicketGenerationHandler_BuildPreview(t *testing.T) {
	t.Parallel()

	t.Run("returns preview", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterT()
		handler := application.NewTicketGenerationHandler(writer)
		model := makeTicketModel([]struct {
			Name           string
			Classification vo.SubdomainClassification
		}{{"Orders", vo.SubdomainCore}})

		preview, err := handler.BuildPreview(model, nil)

		require.NoError(t, err)
		require.NotNil(t, preview)
		assert.NotNil(t, preview.Plan)
		assert.NotEmpty(t, preview.Summary)
	})

	t.Run("does not write", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterT()
		handler := application.NewTicketGenerationHandler(writer)
		model := makeTicketModel([]struct {
			Name           string
			Classification vo.SubdomainClassification
		}{{"Orders", vo.SubdomainCore}})

		handler.BuildPreview(model, nil)

		assert.Equal(t, 0, len(writer.written))
	})

	t.Run("contains bc names", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterT()
		handler := application.NewTicketGenerationHandler(writer)
		model := makeTicketModel([]struct {
			Name           string
			Classification vo.SubdomainClassification
		}{
			{"Orders", vo.SubdomainCore},
			{"Logging", vo.SubdomainGeneric},
		})

		preview, err := handler.BuildPreview(model, nil)

		require.NoError(t, err)
		assert.Contains(t, preview.Summary, "Orders")
		assert.Contains(t, preview.Summary, "Logging")
	})

	t.Run("empty model raises", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterT()
		handler := application.NewTicketGenerationHandler(writer)
		model := ddd.NewDomainModel("empty")

		_, err := handler.BuildPreview(model, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no bounded contexts")
	})

	t.Run("includes validation", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterT()
		handler := application.NewTicketGenerationHandler(writer)
		model := makeTicketModel([]struct {
			Name           string
			Classification vo.SubdomainClassification
		}{{"Orders", vo.SubdomainCore}})

		preview, err := handler.BuildPreview(model, nil)

		require.NoError(t, err)
		assert.Equal(t, len(preview.Plan.Tickets()), len(preview.Validation))
	})

	t.Run("validation ticket ids match plan", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterT()
		handler := application.NewTicketGenerationHandler(writer)
		model := makeTicketModel([]struct {
			Name           string
			Classification vo.SubdomainClassification
		}{
			{"Orders", vo.SubdomainCore},
			{"Logging", vo.SubdomainGeneric},
		})

		preview, err := handler.BuildPreview(model, nil)

		require.NoError(t, err)
		planIDs := make(map[string]bool)
		for _, t := range preview.Plan.Tickets() {
			planIDs[t.TicketID()] = true
		}
		for _, result := range preview.Validation {
			assert.True(t, planIDs[result.TicketID()])
		}
	})
}

// ---------------------------------------------------------------------------
// Tests — Approve and Write
// ---------------------------------------------------------------------------

func TestTicketGenerationHandler_ApproveAndWrite(t *testing.T) {
	t.Parallel()

	t.Run("writes files", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterT()
		handler := application.NewTicketGenerationHandler(writer)
		model := makeTicketModel([]struct {
			Name           string
			Classification vo.SubdomainClassification
		}{{"Orders", vo.SubdomainCore}})
		preview, _ := handler.BuildPreview(model, nil)

		err := handler.ApproveAndWrite(context.Background(), preview, "/project", nil)

		require.NoError(t, err)
		assert.True(t, len(writer.written) >= 2)
		hasSummary := false
		hasTicket := false
		for p := range writer.written {
			if strings.Contains(p, "PLAN_SUMMARY") {
				hasSummary = true
			}
			if strings.Contains(p, "tickets") {
				hasTicket = true
			}
		}
		assert.True(t, hasSummary)
		assert.True(t, hasTicket)
	})

	t.Run("emits event", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterT()
		handler := application.NewTicketGenerationHandler(writer)
		model := makeTicketModel([]struct {
			Name           string
			Classification vo.SubdomainClassification
		}{{"Orders", vo.SubdomainCore}})
		preview, _ := handler.BuildPreview(model, nil)

		handler.ApproveAndWrite(context.Background(), preview, "/project", nil)

		assert.Equal(t, 1, len(preview.Plan.Events()))
	})

	t.Run("approve twice raises", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterT()
		handler := application.NewTicketGenerationHandler(writer)
		model := makeTicketModel([]struct {
			Name           string
			Classification vo.SubdomainClassification
		}{{"Orders", vo.SubdomainCore}})
		preview, _ := handler.BuildPreview(model, nil)

		handler.ApproveAndWrite(context.Background(), preview, "/project", nil)
		err := handler.ApproveAndWrite(context.Background(), preview, "/project", nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "already approved")
	})

	t.Run("approve subset only writes approved", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterT()
		handler := application.NewTicketGenerationHandler(writer)
		model := makeTicketModel([]struct {
			Name           string
			Classification vo.SubdomainClassification
		}{
			{"Orders", vo.SubdomainCore},
			{"Logging", vo.SubdomainGeneric},
		})
		preview, _ := handler.BuildPreview(model, nil)

		tickets := preview.Plan.Tickets()
		firstID := tickets[0].TicketID()
		err := handler.ApproveAndWrite(context.Background(), preview, "/project", []string{firstID})

		require.NoError(t, err)
		// Summary + 1 approved ticket = 2 files
		assert.Equal(t, 2, len(writer.written))
	})
}
