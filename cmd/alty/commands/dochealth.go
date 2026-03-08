package commands

import (
	"context"
	"fmt"

	"github.com/alty-cli/alty/internal/composition"
	"github.com/spf13/cobra"
)

// NewDocHealthCmd creates the "alty doc-health" command.
func NewDocHealthCmd(app *composition.App) *cobra.Command {
	return &cobra.Command{
		Use:   "doc-health [project-dir]",
		Short: "Check documentation freshness and health",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectDir := "."
			if len(args) > 0 {
				projectDir = args[0]
			}

			report, err := app.DocHealthHandler.Handle(
				context.Background(), projectDir,
			)
			if err != nil {
				return fmt.Errorf("doc health: %w", err)
			}

			fmt.Println("Doc Health Report")
			fmt.Println("----------------------------------------")

			for _, status := range report.Statuses() {
				icon := "  ? "
				switch string(status.Status()) {
				case "ok":
					icon = "  OK"
				case "stale":
					icon = "  !!"
				case "missing":
					icon = "  XX"
				case "no_frontmatter":
					icon = "  ! "
				}
				fmt.Printf("%s %-40s %s\n", icon, status.Path(), status.Status())
			}

			fmt.Println()
			fmt.Printf("Summary: %d checked, %d issue(s) found\n",
				report.TotalChecked(), report.IssueCount())

			if report.HasIssues() {
				return fmt.Errorf("documentation health check found issues")
			}

			return nil
		},
	}
}

// NewDocReviewCmd creates the "alty doc-review" command.
func NewDocReviewCmd(_ *composition.App) *cobra.Command {
	return &cobra.Command{
		Use:   "doc-review [doc-path]",
		Short: "Mark documentation as reviewed",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Doc review command requires DocReviewHandler (not yet wired).")
			return nil
		},
	}
}
