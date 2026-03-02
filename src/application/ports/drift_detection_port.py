"""Port for knowledge base drift detection.

Defines the interface for detecting drift in knowledge base entries,
separate from KnowledgeLookupPort (read vs detect — ISP).
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from src.domain.models.drift_detection import DriftReport


@runtime_checkable
class DriftDetectionPort(Protocol):
    """Interface for knowledge base drift detection operations.

    Adapters implement this to scan knowledge entries for staleness,
    version-to-version changes, and doc-vs-code mismatches.
    """

    def detect(self) -> DriftReport:
        """Detect drift across all knowledge entries.

        Returns:
            A DriftReport with all detected drift signals.
        """
        ...
