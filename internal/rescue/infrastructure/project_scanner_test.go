package infrastructure_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	rescueapp "github.com/alty-cli/alty/internal/rescue/application"
	"github.com/alty-cli/alty/internal/rescue/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectScannerImplementsPort(t *testing.T) {
	t.Parallel()
	var _ rescueapp.ProjectScan = (*infrastructure.ProjectScanner)(nil)
}

// ---------------------------------------------------------------------------
// Empty project
// ---------------------------------------------------------------------------

func TestScanEmptyProject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	scanner := &infrastructure.ProjectScanner{}
	scan, err := scanner.Scan(context.Background(), dir, nil)
	require.NoError(t, err)

	assert.Equal(t, dir, scan.ProjectDir())
	assert.Len(t, scan.ExistingDocs(), 0)
	assert.Len(t, scan.ExistingConfigs(), 0)
	assert.Len(t, scan.ExistingStructure(), 0)
	assert.False(t, scan.HasKnowledgeDir())
	assert.False(t, scan.HasAgentsMD())
	assert.False(t, scan.HasGit())
}

// ---------------------------------------------------------------------------
// Docs
// ---------------------------------------------------------------------------

func TestScanFindsPRD(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docs", "PRD.md"), []byte("# PRD"), 0o644))
	scanner := &infrastructure.ProjectScanner{}
	scan, err := scanner.Scan(context.Background(), dir, nil)
	require.NoError(t, err)
	assert.Contains(t, scan.ExistingDocs(), "docs/PRD.md")
}

func TestScanFindsDDD(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docs", "DDD.md"), []byte("# DDD"), 0o644))
	scanner := &infrastructure.ProjectScanner{}
	scan, err := scanner.Scan(context.Background(), dir, nil)
	require.NoError(t, err)
	assert.Contains(t, scan.ExistingDocs(), "docs/DDD.md")
}

func TestScanFindsArchitecture(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docs", "ARCHITECTURE.md"), []byte("# Arch"), 0o644))
	scanner := &infrastructure.ProjectScanner{}
	scan, err := scanner.Scan(context.Background(), dir, nil)
	require.NoError(t, err)
	assert.Contains(t, scan.ExistingDocs(), "docs/ARCHITECTURE.md")
}

func TestScanFindsAgentsMD(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# AGENTS"), 0o644))
	scanner := &infrastructure.ProjectScanner{}
	scan, err := scanner.Scan(context.Background(), dir, nil)
	require.NoError(t, err)
	assert.Contains(t, scan.ExistingDocs(), "AGENTS.md")
	assert.True(t, scan.HasAgentsMD())
}

// ---------------------------------------------------------------------------
// Configs
// ---------------------------------------------------------------------------

func TestScanFindsClaudeMD(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".claude", "CLAUDE.md"), []byte("# Claude"), 0o644))
	scanner := &infrastructure.ProjectScanner{}
	scan, err := scanner.Scan(context.Background(), dir, nil)
	require.NoError(t, err)
	assert.Contains(t, scan.ExistingConfigs(), ".claude/CLAUDE.md")
}

func TestScanFindsBeads(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".beads"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".beads", "issues.jsonl"), []byte("{}"), 0o644))
	scanner := &infrastructure.ProjectScanner{}
	scan, err := scanner.Scan(context.Background(), dir, nil)
	require.NoError(t, err)
	assert.Contains(t, scan.ExistingConfigs(), ".beads/issues.jsonl")
}

func TestScanFindsPyproject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[project]"), 0o644))
	scanner := &infrastructure.ProjectScanner{}
	scan, err := scanner.Scan(context.Background(), dir, nil)
	require.NoError(t, err)
	assert.Contains(t, scan.ExistingConfigs(), "pyproject.toml")
}

// ---------------------------------------------------------------------------
// Structure
// ---------------------------------------------------------------------------

func TestScanFindsDomainDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "src", "domain"), 0o755))
	scanner := &infrastructure.ProjectScanner{}
	scan, err := scanner.Scan(context.Background(), dir, nil)
	require.NoError(t, err)
	assert.Contains(t, scan.ExistingStructure(), "src/domain/")
}

func TestScanFindsApplicationDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "src", "application"), 0o755))
	scanner := &infrastructure.ProjectScanner{}
	scan, err := scanner.Scan(context.Background(), dir, nil)
	require.NoError(t, err)
	assert.Contains(t, scan.ExistingStructure(), "src/application/")
}

func TestScanFindsInfrastructureDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "src", "infrastructure"), 0o755))
	scanner := &infrastructure.ProjectScanner{}
	scan, err := scanner.Scan(context.Background(), dir, nil)
	require.NoError(t, err)
	assert.Contains(t, scan.ExistingStructure(), "src/infrastructure/")
}

// ---------------------------------------------------------------------------
// Special dirs
// ---------------------------------------------------------------------------

func TestScanFindsKnowledgeDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".alty", "knowledge"), 0o755))
	scanner := &infrastructure.ProjectScanner{}
	scan, err := scanner.Scan(context.Background(), dir, nil)
	require.NoError(t, err)
	assert.True(t, scan.HasKnowledgeDir())
}

func TestScanFindsGitDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))
	scanner := &infrastructure.ProjectScanner{}
	scan, err := scanner.Scan(context.Background(), dir, nil)
	require.NoError(t, err)
	assert.True(t, scan.HasGit())
}

func TestScanFindsAltyConfig(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".alty"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".alty", "config.toml"), []byte("[alty]"), 0o644))
	scanner := &infrastructure.ProjectScanner{}
	scan, err := scanner.Scan(context.Background(), dir, nil)
	require.NoError(t, err)
	assert.True(t, scan.HasAltyConfig())
}

func TestScanNoAltyConfig(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	scanner := &infrastructure.ProjectScanner{}
	scan, err := scanner.Scan(context.Background(), dir, nil)
	require.NoError(t, err)
	assert.False(t, scan.HasAltyConfig())
}

func TestScanFindsMaintenanceDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".alty", "maintenance"), 0o755))
	scanner := &infrastructure.ProjectScanner{}
	scan, err := scanner.Scan(context.Background(), dir, nil)
	require.NoError(t, err)
	assert.True(t, scan.HasMaintenanceDir())
}

func TestScanNoMaintenanceDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	scanner := &infrastructure.ProjectScanner{}
	scan, err := scanner.Scan(context.Background(), dir, nil)
	require.NoError(t, err)
	assert.False(t, scan.HasMaintenanceDir())
}

// ---------------------------------------------------------------------------
// Full project
// ---------------------------------------------------------------------------

func TestScanFullProject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Docs
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs"), 0o755))
	for _, doc := range []string{"PRD.md", "DDD.md", "ARCHITECTURE.md"} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "docs", doc), []byte("# "+doc), 0o644))
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# AGENTS"), 0o644))

	// Configs
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".claude", "CLAUDE.md"), []byte("# Claude"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".beads"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".beads", "issues.jsonl"), []byte("{}"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[project]"), 0o644))

	// Structure
	for _, d := range []string{"src/domain", "src/application", "src/infrastructure"} {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, d), 0o755))
	}

	// Special dirs
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".alty", "knowledge"), 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o755))

	scanner := &infrastructure.ProjectScanner{}
	scan, err := scanner.Scan(context.Background(), dir, nil)
	require.NoError(t, err)

	assert.Len(t, scan.ExistingDocs(), 4)
	assert.Len(t, scan.ExistingConfigs(), 3)
	assert.Len(t, scan.ExistingStructure(), 3)
	assert.True(t, scan.HasKnowledgeDir())
	assert.True(t, scan.HasAgentsMD())
	assert.True(t, scan.HasGit())
}
