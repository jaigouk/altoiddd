package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/alty-cli/alty/internal/composition"
)

// NewInitCmd creates the "alty init" command.
func NewInitCmd(app *composition.App) *cobra.Command {
	var (
		existing bool
		dryRun   bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Bootstrap a new project from a README idea",
		Long: `Bootstrap a new project from a README idea.

Use --existing to rescue an existing project (alty init --existing).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if existing {
				return runRescue(cmd, app, dryRun)
			}
			return runInit(app)
		},
	}

	cmd.Flags().BoolVar(&existing, "existing", false, "Rescue an existing project")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show rescue plan without executing")

	return cmd
}

func runInit(app *composition.App) error {
	// Preview bootstrap actions
	session, err := app.BootstrapHandler.Preview(".")
	if err != nil {
		return fmt.Errorf("bootstrap preview: %w", err)
	}
	fmt.Printf("Bootstrap session created: %s\n", session.SessionID())
	fmt.Println("Run 'alty guide' to start the discovery flow.")
	return nil
}

func runRescue(cmd *cobra.Command, app *composition.App, dryRun bool) error {
	ctx := context.Background()
	projectDir := "."

	// 1. Validate preconditions.
	if err := app.RescueHandler.ValidatePreconditions(ctx, projectDir); err != nil {
		return fmt.Errorf("rescue preconditions: %w", err)
	}

	// 2. Run rescue analysis.
	analysis, err := app.RescueHandler.Rescue(ctx, projectDir, nil, true)
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
