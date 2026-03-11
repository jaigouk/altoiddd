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
	ticketdomain "github.com/alty-cli/alty/internal/ticket/domain"
)

// ---------------------------------------------------------------------------
// Fakes
// ---------------------------------------------------------------------------

type fakePublisherT struct {
	published []any
}

func (f *fakePublisherT) Publish(_ context.Context, event any) error {
	f.published = append(f.published, event)
	return nil
}

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

type fakeBeadsWriterT struct {
	epics         []ticketdomain.GeneratedEpic
	tickets       []ticketdomain.GeneratedTicket
	dependencies  []struct{ from, to string }
	epicCounter   int
	ticketCounter int
}

func newFakeBeadsWriterT() *fakeBeadsWriterT {
	return &fakeBeadsWriterT{}
}

func (f *fakeBeadsWriterT) WriteEpic(_ context.Context, epic ticketdomain.GeneratedEpic) (string, error) {
	f.epics = append(f.epics, epic)
	f.epicCounter++
	return "beads-epic-" + epic.EpicID(), nil
}

func (f *fakeBeadsWriterT) WriteTicket(_ context.Context, ticket ticketdomain.GeneratedTicket) (string, error) {
	f.tickets = append(f.tickets, ticket)
	f.ticketCounter++
	return "beads-ticket-" + ticket.TicketID(), nil
}

func (f *fakeBeadsWriterT) SetDependency(_ context.Context, ticketID, dependsOnID string) error {
	f.dependencies = append(f.dependencies, struct{ from, to string }{ticketID, dependsOnID})
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeTicketModel(contexts []struct {
	Name           string
	Classification vo.SubdomainClassification
},
) *ddd.DomainModel {
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
		handler := application.NewTicketGenerationHandler(writer, &fakePublisherT{})
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
		handler := application.NewTicketGenerationHandler(writer, &fakePublisherT{})
		model := makeTicketModel([]struct {
			Name           string
			Classification vo.SubdomainClassification
		}{{"Orders", vo.SubdomainCore}})

		handler.BuildPreview(model, nil)

		assert.Empty(t, writer.written)
	})

	t.Run("contains bc names", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterT()
		handler := application.NewTicketGenerationHandler(writer, &fakePublisherT{})
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
		handler := application.NewTicketGenerationHandler(writer, &fakePublisherT{})
		model := ddd.NewDomainModel("empty")

		_, err := handler.BuildPreview(model, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no bounded contexts")
	})

	t.Run("includes validation", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterT()
		handler := application.NewTicketGenerationHandler(writer, &fakePublisherT{})
		model := makeTicketModel([]struct {
			Name           string
			Classification vo.SubdomainClassification
		}{{"Orders", vo.SubdomainCore}})

		preview, err := handler.BuildPreview(model, nil)

		require.NoError(t, err)
		assert.Len(t, preview.Validation, len(preview.Plan.Tickets()))
	})

	t.Run("validation ticket ids match plan", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterT()
		handler := application.NewTicketGenerationHandler(writer, &fakePublisherT{})
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
		handler := application.NewTicketGenerationHandler(writer, &fakePublisherT{})
		model := makeTicketModel([]struct {
			Name           string
			Classification vo.SubdomainClassification
		}{{"Orders", vo.SubdomainCore}})
		preview, _ := handler.BuildPreview(model, nil)

		err := handler.ApproveAndWrite(context.Background(), preview, "/project", nil)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(writer.written), 2)
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
		handler := application.NewTicketGenerationHandler(writer, &fakePublisherT{})
		model := makeTicketModel([]struct {
			Name           string
			Classification vo.SubdomainClassification
		}{{"Orders", vo.SubdomainCore}})
		preview, _ := handler.BuildPreview(model, nil)

		handler.ApproveAndWrite(context.Background(), preview, "/project", nil)

		assert.Len(t, preview.Plan.Events(), 1)
	})

	t.Run("approve twice raises", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriterT()
		handler := application.NewTicketGenerationHandler(writer, &fakePublisherT{})
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
		handler := application.NewTicketGenerationHandler(writer, &fakePublisherT{})
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
		assert.Len(t, writer.written, 2)
	})
}

func TestTicketGenerationHandler_ApproveAndWrite_PublishesEvent(t *testing.T) {
	t.Parallel()

	pub := &fakePublisherT{}
	writer := newFakeFileWriterT()
	handler := application.NewTicketGenerationHandler(writer, pub)
	model := makeTicketModel([]struct {
		Name           string
		Classification vo.SubdomainClassification
	}{{"Orders", vo.SubdomainCore}})
	preview, err := handler.BuildPreview(model, nil)
	require.NoError(t, err)

	err = handler.ApproveAndWrite(context.Background(), preview, "/project", nil)
	require.NoError(t, err)

	require.Len(t, pub.published, 1)
	_, ok := pub.published[0].(ticketdomain.TicketPlanApproved)
	assert.True(t, ok, "expected TicketPlanApproved, got %T", pub.published[0])
}

// ---------------------------------------------------------------------------
// Tests — Beads Integration (pv2.3)
// ---------------------------------------------------------------------------

func TestTicketGenerationHandler_WriteToBeads(t *testing.T) {
	t.Parallel()

	t.Run("creates epics and tickets in beads", func(t *testing.T) {
		t.Parallel()
		fileWriter := newFakeFileWriterT()
		beadsWriter := newFakeBeadsWriterT()
		handler := application.NewTicketGenerationHandler(fileWriter, &fakePublisherT{})
		handler.SetBeadsWriter(beadsWriter)

		model := makeTicketModel([]struct {
			Name           string
			Classification vo.SubdomainClassification
		}{{"Orders", vo.SubdomainCore}})
		preview, _ := handler.BuildPreview(model, nil)

		err := handler.ApproveAndWriteToBeads(context.Background(), preview, nil)

		require.NoError(t, err)
		assert.Len(t, beadsWriter.epics, 1)
		assert.GreaterOrEqual(t, len(beadsWriter.tickets), 1)
	})

	t.Run("sets dependencies between tickets", func(t *testing.T) {
		t.Parallel()
		fileWriter := newFakeFileWriterT()
		beadsWriter := newFakeBeadsWriterT()
		handler := application.NewTicketGenerationHandler(fileWriter, &fakePublisherT{})
		handler.SetBeadsWriter(beadsWriter)

		// Create a model with upstream/downstream relationship
		model := ddd.NewDomainModel("dep-test")
		story := vo.NewDomainStory("Flow", []string{"User"}, "Start",
			[]string{"User orders", "User pays"}, nil)
		model.AddDomainStory(story)

		model.AddTerm("Order", "Order entity", "Order", nil)
		model.AddTerm("Payment", "Payment entity", "Payment", nil)

		orderBC := vo.NewDomainBoundedContext("Order", "Manages orders", nil, nil, "")
		paymentBC := vo.NewDomainBoundedContext("Payment", "Manages payments", nil, nil, "")
		model.AddBoundedContext(orderBC)
		model.AddBoundedContext(paymentBC)
		model.ClassifySubdomain("Order", vo.SubdomainCore, "test")
		model.ClassifySubdomain("Payment", vo.SubdomainCore, "test")

		// Payment depends on Order (downstream)
		rel := vo.NewContextRelationship("Order", "Payment", "customer-supplier")
		model.AddContextRelationship(rel)

		orderAgg := vo.NewAggregateDesign("OrderRoot", "Order", "OrderRoot", nil, nil, nil, nil)
		paymentAgg := vo.NewAggregateDesign("PaymentRoot", "Payment", "PaymentRoot", nil, nil, nil, nil)
		model.DesignAggregate(orderAgg)
		model.DesignAggregate(paymentAgg)
		model.Finalize()

		preview, _ := handler.BuildPreview(model, nil)
		err := handler.ApproveAndWriteToBeads(context.Background(), preview, nil)

		require.NoError(t, err)
		// Should have dependencies set (Payment tickets depend on Order tickets)
		assert.NotEmpty(t, beadsWriter.dependencies)
	})
}
