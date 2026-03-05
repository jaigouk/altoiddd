"""Tests for tier classification (alty-2j7.11).

Pure function tests for classify_tier() and tier_to_detail_level().
"""

from __future__ import annotations

from src.domain.models.domain_values import SubdomainClassification
from src.domain.models.ticket_values import (
    TicketDetailLevel,
    Tier,
    TierClassification,
    classify_tier,
    tier_to_detail_level,
)

# ---------------------------------------------------------------------------
# classify_tier: Core override
# ---------------------------------------------------------------------------


class TestCoreOverride:
    """Core subdomain is always near-term regardless of depth."""

    def test_core_depth_0_near_term(self):
        result = classify_tier(0, SubdomainClassification.CORE)
        assert result.tier == Tier.NEAR_TERM

    def test_core_depth_5_near_term(self):
        result = classify_tier(5, SubdomainClassification.CORE)
        assert result.tier == Tier.NEAR_TERM

    def test_core_depth_100_near_term(self):
        result = classify_tier(100, SubdomainClassification.CORE)
        assert result.tier == Tier.NEAR_TERM

    def test_core_reason_mentions_core(self):
        result = classify_tier(5, SubdomainClassification.CORE)
        assert "Core" in result.reason or "core" in result.reason


# ---------------------------------------------------------------------------
# classify_tier: depth boundary
# ---------------------------------------------------------------------------


class TestDepthBoundary:
    """Depth <=2 is near-term, >2 is far-term for non-Core."""

    def test_supporting_depth_0_near_term(self):
        result = classify_tier(0, SubdomainClassification.SUPPORTING)
        assert result.tier == Tier.NEAR_TERM

    def test_supporting_depth_2_near_term(self):
        result = classify_tier(2, SubdomainClassification.SUPPORTING)
        assert result.tier == Tier.NEAR_TERM

    def test_supporting_depth_3_far_term(self):
        result = classify_tier(3, SubdomainClassification.SUPPORTING)
        assert result.tier == Tier.FAR_TERM

    def test_generic_depth_1_near_term(self):
        result = classify_tier(1, SubdomainClassification.GENERIC)
        assert result.tier == Tier.NEAR_TERM

    def test_generic_depth_3_far_term(self):
        result = classify_tier(3, SubdomainClassification.GENERIC)
        assert result.tier == Tier.FAR_TERM


# ---------------------------------------------------------------------------
# tier_to_detail_level
# ---------------------------------------------------------------------------


class TestTierToDetailLevel:
    """Tier + classification → TicketDetailLevel mapping."""

    def test_near_term_core_is_full(self):
        tier = TierClassification(Tier.NEAR_TERM, "test")
        assert tier_to_detail_level(tier, SubdomainClassification.CORE) == TicketDetailLevel.FULL

    def test_near_term_supporting_is_standard(self):
        tier = TierClassification(Tier.NEAR_TERM, "test")
        assert (
            tier_to_detail_level(tier, SubdomainClassification.SUPPORTING)
            == TicketDetailLevel.STANDARD
        )

    def test_near_term_generic_is_stub(self):
        tier = TierClassification(Tier.NEAR_TERM, "test")
        assert (
            tier_to_detail_level(tier, SubdomainClassification.GENERIC) == TicketDetailLevel.STUB
        )

    def test_far_term_supporting_is_stub(self):
        tier = TierClassification(Tier.FAR_TERM, "test")
        result = tier_to_detail_level(tier, SubdomainClassification.SUPPORTING)
        assert result == TicketDetailLevel.STUB

    def test_far_term_core_is_full(self):
        """Core override: even FAR_TERM Core stays FULL (domain invariant)."""
        tier = TierClassification(Tier.FAR_TERM, "test")
        assert tier_to_detail_level(tier, SubdomainClassification.CORE) == TicketDetailLevel.FULL

    def test_far_term_generic_is_stub(self):
        tier = TierClassification(Tier.FAR_TERM, "test")
        assert (
            tier_to_detail_level(tier, SubdomainClassification.GENERIC) == TicketDetailLevel.STUB
        )
