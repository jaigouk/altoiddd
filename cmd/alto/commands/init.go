package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cbterm "github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"

	"github.com/alto-cli/alto/internal/bootstrap/domain"
	"github.com/alto-cli/alto/internal/composition"
)

// NewInitCmd creates the "alto init" command.
func NewInitCmd(app *composition.App) *cobra.Command {
	var (
		existing     bool
		dryRun       bool
		yes          bool
		forceBranch  bool
		noCommit     bool
		withScaffold bool
		force        bool
		projectName  string
		ticketPrefix string
		issueTracker string
		boundedCtxs  []string
		primaryTool  string
		noHooks      bool
		forceHooks   bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Bootstrap a new project from a README idea",
		Long: `Bootstrap a new project from a README idea.

Auto-detects whether the current directory contains an existing project
and chooses the appropriate path. Use --existing to force rescue mode.

Use --with-scaffold to extract the embedded alto-scaffold/ workflow scaffold into
the current directory. The five --project-name / --ticket-prefix /
--issue-tracker / --bounded-contexts / --primary-tool flags supply the
template parameters substituted into the scaffold files. Use --force to
overwrite an existing alto-scaffold/ tree (a per-file [OVERWRITE] preview is
emitted before any write occurs).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// --with-scaffold short-circuits the rescue / detect path; it is
			// strictly an "extract the embedded alto-scaffold/" operation. The
			// --existing rescue integration is out of scope for this ticket.
			if withScaffold {
				if noHooks && forceHooks {
					return fmt.Errorf("--no-hooks and --force-hooks are mutually exclusive")
				}
				return runWithScaffold(cmd, app, withScaffoldArgs{
					projectName:  projectName,
					ticketPrefix: ticketPrefix,
					issueTracker: issueTracker,
					boundedCtxs:  boundedCtxs,
					primaryTool:  primaryTool,
					force:        force,
					noHooks:      noHooks,
					forceHooks:   forceHooks,
				})
			}

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
				if !yes {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Found docs/ folder but no source code.")
					_, _ = fmt.Fprint(cmd.OutOrStdout(), "Treat as existing project? [y/N] ")
					scanner := bufio.NewScanner(cmd.InOrStdin())
					if scanner.Scan() {
						answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
						if answer == "y" || answer == "yes" {
							return runRescue(cmd, app, dryRun, forceBranch)
						}
					}
				} else {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Proceeding with fresh init (-y).")
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

	cmd.Flags().BoolVar(&withScaffold, "with-scaffold", false, "Extract the embedded alto-scaffold/ workflow scaffold into the current directory")
	cmd.Flags().BoolVar(&force, "force", false, "With --with-scaffold: overwrite an existing alto-scaffold/ tree (per-file [OVERWRITE] preview emitted)")
	cmd.Flags().StringVar(&projectName, "project-name", "", "With --with-scaffold: project name (substituted into template parameter)")
	cmd.Flags().StringVar(&ticketPrefix, "ticket-prefix", "", "With --with-scaffold: ticket prefix (must end with '-'), e.g. 'demo-'")
	cmd.Flags().StringVar(&issueTracker, "issue-tracker", "beads", "With --with-scaffold: one of {beads, github, linear}")
	cmd.Flags().StringSliceVar(&boundedCtxs, "bounded-contexts", nil, "With --with-scaffold: comma-separated PascalCase context names, e.g. Orders,Catalog")
	cmd.Flags().StringVar(&primaryTool, "primary-tool", "claude", "With --with-scaffold: one of {claude, opencode}")
	cmd.Flags().BoolVar(&noHooks, "no-hooks", false, "With --with-scaffold: skip writing .beads/hooks/post-close (advanced opt-out)")
	cmd.Flags().BoolVar(&forceHooks, "force-hooks", false, "With --with-scaffold: overwrite an existing .beads/hooks/post-close hook")

	return cmd
}

// withScaffoldArgs bundles the --with-scaffold flag values into a single
// parameter object — keeps the run function under the linter's argument
// count limit and documents intent at the call site.
type withScaffoldArgs struct {
	projectName  string
	ticketPrefix string
	issueTracker string
	boundedCtxs  []string
	primaryTool  string
	force        bool
	noHooks      bool
	forceHooks   bool
}

// runWithScaffold builds the ScaffoldParams VO from CLI flags and asks
// the BootstrapHandler to extract the embedded alto-scaffold/ tree. The handler
// chains into the OpenCode adapter when --primary-tool=opencode.
func runWithScaffold(cmd *cobra.Command, app *composition.App, args withScaffoldArgs) error {
	params, err := domain.NewScaffoldParams(
		args.projectName,
		args.ticketPrefix,
		args.issueTracker,
		args.boundedCtxs,
		args.primaryTool,
	)
	if err != nil {
		return fmt.Errorf("scaffold parameters: %w", err)
	}
	if args.noHooks {
		params = params.WithIncludeHooks(false)
	}
	app.BootstrapHandler.SetForceHooks(args.forceHooks)

	if err := app.BootstrapHandler.WriteScaffold(cmd.Context(), ".", params, args.force); err != nil {
		return fmt.Errorf("writing scaffold: %w", err)
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Scaffold extracted to alto-scaffold/")
	return nil
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
	projectName := resolveProjectName(projectDir)
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
		scanner := bufio.NewScanner(cmd.InOrStdin())
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

	// 6. Launch guide flow — detect TTY to decide TUI mode.
	noTUI := !cbterm.IsTerminal(os.Stdin.Fd()) || os.Getenv("ALTO_NO_TUI") == "1"
	return runGuide(cmd.Context(), app, noTUI, false, false, false)
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

// resolveProjectName extracts the project name from the first # heading in
// README.md or README. Falls back to filepath.Base of the absolute project dir.
func resolveProjectName(projectDir string) string {
	for _, candidate := range []string{"README.md", "README"} {
		content, err := os.ReadFile(filepath.Join(projectDir, candidate))
		if err != nil {
			continue
		}

		for _, line := range strings.Split(string(content), "\n") {
			after, ok := strings.CutPrefix(line, "# ")
			if !ok {
				continue
			}

			if name := strings.TrimSpace(after); name != "" {
				return stripSubtitle(name)
			}
		}
	}

	return filepath.Base(mustAbs(projectDir))
}

// stripSubtitle removes a tagline/subtitle after a separator in a heading.
// Separators checked in order: " — ", " - ", ": ", " | ". Earliest match wins.
func stripSubtitle(heading string) string {
	for _, sep := range []string{" — ", " - ", ": ", " | "} {
		if idx := strings.Index(heading, sep); idx > 0 {
			return heading[:idx]
		}
	}

	return heading
}
