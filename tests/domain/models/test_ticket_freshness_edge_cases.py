"""Edge case tests for Ticket Freshness domain models (2j7.9 QA).

BICEP analysis uncovered:
- Boundary: EpicHealthSummary negative fresh_count and stale_count individually
- Boundary: freshness_pct return type consistency
- Boundary: TicketHealthReport with flagged > total_open (logical inconsistency)
- Inverse: EpicHealthSummary with zero-of-each
- Error: ContextDiff with very long summary
"""

from __future__ import annotations

import pytest

from src.domain.models.errors import InvariantViolationError


class TestEpicHealthSummaryNegativeCounts:
    """Each individual count field must be non-negative."""

    def test_rejects_negative_fresh_count(self) -> None:
        from src.domain.models.ticket_freshness import EpicHealthSummary

        with pytest.raises(InvariantViolationError, match="non-negative"):
            EpicHealthSummary(
                epic_id="k7m",
                total_tickets=5,
                fresh_count=-1,
                stale_count=6,
            )

    def test_rejects_negative_stale_count(self) -> None:
        from src.domain.models.ticket_freshness import EpicHealthSummary

        with pytest.raises(InvariantViolationError, match="non-negative"):
            EpicHealthSummary(
                epic_id="k7m",
                total_tickets=5,
                fresh_count=6,
                stale_count=-1,
            )

    def test_rejects_negative_total_and_matching_sum(self) -> None:
        """Even if fresh + stale == total, negative total is invalid."""
        from src.domain.models.ticket_freshness import EpicHealthSummary

        with pytest.raises(InvariantViolationError, match="non-negative"):
            EpicHealthSummary(
                epic_id="k7m",
                total_tickets=-2,
                fresh_count=-1,
                stale_count=-1,
            )


class TestTicketHealthReportFreshnessPctType:
    """freshness_pct must always return a float."""

    def test_freshness_pct_returns_float_when_all_fresh(self) -> None:
        from src.domain.models.ticket_freshness import TicketHealthReport

        report = TicketHealthReport(flagged_tickets=(), total_open=10)
        result = report.freshness_pct
        assert isinstance(result, float)

    def test_freshness_pct_returns_float_when_zero_open(self) -> None:
        from src.domain.models.ticket_freshness import TicketHealthReport

        report = TicketHealthReport(flagged_tickets=(), total_open=0)
        result = report.freshness_pct
        assert isinstance(result, float)

    def test_freshness_pct_returns_float_when_partial(self) -> None:
        from src.domain.models.ticket_freshness import (
            ContextDiff,
            FlaggedTicket,
            FreshnessFlag,
            TicketFreshnessStatus,
            TicketHealthReport,
        )

        diff = ContextDiff(
            summary="Change",
            triggering_ticket_id="k7m.19",
            produced_at="2026-03-01",
        )
        flag = FreshnessFlag(context_diff=diff, flagged_at="2026-03-01T10:00:00")
        flagged = (
            FlaggedTicket(
                ticket_id="k7m.25",
                title="Test",
                flags=(flag,),
                status=TicketFreshnessStatus.REVIEW_NEEDED,
            ),
        )
        report = TicketHealthReport(flagged_tickets=flagged, total_open=3)
        result = report.freshness_pct
        assert isinstance(result, float)
        # 2/3 * 100 = 66.666...
        assert abs(result - 66.66666666666667) < 0.001


class TestTicketHealthReportBoundary:
    """Boundary tests for TicketHealthReport."""

    def test_single_open_no_flags(self) -> None:
        """1 open, 0 flags -> 100%."""
        from src.domain.models.ticket_freshness import TicketHealthReport

        report = TicketHealthReport(flagged_tickets=(), total_open=1)
        assert report.freshness_pct == 100.0

    def test_large_number_of_open_tickets(self) -> None:
        """Performance boundary: large total_open."""
        from src.domain.models.ticket_freshness import TicketHealthReport

        report = TicketHealthReport(flagged_tickets=(), total_open=100_000)
        assert report.freshness_pct == 100.0


class TestContextDiffEdgeCases:
    """Additional edge cases for ContextDiff validation."""

    def test_context_diff_accepts_long_summary(self) -> None:
        """Very long summaries are valid (no upper bound on summary length)."""
        from src.domain.models.ticket_freshness import ContextDiff

        long_summary = "A" * 10_000
        diff = ContextDiff(
            summary=long_summary,
            triggering_ticket_id="k7m.19",
            produced_at="2026-03-01",
        )
        assert len(diff.summary) == 10_000

    def test_context_diff_accepts_special_characters(self) -> None:
        """Summaries with special chars, newlines, unicode are valid."""
        from src.domain.models.ticket_freshness import ContextDiff

        diff = ContextDiff(
            summary="Added module with\nnewlines & 'quotes' + unicode: \u2713",
            triggering_ticket_id="k7m.19",
            produced_at="2026-03-01",
        )
        assert "\n" in diff.summary


class TestFlaggedTicketEdgeCases:
    """Edge cases for FlaggedTicket VO."""

    def test_flagged_ticket_with_zero_flags(self) -> None:
        """A ticket can have zero flags (fresh status)."""
        from src.domain.models.ticket_freshness import (
            FlaggedTicket,
            TicketFreshnessStatus,
        )

        ticket = FlaggedTicket(
            ticket_id="k7m.25",
            title="Test",
            flags=(),
            status=TicketFreshnessStatus.FRESH,
        )
        assert ticket.flag_count == 0

    def test_flagged_ticket_is_frozen(self) -> None:
        from src.domain.models.ticket_freshness import (
            FlaggedTicket,
            TicketFreshnessStatus,
        )

        ticket = FlaggedTicket(
            ticket_id="k7m.25",
            title="Test",
            flags=(),
            status=TicketFreshnessStatus.FRESH,
        )
        with pytest.raises(AttributeError):
            ticket.ticket_id = "changed"  # type: ignore[misc]
