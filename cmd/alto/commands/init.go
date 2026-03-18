package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alto-cli/alto/internal/bootstrap/domain"
	"github.com/alto-cli/alto/internal/composition"
)

// NewInitCmd creates the "alto init" command.
func NewInitCmd(app *composition.App) *cobra.Command {
	var (
		existing    bool
		dryRun      bool
		yes         bool
		forceBranch bool
		noCommit    bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Bootstrap a new project from a README idea",
		Long: `Bootstrap a new project from a README idea.

Auto-detects whether the current directory contains an existing project
and chooses the appropriate path. Use --existing to force rescue mode.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// --existing flag overrides auto-detection.
			if existing {
				return runRescue(cmd, app, dryRun, forceBranch)
			}

			// Auto-detect project state.
			result, err := app.ProjectDetector.Detect(".")
			if err != nil {
				return fmt.Errorf("detecting project state: %w", err)
			}

			if result.IsExistingProject() && !result.IsAmbiguous() {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Detected existing %s project (%s). Running rescue mode.\n",
					result.Language(), result.ManifestPath())
				return runRescue(cmd, app, dryRun, forceBranch)
			}

			if result.IsAmbiguous() {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Found docs/ folder but no source code.")
				_, _ = fmt.Fprint(cmd.OutOrStdout(), "Treat as existing project? [y/N] ")
				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
					if answer == "y" || answer == "yes" {
						return runRescue(cmd, app, dryRun, forceBranch)
					}
				}
			}

			return runInit(cmd, app, dryRun, yes, noCommit, result)
		},
	}

	cmd.Flags().BoolVar(&existing, "existing", false, "Rescue an existing project")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show plan without executing")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&forceBranch, "force-branch", false, "Delete existing alto/init branch before creating a new one")
	cmd.Flags().BoolVar(&noCommit, "no-commit", false, "Skip auto-commit of generated files")

	return cmd
}

func runInit(cmd *cobra.Command, app *composition.App, dryRun bool, yes bool, noCommit bool, detection domain.ProjectDetectionResult) error {
	projectDir := "."

	// Clear GitCommitter when --no-commit is set.
	if noCommit {
		app.BootstrapHandler.SetGitCommitter(nil)
	}

	// 1. Preview bootstrap actions.
	session, err := app.BootstrapHandler.Preview(projectDir)
	if err != nil {
		return fmt.Errorf("bootstrap preview: %w", err)
	}

	// Build ProjectConfig from detection result and detected tools.
	projectName := filepath.Base(mustAbs(projectDir))
	config := domain.NewProjectConfig(
		projectName,
		detection.Language(),
		detection.ModulePath(),
		session.DetectedTools(),
	)
	app.BootstrapHandler.WithProjectConfig(session.SessionID(), config)

	// 2. Display plan.
	preview := session.Preview()
	if preview != nil {
		for _, action := range preview.FileActions() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s", action.ActionType(), action.Path())
			if action.Reason() != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), " (%s)", action.Reason())
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout())
		}
	}

	// 3. Dry-run: show preview and exit.
	if dryRun {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Dry run: no files written.")
		return nil
	}

	// 4. Confirm — require explicit user approval before writing files.
	if !yes {
		_, _ = fmt.Fprint(cmd.OutOrStdout(), "\nProceed? [y/N] ")
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			return fmt.Errorf("bootstrap cancelled")
		}
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if answer != "y" && answer != "yes" {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Cancelled.")
			return nil
		}
	}

	_, err = app.BootstrapHandler.Confirm(session.SessionID())
	if err != nil {
		return fmt.Errorf("bootstrap confirm: %w", err)
	}

	// 5. Execute (writes files).
	_, err = app.BootstrapHandler.Execute(session.SessionID())
	if err != nil {
		return fmt.Errorf("bootstrap execute: %w", err)
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Bootstrap complete. Starting guided discovery...")

	// 6. Launch guide flow.
	return runGuide(cmd.Context(), app, false, false, false)
}

func runRescue(cmd *cobra.Command, app *composition.App, dryRun bool, forceBranch bool) error {
	ctx := context.Background()
	projectDir := "."

	// 1. Validate preconditions.
	if err := app.RescueHandler.ValidatePreconditions(ctx, projectDir, forceBranch); err != nil {
		return fmt.Errorf("rescue preconditions: %w", err)
	}

	// 2. Run rescue analysis.
	analysis, err := app.RescueHandler.Rescue(ctx, projectDir, nil, true, forceBranch)
	if err != nil {
		return fmt.Errorf("rescue analysis: %w", err)
	}

	// 3. Print gap report.
	gaps := analysis.Gaps()
	if len(gaps) == 0 {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No gaps found, project is compliant.")
		return nil
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Gap Analysis Report")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "----------------------------------------")
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%-40s %-20s %s\n", "PATH", "TYPE", "SEVERITY")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "----------------------------------------")
	for _, gap := range gaps {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%-40s %-20s %s\n",
			gap.Path(), gap.GapType(), gap.Severity())
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout())

	// 4. Check for plan (no plan means analyzed but no actionable gaps).
	plan := analysis.Plan()
	if plan == nil {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No gaps found, project is compliant.")
		return nil
	}

	// 5. Dry-run: show plan but don't execute.
	if dryRun {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Dry run: would create %d file(s) on branch %s\n",
			len(plan.Gaps()), plan.BranchName())
		return nil
	}

	// 6. Execute plan.
	if err := app.RescueHandler.ExecutePlan(ctx, analysis); err != nil {
		return fmt.Errorf("execute plan: %w", err)
	}

	// 7. Print results.
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Rescue Complete")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "----------------------------------------")
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Branch: %s\n", plan.BranchName())
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Files created: %d\n", len(plan.Gaps()))

	return nil
}

// mustAbs returns the absolute path of dir, falling back to dir itself on error.
func mustAbs(dir string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return dir
	}
	return abs
}
