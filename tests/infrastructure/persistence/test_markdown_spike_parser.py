"""Tests for MarkdownSpikeReportParser.

RED phase: these tests define the contract for extracting FollowUpIntent
value objects from Markdown spike research reports.
"""

from __future__ import annotations

from pathlib import Path

import pytest

# ── Helpers ──────────────────────────────────────────────────────────


def _write_report(tmp_path: Path, content: str) -> Path:
    """Write a Markdown report to a temp file and return its path."""
    report = tmp_path / "report.md"
    report.write_text(content)
    return report


# ── Heading detection ────────────────────────────────────────────────


class TestHeadingDetection:
    def test_detects_h2_follow_up_tickets(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.markdown_spike_parser import MarkdownSpikeReportParser

        report = _write_report(
            tmp_path,
            "# Spike Report\n\n"
            "## Follow-Up Tickets\n\n"
            "### Ticket 1: Implement feature A\n\n"
            "Description of A.\n",
        )
        parser = MarkdownSpikeReportParser()
        intents = parser.parse(report)
        assert len(intents) == 1
        assert intents[0].title == "Implement feature A"

    def test_detects_h2_follow_up_implementation_tickets(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.markdown_spike_parser import MarkdownSpikeReportParser

        report = _write_report(
            tmp_path,
            "# Report\n\n"
            "## 8. Follow-Up Implementation Tickets\n\n"
            "### Ticket 1: Build the parser\n\n"
            "Details.\n",
        )
        parser = MarkdownSpikeReportParser()
        intents = parser.parse(report)
        assert len(intents) == 1
        assert intents[0].title == "Build the parser"

    def test_detects_h2_follow_up_tickets_needed(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.markdown_spike_parser import MarkdownSpikeReportParser

        report = _write_report(
            tmp_path,
            "# Report\n\n"
            "## Follow-up Tickets Needed\n\n"
            "### 1. Create domain model\n\n"
            "Steps here.\n",
        )
        parser = MarkdownSpikeReportParser()
        intents = parser.parse(report)
        assert len(intents) == 1
        assert intents[0].title == "Create domain model"

    def test_detects_h3_follow_up_heading(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.markdown_spike_parser import MarkdownSpikeReportParser

        report = _write_report(
            tmp_path,
            "# Report\n\n"
            "### Follow-up Tickets Needed\n\n"
            "- **Task A**: Build stuff\n"
            "- **Task B**: Fix stuff\n",
        )
        parser = MarkdownSpikeReportParser()
        intents = parser.parse(report)
        assert len(intents) == 2

    def test_detects_numbered_heading_prefix(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.markdown_spike_parser import MarkdownSpikeReportParser

        report = _write_report(
            tmp_path,
            "## 13. Follow-Up Implementation Tickets\n\n"
            "### Ticket 1: Gap analysis tool\n\n"
            "Build it.\n",
        )
        parser = MarkdownSpikeReportParser()
        intents = parser.parse(report)
        assert len(intents) == 1
        assert intents[0].title == "Gap analysis tool"


# ── Ticket extraction formats ───────────────────────────────────────


class TestTicketExtraction:
    def test_extracts_h3_ticket_heading_format(self, tmp_path: Path) -> None:
        """### Ticket N: Title"""
        from src.infrastructure.persistence.markdown_spike_parser import MarkdownSpikeReportParser

        report = _write_report(
            tmp_path,
            (
                "## Follow-Up Implementation Tickets\n\n"
                "### Ticket 1: Implement SessionStore\n\n"
                "**Type:** Task\n**Priority:** P2\n\n"
                "### Ticket 2: Add serialization\n\n"
                "**Type:** Task\n**Priority:** P3\n"
            ),
        )
        parser = MarkdownSpikeReportParser()
        intents = parser.parse(report)
        assert len(intents) == 2
        assert intents[0].title == "Implement SessionStore"
        assert intents[1].title == "Add serialization"

    def test_extracts_numbered_heading_format(self, tmp_path: Path) -> None:
        """### N. Title"""
        from src.infrastructure.persistence.markdown_spike_parser import MarkdownSpikeReportParser

        report = _write_report(
            tmp_path,
            (
                "## Follow-up Tickets Needed\n\n"
                "### 1. Create domain model\n\nSteps.\n\n"
                "### 2. Wire infrastructure\n\nMore steps.\n"
            ),
        )
        parser = MarkdownSpikeReportParser()
        intents = parser.parse(report)
        assert len(intents) == 2
        assert intents[0].title == "Create domain model"
        assert intents[1].title == "Wire infrastructure"

    def test_extracts_bold_list_item_format(self, tmp_path: Path) -> None:
        """- **Title**: Description"""
        from src.infrastructure.persistence.markdown_spike_parser import MarkdownSpikeReportParser

        report = _write_report(
            tmp_path,
            (
                "## Follow-Up Tickets\n\n"
                "- **Build CLI command**: Wire up the subcommand tree\n"
                "- **Add test coverage**: Cover edge cases\n"
            ),
        )
        parser = MarkdownSpikeReportParser()
        intents = parser.parse(report)
        assert len(intents) == 2
        assert intents[0].title == "Build CLI command"
        assert intents[0].description == "Wire up the subcommand tree"
        assert intents[1].title == "Add test coverage"

    def test_extracts_plain_list_item_format(self, tmp_path: Path) -> None:
        """- Title text (no bold)"""
        from src.infrastructure.persistence.markdown_spike_parser import MarkdownSpikeReportParser

        report = _write_report(
            tmp_path,
            "## Follow-Up Tickets\n\n- Create fitness functions\n- Add import linter\n",
        )
        parser = MarkdownSpikeReportParser()
        intents = parser.parse(report)
        assert len(intents) == 2
        assert intents[0].title == "Create fitness functions"
        assert intents[1].title == "Add import linter"

    def test_extracts_optional_prefix_in_parentheses(self, tmp_path: Path) -> None:
        """### Ticket 2: (Optional) Add Serialization"""
        from src.infrastructure.persistence.markdown_spike_parser import MarkdownSpikeReportParser

        report = _write_report(
            tmp_path,
            "## Follow-Up Implementation Tickets\n\n"
            "### Ticket 2: (Optional) Add Serialization\n\n"
            "Details.\n",
        )
        parser = MarkdownSpikeReportParser()
        intents = parser.parse(report)
        assert len(intents) == 1
        assert intents[0].title == "(Optional) Add Serialization"

    def test_extracts_description_from_body_text(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.markdown_spike_parser import MarkdownSpikeReportParser

        report = _write_report(
            tmp_path,
            (
                "## Follow-Up Tickets\n\n"
                "### Ticket 1: Build parser\n\n"
                "Create a Markdown parser that extracts follow-up intents.\n\n"
                "**Type:** Task\n"
            ),
        )
        parser = MarkdownSpikeReportParser()
        intents = parser.parse(report)
        assert len(intents) == 1
        assert "Markdown parser" in intents[0].description


# ── Edge cases ──────────────────────────────────────────────────────


class TestParserEdgeCases:
    def test_no_follow_up_section_returns_empty(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.markdown_spike_parser import MarkdownSpikeReportParser

        report = _write_report(
            tmp_path,
            "# Spike Report\n\n"
            "## Findings\n\n"
            "Some research results.\n\n"
            "## References\n\n"
            "- Link 1\n",
        )
        parser = MarkdownSpikeReportParser()
        intents = parser.parse(report)
        assert intents == ()

    def test_empty_follow_up_section_returns_empty(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.markdown_spike_parser import MarkdownSpikeReportParser

        report = _write_report(
            tmp_path,
            "# Report\n\n## Follow-Up Tickets\n\n## References\n\n- Link\n",
        )
        parser = MarkdownSpikeReportParser()
        intents = parser.parse(report)
        assert intents == ()

    def test_nonexistent_file_returns_empty(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.markdown_spike_parser import MarkdownSpikeReportParser

        parser = MarkdownSpikeReportParser()
        intents = parser.parse(tmp_path / "does_not_exist.md")
        assert intents == ()

    def test_stops_at_next_same_level_heading(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.markdown_spike_parser import MarkdownSpikeReportParser

        report = _write_report(
            tmp_path,
            (
                "## Follow-Up Tickets\n\n"
                "### Ticket 1: Real ticket\n\nDetails.\n\n"
                "## References\n\n"
                "### Not a ticket: This is a reference\n\nLink.\n"
            ),
        )
        parser = MarkdownSpikeReportParser()
        intents = parser.parse(report)
        assert len(intents) == 1
        assert intents[0].title == "Real ticket"

    def test_multiple_follow_up_sections_uses_first(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.markdown_spike_parser import MarkdownSpikeReportParser

        report = _write_report(
            tmp_path,
            (
                "## Follow-Up Tickets\n\n"
                "### Ticket 1: First section ticket\n\nDetails.\n\n"
                "## Other Stuff\n\nContent.\n\n"
                "## Follow-Up Tickets\n\n"
                "### Ticket 1: Second section ticket\n\nDetails.\n"
            ),
        )
        parser = MarkdownSpikeReportParser()
        intents = parser.parse(report)
        assert len(intents) >= 1
        assert intents[0].title == "First section ticket"

    def test_implements_spike_report_parser_port(self) -> None:
        from src.application.ports.spike_report_parser_port import SpikeReportParserPort
        from src.infrastructure.persistence.markdown_spike_parser import MarkdownSpikeReportParser

        parser = MarkdownSpikeReportParser()
        assert isinstance(parser, SpikeReportParserPort)

    def test_real_report_gap_analysis(self) -> None:
        """Smoke test against a real spike report in docs/research/."""
        from src.infrastructure.persistence.markdown_spike_parser import MarkdownSpikeReportParser

        report_path = Path("docs/research/20260301_mcp_multi_turn_sessions.md")
        if not report_path.exists():
            pytest.skip("Real report not available")
        parser = MarkdownSpikeReportParser()
        intents = parser.parse(report_path)
        assert len(intents) >= 1
        titles = [i.title for i in intents]
        assert any("SessionStore" in t or "Serialization" in t for t in titles)
