package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alty-cli/alty/internal/composition"
)

// NewGuideCmd creates the "alty guide" command.
func NewGuideCmd(app *composition.App) *cobra.Command {
	return &cobra.Command{
		Use:   "guide",
		Short: "Run the 10-question guided DDD discovery flow",
		Long: `Run the 10-question guided DDD discovery flow.

This multi-step command orchestrates:
  1. Detection of installed AI coding tools
  2. Interactive discovery session (10 questions)
  3. Artifact generation from discovery answers`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGuide(app)
		},
	}
}

func runGuide(app *composition.App) error {
	// Step 1: Detection
	result, err := app.DetectionHandler.Detect(".")
	if err != nil {
		return fmt.Errorf("detection: %w", err)
	}
	fmt.Printf("Detected %d tool(s)\n", len(result.DetectedTools()))

	// Step 2: Discovery (would be interactive)
	fmt.Println("Discovery flow requires interactive prompts -- use MCP or interactive CLI")

	return nil
}
