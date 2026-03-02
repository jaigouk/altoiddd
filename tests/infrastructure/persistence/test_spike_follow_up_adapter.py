"""Tests for SpikeFollowUpAdapter.

RED phase: these tests define the contract for the audit adapter that
compares follow-up intents from spike reports against created beads tickets.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from pathlib import Path

# ── Helpers ──────────────────────────────────────────────────────────


def _write_spike_report(project_dir: Path, filename: str, content: str) -> Path:
    """Write a spike report to docs/research/ under a project dir."""
    research_dir = project_dir / "docs" / "research"
    research_dir.mkdir(parents=True, exist_ok=True)
    report = research_dir / filename
    report.write_text(content)
    return report


def _write_beads_issues(project_dir: Path, titles: list[str]) -> None:
    """Write a minimal .beads/issues.jsonl with given ticket titles."""
    import json

    beads_dir = project_dir / ".beads"
    beads_dir.mkdir(parents=True, exist_ok=True)
    issues_file = beads_dir / "issues.jsonl"
    lines = []
    for i, title in enumerate(titles):
        issue = {"id": f"alty-t{i}", "title": title, "status": "open"}
        lines.append(json.dumps(issue))
    issues_file.write_text("\n".join(lines) + "\n")


# ── Happy path ──────────────────────────────────────────────────────


class TestSpikeFollowUpAdapterAudit:
    def test_detects_orphaned_follow_ups(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.spike_follow_up_adapter import SpikeFollowUpAdapter

        _write_spike_report(
            tmp_path,
            "20260223_test_spike.md",
            (
                "# Test Spike\n\n"
                "## Follow-Up Tickets\n\n"
                "### Ticket 1: Build the widget\n\nDetails.\n\n"
                "### Ticket 2: Test the widget\n\nDetails.\n"
            ),
        )
        _write_beads_issues(tmp_path, [])  # No tickets created

        adapter = SpikeFollowUpAdapter()
        result = adapter.audit("test-spike", tmp_path)
        assert result.defined_count == 2
        assert result.orphaned_count == 2
        assert result.has_orphans is True

    def test_all_matched_no_orphans(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.spike_follow_up_adapter import SpikeFollowUpAdapter

        _write_spike_report(
            tmp_path,
            "20260223_test_spike.md",
            (
                "# Spike\n\n## Follow-Up Tickets\n\n"
                "### Ticket 1: Build the widget\n\nDetails.\n\n"
                "### Ticket 2: Test the widget\n\nDetails.\n"
            ),
        )
        _write_beads_issues(tmp_path, ["Build the widget", "Test the widget"])

        adapter = SpikeFollowUpAdapter()
        result = adapter.audit("test-spike", tmp_path)
        assert result.defined_count == 2
        assert result.orphaned_count == 0
        assert result.has_orphans is False
        assert len(result.matched_ticket_ids) == 2

    def test_partial_match_reports_orphans(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.spike_follow_up_adapter import SpikeFollowUpAdapter

        _write_spike_report(
            tmp_path,
            "20260223_test_spike.md",
            (
                "# Spike\n\n## Follow-Up Tickets\n\n"
                "### Ticket 1: Build parser\n\nD.\n\n"
                "### Ticket 2: Build adapter\n\nD.\n\n"
                "### Ticket 3: Wire composition\n\nD.\n"
            ),
        )
        _write_beads_issues(tmp_path, ["Build parser"])

        adapter = SpikeFollowUpAdapter()
        result = adapter.audit("test-spike", tmp_path)
        assert result.defined_count == 3
        assert result.orphaned_count == 2
        assert len(result.matched_ticket_ids) == 1

    def test_fuzzy_match_case_insensitive(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.spike_follow_up_adapter import SpikeFollowUpAdapter

        _write_spike_report(
            tmp_path,
            "20260223_test_spike.md",
            "# Spike\n\n## Follow-Up Tickets\n\n### Ticket 1: Build The Widget\n\nD.\n",
        )
        _write_beads_issues(tmp_path, ["build the widget"])

        adapter = SpikeFollowUpAdapter()
        result = adapter.audit("test-spike", tmp_path)
        assert result.orphaned_count == 0

    def test_fuzzy_match_substring(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.spike_follow_up_adapter import SpikeFollowUpAdapter

        _write_spike_report(
            tmp_path,
            "20260223_test_spike.md",
            "# Spike\n\n## Follow-Up Tickets\n\n### Ticket 1: Implement SessionStore\n\nD.\n",
        )
        _write_beads_issues(
            tmp_path,
            ["Implement SessionStore and MCP discovery tools"],
        )

        adapter = SpikeFollowUpAdapter()
        result = adapter.audit("test-spike", tmp_path)
        assert result.orphaned_count == 0

    def test_report_path_in_result(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.spike_follow_up_adapter import SpikeFollowUpAdapter

        _write_spike_report(
            tmp_path,
            "20260223_gap_analysis.md",
            "# Spike\n\n## Follow-Up Tickets\n\n### Ticket 1: Task\n\nD.\n",
        )
        _write_beads_issues(tmp_path, [])

        adapter = SpikeFollowUpAdapter()
        result = adapter.audit("gap-spike", tmp_path)
        assert "20260223_gap_analysis.md" in result.report_path


# ── Edge cases ──────────────────────────────────────────────────────


class TestSpikeFollowUpAdapterEdgeCases:
    def test_no_research_dir_returns_empty_result(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.spike_follow_up_adapter import SpikeFollowUpAdapter

        adapter = SpikeFollowUpAdapter()
        result = adapter.audit("missing-spike", tmp_path)
        assert result.defined_count == 0
        assert result.orphaned_count == 0

    def test_no_beads_dir_treats_all_as_orphaned(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.spike_follow_up_adapter import SpikeFollowUpAdapter

        _write_spike_report(
            tmp_path,
            "20260223_test.md",
            "# S\n\n## Follow-Up Tickets\n\n### Ticket 1: Task A\n\nD.\n",
        )
        # No .beads/ directory

        adapter = SpikeFollowUpAdapter()
        result = adapter.audit("test", tmp_path)
        assert result.orphaned_count == 1

    def test_spike_with_no_follow_ups_clean_result(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.spike_follow_up_adapter import SpikeFollowUpAdapter

        _write_spike_report(
            tmp_path,
            "20260223_no_followups.md",
            "# Research\n\n## Findings\n\nJust research, no tickets.\n",
        )
        _write_beads_issues(tmp_path, [])

        adapter = SpikeFollowUpAdapter()
        result = adapter.audit("no-followups", tmp_path)
        assert result.defined_count == 0
        assert result.has_orphans is False

    def test_implements_spike_follow_up_port(self) -> None:
        from src.application.ports.spike_follow_up_port import SpikeFollowUpPort
        from src.infrastructure.persistence.spike_follow_up_adapter import SpikeFollowUpAdapter

        adapter = SpikeFollowUpAdapter()
        assert isinstance(adapter, SpikeFollowUpPort)

    def test_multiple_reports_scans_all(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.spike_follow_up_adapter import SpikeFollowUpAdapter

        _write_spike_report(
            tmp_path,
            "20260223_report_a.md",
            "# A\n\n## Follow-Up Tickets\n\n### Ticket 1: Task from A\n\nD.\n",
        )
        _write_spike_report(
            tmp_path,
            "20260223_report_b.md",
            "# B\n\n## Follow-Up Tickets\n\n### Ticket 1: Task from B\n\nD.\n",
        )
        _write_beads_issues(tmp_path, [])

        adapter = SpikeFollowUpAdapter()
        result = adapter.audit("multi", tmp_path)
        # Should find intents from at least one report
        assert result.defined_count >= 1


# ── Improved fuzzy matching (alty-dj7) ─────────────────────────────


class TestFuzzyMatchPrefixStripping:
    """Prefix stripping: common prefixes removed before comparison."""

    def test_task_prefix_stripped(self) -> None:
        from src.infrastructure.persistence.spike_follow_up_adapter import SpikeFollowUpAdapter

        adapter = SpikeFollowUpAdapter()
        result = adapter._fuzzy_match("Task: Build the parser", {"t1": "Build the parser"})
        assert result == "t1"

    def test_spike_prefix_stripped(self) -> None:
        from src.infrastructure.persistence.spike_follow_up_adapter import SpikeFollowUpAdapter

        adapter = SpikeFollowUpAdapter()
        result = adapter._fuzzy_match(
            "Spike: Research caching strategies", {"t1": "Research caching strategies"}
        )
        assert result == "t1"

    def test_optional_prefix_stripped(self) -> None:
        from src.infrastructure.persistence.spike_follow_up_adapter import SpikeFollowUpAdapter

        adapter = SpikeFollowUpAdapter()
        result = adapter._fuzzy_match("(Optional) Add retry logic", {"t1": "Add retry logic"})
        assert result == "t1"

    def test_implement_alty_prefix_stripped(self) -> None:
        """'Implement X' matches 'Implement alty X'."""
        from src.infrastructure.persistence.spike_follow_up_adapter import SpikeFollowUpAdapter

        adapter = SpikeFollowUpAdapter()
        result = adapter._fuzzy_match(
            "Implement fitness function generation",
            {"t1": "Implement alty generate fitness"},
        )
        # Should match via keyword overlap even after prefix strip
        assert result == "t1"

    def test_empty_after_strip_no_match(self) -> None:
        """A title that's only a prefix should not match anything."""
        from src.infrastructure.persistence.spike_follow_up_adapter import SpikeFollowUpAdapter

        adapter = SpikeFollowUpAdapter()
        result = adapter._fuzzy_match("Task:", {"t1": "Implement alty detect"})
        assert result is None


class TestFuzzyMatchKeywordOverlap:
    """Keyword overlap scoring: Jaccard similarity on tokenized titles."""

    def test_keyword_overlap_above_threshold(self) -> None:
        """Titles with rephrased words match via keyword overlap."""
        from src.infrastructure.persistence.spike_follow_up_adapter import SpikeFollowUpAdapter

        adapter = SpikeFollowUpAdapter()
        # "fitness function generation (import-linter + pytestarch)"
        # vs "Implement alty generate fitness (import-linter + pytestarch)"
        # Shared keywords: fitness, import-linter, pytestarch
        result = adapter._fuzzy_match(
            "Implement fitness function generation (import-linter + pytestarch)",
            {"t1": "Implement alty generate fitness (import-linter + pytestarch)"},
        )
        assert result == "t1"

    def test_keyword_overlap_below_threshold_no_match(self) -> None:
        """Unrelated titles with few shared words do not match."""
        from src.infrastructure.persistence.spike_follow_up_adapter import SpikeFollowUpAdapter

        adapter = SpikeFollowUpAdapter()
        result = adapter._fuzzy_match(
            "Build the parser for YAML",
            {"t1": "Deploy the server to production"},
        )
        assert result is None

    def test_keyword_reorder_matches(self) -> None:
        """Same keywords in different order match."""
        from src.infrastructure.persistence.spike_follow_up_adapter import SpikeFollowUpAdapter

        adapter = SpikeFollowUpAdapter()
        result = adapter._fuzzy_match(
            "fitness function generation", {"t1": "generate fitness functions"}
        )
        assert result == "t1"

    def test_short_titles_shared_common_word_no_false_positive(self) -> None:
        """Short titles sharing only 'implement' should not match."""
        from src.infrastructure.persistence.spike_follow_up_adapter import SpikeFollowUpAdapter

        adapter = SpikeFollowUpAdapter()
        result = adapter._fuzzy_match("Implement parser", {"t1": "Implement deploy pipeline"})
        assert result is None


class TestFuzzyMatchParenthetical:
    """Parenthetical content contributes to matching."""

    def test_parenthetical_tools_match(self) -> None:
        """Shared parenthetical tool names boost matching."""
        from src.infrastructure.persistence.spike_follow_up_adapter import SpikeFollowUpAdapter

        adapter = SpikeFollowUpAdapter()
        result = adapter._fuzzy_match(
            "Task: Implement fitness generation (import-linter + pytestarch)",
            {"t1": "Implement alty generate fitness (import-linter + pytestarch)"},
        )
        assert result == "t1"

    def test_no_parenthetical_still_uses_keywords(self) -> None:
        """Without parenthetical, keyword overlap still works."""
        from src.infrastructure.persistence.spike_follow_up_adapter import SpikeFollowUpAdapter

        adapter = SpikeFollowUpAdapter()
        result = adapter._fuzzy_match(
            "Implement drift detection for knowledge base",
            {"t1": "Implement knowledge base drift detection"},
        )
        assert result == "t1"


class TestFuzzyMatchEmptyExisting:
    """Edge case: empty ticket dict."""

    def test_empty_existing_returns_none(self) -> None:
        from src.infrastructure.persistence.spike_follow_up_adapter import SpikeFollowUpAdapter

        adapter = SpikeFollowUpAdapter()
        result = adapter._fuzzy_match("Build something", {})
        assert result is None
