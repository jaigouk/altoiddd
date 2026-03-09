package composition

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	bootstrapdomain "github.com/alty-cli/alty/internal/bootstrap/domain"
	discoverydomain "github.com/alty-cli/alty/internal/discovery/domain"
	fitnessdomain "github.com/alty-cli/alty/internal/fitness/domain"
	shareddomain "github.com/alty-cli/alty/internal/shared/domain"
	"github.com/alty-cli/alty/internal/shared/domain/events"
	"github.com/alty-cli/alty/internal/shared/infrastructure/eventbus"
	ticketdomain "github.com/alty-cli/alty/internal/ticket/domain"
	ttdomain "github.com/alty-cli/alty/internal/tooltranslation/domain"
)

// wireEventSubscribers creates a Subscriber with handlers for all domain events.
// Tier 1: Observability logging (structured slog output for each event).
// Tier 2: Readiness tracking (updates SessionTracker based on workflow events).
func wireEventSubscribers(
	bus *eventbus.Bus,
	logger *slog.Logger,
	tracker *shareddomain.SessionTracker,
) (*eventbus.Subscriber, error) {
	sub := eventbus.NewSubscriber(bus)
	var errs []error

	// ===========================================================================
	// Tier 1 — Observability (logging only)
	// ===========================================================================

	// --- Shared events ---

	errs = append(errs, eventbus.SubscribeTyped(sub, func(ctx context.Context, evt *events.DomainModelGenerated) error {
		logger.InfoContext(ctx, "event.received",
			"type", "DomainModelGenerated",
			"model_id", evt.ModelID(),
			"bounded_contexts", len(evt.BoundedContexts()),
		)
		return nil
	}))

	errs = append(errs, eventbus.SubscribeTyped(sub, func(ctx context.Context, evt *events.GapAnalysisCompleted) error {
		logger.InfoContext(ctx, "event.received",
			"type", "GapAnalysisCompleted",
			"analysis_id", evt.AnalysisID(),
			"gaps_found", evt.GapsFound(),
			"gaps_resolved", evt.GapsResolved(),
		)
		return nil
	}))

	errs = append(errs, eventbus.SubscribeTyped(sub, func(ctx context.Context, evt *events.ConfigsGenerated) error {
		logger.InfoContext(ctx, "event.received",
			"type", "ConfigsGenerated",
			"tool_count", len(evt.ToolNames()),
		)
		return nil
	}))

	// --- Discovery ---

	errs = append(errs, eventbus.SubscribeTyped(sub, func(ctx context.Context, evt *discoverydomain.DiscoveryCompletedEvent) error {
		logger.InfoContext(ctx, "event.received",
			"type", "DiscoveryCompletedEvent",
			"session_id", evt.SessionID(),
			"persona", string(evt.Persona()),
		)
		return nil
	}))

	// --- Bootstrap ---

	errs = append(errs, eventbus.SubscribeTyped(sub, func(ctx context.Context, evt *bootstrapdomain.BootstrapCompletedEvent) error {
		logger.InfoContext(ctx, "event.received",
			"type", "BootstrapCompletedEvent",
			"session_id", evt.SessionID(),
			"project_dir", evt.ProjectDir(),
		)
		return nil
	}))

	// --- Ticket ---

	errs = append(errs, eventbus.SubscribeTyped(sub, func(ctx context.Context, evt *ticketdomain.TicketPlanApproved) error {
		logger.InfoContext(ctx, "event.received",
			"type", "TicketPlanApproved",
			"plan_id", evt.PlanID(),
			"approved", len(evt.ApprovedTicketIDs()),
			"dismissed", len(evt.DismissedTicketIDs()),
		)
		return nil
	}))

	errs = append(errs, eventbus.SubscribeTyped(sub, func(ctx context.Context, evt *ticketdomain.TicketFlagged) error {
		logger.InfoContext(ctx, "event.received",
			"type", "TicketFlagged",
			"review_id", evt.ReviewID(),
			"ticket_id", evt.TicketID(),
		)
		return nil
	}))

	errs = append(errs, eventbus.SubscribeTyped(sub, func(ctx context.Context, evt *ticketdomain.FlagCleared) error {
		logger.InfoContext(ctx, "event.received",
			"type", "FlagCleared",
			"review_id", evt.ReviewID(),
			"ticket_id", evt.TicketID(),
		)
		return nil
	}))

	// --- Fitness ---

	errs = append(errs, eventbus.SubscribeTyped(sub, func(ctx context.Context, evt *fitnessdomain.FitnessTestsGenerated) error {
		logger.InfoContext(ctx, "event.received",
			"type", "FitnessTestsGenerated",
			"suite_id", evt.SuiteID(),
			"contracts", len(evt.Contracts()),
			"arch_rules", len(evt.ArchRules()),
		)
		return nil
	}))

	// --- ToolTranslation ---

	errs = append(errs, eventbus.SubscribeTyped(sub, func(ctx context.Context, evt *ttdomain.ConfigsGeneratedEvent) error {
		logger.InfoContext(ctx, "event.received",
			"type", "ConfigsGeneratedEvent",
			"tool_count", len(evt.ToolNames()),
		)
		return nil
	}))

	// ===========================================================================
	// Tier 2 — Readiness tracking (updates SessionTracker)
	// ===========================================================================

	// DiscoveryCompleted → artifact_generation is ready
	errs = append(errs, eventbus.SubscribeTyped(sub, func(_ context.Context, evt *discoverydomain.DiscoveryCompletedEvent) error {
		tracker.MarkReady(evt.SessionID(), shareddomain.StepArtifactGeneration)
		return nil
	}))

	// DomainModelGenerated → fitness, tickets, configs are ready
	errs = append(errs, eventbus.SubscribeTyped(sub, func(_ context.Context, evt *events.DomainModelGenerated) error {
		tracker.MarkReady(evt.ModelID(),
			shareddomain.StepFitness,
			shareddomain.StepTickets,
			shareddomain.StepConfigs,
		)
		return nil
	}))

	// TicketPlanApproved → ripple_review is ready
	errs = append(errs, eventbus.SubscribeTyped(sub, func(_ context.Context, evt *ticketdomain.TicketPlanApproved) error {
		tracker.MarkReady(evt.PlanID(), shareddomain.StepRippleReview)
		return nil
	}))

	// FitnessTestsGenerated → fitness is completed
	errs = append(errs, eventbus.SubscribeTyped(sub, func(_ context.Context, evt *fitnessdomain.FitnessTestsGenerated) error {
		tracker.MarkCompleted(evt.SuiteID(), shareddomain.StepFitness)
		return nil
	}))

	// ConfigsGenerated (shared) → configs is completed
	errs = append(errs, eventbus.SubscribeTyped(sub, func(_ context.Context, evt *events.ConfigsGenerated) error {
		// ConfigsGenerated doesn't have a session ID, so we use tool names as a proxy
		// In practice, this event is correlated with a session via the caller
		// For now, this is a no-op until we add session context to the event
		_ = evt
		return nil
	}))

	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("wiring event subscribers: %w", err)
	}
	return sub, nil
}
