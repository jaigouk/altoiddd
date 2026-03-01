"""Tests for Ticket Freshness domain value objects (CORE subdomain).

Covers ContextDiff, FreshnessFlag, TicketFreshnessStatus, FlaggedTicket,
TicketHealthReport, and OpenTicketData — all frozen dataclass VOs.
"""

from __future__ import annotations

import pytest

from src.domain.models.errors import InvariantViolationError

# ---------------------------------------------------------------------------
# 1. ContextDiff Value Object
# ---------------------------------------------------------------------------


class TestContextDiff:
    def test_context_diff_valid(self) -> None:
        from src.domain.models.ticket_freshness import ContextDiff

        diff = ContextDiff(
            summary="Added order validation",
            triggering_ticket_id="k7m.19",
            produced_at="2026-03-01",
        )
        assert diff.summary == "Added order validation"
        assert diff.triggering_ticket_id == "k7m.19"
        assert diff.produced_at == "2026-03-01"

    def test_context_diff_rejects_empty_summary(self) -> None:
        from src.domain.models.ticket_freshness import ContextDiff

        with pytest.raises(InvariantViolationError, match="summary"):
            ContextDiff(
                summary="",
                triggering_ticket_id="k7m.19",
                produced_at="2026-03-01",
            )

    def test_context_diff_rejects_whitespace_summary(self) -> None:
        from src.domain.models.ticket_freshness import ContextDiff

        with pytest.raises(InvariantViolationError, match="summary"):
            ContextDiff(
                summary="   \t\n  ",
                triggering_ticket_id="k7m.19",
                produced_at="2026-03-01",
            )

    def test_context_diff_is_frozen(self) -> None:
        from src.domain.models.ticket_freshness import ContextDiff

        diff = ContextDiff(
            summary="Some change",
            triggering_ticket_id="k7m.19",
            produced_at="2026-03-01",
        )
        with pytest.raises(AttributeError):
            diff.summary = "changed"  # type: ignore[misc]


# ---------------------------------------------------------------------------
# 2. FreshnessFlag Value Object
# ---------------------------------------------------------------------------


class TestFreshnessFlag:
    def test_freshness_flag_stores_context_diff(self) -> None:
        from src.domain.models.ticket_freshness import ContextDiff, FreshnessFlag

        diff = ContextDiff(
            summary="Implemented fitness tests",
            triggering_ticket_id="k7m.19",
            produced_at="2026-03-01",
        )
        flag = FreshnessFlag(context_diff=diff, flagged_at="2026-03-01T10:00:00")
        assert flag.context_diff is diff
        assert flag.flagged_at == "2026-03-01T10:00:00"

    def test_freshness_flag_is_frozen(self) -> None:
        from src.domain.models.ticket_freshness import ContextDiff, FreshnessFlag

        diff = ContextDiff(
            summary="Change",
            triggering_ticket_id="k7m.19",
            produced_at="2026-03-01",
        )
        flag = FreshnessFlag(context_diff=diff, flagged_at="2026-03-01T10:00:00")
        with pytest.raises(AttributeError):
            flag.flagged_at = "changed"  # type: ignore[misc]


# ---------------------------------------------------------------------------
# 3. TicketFreshnessStatus Enum
# ---------------------------------------------------------------------------


class TestTicketFreshnessStatus:
    def test_ticket_freshness_status_enum_values(self) -> None:
        from src.domain.models.ticket_freshness import TicketFreshnessStatus

        assert TicketFreshnessStatus.FRESH.value == "fresh"
        assert TicketFreshnessStatus.REVIEW_NEEDED.value == "review_needed"
        assert TicketFreshnessStatus.NEVER_REVIEWED.value == "never_reviewed"


# ---------------------------------------------------------------------------
# 4. FlaggedTicket Value Object
# ---------------------------------------------------------------------------


class TestFlaggedTicket:
    def test_flagged_ticket_with_single_flag(self) -> None:
        from src.domain.models.ticket_freshness import (
            ContextDiff,
            FlaggedTicket,
            FreshnessFlag,
            TicketFreshnessStatus,
        )

        diff = ContextDiff(
            summary="New module added",
            triggering_ticket_id="k7m.20",
            produced_at="2026-03-01",
        )
        flag = FreshnessFlag(context_diff=diff, flagged_at="2026-03-01T12:00:00")
        ticket = FlaggedTicket(
            ticket_id="k7m.25",
            title="Ticket Health",
            flags=(flag,),
            status=TicketFreshnessStatus.REVIEW_NEEDED,
        )
        assert ticket.ticket_id == "k7m.25"
        assert ticket.title == "Ticket Health"
        assert len(ticket.flags) == 1
        assert ticket.status == TicketFreshnessStatus.REVIEW_NEEDED

    def test_flagged_ticket_with_multiple_flags(self) -> None:
        """Invariant 3: flags stack -- a ticket can have multiple flags."""
        from src.domain.models.ticket_freshness import (
            ContextDiff,
            FlaggedTicket,
            FreshnessFlag,
            TicketFreshnessStatus,
        )

        diff1 = ContextDiff(
            summary="Change one",
            triggering_ticket_id="k7m.19",
            produced_at="2026-02-28",
        )
        diff2 = ContextDiff(
            summary="Change two",
            triggering_ticket_id="k7m.20",
            produced_at="2026-03-01",
        )
        flags = (
            FreshnessFlag(context_diff=diff1, flagged_at="2026-02-28T10:00:00"),
            FreshnessFlag(context_diff=diff2, flagged_at="2026-03-01T10:00:00"),
        )
        ticket = FlaggedTicket(
            ticket_id="k7m.25",
            title="Ticket Health",
            flags=flags,
            status=TicketFreshnessStatus.REVIEW_NEEDED,
        )
        assert len(ticket.flags) == 2

    def test_flagged_ticket_flag_count(self) -> None:
        from src.domain.models.ticket_freshness import (
            ContextDiff,
            FlaggedTicket,
            FreshnessFlag,
            TicketFreshnessStatus,
        )

        diff = ContextDiff(
            summary="Change",
            triggering_ticket_id="k7m.19",
            produced_at="2026-03-01",
        )
        flags = tuple(
            FreshnessFlag(context_diff=diff, flagged_at=f"2026-03-01T{i:02d}:00:00")
            for i in range(3)
        )
        ticket = FlaggedTicket(
            ticket_id="k7m.25",
            title="Ticket Health",
            flags=flags,
            status=TicketFreshnessStatus.REVIEW_NEEDED,
        )
        assert ticket.flag_count == 3


# ---------------------------------------------------------------------------
# 5. TicketHealthReport Value Object
# ---------------------------------------------------------------------------


class TestTicketHealthReport:
    def test_ticket_health_report_review_needed_count(self) -> None:
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
        flagged = FlaggedTicket(
            ticket_id="k7m.25",
            title="Ticket Health",
            flags=(flag,),
            status=TicketFreshnessStatus.REVIEW_NEEDED,
        )
        report = TicketHealthReport(
            flagged_tickets=(flagged,),
            total_open=5,
        )
        assert report.review_needed_count == 1

    def test_ticket_health_report_has_issues_true(self) -> None:
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
        flagged = FlaggedTicket(
            ticket_id="k7m.25",
            title="Ticket Health",
            flags=(flag,),
            status=TicketFreshnessStatus.REVIEW_NEEDED,
        )
        report = TicketHealthReport(
            flagged_tickets=(flagged,),
            total_open=5,
        )
        assert report.has_issues is True

    def test_ticket_health_report_has_issues_false(self) -> None:
        from src.domain.models.ticket_freshness import TicketHealthReport

        report = TicketHealthReport(
            flagged_tickets=(),
            total_open=5,
        )
        assert report.has_issues is False

    def test_ticket_health_report_oldest_last_reviewed(self) -> None:
        from src.domain.models.ticket_freshness import TicketHealthReport

        report = TicketHealthReport(
            flagged_tickets=(),
            total_open=10,
            oldest_last_reviewed="2026-01-15",
        )
        assert report.oldest_last_reviewed == "2026-01-15"

    def test_ticket_health_report_oldest_last_reviewed_default_none(self) -> None:
        from src.domain.models.ticket_freshness import TicketHealthReport

        report = TicketHealthReport(
            flagged_tickets=(),
            total_open=0,
        )
        assert report.oldest_last_reviewed is None


# ---------------------------------------------------------------------------
# 6. OpenTicketData Value Object
# ---------------------------------------------------------------------------


class TestOpenTicketData:
    def test_open_ticket_data_fields(self) -> None:
        from src.domain.models.ticket_freshness import OpenTicketData

        data = OpenTicketData(
            ticket_id="k7m.25",
            title="Ticket Health",
            labels=("review_needed", "core"),
            last_reviewed="2026-02-28",
        )
        assert data.ticket_id == "k7m.25"
        assert data.title == "Ticket Health"
        assert data.labels == ("review_needed", "core")
        assert data.last_reviewed == "2026-02-28"

    def test_open_ticket_data_default_last_reviewed(self) -> None:
        from src.domain.models.ticket_freshness import OpenTicketData

        data = OpenTicketData(
            ticket_id="k7m.25",
            title="Ticket Health",
            labels=(),
        )
        assert data.last_reviewed is None

    def test_open_ticket_data_is_frozen(self) -> None:
        from src.domain.models.ticket_freshness import OpenTicketData

        data = OpenTicketData(
            ticket_id="k7m.25",
            title="Ticket Health",
            labels=(),
        )
        with pytest.raises(AttributeError):
            data.ticket_id = "changed"  # type: ignore[misc]
