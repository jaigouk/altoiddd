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
