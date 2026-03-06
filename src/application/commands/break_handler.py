"""BreakHandler — manages DomainModel version snapshots and iteration diffs.

Captures snapshots at iteration breaks, computes diffs between consecutive
versions, and classifies the convergence trend.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.domain.models.artifact_diff import ArtifactDiff, ArtifactVersion
from src.domain.services.diff_service import DiffService

if TYPE_CHECKING:
    from src.domain.models.domain_model import DomainModel


class BreakHandler:
    """Manages version snapshots and computes iteration diffs."""

    def __init__(self) -> None:
        self._versions: list[ArtifactVersion] = []
        self._diffs: list[ArtifactDiff] = []

    def capture_snapshot(self, version_number: int, model: DomainModel) -> ArtifactVersion:
        """Capture a versioned snapshot of a DomainModel.

        Args:
            version_number: Positive integer version (1, 2, 3, ...).
            model: The DomainModel to snapshot.

        Returns:
            The created ArtifactVersion.
        """
        version = ArtifactVersion(version_number=version_number, model=model)
        self._versions.append(version)
        return version

    def compute_diff(self) -> ArtifactDiff | None:
        """Compute diff between the last two snapshots.

        Returns:
            ArtifactDiff if >= 2 snapshots exist, None otherwise.
        """
        if len(self._versions) < 2:
            return None

        before = self._versions[-2]
        after = self._versions[-1]
        result = DiffService.compute(
            before.model,
            after.model,
            from_version=before.version_number,
            to_version=after.version_number,
        )
        self._diffs.append(result)
        return result

    def convergence_trend(self) -> str:
        """Classify the convergence trend based on accumulated diffs.

        Returns:
            'active refinement' if no diffs or increasing changes,
            'stabilizing' if changes are decreasing,
            'converged' if most recent diff has zero changes.
        """
        if not self._diffs:
            return "active refinement"

        latest = self._diffs[-1]
        total_changes = (
            latest.convergence.invariants_delta
            + latest.convergence.terms_delta
            + latest.convergence.stories_delta
            + latest.convergence.canvases_delta
        )

        if total_changes == 0:
            return "converged"

        if len(self._diffs) >= 2:
            previous = self._diffs[-2]
            prev_total = (
                previous.convergence.invariants_delta
                + previous.convergence.terms_delta
                + previous.convergence.stories_delta
                + previous.convergence.canvases_delta
            )
            if total_changes < prev_total:
                return "stabilizing"

        return "active refinement"
