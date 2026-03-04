"""Edge case tests for RippleReview aggregate (2j7.9 QA).

BICEP analysis uncovered:
- Boundary: build_ripple_comment with special characters in context diff
- Boundary: flag the same ticket twice (duplicate flagging / stacking)
- Inverse: clear all flags then verify empty state
- Error: clear_flag after double-flagging same ticket (should remove once)
- Cross-check: comment text includes both closed_ticket_id and summary
"""

from __future__ import annotations

from typing import TYPE_CHECKING

import pytest

from src.domain.models.errors import InvariantViolationError

if TYPE_CHECKING:
    from src.domain.models.ticket_freshness import ContextDiff


def _make_context_diff(summary: str = "Implemented fitness test generation") -> ContextDiff:
    from src.domain.models.ticket_freshness import ContextDiff

    return ContextDiff(
        summary=summary,
        triggering_ticket_id="k7m.19",
        produced_at="2026-03-01",
    )


class TestBuildRippleCommentEdgeCases:
    """build_ripple_comment with special content."""

    def test_comment_includes_closed_ticket_id(self) -> None:
        from src.domain.models.ripple_review import RippleReview

        review = RippleReview(
            review_id="rr-001",
            closed_ticket_id="k7m.42",
            context_diff=_make_context_diff(),
        )
        comment = review.build_ripple_comment()
        assert "k7m.42" in comment

    def test_comment_includes_context_summary(self) -> None:
        from src.domain.models.ripple_review import RippleReview

        review = RippleReview(
            review_id="rr-001",
            closed_ticket_id="k7m.19",
            context_diff=_make_context_diff("Added new StackProfile protocol"),
        )
        comment = review.build_ripple_comment()
        assert "Added new StackProfile protocol" in comment

    def test_comment_with_special_chars_in_summary(self) -> None:
        from src.domain.models.ripple_review import RippleReview

        review = RippleReview(
            review_id="rr-001",
            closed_ticket_id="k7m.19",
            context_diff=_make_context_diff("Added `code` & 'quotes' + <tags>"),
        )
        comment = review.build_ripple_comment()
        assert "`code`" in comment
        assert "<tags>" in comment

    def test_comment_includes_checklist_items(self) -> None:
        from src.domain.models.ripple_review import (
            REVIEW_CHECKLIST_TEMPLATE,
            RippleReview,
        )

        review = RippleReview(
            review_id="rr-001",
            closed_ticket_id="k7m.19",
            context_diff=_make_context_diff(),
        )
        comment = review.build_ripple_comment()
        # Verify checklist is embedded in the comment
        assert REVIEW_CHECKLIST_TEMPLATE in comment


class TestDuplicateFlagging:
    """Flagging the same ticket multiple times (stacking)."""

    def test_flag_same_ticket_twice(self) -> None:
        """Stacking: same ticket flagged twice should appear twice."""
        from src.domain.models.ripple_review import RippleReview

        review = RippleReview(
            review_id="rr-001",
            closed_ticket_id="k7m.19",
            context_diff=_make_context_diff(),
        )
        review.flag_ticket(ticket_id="k7m.25", is_open=True)
        review.flag_ticket(ticket_id="k7m.25", is_open=True)
        # Both flags stored (stacking behavior)
        flagged = review.flagged_tickets
        assert flagged.count("k7m.25") == 2

    def test_clear_one_instance_of_duplicate_flag(self) -> None:
        """Clearing removes only the first occurrence."""
        from src.domain.models.ripple_review import RippleReview

        review = RippleReview(
            review_id="rr-001",
            closed_ticket_id="k7m.19",
            context_diff=_make_context_diff(),
        )
        review.flag_ticket(ticket_id="k7m.25", is_open=True)
        review.flag_ticket(ticket_id="k7m.25", is_open=True)
        review.clear_flag(ticket_id="k7m.25")
        # One instance removed, one remains
        assert review.flagged_tickets.count("k7m.25") == 1


class TestRippleReviewFullLifecycle:
    """Inverse: full lifecycle from flagging to clearing all."""

    def test_flag_three_then_clear_all(self) -> None:
        from src.domain.models.ripple_review import RippleReview

        review = RippleReview(
            review_id="rr-001",
            closed_ticket_id="k7m.19",
            context_diff=_make_context_diff(),
        )
        review.flag_ticket(ticket_id="k7m.25", is_open=True)
        review.flag_ticket(ticket_id="k7m.20", is_open=True)
        review.flag_ticket(ticket_id="k7m.21", is_open=True)
        assert len(review.flagged_tickets) == 3

        review.clear_flag(ticket_id="k7m.25")
        review.clear_flag(ticket_id="k7m.20")
        review.clear_flag(ticket_id="k7m.21")
        assert len(review.flagged_tickets) == 0

    def test_clear_already_cleared_raises(self) -> None:
        """After clearing, clearing again raises InvariantViolationError."""
        from src.domain.models.ripple_review import RippleReview

        review = RippleReview(
            review_id="rr-001",
            closed_ticket_id="k7m.19",
            context_diff=_make_context_diff(),
        )
        review.flag_ticket(ticket_id="k7m.25", is_open=True)
        review.clear_flag(ticket_id="k7m.25")

        with pytest.raises(InvariantViolationError, match="not flagged"):
            review.clear_flag(ticket_id="k7m.25")

    def test_events_accumulate_through_lifecycle(self) -> None:
        """Events from both flag and clear operations are accumulated."""
        from src.domain.events.ticket_freshness_events import FlagCleared, TicketFlagged
        from src.domain.models.ripple_review import RippleReview

        review = RippleReview(
            review_id="rr-001",
            closed_ticket_id="k7m.19",
            context_diff=_make_context_diff(),
        )
        review.flag_ticket(ticket_id="k7m.25", is_open=True)
        review.flag_ticket(ticket_id="k7m.20", is_open=True)
        review.clear_flag(ticket_id="k7m.25")

        events = review.events
        assert len(events) == 3
        assert isinstance(events[0], TicketFlagged)
        assert isinstance(events[1], TicketFlagged)
        assert isinstance(events[2], FlagCleared)
