"""Tests for Rescue domain events.

Verifies GapAnalysisCompleted is frozen and has the expected fields.
"""

from __future__ import annotations

from pathlib import Path

import pytest

from src.domain.events.rescue_events import GapAnalysisCompleted


class TestGapAnalysisCompleted:
    def test_gap_analysis_completed_is_frozen(self) -> None:
        event = GapAnalysisCompleted(
            analysis_id="abc-123",
            project_dir=Path("/tmp/proj"),
            gaps_found=5,
            gaps_resolved=3,
        )
        with pytest.raises(AttributeError):
            event.gaps_found = 10  # type: ignore[misc]

    def test_gap_analysis_completed_fields(self) -> None:
        event = GapAnalysisCompleted(
            analysis_id="abc-123",
            project_dir=Path("/tmp/proj"),
            gaps_found=5,
            gaps_resolved=3,
        )
        assert event.analysis_id == "abc-123"
        assert event.project_dir == Path("/tmp/proj")
        assert event.gaps_found == 5
        assert event.gaps_resolved == 3

    def test_gap_analysis_completed_equality(self) -> None:
        """Frozen dataclasses support structural equality."""
        e1 = GapAnalysisCompleted(
            analysis_id="abc",
            project_dir=Path("/tmp/proj"),
            gaps_found=1,
            gaps_resolved=1,
        )
        e2 = GapAnalysisCompleted(
            analysis_id="abc",
            project_dir=Path("/tmp/proj"),
            gaps_found=1,
            gaps_resolved=1,
        )
        assert e1 == e2
