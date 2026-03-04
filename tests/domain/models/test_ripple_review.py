"""Tests for RippleReview aggregate root (CORE subdomain).

Covers invariants:
1. Non-empty context diff (enforced by ContextDiff VO)
2. Only open tickets can be flagged
3. Stacking allowed (multiple flags per ticket)
4. Explicit review required to clear a flag
"""

from __future__ import annotations

from typing import TYPE_CHECKING

import pytest

from src.domain.models.errors import InvariantViolationError

if TYPE_CHECKING:
    from src.domain.models.ticket_freshness import ContextDiff

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _make_context_diff() -> ContextDiff:
    from src.domain.models.ticket_freshness import ContextDiff

    return ContextDiff(
        summary="Implemented fitness test generation",
        triggering_ticket_id="k7m.19",
        produced_at="2026-03-01",
    )


# ---------------------------------------------------------------------------
# 1. Creation
# ---------------------------------------------------------------------------


class TestRippleReviewCreation:
    def test_ripple_review_creation(self) -> None:
        from src.domain.models.ripple_review import RippleReview

        diff = _make_context_diff()
        review = RippleReview(
            review_id="rr-001",
            closed_ticket_id="k7m.19",
            context_diff=diff,
        )
        assert review.review_id == "rr-001"
        assert review.closed_ticket_id == "k7m.19"
        assert review.context_diff is diff
        assert review.flagged_tickets == ()
        assert review.events == ()


# ---------------------------------------------------------------------------
# 2. Flagging tickets
# ---------------------------------------------------------------------------


class TestFlagTicket:
    def test_ripple_review_flag_open_ticket(self) -> None:
        from src.domain.models.ripple_review import RippleReview

        review = RippleReview(
            review_id="rr-001",
            closed_ticket_id="k7m.19",
            context_diff=_make_context_diff(),
        )
        review.flag_ticket(ticket_id="k7m.25", is_open=True)
        assert "k7m.25" in review.flagged_tickets

    def test_ripple_review_rejects_closed_ticket(self) -> None:
        """Invariant 2: only open tickets can be flagged."""
        from src.domain.models.ripple_review import RippleReview

        review = RippleReview(
            review_id="rr-001",
            closed_ticket_id="k7m.19",
            context_diff=_make_context_diff(),
        )
        with pytest.raises(InvariantViolationError, match="open"):
            review.flag_ticket(ticket_id="k7m.18", is_open=False)

    def test_ripple_review_flag_multiple_tickets(self) -> None:
        """Invariant 3: multiple tickets can be flagged."""
        from src.domain.models.ripple_review import RippleReview

        review = RippleReview(
            review_id="rr-001",
            closed_ticket_id="k7m.19",
            context_diff=_make_context_diff(),
        )
        review.flag_ticket(ticket_id="k7m.25", is_open=True)
        review.flag_ticket(ticket_id="k7m.20", is_open=True)
        assert set(review.flagged_tickets) == {"k7m.25", "k7m.20"}


# ---------------------------------------------------------------------------
# 3. Clearing flags
# ---------------------------------------------------------------------------


class TestClearFlag:
    def test_ripple_review_clear_flag(self) -> None:
        from src.domain.models.ripple_review import RippleReview

        review = RippleReview(
            review_id="rr-001",
            closed_ticket_id="k7m.19",
            context_diff=_make_context_diff(),
        )
        review.flag_ticket(ticket_id="k7m.25", is_open=True)
        review.clear_flag(ticket_id="k7m.25")
        assert "k7m.25" not in review.flagged_tickets

    def test_ripple_review_clear_unflagged_raises(self) -> None:
        """Invariant 4: cannot clear a flag that does not exist."""
        from src.domain.models.ripple_review import RippleReview

        review = RippleReview(
            review_id="rr-001",
            closed_ticket_id="k7m.19",
            context_diff=_make_context_diff(),
        )
        with pytest.raises(InvariantViolationError, match="not flagged"):
            review.clear_flag(ticket_id="k7m.25")


# ---------------------------------------------------------------------------
# 4. Domain events
# ---------------------------------------------------------------------------


class TestRippleReviewEvents:
    def test_ripple_review_emits_ticket_flagged_event(self) -> None:
        from src.domain.events.ticket_freshness_events import TicketFlagged
        from src.domain.models.ripple_review import RippleReview

        review = RippleReview(
            review_id="rr-001",
            closed_ticket_id="k7m.19",
            context_diff=_make_context_diff(),
        )
        review.flag_ticket(ticket_id="k7m.25", is_open=True)

        assert len(review.events) == 1
        event = review.events[0]
        assert isinstance(event, TicketFlagged)
        assert event.review_id == "rr-001"
        assert event.ticket_id == "k7m.25"
        assert event.context_diff is review.context_diff

    def test_ripple_review_emits_flag_cleared_event(self) -> None:
        from src.domain.events.ticket_freshness_events import FlagCleared
        from src.domain.models.ripple_review import RippleReview

        review = RippleReview(
            review_id="rr-001",
            closed_ticket_id="k7m.19",
            context_diff=_make_context_diff(),
        )
        review.flag_ticket(ticket_id="k7m.25", is_open=True)
        review.clear_flag(ticket_id="k7m.25")

        # Should have 2 events: TicketFlagged then FlagCleared
        assert len(review.events) == 2
        event = review.events[1]
        assert isinstance(event, FlagCleared)
        assert event.review_id == "rr-001"
        assert event.ticket_id == "k7m.25"


# ---------------------------------------------------------------------------
# 5. Defensive copies
# ---------------------------------------------------------------------------


class TestDefensiveCopies:
    def test_ripple_review_flagged_tickets_defensive_copy(self) -> None:
        from src.domain.models.ripple_review import RippleReview

        review = RippleReview(
            review_id="rr-001",
            closed_ticket_id="k7m.19",
            context_diff=_make_context_diff(),
        )
        review.flag_ticket(ticket_id="k7m.25", is_open=True)

        # Returned tuple should be a copy, not the internal list
        flagged = review.flagged_tickets
        assert isinstance(flagged, tuple)
        # Modifying the tuple is impossible (it's a tuple), but verify
        # that two calls return equal but independent objects
        assert review.flagged_tickets == flagged

    def test_ripple_review_events_defensive_copy(self) -> None:
        from src.domain.models.ripple_review import RippleReview

        review = RippleReview(
            review_id="rr-001",
            closed_ticket_id="k7m.19",
            context_diff=_make_context_diff(),
        )
        review.flag_ticket(ticket_id="k7m.25", is_open=True)

        events = review.events
        assert isinstance(events, tuple)
        assert review.events == events


# ---------------------------------------------------------------------------
# 6. Review checklist constant (G1)
# ---------------------------------------------------------------------------


class TestReviewChecklist:
    def test_review_checklist_template_exists(self) -> None:
        """G1: REVIEW_CHECKLIST_TEMPLATE constant must exist."""
        from src.domain.models.ripple_review import REVIEW_CHECKLIST_TEMPLATE

        assert isinstance(REVIEW_CHECKLIST_TEMPLATE, str)
        assert len(REVIEW_CHECKLIST_TEMPLATE) > 0

    def test_review_checklist_template_has_key_items(self) -> None:
        """Checklist must contain review guidance items."""
        from src.domain.models.ripple_review import REVIEW_CHECKLIST_TEMPLATE

        # Must mention description/AC review, DDD alignment, and dismiss option
        assert "description" in REVIEW_CHECKLIST_TEMPLATE.lower()
        assert "acceptance criteria" in REVIEW_CHECKLIST_TEMPLATE.lower()
        assert "dismiss" in REVIEW_CHECKLIST_TEMPLATE.lower() or (
            "unchanged" in REVIEW_CHECKLIST_TEMPLATE.lower()
        )

    def test_review_checklist_used_in_build_ripple_comment(self) -> None:
        """The aggregate must produce comments that include the checklist."""
        from src.domain.models.ripple_review import RippleReview

        review = RippleReview(
            review_id="rr-001",
            closed_ticket_id="k7m.19",
            context_diff=_make_context_diff(),
        )
        comment = review.build_ripple_comment()
        assert isinstance(comment, str)
        assert "k7m.19" in comment
        assert "fitness" in comment.lower()
        # Must include checklist items
        assert "description" in comment.lower()
        assert "acceptance criteria" in comment.lower()
