"""Tests for BreakHandler — manages version snapshots and computes diffs.

RED phase: 3 tests covering snapshot capture, diff computation, and
convergence trend detection.
"""

from __future__ import annotations

from src.domain.models.domain_model import DomainModel
from src.domain.models.domain_values import BoundedContext


class TestBreakHandlerCaptureSnapshot:
    """capture_snapshot stores versioned DomainModel snapshots."""

    def test_capture_returns_artifact_version(self) -> None:
        from src.application.commands.break_handler import BreakHandler

        handler = BreakHandler()
        model = DomainModel()
        version = handler.capture_snapshot(1, model)
        assert version.version_number == 1
        assert version.model is model


class TestBreakHandlerComputeDiff:
    """compute_diff returns diff between last two snapshots."""

    def test_returns_none_with_fewer_than_two_snapshots(self) -> None:
        from src.application.commands.break_handler import BreakHandler

        handler = BreakHandler()
        assert handler.compute_diff() is None

        handler.capture_snapshot(1, DomainModel())
        assert handler.compute_diff() is None

    def test_returns_diff_between_last_two(self) -> None:
        from src.application.commands.break_handler import BreakHandler

        handler = BreakHandler()
        model1 = DomainModel()
        model2 = DomainModel()
        model2.add_bounded_context(BoundedContext(name="Billing", responsibility="Payments"))

        handler.capture_snapshot(1, model1)
        handler.capture_snapshot(2, model2)

        diff = handler.compute_diff()
        assert diff is not None
        assert diff.from_version == 1
        assert diff.to_version == 2
        assert len(diff.entries) > 0


class TestBreakHandlerConvergenceTrend:
    """convergence_trend classifies the change pattern."""

    def test_no_snapshots_returns_active_refinement(self) -> None:
        from src.application.commands.break_handler import BreakHandler

        handler = BreakHandler()
        assert handler.convergence_trend() == "active refinement"

    def test_stabilizing_when_changes_decrease(self) -> None:
        from src.application.commands.break_handler import BreakHandler

        handler = BreakHandler()

        # First iteration: many changes
        model1 = DomainModel()
        model2 = DomainModel()
        model2.add_bounded_context(BoundedContext(name="A", responsibility="A"))
        model2.add_bounded_context(BoundedContext(name="B", responsibility="B"))
        model2.add_bounded_context(BoundedContext(name="C", responsibility="C"))
        model2.add_term("X", "def", "A")
        model2.add_term("Y", "def", "B")
        handler.capture_snapshot(1, model1)
        handler.capture_snapshot(2, model2)
        handler.compute_diff()  # record first diff

        # Second iteration: fewer changes
        model3 = DomainModel()
        model3.add_bounded_context(BoundedContext(name="A", responsibility="A"))
        model3.add_bounded_context(BoundedContext(name="B", responsibility="B"))
        model3.add_bounded_context(BoundedContext(name="C", responsibility="C"))
        model3.add_term("X", "def", "A")
        model3.add_term("Y", "def", "B")
        model3.add_term("Z", "def", "C")  # only one new term
        handler.capture_snapshot(3, model3)
        handler.compute_diff()  # record second diff

        assert handler.convergence_trend() == "stabilizing"

    def test_converged_when_no_changes(self) -> None:
        from src.application.commands.break_handler import BreakHandler

        handler = BreakHandler()

        model1 = DomainModel()
        model1.add_bounded_context(BoundedContext(name="A", responsibility="A"))
        model2 = DomainModel()
        model2.add_bounded_context(BoundedContext(name="A", responsibility="A"))

        handler.capture_snapshot(1, model1)
        handler.capture_snapshot(2, model2)
        handler.compute_diff()

        assert handler.convergence_trend() == "converged"

    def test_active_refinement_without_diffs(self) -> None:
        from src.application.commands.break_handler import BreakHandler

        handler = BreakHandler()
        model1 = DomainModel()
        model2 = DomainModel()
        model2.add_bounded_context(BoundedContext(name="A", responsibility="A"))

        handler.capture_snapshot(1, model1)
        handler.capture_snapshot(2, model2)
        # Don't call compute_diff — query should NOT compute as side effect
        assert handler.convergence_trend() == "active refinement"


class TestBreakHandlerMultipleVersions:
    """BreakHandler with 3+ versions diffs the last two."""

    def test_compute_diff_uses_last_two_versions(self) -> None:
        from src.application.commands.break_handler import BreakHandler

        handler = BreakHandler()
        model1 = DomainModel()
        model2 = DomainModel()
        model2.add_bounded_context(BoundedContext(name="A", responsibility="A"))
        model3 = DomainModel()
        model3.add_bounded_context(BoundedContext(name="A", responsibility="A"))
        model3.add_bounded_context(BoundedContext(name="B", responsibility="B"))

        handler.capture_snapshot(1, model1)
        handler.capture_snapshot(2, model2)
        handler.capture_snapshot(3, model3)

        diff = handler.compute_diff()
        assert diff is not None
        assert diff.from_version == 2
        assert diff.to_version == 3
        # Only "B" is new between model2 and model3
        assert len(diff.entries) == 1
