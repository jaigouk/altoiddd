package application

import (
	"context"
	"fmt"
	"path/filepath"

	sharedapp "github.com/alty-cli/alty/internal/shared/application"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
	ticketdomain "github.com/alty-cli/alty/internal/ticket/domain"
)

// TicketPreview holds the generated ticket plan ready for user review.
type TicketPreview struct {
	Plan       *ticketdomain.TicketPlan
	Summary    string
	Validation []ticketdomain.DesignTraceResult
}

// TicketGenerationHandler orchestrates ticket pipeline generation from a DomainModel.
type TicketGenerationHandler struct {
	fileWriter sharedapp.FileWriter
}

// NewTicketGenerationHandler creates a new TicketGenerationHandler.
func NewTicketGenerationHandler(fileWriter sharedapp.FileWriter) *TicketGenerationHandler {
	return &TicketGenerationHandler{fileWriter: fileWriter}
}

// BuildPreview generates a ticket plan for preview without writing files.
func (h *TicketGenerationHandler) BuildPreview(
	model *ddd.DomainModel,
	profile vo.StackProfile,
) (*TicketPreview, error) {
	plan := ticketdomain.NewTicketPlan()
	if err := plan.GeneratePlan(model, profile); err != nil {
		return nil, err
	}

	summary, err := plan.Preview()
	if err != nil {
		return nil, err
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
		return err
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
				return err
			}
		}
	}

	summaryPath := filepath.Join(outputDir, "tickets", "PLAN_SUMMARY.md")
	return h.fileWriter.WriteFile(ctx, summaryPath, preview.Summary)
}
