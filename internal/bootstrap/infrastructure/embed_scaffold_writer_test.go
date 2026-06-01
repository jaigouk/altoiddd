package infrastructure_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bootstrapapp "github.com/alto-cli/alto/internal/bootstrap/application"
	"github.com/alto-cli/alto/internal/bootstrap/domain"
	"github.com/alto-cli/alto/internal/bootstrap/infrastructure"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// validParams returns the canonical happy-path ScaffoldParams.
func validParams(t *testing.T) domain.ScaffoldParams {
	t.Helper()
	p, err := domain.NewScaffoldParams("demo", "demo-", "beads", []string{"Orders"}, "claude")
	require.NoError(t, err)
	return p
}

func TestEmbedScaffoldWriter_CompileTimePortAssertion(t *testing.T) {
	t.Parallel()
	// If the package compiles, the `var _ application.ScaffoldWriter =
	// (*EmbedScaffoldWriter)(nil)` assertion at the top of
	// embed_scaffold_writer.go holds; this test merely records intent and
	// ensures the constructor returns the right type.
	var _ bootstrapapp.ScaffoldWriter = (*infrastructure.EmbedScaffoldWriter)(nil)
	w := infrastructure.NewEmbedScaffoldWriter()
	assert.NotNil(t, w)
}

func TestEmbedScaffoldWriter_WriteScaffold_CreatesAltoDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	w := infrastructure.NewEmbedScaffoldWriter()
	require.NoError(t, w.WriteScaffold(context.Background(), dir, validParams(t), false))

	info, err := os.Stat(filepath.Join(dir, "alto-scaffold"))
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	contextMd := filepath.Join(dir, "alto-scaffold", "CONTEXT.md")
	_, err = os.Stat(contextMd)
	require.NoError(t, err, "alto-scaffold/CONTEXT.md must be written")
}

func TestEmbedScaffoldWriter_WriteScaffold_ZeroUnsubstitutedPlaceholders(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	w := infrastructure.NewEmbedScaffoldWriter()
	require.NoError(t, w.WriteScaffold(context.Background(), dir, validParams(t), false))

	placeholderRe := regexp.MustCompile(`\{\{\s*\.?(ProjectName|TicketPrefix|IssueTracker|BoundedContexts|PrimaryTool)\b`)
	root := filepath.Join(dir, "alto-scaffold")
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return walkErr
		}
		content, err := os.ReadFile(path) //nolint:gosec // test reads its own tempdir
		if err != nil {
			return err
		}
		if placeholderRe.Match(content) {
			t.Errorf("unsubstituted placeholder in %s", path)
		}
		return nil
	})
	require.NoError(t, err)
}

func TestEmbedScaffoldWriter_WriteScaffold_AttackerProjectName_LiteralNotReevaluated(t *testing.T) {
	t.Parallel()
	// Critical security AC: text/template DATA binding renders an
	// attacker-controlled ProjectName as literal output, never as a
	// re-evaluated template expression. Exercised at the renderTemplate
	// boundary because the current embed bodies do not yet contain
	// {{.ProjectName}} substitutions — the security property must hold
	// independently of which embed files use the substitution.
	attacker, err := domain.NewScaffoldParams("{{.Evil}}", "demo-", "beads", nil, "claude")
	require.NoError(t, err)

	body := []byte("project: {{.ProjectName}}")
	rendered, err := infrastructure.RenderTemplateForTest(t, "synthetic.md", body, attacker)
	require.NoError(t, err)
	assert.Equal(t, "project: {{.Evil}}", string(rendered),
		"attacker-controlled {{.Evil}} must round-trip verbatim — text/template data binding does not re-evaluate it")
}

func TestEmbedScaffoldWriter_WriteScaffold_AttackerProjectName_DoesNotCrashRenderer(t *testing.T) {
	t.Parallel()
	// Full-embed run with attacker-controlled name must not panic or
	// error; absence of crash + completion is the binding safety contract
	// when the embed has no {{ }} placeholders to substitute through.
	dir := t.TempDir()
	attacker, err := domain.NewScaffoldParams("{{.Evil}}", "demo-", "beads", nil, "claude")
	require.NoError(t, err)
	w := infrastructure.NewEmbedScaffoldWriter()
	require.NoError(t, w.WriteScaffold(context.Background(), dir, attacker, false))
}

func TestEmbedScaffoldWriter_WriteScaffold_ExistingAltoDirNoForce_ReturnsErrAlreadyExists(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "alto-scaffold"), 0o755))

	w := infrastructure.NewEmbedScaffoldWriter()
	err := w.WriteScaffold(context.Background(), dir, validParams(t), false)
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrAlreadyExists)
}

func TestEmbedScaffoldWriter_WriteScaffold_ExistingAltoDirWithForce_Overwrites(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "alto-scaffold", "CONTEXT.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(target), 0o755))
	require.NoError(t, os.WriteFile(target, []byte("stale content"), 0o644))

	w := infrastructure.NewEmbedScaffoldWriter()
	require.NoError(t, w.WriteScaffold(context.Background(), dir, validParams(t), true))

	content, err := os.ReadFile(target) //nolint:gosec // test reads its own tempdir
	require.NoError(t, err)
	assert.NotContains(t, string(content), "stale content")
}

func TestEmbedScaffoldWriter_WriteScaffold_UsesOExclByDefault(t *testing.T) {
	t.Parallel()
	// If the writer ignored O_EXCL, planting a pre-existing regular file
	// at the alto-scaffold/ path would let the call clobber it. O_EXCL turns
	// that into an error.
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "alto-scaffold"), []byte("x"), 0o644))

	w := infrastructure.NewEmbedScaffoldWriter()
	err := w.WriteScaffold(context.Background(), dir, validParams(t), false)
	require.Error(t, err, "O_EXCL must surface a collision with the pre-planted file")
}

func TestEmbedScaffoldWriter_WriteScaffold_ReadOnlyTargetDir_ReturnsWrappedError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("read-only directory semantics differ on Windows")
	}
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.Chmod(dir, 0o555))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	w := infrastructure.NewEmbedScaffoldWriter()
	err := w.WriteScaffold(context.Background(), dir, validParams(t), false)
	require.Error(t, err)
}

func TestEmbedScaffoldWriter_EmbedCount_MatchesExpectedFileCount(t *testing.T) {
	t.Parallel()
	count := infrastructure.WalkEmbedForTest(t)
	assert.Equal(t, infrastructure.ExpectedEmbedFileCount, count,
		"if the alto-scaffold/ scaffold gained or lost a GENERIC file, update ExpectedEmbedFileCount in lockstep")
}

func TestEmbedScaffoldWriter_PlanContainsNoDotProjectMdFiles(t *testing.T) {
	// The //go:embed directive cannot do per-file exclusion, so the
	// runtime planFiles filter is the binding mechanism. Assert that the
	// post-filter writeset contains zero overlay files.
	t.Parallel()
	bad := infrastructure.PlanFilesMatchingForTest(t, func(name string) bool {
		return strings.HasSuffix(name, ".project.md")
	})
	assert.Empty(t, bad, "*.project.md overlays must not survive the runtime exclusion filter")
}

func TestEmbedScaffoldWriter_PlanContainsNoLifecycleFiles(t *testing.T) {
	// Belt-and-braces: the //go:embed allowlist omits lifecycle/ AND the
	// runtime filter rejects it. Confirm via the runtime planner.
	t.Parallel()
	bad := infrastructure.PlanFilesMatchingForTest(t, func(name string) bool {
		return strings.Contains(name, "/lifecycle/") || strings.HasPrefix(name, "alto-scaffold/lifecycle/")
	})
	assert.Empty(t, bad, "alto-scaffold/lifecycle/ content must be excluded from the writeset")
}

func TestEmbedScaffoldWriter_WriteScaffold_NoDotProjectMdFilesWritten(t *testing.T) {
	// Integration assertion: the written tree must contain zero overlays.
	t.Parallel()
	dir := t.TempDir()
	w := infrastructure.NewEmbedScaffoldWriter()
	require.NoError(t, w.WriteScaffold(context.Background(), dir, validParams(t), false))

	var bad []string
	err := filepath.WalkDir(filepath.Join(dir, "alto-scaffold"), func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return walkErr
		}
		if strings.HasSuffix(d.Name(), ".project.md") {
			bad = append(bad, path)
		}
		return nil
	})
	require.NoError(t, err)
	assert.Empty(t, bad)
}

func TestEmbedScaffoldWriter_NoSecretsInEmbed(t *testing.T) {
	t.Parallel()
	// Build-time fitness: assert no actual secrets (high-precision
	// patterns) ship in the embed. Per Phase 1 contracts §Embed
	// expectations: AWS access-key IDs, GitHub PAT tokens, and PEM
	// private-key blocks. Broad keyword scans (password|secret|token in
	// prose) are intentionally NOT used here — they false-positive on
	// legitimate documentation (e.g. white-hacker persona text
	// discussing "secrets" abstractly, brainstorm template prose).
	awsKeyRe := regexp.MustCompile(`AKIA[0-9A-Z]{16}`)
	ghTokenRe := regexp.MustCompile(`gh[pousr]_[A-Za-z0-9]{36,}`)
	pemRe := regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`)

	files := infrastructure.EmbedFilesMatchingForTest(t, func(_ string) bool { return true })
	for _, name := range files {
		content := infrastructure.EmbedReadForTest(t, name)
		if awsKeyRe.Match(content) {
			t.Errorf("AWS access-key ID in embedded file %s", name)
		}
		if ghTokenRe.Match(content) {
			t.Errorf("GitHub token in embedded file %s", name)
		}
		if pemRe.Match(content) {
			t.Errorf("PEM private-key block in embedded file %s", name)
		}
	}
}

func TestEmbedScaffoldWriter_WriteScaffold_TemplatesAreSiblingOfCommands(t *testing.T) {
	// Embed layout constraint: templates/ MUST be sibling of commands/ so
	// OpenCode adapter can resolve ${CLAUDE_SKILL_DIR}/../templates/<file>.md.
	t.Parallel()
	dir := t.TempDir()
	w := infrastructure.NewEmbedScaffoldWriter()
	require.NoError(t, w.WriteScaffold(context.Background(), dir, validParams(t), false))

	cmdInfo, err := os.Stat(filepath.Join(dir, "alto-scaffold", "commands"))
	require.NoError(t, err)
	assert.True(t, cmdInfo.IsDir())

	tplInfo, err := os.Stat(filepath.Join(dir, "alto-scaffold", "templates"))
	require.NoError(t, err)
	assert.True(t, tplInfo.IsDir())
}

// ---------------------------------------------------------------------------
// Round 1 — Fix 2: targetDir traversal guard (WH-HIGH-1)
// ---------------------------------------------------------------------------

func TestEmbedScaffoldWriter_WriteScaffold_TargetDirTraversal_Rejected(t *testing.T) {
	t.Parallel()
	w := infrastructure.NewEmbedScaffoldWriter()
	err := w.WriteScaffold(context.Background(), "../../tmp/foo", validParams(t), false)
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestEmbedScaffoldWriter_WriteScaffold_TargetDirEmbeddedTraversal_Rejected(t *testing.T) {
	t.Parallel()
	w := infrastructure.NewEmbedScaffoldWriter()
	err := w.WriteScaffold(context.Background(), "foo/../../tmp/bar", validParams(t), false)
	require.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestEmbedScaffoldWriter_WriteScaffold_TargetDirDot_Accepted(t *testing.T) {
	// CLI passes "." (init.go:145 hardcodes it). The guard MUST accept this.
	// Chdir-based; cannot run in parallel with other Chdir tests.
	dir := t.TempDir()
	prev, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(prev) })

	w := infrastructure.NewEmbedScaffoldWriter()
	require.NoError(t, w.WriteScaffold(context.Background(), ".", validParams(t), false))

	_, err = os.Stat(filepath.Join(dir, "alto-scaffold", "CONTEXT.md"))
	require.NoError(t, err)
}

func TestEmbedScaffoldWriter_WriteScaffold_TargetDirEmpty_Accepted(t *testing.T) {
	// Empty must be treated as ".".
	// Chdir-based; cannot run in parallel with other Chdir tests.
	dir := t.TempDir()
	prev, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(prev) })

	w := infrastructure.NewEmbedScaffoldWriter()
	require.NoError(t, w.WriteScaffold(context.Background(), "", validParams(t), false))

	_, err = os.Stat(filepath.Join(dir, "alto-scaffold", "CONTEXT.md"))
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Round 1 — Fix 1: symlink-overwrite defence (WH-CRIT-1)
// ---------------------------------------------------------------------------

func TestEmbedScaffoldWriter_WriteScaffold_SymlinkTarget_ForceMode_Rejected(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires elevated privileges on Windows")
	}
	t.Parallel()
	dir := t.TempDir()
	victimDir := t.TempDir()
	victim := filepath.Join(victimDir, "victim.txt")
	require.NoError(t, os.WriteFile(victim, []byte("untouchable"), 0o644))

	// Plant a symlink at <dir>/alto-scaffold/CONTEXT.md pointing at the victim.
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "alto-scaffold"), 0o755))
	require.NoError(t, os.Symlink(victim, filepath.Join(dir, "alto-scaffold", "CONTEXT.md")))

	w := infrastructure.NewEmbedScaffoldWriter()
	err := w.WriteScaffold(context.Background(), dir, validParams(t), true)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)

	// Victim must be untouched.
	content, err := os.ReadFile(victim) //nolint:gosec // test reads its own tempdir
	require.NoError(t, err)
	assert.Equal(t, "untouchable", string(content),
		"symlink-overwrite defence must not let a follow-write reach the victim")
}

func TestEmbedScaffoldWriter_WriteScaffold_SymlinkTarget_NoForceMode_RejectedByExists(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires elevated privileges on Windows")
	}
	t.Parallel()
	dir := t.TempDir()
	victimDir := t.TempDir()
	victim := filepath.Join(victimDir, "victim.txt")
	require.NoError(t, os.WriteFile(victim, []byte("untouchable"), 0o644))

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "alto-scaffold"), 0o755))
	require.NoError(t, os.Symlink(victim, filepath.Join(dir, "alto-scaffold", "CONTEXT.md")))

	w := infrastructure.NewEmbedScaffoldWriter()
	err := w.WriteScaffold(context.Background(), dir, validParams(t), false)
	// The existing alto-scaffold/ exists-check fires first in this path.
	require.ErrorIs(t, err, domainerrors.ErrAlreadyExists)

	content, err := os.ReadFile(victim) //nolint:gosec // test reads its own tempdir
	require.NoError(t, err)
	assert.Equal(t, "untouchable", string(content))
}

func TestEmbedScaffoldWriter_WriteScaffold_NoSymlinks_ForceMode_Succeeds(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "alto-scaffold"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "alto-scaffold", "CONTEXT.md"), []byte("stale"), 0o644))

	w := infrastructure.NewEmbedScaffoldWriter()
	require.NoError(t, w.WriteScaffold(context.Background(), dir, validParams(t), true))

	content, err := os.ReadFile(filepath.Join(dir, "alto-scaffold", "CONTEXT.md")) //nolint:gosec // test reads its own tempdir
	require.NoError(t, err)
	assert.NotContains(t, string(content), "stale",
		"clean force-overwrite (no symlinks) must succeed without triggering the symlink guard")
}
