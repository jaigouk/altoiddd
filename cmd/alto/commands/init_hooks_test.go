package commands_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/cmd/alto/commands"
	"github.com/alto-cli/alto/internal/composition"
)

func TestInitCmd_NoHooksAndForceHooksFlags_Defined(t *testing.T) {
	t.Parallel()
	app, err := composition.NewApp()
	require.NoError(t, err)
	t.Cleanup(func() { _ = app.Close() })

	cmd := commands.NewInitCmd(app)
	for _, name := range []string{"no-hooks", "force-hooks"} {
		assert.NotNil(t, cmd.Flags().Lookup(name), "flag --%s must be registered", name)
	}
}

func TestInitCmd_WithScaffold_DefaultsToWritingHooks(t *testing.T) {
	// Chdir-based; cannot run in parallel with other Chdir tests.
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# demo\n"), 0o644))

	app, cleanup := newCmdInDir(t, dir)
	defer cleanup()

	cmd := commands.NewInitCmd(app)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{
		"--with-scaffold",
		"--project-name=demo",
		"--ticket-prefix=demo-",
		"--issue-tracker=beads",
		"--primary-tool=claude",
	})

	require.NoError(t, cmd.Execute())

	hookPath := filepath.Join(dir, ".beads", "hooks", "post-close")
	info, err := os.Stat(hookPath)
	require.NoError(t, err, "post-close hook must be written by default")
	assert.False(t, info.IsDir())

	body, _ := os.ReadFile(hookPath)
	assert.Contains(t, string(body), "alto ticket-ripple")
}

func TestInitCmd_WithScaffold_NoHooksSuppressesHookWrite(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# demo\n"), 0o644))

	app, cleanup := newCmdInDir(t, dir)
	defer cleanup()

	cmd := commands.NewInitCmd(app)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{
		"--with-scaffold",
		"--project-name=demo",
		"--ticket-prefix=demo-",
		"--issue-tracker=beads",
		"--primary-tool=claude",
		"--no-hooks",
	})

	require.NoError(t, cmd.Execute())

	_, err := os.Stat(filepath.Join(dir, ".beads", "hooks", "post-close"))
	assert.True(t, os.IsNotExist(err), "--no-hooks must suppress hook file write")
}

func TestInitCmd_WithScaffold_NoHooksAndForceHooks_RejectedAsMutuallyExclusive(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# demo\n"), 0o644))

	app, cleanup := newCmdInDir(t, dir)
	defer cleanup()

	cmd := commands.NewInitCmd(app)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{
		"--with-scaffold",
		"--project-name=demo",
		"--ticket-prefix=demo-",
		"--issue-tracker=beads",
		"--primary-tool=claude",
		"--no-hooks",
		"--force-hooks",
	})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestInitCmd_WithScaffold_ForceHooksOverwritesExistingDifferentHook(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# demo\n"), 0o644))
	hooksDir := filepath.Join(dir, ".beads", "hooks")
	require.NoError(t, os.MkdirAll(hooksDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(hooksDir, "post-close"), []byte("# stale\n"), 0o644))

	app, cleanup := newCmdInDir(t, dir)
	defer cleanup()

	cmd := commands.NewInitCmd(app)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{
		"--with-scaffold",
		"--project-name=demo",
		"--ticket-prefix=demo-",
		"--issue-tracker=beads",
		"--primary-tool=claude",
		"--force-hooks",
	})

	require.NoError(t, cmd.Execute())

	body, _ := os.ReadFile(filepath.Join(hooksDir, "post-close"))
	assert.Contains(t, string(body), "alto ticket-ripple")
	assert.NotContains(t, string(body), "stale")
}
