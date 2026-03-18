package valueobjects_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// PersonaType enum
// ---------------------------------------------------------------------------

func TestPersonaTypeHasFiveValues(t *testing.T) {
	t.Parallel()
	assert.Len(t, vo.AllPersonaTypes(), 5)
}

func TestPersonaTypeValues(t *testing.T) {
	t.Parallel()
	expected := map[string]struct{}{
		"solo_developer":   {},
		"team_lead":        {},
		"ai_tool_switcher": {},
		"product_owner":    {},
		"domain_expert":    {},
	}
	actual := make(map[string]struct{})
	for _, pt := range vo.AllPersonaTypes() {
		actual[string(pt)] = struct{}{}
	}
	assert.Equal(t, expected, actual)
}

// ---------------------------------------------------------------------------
// Register enum
// ---------------------------------------------------------------------------

func TestRegisterValues(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "technical", string(vo.RegisterTechnical))
	assert.Equal(t, "non_technical", string(vo.RegisterNonTechnical))
}

// ---------------------------------------------------------------------------
// PersonaDefinition
// ---------------------------------------------------------------------------

func TestPersonaDefinitionFields(t *testing.T) {
	t.Parallel()
	defn, err := vo.NewPersonaDefinition(
		"Solo Developer",
		vo.PersonaSoloDeveloper,
		vo.RegisterTechnical,
		"Full-stack developer",
		"# Solo\n\nInstructions.",
	)
	require.NoError(t, err)
	assert.Equal(t, "Solo Developer", defn.Name())
	assert.Equal(t, vo.PersonaSoloDeveloper, defn.PersonaType())
	assert.Equal(t, vo.RegisterTechnical, defn.Register())
	assert.Equal(t, "Full-stack developer", defn.Description())
	assert.Equal(t, "# Solo\n\nInstructions.", defn.InstructionsTemplate())
}

func TestPersonaDefinitionRejectsEmptyName(t *testing.T) {
	t.Parallel()
	_, err := vo.NewPersonaDefinition(
		"",
		vo.PersonaSoloDeveloper,
		vo.RegisterTechnical,
		"A test persona",
		"# Test\n\nInstructions.",
	)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	assert.Contains(t, err.Error(), "name cannot be empty")
}

func TestPersonaDefinitionRejectsWhitespaceName(t *testing.T) {
	t.Parallel()
	_, err := vo.NewPersonaDefinition(
		"   ",
		vo.PersonaSoloDeveloper,
		vo.RegisterTechnical,
		"A test persona",
		"# Test\n\nInstructions.",
	)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	assert.Contains(t, err.Error(), "name cannot be empty")
}

func TestPersonaDefinitionRejectsEmptyTemplate(t *testing.T) {
	t.Parallel()
	_, err := vo.NewPersonaDefinition(
		"Test",
		vo.PersonaSoloDeveloper,
		vo.RegisterTechnical,
		"A test persona",
		"",
	)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	assert.Contains(t, err.Error(), "instructions template cannot be empty")
}

func TestPersonaDefinitionRejectsWhitespaceTemplate(t *testing.T) {
	t.Parallel()
	_, err := vo.NewPersonaDefinition(
		"Test",
		vo.PersonaSoloDeveloper,
		vo.RegisterTechnical,
		"A test persona",
		"   ",
	)
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	assert.Contains(t, err.Error(), "instructions template cannot be empty")
}

// ---------------------------------------------------------------------------
// Registry register assignments
// ---------------------------------------------------------------------------

func TestRegistryTechnicalRegisterForSoloDeveloper(t *testing.T) {
	t.Parallel()
	reg := vo.PersonaRegistry()
	assert.Equal(t, vo.RegisterTechnical, reg[vo.PersonaSoloDeveloper].Register())
}

func TestRegistryTechnicalRegisterForTeamLead(t *testing.T) {
	t.Parallel()
	reg := vo.PersonaRegistry()
	assert.Equal(t, vo.RegisterTechnical, reg[vo.PersonaTeamLead].Register())
}

func TestRegistryTechnicalRegisterForAIToolSwitcher(t *testing.T) {
	t.Parallel()
	reg := vo.PersonaRegistry()
	assert.Equal(t, vo.RegisterTechnical, reg[vo.PersonaAIToolSwitcher].Register())
}

func TestRegistryNonTechnicalRegisterForProductOwner(t *testing.T) {
	t.Parallel()
	reg := vo.PersonaRegistry()
	assert.Equal(t, vo.RegisterNonTechnical, reg[vo.PersonaProductOwner].Register())
}

func TestRegistryNonTechnicalRegisterForDomainExpert(t *testing.T) {
	t.Parallel()
	reg := vo.PersonaRegistry()
	assert.Equal(t, vo.RegisterNonTechnical, reg[vo.PersonaDomainExpert].Register())
}

// ---------------------------------------------------------------------------
// PERSONA_REGISTRY completeness
// ---------------------------------------------------------------------------

func TestPersonaRegistryHasFiveEntries(t *testing.T) {
	t.Parallel()
	assert.Len(t, vo.PersonaRegistry(), 5)
}

func TestPersonaRegistryKeysMatchPersonaTypes(t *testing.T) {
	t.Parallel()
	reg := vo.PersonaRegistry()
	allTypes := vo.AllPersonaTypes()
	for _, pt := range allTypes {
		_, ok := reg[pt]
		require.True(t, ok, "missing registry entry for %s", pt)
	}
}

func TestEachRegistryEntryHasMatchingPersonaType(t *testing.T) {
	t.Parallel()
	reg := vo.PersonaRegistry()
	for key, defn := range reg {
		assert.Equal(t, key, defn.PersonaType())
	}
}

// ---------------------------------------------------------------------------
// SUPPORTED_TOOLS and TOOL_TARGET_PATHS
// ---------------------------------------------------------------------------

func TestSupportedToolsList(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []string{"claude-code", "cursor", "roo-code", "opencode"}, vo.SupportedTools())
}

func TestToolTargetPathsContainTools(t *testing.T) {
	t.Parallel()
	paths := vo.ToolTargetPaths()
	for _, tool := range []string{"claude-code", "cursor", "roo-code", "opencode"} {
		_, ok := paths[tool]
		require.True(t, ok, "missing target path for %s", tool)
	}
}

func TestToolTargetPathsContainNamePlaceholder(t *testing.T) {
	t.Parallel()
	paths := vo.ToolTargetPaths()
	for tool, path := range paths {
		assert.Contains(t, path, "{name}",
			"Path for %s must contain {name} placeholder", tool)
	}
}
