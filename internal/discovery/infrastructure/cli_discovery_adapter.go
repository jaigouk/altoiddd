package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alto-cli/alto/internal/discovery/application"
	"github.com/alto-cli/alto/internal/discovery/domain"
)

// CLIDiscoveryAdapter orchestrates the CLI-based storytelling discovery flow.
type CLIDiscoveryAdapter struct {
	handler             *application.DiscoveryHandler
	storytellingHandler *application.StorytellingHandler
	prompter            application.StorytellingPrompter
	projectDir          string
}

// NewCLIDiscoveryAdapter creates a new CLIDiscoveryAdapter.
func NewCLIDiscoveryAdapter(
	handler *application.DiscoveryHandler,
	storytellingHandler *application.StorytellingHandler,
	prompter application.StorytellingPrompter,
	projectDir string,
) *CLIDiscoveryAdapter {
	return &CLIDiscoveryAdapter{
		handler:             handler,
		storytellingHandler: storytellingHandler,
		prompter:            prompter,
		projectDir:          projectDir,
	}
}

// personaChoices are the persona options presented to the user.
var personaChoices = []application.Choice{
	{Key: "1", Label: "Developer", Description: "I write code"},
	{Key: "2", Label: "Product Owner", Description: "I define requirements"},
	{Key: "3", Label: "Domain Expert", Description: "I know the business"},
	{Key: "4", Label: "Mixed Team", Description: "Multiple roles"},
}

// Run executes the storytelling discovery flow:
// read README, start session, select mode, select persona, run story loop, complete.
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

	// Step 3: Select mode (MUST be before DetectPersona — SetMode requires StatusCreated)
	mode, err := a.prompter.SelectMode(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return context.Canceled
		}
		return fmt.Errorf("selecting mode: %w", err)
	}

	if setModeErr := session.SetMode(mode); setModeErr != nil {
		return fmt.Errorf("setting mode: %w", setModeErr)
	}

	// Step 4: Select persona
	choice, err := a.prompter.AskChoice(ctx, "Who are you?", personaChoices, "1")
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return context.Canceled
		}
		return fmt.Errorf("selecting persona: %w", err)
	}

	session, err = a.handler.DetectPersona(session.SessionID(), choice)
	if err != nil {
		return fmt.Errorf("detecting persona: %w", err)
	}

	// Step 5: Create storytelling flow
	flow, err := domain.NewStorytellingFlow(mode)
	if err != nil {
		return fmt.Errorf("creating storytelling flow: %w", err)
	}

	// Step 6: Story loop
	for storyIndex := 1; ; storyIndex++ {
		_, _, err := a.storytellingHandler.RunStory(ctx, session, storyIndex, flow)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return context.Canceled
			}
			return fmt.Errorf("running story %d: %w", storyIndex, err)
		}

		// Check if we have enough stories
		if flow.CheckStoryCompleteness(session.StoryCount()) == nil {
			break
		}

		// Ask if user wants to continue
		continueChoices := []application.Choice{
			{Key: "yes", Label: "Yes", Description: "Tell another domain story"},
			{Key: "no", Label: "No", Description: "Finish discovery"},
		}

		continueChoice, err := a.prompter.AskChoice(ctx, "Would you like to tell another story?", continueChoices, "yes")
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return context.Canceled
			}
			return fmt.Errorf("asking continue: %w", err)
		}

		if continueChoice == "no" {
			break
		}
	}

	// Step 6.5: Confirm boundaries (required by Invariant 9 for storytelling sessions)
	if err := session.ConfirmBoundaries(nil); err != nil {
		return fmt.Errorf("confirming boundaries: %w", err)
	}

	// Step 7: Complete session
	if _, err := a.handler.Complete(session.SessionID()); err != nil { //nolint:contextcheck // Discovery interface deliberately omits context
		return fmt.Errorf("completing session: %w", err)
	}

	return nil
}
