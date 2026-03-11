package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alty-cli/alty/internal/discovery/application"
	"github.com/alty-cli/alty/internal/discovery/domain"
)

// CLIDiscoveryAdapter orchestrates the CLI-based discovery flow.
type CLIDiscoveryAdapter struct {
	handler    *application.DiscoveryHandler
	prompter   application.Prompter
	projectDir string
}

// NewCLIDiscoveryAdapter creates a new CLIDiscoveryAdapter.
func NewCLIDiscoveryAdapter(
	handler *application.DiscoveryHandler,
	prompter application.Prompter,
	projectDir string,
) *CLIDiscoveryAdapter {
	return &CLIDiscoveryAdapter{
		handler:    handler,
		prompter:   prompter,
		projectDir: projectDir,
	}
}

// Run executes the discovery flow: read README, start session, select persona.
func (a *CLIDiscoveryAdapter) Run(ctx context.Context) error {
	// Step 1: Read README
	readmePath := filepath.Join(a.projectDir, "README.md")
	readme, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("reading README.md: %w", err)
	}

	// Step 2: Start session
	session, err := a.handler.StartSession(string(readme))
	if err != nil {
		return fmt.Errorf("starting session: %w", err)
	}

	// Step 3: Persona selection
	choice, err := a.prompter.SelectPersona(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return context.Canceled
		}
		return fmt.Errorf("selecting persona: %w", err)
	}

	// Step 4: Detect persona in domain
	session, err = a.handler.DetectPersona(session.SessionID(), choice)
	if err != nil {
		return fmt.Errorf("detecting persona: %w", err)
	}

	// Step 5: Question loop
	sessionID := session.SessionID()
	register, hasRegister := session.Register()
	if !hasRegister {
		return fmt.Errorf("session has no register after persona detection")
	}

	questions := domain.QuestionCatalog()
	for i, question := range questions {
		// Select text based on register
		var text string
		if register == domain.RegisterTechnical {
			text = question.TechnicalText()
		} else {
			text = question.NonTechnicalText()
		}

		// Display progress and ask question
		fmt.Printf("Q%d/%d\n", i+1, len(questions))
		answer, err := a.prompter.AskQuestion(ctx, text)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return context.Canceled
			}
			return fmt.Errorf("asking question %s: %w", question.ID(), err)
		}

		// Handle answer or skip
		if answer == "" {
			// Skip: ask for reason
			reason, skipErr := a.prompter.AskSkipReason(ctx)
			if skipErr != nil {
				if errors.Is(skipErr, context.Canceled) {
					return context.Canceled
				}
				return fmt.Errorf("asking skip reason: %w", skipErr)
			}
			session, err = a.handler.SkipQuestion(sessionID, question.ID(), reason)
			if err != nil {
				return fmt.Errorf("skipping question %s: %w", question.ID(), err)
			}
		} else {
			// Answer the question
			session, err = a.handler.AnswerQuestion(sessionID, question.ID(), answer)
			if err != nil {
				return fmt.Errorf("answering question %s: %w", question.ID(), err)
			}
		}

		// Check for playback pending (triggered every 3 questions)
		if session.Status() == domain.StatusPlaybackPending {
			// Auto-confirm playback for now (full playback handling in 7u7.5)
			_, err = a.handler.ConfirmPlayback(sessionID, true)
			if err != nil {
				return fmt.Errorf("confirming playback: %w", err)
			}
		}
	}

	return nil
}
