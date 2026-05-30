package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alto-cli/alto/internal/composition"
)

// NewDocHealthCmd creates the "alto doc-health" command.
//
// Modes:
//   - Default: validates docs/ (registered docs + unregistered scan)
//   - --paths: also validates scaffold assets under the given directories
//     (e.g. --paths=.alto/). Multi-value: --paths=.alto/,custom-dir/ or
//     repeat the flag.
//
// Exit code is non-zero when EITHER mode reports issues.
func NewDocHealthCmd(app *composition.App) *cobra.Command {
	var paths []string
	cmd := &cobra.Command{
		Use:   "doc-health [project-dir]",
		Short: "Check documentation freshness and health",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var docHealthFailed, scaffoldFailed bool

			// docs/ freshness runs when (a) no --paths is given (legacy
			// invocation) OR (b) a positional project-dir is provided
			// alongside --paths. `--paths=.alto/` alone is purely scaffold.
			runDocs := len(paths) == 0 || len(args) > 0
			if runDocs {
				projectDir := "."
				if len(args) > 0 {
					projectDir = args[0]
				}
				report, err := app.DocHealthHandler.Handle(context.Background(), projectDir)
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
					docHealthFailed = true
				}
			}

			for _, p := range paths {
				if err := runScaffoldCheck(cmd, app, p, &scaffoldFailed); err != nil {
					return err
				}
			}

			if docHealthFailed || scaffoldFailed {
				return fmt.Errorf("documentation health check found issues")
			}
			return nil
		},
	}
	cmd.Flags().StringSliceVar(&paths, "paths", nil,
		"Comma-separated additional paths to validate (e.g. --paths=.alto/,custom-dir/)")
	return cmd
}

func runScaffoldCheck(cmd *cobra.Command, app *composition.App, altoDir string, failed *bool) error {
	out := cmd.OutOrStdout()
	report, err := app.ScaffoldHealthHandler.Handle(context.Background(), altoDir)
	if err != nil {
		return fmt.Errorf("scaffold health %s: %w", altoDir, err)
	}
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintf(out, "Scaffold Health Report — %s\n", altoDir)
	_, _ = fmt.Fprintln(out, "----------------------------------------")
	for _, v := range report.Violations() {
		_, _ = fmt.Fprintf(out, "  %s  %-32s  %s — %s\n",
			v.Severity(), v.Rule(), v.File(), v.Message())
	}
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintf(out, "Summary: %d violation(s) (%d error, %d warning)\n",
		report.TotalCount(), report.ErrorCount(), report.WarningCount())
	if report.HasErrors() {
		*failed = true
	}
	return nil
}

// NewDocReviewCmd creates the "alto doc-review" command with subcommands.
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
