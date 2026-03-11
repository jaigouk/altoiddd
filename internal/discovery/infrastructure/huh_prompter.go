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
