"""SpikeFollowUpAdapter -- audits spike follow-ups against created tickets.

Implements SpikeFollowUpPort by scanning research reports for follow-up
intents and comparing them against beads tickets using fuzzy title matching.
"""

from __future__ import annotations

import json
import re
from typing import TYPE_CHECKING

from src.domain.models.follow_up_intent import FollowUpAuditResult, FollowUpIntent
from src.infrastructure.persistence.markdown_spike_parser import MarkdownSpikeReportParser

if TYPE_CHECKING:
    from pathlib import Path

# Prefixes stripped before comparison (case-insensitive)
_STRIP_PREFIXES = (
    "task:",
    "spike:",
    "bug:",
    "feature:",
    "(optional)",
)

# Short stop-words excluded from keyword overlap scoring
_STOP_WORDS = frozenset(
    {
        "a",
        "an",
        "the",
        "and",
        "or",
        "of",
        "to",
        "in",
        "for",
        "on",
        "with",
        "is",
        "it",
        "be",
        "as",
        "at",
        "by",
        "from",
        "that",
        "this",
    }
)

# Minimum Jaccard similarity for keyword overlap match
_KEYWORD_OVERLAP_THRESHOLD = 0.4


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

    def _fuzzy_match(self, intent_title: str, existing: dict[str, str]) -> str | None:
        """Find a beads ticket that fuzzy-matches the intent title.

        Matching strategy (ordered by strictness):
        1. Case-insensitive exact match (after prefix stripping)
        2. Case-insensitive substring (intent in ticket or ticket in intent)
        3. Keyword overlap (Jaccard similarity >= threshold)

        Returns the matching ticket ID, or None.
        """
        stripped_intent = self._strip_prefixes(intent_title)
        if not stripped_intent:
            return None

        intent_lower = stripped_intent.lower()

        for ticket_id, title in existing.items():
            stripped_ticket = self._strip_prefixes(title)
            if not stripped_ticket:
                continue
            ticket_lower = stripped_ticket.lower()

            # Tier 1: Exact match
            if intent_lower == ticket_lower:
                return ticket_id

            # Tier 2: Substring
            if intent_lower in ticket_lower or ticket_lower in intent_lower:
                return ticket_id

            # Tier 3: Keyword overlap
            if self._keyword_overlap(intent_lower, ticket_lower) >= _KEYWORD_OVERLAP_THRESHOLD:
                return ticket_id

        return None

    @staticmethod
    def _strip_prefixes(title: str) -> str:
        """Remove common prefixes from a title for comparison."""
        result = title.strip()
        lower = result.lower()
        for prefix in _STRIP_PREFIXES:
            if lower.startswith(prefix):
                result = result[len(prefix) :].strip()
                lower = result.lower()
        return result

    @staticmethod
    def _tokenize(text: str) -> set[str]:
        """Tokenize a title into meaningful keywords.

        Splits on whitespace/punctuation, lowercases, removes stop-words,
        and extracts parenthetical content as additional tokens.
        """
        # Extract parenthetical content as extra tokens
        parens = re.findall(r"\(([^)]+)\)", text)
        paren_tokens: set[str] = set()
        for p in parens:
            for token in re.split(r"[\s,+/]+", p.lower()):
                token = token.strip()
                if token and token not in _STOP_WORDS and len(token) > 1:
                    paren_tokens.add(token)

        # Main tokenization
        words = re.split(r"[\s\-_/,()]+", text.lower())
        tokens = {w for w in words if w and w not in _STOP_WORDS and len(w) > 1}

        return tokens | paren_tokens

    @classmethod
    def _keyword_overlap(cls, a: str, b: str) -> float:
        """Compute fuzzy keyword overlap between two titles.

        Uses prefix matching (first 5 chars) for each token pair to handle
        morphological variants like "generate"/"generation", "function"/"functions".
        Returns a score from 0.0 to 1.0.
        """
        tokens_a = cls._tokenize(a)
        tokens_b = cls._tokenize(b)
        if not tokens_a or not tokens_b:
            return 0.0

        # Count how many tokens in A have a fuzzy match in B
        matched_a = 0
        matched_b_tokens: set[str] = set()
        for ta in tokens_a:
            for tb in tokens_b:
                if tb in matched_b_tokens:
                    continue
                if ta == tb or (len(ta) >= 5 and len(tb) >= 5 and ta[:5] == tb[:5]):
                    matched_a += 1
                    matched_b_tokens.add(tb)
                    break

        total_unique = len(tokens_a) + len(tokens_b) - matched_a
        if total_unique == 0:
            return 0.0
        return matched_a / total_unique
