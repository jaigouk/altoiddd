"""BeadsTicketReader -- anticorruption layer for the Ticket Freshness context.

Translates beads data into domain-layer value objects (OpenTicketData,
FreshnessFlag).  Reads ticket data from JSONL export and enriches with
labels and comments from the ``bd`` CLI (since JSONL lacks both).
Handles missing directories, missing CLI, and corrupted data gracefully.
"""

from __future__ import annotations

import json
import re
import subprocess
from typing import TYPE_CHECKING

from src.domain.models.ticket_freshness import (
    ContextDiff,
    FreshnessFlag,
    OpenTicketData,
)

if TYPE_CHECKING:
    from pathlib import Path

# Matches old-format ripple comments: **Ripple context diff from `<id>`:**
_RIPPLE_PATTERN_OLD = re.compile(
    r"\*\*Ripple context diff from `([^`]+)`:\*\*\s*(.*)",
    re.DOTALL,
)

# Matches current ripple comments: **Ripple review needed** -- `<id>` (title) was closed.
_RIPPLE_TRIGGER_RE = re.compile(
    r"\*\*Ripple review needed\*\*\s*--\s*`([^`]+)`",
)

# Extracts the "What changed:" summary text
_WHAT_CHANGED_RE = re.compile(
    r"\*\*What changed:\*\*\s*(.*?)(?:\n\n|\Z)",
    re.DOTALL,
)

# Parses bd query output lines: ○ alty-2j7.8 [● P2] [task] [review_needed] - title
_BD_QUERY_LINE_RE = re.compile(r"^[○◐●✓❄]\s+(\S+)")

_BD_TIMEOUT_SECONDS = 10


class BeadsTicketReader:
    """Reads beads data and translates it into domain value objects.

    This is an Anti-Corruption Layer (ACL) that shields the domain from
    beads data format details.  JSONL provides basic ticket data; the
    ``bd`` CLI provides labels and comments (which are not in the export).

    Attributes:
        _beads_dir: Path to the .beads directory.
    """

    def __init__(self, beads_dir: Path) -> None:
        self._beads_dir = beads_dir

    def read_open_tickets(self) -> tuple[OpenTicketData, ...]:
        """Read all open tickets from issues.jsonl, enriched with labels.

        Labels are fetched via ``bd query label=review_needed`` since the
        JSONL export does not include label data.

        Returns:
            Tuple of OpenTicketData for each open ticket.  Returns empty
            tuple if the directory or file is missing.
        """
        tickets = self._read_tickets_from_jsonl()
        flagged_ids = self._get_flagged_ids()
        return self._enrich_labels(tickets, flagged_ids)

    def read_flags(self, ticket_id: str) -> tuple[FreshnessFlag, ...]:
        """Read freshness flags from ripple review comments.

        Tries ``bd comments <id>`` first (primary source).  Falls back
        to interactions.jsonl if the CLI is unavailable.

        Args:
            ticket_id: The ticket to read flags for.

        Returns:
            Tuple of FreshnessFlag extracted from ripple comments.
        """
        flags = self._read_flags_from_bd_comments(ticket_id)
        if flags:
            return flags
        return self._read_flags_from_jsonl(ticket_id)

    # ------------------------------------------------------------------
    # JSONL reading (basic ticket data)
    # ------------------------------------------------------------------

    def _read_tickets_from_jsonl(self) -> list[OpenTicketData]:
        """Read open tickets from issues.jsonl (no labels)."""
        jsonl_path = self._beads_dir / "issues.jsonl"
        if not jsonl_path.exists():
            return []

        tickets: list[OpenTicketData] = []
        with jsonl_path.open() as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                try:
                    data = json.loads(line)
                except json.JSONDecodeError:
                    continue

                if data.get("status") != "open":
                    continue

                tickets.append(
                    OpenTicketData(
                        ticket_id=data.get("id", ""),
                        title=data.get("title", ""),
                        labels=(),
                        last_reviewed=None,
                    )
                )

        return tickets

    def _read_flags_from_jsonl(self, ticket_id: str) -> tuple[FreshnessFlag, ...]:
        """Fallback: read flags from interactions.jsonl."""
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
                match = _RIPPLE_PATTERN_OLD.search(body)
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

    # ------------------------------------------------------------------
    # bd CLI integration (labels + comments)
    # ------------------------------------------------------------------

    def _get_flagged_ids(self) -> set[str]:
        """Get IDs of tickets with review_needed label via bd CLI."""
        try:
            result = subprocess.run(
                ["bd", "query", "label=review_needed"],  # noqa: S607
                capture_output=True,
                text=True,
                timeout=_BD_TIMEOUT_SECONDS,
            )
            if result.returncode != 0:
                return set()

            ids: set[str] = set()
            for line in result.stdout.splitlines():
                match = _BD_QUERY_LINE_RE.match(line.strip())
                if match:
                    ids.add(match.group(1))
            return ids
        except (FileNotFoundError, subprocess.TimeoutExpired):
            return set()

    @staticmethod
    def _enrich_labels(
        tickets: list[OpenTicketData],
        flagged_ids: set[str],
    ) -> tuple[OpenTicketData, ...]:
        """Replace empty labels with review_needed for flagged tickets."""
        enriched: list[OpenTicketData] = []
        for t in tickets:
            labels = ("review_needed",) if t.ticket_id in flagged_ids else ()
            enriched.append(
                OpenTicketData(
                    ticket_id=t.ticket_id,
                    title=t.title,
                    labels=labels,
                    last_reviewed=t.last_reviewed,
                )
            )
        return tuple(enriched)

    def _read_flags_from_bd_comments(self, ticket_id: str) -> tuple[FreshnessFlag, ...]:
        """Parse ripple review comments from bd comments output."""
        try:
            result = subprocess.run(  # noqa: S603
                ["bd", "comments", ticket_id],  # noqa: S607
                capture_output=True,
                text=True,
                timeout=_BD_TIMEOUT_SECONDS,
            )
            if result.returncode != 0:
                return ()
        except (FileNotFoundError, subprocess.TimeoutExpired):
            return ()

        return self._parse_bd_comments(result.stdout)

    @staticmethod
    def _parse_bd_comments(output: str) -> tuple[FreshnessFlag, ...]:
        """Extract ripple review flags from bd comments text output.

        Handles both current format (``**Ripple review needed** -- ...``)
        and old format (``**Ripple context diff from ...**``).
        """
        flags: list[FreshnessFlag] = []

        # Split into individual comments by the [Author] at <date> pattern
        comment_blocks = re.split(r"\n\[.+?\] at (\d{4}-\d{2}-\d{2}[\sT]?\d{0,8})", output)

        # comment_blocks[0] is header, then pairs of (date, body)
        i = 1
        while i < len(comment_blocks) - 1:
            date_str = comment_blocks[i].strip()
            body = comment_blocks[i + 1]
            i += 2

            # Try current format first
            trigger_match = _RIPPLE_TRIGGER_RE.search(body)
            changed_match = _WHAT_CHANGED_RE.search(body)

            if trigger_match and changed_match:
                triggering_id = trigger_match.group(1)
                summary = changed_match.group(1).strip()
                if summary:
                    flags.append(
                        FreshnessFlag(
                            context_diff=ContextDiff(
                                summary=summary,
                                triggering_ticket_id=triggering_id,
                                produced_at=date_str,
                            ),
                            flagged_at=date_str,
                        )
                    )
                continue

            # Try old format
            old_match = _RIPPLE_PATTERN_OLD.search(body)
            if old_match:
                triggering_id = old_match.group(1)
                summary = old_match.group(2).strip()
                if summary:
                    flags.append(
                        FreshnessFlag(
                            context_diff=ContextDiff(
                                summary=summary,
                                triggering_ticket_id=triggering_id,
                                produced_at=date_str,
                            ),
                            flagged_at=date_str,
                        )
                    )

        return tuple(flags)
