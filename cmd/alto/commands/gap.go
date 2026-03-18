package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alto-cli/alto/internal/composition"
)

// NewGapCmd creates the "alto gap" command.
func NewGapCmd(app *composition.App) *cobra.Command {
	return &cobra.Command{
		Use:   "gap [project-dir]",
		Short: "Analyze project for structural gaps without modifying anything",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir := "."
			if len(args) > 0 {
				projectDir = args[0]
			}

			ctx := context.Background()

			report, err := app.GapQueryHandler.AnalyzeGaps(ctx, projectDir)
			if err != nil {
				return fmt.Errorf("gap analysis: %w", err)
			}

			_, _ = fmt.Fprint(cmd.OutOrStdout(), report.FormatReport())

			if report.HasRequired {
				return fmt.Errorf("required gaps found")
			}

			return nil
		},
	}
}
