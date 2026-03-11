// Package infrastructure provides adapters for the Rescue bounded context.
package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	rescueapp "github.com/alty-cli/alty/internal/rescue/application"
)

// TestRunnerAdapter implements TestRunner using filesystem detection and subprocess execution.
type TestRunnerAdapter struct{}

// Compile-time interface check.
var _ rescueapp.TestRunner = (*TestRunnerAdapter)(nil)

// Detect identifies the test framework used in a project directory.
// Returns one of TestFramework* constants, or empty string if none detected.
// Priority: Go > npm > pytest.
func (t *TestRunnerAdapter) Detect(ctx context.Context, projectDir string) (string, error) {
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("detecting test framework: %w", ctx.Err())
	default:
	}

	// Check for Go project (go.mod)
	if fileExists(filepath.Join(projectDir, "go.mod")) {
		return rescueapp.TestFrameworkGo, nil
	}

	// Check for npm project (package.json with test script)
	if hasNpmTestScript(projectDir) {
		return rescueapp.TestFrameworkNPM, nil
	}

	// Check for pytest project (pytest.ini or conftest.py)
	if fileExists(filepath.Join(projectDir, "pytest.ini")) ||
		fileExists(filepath.Join(projectDir, "conftest.py")) {
		return rescueapp.TestFrameworkPytest, nil
	}

	return "", nil
}

// Run executes tests using the specified framework.
func (t *TestRunnerAdapter) Run(ctx context.Context, projectDir, framework string) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("running tests: %w", ctx.Err())
	default:
	}

	var cmd *exec.Cmd

	switch framework {
	case rescueapp.TestFrameworkGo:
		cmd = exec.CommandContext(ctx, "go", "test", "./...")
	case rescueapp.TestFrameworkNPM:
		cmd = exec.CommandContext(ctx, "npm", "test")
	case rescueapp.TestFrameworkPytest:
		cmd = exec.CommandContext(ctx, "pytest")
	default:
		return fmt.Errorf("unknown framework: %q", framework)
	}

	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tests failed: %s: %w", string(output), err)
	}

	return nil
}

// hasNpmTestScript checks if package.json exists and has a "test" script.
func hasNpmTestScript(projectDir string) bool {
	packagePath := filepath.Join(projectDir, "package.json")
	data, err := os.ReadFile(packagePath)
	if err != nil {
		return false
	}

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false
	}

	_, hasTest := pkg.Scripts["test"]
	return hasTest
}
