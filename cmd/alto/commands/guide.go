package commands

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alto-cli/alto/internal/composition"
	"github.com/alto-cli/alto/internal/discovery/application"
	"github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/discovery/infrastructure"
)

// NewGuideCmd creates the "alto guide" command.
func NewGuideCmd(app *composition.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "guide",
		Short: "Run the 10-question guided DDD discovery flow",
		Long: `Run the 10-question guided DDD discovery flow.

This multi-step command orchestrates:
  1. Detection of installed AI coding tools
  2. Interactive discovery session (10 questions)
  3. Artifact generation from discovery answers

Use --no-tui for accessibility (screen readers) or CI/scripted input.
Use --continue to resume a previously interrupted session.
Use --agent to output the discovery session as JSONL for AI agent consumption.
Use --agent --ingest <file> to ingest answers from a JSONL file (or "-" for stdin).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			noTUI, _ := cmd.Flags().GetBool("no-tui")
			continueSession, _ := cmd.Flags().GetBool("continue")
			agentMode, _ := cmd.Flags().GetBool("agent")
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

			return runGuide(cmd.Context(), app, noTUI, continueSession, agentMode)
		},
	}
	cmd.Flags().Bool("no-tui", false, "Disable TUI prompts, use plain stdin/stdout (accessibility, CI)")
	cmd.Flags().Bool("continue", false, "Resume a previously interrupted discovery session")
	cmd.Flags().Bool("agent", false, "Output discovery session as JSONL for AI agent consumption")
	cmd.Flags().String("ingest", "", "Ingest answers from JSONL file (or \"-\" for stdin); requires --agent")
	return cmd
}

func runGuide(ctx context.Context, app *composition.App, noTUI bool, continueSession bool, agentMode bool) error {
	if agentMode && continueSession {
		return fmt.Errorf("--agent and --continue are mutually exclusive")
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
	var prompter application.Prompter
	if noTUI || os.Getenv("ALTO_NO_TUI") == "1" {
		prompter = infrastructure.NewStdinPrompter(os.Stdin, os.Stdout)
	} else {
		prompter = infrastructure.NewHuhPrompter()
	}

	// Step 3: Discovery (interactive)
	adapter := infrastructure.NewCLIDiscoveryAdapter(app.DiscoveryHandler, prompter, ".")

	if err := adapter.Run(ctx); err != nil {
		return fmt.Errorf("discovery: %w", err)
	}

	fmt.Println("Discovery complete.")
	return nil
}

func runGuideAgent(ctx context.Context, app *composition.App) error {
	renderer := infrastructure.NewJSONSessionRenderer()
	adapter := infrastructure.NewAgentDiscoveryAdapter(app.DiscoveryHandler, renderer, os.Stdout, ".")
	if err := adapter.Run(ctx); err != nil {
		return fmt.Errorf("agent discovery: %w", err)
	}
	return nil
}

func runGuideContinue(ctx context.Context, app *composition.App, noTUI bool) error {
	// Step 1: Check session exists via repository
	sessionRepo := infrastructure.NewFileSystemSessionRepository(".alto")
	exists, err := sessionRepo.Exists(ctx, "")
	if err != nil {
		return fmt.Errorf("checking session: %w", err)
	}
	if !exists {
		return fmt.Errorf("no session to continue. Run `alto guide` first")
	}

	// Step 2: Load session
	session, err := sessionRepo.Load(ctx, "")
	if err != nil {
		return fmt.Errorf("could not load session: %w", err)
	}

	// Step 3: Check if already completed
	if session.Status() == domain.StatusCompleted {
		return fmt.Errorf("session already complete. Start a new one with `alto guide`")
	}

	// Step 4: Register session in handler for subsequent operations
	session, err = app.DiscoveryHandler.LoadOrGetSession(session.SessionID()) //nolint:contextcheck // Discovery interface deliberately omits context
	if err != nil {
		return fmt.Errorf("loading session into handler: %w", err)
	}

	// Step 5: Display summary
	displaySessionSummary(session)

	// Step 6: Select prompter
	var prompter application.Prompter
	if noTUI || os.Getenv("ALTO_NO_TUI") == "1" {
		prompter = infrastructure.NewStdinPrompter(os.Stdin, os.Stdout)
	} else {
		prompter = infrastructure.NewHuhPrompter()
	}

	// Step 7: Determine which questions need answering
	questions := domain.QuestionCatalog()
	answeredIDs := make(map[string]bool)
	for _, a := range session.Answers() {
		answeredIDs[a.QuestionID()] = true
	}

	register, hasRegister := session.Register()
	if !hasRegister {
		return fmt.Errorf("session has no register — cannot continue")
	}

	sessionID := session.SessionID()

	// Step 8: Unskip all skipped questions so they can be re-asked
	for _, q := range questions {
		if session.SkipReason(q.ID()) != "" {
			if unskipErr := session.UnskipQuestion(q.ID()); unskipErr != nil {
				return fmt.Errorf("unskipping question %s: %w", q.ID(), unskipErr)
			}
		}
	}

	// Step 9: Resume question loop from first unanswered question
	for i, question := range questions {
		if answeredIDs[question.ID()] {
			continue
		}

		var text string
		if register == domain.RegisterTechnical {
			text = question.TechnicalText()
		} else {
			text = question.NonTechnicalText()
		}

		fmt.Printf("Q%d/%d\n", i+1, len(questions))
		answer, askErr := prompter.AskQuestion(ctx, text)
		if askErr != nil {
			return fmt.Errorf("asking question %s: %w", question.ID(), askErr)
		}

		if answer == "" {
			reason, skipErr := prompter.AskSkipReason(ctx)
			if skipErr != nil {
				return fmt.Errorf("asking skip reason: %w", skipErr)
			}
			session, err = app.DiscoveryHandler.SkipQuestion(sessionID, question.ID(), reason) //nolint:contextcheck // Discovery interface deliberately omits context
			if err != nil {
				return fmt.Errorf("skipping question %s: %w", question.ID(), err)
			}
		} else {
			session, err = app.DiscoveryHandler.AnswerQuestion(sessionID, question.ID(), answer) //nolint:contextcheck // Discovery interface deliberately omits context
			if err != nil {
				return fmt.Errorf("answering question %s: %w", question.ID(), err)
			}
		}

		// Check for playback pending
		if session.Status() == domain.StatusPlaybackPending {
			summary := buildContinuePlaybackSummary(session, register)
			confirmed, playbackErr := prompter.ConfirmPlayback(ctx, summary)
			if playbackErr != nil {
				return fmt.Errorf("playback confirmation: %w", playbackErr)
			}
			_, err = app.DiscoveryHandler.ConfirmPlayback(sessionID, confirmed)
			if err != nil {
				return fmt.Errorf("confirming playback: %w", err)
			}
		}
	}

	fmt.Println("Discovery complete.")
	return nil
}

func displaySessionSummary(session *domain.DiscoverySession) {
	fmt.Println("Resuming discovery session...")

	answers := session.Answers()
	if len(answers) > 0 {
		fmt.Println("Previously answered:")
		for _, a := range answers {
			label := a.QuestionID()
			// Truncate long answers for display
			response := a.ResponseText()
			if len(response) > 60 {
				response = response[:60] + "..."
			}
			fmt.Printf("  %s: %s\n", label, response)
		}
	}

	// Show skipped questions
	questions := domain.QuestionCatalog()
	var skippedLines []string
	for _, q := range questions {
		if reason := session.SkipReason(q.ID()); reason != "" {
			skippedLines = append(skippedLines, fmt.Sprintf("  %s (reason: %s)", q.ID(), reason))
		}
	}
	if len(skippedLines) > 0 {
		fmt.Println("\nSkipped (will ask again):")
		for _, line := range skippedLines {
			fmt.Println(line)
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

func buildContinuePlaybackSummary(session *domain.DiscoverySession, register domain.DiscoveryRegister) string {
	answers := session.Answers()
	if len(answers) == 0 {
		return "No answers recorded yet."
	}

	qByID := domain.QuestionByID()
	var sb strings.Builder

	for _, ans := range answers {
		q, ok := qByID[ans.QuestionID()]
		if !ok {
			continue
		}
		var qText string
		if register == domain.RegisterTechnical {
			qText = q.TechnicalText()
		} else {
			qText = q.NonTechnicalText()
		}
		fmt.Fprintf(&sb, "Q: %s\nA: %s\n\n", qText, ans.ResponseText())
	}

	return strings.TrimSpace(sb.String())
}
