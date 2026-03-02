"""SpikeFollowUpAdapter -- audits spike follow-ups against created tickets.

Implements SpikeFollowUpPort by scanning research reports for follow-up
intents and comparing them against beads tickets using fuzzy title matching.
"""

from __future__ import annotations

import json
from typing import TYPE_CHECKING

from src.domain.models.follow_up_intent import FollowUpAuditResult, FollowUpIntent
from src.infrastructure.persistence.markdown_spike_parser import MarkdownSpikeReportParser

if TYPE_CHECKING:
    from pathlib import Path


class SpikeFollowUpAdapter:
    """Implements SpikeFollowUpPort backed by filesystem scanning.

    Locates spike research reports in docs/research/, parses them for
    follow-up intents, and compares against .beads/issues.jsonl.
    """

    def __init__(self) -> None:
        self._parser = MarkdownSpikeReportParser()

    def audit(self, spike_id: str, project_dir: Path) -> FollowUpAuditResult:
        """Audit a spike's follow-up intents against created tickets."""
        research_dir = project_dir / "docs" / "research"
        if not research_dir.exists():
            return FollowUpAuditResult(
                spike_id=spike_id,
                report_path="",
                defined_intents=(),
                matched_ticket_ids=(),
                orphaned_intents=(),
            )

        # Scan all Markdown reports in docs/research/
        all_intents: list[FollowUpIntent] = []
        report_path = ""
        for report_file in sorted(research_dir.glob("*.md")):
            intents = self._parser.parse(report_file)
            if intents:
                all_intents.extend(intents)
                if not report_path:
                    report_path = str(report_file)

        if not all_intents:
            return FollowUpAuditResult(
                spike_id=spike_id,
                report_path=report_path,
                defined_intents=(),
                matched_ticket_ids=(),
                orphaned_intents=(),
            )

        # Load existing beads tickets
        existing_titles = self._load_ticket_titles(project_dir)

        # Match intents against tickets
        matched_ids: list[str] = []
        orphaned: list[FollowUpIntent] = []

        for intent in all_intents:
            ticket_id = self._fuzzy_match(intent.title, existing_titles)
            if ticket_id:
                matched_ids.append(ticket_id)
            else:
                orphaned.append(intent)

        return FollowUpAuditResult(
            spike_id=spike_id,
            report_path=report_path,
            defined_intents=tuple(all_intents),
            matched_ticket_ids=tuple(matched_ids),
            orphaned_intents=tuple(orphaned),
        )

    def _load_ticket_titles(self, project_dir: Path) -> dict[str, str]:
        """Load ticket ID → title mapping from .beads/issues.jsonl."""
        issues_path = project_dir / ".beads" / "issues.jsonl"
        if not issues_path.exists():
            return {}

        titles: dict[str, str] = {}
        for line in issues_path.read_text().splitlines():
            line = line.strip()
            if not line:
                continue
            try:
                issue = json.loads(line)
                ticket_id = issue.get("id", "")
                title = issue.get("title", "")
                if ticket_id and title:
                    titles[ticket_id] = title
            except json.JSONDecodeError:
                continue
        return titles

    def _fuzzy_match(
        self, intent_title: str, existing: dict[str, str]
    ) -> str | None:
        """Find a beads ticket that fuzzy-matches the intent title.

        Matching strategy (ordered by strictness):
        1. Case-insensitive exact match
        2. Case-insensitive substring (intent title contained in ticket title)
        3. Case-insensitive substring (ticket title contained in intent title)

        Returns the matching ticket ID, or None.
        """
        normalized = intent_title.lower().strip()

        for ticket_id, title in existing.items():
            ticket_lower = title.lower().strip()

            # Exact match
            if normalized == ticket_lower:
                return ticket_id

            # Intent title is substring of ticket title
            if normalized in ticket_lower:
                return ticket_id

            # Ticket title is substring of intent title
            if ticket_lower in normalized:
                return ticket_id

        return None
