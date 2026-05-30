package infrastructure

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ttdomain "github.com/alto-cli/alto/internal/tooltranslation/domain"
)

func writeSource(t *testing.T, dir, name, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
}

const validFrontmatter = `---
name: foo
description: test command
kind: command
phase: groom
when_to_use: when testing
tools_required: Read, Bash
bash_substitution_policy: none
license: Apache-2.0
---
`

func TestParseFrontmatter_ValidYAML(t *testing.T) {
	t.Parallel()
	fm, body, err := parseFrontmatter(validFrontmatter + "\n# Body\nhello\n")
	require.NoError(t, err)
	assert.Equal(t, "foo", fm.Name)
	assert.Equal(t, "command", fm.Kind)
	assert.Contains(t, body, "# Body")
}

func TestParseFrontmatter_MissingDelimiter_ReturnsErrInvalidFrontmatter(t *testing.T) {
	t.Parallel()
	_, _, err := parseFrontmatter("no frontmatter here")
	assert.ErrorIs(t, err, ttdomain.ErrInvalidFrontmatter)
}

func TestParseFrontmatter_UnclosedFrontmatter_ReturnsErrInvalidFrontmatter(t *testing.T) {
	t.Parallel()
	_, _, err := parseFrontmatter("---\nname: foo\n")
	assert.ErrorIs(t, err, ttdomain.ErrInvalidFrontmatter)
}

func TestParseFrontmatter_DisableModelInvocationTrue_ParsedCorrectly(t *testing.T) {
	t.Parallel()
	src := `---
name: foo
description: x
kind: command
phase: groom
when_to_use: x
tools_required: Read
bash_substitution_policy: none
license: Apache-2.0
disable_model_invocation: true
---
body`
	fm, _, err := parseFrontmatter(src)
	require.NoError(t, err)
	assert.True(t, fm.DisableModelInvocation)
}

func TestParseFrontmatter_AgentPresentEvenWhenEmpty_HasAgentTrue(t *testing.T) {
	t.Parallel()
	src := `---
name: foo
description: x
kind: command
phase: groom
when_to_use: x
tools_required: Read
bash_substitution_policy: none
license: Apache-2.0
agent: ""
---
body`
	fm, _, err := parseFrontmatter(src)
	require.NoError(t, err)
	assert.True(t, fm.HasAgent, "HasAgent should detect the field's presence even with empty value")
}

func TestParseFrontmatter_AgentAbsent_HasAgentFalse(t *testing.T) {
	t.Parallel()
	fm, _, err := parseFrontmatter(validFrontmatter + "body")
	require.NoError(t, err)
	assert.False(t, fm.HasAgent)
}

func TestLoadAssetSource_NoOverlay(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSource(t, dir, "foo.md", validFrontmatter+"primary body\n")
	src, err := loadAssetSource(dir, "foo")
	require.NoError(t, err)
	assert.Equal(t, "foo", src.name)
	assert.Contains(t, src.body, "primary body")
}

func TestLoadAssetSource_OverlayAppendedNewlineSeparated(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSource(t, dir, "foo.md", validFrontmatter+"primary\n")
	writeSource(t, dir, "foo.project.md", validFrontmatter+"overlay\n")
	src, err := loadAssetSource(dir, "foo")
	require.NoError(t, err)
	assert.Contains(t, src.body, "primary")
	assert.Contains(t, src.body, "overlay")
	assert.Less(t, strings.Index(src.body, "primary"), strings.Index(src.body, "overlay"),
		"overlay must be appended after primary")
}

func TestLoadAssetSource_InvalidName_ReturnsErrInvalidAssetName(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	bad := `---
name: BadName
description: x
kind: command
phase: groom
when_to_use: x
tools_required: Read
bash_substitution_policy: none
license: Apache-2.0
---
body`
	writeSource(t, dir, "foo.md", bad)
	_, err := loadAssetSource(dir, "foo")
	assert.ErrorIs(t, err, ttdomain.ErrInvalidAssetName)
}

func TestStripBashBlocks_InlineBlock_ReplacedWithComment(t *testing.T) {
	t.Parallel()
	got := stripBashBlocks("Run !`ls -la` now")
	assert.Contains(t, got, "<!--")
	assert.Contains(t, got, "ls -la")
	assert.Contains(t, got, "stripped")
	// The substitution is wrapped in an HTML comment — markdown/Claude
	// Code clients don't re-execute commands inside `<!-- ... -->`. The
	// AC preserves the original command verbatim, so the literal `!`ls
	// -la`` survives INSIDE the comment; that's by design.
}

func TestStripBashBlocks_FencedBlock_ReplacedWithComment(t *testing.T) {
	t.Parallel()
	in := "before\n```!\nfind . -name foo\n```\nafter"
	got := stripBashBlocks(in)
	assert.Contains(t, got, "<!--")
	assert.Contains(t, got, "find . -name foo")
	assert.Contains(t, got, "stripped")
}

func TestStripBashBlocks_OriginalCommandPreservedVerbatim(t *testing.T) {
	t.Parallel()
	got := stripBashBlocks("!`echo \"hello world\"`")
	assert.Contains(t, got, `echo "hello world"`)
}

func TestInlineTemplateRefs_PresentTemplate_Inlined(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	tmplDir := filepath.Join(root, "templates")
	require.NoError(t, os.MkdirAll(tmplDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tmplDir, "epic.md"), []byte("TEMPLATE_CONTENT"), 0o600))

	body := "See ${CLAUDE_SKILL_DIR}/../templates/epic.md for details."
	got, errs := inlineTemplateRefs(body, tmplDir)
	assert.Empty(t, errs)
	assert.Contains(t, got, "TEMPLATE_CONTENT")
	assert.NotContains(t, got, "${CLAUDE_SKILL_DIR}")
}

func TestInlineTemplateRefs_MissingTemplate_ReturnsErrMissingTemplate(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	body := "Ref ${CLAUDE_SKILL_DIR}/../templates/nope.md"
	got, errs := inlineTemplateRefs(body, root)
	assert.Contains(t, got, "missing template nope.md")
	require.Len(t, errs, 1)
	assert.ErrorIs(t, errs[0], ttdomain.ErrMissingTemplate)
}

func TestSafeJoin_RejectsTraversal(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	_, err := safeJoin(root, "../escape.md")
	assert.ErrorIs(t, err, ttdomain.ErrPathTraversal)
}

func TestSafeJoin_RejectsEmptyName(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	_, err := safeJoin(root, "")
	assert.ErrorIs(t, err, ttdomain.ErrPathTraversal)
}

func TestSafeJoin_AcceptsLegitName(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	got, err := safeJoin(root, "ok.md")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(root, "ok.md"), got)
}

func TestWriteAtomic_CreatesFileWithContent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "out.md")
	require.NoError(t, writeAtomic(target, []byte("hello")))
	got, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(got))
}

func TestWriteAtomic_LeavesNoTempfileBehindOnSuccess(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "out.md")
	require.NoError(t, writeAtomic(target, []byte("data")))
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	for _, e := range entries {
		assert.False(t, strings.HasPrefix(e.Name(), ".workflow-asset-"),
			"tempfile %s leaked into output dir", e.Name())
	}
}

func TestListPrimaryAssetNames_OrphanOverlay_ReturnsErrOrphanOverlay(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeSource(t, dir, "foo.md", validFrontmatter+"body")
	writeSource(t, dir, "orphan.project.md", validFrontmatter+"overlay")
	names, errs, err := listPrimaryAssetNames(dir)
	require.NoError(t, err)
	assert.Equal(t, []string{"foo"}, names)
	require.Len(t, errs, 1)
	assert.ErrorIs(t, errs[0], ttdomain.ErrOrphanOverlay)
}

func TestAggregateErrors_NilOnEmpty(t *testing.T) {
	t.Parallel()
	assert.NoError(t, aggregateErrors(nil))
	assert.NoError(t, aggregateErrors([]error{}))
}

func TestAggregateErrors_JoinsPreservesIsCheck(t *testing.T) {
	t.Parallel()
	joined := aggregateErrors([]error{
		ttdomain.ErrInvocationProtectionNotSupported,
		ttdomain.ErrMissingTemplate,
	})
	require.Error(t, joined)
	require.ErrorIs(t, joined, ttdomain.ErrInvocationProtectionNotSupported)
	require.ErrorIs(t, joined, ttdomain.ErrMissingTemplate)
}
