package application

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/alty-cli/alty/internal/fitness/domain"
	sharedapp "github.com/alty-cli/alty/internal/shared/application"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// FitnessPreview holds the preview data for fitness test generation.
type FitnessPreview struct {
	Suite       *domain.FitnessTestSuite
	TomlContent string
	TestContent string
	Summary     string
}

// FitnessGenerationHandler orchestrates fitness test suite generation
// using the preview-before-action pattern.
type FitnessGenerationHandler struct {
	fileWriter sharedapp.FileWriter
	publisher  sharedapp.EventPublisher
}

// NewFitnessGenerationHandler creates a new FitnessGenerationHandler.
func NewFitnessGenerationHandler(fileWriter sharedapp.FileWriter, publisher sharedapp.EventPublisher) *FitnessGenerationHandler {
	return &FitnessGenerationHandler{
		fileWriter: fileWriter,
		publisher:  publisher,
	}
}

// BuildPreview generates a fitness test suite preview without writing files.
// Returns nil preview (no error) when the profile does not support fitness tests.
func (h *FitnessGenerationHandler) BuildPreview(
	model *ddd.DomainModel,
	projectName string,
	profile vo.StackProfile,
) (*FitnessPreview, error) {
	// If a profile is provided and fitness is not available, return nil.
	if profile != nil && !profile.FitnessAvailable() {
		return nil, nil
	}

	bcs := model.BoundedContexts()
	if len(bcs) == 0 {
		return nil, fmt.Errorf("no bounded contexts found in model: %w",
			domainerrors.ErrInvariantViolation)
	}

	// Determine root package
	rootPackage := projectName
	if profile != nil {
		rootPackage = profile.ToRootPackage(projectName)
	}

	// Create the suite and generate contracts
	suite := domain.NewFitnessTestSuite(rootPackage)
	inputs := make([]domain.BoundedContextInput, 0, len(bcs))
	for _, bc := range bcs {
		inputs = append(inputs, domain.BoundedContextInput{
			Name:           bc.Name(),
			Responsibility: bc.Responsibility(),
			Classification: bc.Classification(),
		})
	}

	if err := suite.GenerateContracts(inputs); err != nil {
		return nil, fmt.Errorf("generate contracts: %w", err)
	}

	tomlContent, err := suite.RenderImportLinterTOML()
	if err != nil {
		return nil, fmt.Errorf("render import linter TOML: %w", err)
	}

	testContent, err := suite.RenderPytestarchTests()
	if err != nil {
		return nil, fmt.Errorf("render pytestarch tests: %w", err)
	}

	summary, err := suite.Preview()
	if err != nil {
		return nil, fmt.Errorf("generate preview: %w", err)
	}

	// Add BC names to summary
	var bcNames []string
	for _, bc := range bcs {
		bcNames = append(bcNames, bc.Name())
	}
	summary += "\nBounded contexts: " + strings.Join(bcNames, ", ")

	return &FitnessPreview{
		Suite:       suite,
		TomlContent: tomlContent,
		TestContent: testContent,
		Summary:     summary,
	}, nil
}

// WriteFiles writes the preview content to disk without approving the suite.
func (h *FitnessGenerationHandler) WriteFiles(
	ctx context.Context,
	preview *FitnessPreview,
	projectDir string,
) error {
	return h.writePreviewFiles(ctx, preview, projectDir)
}

// ApproveAndWrite approves the suite (emitting domain events) and writes files.
func (h *FitnessGenerationHandler) ApproveAndWrite(
	ctx context.Context,
	preview *FitnessPreview,
	projectDir string,
) error {
	if err := preview.Suite.Approve(); err != nil {
		return fmt.Errorf("approve suite: %w", err)
	}
	if err := h.writePreviewFiles(ctx, preview, projectDir); err != nil {
		return err
	}
	for _, event := range preview.Suite.Events() {
		_ = h.publisher.Publish(ctx, event)
	}
	return nil
}

func (h *FitnessGenerationHandler) writePreviewFiles(
	ctx context.Context,
	preview *FitnessPreview,
	projectDir string,
) error {
	tomlPath := filepath.Join(projectDir, ".importlinter")
	if err := h.fileWriter.WriteFile(ctx, tomlPath, preview.TomlContent); err != nil {
		return fmt.Errorf("write import linter config: %w", err)
	}

	testPath := filepath.Join(projectDir, "tests", "architecture", "test_fitness.py")
	if err := h.fileWriter.WriteFile(ctx, testPath, preview.TestContent); err != nil {
		return fmt.Errorf("write fitness tests: %w", err)
	}

	return nil
}
