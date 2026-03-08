package infrastructure_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	fitnessapp "github.com/alty-cli/alty/internal/fitness/application"
	"github.com/alty-cli/alty/internal/fitness/infrastructure"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// SubprocessGateRunner
// ---------------------------------------------------------------------------

func TestSubprocessGateRunnerImplementsPort(t *testing.T) {
	t.Parallel()
	var _ fitnessapp.GateRunner = (*infrastructure.SubprocessGateRunner)(nil)
}

func TestFitnessSkipsWhenNoArchitectureDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	runner := infrastructure.NewSubprocessGateRunner(dir, nil)

	result, err := runner.Run(context.Background(), vo.QualityGateFitness)
	require.NoError(t, err)
	assert.Equal(t, vo.QualityGateFitness, result.Gate())
	assert.True(t, result.Passed())
	assert.Contains(t, result.Output(), "Skipped")
	assert.Equal(t, 0, result.DurationMS())
}

func TestFitnessRunsWhenArchitectureDirExists(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	archDir := filepath.Join(dir, "tests", "architecture")
	require.NoError(t, os.MkdirAll(archDir, 0o755))

	runner := infrastructure.NewSubprocessGateRunner(dir, nil)
	result, err := runner.Run(context.Background(), vo.QualityGateFitness)
	require.NoError(t, err)

	assert.Equal(t, vo.QualityGateFitness, result.Gate())
	assert.GreaterOrEqual(t, result.DurationMS(), 0)
}

func TestGateResultHasCorrectGate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	runner := infrastructure.NewSubprocessGateRunner(dir, nil)
	result, err := runner.Run(context.Background(), vo.QualityGateFitness)
	require.NoError(t, err)
	assert.Equal(t, vo.QualityGateFitness, result.Gate())
}

func TestDurationIsNonNegative(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	runner := infrastructure.NewSubprocessGateRunner(dir, nil)
	result, err := runner.Run(context.Background(), vo.QualityGateFitness)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.DurationMS(), 0)
}

func TestGateSkipsWhenNoCommand(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	runner := infrastructure.NewSubprocessGateRunner(dir, vo.GenericProfile{})
	result, err := runner.Run(context.Background(), vo.QualityGateLint)
	require.NoError(t, err)
	assert.True(t, result.Passed())
	assert.Contains(t, result.Output(), "Skipped")
}

// ---------------------------------------------------------------------------
// CodebasePortScanner
// ---------------------------------------------------------------------------

func TestPortScannerFindsInterfaces(t *testing.T) {
	t.Parallel()
	// Scan the real fitness ports directory
	portsDir := filepath.Join("..", "application")
	scanner := infrastructure.CodebasePortScanner{}
	result := scanner.Scan(portsDir)

	assert.NotEmpty(t, result)
	_, hasGateRunner := result["GateRunner"]
	assert.True(t, hasGateRunner, "expected to find GateRunner interface")
}

func TestPortScannerExtractsMethodSignatures(t *testing.T) {
	t.Parallel()
	portsDir := filepath.Join("..", "application")
	scanner := infrastructure.CodebasePortScanner{}
	result := scanner.Scan(portsDir)

	gateRunner, ok := result["GateRunner"]
	require.True(t, ok, "GateRunner not found")
	methodNames := make([]string, len(gateRunner.Methods()))
	for i, m := range gateRunner.Methods() {
		methodNames[i] = m.Name()
	}
	assert.Contains(t, methodNames, "Run")
}

func TestPortScannerEmptyDirectory(t *testing.T) {
	t.Parallel()
	scanner := infrastructure.CodebasePortScanner{}
	result := scanner.Scan(t.TempDir())
	assert.Empty(t, result)
}

func TestPortScannerMalformedFileSkipped(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	badFile := filepath.Join(dir, "broken.go")
	require.NoError(t, os.WriteFile(badFile, []byte("package broken\n\nfunc broken( {"), 0o644))

	scanner := infrastructure.CodebasePortScanner{}
	result := scanner.Scan(dir)
	assert.Empty(t, result)
}

func TestPortScannerNonExistentDir(t *testing.T) {
	t.Parallel()
	scanner := infrastructure.CodebasePortScanner{}
	result := scanner.Scan(filepath.Join(t.TempDir(), "nonexistent"))
	assert.Empty(t, result)
}
