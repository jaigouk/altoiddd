"""Tests for artifact diff value objects.

RED phase: These tests define the contracts for DiffType, DiffEntry,
ConvergenceMetric, ArtifactVersion, and ArtifactDiff.
"""

from __future__ import annotations

import pytest

from src.domain.models.artifact_diff import (
    ArtifactDiff,
    ArtifactVersion,
    ConvergenceMetric,
    DiffEntry,
    DiffType,
)
from src.domain.models.domain_model import DomainModel
from src.domain.models.errors import InvariantViolationError
from src.domain.models.research import TrustLevel


class TestDiffType:
    """DiffType enum has exactly four members."""

    def test_has_added(self) -> None:
        assert DiffType.ADDED.value == "added"

    def test_has_modified(self) -> None:
        assert DiffType.MODIFIED.value == "modified"

    def test_has_removed(self) -> None:
        assert DiffType.REMOVED.value == "removed"

    def test_has_disambiguated(self) -> None:
        assert DiffType.DISAMBIGUATED.value == "disambiguated"


class TestDiffEntry:
    """DiffEntry is a frozen VO with provenance."""

    def test_creation(self) -> None:
        entry = DiffEntry(
            diff_type=DiffType.ADDED,
            section="Ubiquitous Language",
            description="Added term 'Order'",
            provenance=TrustLevel.AI_INFERRED,
        )
        assert entry.diff_type == DiffType.ADDED
        assert entry.section == "Ubiquitous Language"
        assert entry.description == "Added term 'Order'"
        assert entry.provenance == TrustLevel.AI_INFERRED

    def test_frozen(self) -> None:
        entry = DiffEntry(
            diff_type=DiffType.ADDED,
            section="Bounded Contexts",
            description="Added context 'Billing'",
            provenance=TrustLevel.AI_INFERRED,
        )
        with pytest.raises(AttributeError):
            entry.section = "Other"  # type: ignore[misc]

    def test_empty_description_rejected(self) -> None:
        with pytest.raises(InvariantViolationError):
            DiffEntry(
                diff_type=DiffType.ADDED,
                section="Bounded Contexts",
                description="",
                provenance=TrustLevel.AI_INFERRED,
            )

    def test_empty_section_rejected(self) -> None:
        with pytest.raises(InvariantViolationError):
            DiffEntry(
                diff_type=DiffType.ADDED,
                section="",
                description="Added something",
                provenance=TrustLevel.AI_INFERRED,
            )


class TestConvergenceMetric:
    """ConvergenceMetric is a frozen VO with non-negative deltas."""

    def test_creation(self) -> None:
        metric = ConvergenceMetric(
            invariants_delta=2,
            terms_delta=3,
            stories_delta=1,
            canvases_delta=0,
        )
        assert metric.invariants_delta == 2
        assert metric.terms_delta == 3
        assert metric.stories_delta == 1
        assert metric.canvases_delta == 0

    def test_frozen(self) -> None:
        metric = ConvergenceMetric(
            invariants_delta=0, terms_delta=0, stories_delta=0, canvases_delta=0
        )
        with pytest.raises(AttributeError):
            metric.terms_delta = 5  # type: ignore[misc]

    def test_negative_delta_rejected(self) -> None:
        with pytest.raises(InvariantViolationError):
            ConvergenceMetric(
                invariants_delta=-1, terms_delta=0, stories_delta=0, canvases_delta=0
            )


class TestArtifactVersion:
    """ArtifactVersion holds a version number and a DomainModel."""

    def test_creation(self) -> None:
        model = DomainModel()
        version = ArtifactVersion(version_number=1, model=model)
        assert version.version_number == 1
        assert version.model is model

    def test_version_zero_rejected(self) -> None:
        with pytest.raises(InvariantViolationError):
            ArtifactVersion(version_number=0, model=DomainModel())

    def test_immutable(self) -> None:
        model = DomainModel()
        version = ArtifactVersion(version_number=1, model=model)
        with pytest.raises(AttributeError, match="immutable"):
            version.version_number = 2  # type: ignore[misc]

    def test_negative_version_rejected(self) -> None:
        with pytest.raises(InvariantViolationError):
            ArtifactVersion(version_number=-1, model=DomainModel())


class TestArtifactDiff:
    """ArtifactDiff is a frozen VO with version ordering invariant."""

    def test_creation(self) -> None:
        diff = ArtifactDiff(
            from_version=1,
            to_version=2,
            entries=(),
            convergence=ConvergenceMetric(
                invariants_delta=0, terms_delta=0, stories_delta=0, canvases_delta=0
            ),
        )
        assert diff.from_version == 1
        assert diff.to_version == 2
        assert diff.entries == ()

    def test_from_must_be_less_than_to(self) -> None:
        with pytest.raises(InvariantViolationError):
            ArtifactDiff(
                from_version=2,
                to_version=1,
                entries=(),
                convergence=ConvergenceMetric(
                    invariants_delta=0,
                    terms_delta=0,
                    stories_delta=0,
                    canvases_delta=0,
                ),
            )

    def test_equal_versions_rejected(self) -> None:
        with pytest.raises(InvariantViolationError):
            ArtifactDiff(
                from_version=1,
                to_version=1,
                entries=(),
                convergence=ConvergenceMetric(
                    invariants_delta=0,
                    terms_delta=0,
                    stories_delta=0,
                    canvases_delta=0,
                ),
            )

    def test_frozen(self) -> None:
        diff = ArtifactDiff(
            from_version=1,
            to_version=2,
            entries=(),
            convergence=ConvergenceMetric(
                invariants_delta=0, terms_delta=0, stories_delta=0, canvases_delta=0
            ),
        )
        with pytest.raises(AttributeError):
            diff.from_version = 3  # type: ignore[misc]
