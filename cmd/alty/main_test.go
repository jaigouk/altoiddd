package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/composition"
)

func TestNewRootCmd_HasExpectedSubcommands(t *testing.T) {
	t.Parallel()
	app, err := composition.NewApp()
	require.NoError(t, err)
	defer app.Close()

	rootCmd := newRootCmd(app)

	expectedNames := []string{
		"version",
		"init",
		"guide",
		"detect",
		"check",
		"kb",
		"doc-health",
		"doc-review",
		"ticket-health",
		"generate",
		"persona",
	}

	cmdNames := make([]string, 0)
	for _, cmd := range rootCmd.Commands() {
		cmdNames = append(cmdNames, cmd.Name())
	}

	for _, expected := range expectedNames {
		assert.Contains(t, cmdNames, expected, "missing subcommand: %s", expected)
	}
}

func TestVersionCmd_PrintsVersion(t *testing.T) {
	t.Parallel()
	app, err := composition.NewApp()
	require.NoError(t, err)
	defer app.Close()

	rootCmd := newRootCmd(app)

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"version"})
	err = rootCmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "dev")
}

func TestHelpCmd_ShowsUsage(t *testing.T) {
	t.Parallel()
	app, err := composition.NewApp()
	require.NoError(t, err)
	defer app.Close()

	rootCmd := newRootCmd(app)

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"help"})
	err = rootCmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "alty")
	assert.Contains(t, output, "init")
	assert.Contains(t, output, "guide")
	assert.Contains(t, output, "detect")
	assert.Contains(t, output, "check")
}

func TestGenerateCmd_HasSubcommands(t *testing.T) {
	t.Parallel()
	app, err := composition.NewApp()
	require.NoError(t, err)
	defer app.Close()

	rootCmd := newRootCmd(app)

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"generate", "--help"})
	err = rootCmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "artifacts")
	assert.Contains(t, output, "fitness")
	assert.Contains(t, output, "tickets")
	assert.Contains(t, output, "configs")
}

func TestPersonaCmd_HasSubcommands(t *testing.T) {
	t.Parallel()
	app, err := composition.NewApp()
	require.NoError(t, err)
	defer app.Close()

	rootCmd := newRootCmd(app)

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"persona", "--help"})
	err = rootCmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "list")
	assert.Contains(t, output, "generate")
}

func TestDetectCmd_Runs(t *testing.T) {
	t.Parallel()
	app, err := composition.NewApp()
	require.NoError(t, err)
	defer app.Close()

	rootCmd := newRootCmd(app)

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"detect"})
	// Should not error -- may detect tools or not
	err = rootCmd.Execute()
	assert.NoError(t, err)
}

func TestPersonaListCmd_Runs(t *testing.T) {
	t.Parallel()
	app, err := composition.NewApp()
	require.NoError(t, err)
	defer app.Close()

	rootCmd := newRootCmd(app)

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"persona", "list"})
	err = rootCmd.Execute()
	assert.NoError(t, err)
}
