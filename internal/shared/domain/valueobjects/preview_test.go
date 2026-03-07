package valueobjects_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// FileAction
// ---------------------------------------------------------------------------

func TestFileActionCreate(t *testing.T) {
	t.Parallel()
	action := vo.NewFileAction("docs/PRD.md", vo.FileActionCreate, "", "")
	assert.Equal(t, "docs/PRD.md", action.Path())
	assert.Equal(t, vo.FileActionCreate, action.ActionType())
	assert.Empty(t, action.Reason())
	assert.Empty(t, action.RenamedPath())
}

func TestFileActionSkipForExistingFile(t *testing.T) {
	t.Parallel()
	action := vo.NewFileAction("README.md", vo.FileActionSkip, "already exists", "")
	assert.Equal(t, vo.FileActionSkip, action.ActionType())
	assert.Equal(t, "already exists", action.Reason())
}

func TestFileActionConflictRenameUsesAltySuffix(t *testing.T) {
	t.Parallel()
	action := vo.NewFileAction(
		".claude/CLAUDE.md",
		vo.FileActionConflictRename,
		"existing file conflicts with template",
		".claude/CLAUDE_alty.md",
	)
	assert.Equal(t, vo.FileActionConflictRename, action.ActionType())
	assert.Equal(t, ".claude/CLAUDE_alty.md", action.RenamedPath())
	assert.Contains(t, action.RenamedPath(), "_alty")
}

// ---------------------------------------------------------------------------
// Preview
// ---------------------------------------------------------------------------

func TestPreviewContainsFileActions(t *testing.T) {
	t.Parallel()
	actions := []vo.FileAction{
		vo.NewFileAction("docs/PRD.md", vo.FileActionCreate, "", ""),
		vo.NewFileAction("README.md", vo.FileActionSkip, "already exists", ""),
	}
	preview := vo.NewPreview(actions, nil, nil)
	assert.Len(t, preview.FileActions(), 2)
	assert.Equal(t, vo.FileActionCreate, preview.FileActions()[0].ActionType())
	assert.Equal(t, vo.FileActionSkip, preview.FileActions()[1].ActionType())
}

func TestPreviewNeverHasOverwriteAction(t *testing.T) {
	t.Parallel()
	actionNames := vo.AllFileActionTypes()
	for _, a := range actionNames {
		assert.NotEqual(t, "overwrite", string(a))
	}
}

func TestPreviewDefaultConflictDescriptionsEmpty(t *testing.T) {
	t.Parallel()
	preview := vo.NewPreview(
		[]vo.FileAction{vo.NewFileAction("a.txt", vo.FileActionCreate, "", "")},
		nil, nil,
	)
	assert.Empty(t, preview.ConflictDescriptions())
}

func TestPreviewWithConflictDescriptions(t *testing.T) {
	t.Parallel()
	preview := vo.NewPreview(
		[]vo.FileAction{vo.NewFileAction("a.txt", vo.FileActionCreate, "", "")},
		nil,
		[]string{"Global cursor setting overrides local"},
	)
	assert.Len(t, preview.ConflictDescriptions(), 1)
	assert.Contains(t, preview.ConflictDescriptions()[0], "cursor")
}

// ---------------------------------------------------------------------------
// GlobalSettingConflict
// ---------------------------------------------------------------------------

func TestGlobalSettingConflictHasToolAndPaths(t *testing.T) {
	t.Parallel()
	conflict := vo.NewGlobalSettingConflict(
		"cursor",
		"/home/user/.cursor/settings.json",
		`{"theme": "dark"}`,
		`{"theme": "light"}`,
	)
	assert.Equal(t, "cursor", conflict.Tool())
	assert.Contains(t, conflict.GlobalPath(), "settings.json")
	assert.NotEqual(t, conflict.GlobalValue(), conflict.LocalValue())
}

func TestConflictResolutionOptions(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "keep_global", string(vo.ConflictKeepGlobal))
	assert.Equal(t, "update_global", string(vo.ConflictUpdateGlobal))
	assert.Equal(t, "set_local_with_warning", string(vo.ConflictSetLocalWithWarning))
}
