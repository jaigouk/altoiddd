// Package infrastructure provides adapters for the Ticket bounded context.
package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/alty-cli/alty/internal/ticket/application"
	ticketdomain "github.com/alty-cli/alty/internal/ticket/domain"
)

// Compile-time interface satisfaction check.
var _ application.BeadsWriter = (*BeadsCLIWriter)(nil)

// issueIDRegex matches "Created issue: <id> — <title>" output from bd create.
var issueIDRegex = regexp.MustCompile(`Created issue:\s*(\S+)\s*—`)

// BeadsCLIWriter implements BeadsWriter by shelling out to the bd CLI.
type BeadsCLIWriter struct {
	projectDir string
}

// NewBeadsCLIWriter creates a new BeadsCLIWriter for the given project directory.
func NewBeadsCLIWriter(projectDir string) *BeadsCLIWriter {
	return &BeadsCLIWriter{projectDir: projectDir}
}

// WriteEpic creates an epic in beads and returns the assigned ID.
func (w *BeadsCLIWriter) WriteEpic(ctx context.Context, epic ticketdomain.GeneratedEpic) (string, error) {
	args := []string{
		"create",
		"--title", epic.Title(),
		"--description", epic.Description(),
		"--type", "epic",
	}

	return w.runBdCreate(ctx, args)
}

// WriteTicket creates a ticket in beads and returns the assigned ID.
// Uses the ticket's TemplateType() to determine whether to create a task or spike.
func (w *BeadsCLIWriter) WriteTicket(ctx context.Context, ticket ticketdomain.GeneratedTicket) (string, error) {
	ticketType := "task"
	if ticket.TemplateType() == ticketdomain.TemplateSpike {
		ticketType = "spike"
	}

	args := []string{
		"create",
		"--title", ticket.Title(),
		"--description", ticket.Description(),
		"--type", ticketType,
	}

	// Add parent if this ticket belongs to an epic
	if ticket.EpicID() != "" {
		args = append(args, "--parent", ticket.EpicID())
	}

	return w.runBdCreate(ctx, args)
}

// SetDependency sets a dependency between two tickets.
func (w *BeadsCLIWriter) SetDependency(ctx context.Context, ticketID, dependsOnID string) error {
	cmd := exec.CommandContext(ctx, "bd", "dep", "add", ticketID, dependsOnID)
	cmd.Dir = w.projectDir

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("setting dependency %s -> %s: %w", ticketID, dependsOnID, err)
	}

	return nil
}

// runBdCreate executes bd create with the given args and returns the issue ID.
func (w *BeadsCLIWriter) runBdCreate(ctx context.Context, args []string) (string, error) {
	cmd := exec.CommandContext(ctx, "bd", args...)
	cmd.Dir = w.projectDir

	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", fmt.Errorf("bd create failed: %s: %w", string(exitErr.Stderr), err)
		}
		return "", fmt.Errorf("bd create failed: %w", err)
	}

	return ParseIssueID(string(output))
}

// ParseIssueID extracts the issue ID from bd create output.
// Output format: "Created issue: <id> — <title>"
func ParseIssueID(output string) (string, error) {
	matches := issueIDRegex.FindStringSubmatch(strings.TrimSpace(output))
	if len(matches) < 2 {
		return "", fmt.Errorf("could not parse issue ID from output: %q", output)
	}
	return matches[1], nil
}
