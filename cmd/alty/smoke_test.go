//go:build smoke

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// binaryPath is the path to the compiled alty binary.
// Set by TestMain.
var binaryPath string

func TestMain(m *testing.M) {
	// Build binary to temp location for testing.
	tmpDir, err := os.MkdirTemp("", "alty-smoke-*")
	if err != nil {
		panic("creating temp dir: " + err.Error())
	}
	defer os.RemoveAll(tmpDir)

	binaryPath = filepath.Join(tmpDir, "alty")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = filepath.Join(".") // cmd/alty directory
	if out, err := cmd.CombinedOutput(); err != nil {
		panic("building binary: " + err.Error() + "\n" + string(out))
	}

	os.Exit(m.Run())
}

// runAlty executes the alty binary with the given args and returns stdout, stderr, and exit code.
func runAlty(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	cmd := exec.Command(binaryPath, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	// Set timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		stdout = outBuf.String()
		stderr = errBuf.String()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				t.Fatalf("running alty: %v", err)
			}
		}
	case <-time.After(10 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("command timed out after 10s")
	}

	return stdout, stderr, exitCode
}

// assertNoPanic checks that stderr does not contain panic indicators.
func assertNoPanic(t *testing.T, stderr string) {
	t.Helper()
	assert.NotContains(t, stderr, "panic:", "command should not panic")
	assert.NotContains(t, stderr, "runtime error", "command should not have runtime errors")
}

// --- Version ---

func TestSmoke_Version(t *testing.T) {
	stdout, stderr, exitCode := runAlty(t, "version")

	assert.Equal(t, 0, exitCode, "exit code should be 0")
	// Version output may be in stdout or combined output
	combined := stdout + stderr
	assert.NotEmpty(t, strings.TrimSpace(combined), "should print version")
	assertNoPanic(t, stderr)
}

// --- Help ---

func TestSmoke_Help(t *testing.T) {
	stdout, stderr, exitCode := runAlty(t, "help")

	assert.Equal(t, 0, exitCode, "exit code should be 0")
	assert.Contains(t, stdout, "init", "help should mention init")
	assertNoPanic(t, stderr)
}

func TestSmoke_HelpFlag(t *testing.T) {
	stdout, stderr, exitCode := runAlty(t, "--help")

	assert.Equal(t, 0, exitCode, "exit code should be 0")
	assert.Contains(t, stdout, "Available Commands", "should show Cobra help")
	assertNoPanic(t, stderr)
}

func TestSmoke_NoArgs(t *testing.T) {
	stdout, stderr, exitCode := runAlty(t)

	assert.Equal(t, 0, exitCode, "exit code should be 0")
	assert.NotEmpty(t, stdout, "should print help or usage")
	assertNoPanic(t, stderr)
}

// --- Detect ---

func TestSmoke_Detect(t *testing.T) {
	stdout, stderr, exitCode := runAlty(t, "detect")

	assert.Equal(t, 0, exitCode, "exit code should be 0")
	// Should detect at least one tool or print "no tools detected"
	assert.True(t,
		strings.Contains(stdout, "Detected") || strings.Contains(stdout, "No AI coding tools"),
		"should report detection results",
	)
	assertNoPanic(t, stderr)
}

// --- Check ---

func TestSmoke_Check(t *testing.T) {
	stdout, stderr, exitCode := runAlty(t, "check")

	// Exit 0 (all pass) or 1 (some fail) are both valid
	assert.True(t, exitCode == 0 || exitCode == 1, "exit code should be 0 or 1")
	assert.Contains(t, stdout, "[", "should show gate results")
	assertNoPanic(t, stderr)
}

// --- Persona ---

func TestSmoke_PersonaList(t *testing.T) {
	stdout, stderr, exitCode := runAlty(t, "persona", "list")

	assert.Equal(t, 0, exitCode, "exit code should be 0")
	assert.Contains(t, stdout, "Available Personas", "should list personas")
	assert.Contains(t, stdout, "Solo Developer", "should include Solo Developer persona")
	assertNoPanic(t, stderr)
}

// --- KB ---

func TestSmoke_KB(t *testing.T) {
	stdout, stderr, exitCode := runAlty(t, "kb")

	assert.Equal(t, 0, exitCode, "exit code should be 0")
	assert.Contains(t, stdout, "Knowledge Base Categories", "should list KB categories")
	assert.Contains(t, stdout, "ddd", "should include ddd category")
	assertNoPanic(t, stderr)
}

// --- Doc Health ---

func TestSmoke_DocHealth(t *testing.T) {
	stdout, stderr, exitCode := runAlty(t, "doc-health")

	// Exit 0 (all ok) or 1 (issues found) are both valid
	assert.True(t, exitCode == 0 || exitCode == 1, "exit code should be 0 or 1")
	assert.Contains(t, stdout, "Doc Health", "should show doc health header")
	assertNoPanic(t, stderr)
}

// --- Doc Review ---

func TestSmoke_DocReviewList(t *testing.T) {
	stdout, stderr, exitCode := runAlty(t, "doc-review", "list")

	assert.Equal(t, 0, exitCode, "exit code should be 0")
	// Either shows docs or "No docs due for review"
	assert.True(t,
		strings.Contains(stdout, "docs") || strings.Contains(stdout, "No docs"),
		"should show review status",
	)
	assertNoPanic(t, stderr)
}

// --- Init (dry-run) ---

func TestSmoke_InitExistingDryRun(t *testing.T) {
	// This test requires being in a git repo, which the project root is
	stdout, stderr, exitCode := runAlty(t, "init", "--existing", "--dry-run")

	// May succeed (exit 0) or fail due to preconditions (exit 1)
	// Either way, should not panic
	assert.True(t, exitCode == 0 || exitCode == 1, "exit code should be 0 or 1")
	assertNoPanic(t, stderr)
	_ = stdout // stdout content varies by project state
}

// --- Invalid Command ---

func TestSmoke_InvalidCommand(t *testing.T) {
	stdout, stderr, exitCode := runAlty(t, "nonexistent-command")

	assert.NotEqual(t, 0, exitCode, "exit code should be non-zero for invalid command")
	combined := stdout + stderr
	assert.Contains(t, combined, "unknown command", "should report unknown command")
	assertNoPanic(t, stderr)
}

// --- Comprehensive No-Panic Test ---

func TestSmoke_NoPanicAcrossCommands(t *testing.T) {
	commands := [][]string{
		{"version"},
		{"help"},
		{"--help"},
		{},
		{"detect"},
		{"persona", "list"},
		{"kb"},
		{"doc-health"},
		{"doc-review", "list"},
	}

	for _, args := range commands {
		name := "no-args"
		if len(args) > 0 {
			name = strings.Join(args, "_")
		}
		t.Run(name, func(t *testing.T) {
			_, stderr, _ := runAlty(t, args...)
			assertNoPanic(t, stderr)
		})
	}
}
