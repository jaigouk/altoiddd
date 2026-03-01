"""Tests for TicketHealthHandler query handler."""

from __future__ import annotations

from src.domain.models.ticket_freshness import (
    ContextDiff,
    FreshnessFlag,
    OpenTicketData,
    TicketFreshnessStatus,
)

# ---------------------------------------------------------------------------
# Fake reader for testing
# ---------------------------------------------------------------------------


class FakeTicketReader:
    """In-memory stub implementing TicketReaderProtocol."""

    def __init__(
        self,
        open_tickets: tuple[OpenTicketData, ...] = (),
        flags_by_ticket: dict[str, tuple[FreshnessFlag, ...]] | None = None,
    ) -> None:
        self._open_tickets = open_tickets
        self._flags_by_ticket = flags_by_ticket or {}

    def read_open_tickets(self) -> tuple[OpenTicketData, ...]:
        return self._open_tickets

    def read_flags(self, ticket_id: str) -> tuple[FreshnessFlag, ...]:
        return self._flags_by_ticket.get(ticket_id, ())


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _make_diff(summary: str = "Change") -> ContextDiff:
    return ContextDiff(
        summary=summary,
        triggering_ticket_id="k7m.19",
        produced_at="2026-03-01",
    )


def _make_flag(summary: str = "Change") -> FreshnessFlag:
    return FreshnessFlag(
        context_diff=_make_diff(summary),
        flagged_at="2026-03-01T10:00:00",
    )


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------


class TestTicketHealthHandler:
    def test_handler_reads_open_tickets_and_flags(self) -> None:
        from src.application.queries.ticket_health_handler import TicketHealthHandler

        flag = _make_flag()
        reader = FakeTicketReader(
            open_tickets=(
                OpenTicketData(
                    ticket_id="k7m.25",
                    title="Ticket Health",
                    labels=("review_needed",),
                    last_reviewed="2026-02-28",
                ),
                OpenTicketData(
                    ticket_id="k7m.20",
                    title="Ticket Gen",
                    labels=(),
                    last_reviewed="2026-02-25",
                ),
            ),
            flags_by_ticket={
                "k7m.25": (flag,),
            },
        )
        handler = TicketHealthHandler(reader=reader)
        report = handler.report()

        assert report.total_open == 2
        assert report.review_needed_count == 1
        assert report.flagged_tickets[0].ticket_id == "k7m.25"

    def test_handler_empty_report_no_open_tickets(self) -> None:
        from src.application.queries.ticket_health_handler import TicketHealthHandler

        reader = FakeTicketReader(open_tickets=())
        handler = TicketHealthHandler(reader=reader)
        report = handler.report()

        assert report.total_open == 0
        assert report.review_needed_count == 0
        assert report.has_issues is False

    def test_handler_includes_context_diffs(self) -> None:
        from src.application.queries.ticket_health_handler import TicketHealthHandler

        flag = _make_flag("Added new aggregate")
        reader = FakeTicketReader(
            open_tickets=(
                OpenTicketData(
                    ticket_id="k7m.25",
                    title="Ticket Health",
                    labels=("review_needed",),
                ),
            ),
            flags_by_ticket={
                "k7m.25": (flag,),
            },
        )
        handler = TicketHealthHandler(reader=reader)
        report = handler.report()

        assert len(report.flagged_tickets) == 1
        flagged = report.flagged_tickets[0]
        assert flagged.flags[0].context_diff.summary == "Added new aggregate"

    def test_handler_excludes_non_flagged_tickets(self) -> None:
        from src.application.queries.ticket_health_handler import TicketHealthHandler

        reader = FakeTicketReader(
            open_tickets=(
                OpenTicketData(
                    ticket_id="k7m.20",
                    title="No flags",
                    labels=(),
                    last_reviewed="2026-02-28",
                ),
                OpenTicketData(
                    ticket_id="k7m.21",
                    title="Also no flags",
                    labels=("some_other_label",),
                    last_reviewed="2026-02-27",
                ),
            ),
        )
        handler = TicketHealthHandler(reader=reader)
        report = handler.report()

        assert report.total_open == 2
        assert report.review_needed_count == 0
        assert report.has_issues is False

    def test_handler_finds_oldest_last_reviewed(self) -> None:
        from src.application.queries.ticket_health_handler import TicketHealthHandler

        reader = FakeTicketReader(
            open_tickets=(
                OpenTicketData(
                    ticket_id="k7m.20",
                    title="Old",
                    labels=(),
                    last_reviewed="2026-01-15",
                ),
                OpenTicketData(
                    ticket_id="k7m.21",
                    title="Newer",
                    labels=(),
                    last_reviewed="2026-02-28",
                ),
                OpenTicketData(
                    ticket_id="k7m.22",
                    title="Never reviewed",
                    labels=(),
                    last_reviewed=None,
                ),
            ),
        )
        handler = TicketHealthHandler(reader=reader)
        report = handler.report()

        assert report.oldest_last_reviewed == "2026-01-15"

    def test_handler_oldest_last_reviewed_none_when_all_none(self) -> None:
        from src.application.queries.ticket_health_handler import TicketHealthHandler

        reader = FakeTicketReader(
            open_tickets=(
                OpenTicketData(
                    ticket_id="k7m.20",
                    title="Never reviewed",
                    labels=(),
                    last_reviewed=None,
                ),
            ),
        )
        handler = TicketHealthHandler(reader=reader)
        report = handler.report()

        assert report.oldest_last_reviewed is None

    def test_handler_flagged_ticket_status_is_review_needed(self) -> None:
        from src.application.queries.ticket_health_handler import TicketHealthHandler

        flag = _make_flag()
        reader = FakeTicketReader(
            open_tickets=(
                OpenTicketData(
                    ticket_id="k7m.25",
                    title="Ticket Health",
                    labels=("review_needed",),
                ),
            ),
            flags_by_ticket={"k7m.25": (flag,)},
        )
        handler = TicketHealthHandler(reader=reader)
        report = handler.report()

        assert report.flagged_tickets[0].status == TicketFreshnessStatus.REVIEW_NEEDED
