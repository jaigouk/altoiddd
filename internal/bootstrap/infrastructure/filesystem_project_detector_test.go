package infrastructure

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bootstrapapp "github.com/alto-cli/alto/internal/bootstrap/application"
)

// Compile-time check that FileSystemProjectDetector satisfies the port.
var _ bootstrapapp.ProjectDetector = (*FileSystemProjectDetector)(nil)

func TestFileSystemProjectDetector_Detect_WhenEmptyDirectory_ExpectNewProject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	detector := &FileSystemProjectDetector{}
	result, err := detector.Detect(dir)

	require.NoError(t, err)
	assert.False(t, result.IsExistingProject())
	assert.False(t, result.HasSourceCode())
	assert.Empty(t, result.Language())
	assert.Empty(t, result.ManifestPath())
}

func TestFileSystemProjectDetector_Detect_WhenGoModPresent_ExpectGoProject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/user/test\n\ngo 1.22\n"), 0o644))

	detector := &FileSystemProjectDetector{}
	result, err := detector.Detect(dir)

	require.NoError(t, err)
	assert.True(t, result.HasSourceCode())
	assert.True(t, result.IsExistingProject())
	assert.Equal(t, "go", result.Language())
	assert.Equal(t, "go.mod", result.ManifestPath())
	assert.Equal(t, "github.com/user/test", result.ModulePath())
}

func TestFileSystemProjectDetector_Detect_WhenPyprojectPresent_ExpectPythonProject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[project]"), 0o644))

	detector := &FileSystemProjectDetector{}
	result, err := detector.Detect(dir)

	require.NoError(t, err)
	assert.True(t, result.HasSourceCode())
	assert.Equal(t, "python", result.Language())
	assert.Equal(t, "pyproject.toml", result.ManifestPath())
}

func TestFileSystemProjectDetector_Detect_WhenRequirementsTxtPresent_ExpectPythonProject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("flask"), 0o644))

	detector := &FileSystemProjectDetector{}
	result, err := detector.Detect(dir)

	require.NoError(t, err)
	assert.True(t, result.HasSourceCode())
	assert.Equal(t, "python", result.Language())
	assert.Equal(t, "requirements.txt", result.ManifestPath())
}

func TestFileSystemProjectDetector_Detect_WhenPackageJsonPresent_ExpectTypescriptProject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0o644))

	detector := &FileSystemProjectDetector{}
	result, err := detector.Detect(dir)

	require.NoError(t, err)
	assert.True(t, result.HasSourceCode())
	assert.Equal(t, "typescript", result.Language())
	assert.Equal(t, "package.json", result.ManifestPath())
}

func TestFileSystemProjectDetector_Detect_WhenDocsFolderPresent_ExpectDocsDetected(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs"), 0o755))

	detector := &FileSystemProjectDetector{}
	result, err := detector.Detect(dir)

	require.NoError(t, err)
	assert.True(t, result.HasDocsFolder())
	assert.True(t, result.IsExistingProject())
	assert.True(t, result.IsAmbiguous(), "docs without source should be ambiguous")
}

func TestFileSystemProjectDetector_Detect_WhenAltoConfigPresent_ExpectAltoDetected(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".alto"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".alto", "config.toml"), []byte(""), 0o644))

	detector := &FileSystemProjectDetector{}
	result, err := detector.Detect(dir)

	require.NoError(t, err)
	assert.True(t, result.HasAltoConfig())
}

func TestFileSystemProjectDetector_Detect_WhenClaudeDirPresent_ExpectAIToolDetected(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0o755))

	detector := &FileSystemProjectDetector{}
	result, err := detector.Detect(dir)

	require.NoError(t, err)
	assert.True(t, result.HasAIToolConfig())
}

func TestFileSystemProjectDetector_Detect_WhenCursorDirPresent_ExpectAIToolDetected(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".cursor"), 0o755))

	detector := &FileSystemProjectDetector{}
	result, err := detector.Detect(dir)

	require.NoError(t, err)
	assert.True(t, result.HasAIToolConfig())
}

func TestFileSystemProjectDetector_Detect_WhenClaudeMdPresent_ExpectAIToolDetected(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Claude"), 0o644))

	detector := &FileSystemProjectDetector{}
	result, err := detector.Detect(dir)

	require.NoError(t, err)
	assert.True(t, result.HasAIToolConfig())
}

func TestFileSystemProjectDetector_Detect_WhenFullProject_ExpectAllDetected(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Given: a fully set up Go project with docs and AI tools
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".alto"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".alto", "config.toml"), []byte(""), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".claude"), 0o755))

	detector := &FileSystemProjectDetector{}
	result, err := detector.Detect(dir)

	require.NoError(t, err)
	assert.True(t, result.HasSourceCode())
	assert.Equal(t, "go", result.Language())
	assert.Equal(t, "go.mod", result.ManifestPath())
	assert.True(t, result.HasDocsFolder())
	assert.True(t, result.HasAltoConfig())
	assert.True(t, result.HasAIToolConfig())
	assert.True(t, result.IsExistingProject())
	assert.False(t, result.IsAmbiguous())
}

// ---------------------------------------------------------------------------
// extractModulePath tests
// ---------------------------------------------------------------------------

func Test_extractModulePath_WhenGoMod_ExpectModuleName(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"),
		[]byte("module github.com/user/my-service\n\ngo 1.22\n"), 0o644))

	got := extractModulePath(dir, "go.mod")
	assert.Equal(t, "github.com/user/my-service", got)
}

func Test_extractModulePath_WhenPyprojectToml_ExpectProjectName(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	content := `[project]
name = "my-python-app"
version = "1.0.0"
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte(content), 0o644))

	got := extractModulePath(dir, "pyproject.toml")
	assert.Equal(t, "my-python-app", got)
}

func Test_extractModulePath_WhenPackageJson_ExpectPackageName(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	content := `{
  "name": "@org/my-ts-app",
  "version": "1.0.0"
}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(content), 0o644))

	got := extractModulePath(dir, "package.json")
	assert.Equal(t, "@org/my-ts-app", got)
}

func Test_extractModulePath_WhenUnknownManifest_ExpectEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	got := extractModulePath(dir, "Cargo.toml")
	assert.Empty(t, got)
}

func Test_extractModulePath_WhenEmptyManifestPath_ExpectEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	got := extractModulePath(dir, "")
	assert.Empty(t, got)
}

func Test_extractModulePath_WhenFileNotFound_ExpectEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	got := extractModulePath(dir, "go.mod")
	assert.Empty(t, got)
}

func Test_extractModulePath_WhenGoModMalformed_ExpectEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("not a valid go.mod"), 0o644))

	got := extractModulePath(dir, "go.mod")
	assert.Empty(t, got)
}

func Test_extractModulePath_WhenGoModHasCommentsBeforeModule_ExpectModuleName(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	content := "// Comment line\n// Another comment\nmodule github.com/user/my-project\n\ngo 1.22\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644))

	got := extractModulePath(dir, "go.mod")
	assert.Equal(t, "github.com/user/my-project", got)
}
