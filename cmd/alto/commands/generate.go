package commands

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/alto-cli/alto/internal/composition"
	"github.com/alto-cli/alto/internal/shared/infrastructure/stack"
	ttdomain "github.com/alto-cli/alto/internal/tooltranslation/domain"
)

// NewGenerateCmd creates the "alto generate" command group.
func NewGenerateCmd(app *composition.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate project artifacts",
		Long:  "Generate various project artifacts (DDD docs, fitness tests, tickets, configs).",
	}

	cmd.AddCommand(
		newGenerateArtifactsCmd(app),
		newGenerateFitnessCmd(app),
		newGenerateTicketsCmd(app),
		newGenerateConfigsCmd(app),
	)

	return cmd
}

func newGenerateArtifactsCmd(app *composition.App) *cobra.Command {
	return &cobra.Command{
		Use:   "artifacts",
		Short: "Generate DDD artifacts (PRD, DDD.md, ARCHITECTURE.md)",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = app.ArtifactGenerationHandler
			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintln(w, "Artifact generation requires a completed discovery session.")
			_, _ = fmt.Fprintln(w, "Run 'alto guide' first.")
			return nil
		},
	}
}

func newGenerateFitnessCmd(app *composition.App) *cobra.Command {
	var fromDocs bool
	var docsDir string

	cmd := &cobra.Command{
		Use:   "fitness",
		Short: "Generate architecture fitness tests",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !fromDocs {
				return fmt.Errorf("requires a domain model. Use --from-docs or run \"alto guide\" first")
			}

			result, err := app.DocImportHandler.Import(cmd.Context(), docsDir)
			if err != nil {
				return fmt.Errorf("importing docs: %w", err)
			}
			model := result.Model()

			profile := stack.DetectProfile("")
			projectName := "."
			cwd, absErr := filepath.Abs(".")
			if absErr == nil {
				projectName = filepath.Base(cwd)
			}

			preview, err := app.FitnessGenerationHandler.BuildPreview(model, projectName, profile, nil)
			if err != nil {
				return fmt.Errorf("building fitness preview: %w", err)
			}

			w := cmd.OutOrStdout()
			if preview == nil {
				_, _ = fmt.Fprintln(w, "Fitness tests not available for the detected stack profile.")
				return nil
			}
			_, _ = fmt.Fprintln(w, preview.Summary)
			return nil
		},
	}

	cmd.Flags().BoolVar(&fromDocs, "from-docs", false, "Import model from docs/DDD.md instead of discovery")
	cmd.Flags().StringVar(&docsDir, "docs-dir", "docs", "Directory containing DDD.md")

	return cmd
}

func newGenerateTicketsCmd(app *composition.App) *cobra.Command {
	var fromDocs bool
	var docsDir string

	cmd := &cobra.Command{
		Use:   "tickets",
		Short: "Generate dependency-ordered beads tickets from DDD",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !fromDocs {
				return fmt.Errorf("requires a domain model. Use --from-docs or run \"alto guide\" first")
			}

			result, err := app.DocImportHandler.Import(cmd.Context(), docsDir)
			if err != nil {
				return fmt.Errorf("importing docs: %w", err)
			}
			model := result.Model()

			profile := stack.DetectProfile("")

			preview, err := app.TicketGenerationHandler.BuildPreview(model, profile)
			if err != nil {
				return fmt.Errorf("building ticket preview: %w", err)
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), preview.Summary)
			return nil
		},
	}

	cmd.Flags().BoolVar(&fromDocs, "from-docs", false, "Import model from docs/DDD.md instead of discovery")
	cmd.Flags().StringVar(&docsDir, "docs-dir", "docs", "Directory containing DDD.md")

	return cmd
}

func newGenerateConfigsCmd(app *composition.App) *cobra.Command {
	var fromDocs bool
	var docsDir string

	cmd := &cobra.Command{
		Use:   "configs",
		Short: "Generate tool-native configurations for AI coding tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !fromDocs {
				return fmt.Errorf("requires a domain model. Use --from-docs or run \"alto guide\" first")
			}

			result, err := app.DocImportHandler.Import(cmd.Context(), docsDir)
			if err != nil {
				return fmt.Errorf("importing docs: %w", err)
			}
			model := result.Model()

			profile := stack.DetectProfile("")

			// Detect installed tools, fall back to all supported tools
			tools := detectSupportedTools(app)

			preview, err := app.ConfigGenerationHandler.BuildPreview(model, tools, profile)
			if err != nil {
				return fmt.Errorf("building config preview: %w", err)
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), preview.Summary)
			return nil
		},
	}

	cmd.Flags().BoolVar(&fromDocs, "from-docs", false, "Import model from docs/DDD.md instead of discovery")
	cmd.Flags().StringVar(&docsDir, "docs-dir", "docs", "Directory containing DDD.md")

	return cmd
}

// detectSupportedTools tries to detect installed AI tools, falling back to all supported tools.
func detectSupportedTools(app *composition.App) []ttdomain.SupportedTool {
	detectionResult, err := app.DetectionHandler.Detect(".")
	if err == nil {
		detected := detectionResult.DetectedTools()
		if len(detected) > 0 {
			tools := make([]ttdomain.SupportedTool, 0, len(detected))
			for _, dt := range detected {
				tools = append(tools, ttdomain.SupportedTool(dt.Name()))
			}
			return tools
		}
	}
	return ttdomain.AllSupportedTools()
}
