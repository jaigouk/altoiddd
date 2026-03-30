package commands_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/cmd/alto/commands"
	"github.com/alto-cli/alto/internal/composition"
)

func TestNewGuideCmd_ContinueFlagMentionsAgent(t *testing.T) {
	t.Parallel()

	app, err := composition.NewApp()
	require.NoError(t, err)
	defer func() { _ = app.Close() }()

	cmd := commands.NewGuideCmd(app)

	continueFlag := cmd.Flags().Lookup("continue")
	require.NotNil(t, continueFlag, "--continue flag must exist")
	assert.Contains(t, continueFlag.Usage, "agent",
		"--continue flag usage should mention agent-mode sessions")
}

func TestNewGuideCmd_LongHelpMentionsContinueRequiresAgent(t *testing.T) {
	t.Parallel()

	app, err := composition.NewApp()
	require.NoError(t, err)
	defer func() { _ = app.Close() }()

	cmd := commands.NewGuideCmd(app)

	assert.Contains(t, cmd.Long, "--continue",
		"Long help text should mention --continue")
	assert.Contains(t, cmd.Long, "started with --agent",
		"Long help text should clarify that --continue requires a session started with --agent")
}

func TestNewGuideCmd_ContinueNoSession_ErrorMentionsAgent(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Create .alto/config.toml so the init guard passes
	altoDir := filepath.Join(tmpDir, ".alto")
	require.NoError(t, os.MkdirAll(altoDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(altoDir, "config.toml"), nil, 0o644))

	app, err := composition.NewApp()
	require.NoError(t, err)
	defer func() { _ = app.Close() }()

	cmd := commands.NewGuideCmd(app)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--continue"})

	err = cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "alto guide --agent",
		"error message should tell users to start with alto guide --agent")
}
