package infrastructure_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rescueapp "github.com/alto-cli/alto/internal/rescue/application"
	"github.com/alto-cli/alto/internal/rescue/infrastructure"
)

// Compile-time interface check.
var _ rescueapp.TestRunner = (*infrastructure.TestRunnerAdapter)(nil)

// ---------------------------------------------------------------------------
// Detect - Go projects
// ---------------------------------------------------------------------------

func TestDetect_GoProject_ReturnsGo(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0o644))

	adapter := &infrastructure.TestRunnerAdapter{}
	framework, err := adapter.Detect(context.Background(), dir)

	require.NoError(t, err)
	assert.Equal(t, rescueapp.TestFrameworkGo, framework)
}

// ---------------------------------------------------------------------------
// Detect - npm projects
// ---------------------------------------------------------------------------

func TestDetect_NpmProject_WithTestScript_ReturnsNpm(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	packageJSON := `{"name": "test", "scripts": {"test": "jest"}}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0o644))

	adapter := &infrastructure.TestRunnerAdapter{}
	framework, err := adapter.Detect(context.Background(), dir)

	require.NoError(t, err)
	assert.Equal(t, rescueapp.TestFrameworkNPM, framework)
}

func TestDetect_NpmProject_WithoutTestScript_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	packageJSON := `{"name": "test", "scripts": {"build": "tsc"}}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0o644))

	adapter := &infrastructure.TestRunnerAdapter{}
	framework, err := adapter.Detect(context.Background(), dir)

	require.NoError(t, err)
	assert.Empty(t, framework)
}

func TestDetect_NpmProject_NoScripts_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	packageJSON := `{"name": "test"}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0o644))

	adapter := &infrastructure.TestRunnerAdapter{}
	framework, err := adapter.Detect(context.Background(), dir)

	require.NoError(t, err)
	assert.Empty(t, framework)
}

// ---------------------------------------------------------------------------
// Detect - pytest projects
// ---------------------------------------------------------------------------

func TestDetect_PytestProject_WithPytestIni_ReturnsPytest(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pytest.ini"), []byte("[pytest]"), 0o644))

	adapter := &infrastructure.TestRunnerAdapter{}
	framework, err := adapter.Detect(context.Background(), dir)

	require.NoError(t, err)
	assert.Equal(t, rescueapp.TestFrameworkPytest, framework)
}

func TestDetect_PytestProject_WithConftest_ReturnsPytest(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "conftest.py"), []byte("# conftest"), 0o644))

	adapter := &infrastructure.TestRunnerAdapter{}
	framework, err := adapter.Detect(context.Background(), dir)

	require.NoError(t, err)
	assert.Equal(t, rescueapp.TestFrameworkPytest, framework)
}

// ---------------------------------------------------------------------------
// Detect - No framework
// ---------------------------------------------------------------------------

func TestDetect_NoFramework_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	adapter := &infrastructure.TestRunnerAdapter{}
	framework, err := adapter.Detect(context.Background(), dir)

	require.NoError(t, err)
	assert.Empty(t, framework)
}

// ---------------------------------------------------------------------------
// Detect - Priority (Go > npm > pytest)
// ---------------------------------------------------------------------------

func TestDetect_MultipleFrameworks_PrefersGo(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0o644))
	packageJSON := `{"name": "test", "scripts": {"test": "jest"}}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(packageJSON), 0o644))

	adapter := &infrastructure.TestRunnerAdapter{}
	framework, err := adapter.Detect(context.Background(), dir)

	require.NoError(t, err)
	assert.Equal(t, rescueapp.TestFrameworkGo, framework)
}

// ---------------------------------------------------------------------------
// Run - Unknown framework
// ---------------------------------------------------------------------------

func TestRun_UnknownFramework_ReturnsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	adapter := &infrastructure.TestRunnerAdapter{}
	err := adapter.Run(context.Background(), dir, "unknown")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown framework")
}

func TestRun_EmptyFramework_ReturnsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	adapter := &infrastructure.TestRunnerAdapter{}
	err := adapter.Run(context.Background(), dir, "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown framework")
}

// ---------------------------------------------------------------------------
// Detect - Context cancellation
// ---------------------------------------------------------------------------

func TestDetect_CancelledContext_ReturnsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	adapter := &infrastructure.TestRunnerAdapter{}
	_, err := adapter.Detect(ctx, dir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "detecting test framework")
}

// ---------------------------------------------------------------------------
// Run - Context cancellation
// ---------------------------------------------------------------------------

func TestRun_CancelledContext_ReturnsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0o644))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	adapter := &infrastructure.TestRunnerAdapter{}
	err := adapter.Run(ctx, dir, rescueapp.TestFrameworkGo)

	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Run - Go framework integration test
// ---------------------------------------------------------------------------

func TestRun_GoFramework_HappyPath(t *testing.T) {
	t.Parallel()

	// Create a minimal Go project with a passing test
	dir := t.TempDir()
	goMod := `module testproject

go 1.21
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644))

	// Create a simple passing test file
	testFile := `package testproject

import "testing"

func TestPass(t *testing.T) {
	// This test always passes
}
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pass_test.go"), []byte(testFile), 0o644))

	adapter := &infrastructure.TestRunnerAdapter{}
	err := adapter.Run(context.Background(), dir, rescueapp.TestFrameworkGo)

	require.NoError(t, err)
}
