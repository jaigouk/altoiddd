package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alto-cli/alto/internal/composition"
	fitnessapp "github.com/alto-cli/alto/internal/fitness/application"
	fitnessdomain "github.com/alto-cli/alto/internal/fitness/domain"
	fitnessinfra "github.com/alto-cli/alto/internal/fitness/infrastructure"
	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
	"github.com/alto-cli/alto/internal/shared/infrastructure/stack"
)

// NewFitnessCmd creates the "alto fitness" command group.
func NewFitnessCmd(app *composition.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fitness",
		Short: "Architecture fitness testing commands",
		Long:  "Commands for generating and running architecture fitness tests.",
	}

	cmd.AddCommand(newFitnessGenerateCmd(app))

	return cmd
}

func newFitnessGenerateCmd(app *composition.App) *cobra.Command {
	var (
		preview    bool
		brownfield bool
		dir        string
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate architecture fitness test configuration",
		Long: `Generate arch-go.yml configuration from the bounded context map.

Reads .alto/bounded_context_map.yaml and generates arch-go.yml for
architecture fitness testing.

Examples:
  alto fitness generate              # Generate with user confirmation
  alto fitness generate --preview    # Show what would be generated
  alto fitness generate --brownfield # Use 80% compliance threshold`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFitnessGenerate(cmd.Context(), app, dir, preview, brownfield)
		},
	}

	cmd.Flags().BoolVar(&preview, "preview", false, "Show what would be generated without writing")
	cmd.Flags().BoolVar(&brownfield, "brownfield", false, "Use 80% compliance threshold for existing projects")
	cmd.Flags().StringVar(&dir, "dir", ".", "Project directory (default: current directory)")

	return cmd
}

func runFitnessGenerate(ctx context.Context, app *composition.App, projectDir string, preview, brownfield bool) error {
	// Resolve absolute path
	absDir, err := filepath.Abs(projectDir)
	if err != nil {
		return fmt.Errorf("resolving project directory: %w", err)
	}

	// Check for bounded context map
	bcMapPath := filepath.Join(absDir, ".alto", "bounded_context_map.yaml")
	if _, statErr := os.Stat(bcMapPath); os.IsNotExist(statErr) {
		return fmt.Errorf("bounded_context_map.yaml not found at %s\nRun 'alto guide' first to generate it", bcMapPath)
	}

	// Detect stack profile
	profile := DetectStackProfile(absDir)
	if !profile.FitnessAvailable() {
		return fmt.Errorf("fitness tests not available for %s stack\nSupported: Go (go.mod), Python (pyproject.toml)", profile.StackID())
	}

	// Load domain model from bounded context map
	loadResult, err := LoadDomainModelFromBCMap(ctx, bcMapPath)
	if err != nil {
		return fmt.Errorf("loading bounded context map: %w", err)
	}

	// Build preview options
	var opts *fitnessapp.BuildPreviewOptions
	if brownfield {
		opts = &fitnessapp.BuildPreviewOptions{Threshold: 80}
	}

	// Get project name from model
	projectName := loadResult.Model.ModelID()

	// Build preview - pass BC map for Go stack to preserve module_path values
	fitnessPreview, err := app.FitnessGenerationHandler.BuildPreviewWithBCMap(
		loadResult.Model, projectName, profile, loadResult.BCMap, opts)
	if err != nil {
		return fmt.Errorf("building fitness preview: %w", err)
	}

	if fitnessPreview == nil {
		return fmt.Errorf("fitness tests not available for this stack")
	}

	// Show preview
	fmt.Println("=== Architecture Fitness Configuration Preview ===")
	fmt.Println()
	fmt.Println(fitnessPreview.Summary)
	fmt.Println()
	fmt.Println("--- arch-go.yml ---")
	fmt.Println(fitnessPreview.YAMLContent)

	if preview {
		fmt.Println("(Preview mode - no files written)")
		return nil
	}

	// Confirm with user
	fmt.Print("Write arch-go.yml? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("reading confirmation: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}

	// Write files
	if err := app.FitnessGenerationHandler.WriteFiles(ctx, fitnessPreview, absDir); err != nil {
		return fmt.Errorf("writing fitness configuration: %w", err)
	}

	fmt.Println("Wrote arch-go.yml")
	fmt.Println()
	fmt.Println("Run 'arch-go' to validate architecture fitness.")

	return nil
}

// DetectStackProfile detects the stack profile from project files.
// Delegates to stack.DetectProfile for shared implementation.
func DetectStackProfile(projectDir string) vo.StackProfile {
	return stack.DetectProfile(projectDir)
}

// LoadResult contains both the DomainModel and the raw BoundedContextMap.
// The BC map preserves original module_path values needed for arch-go generation.
type LoadResult struct {
	Model *ddd.DomainModel
	BCMap *fitnessdomain.BoundedContextMap
}

// LoadDomainModelFromBCMap parses a bounded context map YAML and creates a minimal DomainModel.
// The model is not finalized (Finalize() not called) since BuildPreview only needs BoundedContexts.
// Also returns the parsed BC map to preserve module_path values for arch-go rendering.
func LoadDomainModelFromBCMap(ctx context.Context, bcMapPath string) (*LoadResult, error) {
	parser := fitnessinfra.NewBoundedContextMapParser()
	bcMap, err := parser.Parse(ctx, bcMapPath)
	if err != nil {
		return nil, fmt.Errorf("parsing bounded context map: %w", err)
	}

	// Create minimal domain model
	model := ddd.NewDomainModel(bcMap.ProjectName())

	// Add bounded contexts with classifications
	for _, entry := range bcMap.Contexts() {
		classification := entry.Classification()
		bc := vo.NewDomainBoundedContext(
			entry.Name(),
			"",  // responsibility not needed for fitness
			nil, // key domain objects not needed
			&classification,
			"", // rationale not needed
		)
		if err := model.AddBoundedContext(bc); err != nil {
			return nil, fmt.Errorf("adding bounded context %s: %w", entry.Name(), err)
		}
	}

	return &LoadResult{Model: model, BCMap: bcMap}, nil
}
