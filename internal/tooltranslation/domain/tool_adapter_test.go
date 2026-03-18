package domain_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
	"github.com/alto-cli/alto/internal/tooltranslation/domain"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeModel(t *testing.T) *ddd.DomainModel {
	t.Helper()
	return makeModelWithContexts(t, []struct {
		name           string
		classification vo.SubdomainClassification
	}{{"Orders", vo.SubdomainCore}})
}

func makeMultiContextModel(t *testing.T) *ddd.DomainModel {
	t.Helper()
	return makeModelWithContexts(t, []struct {
		name           string
		classification vo.SubdomainClassification
	}{
		{"Orders", vo.SubdomainCore},
		{"Notifications", vo.SubdomainSupporting},
	})
}

func sectionByPath(sections []domain.ConfigSection, path string) *domain.ConfigSection {
	for _, s := range sections {
		if s.FilePath() == path {
			return &s
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// ClaudeCodeAdapter
// ---------------------------------------------------------------------------

func TestClaudeCodeAdapter(t *testing.T) {
	t.Parallel()
	profile := vo.PythonUvProfile{}

	t.Run("produces claude md and memory md", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		adapter := domain.NewClaudeCodeAdapter()
		sections := adapter.Translate(m, profile)
		assert.GreaterOrEqual(t, len(sections), 6)
		assert.Equal(t, ".claude/CLAUDE.md", sections[0].FilePath())
		assert.Equal(t, ".claude/memory/MEMORY.md", sections[1].FilePath())
	})

	t.Run("content includes ubiquitous language", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewClaudeCodeAdapter().Translate(m, profile)
		assert.Contains(t, sections[0].Content(), "Orders")
		assert.Contains(t, sections[0].Content(), "Ubiquitous Language")
	})

	t.Run("content includes bounded contexts", func(t *testing.T) {
		t.Parallel()
		m := makeMultiContextModel(t)
		sections := domain.NewClaudeCodeAdapter().Translate(m, profile)
		assert.Contains(t, sections[0].Content(), "Orders")
		assert.Contains(t, sections[0].Content(), "Notifications")
		assert.Contains(t, sections[0].Content(), "Bounded Contexts")
	})

	t.Run("content includes ddd layer rules", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewClaudeCodeAdapter().Translate(m, profile)
		assert.Contains(t, sections[0].Content(), "DDD Layer Rules")
	})

	t.Run("includes after close protocol", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewClaudeCodeAdapter().Translate(m, profile)
		content := sections[0].Content()
		assert.Contains(t, content, "After-Close Protocol")
		assert.Contains(t, content, "bd-ripple")
		assert.Contains(t, content, "review_needed")
		assert.Contains(t, content, "bd ready")
	})

	t.Run("follow up step", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewClaudeCodeAdapter().Translate(m, profile)
		content := sections[0].Content()
		assert.True(t, strings.Contains(content, "Follow-up") || strings.Contains(content, "follow-up"))
	})
}

// ---------------------------------------------------------------------------
// ClaudeCodeAdapter — Agent Personas
// ---------------------------------------------------------------------------

func TestClaudeCodeAdapter_AgentPersonas(t *testing.T) {
	t.Parallel()
	profile := vo.PythonUvProfile{}

	t.Run("produces agent persona files", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewClaudeCodeAdapter().Translate(m, profile)

		// Base 2 + 4 personas = 6 total
		assert.GreaterOrEqual(t, len(sections), 6)

		// Check developer persona exists
		dev := sectionByPath(sections, ".claude/agents/developer.md")
		require.NotNil(t, dev, "missing developer.md")
		assert.Contains(t, dev.Content(), "name: developer")
	})

	t.Run("developer persona has bounded contexts", func(t *testing.T) {
		t.Parallel()
		m := makeMultiContextModel(t)
		sections := domain.NewClaudeCodeAdapter().Translate(m, profile)
		dev := sectionByPath(sections, ".claude/agents/developer.md")
		require.NotNil(t, dev)
		assert.Contains(t, dev.Content(), "Orders")
		assert.Contains(t, dev.Content(), "Notifications")
	})

	t.Run("personas contain ubiquitous language", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewClaudeCodeAdapter().Translate(m, profile)
		dev := sectionByPath(sections, ".claude/agents/developer.md")
		require.NotNil(t, dev)
		assert.Contains(t, dev.Content(), "Ubiquitous Language")
	})

	t.Run("personas have yaml frontmatter", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewClaudeCodeAdapter().Translate(m, profile)
		for _, persona := range []string{"developer", "tech-lead", "qa-engineer", "researcher"} {
			path := ".claude/agents/" + persona + ".md"
			sec := sectionByPath(sections, path)
			require.NotNil(t, sec, "missing %s", path)
			assert.True(t, strings.HasPrefix(sec.Content(), "---"), "%s missing YAML frontmatter", persona)
		}
	})

	t.Run("personas include quality gates when profile has them", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewClaudeCodeAdapter().Translate(m, vo.PythonUvProfile{})
		dev := sectionByPath(sections, ".claude/agents/developer.md")
		require.NotNil(t, dev)
		assert.Contains(t, dev.Content(), "Quality Gates")
	})

	t.Run("personas omit quality gates for generic profile", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewClaudeCodeAdapter().Translate(m, vo.GenericProfile{})
		dev := sectionByPath(sections, ".claude/agents/developer.md")
		require.NotNil(t, dev)
		assert.NotContains(t, dev.Content(), "Quality Gates")
	})
}

// ---------------------------------------------------------------------------
// CursorAdapter
// ---------------------------------------------------------------------------

func TestCursorAdapter(t *testing.T) {
	t.Parallel()
	profile := vo.PythonUvProfile{}

	t.Run("produces at least two sections", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewCursorAdapter().Translate(m, profile)
		assert.GreaterOrEqual(t, len(sections), 2)
		paths := make([]string, len(sections))
		for i, s := range sections {
			paths[i] = s.FilePath()
		}
		assert.Contains(t, paths, "AGENTS.md")
		assert.Contains(t, paths, ".cursor/rules/project-conventions.mdc")
	})

	t.Run("mdc has frontmatter", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewCursorAdapter().Translate(m, profile)
		mdc := sectionByPath(sections, ".cursor/rules/project-conventions.mdc")
		require.NotNil(t, mdc)
		assert.True(t, strings.HasPrefix(mdc.Content(), "---"))
	})
}

// ---------------------------------------------------------------------------
// CursorAdapter — Domain Expert Rule
// ---------------------------------------------------------------------------

func TestCursorAdapter_DomainExpertRule(t *testing.T) {
	t.Parallel()
	profile := vo.PythonUvProfile{}

	t.Run("produces domain-expert.mdc with terms", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewCursorAdapter().Translate(m, profile)

		expert := sectionByPath(sections, ".cursor/rules/domain-expert.mdc")
		require.NotNil(t, expert, "missing domain-expert.mdc")
		assert.Contains(t, expert.Content(), "Domain Expert Reference")
		assert.Contains(t, expert.Content(), "Orders")
	})

	t.Run("domain-expert has mdc frontmatter", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewCursorAdapter().Translate(m, profile)
		expert := sectionByPath(sections, ".cursor/rules/domain-expert.mdc")
		require.NotNil(t, expert)
		assert.True(t, strings.HasPrefix(expert.Content(), "---"))
		assert.Contains(t, expert.Content(), "alwaysApply: true")
	})

	t.Run("skips domain-expert when no terms", func(t *testing.T) {
		t.Parallel()
		m := ddd.NewDomainModel("empty")
		sections := domain.NewCursorAdapter().Translate(m, profile)
		expert := sectionByPath(sections, ".cursor/rules/domain-expert.mdc")
		assert.Nil(t, expert, "should not produce domain-expert.mdc for empty model")
	})
}

// ---------------------------------------------------------------------------
// RooCodeAdapter
// ---------------------------------------------------------------------------

func TestRooCodeAdapter(t *testing.T) {
	t.Parallel()
	profile := vo.PythonUvProfile{}

	t.Run("produces three sections", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewRooCodeAdapter().Translate(m, profile)
		assert.Len(t, sections, 3)
		paths := make([]string, len(sections))
		for i, s := range sections {
			paths[i] = s.FilePath()
		}
		assert.Contains(t, paths, "AGENTS.md")
		assert.Contains(t, paths, ".roomodes")
		assert.Contains(t, paths, ".roo/rules/project-conventions.md")
	})

	t.Run("roomodes is valid json", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewRooCodeAdapter().Translate(m, profile)
		roomodes := sectionByPath(sections, ".roomodes")
		require.NotNil(t, roomodes)
		var parsed map[string]interface{}
		err := json.Unmarshal([]byte(roomodes.Content()), &parsed)
		require.NoError(t, err)
		assert.Contains(t, parsed, "customModes")
	})
}

// ---------------------------------------------------------------------------
// RooCodeAdapter — Enhanced Roomodes
// ---------------------------------------------------------------------------

func TestRooCodeAdapter_EnhancedRoomodes(t *testing.T) {
	t.Parallel()
	profile := vo.PythonUvProfile{}

	t.Run("roomodes has roleDefinition", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewRooCodeAdapter().Translate(m, profile)
		roomodes := sectionByPath(sections, ".roomodes")
		require.NotNil(t, roomodes)
		assert.Contains(t, roomodes.Content(), "roleDefinition")
	})

	t.Run("roleDefinition contains bounded contexts", func(t *testing.T) {
		t.Parallel()
		m := makeMultiContextModel(t)
		sections := domain.NewRooCodeAdapter().Translate(m, profile)
		roomodes := sectionByPath(sections, ".roomodes")
		require.NotNil(t, roomodes)
		assert.Contains(t, roomodes.Content(), "Orders")
		assert.Contains(t, roomodes.Content(), "Notifications")
	})

	t.Run("generates context-specific modes", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewRooCodeAdapter().Translate(m, profile)
		roomodes := sectionByPath(sections, ".roomodes")
		require.NotNil(t, roomodes)

		var parsed map[string]interface{}
		err := json.Unmarshal([]byte(roomodes.Content()), &parsed)
		require.NoError(t, err)

		modes := parsed["customModes"].([]interface{})
		assert.GreaterOrEqual(t, len(modes), 2)
	})

	t.Run("context mode has classification", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewRooCodeAdapter().Translate(m, profile)
		roomodes := sectionByPath(sections, ".roomodes")
		require.NotNil(t, roomodes)
		assert.Contains(t, roomodes.Content(), "core")
	})
}

// ---------------------------------------------------------------------------
// OpenCodeAdapter
// ---------------------------------------------------------------------------

func TestOpenCodeAdapter(t *testing.T) {
	t.Parallel()
	profile := vo.PythonUvProfile{}

	t.Run("produces three sections", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewOpenCodeAdapter().Translate(m, profile)
		assert.Len(t, sections, 3)
		paths := make([]string, len(sections))
		for i, s := range sections {
			paths[i] = s.FilePath()
		}
		assert.Contains(t, paths, "AGENTS.md")
		assert.Contains(t, paths, "opencode.json")
		assert.Contains(t, paths, ".opencode/rules/project-conventions.md")
	})

	t.Run("opencode json is valid", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewOpenCodeAdapter().Translate(m, profile)
		ocJSON := sectionByPath(sections, "opencode.json")
		require.NotNil(t, ocJSON)
		var parsed map[string]interface{}
		err := json.Unmarshal([]byte(ocJSON.Content()), &parsed)
		require.NoError(t, err)
		assert.Contains(t, parsed, "context")
	})
}

// ---------------------------------------------------------------------------
// Adapter content tests
// ---------------------------------------------------------------------------

func TestAdapterContent(t *testing.T) {
	t.Parallel()
	profile := vo.PythonUvProfile{}

	t.Run("agents md includes terms", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewCursorAdapter().Translate(m, profile)
		agents := sectionByPath(sections, "AGENTS.md")
		require.NotNil(t, agents)
		assert.Contains(t, agents.Content(), "Orders")
	})

	t.Run("agents md includes classification", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewRooCodeAdapter().Translate(m, profile)
		agents := sectionByPath(sections, "AGENTS.md")
		require.NotNil(t, agents)
		assert.Contains(t, agents.Content(), "core")
	})

	t.Run("agents md includes after close protocol", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		adapters := []domain.ToolAdapter{
			domain.NewCursorAdapter(),
			domain.NewRooCodeAdapter(),
			domain.NewOpenCodeAdapter(),
		}
		for _, adapter := range adapters {
			sections := adapter.Translate(m, profile)
			agents := sectionByPath(sections, "AGENTS.md")
			require.NotNil(t, agents)
			assert.Contains(t, agents.Content(), "After-Close Protocol")
		}
	})

	t.Run("claude md and memory md protocol consistent", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewClaudeCodeAdapter().Translate(m, profile)
		claudeMD := sections[0].Content()
		memoryMD := sections[1].Content()
		for _, keyword := range []string{"bd-ripple", "review_needed", "Follow-up", "bd ready"} {
			assert.Contains(t, claudeMD, keyword, "CLAUDE.md missing %s", keyword)
			assert.Contains(t, memoryMD, keyword, "MEMORY.md missing %s", keyword)
		}
	})
}

// ---------------------------------------------------------------------------
// Profile integration tests
// ---------------------------------------------------------------------------

func TestAdapterProfileIntegration(t *testing.T) {
	t.Parallel()

	t.Run("python profile includes quality gates", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewClaudeCodeAdapter().Translate(m, vo.PythonUvProfile{})
		assert.Contains(t, sections[0].Content(), "Quality Gates")
		assert.Contains(t, sections[0].Content(), "uv run ruff")
	})

	t.Run("generic profile omits quality gates", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewClaudeCodeAdapter().Translate(m, vo.GenericProfile{})
		assert.NotContains(t, sections[0].Content(), "Quality Gates")
		assert.NotContains(t, sections[0].Content(), "uv run")
	})

	t.Run("cursor uses profile file glob", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewCursorAdapter().Translate(m, vo.PythonUvProfile{})
		mdc := sectionByPath(sections, ".cursor/rules/project-conventions.mdc")
		require.NotNil(t, mdc)
		assert.Contains(t, mdc.Content(), "globs: **/*.py")
	})

	t.Run("cursor generic profile uses star glob", func(t *testing.T) {
		t.Parallel()
		m := makeModel(t)
		sections := domain.NewCursorAdapter().Translate(m, vo.GenericProfile{})
		mdc := sectionByPath(sections, ".cursor/rules/project-conventions.mdc")
		require.NotNil(t, mdc)
		assert.Contains(t, mdc.Content(), "globs: *")
		assert.NotContains(t, mdc.Content(), "globs: **/*.py")
	})
}

// ---------------------------------------------------------------------------
// MEMORY.md tests
// ---------------------------------------------------------------------------

func TestMemoryMd(t *testing.T) {
	t.Parallel()

	getMemoryContent := func(t *testing.T, m *ddd.DomainModel, profile vo.StackProfile) string {
		t.Helper()
		sections := domain.NewClaudeCodeAdapter().Translate(m, profile)
		memory := sectionByPath(sections, ".claude/memory/MEMORY.md")
		require.NotNil(t, memory)
		return memory.Content()
	}

	t.Run("has after close protocol", func(t *testing.T) {
		t.Parallel()
		content := getMemoryContent(t, makeModel(t), vo.PythonUvProfile{})
		assert.Contains(t, content, "After-Close Protocol")
		assert.Contains(t, content, "bin/bd-ripple")
	})

	t.Run("has grooming checklist", func(t *testing.T) {
		t.Parallel()
		content := getMemoryContent(t, makeModel(t), vo.PythonUvProfile{})
		assert.Contains(t, content, "Grooming Checklist")
	})

	t.Run("has beads workflow", func(t *testing.T) {
		t.Parallel()
		content := getMemoryContent(t, makeModel(t), vo.PythonUvProfile{})
		assert.Contains(t, content, "bd ready")
		assert.Contains(t, content, "bd show")
		assert.Contains(t, content, "bd close")
	})

	t.Run("has bounded contexts", func(t *testing.T) {
		t.Parallel()
		content := getMemoryContent(t, makeModel(t), vo.PythonUvProfile{})
		assert.Contains(t, content, "Bounded Contexts")
		assert.Contains(t, content, "Orders")
	})

	t.Run("has ubiquitous language", func(t *testing.T) {
		t.Parallel()
		content := getMemoryContent(t, makeModel(t), vo.PythonUvProfile{})
		assert.Contains(t, content, "Ubiquitous Language")
	})

	t.Run("generic profile omits quality gates", func(t *testing.T) {
		t.Parallel()
		content := getMemoryContent(t, makeModel(t), vo.GenericProfile{})
		assert.NotContains(t, content, "Quality Gates")
		assert.NotContains(t, content, "uv run")
		assert.NotContains(t, content, "pytest")
	})

	t.Run("python profile includes quality gates", func(t *testing.T) {
		t.Parallel()
		content := getMemoryContent(t, makeModel(t), vo.PythonUvProfile{})
		assert.Contains(t, content, "Quality Gates")
		assert.Contains(t, content, "uv run ruff")
		assert.Contains(t, content, "uv run mypy")
		assert.Contains(t, content, "uv run pytest")
	})

	t.Run("under 200 lines", func(t *testing.T) {
		t.Parallel()
		m := makeMultiContextModel(t)
		for _, profile := range []vo.StackProfile{vo.PythonUvProfile{}, vo.GenericProfile{}} {
			content := getMemoryContent(t, m, profile)
			lineCount := len(strings.Split(content, "\n"))
			assert.Less(t, lineCount, 200, "MEMORY.md has %d lines, must be under 200", lineCount)
		}
	})

	t.Run("multi context model includes all contexts", func(t *testing.T) {
		t.Parallel()
		m := makeMultiContextModel(t)
		content := getMemoryContent(t, m, vo.PythonUvProfile{})
		assert.Contains(t, content, "Orders")
		assert.Contains(t, content, "Notifications")
	})

	t.Run("grooming has template compliance", func(t *testing.T) {
		t.Parallel()
		content := getMemoryContent(t, makeModel(t), vo.PythonUvProfile{})
		lower := strings.ToLower(content)
		assert.Contains(t, lower, "template compliance")
	})

	t.Run("grooming has prd traceability", func(t *testing.T) {
		t.Parallel()
		content := getMemoryContent(t, makeModel(t), vo.PythonUvProfile{})
		assert.True(t, strings.Contains(content, "PRD traceability") || strings.Contains(content, "prd-traceability"))
	})
}

// ---------------------------------------------------------------------------
// Roo/OpenCode generic profile tests
// ---------------------------------------------------------------------------

func TestRooCodeGenericProfile(t *testing.T) {
	t.Parallel()
	m := makeModel(t)
	sections := domain.NewRooCodeAdapter().Translate(m, vo.GenericProfile{})
	rules := sectionByPath(sections, ".roo/rules/project-conventions.md")
	require.NotNil(t, rules)
	assert.NotContains(t, rules.Content(), "Quality Gates")
}

func TestOpenCodeGenericProfile(t *testing.T) {
	t.Parallel()
	m := makeModel(t)
	sections := domain.NewOpenCodeAdapter().Translate(m, vo.GenericProfile{})
	rules := sectionByPath(sections, ".opencode/rules/project-conventions.md")
	require.NotNil(t, rules)
	assert.NotContains(t, rules.Content(), "Quality Gates")
}

func TestCursorGenericProfile(t *testing.T) {
	t.Parallel()
	m := makeModel(t)
	sections := domain.NewCursorAdapter().Translate(m, vo.GenericProfile{})

	mdc := sectionByPath(sections, ".cursor/rules/project-conventions.mdc")
	require.NotNil(t, mdc)
	assert.NotContains(t, mdc.Content(), "Quality Gates")

	agents := sectionByPath(sections, "AGENTS.md")
	require.NotNil(t, agents)
	assert.NotContains(t, agents.Content(), "Quality Gates")
}

func TestPythonProfileOutputFormat(t *testing.T) {
	t.Parallel()
	m := makeModel(t)
	sections := domain.NewClaudeCodeAdapter().Translate(m, vo.PythonUvProfile{})
	content := sections[0].Content()
	assert.Contains(t, content, "```bash")
	assert.Contains(t, content, "uv run ruff check .")
	assert.Contains(t, content, "uv run mypy .")
	assert.Contains(t, content, "uv run pytest")
}

func TestAllAdaptersAcceptProfile(t *testing.T) {
	t.Parallel()
	m := makeModel(t)
	profile := vo.PythonUvProfile{}
	adapters := []struct {
		name    string
		adapter domain.ToolAdapter
	}{
		{"ClaudeCode", domain.NewClaudeCodeAdapter()},
		{"Cursor", domain.NewCursorAdapter()},
		{"RooCode", domain.NewRooCodeAdapter()},
		{"OpenCode", domain.NewOpenCodeAdapter()},
	}
	for _, tt := range adapters {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sections := tt.adapter.Translate(m, profile)
			assert.GreaterOrEqual(t, len(sections), 1)
		})
	}
}
