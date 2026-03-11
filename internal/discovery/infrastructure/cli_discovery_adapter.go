package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alty-cli/alty/internal/discovery/application"
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
	if _, err := a.handler.DetectPersona(session.SessionID(), choice); err != nil {
		return fmt.Errorf("detecting persona: %w", err)
	}

	// TODO: Question loop (alty-cli-7u7.4)
	return nil
}
