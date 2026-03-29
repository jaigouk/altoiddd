package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alto-cli/alto/internal/discovery/application"
	"github.com/alto-cli/alto/internal/discovery/domain"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// CLIDiscoveryAdapter orchestrates the CLI-based storytelling discovery flow.
type CLIDiscoveryAdapter struct {
	handler                  *application.DiscoveryHandler
	storytellingHandler      *application.StorytellingHandler
	boundaryDetectionHandler *application.BoundaryDetectionHandler
	boundaryPrompter         application.BoundaryPrompter
	contextMapWriter         application.ContextMapWriter
	prompter                 application.StorytellingPrompter
	projectDir               string
}

// NewCLIDiscoveryAdapter creates a new CLIDiscoveryAdapter.
func NewCLIDiscoveryAdapter(
	handler *application.DiscoveryHandler,
	storytellingHandler *application.StorytellingHandler,
	boundaryDetectionHandler *application.BoundaryDetectionHandler,
	boundaryPrompter application.BoundaryPrompter,
	contextMapWriter application.ContextMapWriter,
	prompter application.StorytellingPrompter,
	projectDir string,
) *CLIDiscoveryAdapter {
	return &CLIDiscoveryAdapter{
		handler:                  handler,
		storytellingHandler:      storytellingHandler,
		boundaryDetectionHandler: boundaryDetectionHandler,
		boundaryPrompter:         boundaryPrompter,
		contextMapWriter:         contextMapWriter,
		prompter:                 prompter,
		projectDir:               projectDir,
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

	// Step 6: Story loop — accumulate stories for boundary detection
	var accumulatedStories []*domain.DomainStory

	for storyIndex := 1; ; storyIndex++ {
		story, _, err := a.storytellingHandler.RunStory(ctx, session, storyIndex, flow)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return context.Canceled
			}

			return fmt.Errorf("running story %d: %w", storyIndex, err)
		}

		accumulatedStories = append(accumulatedStories, story)

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

	// Step 6.5: Boundary detection + confirmation (Invariant 9)
	if flow.CanRunBoundaryDetection(session.StoryCount()) {
		if err := a.runBoundaryDetection(ctx, session, accumulatedStories, mode); err != nil {
			return err
		}
	} else {
		// Single-context domain — fewer than minimum stories for boundary detection
		if err := session.ConfirmBoundaries(nil); err != nil {
			return fmt.Errorf("confirming boundaries: %w", err)
		}
	}

	// Step 7: Complete session
	if _, err := a.handler.Complete(session.SessionID()); err != nil { //nolint:contextcheck // Discovery interface deliberately omits context
		return fmt.Errorf("completing session: %w", err)
	}

	return nil
}

// runBoundaryDetection detects, displays, and confirms boundary sketches,
// then persists the context map to .alto/context-map.yaml.
func (a *CLIDiscoveryAdapter) runBoundaryDetection(
	ctx context.Context,
	session *domain.DiscoverySession,
	stories []*domain.DomainStory,
	mode domain.DiscoveryMode,
) error {
	// 1. Detect boundaries
	sketches, err := a.boundaryDetectionHandler.Detect(ctx, stories, mode)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return context.Canceled
		}

		return fmt.Errorf("detecting boundaries: %w", err)
	}

	// 2. Display proposals for user review
	acceptedNames, err := a.boundaryPrompter.DisplayBoundaryProposals(ctx, sketches)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return context.Canceled
		}

		return fmt.Errorf("displaying boundary proposals: %w", err)
	}

	// 3. Ask for missing context
	missingName, err := a.boundaryPrompter.AskMissingContext(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return context.Canceled
		}

		return fmt.Errorf("asking missing context: %w", err)
	}

	// 4. Build confirmed sketch list
	acceptedSet := make(map[string]struct{}, len(acceptedNames))
	for _, name := range acceptedNames {
		acceptedSet[name] = struct{}{}
	}

	var confirmedSketches []domain.BoundedContextSketch

	for _, sketch := range sketches {
		if _, ok := acceptedSet[sketch.Name()]; ok {
			confirmedSketches = append(confirmedSketches, sketch)
		}
	}

	// 5. Add user-stated stub if missing context provided
	if missingName != "" {
		stub, stubErr := domain.NewBoundedContextSketch(
			missingName, vo.SubdomainGeneric, 0.50,
			nil, nil, nil, nil, vo.UserStated,
		)
		if stubErr != nil {
			return fmt.Errorf("creating missing context sketch: %w", stubErr)
		}

		confirmedSketches = append(confirmedSketches, stub)
	}

	// 6. Confirm boundaries on session
	if confirmErr := session.ConfirmBoundaries(confirmedSketches); confirmErr != nil {
		return fmt.Errorf("confirming boundaries: %w", confirmErr)
	}

	// 7. Resolve project name from projectDir
	projectName, err := a.resolveProjectName()
	if err != nil {
		return fmt.Errorf("resolving project name: %w", err)
	}

	// 8. Build and write context map
	cm, err := domain.NewContextMap(projectName, confirmedSketches, nil)
	if err != nil {
		return fmt.Errorf("creating context map: %w", err)
	}

	contextMapPath := filepath.Join(a.projectDir, ".alto", "context-map.yaml")

	if err := a.contextMapWriter.Write(ctx, contextMapPath, cm); err != nil {
		return fmt.Errorf("writing context map: %w", err)
	}

	return nil
}

// Resume resumes a previously interrupted storytelling session from the checkpoint.
func (a *CLIDiscoveryAdapter) Resume(ctx context.Context, session *domain.DiscoverySession) error {
	// 1. Register session in handler so Complete can find it
	a.handler.RegisterSession(session)

	// 2. Compute resume checkpoint
	checkpoint, err := session.ComputeResumeCheckpoint()
	if err != nil {
		return fmt.Errorf("computing resume checkpoint: %w", err)
	}

	// 3. Create storytelling flow from session mode
	flow, err := domain.NewStorytellingFlow(session.Mode())
	if err != nil {
		return fmt.Errorf("creating storytelling flow: %w", err)
	}

	// 4. Story loop from checkpoint — only if stories are still needed
	var accumulatedStories []*domain.DomainStory

	if flow.CheckStoryCompleteness(session.StoryCount()) != nil {
		for storyIndex := checkpoint.StoryIndex; ; storyIndex++ {
			story, _, err := a.storytellingHandler.RunStory(ctx, session, storyIndex, flow)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return context.Canceled
				}

				return fmt.Errorf("running story %d: %w", storyIndex, err)
			}

			accumulatedStories = append(accumulatedStories, story)

			if flow.CheckStoryCompleteness(session.StoryCount()) == nil {
				break
			}

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
	}

	// 5. Boundary detection (only if not already done)
	if !checkpoint.BoundariesDone {
		if flow.CanRunBoundaryDetection(session.StoryCount()) {
			if err := a.runBoundaryDetection(ctx, session, accumulatedStories, session.Mode()); err != nil {
				return err
			}
		} else {
			if err := session.ConfirmBoundaries(nil); err != nil {
				return fmt.Errorf("confirming boundaries: %w", err)
			}
		}
	}

	// 6. Complete session
	if _, err := a.handler.Complete(session.SessionID()); err != nil { //nolint:contextcheck // Discovery interface deliberately omits context
		return fmt.Errorf("completing session: %w", err)
	}

	return nil
}

// resolveProjectName returns the project directory name.
// If projectDir is ".", it resolves to the actual directory name via os.Getwd.
func (a *CLIDiscoveryAdapter) resolveProjectName() (string, error) {
	dir := a.projectDir
	if dir == "." {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("getting working directory: %w", err)
		}

		dir = wd
	}

	return filepath.Base(dir), nil
}
