"""Port for spike follow-up auditing (Ticket Freshness bounded context).

Defines the interface for auditing whether spike-defined follow-up
intents have been created as beads tickets.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from pathlib import Path

    from src.domain.models.follow_up_intent import FollowUpAuditResult


@runtime_checkable
class SpikeFollowUpPort(Protocol):
    """Interface for spike follow-up auditing.

    Adapters implement this to compare follow-up intents from spike
    reports against actually created beads tickets.
    """

    def audit(self, spike_id: str, project_dir: Path) -> FollowUpAuditResult:
        """Audit a spike's follow-up intents against created tickets.

        Locates the spike's research report, extracts follow-up intents,
        and compares them against existing beads tickets using fuzzy matching.

        Args:
            spike_id: The spike ticket identifier.
            project_dir: The project directory containing docs/research/.

        Returns:
            A FollowUpAuditResult with defined vs orphaned intents.
        """
        ...
