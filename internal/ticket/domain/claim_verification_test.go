package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/ticket/domain"
)

// ---------------------------------------------------------------------------
// VerifiableClaim value object tests
// ---------------------------------------------------------------------------

func TestNewVerifiableClaim_ValidatesTicketID(t *testing.T) {
	t.Parallel()

	_, err := domain.NewVerifiableClaim("", "14 findings", "deadcode ./...", "14", "line 47")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ticket ID")
}

func TestNewVerifiableClaim_ValidatesClaimText(t *testing.T) {
	t.Parallel()

	_, err := domain.NewVerifiableClaim("t-123", "", "deadcode ./...", "14", "line 47")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "claim text")
}

func TestNewVerifiableClaim_AllowsEmptyCommand(t *testing.T) {
	t.Parallel()

	claim, err := domain.NewVerifiableClaim("t-123", "14 findings", "", "14", "line 47")
	require.NoError(t, err)
	assert.Empty(t, claim.Command())
}

func TestVerifiableClaim_Accessors(t *testing.T) {
	t.Parallel()

	claim, err := domain.NewVerifiableClaim("t-123", "14 production findings", "deadcode ./cmd/...", "14", "line 47")
	require.NoError(t, err)

	assert.Equal(t, "t-123", claim.TicketID())
	assert.Equal(t, "14 production findings", claim.ClaimText())
	assert.Equal(t, "deadcode ./cmd/...", claim.Command())
	assert.Equal(t, "14", claim.ClaimedValue())
	assert.Equal(t, "line 47", claim.Location())
}

func TestVerifiableClaim_IsVerifiable(t *testing.T) {
	t.Parallel()

	withCmd, _ := domain.NewVerifiableClaim("t-123", "14 findings", "deadcode ./...", "14", "line 47")
	assert.True(t, withCmd.IsVerifiable())

	withoutCmd, _ := domain.NewVerifiableClaim("t-123", "14 findings", "", "14", "line 47")
	assert.False(t, withoutCmd.IsVerifiable())
}

// ---------------------------------------------------------------------------
// VerificationResult value object tests
// ---------------------------------------------------------------------------

func TestNewVerificationResult_Match(t *testing.T) {
	t.Parallel()

	claim, _ := domain.NewVerifiableClaim("t-123", "14 findings", "deadcode ./...", "14", "line 47")
	result := domain.NewVerificationResult(claim, "14", nil)

	assert.True(t, result.Match())
	assert.Empty(t, result.Discrepancy())
	assert.Equal(t, "14", result.ActualValue())
}

func TestNewVerificationResult_Mismatch(t *testing.T) {
	t.Parallel()

	claim, _ := domain.NewVerifiableClaim("t-123", "14 findings", "deadcode ./...", "14", "line 47")
	result := domain.NewVerificationResult(claim, "288", nil)

	assert.False(t, result.Match())
	assert.Contains(t, result.Discrepancy(), "claimed 14")
	assert.Contains(t, result.Discrepancy(), "actual 288")
	assert.Contains(t, result.Discrepancy(), "x difference")
}

func TestNewVerificationResult_CommandError(t *testing.T) {
	t.Parallel()

	claim, _ := domain.NewVerifiableClaim("t-123", "14 findings", "deadcode ./...", "14", "line 47")
	result := domain.NewVerificationResult(claim, "", assert.AnError)

	assert.False(t, result.Match())
	assert.Contains(t, result.Discrepancy(), "error")
}

func TestNewVerificationResult_Unverifiable(t *testing.T) {
	t.Parallel()

	claim, _ := domain.NewVerifiableClaim("t-123", "14 findings", "", "14", "line 47")
	result := domain.NewVerificationResult(claim, "", nil)

	assert.False(t, result.Match())
	assert.Contains(t, result.Discrepancy(), "UNVERIFIED")
}

// ---------------------------------------------------------------------------
// ClaimVerifier domain service tests
// ---------------------------------------------------------------------------

func TestClaimVerifier_ParsesQuantitativeClaims(t *testing.T) {
	t.Parallel()

	markdown := `## Quality Gates
The analysis revealed **14 production findings** that need attention.
`
	verifier := domain.NewClaimVerifier()
	claims := verifier.ParseClaims("t-123", markdown)

	require.Len(t, claims, 1)
	assert.Equal(t, "14", claims[0].ClaimedValue())
	assert.Contains(t, claims[0].ClaimText(), "14")
}

func TestClaimVerifier_ExtractsCommandFromCodeBlock(t *testing.T) {
	t.Parallel()

	markdown := "## Analysis\n\n```bash\ndeadcode ./cmd/... | wc -l\n```\n\nThis revealed **288 functions** to review.\n"

	verifier := domain.NewClaimVerifier()
	claims := verifier.ParseClaims("t-123", markdown)

	require.Len(t, claims, 1)
	assert.Equal(t, "288", claims[0].ClaimedValue())
	assert.Contains(t, claims[0].Command(), "deadcode")
}

func TestClaimVerifier_MultipleClaimsInTicket(t *testing.T) {
	t.Parallel()

	markdown := `Found **14 critical issues** and **5 warnings**.`

	verifier := domain.NewClaimVerifier()
	claims := verifier.ParseClaims("t-123", markdown)

	assert.Len(t, claims, 2)
}

func TestClaimVerifier_NoClaimsFound(t *testing.T) {
	t.Parallel()

	markdown := `## Design
This ticket implements a new feature without quantitative claims.
`
	verifier := domain.NewClaimVerifier()
	claims := verifier.ParseClaims("t-123", markdown)

	assert.Empty(t, claims)
}

func TestClaimVerifier_IgnoresNonQuantitativeBold(t *testing.T) {
	t.Parallel()

	markdown := `The **important** thing is to **focus** on quality.`

	verifier := domain.NewClaimVerifier()
	claims := verifier.ParseClaims("t-123", markdown)

	assert.Empty(t, claims)
}
