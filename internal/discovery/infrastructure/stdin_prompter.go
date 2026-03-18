// Package infrastructure provides adapters for the Discovery bounded context.
package infrastructure

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/alto-cli/alto/internal/discovery/application"
)

// Compile-time interface satisfaction check.
var _ application.Prompter = (*StdinPrompter)(nil)

// StdinPrompter implements Prompter using plain stdin/stdout for accessibility and CI.
type StdinPrompter struct {
	scanner *bufio.Scanner
	writer  io.Writer
}

// NewStdinPrompter creates a new StdinPrompter with the given reader and writer.
func NewStdinPrompter(r io.Reader, w io.Writer) *StdinPrompter {
	return &StdinPrompter{
		scanner: bufio.NewScanner(r),
		writer:  w,
	}
}

// scanOrCancel reads the next line, returning context.Canceled on EOF
// or wrapping any scanner error.
func (p *StdinPrompter) scanOrCancel() (string, error) {
	if !p.scanner.Scan() {
		if err := p.scanner.Err(); err != nil {
			return "", fmt.Errorf("reading input: %w", err)
		}
		return "", context.Canceled // EOF
	}
	return strings.TrimSpace(p.scanner.Text()), nil
}

// SelectPersona displays numbered persona choices and returns the selected choice ("1"-"4").
func (p *StdinPrompter) SelectPersona(_ context.Context) (string, error) {
	_, _ = fmt.Fprintln(p.writer, "Which best describes you?")
	_, _ = fmt.Fprintln(p.writer, "1. Developer (technical background)")
	_, _ = fmt.Fprintln(p.writer, "2. Product Owner (defines what to build)")
	_, _ = fmt.Fprintln(p.writer, "3. Domain Expert (business knowledge)")
	_, _ = fmt.Fprintln(p.writer, "4. Mixed / Other")
	_, _ = fmt.Fprint(p.writer, "Enter choice (1-4): ")

	return p.scanOrCancel()
}

// AskQuestion displays a question and returns the user's answer.
// Returns empty string if the user wants to skip (presses Enter with no input).
func (p *StdinPrompter) AskQuestion(_ context.Context, question string) (string, error) {
	_, _ = fmt.Fprintln(p.writer, question)
	_, _ = fmt.Fprintln(p.writer, "(Press Enter with empty input to skip)")
	_, _ = fmt.Fprint(p.writer, "> ")

	return p.scanOrCancel()
}

// AskSkipReason prompts for a reason when skipping a question.
func (p *StdinPrompter) AskSkipReason(_ context.Context) (string, error) {
	_, _ = fmt.Fprint(p.writer, "Reason for skipping? ")

	return p.scanOrCancel()
}

// ConfirmPlayback displays a summary and asks for confirmation.
// Returns true if confirmed (y/yes), false otherwise.
func (p *StdinPrompter) ConfirmPlayback(_ context.Context, summary string) (bool, error) {
	_, _ = fmt.Fprintln(p.writer, "Review your answers:")
	_, _ = fmt.Fprintln(p.writer, summary)
	_, _ = fmt.Fprint(p.writer, "Continue? (y/n): ")

	answer, err := p.scanOrCancel()
	if err != nil {
		return false, err
	}
	answer = strings.ToLower(answer)
	return answer == "y" || answer == "yes", nil
}
