package infrastructure

import (
	"context"
	"os"
	"path/filepath"

	rescueapp "github.com/alto-cli/alto/internal/rescue/application"
	rescuedomain "github.com/alto-cli/alto/internal/rescue/domain"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

var (
	docTargets = []string{
		"docs/PRD.md",
		"docs/DDD.md",
		"docs/ARCHITECTURE.md",
		"AGENTS.md",
	}

	altoConfigTargets = []string{
		".claude/CLAUDE.md",
		".beads/issues.jsonl",
	}

	defaultStructureTargets = []string{
		"src/domain/",
		"src/application/",
		"src/infrastructure/",
	}

	defaultManifest = "pyproject.toml"
)

// ProjectScanner scans an existing project directory and returns a frozen
// ProjectScan value object describing its current state.
type ProjectScanner struct{}

// Compile-time interface check.
var _ rescueapp.ProjectScan = (*ProjectScanner)(nil)

// Scan scans a project directory and returns a frozen snapshot.
func (s *ProjectScanner) Scan(
	_ context.Context,
	projectDir string,
	profile vo.StackProfile,
) (rescuedomain.ProjectScan, error) {
	var existingDocs []string
	for _, doc := range docTargets {
		if fileExists(filepath.Join(projectDir, doc)) {
			existingDocs = append(existingDocs, doc)
		}
	}

	// Build config targets
	manifest := defaultManifest
	if profile != nil {
		manifest = profile.ProjectManifest()
	}
	configTargets := make([]string, len(altoConfigTargets))
	copy(configTargets, altoConfigTargets)
	if manifest != "" {
		configTargets = append(configTargets, manifest)
	}

	var existingConfigs []string
	for _, cfg := range configTargets {
		if fileExists(filepath.Join(projectDir, cfg)) {
			existingConfigs = append(existingConfigs, cfg)
		}
	}

	// Build structure targets
	structureTargets := defaultStructureTargets
	if profile != nil && len(profile.SourceLayout()) > 0 {
		structureTargets = profile.SourceLayout()
	}

	var existingStructure []string
	for _, dir := range structureTargets {
		if dirExists(filepath.Join(projectDir, dir)) {
			existingStructure = append(existingStructure, dir)
		}
	}

	hasKnowledgeDir := dirExists(filepath.Join(projectDir, ".alto", "knowledge"))
	hasAgentsMD := fileExists(filepath.Join(projectDir, "AGENTS.md"))
	hasGit := pathExists(filepath.Join(projectDir, ".git"))
	hasAltoConfig := fileExists(filepath.Join(projectDir, ".alto", "config.toml"))
	hasMaintenanceDir := dirExists(filepath.Join(projectDir, ".alto", "maintenance"))

	return rescuedomain.NewProjectScan(
		projectDir,
		existingDocs,
		existingConfigs,
		existingStructure,
		hasKnowledgeDir,
		hasAgentsMD,
		hasGit,
		hasAltoConfig,
		hasMaintenanceDir,
	), nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
