package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alty-cli/alty/internal/composition"
	"github.com/alty-cli/alty/internal/discovery/application"
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

Use --no-tui for accessibility (screen readers) or CI/scripted input.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			noTUI, _ := cmd.Flags().GetBool("no-tui")
			return runGuide(cmd.Context(), app, noTUI)
		},
	}
	cmd.Flags().Bool("no-tui", false, "Disable TUI prompts, use plain stdin/stdout (accessibility, CI)")
	return cmd
}

func runGuide(ctx context.Context, app *composition.App, noTUI bool) error {
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
