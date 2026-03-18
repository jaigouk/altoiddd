package application

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/alto-cli/alto/internal/fitness/domain"
	sharedapp "github.com/alto-cli/alto/internal/shared/application"
	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	"github.com/alto-cli/alto/internal/shared/domain/stringutil"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// FitnessPreview holds the preview data for fitness test generation.
type FitnessPreview struct {
	Suite       *domain.FitnessTestSuite
	TomlContent string // Python: import-linter config
	TestContent string // Python: pytestarch tests
	YAMLContent string // Go: arch-go.yml config
	StackID     string // Stack identifier (e.g., "go-mod", "python-uv")
	Summary     string
	warnings    []string
}

// Warnings returns a defensive copy of generation warnings.
func (p *FitnessPreview) Warnings() []string {
	out := make([]string, len(p.warnings))
	copy(out, p.warnings)
	return out
}

// BuildPreviewOptions configures BuildPreview behavior.
type BuildPreviewOptions struct {
	// Threshold sets the compliance percentage (0 = default 100 for greenfield, 80 for brownfield).
	Threshold int
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
// Pass nil for opts to use defaults (100% threshold for greenfield).
func (h *FitnessGenerationHandler) BuildPreview(
	model *ddd.DomainModel,
	projectName string,
	profile vo.StackProfile,
	opts *BuildPreviewOptions,
) (*FitnessPreview, error) {
	return h.BuildPreviewWithBCMap(model, projectName, profile, nil, opts)
}

// BuildPreviewWithBCMap generates a fitness test suite preview without writing files.
// For Go stack, if bcMap is provided, it's used directly to preserve original module_path values.
// Returns nil preview (no error) when the profile does not support fitness tests.
// Pass nil for opts to use defaults (100% threshold for greenfield).
func (h *FitnessGenerationHandler) BuildPreviewWithBCMap(
	model *ddd.DomainModel,
	projectName string,
	profile vo.StackProfile,
	bcMap *domain.BoundedContextMap,
	opts *BuildPreviewOptions,
) (*FitnessPreview, error) {
	// If a profile is provided and fitness is not available, return nil.
	if profile != nil && !profile.FitnessAvailable() {
		return nil, nil
	}

	if model.IsEmpty() {
		return nil, fmt.Errorf("model is empty, nothing to generate; run 'alto guide' or 'alto import' first: %w",
			domainerrors.ErrInvariantViolation)
	}

	bcs := model.BoundedContexts()
	if len(bcs) == 0 {
		return &FitnessPreview{
			Summary:  "Partial generation — see warnings",
			warnings: []string{"no bounded contexts found in model, skipping fitness test generation"},
		}, nil
	}

	// Apply defaults
	threshold := 100
	if opts != nil && opts.Threshold > 0 {
		threshold = opts.Threshold
	}

	// Determine root package and stack
	rootPackage := projectName
	stackID := ""
	if profile != nil {
		rootPackage = profile.ToRootPackage(projectName)
		stackID = profile.StackID()
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

	// Add BC names to summary
	var bcNames []string
	for _, bc := range bcs {
		bcNames = append(bcNames, bc.Name())
	}

	// Branch based on stack
	if stackID == "go-mod" {
		// Go stack: generate arch-go.yml
		// Use provided bcMap if available, otherwise convert from model
		var archBCMap *domain.BoundedContextMap
		if bcMap != nil {
			archBCMap = bcMap
		} else {
			archBCMap = convertToBoundedContextMap(model, rootPackage)
		}
		yamlContent, err := suite.RenderArchGoYAML(archBCMap, threshold)
		if err != nil {
			return nil, fmt.Errorf("render arch-go YAML: %w", err)
		}

		summary, err := suite.Preview()
		if err != nil {
			return nil, fmt.Errorf("generate preview: %w", err)
		}
		summary += "\nBounded contexts: " + strings.Join(bcNames, ", ")

		return &FitnessPreview{
			Suite:       suite,
			YAMLContent: yamlContent,
			StackID:     stackID,
			Summary:     summary,
		}, nil
	}

	// Python stack (default): generate import-linter and pytestarch
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
	summary += "\nBounded contexts: " + strings.Join(bcNames, ", ")

	return &FitnessPreview{
		Suite:       suite,
		TomlContent: tomlContent,
		TestContent: testContent,
		StackID:     stackID,
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
	// Branch based on stack
	if preview.StackID == "go-mod" {
		// Go stack: write arch-go.yml
		yamlPath := filepath.Join(projectDir, "arch-go.yml")
		if err := h.fileWriter.WriteFile(ctx, yamlPath, preview.YAMLContent); err != nil {
			return fmt.Errorf("write arch-go config: %w", err)
		}
		return nil
	}

	// Python stack (default): write import-linter and pytestarch files
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

// convertToBoundedContextMap converts a DomainModel to a BoundedContextMap for arch-go.
func convertToBoundedContextMap(model *ddd.DomainModel, rootPackage string) *domain.BoundedContextMap {
	bcs := model.BoundedContexts()
	rels := model.ContextRelationships()

	// Build relationship lookup: downstream -> list of upstreams
	// vo.ContextRelationship has Upstream() and Downstream()
	relMap := make(map[string][]vo.ContextRelationship)
	for _, rel := range rels {
		relMap[rel.Downstream()] = append(relMap[rel.Downstream()], rel)
	}

	// Convert bounded contexts
	entries := make([]domain.BoundedContextEntry, 0, len(bcs))
	for _, bc := range bcs {
		modulePath := stringutil.ToSnakeCase(bc.Name())
		classification := vo.SubdomainGeneric // default
		if bc.Classification() != nil {
			classification = *bc.Classification()
		}

		// Convert relationships for this context
		var contextRels []domain.ContextRelationship
		for _, rel := range relMap[bc.Name()] {
			// This context is downstream, so the relationship target is upstream
			pattern := mapIntegrationPattern(rel.IntegrationPattern())
			contextRels = append(contextRels, domain.NewContextRelationship(
				rel.Upstream(),
				domain.RelationshipUpstream,
				pattern,
			))
		}

		entry := domain.NewBoundedContextEntry(
			bc.Name(),
			modulePath,
			classification,
			[]string{"domain", "application", "infrastructure"},
			contextRels,
		)
		entries = append(entries, entry)
	}

	bcMap := domain.NewBoundedContextMap("", rootPackage, entries)
	return &bcMap
}

// mapIntegrationPattern maps a string integration pattern to domain constant.
func mapIntegrationPattern(pattern string) domain.RelationshipPattern {
	lower := strings.ToLower(pattern)
	switch {
	case strings.Contains(lower, "event"):
		return domain.PatternDomainEvent
	case strings.Contains(lower, "shared"):
		return domain.PatternSharedKernel
	case strings.Contains(lower, "acl"):
		return domain.PatternACL
	case strings.Contains(lower, "open"):
		return domain.PatternOpenHost
	default:
		return domain.PatternDomainEvent
	}
}
