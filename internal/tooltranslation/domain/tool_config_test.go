package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
	"github.com/alty-cli/alty/internal/tooltranslation/domain"
)

// ---------------------------------------------------------------------------
// SupportedTool enum
// ---------------------------------------------------------------------------

func TestSupportedTool(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "claude-code", string(domain.ToolClaudeCode))
	assert.Equal(t, "cursor", string(domain.ToolCursor))
	assert.Equal(t, "roo-code", string(domain.ToolRooCode))
	assert.Equal(t, "opencode", string(domain.ToolOpenCode))
	assert.Len(t, domain.AllSupportedTools(), 4)
}

// ---------------------------------------------------------------------------
// ConfigSection
// ---------------------------------------------------------------------------

func TestConfigSection(t *testing.T) {
	t.Parallel()

	t.Run("stores fields", func(t *testing.T) {
		t.Parallel()
		s := domain.NewConfigSection("AGENTS.md", "content here", "agents file")
		assert.Equal(t, "AGENTS.md", s.FilePath())
		assert.Equal(t, "content here", s.Content())
		assert.Equal(t, "agents file", s.SectionName())
	})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeModelWithContexts(t *testing.T, contexts []struct {
	name           string
	classification vo.SubdomainClassification
},
) *ddd.DomainModel {
	t.Helper()
	m := ddd.NewDomainModel("test-toolconfig")

	var steps []string
	for _, ctx := range contexts {
		steps = append(steps, "User manages "+ctx.name)
	}
	err := m.AddDomainStory(vo.NewDomainStory("Test flow", []string{"User"}, "User starts", steps, nil))
	require.NoError(t, err)

	for _, ctx := range contexts {
		err = m.AddTerm(ctx.name, ctx.name+" domain", ctx.name, nil)
		require.NoError(t, err)
		err = m.AddBoundedContext(vo.NewDomainBoundedContext(ctx.name, "Manages "+ctx.name, nil, nil, ""))
		require.NoError(t, err)
		cl := ctx.classification
		err = m.ClassifySubdomain(ctx.name, cl, "")
		require.NoError(t, err)
	}
	for _, ctx := range contexts {
		if ctx.classification == vo.SubdomainCore {
			err = m.DesignAggregate(vo.NewAggregateDesign(
				ctx.name+"Root", ctx.name, ctx.name+"Root", nil,
				[]string{"must be valid"}, nil, nil,
			))
			require.NoError(t, err)
		}
	}
	return m
}

type fakeAdapter struct {
	sections []domain.ConfigSection
}

func (f *fakeAdapter) Translate(model *ddd.DomainModel, profile vo.StackProfile) []domain.ConfigSection {
	return f.sections
}

// ---------------------------------------------------------------------------
// ToolConfig creation
// ---------------------------------------------------------------------------

func TestToolConfigCreation(t *testing.T) {
	t.Parallel()

	t.Run("new config has empty sections", func(t *testing.T) {
		t.Parallel()
		config := domain.NewToolConfig(domain.ToolClaudeCode)
		assert.Empty(t, config.Sections())
		assert.Empty(t, config.Events())
	})

	t.Run("config has unique id", func(t *testing.T) {
		t.Parallel()
		a := domain.NewToolConfig(domain.ToolClaudeCode)
		b := domain.NewToolConfig(domain.ToolClaudeCode)
		assert.NotEqual(t, a.ConfigID(), b.ConfigID())
	})

	t.Run("config stores tool", func(t *testing.T) {
		t.Parallel()
		config := domain.NewToolConfig(domain.ToolCursor)
		assert.Equal(t, domain.ToolCursor, config.Tool())
	})
}

// ---------------------------------------------------------------------------
// build_sections
// ---------------------------------------------------------------------------

func TestBuildSections(t *testing.T) {
	t.Parallel()

	t.Run("builds sections from adapter", func(t *testing.T) {
		t.Parallel()
		m := makeModelWithContexts(t, []struct {
			name           string
			classification vo.SubdomainClassification
		}{{"Orders", vo.SubdomainCore}})
		config := domain.NewToolConfig(domain.ToolClaudeCode)
		adapter := &fakeAdapter{sections: []domain.ConfigSection{
			domain.NewConfigSection(".claude/CLAUDE.md", "# Test", "Test section"),
		}}
		err := config.BuildSections(m, adapter, vo.PythonUvProfile{})
		require.NoError(t, err)
		assert.Len(t, config.Sections(), 1)
		assert.Equal(t, ".claude/CLAUDE.md", config.Sections()[0].FilePath())
	})

	t.Run("build clears previous sections", func(t *testing.T) {
		t.Parallel()
		m := makeModelWithContexts(t, []struct {
			name           string
			classification vo.SubdomainClassification
		}{{"Orders", vo.SubdomainCore}})
		config := domain.NewToolConfig(domain.ToolClaudeCode)
		adapter1 := &fakeAdapter{sections: []domain.ConfigSection{
			domain.NewConfigSection(".claude/CLAUDE.md", "# Test", "Test section"),
		}}
		err := config.BuildSections(m, adapter1, vo.PythonUvProfile{})
		require.NoError(t, err)
		assert.Len(t, config.Sections(), 1)

		adapter2 := &fakeAdapter{sections: []domain.ConfigSection{
			domain.NewConfigSection("a.md", "a", "A"),
			domain.NewConfigSection("b.md", "b", "B"),
		}}
		err = config.BuildSections(m, adapter2, vo.PythonUvProfile{})
		require.NoError(t, err)
		assert.Len(t, config.Sections(), 2)
	})

	t.Run("cannot build after approve", func(t *testing.T) {
		t.Parallel()
		m := makeModelWithContexts(t, []struct {
			name           string
			classification vo.SubdomainClassification
		}{{"Orders", vo.SubdomainCore}})
		config := domain.NewToolConfig(domain.ToolClaudeCode)
		adapter := &fakeAdapter{sections: []domain.ConfigSection{
			domain.NewConfigSection(".claude/CLAUDE.md", "# Test", "Test"),
		}}
		_ = config.BuildSections(m, adapter, vo.PythonUvProfile{})
		_ = config.Approve()

		err := config.BuildSections(m, adapter, vo.PythonUvProfile{})
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})
}

// ---------------------------------------------------------------------------
// preview
// ---------------------------------------------------------------------------

func TestPreview(t *testing.T) {
	t.Parallel()

	t.Run("preview returns string", func(t *testing.T) {
		t.Parallel()
		m := makeModelWithContexts(t, []struct {
			name           string
			classification vo.SubdomainClassification
		}{{"Orders", vo.SubdomainCore}})
		config := domain.NewToolConfig(domain.ToolClaudeCode)
		adapter := &fakeAdapter{sections: []domain.ConfigSection{
			domain.NewConfigSection(".claude/CLAUDE.md", "# Test", "Test"),
		}}
		_ = config.BuildSections(m, adapter, vo.PythonUvProfile{})
		preview, err := config.Preview()
		require.NoError(t, err)
		assert.Contains(t, preview, "claude-code")
	})

	t.Run("preview without sections raises", func(t *testing.T) {
		t.Parallel()
		config := domain.NewToolConfig(domain.ToolClaudeCode)
		_, err := config.Preview()
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})
}

// ---------------------------------------------------------------------------
// approve
// ---------------------------------------------------------------------------

func TestApprove(t *testing.T) {
	t.Parallel()

	t.Run("approve emits event", func(t *testing.T) {
		t.Parallel()
		m := makeModelWithContexts(t, []struct {
			name           string
			classification vo.SubdomainClassification
		}{{"Orders", vo.SubdomainCore}})
		config := domain.NewToolConfig(domain.ToolClaudeCode)
		adapter := &fakeAdapter{sections: []domain.ConfigSection{
			domain.NewConfigSection(".claude/CLAUDE.md", "# Test", "Test"),
		}}
		_ = config.BuildSections(m, adapter, vo.PythonUvProfile{})
		err := config.Approve()
		require.NoError(t, err)
		assert.Len(t, config.Events(), 1)
	})

	t.Run("event contains tool name", func(t *testing.T) {
		t.Parallel()
		m := makeModelWithContexts(t, []struct {
			name           string
			classification vo.SubdomainClassification
		}{{"Orders", vo.SubdomainCore}})
		config := domain.NewToolConfig(domain.ToolCursor)
		adapter := &fakeAdapter{sections: []domain.ConfigSection{
			domain.NewConfigSection("AGENTS.md", "# Test", "Test"),
		}}
		_ = config.BuildSections(m, adapter, vo.PythonUvProfile{})
		_ = config.Approve()
		assert.Equal(t, []string{"cursor"}, config.Events()[0].ToolNames())
	})

	t.Run("event contains output paths", func(t *testing.T) {
		t.Parallel()
		m := makeModelWithContexts(t, []struct {
			name           string
			classification vo.SubdomainClassification
		}{{"Orders", vo.SubdomainCore}})
		config := domain.NewToolConfig(domain.ToolClaudeCode)
		adapter := &fakeAdapter{sections: []domain.ConfigSection{
			domain.NewConfigSection(".claude/CLAUDE.md", "# Test", "Test"),
		}}
		_ = config.BuildSections(m, adapter, vo.PythonUvProfile{})
		_ = config.Approve()
		assert.Contains(t, config.Events()[0].OutputPaths(), ".claude/CLAUDE.md")
	})

	t.Run("cannot approve without sections", func(t *testing.T) {
		t.Parallel()
		config := domain.NewToolConfig(domain.ToolClaudeCode)
		err := config.Approve()
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})

	t.Run("cannot approve twice", func(t *testing.T) {
		t.Parallel()
		m := makeModelWithContexts(t, []struct {
			name           string
			classification vo.SubdomainClassification
		}{{"Orders", vo.SubdomainCore}})
		config := domain.NewToolConfig(domain.ToolClaudeCode)
		adapter := &fakeAdapter{sections: []domain.ConfigSection{
			domain.NewConfigSection(".claude/CLAUDE.md", "# Test", "Test"),
		}}
		_ = config.BuildSections(m, adapter, vo.PythonUvProfile{})
		_ = config.Approve()
		err := config.Approve()
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})
}

// ---------------------------------------------------------------------------
// Defensive copies
// ---------------------------------------------------------------------------

func TestDefensiveCopies(t *testing.T) {
	t.Parallel()

	t.Run("sections returns copy", func(t *testing.T) {
		t.Parallel()
		config := domain.NewToolConfig(domain.ToolClaudeCode)
		s := config.Sections()
		assert.Empty(t, s)
	})

	t.Run("events returns copy", func(t *testing.T) {
		t.Parallel()
		config := domain.NewToolConfig(domain.ToolClaudeCode)
		e := config.Events()
		assert.Empty(t, e)
	})
}
