"""Tests for TicketFlagged and FlagCleared domain events."""

from __future__ import annotations

import pytest

from src.domain.models.ticket_freshness import ContextDiff


class TestTicketFlagged:
    def test_stores_fields(self) -> None:
        from src.domain.events.ticket_freshness_events import TicketFlagged

        diff = ContextDiff(
            summary="New fitness tests",
            triggering_ticket_id="k7m.19",
            produced_at="2026-03-01",
        )
        event = TicketFlagged(
            review_id="rr-001",
            ticket_id="k7m.25",
            context_diff=diff,
            flagged_at="2026-03-01T10:00:00",
        )
        assert event.review_id == "rr-001"
        assert event.ticket_id == "k7m.25"
        assert event.context_diff is diff
        assert event.flagged_at == "2026-03-01T10:00:00"

    def test_frozen(self) -> None:
        from src.domain.events.ticket_freshness_events import TicketFlagged

        diff = ContextDiff(
            summary="Change",
            triggering_ticket_id="k7m.19",
            produced_at="2026-03-01",
        )
        event = TicketFlagged(
            review_id="rr-001",
            ticket_id="k7m.25",
            context_diff=diff,
            flagged_at="2026-03-01T10:00:00",
        )
        with pytest.raises(AttributeError):
            event.review_id = "changed"  # type: ignore[misc]


class TestFlagCleared:
    def test_stores_fields(self) -> None:
        from src.domain.events.ticket_freshness_events import FlagCleared

        event = FlagCleared(
            review_id="rr-001",
            ticket_id="k7m.25",
            cleared_at="2026-03-01T11:00:00",
        )
        assert event.review_id == "rr-001"
        assert event.ticket_id == "k7m.25"
        assert event.cleared_at == "2026-03-01T11:00:00"

    def test_frozen(self) -> None:
        from src.domain.events.ticket_freshness_events import FlagCleared

        event = FlagCleared(
            review_id="rr-001",
            ticket_id="k7m.25",
            cleared_at="2026-03-01T11:00:00",
        )
        with pytest.raises(AttributeError):
            event.review_id = "changed"  # type: ignore[misc]
