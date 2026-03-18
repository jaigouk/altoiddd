// Package infrastructure provides adapters for the Fitness bounded context.
package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	fitnessapp "github.com/alto-cli/alto/internal/fitness/application"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

const defaultTimeoutSeconds = 300

// SubprocessGateRunner executes quality gate commands as subprocesses.
// Commands are read from a StackProfile so the runner works for any supported stack.
type SubprocessGateRunner struct {
	projectDir string
	stackID    string
	commands   map[vo.QualityGate][]string
}

// Compile-time interface check.
var _ fitnessapp.GateRunner = (*SubprocessGateRunner)(nil)

// NewSubprocessGateRunner creates a SubprocessGateRunner.
// If projectDir is empty, defaults to the current working directory.
// If profile is nil, defaults to PythonUvProfile.
func NewSubprocessGateRunner(projectDir string, profile vo.StackProfile) *SubprocessGateRunner {
	if projectDir == "" {
		dir, err := os.Getwd()
		if err == nil {
			projectDir = dir
		}
	}
	if profile == nil {
		profile = vo.PythonUvProfile{}
	}
	return &SubprocessGateRunner{
		projectDir: projectDir,
		stackID:    profile.StackID(),
		commands:   profile.QualityGateCommands(),
	}
}

// Run executes a single quality gate as a subprocess.
func (r *SubprocessGateRunner) Run(
	ctx context.Context,
	gate vo.QualityGate,
) (vo.GateResult, error) {
	// Skip gracefully if profile has no command for this gate
	cmd, ok := r.commands[gate]
	if !ok {
		return vo.NewGateResult(gate, true,
			fmt.Sprintf("Skipped: no %s command for this stack", string(gate)), 0), nil
	}

	// FITNESS gate: skip gracefully if fitness config not found
	if gate == vo.QualityGateFitness {
		if skipResult := r.checkFitnessSkip(gate); skipResult != nil {
			return *skipResult, nil
		}
	}

	start := time.Now()

	execCtx, cancel := context.WithTimeout(ctx, defaultTimeoutSeconds*time.Second)
	defer cancel()

	command := exec.CommandContext(execCtx, cmd[0], cmd[1:]...)
	command.Dir = r.projectDir
	output, err := command.CombinedOutput()
	durationMS := int(time.Since(start).Milliseconds())

	if execCtx.Err() != nil {
		return vo.NewGateResult(gate, false,
			fmt.Sprintf("Timed out after %ds", defaultTimeoutSeconds), durationMS), nil
	}

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return vo.NewGateResult(gate, false, string(output), durationMS), nil
		}
		var execErr *exec.Error
		if errors.As(err, &execErr) {
			return vo.NewGateResult(gate, false,
				fmt.Sprintf("Command not found: %s", cmd[0]), durationMS), nil
		}
		return vo.NewGateResult(gate, false, string(output), durationMS), nil
	}

	return vo.NewGateResult(gate, true, string(output), durationMS), nil
}

// checkFitnessSkip returns a skip result if fitness tests should be skipped.
// For Go projects: skip if arch-go.yml is missing.
// For Python projects: skip if tests/architecture/ is missing.
func (r *SubprocessGateRunner) checkFitnessSkip(gate vo.QualityGate) *vo.GateResult {
	switch r.stackID {
	case "go-mod":
		archGoYAML := filepath.Join(r.projectDir, "arch-go.yml")
		if _, err := os.Stat(archGoYAML); os.IsNotExist(err) {
			result := vo.NewGateResult(gate, true, "Skipped: arch-go.yml not found", 0)
			return &result
		}
	default:
		// Python and other stacks use tests/architecture/
		archDir := filepath.Join(r.projectDir, "tests", "architecture")
		if _, err := os.Stat(archDir); os.IsNotExist(err) {
			result := vo.NewGateResult(gate, true, "Skipped: tests/architecture/ directory not found", 0)
			return &result
		}
	}
	return nil
}
