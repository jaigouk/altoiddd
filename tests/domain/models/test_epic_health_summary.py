"""Tests for EpicHealthSummary Value Object (G8).

Covers construction, immutability, and freshness_pct computation.
"""

from __future__ import annotations

import pytest


class TestEpicHealthSummary:
    def test_epic_health_summary_creation(self) -> None:
        """Valid construction with all fields."""
        from src.domain.models.ticket_freshness import EpicHealthSummary

        summary = EpicHealthSummary(
            epic_id="k7m",
            total_tickets=10,
            fresh_count=7,
            stale_count=3,
        )
        assert summary.epic_id == "k7m"
        assert summary.total_tickets == 10
        assert summary.fresh_count == 7
        assert summary.stale_count == 3

    def test_epic_health_summary_is_frozen(self) -> None:
        """VO must be immutable."""
        from src.domain.models.ticket_freshness import EpicHealthSummary

        summary = EpicHealthSummary(
            epic_id="k7m",
            total_tickets=10,
            fresh_count=7,
            stale_count=3,
        )
        with pytest.raises(AttributeError):
            summary.epic_id = "changed"  # type: ignore[misc]

    def test_epic_health_summary_freshness_pct(self) -> None:
        """freshness_pct = fresh_count / total_tickets * 100."""
        from src.domain.models.ticket_freshness import EpicHealthSummary

        summary = EpicHealthSummary(
            epic_id="k7m",
            total_tickets=10,
            fresh_count=7,
            stale_count=3,
        )
        assert summary.freshness_pct == 70.0

    def test_epic_health_summary_all_fresh(self) -> None:
        """fresh_count == total_tickets -> 100%."""
        from src.domain.models.ticket_freshness import EpicHealthSummary

        summary = EpicHealthSummary(
            epic_id="k7m",
            total_tickets=5,
            fresh_count=5,
            stale_count=0,
        )
        assert summary.freshness_pct == 100.0

    def test_epic_health_summary_zero_tickets(self) -> None:
        """total_tickets=0 -> freshness_pct=100.0 (no division by zero)."""
        from src.domain.models.ticket_freshness import EpicHealthSummary

        summary = EpicHealthSummary(
            epic_id="k7m",
            total_tickets=0,
            fresh_count=0,
            stale_count=0,
        )
        assert summary.freshness_pct == 100.0

    def test_epic_health_summary_all_stale(self) -> None:
        """All tickets stale -> 0%."""
        from src.domain.models.ticket_freshness import EpicHealthSummary

        summary = EpicHealthSummary(
            epic_id="k7m",
            total_tickets=5,
            fresh_count=0,
            stale_count=5,
        )
        assert summary.freshness_pct == 0.0

    def test_epic_health_summary_rejects_mismatched_counts(self) -> None:
        """fresh_count + stale_count must equal total_tickets."""
        from src.domain.models.errors import InvariantViolationError
        from src.domain.models.ticket_freshness import EpicHealthSummary

        with pytest.raises(InvariantViolationError, match="must equal total_tickets"):
            EpicHealthSummary(
                epic_id="k7m",
                total_tickets=10,
                fresh_count=3,
                stale_count=2,
            )

    def test_epic_health_summary_rejects_negative_counts(self) -> None:
        """Counts must be non-negative."""
        from src.domain.models.errors import InvariantViolationError
        from src.domain.models.ticket_freshness import EpicHealthSummary

        with pytest.raises(InvariantViolationError, match="non-negative"):
            EpicHealthSummary(
                epic_id="k7m",
                total_tickets=-1,
                fresh_count=0,
                stale_count=0,
            )
