// Package integration provides cross-context integration tests.
package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
	"github.com/alto-cli/alto/internal/ticket/application"
	ticketdomain "github.com/alto-cli/alto/internal/ticket/domain"
)

// ---------------------------------------------------------------------------
// Fakes for ticket generation tests
// ---------------------------------------------------------------------------

type fakeEventPublisher struct {
	published []any
}

func (f *fakeEventPublisher) Publish(_ context.Context, event any) error {
	f.published = append(f.published, event)
	return nil
}

type fakeFileWriter struct {
	written map[string]string
}

func newFakeFileWriter() *fakeFileWriter {
	return &fakeFileWriter{written: make(map[string]string)}
}

func (f *fakeFileWriter) WriteFile(_ context.Context, path, content string) error {
	f.written[path] = content
	return nil
}

type recordingBeadsWriter struct {
	epics        []ticketdomain.GeneratedEpic
	tickets      []ticketdomain.GeneratedTicket
	dependencies []struct{ from, to string }
	epicIDMap    map[string]string
	ticketIDMap  map[string]string
}

func newRecordingBeadsWriter() *recordingBeadsWriter {
	return &recordingBeadsWriter{
		epicIDMap:   make(map[string]string),
		ticketIDMap: make(map[string]string),
	}
}

func (w *recordingBeadsWriter) WriteEpic(_ context.Context, epic ticketdomain.GeneratedEpic) (string, error) {
	w.epics = append(w.epics, epic)
	beadsID := "beads-epic-" + epic.EpicID()
	w.epicIDMap[epic.EpicID()] = beadsID
	return beadsID, nil
}

func (w *recordingBeadsWriter) WriteTicket(_ context.Context, ticket ticketdomain.GeneratedTicket) (string, error) {
	w.tickets = append(w.tickets, ticket)
	beadsID := "beads-ticket-" + ticket.TicketID()
	w.ticketIDMap[ticket.TicketID()] = beadsID
	return beadsID, nil
}

func (w *recordingBeadsWriter) SetDependency(_ context.Context, ticketID, dependsOnID string) error {
	w.dependencies = append(w.dependencies, struct{ from, to string }{ticketID, dependsOnID})
	return nil
}

// ---------------------------------------------------------------------------
// Full Pipeline Integration Tests
// ---------------------------------------------------------------------------

func TestTicketGeneration_FullPipeline(t *testing.T) {
	t.Parallel()

	t.Run("creates epics and tickets from domain model", func(t *testing.T) {
		t.Parallel()
		model := makeMultiContextModel(t)
		beadsWriter := newRecordingBeadsWriter()
		handler := makeHandler(beadsWriter)

		preview, err := handler.BuildPreview(model, nil)
		require.NoError(t, err)

		err = handler.ApproveAndWriteToBeads(context.Background(), preview, nil)
		require.NoError(t, err)

		// Should have 2 epics (Order + Payment contexts)
		assert.Len(t, beadsWriter.epics, 2)

		// Should have at least 2 tickets (one per context)
		assert.GreaterOrEqual(t, len(beadsWriter.tickets), 2)

		// Verify epic names
		epicNames := make([]string, len(beadsWriter.epics))
		for i, epic := range beadsWriter.epics {
			epicNames[i] = epic.Title()
		}
		assert.Contains(t, epicNames, "Order Epic")
		assert.Contains(t, epicNames, "Payment Epic")
	})

	t.Run("sets dependencies between contexts", func(t *testing.T) {
		t.Parallel()
		model := makeMultiContextModel(t)
		beadsWriter := newRecordingBeadsWriter()
		handler := makeHandler(beadsWriter)

		preview, err := handler.BuildPreview(model, nil)
		require.NoError(t, err)

		err = handler.ApproveAndWriteToBeads(context.Background(), preview, nil)
		require.NoError(t, err)

		// Should have dependencies (Payment depends on Order)
		assert.NotEmpty(t, beadsWriter.dependencies, "expected cross-context dependencies")
	})

	t.Run("respects detail level from subdomain classification", func(t *testing.T) {
		t.Parallel()
		model := makeMixedClassificationModel(t)
		beadsWriter := newRecordingBeadsWriter()
		handler := makeHandler(beadsWriter)

		preview, err := handler.BuildPreview(model, nil)
		require.NoError(t, err)

		err = handler.ApproveAndWriteToBeads(context.Background(), preview, nil)
		require.NoError(t, err)

		// Find tickets by context and verify detail levels
		var coreTicket, genericTicket ticketdomain.GeneratedTicket
		for _, ticket := range beadsWriter.tickets {
			switch ticket.BoundedContextName() {
			case "Order":
				coreTicket = ticket
			case "Logging":
				genericTicket = ticket
			}
		}

		// Core context should have FULL detail
		assert.Equal(t, vo.TicketDetailFull, coreTicket.DetailLevel())
		// Generic context should have STUB detail
		assert.Equal(t, vo.TicketDetailStub, genericTicket.DetailLevel())
	})

	t.Run("publishes TicketPlanApproved event", func(t *testing.T) {
		t.Parallel()
		model := makeMultiContextModel(t)
		beadsWriter := newRecordingBeadsWriter()
		publisher := &fakeEventPublisher{}
		fileWriter := newFakeFileWriter()

		handler := application.NewTicketGenerationHandler(fileWriter, publisher)
		handler.SetBeadsWriter(beadsWriter)

		preview, err := handler.BuildPreview(model, nil)
		require.NoError(t, err)

		err = handler.ApproveAndWriteToBeads(context.Background(), preview, nil)
		require.NoError(t, err)

		require.Len(t, publisher.published, 1)
		_, ok := publisher.published[0].(ticketdomain.TicketPlanApproved)
		assert.True(t, ok, "expected TicketPlanApproved event")
	})

	t.Run("partial approval only creates approved tickets", func(t *testing.T) {
		t.Parallel()
		model := makeMultiContextModel(t)
		beadsWriter := newRecordingBeadsWriter()
		handler := makeHandler(beadsWriter)

		preview, err := handler.BuildPreview(model, nil)
		require.NoError(t, err)

		// Approve only the first ticket
		firstTicketID := preview.Plan.Tickets()[0].TicketID()
		err = handler.ApproveAndWriteToBeads(context.Background(), preview, []string{firstTicketID})
		require.NoError(t, err)

		// Should have all epics but only 1 ticket
		assert.Len(t, beadsWriter.epics, 2) // epics are always created
		assert.Len(t, beadsWriter.tickets, 1)
	})
}

func TestTicketGeneration_TemplateTypes(t *testing.T) {
	t.Parallel()

	t.Run("epics have TemplateEpic type", func(t *testing.T) {
		t.Parallel()
		model := makeMultiContextModel(t)
		beadsWriter := newRecordingBeadsWriter()
		handler := makeHandler(beadsWriter)

		preview, err := handler.BuildPreview(model, nil)
		require.NoError(t, err)

		err = handler.ApproveAndWriteToBeads(context.Background(), preview, nil)
		require.NoError(t, err)

		for _, epic := range beadsWriter.epics {
			assert.Equal(t, ticketdomain.TemplateEpic, epic.TemplateType())
		}
	})

	t.Run("regular tickets have TemplateTask type", func(t *testing.T) {
		t.Parallel()
		model := makeMultiContextModel(t)
		beadsWriter := newRecordingBeadsWriter()
		handler := makeHandler(beadsWriter)

		preview, err := handler.BuildPreview(model, nil)
		require.NoError(t, err)

		err = handler.ApproveAndWriteToBeads(context.Background(), preview, nil)
		require.NoError(t, err)

		for _, ticket := range beadsWriter.tickets {
			// All generated tickets from aggregates are tasks
			assert.Equal(t, ticketdomain.TemplateTask, ticket.TemplateType())
		}
	})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeHandler(beadsWriter *recordingBeadsWriter) *application.TicketGenerationHandler {
	handler := application.NewTicketGenerationHandler(newFakeFileWriter(), &fakeEventPublisher{})
	handler.SetBeadsWriter(beadsWriter)
	return handler
}

func makeMultiContextModel(t *testing.T) *ddd.DomainModel {
	t.Helper()
	model := ddd.NewDomainModel("multi-context-test")

	// Add domain story
	story := vo.NewDomainStory(
		"Order and Payment Flow",
		[]string{"Customer"},
		"Customer places order",
		[]string{"Customer creates order", "Customer makes payment"},
		nil,
	)
	model.AddDomainStory(story)

	// Add terms
	model.AddTerm("Order", "A customer order", "Order", nil)
	model.AddTerm("Payment", "A payment for an order", "Payment", nil)

	// Add bounded contexts
	orderBC := vo.NewDomainBoundedContext("Order", "Manages orders", nil, nil, "")
	paymentBC := vo.NewDomainBoundedContext("Payment", "Manages payments", nil, nil, "")
	model.AddBoundedContext(orderBC)
	model.AddBoundedContext(paymentBC)

	// Classify subdomains
	model.ClassifySubdomain("Order", vo.SubdomainCore, "core business")
	model.ClassifySubdomain("Payment", vo.SubdomainCore, "core business")

	// Add context relationship (Payment depends on Order)
	rel := vo.NewContextRelationship("Order", "Payment", "customer-supplier")
	model.AddContextRelationship(rel)

	// Design aggregates
	orderAgg := vo.NewAggregateDesign("OrderRoot", "Order", "OrderRoot",
		nil, []string{"order must have items"}, nil, nil)
	paymentAgg := vo.NewAggregateDesign("PaymentRoot", "Payment", "PaymentRoot",
		nil, []string{"payment must be positive"}, nil, nil)
	model.DesignAggregate(orderAgg)
	model.DesignAggregate(paymentAgg)

	model.Finalize()
	return model
}

func makeMixedClassificationModel(t *testing.T) *ddd.DomainModel {
	t.Helper()
	model := ddd.NewDomainModel("mixed-classification-test")

	// Add domain story
	story := vo.NewDomainStory(
		"Order with Logging",
		[]string{"System"},
		"System processes order",
		[]string{"System creates order", "System logs event"},
		nil,
	)
	model.AddDomainStory(story)

	// Add terms
	model.AddTerm("Order", "A customer order", "Order", nil)
	model.AddTerm("Logging", "System logging", "Logging", nil)

	// Add bounded contexts
	orderBC := vo.NewDomainBoundedContext("Order", "Manages orders", nil, nil, "")
	loggingBC := vo.NewDomainBoundedContext("Logging", "Logging infrastructure", nil, nil, "")
	model.AddBoundedContext(orderBC)
	model.AddBoundedContext(loggingBC)

	// Classify subdomains - Order is Core, Logging is Generic
	model.ClassifySubdomain("Order", vo.SubdomainCore, "core business")
	model.ClassifySubdomain("Logging", vo.SubdomainGeneric, "off-the-shelf")

	// Design aggregates
	orderAgg := vo.NewAggregateDesign("OrderRoot", "Order", "OrderRoot",
		nil, []string{"order must have items"}, nil, nil)
	model.DesignAggregate(orderAgg)

	model.Finalize()
	return model
}
