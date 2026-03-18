package ddd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// AddTerm
// ---------------------------------------------------------------------------

func TestAddSingleTerm(t *testing.T) {
	t.Parallel()
	ul := ddd.NewUbiquitousLanguage()
	require.NoError(t, ul.AddTerm("Order", "A customer purchase", "Sales", nil))
	assert.Len(t, ul.Terms(), 1)
	assert.Equal(t, "Order", ul.Terms()[0].Term())
}

func TestAddMultipleTerms(t *testing.T) {
	t.Parallel()
	ul := ddd.NewUbiquitousLanguage()
	require.NoError(t, ul.AddTerm("Order", "A purchase", "Sales", nil))
	require.NoError(t, ul.AddTerm("Product", "An item for sale", "Catalog", nil))
	assert.Len(t, ul.Terms(), 2)
}

func TestEmptyTermRaises(t *testing.T) {
	t.Parallel()
	ul := ddd.NewUbiquitousLanguage()
	err := ul.AddTerm("", "Definition", "Context", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "term cannot be empty")
}

func TestWhitespaceTermRaises(t *testing.T) {
	t.Parallel()
	ul := ddd.NewUbiquitousLanguage()
	err := ul.AddTerm("   ", "Definition", "Context", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "term cannot be empty")
}

func TestEmptyDefinitionRaises(t *testing.T) {
	t.Parallel()
	ul := ddd.NewUbiquitousLanguage()
	err := ul.AddTerm("Order", "", "Context", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "definition cannot be empty")
}

func TestTermStripped(t *testing.T) {
	t.Parallel()
	ul := ddd.NewUbiquitousLanguage()
	require.NoError(t, ul.AddTerm("  Order  ", "A purchase", "Sales", nil))
	assert.Equal(t, "Order", ul.Terms()[0].Term())
}

func TestSourceQuestionIDs(t *testing.T) {
	t.Parallel()
	ul := ddd.NewUbiquitousLanguage()
	require.NoError(t, ul.AddTerm("Order", "A purchase", "Sales", []string{"Q1", "Q2"}))
	assert.Equal(t, []string{"Q1", "Q2"}, ul.Terms()[0].SourceQuestionIDs())
}

// ---------------------------------------------------------------------------
// GetTermsForContext
// ---------------------------------------------------------------------------

func TestFilterByContext(t *testing.T) {
	t.Parallel()
	ul := ddd.NewUbiquitousLanguage()
	require.NoError(t, ul.AddTerm("Order", "A purchase", "Sales", nil))
	require.NoError(t, ul.AddTerm("Product", "An item", "Catalog", nil))
	require.NoError(t, ul.AddTerm("Invoice", "A bill", "Sales", nil))

	salesTerms := ul.GetTermsForContext("Sales")
	assert.Len(t, salesTerms, 2)
	names := map[string]struct{}{}
	for _, te := range salesTerms {
		names[te.Term()] = struct{}{}
	}
	assert.Equal(t, map[string]struct{}{"Order": {}, "Invoice": {}}, names)
}

func TestNoTermsInContext(t *testing.T) {
	t.Parallel()
	ul := ddd.NewUbiquitousLanguage()
	require.NoError(t, ul.AddTerm("Order", "A purchase", "Sales", nil))
	assert.Empty(t, ul.GetTermsForContext("Unknown"))
}

// ---------------------------------------------------------------------------
// FindAmbiguousTerms
// ---------------------------------------------------------------------------

func TestNoAmbiguity(t *testing.T) {
	t.Parallel()
	ul := ddd.NewUbiquitousLanguage()
	require.NoError(t, ul.AddTerm("Order", "A purchase", "Sales", nil))
	require.NoError(t, ul.AddTerm("Product", "An item", "Catalog", nil))
	assert.Empty(t, ul.FindAmbiguousTerms())
}

func TestDetectsAmbiguity(t *testing.T) {
	t.Parallel()
	ul := ddd.NewUbiquitousLanguage()
	require.NoError(t, ul.AddTerm("Config", "App settings", "Bootstrap", nil))
	require.NoError(t, ul.AddTerm("Config", "Tool configuration", "Tool Translation", nil))
	assert.Equal(t, []string{"config"}, ul.FindAmbiguousTerms())
}

func TestAmbiguityCaseInsensitive(t *testing.T) {
	t.Parallel()
	ul := ddd.NewUbiquitousLanguage()
	require.NoError(t, ul.AddTerm("Order", "Sales order", "Sales", nil))
	require.NoError(t, ul.AddTerm("order", "Work order", "Manufacturing", nil))
	assert.Equal(t, []string{"order"}, ul.FindAmbiguousTerms())
}

// ---------------------------------------------------------------------------
// HasPerContextDefinitions
// ---------------------------------------------------------------------------

func TestAmbiguousWithDefinitions(t *testing.T) {
	t.Parallel()
	ul := ddd.NewUbiquitousLanguage()
	require.NoError(t, ul.AddTerm("Config", "App settings", "Bootstrap", nil))
	require.NoError(t, ul.AddTerm("Config", "Tool config", "Tool Translation", nil))
	assert.True(t, ul.HasPerContextDefinitions("Config"))
}

func TestAmbiguousMissingDefinition(t *testing.T) {
	t.Parallel()
	ul := ddd.NewUbiquitousLanguage()
	require.NoError(t, ul.AddTerm("Config", "App settings", "Bootstrap", nil))
	// Bypass validation to simulate empty definition.
	ul.AddTermEntry(vo.NewTermEntry("Config", "", "Tool Translation", nil))
	assert.False(t, ul.HasPerContextDefinitions("Config"))
}

// ---------------------------------------------------------------------------
// AllTermNames
// ---------------------------------------------------------------------------

func TestAllTermNames(t *testing.T) {
	t.Parallel()
	ul := ddd.NewUbiquitousLanguage()
	require.NoError(t, ul.AddTerm("Order", "A purchase", "Sales", nil))
	require.NoError(t, ul.AddTerm("Product", "An item", "Catalog", nil))
	expected := map[string]struct{}{"order": {}, "product": {}}
	assert.Equal(t, expected, ul.AllTermNames())
}

func TestAllTermNamesEmpty(t *testing.T) {
	t.Parallel()
	ul := ddd.NewUbiquitousLanguage()
	assert.Empty(t, ul.AllTermNames())
}

// ---------------------------------------------------------------------------
// Terms defensive copy
// ---------------------------------------------------------------------------

func TestTermsDefensiveCopy(t *testing.T) {
	t.Parallel()
	ul := ddd.NewUbiquitousLanguage()
	require.NoError(t, ul.AddTerm("Order", "A purchase", "Sales", nil))
	terms1 := ul.Terms()
	terms2 := ul.Terms()
	assert.Equal(t, terms1, terms2)
	// Verify they are different slices.
	assert.NotSame(t, &terms1[0], &terms2[0])
}
