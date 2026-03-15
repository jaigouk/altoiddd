package application

import (
	"context"
	"fmt"
	"path/filepath"

	sharedapp "github.com/alty-cli/alty/internal/shared/application"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
	ticketdomain "github.com/alty-cli/alty/internal/ticket/domain"
)

// TicketPreview holds the generated ticket plan ready for user review.
type TicketPreview struct {
	Plan       *ticketdomain.TicketPlan
	Summary    string
	Validation []ticketdomain.DesignTraceResult
	warnings   []string
}

// Warnings returns a defensive copy of generation warnings.
func (p *TicketPreview) Warnings() []string {
	out := make([]string, len(p.warnings))
	copy(out, p.warnings)
	return out
}

// TicketGenerationHandler orchestrates ticket pipeline generation from a DomainModel.
type TicketGenerationHandler struct {
	fileWriter  sharedapp.FileWriter
	beadsWriter BeadsWriter
	publisher   sharedapp.EventPublisher
}

// NewTicketGenerationHandler creates a new TicketGenerationHandler.
func NewTicketGenerationHandler(fileWriter sharedapp.FileWriter, publisher sharedapp.EventPublisher) *TicketGenerationHandler {
	return &TicketGenerationHandler{fileWriter: fileWriter, publisher: publisher}
}

// SetBeadsWriter sets the optional BeadsWriter for creating beads issues.
func (h *TicketGenerationHandler) SetBeadsWriter(writer BeadsWriter) {
	h.beadsWriter = writer
}

// BuildPreview generates a ticket plan for preview without writing files.
// Returns a partial preview with warnings when the model is incomplete.
// Returns an error only when the model is truly empty (zero contexts, aggregates, and stories).
func (h *TicketGenerationHandler) BuildPreview(
	model *ddd.DomainModel,
	profile vo.StackProfile,
) (*TicketPreview, error) {
	if model.IsEmpty() {
		return nil, fmt.Errorf("model is empty, nothing to generate; run 'alty guide' or 'alty import' first: %w",
			domainerrors.ErrInvariantViolation)
	}

	plan := ticketdomain.NewTicketPlan()
	if err := plan.GeneratePlan(model, profile); err != nil {
		// Partial model — convert generation error to warning
		return &TicketPreview{
			Plan:     plan,
			Summary:  "Partial generation — see warnings",
			warnings: []string{fmt.Sprintf("ticket generation limited: %s", err.Error())},
		}, nil
	}

	summary, err := plan.Preview()
	if err != nil {
		return nil, fmt.Errorf("previewing ticket plan: %w", err)
	}

	validation := ticketdomain.ValidateImplementabilityPlan(plan.Tickets())

	return &TicketPreview{
		Plan:       plan,
		Summary:    summary,
		Validation: validation,
	}, nil
}

// ApproveAndWrite approves the plan and writes tickets to disk.
func (h *TicketGenerationHandler) ApproveAndWrite(
	ctx context.Context,
	preview *TicketPreview,
	outputDir string,
	approvedIDs []string,
) error {
	if err := preview.Plan.Approve(approvedIDs); err != nil {
		return fmt.Errorf("approving ticket plan: %w", err)
	}

	events := preview.Plan.Events()
	if len(events) == 0 {
		return fmt.Errorf("no approval event found after approve")
	}
	lastEvent := events[len(events)-1]
	approvedSet := make(map[string]bool)
	for _, id := range lastEvent.ApprovedTicketIDs() {
		approvedSet[id] = true
	}

	for _, ticket := range preview.Plan.Tickets() {
		if approvedSet[ticket.TicketID()] {
			ticketPath := filepath.Join(outputDir, "tickets", ticket.TicketID()+".md")
			if err := h.fileWriter.WriteFile(ctx, ticketPath, ticket.Description()); err != nil {
				return fmt.Errorf("writing ticket %s: %w", ticket.TicketID(), err)
			}
		}
	}

	summaryPath := filepath.Join(outputDir, "tickets", "PLAN_SUMMARY.md")
	if err := h.fileWriter.WriteFile(ctx, summaryPath, preview.Summary); err != nil {
		return fmt.Errorf("writing plan summary: %w", err)
	}
	for _, event := range preview.Plan.Events() {
		_ = h.publisher.Publish(ctx, event)
	}
	return nil
}

// ApproveAndWriteToBeads approves the plan and creates beads issues.
// Requires SetBeadsWriter to have been called first.
func (h *TicketGenerationHandler) ApproveAndWriteToBeads(
	ctx context.Context,
	preview *TicketPreview,
	approvedIDs []string,
) error {
	if h.beadsWriter == nil {
		return fmt.Errorf("beads writer not configured")
	}

	if err := preview.Plan.Approve(approvedIDs); err != nil {
		return fmt.Errorf("approving ticket plan: %w", err)
	}

	events := preview.Plan.Events()
	if len(events) == 0 {
		return fmt.Errorf("no approval event found after approve")
	}
	lastEvent := events[len(events)-1]
	approvedSet := make(map[string]bool)
	for _, id := range lastEvent.ApprovedTicketIDs() {
		approvedSet[id] = true
	}

	// Map from generated IDs to beads IDs
	idMapping := make(map[string]string)

	// First, create all epics
	for _, epic := range preview.Plan.Epics() {
		beadsID, err := h.beadsWriter.WriteEpic(ctx, epic)
		if err != nil {
			return fmt.Errorf("creating epic %s: %w", epic.Title(), err)
		}
		idMapping[epic.EpicID()] = beadsID
	}

	// Then, create all tickets
	for _, ticket := range preview.Plan.Tickets() {
		if !approvedSet[ticket.TicketID()] {
			continue
		}
		beadsID, err := h.beadsWriter.WriteTicket(ctx, ticket)
		if err != nil {
			return fmt.Errorf("creating ticket %s: %w", ticket.Title(), err)
		}
		idMapping[ticket.TicketID()] = beadsID
	}

	// Finally, set dependencies using beads IDs
	for _, ticket := range preview.Plan.Tickets() {
		if !approvedSet[ticket.TicketID()] {
			continue
		}
		ticketBeadsID := idMapping[ticket.TicketID()]
		for _, depID := range ticket.Dependencies() {
			depBeadsID, ok := idMapping[depID]
			if !ok {
				// Dependency not in this plan (external or dismissed)
				continue
			}
			if err := h.beadsWriter.SetDependency(ctx, ticketBeadsID, depBeadsID); err != nil {
				return fmt.Errorf("setting dependency %s -> %s: %w", ticketBeadsID, depBeadsID, err)
			}
		}
	}

	for _, event := range preview.Plan.Events() {
		_ = h.publisher.Publish(ctx, event)
	}
	return nil
}
