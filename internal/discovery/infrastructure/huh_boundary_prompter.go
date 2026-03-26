package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"charm.land/huh/v2"

	"github.com/alto-cli/alto/internal/discovery/application"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
)

// Compile-time interface satisfaction check.
var _ application.BoundaryPrompter = (*HuhBoundaryPrompter)(nil)

// HuhBoundaryPrompter implements BoundaryPrompter using charmbracelet/huh v2
// for interactive TUI prompts during the boundary detection review step.
type HuhBoundaryPrompter struct{}

// NewHuhBoundaryPrompter creates a new HuhBoundaryPrompter.
func NewHuhBoundaryPrompter() *HuhBoundaryPrompter {
	return &HuhBoundaryPrompter{}
}

// DisplayBoundaryProposals presents each BoundedContextSketch for per-sketch accept/reject.
// Returns the names of accepted sketches.
func (p *HuhBoundaryPrompter) DisplayBoundaryProposals(
	ctx context.Context,
	proposals []discoverydomain.BoundedContextSketch,
) ([]string, error) {
	if ctx.Err() != nil {
		return nil, context.Canceled
	}

	if len(proposals) == 0 {
		return []string{}, nil
	}

	var accepted []string

	for _, sketch := range proposals {
		description := formatSketchDescription(sketch)

		var confirmed bool

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(fmt.Sprintf("Accept bounded context: %s?", sketch.Name())).
					Description(description).
					Affirmative("Accept").
					Negative("Reject").
					Value(&confirmed),
			),
		)

		if err := form.RunWithContext(ctx); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil, context.Canceled
			}

			return nil, fmt.Errorf("displaying boundary proposal %q: %w", sketch.Name(), err)
		}

		if confirmed {
			accepted = append(accepted, sketch.Name())
		}
	}

	return accepted, nil
}

// AskMissingContext asks if the user sees a missing bounded context area.
// Returns non-empty string if yes, empty if the user has nothing to add.
func (p *HuhBoundaryPrompter) AskMissingContext(ctx context.Context) (string, error) {
	if ctx.Err() != nil {
		return "", context.Canceled
	}

	var missing string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Is there a major area I missed?").
				Description("Enter a name for the missing bounded context, or leave empty to continue.").
				Value(&missing),
		),
	)

	if err := form.RunWithContext(ctx); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", context.Canceled
		}

		return "", fmt.Errorf("asking missing context: %w", err)
	}

	return missing, nil
}

// formatSketchDescription builds a human-readable description of a sketch for display.
func formatSketchDescription(sketch discoverydomain.BoundedContextSketch) string {
	desc := fmt.Sprintf("Confidence: %.0f%% (%s)\nTrust: %s",
		sketch.Confidence()*100, sketch.ConfidenceLevel(), sketch.Trust())

	if actors := sketch.Actors(); len(actors) > 0 {
		desc += fmt.Sprintf("\nActors: %s", joinMax(actors, 5))
	}

	if signals := sketch.Signals(); len(signals) > 0 {
		desc += fmt.Sprintf("\nSignals: %d boundary signal(s)", len(signals))
	}

	if workObjects := sketch.WorkObjects(); len(workObjects) > 0 {
		desc += fmt.Sprintf("\nWork Objects: %s", joinMax(workObjects, 5))
	}

	return desc
}

// joinMax joins up to max items with ", " and appends "..." if truncated.
func joinMax(items []string, max int) string {
	if len(items) <= max {
		return strings.Join(items, ", ")
	}

	return strings.Join(items[:max], ", ") + ", ..."
}
