package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
	"github.com/alty-cli/alty/internal/ticket/domain"
)

func TestFindingSeverityValues(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "critical", string(domain.FindingSeverityCritical))
	assert.Equal(t, "major", string(domain.FindingSeverityMajor))
	assert.Equal(t, "minor", string(domain.FindingSeverityMinor))
}

func TestImplementabilityFindingValid(t *testing.T) {
	t.Parallel()
	f, err := domain.NewImplementabilityFinding(domain.FindingSeverityCritical, "Design", "Missing port reference for web search")
	require.NoError(t, err)
	assert.Equal(t, domain.FindingSeverityCritical, f.Severity())
	assert.Equal(t, "Design", f.Location())
	assert.Contains(t, f.Description(), "Missing port")
}

func TestImplementabilityFindingRejectsEmpty(t *testing.T) {
	t.Parallel()
	_, err := domain.NewImplementabilityFinding(domain.FindingSeverityMinor, "AC", "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestImplementabilityFindingRejectsWhitespace(t *testing.T) {
	t.Parallel()
	_, err := domain.NewImplementabilityFinding(domain.FindingSeverityMinor, "AC", "   ")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestInterfaceMismatch(t *testing.T) {
	t.Parallel()
	m := domain.NewInterfaceMismatch("ISP", "Sequence Diagram", "research() signature differs")
	assert.Equal(t, "ISP", m.SectionA())
	assert.Equal(t, "Sequence Diagram", m.SectionB())
	assert.Contains(t, m.Description(), "signature")
}

func TestUnresolvedDependencyValid(t *testing.T) {
	t.Parallel()
	d, err := domain.NewUnresolvedDependency("WebSearchPort", "Design", "No such port exists")
	require.NoError(t, err)
	assert.Equal(t, "WebSearchPort", d.PortName())
	assert.Equal(t, "Design", d.Location())
}

func TestUnresolvedDependencyRejectsEmpty(t *testing.T) {
	t.Parallel()
	_, err := domain.NewUnresolvedDependency("", "Design", "Missing port")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestTicketSection(t *testing.T) {
	t.Parallel()
	s := domain.NewTicketSection("## Design", "Some design text")
	assert.Equal(t, "## Design", s.Heading())
	assert.Equal(t, "Some design text", s.Content())
}

func TestTicketStructureGetSectionFound(t *testing.T) {
	t.Parallel()
	sections := []domain.TicketSection{
		domain.NewTicketSection("## Goal", "goal text"),
		domain.NewTicketSection("## Design", "design text"),
	}
	structure := domain.NewTicketStructure(sections)
	result := structure.GetSection("## Design")
	require.NotNil(t, result)
	assert.Equal(t, "design text", result.Content())
}

func TestTicketStructureGetSectionMissing(t *testing.T) {
	t.Parallel()
	structure := domain.NewTicketStructure(nil)
	assert.Nil(t, structure.GetSection("## Missing"))
}

func TestDesignTraceResultIsValidEmpty(t *testing.T) {
	t.Parallel()
	result := domain.NewDesignTraceResult("t1", nil)
	assert.True(t, result.IsValid())
}

func TestDesignTraceResultIsInvalid(t *testing.T) {
	t.Parallel()
	f, _ := domain.NewImplementabilityFinding(domain.FindingSeverityMinor, "Goal", "Vague goal")
	result := domain.NewDesignTraceResult("t1", []domain.ImplementabilityFinding{f})
	assert.False(t, result.IsValid())
}

func TestDesignTraceResultCriticalCount(t *testing.T) {
	t.Parallel()
	f1, _ := domain.NewImplementabilityFinding(domain.FindingSeverityCritical, "Design", "Missing port")
	f2, _ := domain.NewImplementabilityFinding(domain.FindingSeverityMajor, "AC", "No checkboxes")
	f3, _ := domain.NewImplementabilityFinding(domain.FindingSeverityCritical, "SOLID", "Mismatch")
	result := domain.NewDesignTraceResult("t1", []domain.ImplementabilityFinding{f1, f2, f3})
	assert.Equal(t, 2, result.CriticalCount())
}

func TestDesignTraceResultTicketID(t *testing.T) {
	t.Parallel()
	result := domain.NewDesignTraceResult("t1", nil)
	assert.Equal(t, "t1", result.TicketID())
}

func TestDesignTraceResultFindings(t *testing.T) {
	t.Parallel()
	f1, _ := domain.NewImplementabilityFinding(domain.FindingSeverityCritical, "Design", "Missing port")
	f2, _ := domain.NewImplementabilityFinding(domain.FindingSeverityMajor, "AC", "No checkboxes")
	result := domain.NewDesignTraceResult("t1", []domain.ImplementabilityFinding{f1, f2})
	assert.Len(t, result.Findings(), 2)
}
