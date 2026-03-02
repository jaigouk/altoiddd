"""Port for parsing spike research reports into follow-up intents.

Defines the interface for extracting FollowUpIntent value objects
from any report format. Infrastructure adapters implement this
for specific formats (Markdown, YAML, etc.).
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from pathlib import Path

    from src.domain.models.follow_up_intent import FollowUpIntent


@runtime_checkable
class SpikeReportParserPort(Protocol):
    """Interface for parsing follow-up intents from spike reports.

    Adapters implement this for each supported report format.
    The domain layer never depends on the format directly.
    """

    def parse(self, report_path: Path) -> tuple[FollowUpIntent, ...]:
        """Extract follow-up intents from a spike research report.

        Args:
            report_path: Path to the spike report file.

        Returns:
            Tuple of FollowUpIntent value objects found in the report.
            Returns empty tuple if no follow-up section is found.
        """
        ...
