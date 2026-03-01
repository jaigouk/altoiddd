"""BeadsTicketReader -- anticorruption layer for the Ticket Freshness context.

Translates beads JSONL data (issues.jsonl and interactions.jsonl) into
domain-layer value objects (OpenTicketData, FreshnessFlag). Handles
missing directories and corrupted lines gracefully.
"""

from __future__ import annotations

import json
import re
from typing import TYPE_CHECKING

from src.domain.models.ticket_freshness import (
    ContextDiff,
    FreshnessFlag,
    OpenTicketData,
)

if TYPE_CHECKING:
    from pathlib import Path

# Matches ripple context diff comments: **Ripple context diff from `<id>`:**
_RIPPLE_PATTERN = re.compile(
    r"\*\*Ripple context diff from `([^`]+)`:\*\*\s*(.*)",
    re.DOTALL,
)


class BeadsTicketReader:
    """Reads beads JSONL files and translates them into domain value objects.

    This is an Anti-Corruption Layer (ACL) that shields the domain from
    beads data format details.

    Attributes:
        _beads_dir: Path to the .beads directory.
    """

    def __init__(self, beads_dir: Path) -> None:
        self._beads_dir = beads_dir

    def read_open_tickets(self) -> tuple[OpenTicketData, ...]:
        """Read all open tickets from issues.jsonl.

        Returns:
            Tuple of OpenTicketData for each open ticket. Returns empty
            tuple if the directory or file is missing.
        """
        jsonl_path = self._beads_dir / "issues.jsonl"
        if not jsonl_path.exists():
            return ()

        tickets: list[OpenTicketData] = []
        with jsonl_path.open() as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                try:
                    data = json.loads(line)
                except json.JSONDecodeError:
                    continue  # Skip corrupted lines

                if data.get("status") != "open":
                    continue

                tickets.append(
                    OpenTicketData(
                        ticket_id=data.get("id", ""),
                        title=data.get("title", ""),
                        labels=(),  # Labels live in Dolt, not JSONL export
                        last_reviewed=None,  # Derived from comments if needed
                    )
                )

        return tuple(tickets)

    def read_flags(self, ticket_id: str) -> tuple[FreshnessFlag, ...]:
        """Read freshness flags from ripple review comments in interactions.jsonl.

        Args:
            ticket_id: The ticket to read flags for.

        Returns:
            Tuple of FreshnessFlag extracted from ripple comments.
            Returns empty tuple if no interactions file or no matching comments.
        """
        interactions_path = self._beads_dir / "interactions.jsonl"
        if not interactions_path.exists():
            return ()

        flags: list[FreshnessFlag] = []
        with interactions_path.open() as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                try:
                    data = json.loads(line)
                except json.JSONDecodeError:
                    continue

                if data.get("issue_id") != ticket_id:
                    continue

                body = data.get("body", "")
                match = _RIPPLE_PATTERN.search(body)
                if not match:
                    continue

                triggering_id = match.group(1)
                summary = match.group(2).strip()
                if not summary:
                    continue

                created_at = data.get("created_at", "")

                flags.append(
                    FreshnessFlag(
                        context_diff=ContextDiff(
                            summary=summary,
                            triggering_ticket_id=triggering_id,
                            produced_at=created_at,
                        ),
                        flagged_at=created_at,
                    )
                )

        return tuple(flags)
