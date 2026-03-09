package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alty-cli/alty/internal/composition"
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

// NewDocReviewCmd creates the "alty doc-review" command with subcommands.
func NewDocReviewCmd(app *composition.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doc-review",
		Short: "Manage documentation review status",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default to list when no subcommand provided.
			return runDocReviewList(cmd, app)
		},
	}

	cmd.AddCommand(newDocReviewListCmd(app))
	cmd.AddCommand(newDocReviewMarkCmd(app))
	cmd.AddCommand(newDocReviewMarkAllCmd(app))

	return cmd
}

func newDocReviewListCmd(app *composition.App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List documents due for review",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDocReviewList(cmd, app)
		},
	}
}

func runDocReviewList(cmd *cobra.Command, app *composition.App) error {
	docs, err := app.DocReviewHandler.ReviewableDocs(context.Background(), ".")
	if err != nil {
		return fmt.Errorf("doc review list: %w", err)
	}

	if len(docs) == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No docs due for review.")
		return nil
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Docs Due for Review")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "----------------------------------------")
	for _, doc := range docs {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %-40s %s\n", doc.Path(), doc.Status())
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nTotal: %d doc(s) due for review\n", len(docs))
	return nil
}

func newDocReviewMarkCmd(app *composition.App) *cobra.Command {
	return &cobra.Command{
		Use:   "mark <doc-path>",
		Short: "Mark a document as reviewed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			docPath := args[0]
			result, err := app.DocReviewHandler.MarkReviewed(
				context.Background(), docPath, ".", nil,
			)
			if err != nil {
				return fmt.Errorf("doc review mark: %w", err)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Marked %s as reviewed on %s\n",
				result.Path(), result.NewDate().Format("2006-01-02"))
			return nil
		},
	}
}

func newDocReviewMarkAllCmd(app *composition.App) *cobra.Command {
	return &cobra.Command{
		Use:   "mark-all",
		Short: "Mark all stale documents as reviewed",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := app.DocReviewHandler.MarkAllReviewed(
				context.Background(), ".", nil,
			)
			if err != nil {
				return fmt.Errorf("doc review mark-all: %w", err)
			}

			if len(results) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No docs needed marking.")
				return nil
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Marked as Reviewed")
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "----------------------------------------")
			for _, r := range results {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s (%s)\n",
					r.Path(), r.NewDate().Format("2006-01-02"))
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nTotal: %d doc(s) marked as reviewed\n", len(results))
			return nil
		},
	}
}
