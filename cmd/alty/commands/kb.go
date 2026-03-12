package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alty-cli/alty/internal/composition"
	"github.com/alty-cli/alty/internal/knowledge/domain"
)

// NewKBCmd creates the "alty kb" command with subcommands.
func NewKBCmd(app *composition.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kb",
		Short: "Knowledge base operations",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default: list categories
			return runKBList(cmd, app)
		},
	}

	cmd.AddCommand(newKBLookupCmd(app))
	cmd.AddCommand(newKBDriftCmd(app))

	return cmd
}

// newKBLookupCmd creates the "alty kb lookup <topic>" subcommand.
func newKBLookupCmd(app *composition.App) *cobra.Command {
	return &cobra.Command{
		Use:   "lookup <topic>",
		Short: "Look up a topic in the RLM knowledge base",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			topic := args[0]
			entry, err := app.KnowledgeLookupHandler.Lookup(
				context.Background(), topic, "",
			)
			if err != nil {
				return fmt.Errorf("lookup %q: %w", topic, err)
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), entry.Content())
			return nil
		},
	}
}

// newKBDriftCmd creates the "alty kb drift [tool]" subcommand.
func newKBDriftCmd(app *composition.App) *cobra.Command {
	return &cobra.Command{
		Use:   "drift [tool]",
		Short: "Detect drift in knowledge base entries",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var toolFilter *string
			if len(args) > 0 {
				toolFilter = &args[0]
			}

			report, err := app.DriftDetectionHandler.DetectDrift(
				context.Background(), toolFilter,
			)
			if err != nil {
				return fmt.Errorf("detect drift: %w", err)
			}

			if !report.HasDrift() {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No drift detected.")
				return nil
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Drift Report")
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "----------------------------------------")

			hasError := false
			for _, sig := range report.Signals() {
				icon := severityIcon(sig.Severity())
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s %-10s %-30s %s\n",
					icon, sig.SignalType(), sig.EntryPath(), sig.Description())
				if sig.Severity() == domain.SeverityError {
					hasError = true
				}
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout())
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Total: %d issue(s) found\n", report.TotalCount())

			if hasError {
				return fmt.Errorf("drift detected with errors")
			}
			return nil
		},
	}
}

// runKBList lists available knowledge base categories.
func runKBList(cmd *cobra.Command, app *composition.App) error {
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Knowledge Base Categories")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "----------------------------------------")
	for _, cat := range app.KnowledgeLookupHandler.ListCategories() {
		topics, err := app.KnowledgeLookupHandler.ListTopics(
			context.Background(), cat, nil,
		)
		if err != nil {
			topics = nil
		}
		if len(topics) > 0 {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s: %s\n", cat, joinStrings(topics))
		} else {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s: (empty)\n", cat)
		}
	}
	return nil
}

// severityIcon returns a display icon for a drift severity.
func severityIcon(severity domain.DriftSeverity) string {
	switch severity {
	case domain.SeverityError:
		return "[ERR]"
	case domain.SeverityWarning:
		return "[WRN]"
	case domain.SeverityInfo:
		return "[INF]"
	default:
		return "[???]"
	}
}

func joinStrings(items []string) string {
	result := ""
	for i, s := range items {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}
