// Package infrastructure provides adapters for the Discovery bounded context.
package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"charm.land/huh/v2"

	"github.com/alty-cli/alty/internal/discovery/application"
)

// Compile-time interface satisfaction check.
var _ application.Prompter = (*HuhPrompter)(nil)

// HuhPrompter implements Prompter using charmbracelet/huh v2 for interactive TUI prompts.
type HuhPrompter struct{}

// NewHuhPrompter creates a new HuhPrompter.
func NewHuhPrompter() *HuhPrompter {
	return &HuhPrompter{}
}

// personaOptions maps display text to domain choice values ("1"-"4").
var personaOptions = []huh.Option[string]{
	huh.NewOption("Developer (technical background)", "1"),
	huh.NewOption("Product Owner (defines what to build)", "2"),
	huh.NewOption("Domain Expert (business knowledge)", "3"),
	huh.NewOption("Mixed / Other", "4"),
}

// SelectPersona displays persona choices and returns the selected choice ("1"-"4").
func (p *HuhPrompter) SelectPersona(ctx context.Context) (string, error) {
	var choice string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Which best describes you?").
				Options(personaOptions...).
				Value(&choice),
		),
	)

	if err := form.RunWithContext(ctx); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", context.Canceled
		}
		return "", fmt.Errorf("running persona form: %w", err)
	}

	return choice, nil
}

// AskQuestion displays a question and returns the user's answer.
// Returns empty string if the user wants to skip (presses Enter with no input).
func (p *HuhPrompter) AskQuestion(ctx context.Context, question string) (string, error) {
	var answer string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewText().
				Title(question).
				Description("Press Enter with empty input to skip").
				Lines(6).
				Value(&answer),
		),
	)

	if err := form.RunWithContext(ctx); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", context.Canceled
		}
		return "", fmt.Errorf("running question form: %w", err)
	}

	return answer, nil
}

// AskSkipReason prompts for a reason when skipping a question.
func (p *HuhPrompter) AskSkipReason(ctx context.Context) (string, error) {
	var reason string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Reason for skipping?").
				Placeholder("e.g., not relevant to my project").
				Value(&reason),
		),
	)

	if err := form.RunWithContext(ctx); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", context.Canceled
		}
		return "", fmt.Errorf("running skip reason form: %w", err)
	}

	return reason, nil
}
