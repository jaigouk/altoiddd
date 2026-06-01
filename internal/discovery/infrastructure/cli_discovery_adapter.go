package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	// Artifact pipeline (optional — nil when not wired).
	artifactGen      *application.ArtifactGenerationHandler
	glossaryExport   *application.GlossaryExportHandler
	discoveryReport  *application.DiscoveryReportHandler
	contextMapReader application.ContextMapReader
	storyReader      application.StoryReader
}

// CLIDiscoveryAdapterOption configures optional behavior on CLIDiscoveryAdapter.
type CLIDiscoveryAdapterOption func(*CLIDiscoveryAdapter)

// WithArtifactPipeline wires the artifact generation pipeline that runs after
// storytelling completes. When not provided, the adapter completes without
// producing artifacts (backward-compatible).
func WithArtifactPipeline(
	artifactGen *application.ArtifactGenerationHandler,
	glossaryExport *application.GlossaryExportHandler,
	discoveryReport *application.DiscoveryReportHandler,
	contextMapReader application.ContextMapReader,
	storyReader application.StoryReader,
) CLIDiscoveryAdapterOption {
	return func(a *CLIDiscoveryAdapter) {
		a.artifactGen = artifactGen
		a.glossaryExport = glossaryExport
		a.discoveryReport = discoveryReport
		a.contextMapReader = contextMapReader
		a.storyReader = storyReader
	}
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
	opts ...CLIDiscoveryAdapterOption,
) *CLIDiscoveryAdapter {
	a := &CLIDiscoveryAdapter{
		handler:                  handler,
		storytellingHandler:      storytellingHandler,
		boundaryDetectionHandler: boundaryDetectionHandler,
		boundaryPrompter:         boundaryPrompter,
		contextMapWriter:         contextMapWriter,
		prompter:                 prompter,
		projectDir:               projectDir,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
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
		remaining := flow.RequiredStoryCount() - session.StoryCount()
		basePrompt := fmt.Sprintf("Story %d of %d completed. Tell another?", session.StoryCount(), flow.RequiredStoryCount())

		continueChoices := []application.Choice{
			{Key: "yes", Label: "Yes", Description: "Tell another domain story"},
			{Key: "no", Label: "No", Description: "Finish discovery"},
		}

		if remaining > 0 {
			continueChoices[0].Description = fmt.Sprintf("Tell another domain story (%d more required)", remaining)
			continueChoices[1].Description = fmt.Sprintf("Finish discovery (need %d more to complete)", remaining)
		}

		rctx := application.NewRegroundingContext(session)
		groundedPrompt := application.BuildRegroundingPrompt(rctx, basePrompt)

		continueChoice, err := a.prompter.AskChoice(ctx, groundedPrompt, continueChoices, "yes")
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
	completedSession, completeErr := a.handler.Complete(session.SessionID()) //nolint:contextcheck // Discovery interface deliberately omits context
	if completeErr != nil {
		return fmt.Errorf("completing session: %w", completeErr)
	}

	// Step 8: Artifact pipeline
	if pipelineErr := a.runArtifactPipeline(ctx, completedSession); pipelineErr != nil {
		return fmt.Errorf("artifact pipeline: %w", pipelineErr)
	}

	return nil
}

// runArtifactPipeline runs glossary export, artifact generation, and discovery report
// after storytelling completes. No-op when pipeline is not wired.
func (a *CLIDiscoveryAdapter) runArtifactPipeline(ctx context.Context, session *domain.DiscoverySession) error {
	if a.artifactGen == nil {
		return nil // Pipeline not wired — backward-compatible.
	}

	storyRefs := session.StoryRefs()
	if len(storyRefs) == 0 {
		return nil // No stories to process.
	}

	// Resolve paths.
	glossaryPath := filepath.Join(a.projectDir, "alto-scaffold", "glossary.yaml")
	contextMapPath := filepath.Join(a.projectDir, "alto-scaffold", "context-map.yaml")
	docsDir := filepath.Join(a.projectDir, "docs")

	// Resolve project name.
	projectName, err := a.resolveProjectName()
	if err != nil {
		return fmt.Errorf("resolving project name for artifacts: %w", err)
	}
	_ = projectName // Used as fallback; GenerateFromStories derives from projectDir.

	// Read context map if available.
	var contextMap *domain.ContextMap
	if len(session.ConfirmedSketches()) > 0 && a.contextMapReader != nil {
		contextMap, err = a.contextMapReader.Read(ctx, contextMapPath)
		if err != nil {
			return fmt.Errorf("reading context map: %w", err)
		}
	}

	// Step 1: Glossary export.
	if a.glossaryExport != nil {
		if _, exportErr := a.glossaryExport.Export(ctx, storyRefs, contextMap, glossaryPath); exportErr != nil {
			return fmt.Errorf("exporting glossary: %w", exportErr)
		}
	}

	// Step 2: Artifact generation (PRD, DDD, ARCHITECTURE).
	if a.storyReader != nil {
		if genErr := a.artifactGen.GenerateFromStories(ctx, a.storyReader, storyRefs, contextMap, docsDir, a.projectDir); genErr != nil {
			return fmt.Errorf("generating artifacts: %w", genErr)
		}
	}

	// Step 3: Discovery report (only for multi-context — needs context map).
	if len(session.ConfirmedSketches()) > 0 && a.discoveryReport != nil {
		if reportErr := a.discoveryReport.GenerateReport(ctx, storyRefs, glossaryPath, contextMapPath, a.projectDir); reportErr != nil {
			return fmt.Errorf("generating discovery report: %w", reportErr)
		}
	}

	return nil
}

// runBoundaryDetection detects, displays, and confirms boundary sketches,
// then persists the context map to alto-scaffold/context-map.yaml.
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

	contextMapPath := filepath.Join(a.projectDir, "alto-scaffold", "context-map.yaml")

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

			remaining := flow.RequiredStoryCount() - session.StoryCount()
			basePrompt := fmt.Sprintf("Story %d of %d completed. Tell another?", session.StoryCount(), flow.RequiredStoryCount())

			continueChoices := []application.Choice{
				{Key: "yes", Label: "Yes", Description: "Tell another domain story"},
				{Key: "no", Label: "No", Description: "Finish discovery"},
			}

			if remaining > 0 {
				continueChoices[0].Description = fmt.Sprintf("Tell another domain story (%d more required)", remaining)
				continueChoices[1].Description = fmt.Sprintf("Finish discovery (need %d more to complete)", remaining)
			}

			rctx := application.NewRegroundingContext(session)
			groundedPrompt := application.BuildRegroundingPrompt(rctx, basePrompt)

			continueChoice, err := a.prompter.AskChoice(ctx, groundedPrompt, continueChoices, "yes")
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
	completedSession, completeErr := a.handler.Complete(session.SessionID()) //nolint:contextcheck // Discovery interface deliberately omits context
	if completeErr != nil {
		return fmt.Errorf("completing session: %w", completeErr)
	}

	// 7. Artifact pipeline
	if pipelineErr := a.runArtifactPipeline(ctx, completedSession); pipelineErr != nil {
		return fmt.Errorf("artifact pipeline: %w", pipelineErr)
	}

	return nil
}

// resolveProjectName extracts the project name from the first # heading in
// README.md or README, falling back to the directory base name.
func (a *CLIDiscoveryAdapter) resolveProjectName() (string, error) {
	dir := a.projectDir
	if dir == "." {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("getting working directory: %w", err)
		}

		dir = wd
	}

	for _, candidate := range []string{"README.md", "README"} {
		content, err := os.ReadFile(filepath.Join(dir, candidate))
		if err != nil {
			continue
		}

		for _, line := range strings.Split(string(content), "\n") {
			after, ok := strings.CutPrefix(line, "# ")
			if !ok {
				continue
			}

			if name := strings.TrimSpace(after); name != "" {
				return stripSubtitle(name), nil
			}
		}
	}

	return filepath.Base(dir), nil
}

// stripSubtitle removes a tagline/subtitle after a separator in a heading.
// Separators checked in order: " — ", " - ", ": ", " | ". Earliest match wins.
func stripSubtitle(heading string) string {
	for _, sep := range []string{" — ", " - ", ": ", " | "} {
		if idx := strings.Index(heading, sep); idx > 0 {
			return heading[:idx]
		}
	}

	return heading
}
