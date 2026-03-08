package composition

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	bootstrapdomain "github.com/alty-cli/alty/internal/bootstrap/domain"
	discoverydomain "github.com/alty-cli/alty/internal/discovery/domain"
	fitnessdomain "github.com/alty-cli/alty/internal/fitness/domain"
	"github.com/alty-cli/alty/internal/shared/domain/events"
	"github.com/alty-cli/alty/internal/shared/infrastructure/eventbus"
	ticketdomain "github.com/alty-cli/alty/internal/ticket/domain"
	ttdomain "github.com/alty-cli/alty/internal/tooltranslation/domain"
)

// wireEventSubscribers creates a Subscriber with observability logging handlers
// for all domain events. Each handler logs the event type and key scalar fields
// using structured logging (Tier 1 — observability only).
func wireEventSubscribers(bus *eventbus.Bus, logger *slog.Logger) (*eventbus.Subscriber, error) {
	sub := eventbus.NewSubscriber(bus)
	var errs []error

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

	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("wiring event subscribers: %w", err)
	}
	return sub, nil
}
