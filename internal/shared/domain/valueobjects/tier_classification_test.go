package valueobjects_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// classify_tier: Core override
// ---------------------------------------------------------------------------

func TestCoreOverride(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		depth int
	}{
		{"core depth 0", 0},
		{"core depth 5", 5},
		{"core depth 100", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := vo.ClassifyTier(tt.depth, vo.SubdomainCore)
			assert.Equal(t, vo.TierNearTerm, result.Tier())
		})
	}
}

func TestCoreReasonMentionsCore(t *testing.T) {
	t.Parallel()
	result := vo.ClassifyTier(5, vo.SubdomainCore)
	reason := strings.ToLower(result.Reason())
	assert.Contains(t, reason, "core")
}

// ---------------------------------------------------------------------------
// classify_tier: depth boundary
// ---------------------------------------------------------------------------

func TestDepthBoundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		classification vo.SubdomainClassification
		wantTier       vo.Tier
		depth          int
	}{
		{name: "supporting depth 0 near term", depth: 0, classification: vo.SubdomainSupporting, wantTier: vo.TierNearTerm},
		{name: "supporting depth 2 near term", depth: 2, classification: vo.SubdomainSupporting, wantTier: vo.TierNearTerm},
		{name: "supporting depth 3 far term", depth: 3, classification: vo.SubdomainSupporting, wantTier: vo.TierFarTerm},
		{name: "generic depth 1 near term", depth: 1, classification: vo.SubdomainGeneric, wantTier: vo.TierNearTerm},
		{name: "generic depth 3 far term", depth: 3, classification: vo.SubdomainGeneric, wantTier: vo.TierFarTerm},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := vo.ClassifyTier(tt.depth, tt.classification)
			assert.Equal(t, tt.wantTier, result.Tier())
		})
	}
}

// ---------------------------------------------------------------------------
// tier_to_detail_level
// ---------------------------------------------------------------------------

func TestTierToDetailLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		tier           vo.Tier
		classification vo.SubdomainClassification
		want           vo.TicketDetailLevel
	}{
		{"near term core is full", vo.TierNearTerm, vo.SubdomainCore, vo.TicketDetailFull},
		{"near term supporting is standard", vo.TierNearTerm, vo.SubdomainSupporting, vo.TicketDetailStandard},
		{"near term generic is stub", vo.TierNearTerm, vo.SubdomainGeneric, vo.TicketDetailStub},
		{"far term supporting is stub", vo.TierFarTerm, vo.SubdomainSupporting, vo.TicketDetailStub},
		{"far term core is full", vo.TierFarTerm, vo.SubdomainCore, vo.TicketDetailFull},
		{"far term generic is stub", vo.TierFarTerm, vo.SubdomainGeneric, vo.TicketDetailStub},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tier := vo.NewTierClassification(tt.tier, "test")
			result := vo.TierToDetailLevel(tier, tt.classification)
			assert.Equal(t, tt.want, result)
		})
	}
}
