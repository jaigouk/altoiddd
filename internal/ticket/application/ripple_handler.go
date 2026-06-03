package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	ticketdomain "github.com/alto-cli/alto/internal/ticket/domain"
)

// RippleReviewResult is the value-object summary returned by
// RippleHandler.Handle. FlaggedCount is the number of distinct tickets
// that received a review_needed label + ripple comment. EventCount is
// the number of domain events emitted by the underlying RippleReview
// aggregate (one TicketFlagged per flag + one RippleReviewCreated on
// Complete()).
type RippleReviewResult struct {
	FlaggedCount int
	EventCount   int
}

// RippleHandler orchestrates the post-close ripple-review flow:
//  1. Resolve the context summary (override > close_reason).
//  2. Walk the dependency graph (siblings, dependents, related).
//  3. Deduplicate and filter to open tickets only.
//  4. Construct a RippleReview aggregate and FlagTicket() each.
//  5. Apply the review_needed label + ripple comment via adapters.
//  6. Complete() the aggregate to emit RippleReviewCreated.
//
// The handler depends on three narrow ports (BeadsGraphReader,
// BeadsLabelWriter, BeadsCommentWriter) and never reaches into the
// infrastructure layer directly.
type RippleHandler struct {
	graphReader   BeadsGraphReader
	labelWriter   BeadsLabelWriter
	commentWriter BeadsCommentWriter
	clock         func() time.Time
	idGen         func() string
}

// NewRippleHandler constructs a RippleHandler with injected ports.
func NewRippleHandler(
	graphReader BeadsGraphReader,
	labelWriter BeadsLabelWriter,
	commentWriter BeadsCommentWriter,
) *RippleHandler {
	return &RippleHandler{
		graphReader:   graphReader,
		labelWriter:   labelWriter,
		commentWriter: commentWriter,
		clock:         time.Now,
		idGen:         uuid.NewString,
	}
}

// WithClock injects a deterministic clock for testing.
func (h *RippleHandler) WithClock(clock func() time.Time) *RippleHandler {
	h.clock = clock
	return h
}

// WithIDGenerator injects a deterministic ID generator for testing.
func (h *RippleHandler) WithIDGenerator(idGen func() string) *RippleHandler {
	h.idGen = idGen
	return h
}

// Handle runs the ripple-review flow for closedTicketID. When
// contextOverride is non-nil its value is used as the ContextDiff
// summary; otherwise the closed ticket's close_reason is used. An
// empty summary (both override and close_reason missing) is a hard
// error — the DDD invariant on ContextDiff forbids whitespace-only
// summaries.
func (h *RippleHandler) Handle(
	ctx context.Context,
	closedTicketID string,
	contextOverride *string,
) (RippleReviewResult, error) {
	if closedTicketID == "" {
		return RippleReviewResult{}, fmt.Errorf("closed ticket ID cannot be empty")
	}

	summary, err := h.resolveSummary(ctx, closedTicketID, contextOverride)
	if err != nil {
		return RippleReviewResult{}, err
	}

	now := h.clock().UTC().Format(time.RFC3339)
	diff, err := ticketdomain.NewContextDiff(summary, closedTicketID, now)
	if err != nil {
		return RippleReviewResult{}, fmt.Errorf("building context diff: %w", err)
	}

	toFlag, err := h.collectFlagTargets(ctx, closedTicketID)
	if err != nil {
		return RippleReviewResult{}, err
	}

	review := ticketdomain.NewRippleReview(h.idGen(), closedTicketID, diff)
	comment := review.BuildRippleComment()

	for _, ticketID := range toFlag {
		if err := review.FlagTicket(ticketID, true, now); err != nil {
			return RippleReviewResult{}, fmt.Errorf("flagging %s: %w", ticketID, err)
		}
		if err := h.labelWriter.AddLabel(ctx, ticketID, "review_needed"); err != nil {
			return RippleReviewResult{}, fmt.Errorf("adding review_needed label to %s: %w", ticketID, err)
		}
		if err := h.commentWriter.AddComment(ctx, ticketID, comment); err != nil {
			return RippleReviewResult{}, fmt.Errorf("adding ripple comment to %s: %w", ticketID, err)
		}
	}

	if err := review.Complete(); err != nil {
		return RippleReviewResult{}, fmt.Errorf("completing review: %w", err)
	}

	return RippleReviewResult{
		FlaggedCount: len(review.FlaggedTickets()),
		EventCount:   len(review.Events()),
	}, nil
}

// resolveSummary returns the override when non-nil and non-empty,
// otherwise falls back to the close_reason. Empty summaries are
// rejected so the downstream NewContextDiff invariant is preserved.
func (h *RippleHandler) resolveSummary(
	ctx context.Context,
	closedTicketID string,
	override *string,
) (string, error) {
	if override != nil && *override != "" {
		return *override, nil
	}
	reason, err := h.graphReader.ReadCloseContext(ctx, closedTicketID)
	if err != nil {
		return "", fmt.Errorf("reading close context: %w", err)
	}
	if reason == "" || reason == "Closed" {
		return "", fmt.Errorf("non-empty context summary required for ripple review; provide a summary or close with --reason")
	}
	return reason, nil
}

// collectFlagTargets walks the dependency graph and returns the
// deduplicated list of OPEN ticket IDs that should be flagged. Order is
// stable: siblings, then dependents, then related, preserving first
// occurrence — the handler owns this rule, not the adapter.
func (h *RippleHandler) collectFlagTargets(
	ctx context.Context,
	closedTicketID string,
) ([]string, error) {
	parentID, perr := h.graphReader.ReadParent(ctx, closedTicketID)
	if perr != nil {
		return nil, fmt.Errorf("reading parent: %w", perr)
	}

	var sources [][]BeadsGraphTicket

	if parentID != "" {
		siblings, serr := h.graphReader.ReadSiblings(ctx, parentID, closedTicketID)
		if serr != nil {
			return nil, fmt.Errorf("reading siblings: %w", serr)
		}
		sources = append(sources, siblings)
	}

	dependents, derr := h.graphReader.ReadDependents(ctx, closedTicketID)
	if derr != nil {
		return nil, fmt.Errorf("reading dependents: %w", derr)
	}
	sources = append(sources, dependents)

	related, rerr := h.graphReader.ReadRelated(ctx, closedTicketID)
	if rerr != nil {
		return nil, fmt.Errorf("reading related: %w", rerr)
	}
	sources = append(sources, related)

	seen := make(map[string]struct{})
	var ordered []string
	for _, source := range sources {
		for _, t := range source {
			if t.ID == "" || t.ID == closedTicketID {
				continue
			}
			if !t.IsOpen() {
				continue
			}
			if _, dup := seen[t.ID]; dup {
				continue
			}
			seen[t.ID] = struct{}{}
			ordered = append(ordered, t.ID)
		}
	}
	return ordered, nil
}
