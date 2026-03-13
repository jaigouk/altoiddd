package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// -- ClassificationResult value object tests --

func TestClassificationResult_Creation(t *testing.T) {
	t.Parallel()
	result := NewClassificationResult(vo.SubdomainCore, "Complex pricing rules")
	assert.Equal(t, vo.SubdomainCore, result.Classification())
	assert.Equal(t, "Complex pricing rules", result.Rationale())
}

func TestClassificationResult_Equality(t *testing.T) {
	t.Parallel()
	r1 := NewClassificationResult(vo.SubdomainCore, "Reason")
	r2 := NewClassificationResult(vo.SubdomainCore, "Reason")
	assert.True(t, r1.Equal(r2))
}

func TestClassificationResult_Inequality(t *testing.T) {
	t.Parallel()
	r1 := NewClassificationResult(vo.SubdomainCore, "Reason")
	r2 := NewClassificationResult(vo.SubdomainSupporting, "Reason")
	assert.False(t, r1.Equal(r2))
}

// -- ClassificationDecisionTree tests (Khononov decision tree) --

func TestClassificationDecisionTree_BuyYes_ReturnsGeneric(t *testing.T) {
	t.Parallel()
	tree := NewClassificationDecisionTree()
	result := tree.Classify(true, false, false)
	assert.Equal(t, vo.SubdomainGeneric, result.Classification())
	assert.Contains(t, result.Rationale(), "off-the-shelf")
}

func TestClassificationDecisionTree_BuyNo_ComplexitySimple_ReturnsSupporting(t *testing.T) {
	t.Parallel()
	tree := NewClassificationDecisionTree()
	// buyYes=false, complexRules=false (simple CRUD), competitorThreat=false
	result := tree.Classify(false, false, false)
	assert.Equal(t, vo.SubdomainSupporting, result.Classification())
	assert.Contains(t, result.Rationale(), "Necessary")
}

func TestClassificationDecisionTree_BuyNo_ComplexRules_CompetitorYes_ReturnsCore(t *testing.T) {
	t.Parallel()
	tree := NewClassificationDecisionTree()
	// buyYes=false, complexRules=true, competitorThreat=true
	result := tree.Classify(false, true, true)
	assert.Equal(t, vo.SubdomainCore, result.Classification())
	assert.Contains(t, result.Rationale(), "competitive")
}

func TestClassificationDecisionTree_BuyNo_ComplexRules_CompetitorNo_ReturnsSupporting(t *testing.T) {
	t.Parallel()
	tree := NewClassificationDecisionTree()
	// buyYes=false, complexRules=true, competitorThreat=false
	result := tree.Classify(false, true, false)
	assert.Equal(t, vo.SubdomainSupporting, result.Classification())
	assert.Contains(t, result.Rationale(), "complex but not differentiating")
}

// -- DiscoverySession classification command tests --

func TestDiscoverySession_ClassifyBoundedContext_StoresClassification(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	answerAllQuestions(session)
	require.NoError(t, session.Complete())

	result := NewClassificationResult(vo.SubdomainCore, "Competitive advantage")
	require.NoError(t, session.ClassifyBoundedContext("Orders", result))

	classifications := session.ContextClassifications()
	assert.Contains(t, classifications, "Orders")
	assert.Equal(t, vo.SubdomainCore, classifications["Orders"].Classification())
}

func TestDiscoverySession_ClassifyBoundedContext_BeforeCompleteReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	result := NewClassificationResult(vo.SubdomainCore, "Reason")
	err := session.ClassifyBoundedContext("Orders", result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot classify")
}

func TestDiscoverySession_ClassifyBoundedContext_EmptyNameReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	answerAllQuestions(session)
	require.NoError(t, session.Complete())

	result := NewClassificationResult(vo.SubdomainCore, "Reason")
	err := session.ClassifyBoundedContext("", result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context name")
}

func TestDiscoverySession_ClassifyBoundedContext_DuplicateReturnsError(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	answerAllQuestions(session)
	require.NoError(t, session.Complete())

	result := NewClassificationResult(vo.SubdomainCore, "Reason")
	require.NoError(t, session.ClassifyBoundedContext("Orders", result))
	err := session.ClassifyBoundedContext("Orders", result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already classified")
}

func TestDiscoverySession_ContextClassifications_ReturnsDefensiveCopy(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	answerAllQuestions(session)
	require.NoError(t, session.Complete())

	result := NewClassificationResult(vo.SubdomainCore, "Reason")
	require.NoError(t, session.ClassifyBoundedContext("Orders", result))

	copy1 := session.ContextClassifications()
	copy1["Orders"] = NewClassificationResult(vo.SubdomainGeneric, "Modified")
	copy2 := session.ContextClassifications()
	assert.Equal(t, vo.SubdomainCore, copy2["Orders"].Classification())
}

func TestDiscoverySession_SnapshotRoundTrip_PreservesClassifications(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	answerAllQuestions(session)
	require.NoError(t, session.Complete())

	result := NewClassificationResult(vo.SubdomainCore, "Competitive advantage")
	require.NoError(t, session.ClassifyBoundedContext("Orders", result))

	snapshot := session.ToSnapshot()
	restored, err := FromSnapshot(snapshot)
	require.NoError(t, err)

	classifications := restored.ContextClassifications()
	assert.Contains(t, classifications, "Orders")
	assert.Equal(t, vo.SubdomainCore, classifications["Orders"].Classification())
	assert.Equal(t, "Competitive advantage", classifications["Orders"].Rationale())
}

func TestDiscoverySession_ClassifyBoundedContext_EmitsBoundedContextClassified(t *testing.T) {
	t.Parallel()
	session := sessionWithPersona("1")
	answerAllQuestions(session)
	require.NoError(t, session.Complete())

	result := NewClassificationResult(vo.SubdomainCore, "Competitive advantage")
	require.NoError(t, session.ClassifyBoundedContext("Orders", result))

	events := session.ClassificationEvents()
	require.Len(t, events, 1)
	assert.Equal(t, session.SessionID(), events[0].SessionID())
	assert.Equal(t, "Orders", events[0].ContextName())
	assert.Equal(t, vo.SubdomainCore, events[0].Classification())
	assert.Equal(t, "Competitive advantage", events[0].Rationale())
}
