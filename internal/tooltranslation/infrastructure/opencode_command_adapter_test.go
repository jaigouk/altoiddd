package infrastructure

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ttapp "github.com/alto-cli/alto/internal/tooltranslation/application"
	ttdomain "github.com/alto-cli/alto/internal/tooltranslation/domain"
)

func TestOpenCodeCommandAdapter_CompileTimeAssertion(t *testing.T) {
	t.Parallel()
	var _ ttapp.WorkflowAssetGeneration = (*OpenCodeCommandAdapter)(nil)
}

// setupAltoFixture creates a tempdir with .alto/commands/ + .alto/templates/
// layout and returns (altoDir, projectRoot).
func setupAltoFixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	altoDir := filepath.Join(root, ".alto")
	commandsDir := filepath.Join(altoDir, "commands")
	tmplDir := filepath.Join(altoDir, "templates")
	require.NoError(t, os.MkdirAll(commandsDir, 0o755))
	require.NoError(t, os.MkdirAll(tmplDir, 0o755))
	return commandsDir, root
}

func TestOpenCodeCommandAdapter_RendersSimpleCommand(t *testing.T) {
	t.Parallel()
	commandsDir, projectRoot := setupAltoFixture(t)
	writeSource(t, commandsDir, "foo.md", validFrontmatter+"# Foo\n\nbody.\n")

	a := NewOpenCodeCommandAdapter()
	require.NoError(t, a.GenerateFromAssets(context.TODO(), commandsDir, projectRoot))

	out, err := os.ReadFile(filepath.Join(projectRoot, ".opencode", "commands", "foo.md"))
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "description: test command")
	assert.NotContains(t, s, "agent:", "agent must NOT be emitted when source has none")
	assert.Contains(t, s, "# Foo")
}

func TestOpenCodeCommandAdapter_EmitsAgentOnlyWhenPresent(t *testing.T) {
	t.Parallel()
	commandsDir, projectRoot := setupAltoFixture(t)
	withAgent := strings.Replace(validFrontmatter, "license: Apache-2.0\n---", "license: Apache-2.0\nagent: build\n---", 1)
	writeSource(t, commandsDir, "foo.md", withAgent+"# Foo\n")

	a := NewOpenCodeCommandAdapter()
	require.NoError(t, a.GenerateFromAssets(context.TODO(), commandsDir, projectRoot))
	out, err := os.ReadFile(filepath.Join(projectRoot, ".opencode", "commands", "foo.md"))
	require.NoError(t, err)
	assert.Contains(t, string(out), "agent: build")
}

func TestOpenCodeCommandAdapter_DisableModelInvocation_SkippedAndAggregated(t *testing.T) {
	t.Parallel()
	commandsDir, projectRoot := setupAltoFixture(t)
	protected := strings.Replace(validFrontmatter, "license: Apache-2.0\n---", "license: Apache-2.0\ndisable_model_invocation: true\n---", 1)
	writeSource(t, commandsDir, "secret.md", protected+"# Secret\n")

	a := NewOpenCodeCommandAdapter()
	err := a.GenerateFromAssets(context.TODO(), commandsDir, projectRoot)
	require.Error(t, err)
	require.ErrorIs(t, err, ttdomain.ErrInvocationProtectionNotSupported)

	// File must NOT be emitted.
	_, statErr := os.Stat(filepath.Join(projectRoot, ".opencode", "commands", "secret.md"))
	assert.True(t, os.IsNotExist(statErr), "protected command must not be emitted")
}

func TestOpenCodeCommandAdapter_DisableModelInvocationIsNonFatal_OthersStillRender(t *testing.T) {
	t.Parallel()
	commandsDir, projectRoot := setupAltoFixture(t)
	protected := strings.Replace(validFrontmatter, "license: Apache-2.0\n---", "license: Apache-2.0\ndisable_model_invocation: true\n---", 1)
	writeSource(t, commandsDir, "secret.md", protected+"# Secret\n")
	writeSource(t, commandsDir, "ok.md", strings.Replace(validFrontmatter, "name: foo", "name: ok", 1)+"# Ok\n")

	a := NewOpenCodeCommandAdapter()
	err := a.GenerateFromAssets(context.TODO(), commandsDir, projectRoot)
	require.Error(t, err) // aggregated non-fatal — error returned but other commands proceeded.
	require.ErrorIs(t, err, ttdomain.ErrInvocationProtectionNotSupported)

	_, statErr := os.Stat(filepath.Join(projectRoot, ".opencode", "commands", "ok.md"))
	assert.NoError(t, statErr, "non-protected command must render despite sibling protection")
}

func TestOpenCodeCommandAdapter_StripsInlineBashBlocks(t *testing.T) {
	t.Parallel()
	commandsDir, projectRoot := setupAltoFixture(t)
	body := "Pre: !`ls -la` Post"
	writeSource(t, commandsDir, "foo.md", validFrontmatter+body+"\n")
	a := NewOpenCodeCommandAdapter()
	require.NoError(t, a.GenerateFromAssets(context.TODO(), commandsDir, projectRoot))
	out, _ := os.ReadFile(filepath.Join(projectRoot, ".opencode", "commands", "foo.md"))
	s := string(out)
	assert.Contains(t, s, "<!--")
	assert.Contains(t, s, "ls -la")
	assert.Contains(t, s, "stripped")
	// Original command is preserved verbatim inside the HTML comment per
	// the binding AC ("naming the original command verbatim — no silent drop").
}

func TestOpenCodeCommandAdapter_InlinesTemplateRefs(t *testing.T) {
	t.Parallel()
	commandsDir, projectRoot := setupAltoFixture(t)
	tmplPath := filepath.Join(projectRoot, ".alto", "templates", "epic.md")
	require.NoError(t, os.WriteFile(tmplPath, []byte("EPIC_TEMPLATE_INLINED"), 0o600))
	body := "Use ${CLAUDE_SKILL_DIR}/../templates/epic.md to draft.\n"
	writeSource(t, commandsDir, "foo.md", validFrontmatter+body)

	a := NewOpenCodeCommandAdapter()
	require.NoError(t, a.GenerateFromAssets(context.TODO(), commandsDir, projectRoot))
	out, _ := os.ReadFile(filepath.Join(projectRoot, ".opencode", "commands", "foo.md"))
	assert.Contains(t, string(out), "EPIC_TEMPLATE_INLINED")
	assert.NotContains(t, string(out), "${CLAUDE_SKILL_DIR}")
}

func TestOpenCodeCommandAdapter_MissingTemplate_AggregatesErrMissingTemplate(t *testing.T) {
	t.Parallel()
	commandsDir, projectRoot := setupAltoFixture(t)
	body := "Ref ${CLAUDE_SKILL_DIR}/../templates/nope.md"
	writeSource(t, commandsDir, "foo.md", validFrontmatter+body+"\n")

	a := NewOpenCodeCommandAdapter()
	err := a.GenerateFromAssets(context.TODO(), commandsDir, projectRoot)
	require.Error(t, err)
	require.ErrorIs(t, err, ttdomain.ErrMissingTemplate)

	out, _ := os.ReadFile(filepath.Join(projectRoot, ".opencode", "commands", "foo.md"))
	assert.Contains(t, string(out), "missing template nope.md")
}

func TestOpenCodeCommandAdapter_AtomicWrite_NoTempfileLeak(t *testing.T) {
	t.Parallel()
	commandsDir, projectRoot := setupAltoFixture(t)
	writeSource(t, commandsDir, "foo.md", validFrontmatter+"body\n")
	a := NewOpenCodeCommandAdapter()
	require.NoError(t, a.GenerateFromAssets(context.TODO(), commandsDir, projectRoot))

	outDir := filepath.Join(projectRoot, ".opencode", "commands")
	entries, err := os.ReadDir(outDir)
	require.NoError(t, err)
	for _, e := range entries {
		assert.False(t, strings.HasPrefix(e.Name(), "."), "no dotfile/tempfile leftovers, got %s", e.Name())
	}
}

func TestOpenCodeCommandAdapter_InvalidName_ReturnsErrInvalidAssetName(t *testing.T) {
	t.Parallel()
	commandsDir, projectRoot := setupAltoFixture(t)
	bad := strings.Replace(validFrontmatter, "name: foo", "name: BAD", 1)
	writeSource(t, commandsDir, "foo.md", bad+"body\n")

	a := NewOpenCodeCommandAdapter()
	err := a.GenerateFromAssets(context.TODO(), commandsDir, projectRoot)
	require.Error(t, err)
	assert.ErrorIs(t, err, ttdomain.ErrInvalidAssetName)
}

func TestOpenCodeCommandAdapter_OrphanOverlay_ReturnsErrOrphanOverlay(t *testing.T) {
	t.Parallel()
	commandsDir, projectRoot := setupAltoFixture(t)
	writeSource(t, commandsDir, "orphan.project.md", validFrontmatter+"overlay body")

	a := NewOpenCodeCommandAdapter()
	err := a.GenerateFromAssets(context.TODO(), commandsDir, projectRoot)
	require.Error(t, err)
	assert.ErrorIs(t, err, ttdomain.ErrOrphanOverlay)
}

func TestOpenCodeCommandAdapter_OverlayMerge_BodyAppended(t *testing.T) {
	t.Parallel()
	commandsDir, projectRoot := setupAltoFixture(t)
	writeSource(t, commandsDir, "foo.md", validFrontmatter+"primary body\n")
	writeSource(t, commandsDir, "foo.project.md", validFrontmatter+"overlay body\n")

	a := NewOpenCodeCommandAdapter()
	require.NoError(t, a.GenerateFromAssets(context.TODO(), commandsDir, projectRoot))

	out, _ := os.ReadFile(filepath.Join(projectRoot, ".opencode", "commands", "foo.md"))
	s := string(out)
	primaryIdx := strings.Index(s, "primary body")
	overlayIdx := strings.Index(s, "overlay body")
	require.GreaterOrEqual(t, primaryIdx, 0)
	require.GreaterOrEqual(t, overlayIdx, 0)
	assert.Less(t, primaryIdx, overlayIdx, "overlay must follow primary")
}

func TestOpenCodeCommandAdapter_DescriptionWithColon_QuotedSafely(t *testing.T) {
	t.Parallel()
	commandsDir, projectRoot := setupAltoFixture(t)
	src := strings.Replace(validFrontmatter, "description: test command", `description: "with: colon"`, 1)
	writeSource(t, commandsDir, "foo.md", src+"body\n")

	a := NewOpenCodeCommandAdapter()
	require.NoError(t, a.GenerateFromAssets(context.TODO(), commandsDir, projectRoot))
	out, _ := os.ReadFile(filepath.Join(projectRoot, ".opencode", "commands", "foo.md"))
	assert.Contains(t, string(out), `description: "with: colon"`)
}
