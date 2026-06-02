package commands_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/cmd/alto/commands"
	"github.com/alto-cli/alto/internal/composition"
	dochealthapp "github.com/alto-cli/alto/internal/dochealth/application"
	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
	dochealthinfra "github.com/alto-cli/alto/internal/dochealth/infrastructure"
)

// validScaffoldFrontmatter is a minimal 8-field scaffold asset header used by
// the secret-patterns CLI integration tests. Keeping it inline avoids
// cross-package fixture sharing for a single-file test surface.
const validScaffoldFrontmatter = `---
name: foo
description: x
kind: command
phase: groom
when_to_use: test
tools: Read
bash_substitution_policy: none
license: Apache-2.0
---
`

// writeScaffoldAsset is a tiny helper that drops a scaffold asset at
// <root>/commands/<name>.md with the canonical frontmatter and the supplied
// body. Returns the absolute path written.
func writeScaffoldAsset(t *testing.T, root, name, body string) string {
	t.Helper()
	cmdDir := filepath.Join(root, "commands")
	require.NoError(t, os.MkdirAll(cmdDir, 0o755))
	path := filepath.Join(cmdDir, name+".md")
	content := validScaffoldFrontmatter + body
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

// defaultAppForScaffold wires a ScaffoldHealthHandler backed by the default
// (binding-floor) rules — used as the "before" leg in tests that need to
// distinguish default vs custom-pattern behaviour.
func defaultAppForScaffold(t *testing.T) *composition.App {
	t.Helper()
	params, err := dochealthdomain.NewScaffoldParams(30, nil)
	require.NoError(t, err)
	walker := dochealthinfra.NewFilesystemScaffoldWalker()
	handler := dochealthapp.NewScaffoldHealthHandler(walker, dochealthinfra.DefaultScaffoldRules(params))
	return &composition.App{ScaffoldHealthHandler: handler}
}

// runScaffoldOverDir runs the doc-health command against altoDir using the
// supplied app + optional --secret-patterns flag value. Returns combined
// stdout+stderr and the cobra error so tests can assert both report content
// and exit status.
func runScaffoldOverDir(t *testing.T, app *composition.App, altoDir, secretPatternsPath string) (string, error) {
	t.Helper()
	cmd := commands.NewDocHealthCmd(app)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	args := []string{"--paths=" + altoDir}
	if secretPatternsPath != "" {
		args = append(args, "--secret-patterns="+secretPatternsPath)
	}
	cmd.SetArgs(args)
	cmd.SetContext(context.Background())
	err := cmd.Execute()
	return buf.String(), err
}

// TestDocHealthCmd_SecretPatternsFlag_Defined locks the flag's existence + type.
// If the flag goes missing or changes shape, this fails immediately.
func TestDocHealthCmd_SecretPatternsFlag_Defined(t *testing.T) {
	t.Parallel()
	app := defaultAppForScaffold(t)
	cmd := commands.NewDocHealthCmd(app)
	f := cmd.Flags().Lookup("secret-patterns")
	require.NotNil(t, f, "doc-health must declare --secret-patterns flag")
	assert.Equal(t, "string", f.Value.Type())
}

// TestDocHealthCmd_SecretPatternsFlag_LoadsAndOverrides exercises the full
// file-read -> YAML-parse -> ScaffoldParams -> handler-build path. Asserts
// OVERRIDE semantics: with a custom pattern set, the binding-floor defaults
// do NOT also fire.
func TestDocHealthCmd_SecretPatternsFlag_LoadsAndOverrides(t *testing.T) {
	t.Parallel()
	// Build a scaffold tree containing TWO synthetic markers:
	//   1. "the password is hunter2"  — matches default keyword set
	//   2. "ZZZZ-12345"               — matches our custom regex only
	altoDir := t.TempDir()
	writeScaffoldAsset(t, altoDir, "alpha", "the password is hunter2 — ZZZZ-12345\n")

	// Author a YAML file with ONE custom pattern. The default keyword set
	// (`password`, `secret`, ...) must NOT apply when overrides are supplied.
	yamlDir := t.TempDir()
	yamlPath := filepath.Join(yamlDir, "patterns.yaml")
	require.NoError(t, os.WriteFile(yamlPath, []byte(
		"- name: custom_marker\n  pattern: 'ZZZZ-[0-9]+'\n",
	), 0o600))

	// First leg: default app sees BOTH the keyword hit AND no custom hit (no
	// override). Establishes the baseline.
	defaultOut, defaultErr := runScaffoldOverDir(t, defaultAppForScaffold(t), altoDir, "")
	require.NoError(t, defaultErr)
	assert.Contains(t, defaultOut, "secrets_grep",
		"default rule must fire on keyword 'password'")
	assert.Contains(t, defaultOut, "keyword",
		"default match cites the keyword pattern by name")
	assert.NotContains(t, defaultOut, "custom_marker",
		"default rule must NOT report the custom pattern's name")

	// Second leg: CLI flag flow loads the YAML, builds the override handler,
	// runs it. Custom pattern fires; default keyword pattern does NOT.
	overrideOut, overrideErr := runScaffoldOverDir(t, defaultAppForScaffold(t), altoDir, yamlPath)
	require.NoError(t, overrideErr)
	assert.Contains(t, overrideOut, "custom_marker",
		"--secret-patterns override must invoke the custom pattern")
	assert.Contains(t, overrideOut, "ZZZZ-12345",
		"violation message includes the matched substring")
	assert.NotContains(t, overrideOut, "keyword",
		"override semantics: defaults must NOT also fire alongside custom patterns")
}

// TestDocHealthCmd_SecretPatternsFlag_MissingFile_WrapsErr ensures the loader
// surfaces a wrapped error (not a panic) when the path does not exist.
func TestDocHealthCmd_SecretPatternsFlag_MissingFile_WrapsErr(t *testing.T) {
	t.Parallel()
	altoDir := t.TempDir()
	writeScaffoldAsset(t, altoDir, "alpha", "body\n")

	app := defaultAppForScaffold(t)
	cmd := commands.NewDocHealthCmd(app)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{
		"--paths=" + altoDir,
		"--secret-patterns=/nonexistent/path/patterns.yaml",
	})
	cmd.SetContext(context.Background())

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading secret patterns",
		"error must be wrapped with the loader context")
}

// TestDocHealthCmd_SecretPatternsFlag_InvalidYAML_WrapsErr ensures malformed
// YAML produces a wrapped parse error.
func TestDocHealthCmd_SecretPatternsFlag_InvalidYAML_WrapsErr(t *testing.T) {
	t.Parallel()
	altoDir := t.TempDir()
	writeScaffoldAsset(t, altoDir, "alpha", "body\n")

	yamlDir := t.TempDir()
	yamlPath := filepath.Join(yamlDir, "bad.yaml")
	// Hard YAML parse failure — unmatched bracket at top level.
	require.NoError(t, os.WriteFile(yamlPath, []byte("[ - this is not valid yaml\n"), 0o600))

	app := defaultAppForScaffold(t)
	out, err := runScaffoldOverDir(t, app, altoDir, yamlPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading secret patterns")
	_ = out
}

// TestDocHealthCmd_SecretPatternsFlag_BadRegex_WrapsErr ensures the loader
// rejects a YAML entry whose regex fails to compile.
func TestDocHealthCmd_SecretPatternsFlag_BadRegex_WrapsErr(t *testing.T) {
	t.Parallel()
	altoDir := t.TempDir()
	writeScaffoldAsset(t, altoDir, "alpha", "body\n")

	yamlDir := t.TempDir()
	yamlPath := filepath.Join(yamlDir, "bad.yaml")
	require.NoError(t, os.WriteFile(yamlPath, []byte(
		"- name: broken\n  pattern: '[unclosed'\n",
	), 0o600))

	app := defaultAppForScaffold(t)
	out, err := runScaffoldOverDir(t, app, altoDir, yamlPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading secret patterns",
		"compile failure must surface through the loader wrap")
	_ = out
}
