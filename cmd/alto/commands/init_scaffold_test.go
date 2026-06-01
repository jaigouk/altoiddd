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

func newCmdInDir(t *testing.T, dir string) (*composition.App, func()) {
	t.Helper()
	app, err := composition.NewApp()
	require.NoError(t, err)

	prev, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	cleanup := func() {
		_ = os.Chdir(prev)
		_ = app.Close()
	}
	return app, cleanup
}

func TestInitCmd_WithScaffoldFlag_Defined(t *testing.T) {
	t.Parallel()
	app, err := composition.NewApp()
	require.NoError(t, err)
	t.Cleanup(func() { _ = app.Close() })

	cmd := commands.NewInitCmd(app)
	for _, name := range []string{
		"with-scaffold", "project-name", "ticket-prefix",
		"issue-tracker", "bounded-contexts", "primary-tool", "force",
	} {
		assert.NotNil(t, cmd.Flags().Lookup(name), "flag --%s must be registered", name)
	}
}

func TestInitCmd_WithScaffold_WritesAltoTree(t *testing.T) {
	// Chdir-based; cannot run in parallel with other Chdir tests.
	dir := t.TempDir()
	// Seed a README so any auto-detect path that would consult it has something to read.
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
		"--bounded-contexts=Orders,Catalog",
		"--primary-tool=claude",
	})

	require.NoError(t, cmd.Execute())

	info, err := os.Stat(filepath.Join(dir, "alto-scaffold"))
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestInitCmd_WithScaffold_CursorPrimaryTool_Rejected(t *testing.T) {
	// Chdir-based; cannot run in parallel with other Chdir tests.
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# x\n"), 0o644))

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
		"--primary-tool=cursor",
	})

	err := cmd.Execute()
	require.Error(t, err)
}

func TestInitCmd_WithScaffold_RooPrimaryTool_Rejected(t *testing.T) {
	// Chdir-based; cannot run in parallel with other Chdir tests.
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# x\n"), 0o644))

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
		"--primary-tool=roo",
	})

	err := cmd.Execute()
	require.Error(t, err)
}

func TestInitCmd_WithScaffold_NoForceExistingAlto_Refuses(t *testing.T) {
	// Chdir-based; cannot run in parallel with other Chdir tests.
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# x\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "alto-scaffold"), 0o755))

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

	err := cmd.Execute()
	require.Error(t, err, "missing --force must refuse to overwrite existing alto-scaffold/")
}

func TestInitCmd_WithScaffold_ForceExistingAlto_Overwrites(t *testing.T) {
	// Chdir-based; cannot run in parallel with other Chdir tests.
	// The [OVERWRITE] line emission is covered by
	// TestEmbedScaffoldWriter_WriteScaffold_ExistingAltoDirWithForce_Overwrites
	// at the adapter level (the lines go to os.Stderr directly, which
	// cobra's SetErr does not intercept). This test verifies the CLI
	// reaches the adapter and the stale file is overwritten end-to-end.
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# x\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "alto-scaffold"), 0o755))
	stalePath := filepath.Join(dir, "alto-scaffold", "CONTEXT.md")
	require.NoError(t, os.WriteFile(stalePath, []byte("stale content marker"), 0o644))

	app, cleanup := newCmdInDir(t, dir)
	defer cleanup()

	cmd := commands.NewInitCmd(app)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{
		"--with-scaffold",
		"--force",
		"--project-name=demo",
		"--ticket-prefix=demo-",
		"--issue-tracker=beads",
		"--primary-tool=claude",
	})

	require.NoError(t, cmd.Execute())

	content, err := os.ReadFile(stalePath) //nolint:gosec // test reads its own tempdir
	require.NoError(t, err)
	assert.NotContains(t, string(content), "stale content marker",
		"--force must overwrite the stale alto-scaffold/CONTEXT.md")
}
