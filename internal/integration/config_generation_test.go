// Package integration provides cross-context integration tests.
package integration

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
	"github.com/alty-cli/alty/internal/shared/infrastructure/persistence"
	tooltranslationdomain "github.com/alty-cli/alty/internal/tooltranslation/domain"
)

// ---------------------------------------------------------------------------
// Full Config Generation Flow Tests
// ---------------------------------------------------------------------------

func TestFullConfigGenerationFlow_ClaudeCode(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	model := makeFinalizedDomainModel(t)
	profile := vo.PythonUvProfile{}

	// Generate configs via ClaudeCodeAdapter
	adapter := tooltranslationdomain.NewClaudeCodeAdapter()
	sections := adapter.Translate(model, profile)

	// Write files using real FileWriter
	writer := persistence.NewFilesystemFileWriter()
	for _, section := range sections {
		fullPath := filepath.Join(tmpDir, section.FilePath())
		require.NoError(t, writer.WriteFile(context.Background(), fullPath, section.Content()))
	}

	// Verify expected files exist (base + 4 personas = 6 minimum)
	expectedFiles := []string{
		".claude/CLAUDE.md",
		".claude/memory/MEMORY.md",
		".claude/agents/developer.md",
		".claude/agents/tech-lead.md",
		".claude/agents/qa-engineer.md",
		".claude/agents/researcher.md",
	}
	for _, f := range expectedFiles {
		path := filepath.Join(tmpDir, f)
		assert.FileExists(t, path, "missing %s", f)
	}

	// Verify content correctness
	t.Run("CLAUDE.md has bounded contexts", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(tmpDir, ".claude/CLAUDE.md"))
		require.NoError(t, err)
		assert.Contains(t, string(content), "OrderManagement")
		assert.Contains(t, string(content), "Bounded Contexts")
	})

	t.Run("developer persona has correct frontmatter", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(tmpDir, ".claude/agents/developer.md"))
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(string(content), "---"))
		assert.Contains(t, string(content), "name: developer")
		assert.Contains(t, string(content), "permissionMode: acceptEdits")
	})

	t.Run("personas include domain model content", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(tmpDir, ".claude/agents/developer.md"))
		require.NoError(t, err)
		assert.Contains(t, string(content), "OrderManagement")
		assert.Contains(t, string(content), "Ubiquitous Language")
	})
}

func TestFullConfigGenerationFlow_Cursor(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	model := makeFinalizedDomainModel(t)
	profile := vo.PythonUvProfile{}

	adapter := tooltranslationdomain.NewCursorAdapter()
	sections := adapter.Translate(model, profile)

	writer := persistence.NewFilesystemFileWriter()
	for _, section := range sections {
		fullPath := filepath.Join(tmpDir, section.FilePath())
		require.NoError(t, writer.WriteFile(context.Background(), fullPath, section.Content()))
	}

	// Cursor produces: AGENTS.md, project-conventions.mdc, domain-expert.mdc
	expectedFiles := []string{
		"AGENTS.md",
		".cursor/rules/project-conventions.mdc",
		".cursor/rules/domain-expert.mdc",
	}
	for _, f := range expectedFiles {
		path := filepath.Join(tmpDir, f)
		assert.FileExists(t, path, "missing %s", f)
	}

	t.Run("domain-expert.mdc has ubiquitous language", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(tmpDir, ".cursor/rules/domain-expert.mdc"))
		require.NoError(t, err)
		assert.Contains(t, string(content), "order")
		assert.Contains(t, string(content), "alwaysApply: true")
	})
}

func TestFullConfigGenerationFlow_RooCode(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	model := makeFinalizedDomainModel(t)
	profile := vo.PythonUvProfile{}

	adapter := tooltranslationdomain.NewRooCodeAdapter()
	sections := adapter.Translate(model, profile)

	writer := persistence.NewFilesystemFileWriter()
	for _, section := range sections {
		fullPath := filepath.Join(tmpDir, section.FilePath())
		require.NoError(t, writer.WriteFile(context.Background(), fullPath, section.Content()))
	}

	expectedFiles := []string{
		"AGENTS.md",
		".roomodes",
		".roo/rules/project-conventions.md",
	}
	for _, f := range expectedFiles {
		path := filepath.Join(tmpDir, f)
		assert.FileExists(t, path, "missing %s", f)
	}

	t.Run(".roomodes is valid JSON with roleDefinition", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(tmpDir, ".roomodes"))
		require.NoError(t, err)

		var parsed map[string]interface{}
		require.NoError(t, json.Unmarshal(content, &parsed))
		assert.Contains(t, parsed, "customModes")

		// Verify roleDefinition exists
		assert.Contains(t, string(content), "roleDefinition")
		assert.Contains(t, string(content), "OrderManagement")
	})

	t.Run(".roomodes has context-specific modes", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(tmpDir, ".roomodes"))
		require.NoError(t, err)

		var parsed map[string]interface{}
		require.NoError(t, json.Unmarshal(content, &parsed))

		modes := parsed["customModes"].([]interface{})
		// ddd-developer + context modes
		assert.GreaterOrEqual(t, len(modes), 2)
	})
}

func TestFullConfigGenerationFlow_OpenCode(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	model := makeFinalizedDomainModel(t)
	profile := vo.PythonUvProfile{}

	adapter := tooltranslationdomain.NewOpenCodeAdapter()
	sections := adapter.Translate(model, profile)

	writer := persistence.NewFilesystemFileWriter()
	for _, section := range sections {
		fullPath := filepath.Join(tmpDir, section.FilePath())
		require.NoError(t, writer.WriteFile(context.Background(), fullPath, section.Content()))
	}

	expectedFiles := []string{
		"AGENTS.md",
		"opencode.json",
		".opencode/rules/project-conventions.md",
	}
	for _, f := range expectedFiles {
		path := filepath.Join(tmpDir, f)
		assert.FileExists(t, path, "missing %s", f)
	}

	t.Run("opencode.json has context.include", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(tmpDir, "opencode.json"))
		require.NoError(t, err)

		var parsed map[string]interface{}
		require.NoError(t, json.Unmarshal(content, &parsed))
		assert.Contains(t, parsed, "context")
	})
}

func TestFullConfigGenerationFlow_AllTools(t *testing.T) {
	t.Parallel()
	model := makeFinalizedDomainModel(t)
	profile := vo.PythonUvProfile{}

	tools := []struct {
		name         string
		adapter      tooltranslationdomain.ToolAdapter
		minSections  int
		requiredFile string
	}{
		{"ClaudeCode", tooltranslationdomain.NewClaudeCodeAdapter(), 6, ".claude/CLAUDE.md"},
		{"Cursor", tooltranslationdomain.NewCursorAdapter(), 3, "AGENTS.md"},
		{"RooCode", tooltranslationdomain.NewRooCodeAdapter(), 3, ".roomodes"},
		{"OpenCode", tooltranslationdomain.NewOpenCodeAdapter(), 3, "opencode.json"},
	}

	for _, tc := range tools {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()

			sections := tc.adapter.Translate(model, profile)
			assert.GreaterOrEqual(t, len(sections), tc.minSections,
				"%s should produce at least %d sections", tc.name, tc.minSections)

			writer := persistence.NewFilesystemFileWriter()
			for _, section := range sections {
				fullPath := filepath.Join(tmpDir, section.FilePath())
				require.NoError(t, writer.WriteFile(context.Background(), fullPath, section.Content()))
			}

			// Verify required file exists
			assert.FileExists(t, filepath.Join(tmpDir, tc.requiredFile))
		})
	}
}

// ---------------------------------------------------------------------------
// Conflict Detection Tests
// ---------------------------------------------------------------------------

func TestConfigGeneration_ConflictRename(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Pre-create a CLAUDE.md file (simulates existing user content)
	existingPath := filepath.Join(tmpDir, ".claude/CLAUDE.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(existingPath), 0o755))
	require.NoError(t, os.WriteFile(existingPath, []byte("# User's existing config"), 0o644))

	model := makeFinalizedDomainModel(t)
	profile := vo.PythonUvProfile{}

	adapter := tooltranslationdomain.NewClaudeCodeAdapter()
	sections := adapter.Translate(model, profile)

	// Use ConflictDetectingFileWriter with Rename strategy
	inner := persistence.NewFilesystemFileWriter()
	writer := persistence.NewConflictDetectingFileWriter(inner, vo.ConflictStrategyRename)

	for _, section := range sections {
		fullPath := filepath.Join(tmpDir, section.FilePath())
		require.NoError(t, writer.WriteFile(context.Background(), fullPath, section.Content()))
	}

	// Original file should be untouched
	original, err := os.ReadFile(existingPath)
	require.NoError(t, err)
	assert.Equal(t, "# User's existing config", string(original))

	// Renamed file should exist
	renamedPath := filepath.Join(tmpDir, ".claude/CLAUDE_alty.md")
	assert.FileExists(t, renamedPath)

	// Renamed file should have generated content
	renamed, err := os.ReadFile(renamedPath)
	require.NoError(t, err)
	assert.Contains(t, string(renamed), "Bounded Contexts")

	// Conflicts should be recorded
	conflicts := writer.Conflicts()
	assert.NotEmpty(t, conflicts)
}

func TestConfigGeneration_ConflictSkip(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Pre-create a CLAUDE.md file
	existingPath := filepath.Join(tmpDir, ".claude/CLAUDE.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(existingPath), 0o755))
	require.NoError(t, os.WriteFile(existingPath, []byte("# User's config"), 0o644))

	model := makeFinalizedDomainModel(t)
	profile := vo.PythonUvProfile{}

	adapter := tooltranslationdomain.NewClaudeCodeAdapter()
	sections := adapter.Translate(model, profile)

	// Use ConflictDetectingFileWriter with Skip strategy
	inner := persistence.NewFilesystemFileWriter()
	writer := persistence.NewConflictDetectingFileWriter(inner, vo.ConflictStrategySkip)

	for _, section := range sections {
		fullPath := filepath.Join(tmpDir, section.FilePath())
		require.NoError(t, writer.WriteFile(context.Background(), fullPath, section.Content()))
	}

	// Original file should be untouched
	original, err := os.ReadFile(existingPath)
	require.NoError(t, err)
	assert.Equal(t, "# User's config", string(original))

	// No renamed file should exist
	renamedPath := filepath.Join(tmpDir, ".claude/CLAUDE_alty.md")
	assert.NoFileExists(t, renamedPath)

	// Conflicts should still be recorded
	conflicts := writer.Conflicts()
	assert.NotEmpty(t, conflicts)
}

// ---------------------------------------------------------------------------
// Empty Model Edge Case
// ---------------------------------------------------------------------------

func TestConfigGeneration_EmptyModel(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Empty model (no bounded contexts, no terms)
	model := ddd.NewDomainModel("empty-model")
	profile := vo.GenericProfile{}

	adapter := tooltranslationdomain.NewClaudeCodeAdapter()
	sections := adapter.Translate(model, profile)

	// Should still produce base files + personas
	assert.GreaterOrEqual(t, len(sections), 6)

	writer := persistence.NewFilesystemFileWriter()
	for _, section := range sections {
		fullPath := filepath.Join(tmpDir, section.FilePath())
		require.NoError(t, writer.WriteFile(context.Background(), fullPath, section.Content()))
	}

	// Base files should exist
	assert.FileExists(t, filepath.Join(tmpDir, ".claude/CLAUDE.md"))
	assert.FileExists(t, filepath.Join(tmpDir, ".claude/agents/developer.md"))

	// Content should have empty sections (no crash)
	content, err := os.ReadFile(filepath.Join(tmpDir, ".claude/CLAUDE.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "Bounded Contexts")
}

func TestConfigGeneration_CursorSkipsDomainExpertWhenNoTerms(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Empty model (no terms)
	model := ddd.NewDomainModel("no-terms")
	profile := vo.GenericProfile{}

	adapter := tooltranslationdomain.NewCursorAdapter()
	sections := adapter.Translate(model, profile)

	// Should have only 2 sections (no domain-expert.mdc)
	assert.Len(t, sections, 2)

	writer := persistence.NewFilesystemFileWriter()
	for _, section := range sections {
		fullPath := filepath.Join(tmpDir, section.FilePath())
		require.NoError(t, writer.WriteFile(context.Background(), fullPath, section.Content()))
	}

	// domain-expert.mdc should NOT exist
	assert.NoFileExists(t, filepath.Join(tmpDir, ".cursor/rules/domain-expert.mdc"))
}
