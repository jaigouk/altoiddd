"""MarkdownSpikeReportParser -- extracts FollowUpIntents from Markdown reports.

Implements SpikeReportParserPort for the Markdown format used by alty's
spike research reports (docs/research/YYYYMMDD_*.md).

Heading patterns detected (case-insensitive, with optional numbered prefix):
    ## Follow-Up Tickets
    ## N. Follow-Up Implementation Tickets
    ### Follow-up Tickets Needed

Ticket item formats detected:
    ### Ticket N: Title
    ### N. Title
    - **Title**: Description
    - Plain title text
"""

from __future__ import annotations

import re
from typing import TYPE_CHECKING

from src.domain.models.follow_up_intent import FollowUpIntent

if TYPE_CHECKING:
    from pathlib import Path

# ── Heading patterns ────────────────────────────────────────────────

# Matches ## or ### with optional "N." prefix, containing "follow-up" and "ticket"
_FOLLOWUP_HEADING_RE = re.compile(
    r"^(#{2,3})\s+(?:\d+\.\s+)?(?:Step\s+\d+:\s+)?(.+)$",
    re.IGNORECASE,
)

_FOLLOWUP_KEYWORDS = re.compile(r"follow.?up", re.IGNORECASE)
_TICKET_KEYWORDS = re.compile(r"ticket|implementation|work", re.IGNORECASE)

# ── Ticket item patterns ────────────────────────────────────────────

# ### Ticket N: Title  or  ### Ticket N: (Optional) Title
_TICKET_HEADING_RE = re.compile(
    r"^#{3,4}\s+(?:Ticket\s+\d+:\s*)(.+)$",
    re.IGNORECASE,
)

# ### N. Title
_NUMBERED_HEADING_RE = re.compile(
    r"^#{3,4}\s+\d+\.\s+(.+)$",
)

# - **Title**: Description  or  - **Title** — Description
_BOLD_LIST_RE = re.compile(
    r"^[-*]\s+\*\*(.+?)\*\*(?:\s*[:\u2014\u2013-]\s*(.*))?$",
)

# - Plain title text
_PLAIN_LIST_RE = re.compile(
    r"^[-*]\s+(.+)$",
)

# Any heading at level 2+ (used to detect section boundaries)
_ANY_HEADING_RE = re.compile(r"^(#{1,6})\s+")


def _is_followup_heading(text: str) -> bool:
    """Check if heading text indicates a follow-up tickets section."""
    return bool(_FOLLOWUP_KEYWORDS.search(text) and _TICKET_KEYWORDS.search(text))


class MarkdownSpikeReportParser:
    """Parses Markdown spike reports to extract follow-up intents."""

    def parse(self, report_path: Path) -> tuple[FollowUpIntent, ...]:
        """Extract follow-up intents from a Markdown spike report.

        Returns empty tuple if the file doesn't exist or has no follow-up section.
        """
        if not report_path.exists():
            return ()

        content = report_path.read_text()
        lines = content.splitlines()

        # Find the follow-up section
        section_start, section_level = self._find_followup_section(lines)
        if section_start is None:
            return ()

        # Extract the section content (until next same-level heading)
        section_lines = self._extract_section(lines, section_start, section_level)

        # Parse ticket items from the section
        return self._parse_items(section_lines)

    def _find_followup_section(
        self, lines: list[str]
    ) -> tuple[int | None, int]:
        """Find the first follow-up tickets heading.

        Returns (line_index, heading_level) or (None, 0).
        """
        for i, line in enumerate(lines):
            match = _FOLLOWUP_HEADING_RE.match(line)
            if match:
                level = len(match.group(1))
                heading_text = match.group(2)
                if _is_followup_heading(heading_text):
                    return i, level
        return None, 0

    def _extract_section(
        self, lines: list[str], start: int, level: int
    ) -> list[str]:
        """Extract lines from start+1 until the next heading at same or higher level."""
        result: list[str] = []
        for line in lines[start + 1 :]:
            heading_match = _ANY_HEADING_RE.match(line)
            if heading_match:
                current_level = len(heading_match.group(1))
                if current_level <= level:
                    break
            result.append(line)
        return result

    def _parse_items(self, lines: list[str]) -> tuple[FollowUpIntent, ...]:
        """Parse ticket items from section lines.

        Supports multiple formats: ### Ticket N: Title, ### N. Title,
        - **Title**: Description, - Plain text.
        """
        intents: list[FollowUpIntent] = []
        i = 0
        while i < len(lines):
            line = lines[i].strip()

            # Try ### Ticket N: Title
            match = _TICKET_HEADING_RE.match(lines[i])
            if match:
                title = match.group(1).strip()
                desc = self._collect_description(lines, i + 1)
                intents.append(FollowUpIntent(title=title, description=desc))
                i += 1
                continue

            # Try ### N. Title
            match = _NUMBERED_HEADING_RE.match(lines[i])
            if match:
                title = match.group(1).strip()
                desc = self._collect_description(lines, i + 1)
                intents.append(FollowUpIntent(title=title, description=desc))
                i += 1
                continue

            # Try - **Title**: Description
            match = _BOLD_LIST_RE.match(line)
            if match:
                title = match.group(1).strip()
                desc = (match.group(2) or "").strip()
                intents.append(FollowUpIntent(title=title, description=desc))
                i += 1
                continue

            # Try - Plain title (only if no other format matched)
            match = _PLAIN_LIST_RE.match(line)
            if match and not line.startswith("- http") and not line.startswith("- ["):
                title = match.group(1).strip()
                # Skip lines that look like metadata, not ticket titles
                if not title.startswith("**") and len(title) > 3:
                    intents.append(FollowUpIntent(title=title, description=""))
                    i += 1
                    continue

            i += 1

        return tuple(intents)

    def _collect_description(self, lines: list[str], start: int) -> str:
        """Collect description text after a heading until the next heading or list item."""
        desc_lines: list[str] = []
        for line in lines[start:]:
            stripped = line.strip()
            # Stop at next heading or ticket item
            if stripped.startswith("#") or _BOLD_LIST_RE.match(stripped):
                break
            # Stop at metadata fields
            _meta = ("**Type:", "**Priority:", "**Depends on:", "**Steps:", "**Bounded Context:")
            if stripped.startswith(_meta):
                break
            # Collect non-empty content lines
            if stripped:
                desc_lines.append(stripped)
        return " ".join(desc_lines)
