"""Domain events for the Rescue bounded context.

GapAnalysisCompleted is emitted when a gap analysis finishes executing
all planned migration actions for an existing project.
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from pathlib import Path


@dataclass(frozen=True)
class GapAnalysisCompleted:
    """Emitted when a rescue gap analysis completes execution.

    Attributes:
        analysis_id: Unique identifier of the completed analysis.
        project_dir: The directory that was analyzed and migrated.
        gaps_found: Total number of gaps identified during analysis.
        gaps_resolved: Number of gaps resolved during execution.
    """

    analysis_id: str
    project_dir: Path
    gaps_found: int
    gaps_resolved: int
