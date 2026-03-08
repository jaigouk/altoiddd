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

	fitnessapp "github.com/alty-cli/alty/internal/fitness/application"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

const defaultTimeoutSeconds = 300

// SubprocessGateRunner executes quality gate commands as subprocesses.
// Commands are read from a StackProfile so the runner works for any supported stack.
type SubprocessGateRunner struct {
	projectDir string
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

	// FITNESS gate: skip gracefully if tests/architecture/ does not exist
	if gate == vo.QualityGateFitness {
		archDir := filepath.Join(r.projectDir, "tests", "architecture")
		if _, err := os.Stat(archDir); os.IsNotExist(err) {
			return vo.NewGateResult(gate, true,
				"Skipped: tests/architecture/ directory not found", 0), nil
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
