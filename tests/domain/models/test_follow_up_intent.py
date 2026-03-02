"""Tests for FollowUpIntent and FollowUpAuditResult value objects.

RED phase: these tests define the domain model contract for the
spike-to-ticket audit feature (Ticket Freshness bounded context).
"""

from __future__ import annotations

import pytest

from src.domain.models.errors import InvariantViolationError

# ── FollowUpIntent ──────────────────────────────────────────────────


class TestFollowUpIntent:
    def test_creates_with_title_and_description(self) -> None:
        from src.domain.models.follow_up_intent import FollowUpIntent

        intent = FollowUpIntent(
            title="Implement SessionStore",
            description="Create in-memory store with TTL",
        )
        assert intent.title == "Implement SessionStore"
        assert intent.description == "Create in-memory store with TTL"

    def test_is_frozen(self) -> None:
        from src.domain.models.follow_up_intent import FollowUpIntent

        intent = FollowUpIntent(title="Task", description="Details")
        with pytest.raises(AttributeError):
            intent.title = "Changed"  # type: ignore[misc]

    def test_empty_title_raises(self) -> None:
        from src.domain.models.follow_up_intent import FollowUpIntent

        with pytest.raises(InvariantViolationError):
            FollowUpIntent(title="", description="Details")

    def test_whitespace_only_title_raises(self) -> None:
        from src.domain.models.follow_up_intent import FollowUpIntent

        with pytest.raises(InvariantViolationError):
            FollowUpIntent(title="   ", description="Details")

    def test_empty_description_allowed(self) -> None:
        from src.domain.models.follow_up_intent import FollowUpIntent

        intent = FollowUpIntent(title="Task", description="")
        assert intent.description == ""

    def test_equality_by_value(self) -> None:
        from src.domain.models.follow_up_intent import FollowUpIntent

        a = FollowUpIntent(title="Task A", description="Desc")
        b = FollowUpIntent(title="Task A", description="Desc")
        assert a == b

    def test_different_intents_not_equal(self) -> None:
        from src.domain.models.follow_up_intent import FollowUpIntent

        a = FollowUpIntent(title="Task A", description="Desc")
        b = FollowUpIntent(title="Task B", description="Desc")
        assert a != b


# ── FollowUpAuditResult ────────────────────────────────────────────


class TestFollowUpAuditResult:
    def test_creates_with_all_fields(self) -> None:
        from src.domain.models.follow_up_intent import FollowUpAuditResult, FollowUpIntent

        intent = FollowUpIntent(title="Task 1", description="")
        result = FollowUpAuditResult(
            spike_id="k7m.8",
            report_path="docs/research/20260223_gap_analysis_design.md",
            defined_intents=(intent,),
            matched_ticket_ids=(),
            orphaned_intents=(intent,),
        )
        assert result.spike_id == "k7m.8"
        assert result.orphaned_count == 1

    def test_is_frozen(self) -> None:
        from src.domain.models.follow_up_intent import FollowUpAuditResult

        result = FollowUpAuditResult(
            spike_id="k7m.8",
            report_path="report.md",
            defined_intents=(),
            matched_ticket_ids=(),
            orphaned_intents=(),
        )
        with pytest.raises(AttributeError):
            result.spike_id = "changed"  # type: ignore[misc]

    def test_orphaned_count_property(self) -> None:
        from src.domain.models.follow_up_intent import FollowUpAuditResult, FollowUpIntent

        intents = tuple(FollowUpIntent(title=f"Task {i}", description="") for i in range(5))
        result = FollowUpAuditResult(
            spike_id="k7m.8",
            report_path="report.md",
            defined_intents=intents,
            matched_ticket_ids=("alty-abc", "alty-def"),
            orphaned_intents=intents[:3],
        )
        assert result.orphaned_count == 3

    def test_has_orphans_true(self) -> None:
        from src.domain.models.follow_up_intent import FollowUpAuditResult, FollowUpIntent

        intent = FollowUpIntent(title="Lost task", description="")
        result = FollowUpAuditResult(
            spike_id="k7m.8",
            report_path="report.md",
            defined_intents=(intent,),
            matched_ticket_ids=(),
            orphaned_intents=(intent,),
        )
        assert result.has_orphans is True

    def test_has_orphans_false_when_all_matched(self) -> None:
        from src.domain.models.follow_up_intent import FollowUpAuditResult, FollowUpIntent

        intent = FollowUpIntent(title="Created task", description="")
        result = FollowUpAuditResult(
            spike_id="k7m.8",
            report_path="report.md",
            defined_intents=(intent,),
            matched_ticket_ids=("alty-abc",),
            orphaned_intents=(),
        )
        assert result.has_orphans is False

    def test_no_intents_means_no_orphans(self) -> None:
        from src.domain.models.follow_up_intent import FollowUpAuditResult

        result = FollowUpAuditResult(
            spike_id="k7m.8",
            report_path="report.md",
            defined_intents=(),
            matched_ticket_ids=(),
            orphaned_intents=(),
        )
        assert result.orphaned_count == 0
        assert result.has_orphans is False

    def test_defined_count_property(self) -> None:
        from src.domain.models.follow_up_intent import FollowUpAuditResult, FollowUpIntent

        intents = tuple(FollowUpIntent(title=f"T{i}", description="") for i in range(17))
        result = FollowUpAuditResult(
            spike_id="k7m.8",
            report_path="report.md",
            defined_intents=intents,
            matched_ticket_ids=(),
            orphaned_intents=intents,
        )
        assert result.defined_count == 17
