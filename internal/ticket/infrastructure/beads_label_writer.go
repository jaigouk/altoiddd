package infrastructure

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

const labelWriterTimeout = 5 * time.Second

// BeadsLabelWriter manages labels on beads tickets via the bd CLI.
type BeadsLabelWriter struct {
	timeout time.Duration
}

// NewBeadsLabelWriter creates a BeadsLabelWriter with default settings.
func NewBeadsLabelWriter() *BeadsLabelWriter {
	return &BeadsLabelWriter{
		timeout: labelWriterTimeout,
	}
}

// Timeout returns the configured timeout for label operations.
func (w *BeadsLabelWriter) Timeout() time.Duration {
	return w.timeout
}

// AddLabel adds a label to a ticket.
func (w *BeadsLabelWriter) AddLabel(ctx context.Context, ticketID, label string) error {
	if ticketID == "" {
		return fmt.Errorf("ticket ID cannot be empty")
	}
	if label == "" {
		return fmt.Errorf("label cannot be empty")
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bd", "label", "add", ticketID, label)
	if _, err := cmd.Output(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("label add timed out after %v", w.timeout)
		}
		return fmt.Errorf("adding label: %w", err)
	}

	return nil
}

// RemoveLabel removes a label from a ticket.
func (w *BeadsLabelWriter) RemoveLabel(ctx context.Context, ticketID, label string) error {
	if ticketID == "" {
		return fmt.Errorf("ticket ID cannot be empty")
	}
	if label == "" {
		return fmt.Errorf("label cannot be empty")
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bd", "label", "remove", ticketID, label)
	if _, err := cmd.Output(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("label remove timed out after %v", w.timeout)
		}
		return fmt.Errorf("removing label: %w", err)
	}

	return nil
}
