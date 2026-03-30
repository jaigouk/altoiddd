package commands

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/alto-cli/alto/internal/composition"
	"github.com/alto-cli/alto/internal/discovery/application"
	"github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/discovery/infrastructure"
)

// Deprecation warning constants for legacy discovery flow.
const (
	legacyFlagWarning     = "warning: --legacy uses the deprecated question-based discovery flow; use `alto guide` for the new storytelling flow"
	legacyContinueWarning = "warning: this session uses legacy discovery mode %q; the question-based flow is deprecated"
)

// NewGuideCmd creates the "alto guide" command.
func NewGuideCmd(app *composition.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "guide",
		Short: "Run the Domain Storytelling guided discovery flow",
		Long: `Run the Domain Storytelling guided discovery flow.

alto acts as a domain consultant, moderating a structured conversation:
  1. Detection of installed AI coding tools
  2. Domain Storytelling moderator conversation (actors, activities, work objects
     structured into stories through opening → narration → deepening → closing)
  3. Boundary detection on accumulated stories (bounded context sketches)
  4. Artifact generation from completed stories

Flags:
  --no-tui     Disable TUI prompts, use plain stdin/stdout (accessibility, CI)
  --continue   Resume a session started with --agent
  --agent      Output discovery session as JSONL for AI agent consumption
  --ingest     Ingest answers from JSONL file (or "-" for stdin); requires --agent
  --legacy     Use deprecated question-based flow (prints deprecation warning)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			noTUI, _ := cmd.Flags().GetBool("no-tui")
			continueSession, _ := cmd.Flags().GetBool("continue")
			agentMode, _ := cmd.Flags().GetBool("agent")
			legacyMode, _ := cmd.Flags().GetBool("legacy")
			ingestPath, _ := cmd.Flags().GetString("ingest")

			if ingestPath != "" && !agentMode {
				return fmt.Errorf("--ingest requires --agent")
			}

			if agentMode && ingestPath != "" {
				if ingestPath == "-" {
					return runGuideAgentIngestFromReader(cmd.Context(), app, os.Stdin, ".alto", cmd.OutOrStdout())
				}
				return runGuideAgentIngest(cmd.Context(), app, ingestPath, ".alto", cmd.OutOrStdout())
			}

			return runGuide(cmd.Context(), app, noTUI, continueSession, agentMode, legacyMode)
		},
	}
	cmd.Flags().Bool("no-tui", false, "Disable TUI prompts, use plain stdin/stdout (accessibility, CI)")
	cmd.Flags().Bool("continue", false, "Resume a previously interrupted agent-mode session")
	cmd.Flags().Bool("agent", false, "Output discovery session as JSONL for AI agent consumption")
	cmd.Flags().String("ingest", "", "Ingest answers from JSONL file (or \"-\" for stdin); requires --agent")
	cmd.Flags().Bool("legacy", false, "Use deprecated question-based discovery flow")
	return cmd
}

func runGuide(ctx context.Context, app *composition.App, noTUI bool, continueSession bool, agentMode bool, legacyMode bool) error {
	if agentMode && continueSession {
		return fmt.Errorf("--agent and --continue are mutually exclusive")
	}
	if legacyMode && agentMode {
		return fmt.Errorf("--legacy and --agent are mutually exclusive")
	}
	if legacyMode && continueSession {
		return fmt.Errorf("--legacy and --continue are mutually exclusive")
	}

	// Guard: project must be initialized before any guide flow.
	if _, err := os.Stat(filepath.Join(".", ".alto", "config.toml")); os.IsNotExist(err) {
		return fmt.Errorf("project not initialized: run `alto init` first")
	}

	if legacyMode {
		return runLegacyFlow(ctx, app, noTUI)
	}
	if agentMode {
		return runGuideAgent(ctx, app)
	}

	if continueSession {
		return runGuideContinue(ctx, app, noTUI)
	}

	// Step 1: Detection
	result, err := app.DetectionHandler.Detect(".")
	if err != nil {
		return fmt.Errorf("detection: %w", err)
	}
	fmt.Printf("Detected %d tool(s)\n", len(result.DetectedTools()))

	// Step 2: Select prompter based on flag or env var
	var prompter application.StorytellingPrompter
	var boundaryPrompter application.BoundaryPrompter
	if noTUI || os.Getenv("ALTO_NO_TUI") == "1" {
		stdinScanner := bufio.NewScanner(os.Stdin)
		prompter = infrastructure.NewStdinStorytellingPrompterFromScanner(stdinScanner, os.Stdout)
		boundaryPrompter = infrastructure.NewStdinBoundaryPrompterFromScanner(stdinScanner, os.Stdout)
	} else {
		prompter = infrastructure.NewHuhStorytellingPrompter()
		boundaryPrompter = infrastructure.NewHuhBoundaryPrompter()
	}

	// Step 3: Create StorytellingHandler
	storyWriter := &infrastructure.StoryYAMLParser{}
	storytellingHandler := application.NewStorytellingHandler(storyWriter, prompter, nil)

	// Step 3.5: Create boundary detection dependencies
	boundaryDetectionHandler := application.NewBoundaryDetectionHandler(app.BoundaryDetector)
	contextMapWriter := &infrastructure.ContextMapYAMLParser{}

	// Step 4: Discovery (interactive storytelling)
	adapter := infrastructure.NewCLIDiscoveryAdapter(
		app.DiscoveryHandler,
		storytellingHandler,
		boundaryDetectionHandler,
		boundaryPrompter,
		contextMapWriter,
		prompter,
		".",
		infrastructure.WithArtifactPipeline(
			app.ArtifactGenerationHandler,
			app.GlossaryExportHandler,
			app.DiscoveryReportHandler,
			contextMapWriter,
			storyWriter,
		),
	)

	if err := adapter.Run(ctx); err != nil {
		return fmt.Errorf("discovery: %w", err)
	}

	fmt.Println("Discovery complete.")
	return nil
}

func runGuideAgent(ctx context.Context, app *composition.App) error {
	prompter := infrastructure.NewAgentStorytellingPrompter(domain.ModeRapid, nil)
	storyWriter := &infrastructure.StoryYAMLParser{}
	transformer := application.NewResearchToStoryTransformer()
	storytellingHandler := application.NewStorytellingHandler(storyWriter, prompter, transformer)

	var boundaryHandler *application.BoundaryDetectionHandler
	if app.BoundaryDetector != nil {
		boundaryHandler = application.NewBoundaryDetectionHandler(app.BoundaryDetector)
	}

	adapter := infrastructure.NewAgentStorytellingAdapter(
		app.DiscoveryHandler, storytellingHandler, boundaryHandler,
		app.DomainResearcher,
		os.Stdout, ".",
	)

	if err := adapter.Run(ctx); err != nil {
		return fmt.Errorf("agent storytelling: %w", err)
	}

	return nil
}

func runGuideContinue(ctx context.Context, app *composition.App, noTUI bool) error {
	// Step 1: Check session exists
	sessionRepo := infrastructure.NewFileSystemSessionRepository(".alto")
	exists, err := sessionRepo.Exists(ctx, "")
	if err != nil {
		return fmt.Errorf("checking session: %w", err)
	}
	if !exists {
		return fmt.Errorf("--continue requires a session started with `alto guide --agent`")
	}

	// Step 2: Load session
	session, err := sessionRepo.Load(ctx, "")
	if err != nil {
		return fmt.Errorf("could not load session: %w", err)
	}

	// Step 3: Guard completed
	if session.Status() == domain.StatusCompleted {
		return fmt.Errorf("session already complete. Start a new one with `alto guide`")
	}

	// Step 3.5: Warn if resuming a legacy-mode session
	if isLegacyMode(session.Mode()) {
		fmt.Fprintf(os.Stderr, legacyContinueWarning+"\n", session.Mode())
	}

	// Step 4: Fail fast if session is non-resumable (e.g. legacy mode without storytelling flow)
	if _, checkpointErr := session.ComputeResumeCheckpoint(); checkpointErr != nil {
		return fmt.Errorf("cannot resume session: %w", checkpointErr)
	}

	// Step 5: Display resume summary
	displayStoryResumeSummary(session)

	// Step 6: Register session in handler
	session, err = app.DiscoveryHandler.LoadOrGetSession(session.SessionID()) //nolint:contextcheck // Discovery interface deliberately omits context
	if err != nil {
		return fmt.Errorf("loading session into handler: %w", err)
	}

	// Step 7: Wire storytelling handler + adapter (match runGuide :84-110 pattern)
	var prompter application.StorytellingPrompter
	var boundaryPrompter application.BoundaryPrompter
	if noTUI || os.Getenv("ALTO_NO_TUI") == "1" {
		stdinScanner := bufio.NewScanner(os.Stdin)
		prompter = infrastructure.NewStdinStorytellingPrompterFromScanner(stdinScanner, os.Stdout)
		boundaryPrompter = infrastructure.NewStdinBoundaryPrompterFromScanner(stdinScanner, os.Stdout)
	} else {
		prompter = infrastructure.NewHuhStorytellingPrompter()
		boundaryPrompter = infrastructure.NewHuhBoundaryPrompter()
	}

	storyWriter := &infrastructure.StoryYAMLParser{}
	storytellingHandler := application.NewStorytellingHandler(storyWriter, prompter, nil)

	boundaryDetectionHandler := application.NewBoundaryDetectionHandler(app.BoundaryDetector)
	contextMapWriter := &infrastructure.ContextMapYAMLParser{}

	adapter := infrastructure.NewCLIDiscoveryAdapter(
		app.DiscoveryHandler,
		storytellingHandler,
		boundaryDetectionHandler,
		boundaryPrompter,
		contextMapWriter,
		prompter,
		".",
		infrastructure.WithArtifactPipeline(
			app.ArtifactGenerationHandler,
			app.GlossaryExportHandler,
			app.DiscoveryReportHandler,
			contextMapWriter,
			storyWriter,
		),
	)

	// Step 8: Resume
	if err := adapter.Resume(ctx, session); err != nil {
		return fmt.Errorf("resuming discovery: %w", err)
	}

	fmt.Println("Discovery resumed and complete.")
	return nil
}

func displayStoryResumeSummary(session *domain.DiscoverySession) {
	fmt.Println("Resuming storytelling session...")
	refs := session.StoryRefs()
	if len(refs) > 0 {
		fmt.Printf("Previously completed %d story(ies):\n", len(refs))
		for _, ref := range refs {
			fmt.Printf("  %s\n", ref)
		}
	}
	fmt.Println()
}

// responseEnvelope is the JSONL wrapper for ingest lines.
// Duplicated from agent_discovery_adapter.go because the original is unexported.
type responseEnvelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func runGuideAgentIngest(ctx context.Context, app *composition.App, ingestPath, altoDir string, w io.Writer) error {
	f, err := os.Open(ingestPath)
	if err != nil {
		return fmt.Errorf("opening ingest file: %w", err)
	}
	defer func() { _ = f.Close() }()
	return runGuideAgentIngestFromReader(ctx, app, f, altoDir, w)
}

func runGuideAgentIngestFromReader(ctx context.Context, app *composition.App, r io.Reader, altoDir string, w io.Writer) error {
	if _, err := os.Stat(filepath.Join(altoDir, "config.toml")); os.IsNotExist(err) {
		return fmt.Errorf("project not initialized: run `alto init` first")
	}
	renderer := infrastructure.NewJSONSessionRenderer()
	sessionRepo := infrastructure.NewFileSystemSessionRepository(altoDir)

	// Load persisted session
	exists, err := sessionRepo.Exists(ctx, "")
	if err != nil {
		return fmt.Errorf("checking session: %w", err)
	}
	if !exists {
		return fmt.Errorf("no session found. Run `alto guide --agent` first")
	}

	session, err := sessionRepo.Load(ctx, "")
	if err != nil {
		return fmt.Errorf("loading session: %w", err)
	}

	// Register in handler's in-memory map
	sessionID := session.SessionID()
	session, err = app.DiscoveryHandler.LoadOrGetSession(sessionID) //nolint:contextcheck // Discovery interface deliberately omits context
	if err != nil {
		return fmt.Errorf("loading session into handler: %w", err)
	}

	// Process JSONL lines
	scanner := bufio.NewScanner(r)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var env responseEnvelope
		if unmarshalErr := json.Unmarshal(line, &env); unmarshalErr != nil {
			return fmt.Errorf("line %d: invalid JSON: %w", lineNum, unmarshalErr)
		}

		switch env.Type {
		case "persona_response":
			pr, parseErr := renderer.ParsePersonaResponse(env.Data)
			if parseErr != nil {
				return fmt.Errorf("line %d: %w", lineNum, parseErr)
			}
			if pr.SessionID != sessionID {
				return fmt.Errorf("line %d: session ID mismatch: expected %s, got %s", lineNum, sessionID, pr.SessionID)
			}
			session, err = app.DiscoveryHandler.DetectPersona(sessionID, pr.Choice) //nolint:contextcheck // Discovery interface deliberately omits context
			if err != nil {
				return fmt.Errorf("line %d: detecting persona: %w", lineNum, err)
			}

		case "answer":
			ai, parseErr := renderer.ParseAnswerInput(env.Data)
			if parseErr != nil {
				return fmt.Errorf("line %d: %w", lineNum, parseErr)
			}
			if ai.SessionID != sessionID {
				return fmt.Errorf("line %d: session ID mismatch: expected %s, got %s", lineNum, sessionID, ai.SessionID)
			}

			if ai.Skipped {
				reason := ai.SkipReason
				if reason == "" {
					reason = "skipped by agent"
				}
				session, err = app.DiscoveryHandler.SkipQuestion(sessionID, ai.QuestionID, reason) //nolint:contextcheck // Discovery interface deliberately omits context
				if err != nil {
					return fmt.Errorf("line %d: skipping question %s: %w", lineNum, ai.QuestionID, err)
				}
			} else {
				session, err = app.DiscoveryHandler.AnswerQuestion(sessionID, ai.QuestionID, ai.Answer) //nolint:contextcheck // Discovery interface deliberately omits context
				if err != nil {
					return fmt.Errorf("line %d: answering question %s: %w", lineNum, ai.QuestionID, err)
				}
				// Auto-confirm playback if triggered
				if session.Status() == domain.StatusPlaybackPending {
					session, err = app.DiscoveryHandler.ConfirmPlayback(sessionID, true) //nolint:contextcheck // Discovery interface deliberately omits context
					if err != nil {
						return fmt.Errorf("line %d: auto-confirming playback: %w", lineNum, err)
					}
				}
			}

		case "playback_response":
			pr, parseErr := renderer.ParsePlaybackResponse(env.Data)
			if parseErr != nil {
				return fmt.Errorf("line %d: %w", lineNum, parseErr)
			}
			if pr.SessionID != sessionID {
				return fmt.Errorf("line %d: session ID mismatch: expected %s, got %s", lineNum, sessionID, pr.SessionID)
			}
			session, err = app.DiscoveryHandler.ConfirmPlayback(sessionID, pr.Confirmed) //nolint:contextcheck // Discovery interface deliberately omits context
			if err != nil {
				return fmt.Errorf("line %d: confirming playback: %w", lineNum, err)
			}

		default:
			return fmt.Errorf("line %d: unknown response type %q", lineNum, env.Type)
		}
	}

	if scanErr := scanner.Err(); scanErr != nil {
		return fmt.Errorf("reading ingest input: %w", scanErr)
	}

	// Attempt completion if session is in answering state
	if session.Status() == domain.StatusAnswering {
		completed, completeErr := app.DiscoveryHandler.Complete(sessionID) //nolint:contextcheck // Discovery interface deliberately omits context
		if completeErr == nil {
			session = completed
		}
		// If Complete fails (e.g., MVP questions not all answered), that's fine — partial state
	}

	// Emit final session status
	statusData, err := renderer.RenderSessionStatus(session)
	if err != nil {
		return fmt.Errorf("rendering final status: %w", err)
	}
	finalEnv := responseEnvelope{
		Type: "session_status",
		Data: json.RawMessage(statusData),
	}
	finalLine, err := json.Marshal(finalEnv)
	if err != nil {
		return fmt.Errorf("marshaling final status: %w", err)
	}
	if _, err := fmt.Fprintf(w, "%s\n", finalLine); err != nil {
		return fmt.Errorf("writing final status: %w", err)
	}

	return nil
}

// isLegacyMode returns true for the deprecated question-based discovery modes.
func isLegacyMode(mode domain.DiscoveryMode) bool {
	switch mode { //nolint:exhaustive // Only legacy modes return true; all others (including future) default to false.
	case domain.ModeExpress, domain.ModeDeep, domain.ModeConversational:
		return true
	default:
		return false
	}
}

// runLegacyFlow runs the deprecated question-based discovery flow.
func runLegacyFlow(ctx context.Context, app *composition.App, noTUI bool) error {
	var prompter application.Prompter
	if noTUI || os.Getenv("ALTO_NO_TUI") == "1" {
		prompter = infrastructure.NewStdinPrompter(os.Stdin, os.Stdout)
	} else {
		prompter = infrastructure.NewHuhPrompter()
	}
	return runLegacyFlowWithDeps(ctx, app, prompter, os.Stderr)
}

// runLegacyFlowWithDeps runs the legacy flow with injected dependencies (testable).
func runLegacyFlowWithDeps(ctx context.Context, app *composition.App, prompter application.Prompter, stderr io.Writer) error {
	if _, err := fmt.Fprintln(stderr, legacyFlagWarning); err != nil {
		return fmt.Errorf("writing deprecation warning: %w", err)
	}

	session, err := app.DiscoveryHandler.StartSession("")
	if err != nil {
		return fmt.Errorf("starting legacy session: %w", err)
	}
	sessionID := session.SessionID()

	// Persona selection
	choice, err := prompter.SelectPersona(ctx)
	if err != nil {
		return fmt.Errorf("selecting persona: %w", err)
	}
	if _, personaErr := app.DiscoveryHandler.DetectPersona(sessionID, choice); personaErr != nil { //nolint:contextcheck // Discovery interface deliberately omits context
		return fmt.Errorf("detecting persona: %w", personaErr)
	}

	// Question loop
	questions := domain.QuestionCatalog()
	for _, q := range questions {
		// Handle playback gate before next question
		session, err = app.DiscoveryHandler.GetSession(sessionID)
		if err != nil {
			return fmt.Errorf("getting session: %w", err)
		}
		if session.Status() == domain.StatusPlaybackPending {
			summary := fmt.Sprintf("Review your answers so far (%d answered)", len(session.Answers()))
			confirmed, pbErr := prompter.ConfirmPlayback(ctx, summary)
			if pbErr != nil {
				return fmt.Errorf("confirming playback: %w", pbErr)
			}
			if _, pbConfirmErr := app.DiscoveryHandler.ConfirmPlayback(sessionID, confirmed); pbConfirmErr != nil { //nolint:contextcheck // Discovery interface deliberately omits context
				return fmt.Errorf("confirming playback: %w", pbConfirmErr)
			}
		}

		answer, askErr := prompter.AskQuestion(ctx, q.TechnicalText())
		if askErr != nil {
			return fmt.Errorf("asking question %s: %w", q.ID(), askErr)
		}
		if answer == "" {
			reason, skipErr := prompter.AskSkipReason(ctx)
			if skipErr != nil {
				return fmt.Errorf("asking skip reason for %s: %w", q.ID(), skipErr)
			}
			if _, skipQErr := app.DiscoveryHandler.SkipQuestion(sessionID, q.ID(), reason); skipQErr != nil { //nolint:contextcheck // Discovery interface deliberately omits context
				return fmt.Errorf("skipping question %s: %w", q.ID(), skipQErr)
			}
		} else {
			if _, answerErr := app.DiscoveryHandler.AnswerQuestion(sessionID, q.ID(), answer); answerErr != nil { //nolint:contextcheck // Discovery interface deliberately omits context
				return fmt.Errorf("answering question %s: %w", q.ID(), answerErr)
			}
		}
	}

	// Complete
	if _, completeErr := app.DiscoveryHandler.Complete(sessionID); completeErr != nil { //nolint:contextcheck // Discovery interface deliberately omits context
		return fmt.Errorf("completing legacy session: %w", completeErr)
	}

	fmt.Println("Legacy discovery complete.")
	return nil
}
