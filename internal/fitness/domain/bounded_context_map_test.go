package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/fitness/domain"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// RelationshipDirection
// ---------------------------------------------------------------------------

func TestRelationshipDirectionConstants(t *testing.T) {
	t.Parallel()
	assert.Equal(t, domain.RelationshipUpstream, domain.RelationshipDirection("upstream"))
	assert.Equal(t, domain.RelationshipDownstream, domain.RelationshipDirection("downstream"))
}

func TestAllRelationshipDirections(t *testing.T) {
	t.Parallel()
	all := domain.AllRelationshipDirections()
	assert.Len(t, all, 2)
	assert.Contains(t, all, domain.RelationshipUpstream)
	assert.Contains(t, all, domain.RelationshipDownstream)
}

// ---------------------------------------------------------------------------
// RelationshipPattern
// ---------------------------------------------------------------------------

func TestRelationshipPatternConstants(t *testing.T) {
	t.Parallel()
	assert.Equal(t, domain.PatternDomainEvent, domain.RelationshipPattern("domain_event"))
	assert.Equal(t, domain.PatternSharedKernel, domain.RelationshipPattern("shared_kernel"))
	assert.Equal(t, domain.PatternACL, domain.RelationshipPattern("acl"))
	assert.Equal(t, domain.PatternOpenHost, domain.RelationshipPattern("open_host"))
}

func TestAllRelationshipPatterns(t *testing.T) {
	t.Parallel()
	all := domain.AllRelationshipPatterns()
	assert.Len(t, all, 4)
	assert.Contains(t, all, domain.PatternDomainEvent)
	assert.Contains(t, all, domain.PatternSharedKernel)
	assert.Contains(t, all, domain.PatternACL)
	assert.Contains(t, all, domain.PatternOpenHost)
}

// ---------------------------------------------------------------------------
// ContextRelationship (Value Object)
// ---------------------------------------------------------------------------

func TestNewContextRelationship(t *testing.T) {
	t.Parallel()

	rel := domain.NewContextRelationship("Discovery", domain.RelationshipDownstream, domain.PatternDomainEvent)

	assert.Equal(t, "Discovery", rel.Target())
	assert.Equal(t, domain.RelationshipDownstream, rel.Direction())
	assert.Equal(t, domain.PatternDomainEvent, rel.Pattern())
}

// ---------------------------------------------------------------------------
// BoundedContextEntry (Value Object)
// ---------------------------------------------------------------------------

func TestNewBoundedContextEntry(t *testing.T) {
	t.Parallel()

	classification := vo.SubdomainCore
	layers := []string{"domain", "application", "infrastructure"}
	relationships := []domain.ContextRelationship{
		domain.NewContextRelationship("Discovery", domain.RelationshipDownstream, domain.PatternDomainEvent),
	}

	entry := domain.NewBoundedContextEntry(
		"Bootstrap",
		"bootstrap",
		classification,
		layers,
		relationships,
	)

	assert.Equal(t, "Bootstrap", entry.Name())
	assert.Equal(t, "bootstrap", entry.ModulePath())
	assert.Equal(t, vo.SubdomainCore, entry.Classification())
	assert.Equal(t, layers, entry.Layers())
	assert.Len(t, entry.Relationships(), 1)
}

func TestBoundedContextEntryDefensiveCopy(t *testing.T) {
	t.Parallel()

	layers := []string{"domain", "application"}
	relationships := []domain.ContextRelationship{
		domain.NewContextRelationship("Other", domain.RelationshipUpstream, domain.PatternACL),
	}

	entry := domain.NewBoundedContextEntry("Test", "test", vo.SubdomainSupporting, layers, relationships)

	// Mutate original slices
	layers[0] = "mutated"
	relationships[0] = domain.NewContextRelationship("Mutated", domain.RelationshipDownstream, domain.PatternSharedKernel)

	// Entry should be unaffected
	assert.Equal(t, "domain", entry.Layers()[0])
	assert.Equal(t, "Other", entry.Relationships()[0].Target())
}

// ---------------------------------------------------------------------------
// BoundedContextMap (Value Object)
// ---------------------------------------------------------------------------

func TestNewBoundedContextMap(t *testing.T) {
	t.Parallel()

	contexts := []domain.BoundedContextEntry{
		domain.NewBoundedContextEntry("Bootstrap", "bootstrap", vo.SubdomainSupporting, []string{"domain", "application", "infrastructure"}, nil),
		domain.NewBoundedContextEntry("Discovery", "discovery", vo.SubdomainCore, []string{"domain", "application", "infrastructure"}, nil),
	}

	bcMap := domain.NewBoundedContextMap("alty", "github.com/alty-cli/alty", contexts)

	assert.Equal(t, "alty", bcMap.ProjectName())
	assert.Equal(t, "github.com/alty-cli/alty", bcMap.RootPackage())
	assert.Len(t, bcMap.Contexts(), 2)
}

func TestBoundedContextMapDefensiveCopy(t *testing.T) {
	t.Parallel()

	contexts := []domain.BoundedContextEntry{
		domain.NewBoundedContextEntry("Test", "test", vo.SubdomainGeneric, nil, nil),
	}

	bcMap := domain.NewBoundedContextMap("proj", "github.com/org/proj", contexts)

	// Mutate original slice
	contexts[0] = domain.NewBoundedContextEntry("Mutated", "mutated", vo.SubdomainCore, nil, nil)

	// Map should be unaffected
	assert.Equal(t, "Test", bcMap.Contexts()[0].Name())
}

func TestBoundedContextMapFindContext(t *testing.T) {
	t.Parallel()

	contexts := []domain.BoundedContextEntry{
		domain.NewBoundedContextEntry("Bootstrap", "bootstrap", vo.SubdomainSupporting, nil, nil),
		domain.NewBoundedContextEntry("Discovery", "discovery", vo.SubdomainCore, nil, nil),
	}

	bcMap := domain.NewBoundedContextMap("alty", "github.com/alty-cli/alty", contexts)

	// Found
	entry, found := bcMap.FindContext("Bootstrap")
	require.True(t, found)
	assert.Equal(t, "Bootstrap", entry.Name())

	// Not found
	_, found = bcMap.FindContext("NonExistent")
	assert.False(t, found)
}

func TestBoundedContextMapContextNames(t *testing.T) {
	t.Parallel()

	contexts := []domain.BoundedContextEntry{
		domain.NewBoundedContextEntry("Bootstrap", "bootstrap", vo.SubdomainSupporting, nil, nil),
		domain.NewBoundedContextEntry("Discovery", "discovery", vo.SubdomainCore, nil, nil),
	}

	bcMap := domain.NewBoundedContextMap("alty", "github.com/alty-cli/alty", contexts)

	names := bcMap.ContextNames()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "Bootstrap")
	assert.Contains(t, names, "Discovery")
}

func TestBoundedContextMapContextsWithClassification(t *testing.T) {
	t.Parallel()

	contexts := []domain.BoundedContextEntry{
		domain.NewBoundedContextEntry("Bootstrap", "bootstrap", vo.SubdomainSupporting, nil, nil),
		domain.NewBoundedContextEntry("Discovery", "discovery", vo.SubdomainCore, nil, nil),
		domain.NewBoundedContextEntry("Challenge", "challenge", vo.SubdomainCore, nil, nil),
		domain.NewBoundedContextEntry("FileGen", "filegen", vo.SubdomainGeneric, nil, nil),
	}

	bcMap := domain.NewBoundedContextMap("alty", "github.com/alty-cli/alty", contexts)

	coreContexts := bcMap.ContextsWithClassification(vo.SubdomainCore)
	assert.Len(t, coreContexts, 2)

	supportingContexts := bcMap.ContextsWithClassification(vo.SubdomainSupporting)
	assert.Len(t, supportingContexts, 1)

	genericContexts := bcMap.ContextsWithClassification(vo.SubdomainGeneric)
	assert.Len(t, genericContexts, 1)
}
