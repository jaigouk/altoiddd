"""Tests for TicketPlanApproved domain event."""

from __future__ import annotations

import pytest

from src.domain.events.ticket_events import TicketPlanApproved


class TestTicketPlanApproved:
    def test_frozen(self):
        event = TicketPlanApproved(
            plan_id="plan-1",
            approved_ticket_ids=("t-1", "t-2"),
            dismissed_ticket_ids=("t-3",),
        )
        with pytest.raises(AttributeError):
            event.plan_id = "changed"  # type: ignore[misc]

    def test_stores_fields(self):
        event = TicketPlanApproved(
            plan_id="plan-1",
            approved_ticket_ids=("t-1", "t-2"),
            dismissed_ticket_ids=("t-3",),
        )
        assert event.plan_id == "plan-1"
        assert event.approved_ticket_ids == ("t-1", "t-2")
        assert event.dismissed_ticket_ids == ("t-3",)

    def test_empty_tuples_allowed(self):
        event = TicketPlanApproved(
            plan_id="plan-1",
            approved_ticket_ids=(),
            dismissed_ticket_ids=(),
        )
        assert event.approved_ticket_ids == ()
        assert event.dismissed_ticket_ids == ()
