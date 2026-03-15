package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alty-cli/alty/internal/composition"
	"github.com/alty-cli/alty/internal/discovery/application"
	"github.com/alty-cli/alty/internal/discovery/domain"
	"github.com/alty-cli/alty/internal/discovery/infrastructure"
)

// NewGuideCmd creates the "alty guide" command.
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
Use --continue to resume a previously interrupted session.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			noTUI, _ := cmd.Flags().GetBool("no-tui")
			continueSession, _ := cmd.Flags().GetBool("continue")
			return runGuide(cmd.Context(), app, noTUI, continueSession)
		},
	}
	cmd.Flags().Bool("no-tui", false, "Disable TUI prompts, use plain stdin/stdout (accessibility, CI)")
	cmd.Flags().Bool("continue", false, "Resume a previously interrupted discovery session")
	return cmd
}

func runGuide(ctx context.Context, app *composition.App, noTUI bool, continueSession bool) error {
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
	if noTUI || os.Getenv("ALTY_NO_TUI") == "1" {
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

func runGuideContinue(ctx context.Context, app *composition.App, noTUI bool) error {
	// Step 1: Check session exists via repository
	sessionRepo := infrastructure.NewFileSystemSessionRepository(".alty")
	exists, err := sessionRepo.Exists(ctx, "")
	if err != nil {
		return fmt.Errorf("checking session: %w", err)
	}
	if !exists {
		return fmt.Errorf("no session to continue. Run `alty guide` first")
	}

	// Step 2: Load session
	session, err := sessionRepo.Load(ctx, "")
	if err != nil {
		return fmt.Errorf("could not load session: %w", err)
	}

	// Step 3: Check if already completed
	if session.Status() == domain.StatusCompleted {
		return fmt.Errorf("session already complete. Start a new one with `alty guide`")
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
	if noTUI || os.Getenv("ALTY_NO_TUI") == "1" {
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
