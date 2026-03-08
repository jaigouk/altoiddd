"""Tests for CLI diff rendering functions.

RED phase: 3 tests covering added/modified/removed markers,
convergence trend display, and large diff truncation.
"""

from __future__ import annotations

from src.domain.models.artifact_diff import (
    ArtifactDiff,
    ConvergenceMetric,
    DiffEntry,
    DiffType,
)
from src.domain.models.research import TrustLevel
from src.infrastructure.cli.diff_rendering import format_diff


def _zero_convergence() -> ConvergenceMetric:
    return ConvergenceMetric(invariants_delta=0, terms_delta=0, stories_delta=0, canvases_delta=0)


class TestFormatDiff:
    """format_diff returns a formatted string for an ArtifactDiff."""

    def test_markers_for_added_modified_removed(self) -> None:
        diff = ArtifactDiff(
            from_version=1,
            to_version=2,
            entries=(
                DiffEntry(
                    diff_type=DiffType.ADDED,
                    section="Bounded Contexts",
                    description="Added context 'Billing'",
                    provenance=TrustLevel.AI_INFERRED,
                ),
                DiffEntry(
                    diff_type=DiffType.MODIFIED,
                    section="Ubiquitous Language",
                    description="Modified term 'Order'",
                    provenance=TrustLevel.AI_INFERRED,
                ),
                DiffEntry(
                    diff_type=DiffType.REMOVED,
                    section="Domain Stories",
                    description="Removed story 'Old Flow'",
                    provenance=TrustLevel.AI_INFERRED,
                ),
                DiffEntry(
                    diff_type=DiffType.DISAMBIGUATED,
                    section="Ubiquitous Language",
                    description="Disambiguated term 'Policy'",
                    provenance=TrustLevel.AI_INFERRED,
                ),
            ),
            convergence=_zero_convergence(),
        )

        output = format_diff(diff, trend="active refinement")
        assert "+ Added context 'Billing'" in output
        assert "~ Modified term 'Order'" in output
        assert "- Removed story 'Old Flow'" in output
        assert "~ Disambiguated term 'Policy'" in output

    def test_shows_convergence_trend(self) -> None:
        diff = ArtifactDiff(
            from_version=1,
            to_version=2,
            entries=(),
            convergence=_zero_convergence(),
        )

        output = format_diff(diff, trend="converged")
        assert "converged" in output.lower()

    def test_groups_by_section(self) -> None:
        entries = tuple(
            DiffEntry(
                diff_type=DiffType.ADDED,
                section="Bounded Contexts",
                description=f"Added context '{i}'",
                provenance=TrustLevel.AI_INFERRED,
            )
            for i in range(5)
        )

        diff = ArtifactDiff(
            from_version=1,
            to_version=2,
            entries=entries,
            convergence=ConvergenceMetric(
                invariants_delta=0, terms_delta=0, stories_delta=0, canvases_delta=5
            ),
        )

        output = format_diff(diff, trend="active refinement")
        # Should have a section header
        assert "Bounded Contexts" in output

    def test_empty_diff_includes_version_range(self) -> None:
        diff = ArtifactDiff(
            from_version=3,
            to_version=4,
            entries=(),
            convergence=_zero_convergence(),
        )
        output = format_diff(diff, trend="converged")
        assert "3" in output
        assert "4" in output

    def test_multiple_sections_all_appear(self) -> None:
        diff = ArtifactDiff(
            from_version=1,
            to_version=2,
            entries=(
                DiffEntry(
                    diff_type=DiffType.ADDED,
                    section="Bounded Contexts",
                    description="Added 'Sales'",
                    provenance=TrustLevel.AI_INFERRED,
                ),
                DiffEntry(
                    diff_type=DiffType.ADDED,
                    section="Domain Stories",
                    description="Added 'Checkout'",
                    provenance=TrustLevel.AI_INFERRED,
                ),
                DiffEntry(
                    diff_type=DiffType.ADDED,
                    section="Ubiquitous Language",
                    description="Added 'Order'",
                    provenance=TrustLevel.AI_INFERRED,
                ),
            ),
            convergence=ConvergenceMetric(
                invariants_delta=0, terms_delta=1, stories_delta=1, canvases_delta=1
            ),
        )
        output = format_diff(diff, trend="active refinement")
        assert "Bounded Contexts" in output
        assert "Domain Stories" in output
        assert "Ubiquitous Language" in output
