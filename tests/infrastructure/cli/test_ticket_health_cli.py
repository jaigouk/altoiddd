"""Tests for the ``alty ticket-health`` CLI command (2j7.10).

Covers: freshness report, flagged ticket listing, threshold labels,
exit codes, and --epic flag.
"""

from __future__ import annotations

from unittest.mock import patch

from typer.testing import CliRunner

from src.domain.models.ticket_freshness import (
    ContextDiff,
    FlaggedTicket,
    FreshnessFlag,
    TicketFreshnessStatus,
    TicketHealthReport,
)
from src.infrastructure.cli.main import app

runner = CliRunner()


def _make_report(
    *,
    total_open: int = 10,
    flagged: int = 0,
) -> TicketHealthReport:
    """Build a TicketHealthReport with N flagged tickets."""
    flagged_tickets = tuple(
        FlaggedTicket(
            ticket_id=f"alty-t.{i}",
            title=f"Test ticket {i}",
            flags=(
                FreshnessFlag(
                    context_diff=ContextDiff(
                        summary=f"Changed something for ticket {i}",
                        triggering_ticket_id=f"alty-t.{i + 100}",
                        produced_at="2026-03-01",
                    ),
                    flagged_at="2026-03-01",
                ),
            ),
            status=TicketFreshnessStatus.REVIEW_NEEDED,
        )
        for i in range(flagged)
    )
    return TicketHealthReport(
        flagged_tickets=flagged_tickets,
        total_open=total_open,
        oldest_last_reviewed="2026-02-20",
    )


class TestTicketHealthReport:
    """alty ticket-health displays a formatted health report."""

    def test_shows_freshness_percentage(self) -> None:
        """Report includes the freshness percentage."""
        report = _make_report(total_open=10, flagged=2)
        with patch(
            "src.infrastructure.cli.main._build_ticket_health_report",
            return_value=report,
        ):
            result = runner.invoke(app, ["ticket-health"])
        assert "80.0%" in result.output

    def test_shows_healthy_threshold(self) -> None:
        """90-100% freshness labeled 'healthy'."""
        report = _make_report(total_open=10, flagged=0)
        with patch(
            "src.infrastructure.cli.main._build_ticket_health_report",
            return_value=report,
        ):
            result = runner.invoke(app, ["ticket-health"])
        assert "healthy" in result.output.lower()

    def test_shows_acceptable_threshold(self) -> None:
        """70-89% freshness labeled 'acceptable'."""
        report = _make_report(total_open=10, flagged=2)
        with patch(
            "src.infrastructure.cli.main._build_ticket_health_report",
            return_value=report,
        ):
            result = runner.invoke(app, ["ticket-health"])
        assert "acceptable" in result.output.lower()

    def test_shows_action_needed_threshold(self) -> None:
        """<70% freshness labeled 'action needed'."""
        report = _make_report(total_open=10, flagged=4)
        with patch(
            "src.infrastructure.cli.main._build_ticket_health_report",
            return_value=report,
        ):
            result = runner.invoke(app, ["ticket-health"])
        assert "action needed" in result.output.lower()

    def test_shows_flagged_ticket_list(self) -> None:
        """Flagged tickets listed with IDs and context."""
        report = _make_report(total_open=10, flagged=2)
        with patch(
            "src.infrastructure.cli.main._build_ticket_health_report",
            return_value=report,
        ):
            result = runner.invoke(app, ["ticket-health"])
        assert "alty-t.0" in result.output
        assert "alty-t.1" in result.output
        assert "Changed something" in result.output

    def test_shows_open_and_flagged_counts(self) -> None:
        """Report shows total open and flagged counts."""
        report = _make_report(total_open=20, flagged=3)
        with patch(
            "src.infrastructure.cli.main._build_ticket_health_report",
            return_value=report,
        ):
            result = runner.invoke(app, ["ticket-health"])
        assert "20" in result.output
        assert "3" in result.output


class TestTicketHealthExitCodes:
    """Exit code behavior for alty ticket-health."""

    def test_exit_0_when_healthy(self) -> None:
        """No flagged tickets → exit 0."""
        report = _make_report(total_open=10, flagged=0)
        with patch(
            "src.infrastructure.cli.main._build_ticket_health_report",
            return_value=report,
        ):
            result = runner.invoke(app, ["ticket-health"])
        assert result.exit_code == 0

    def test_exit_1_when_issues(self) -> None:
        """Flagged tickets → exit 1."""
        report = _make_report(total_open=10, flagged=2)
        with patch(
            "src.infrastructure.cli.main._build_ticket_health_report",
            return_value=report,
        ):
            result = runner.invoke(app, ["ticket-health"])
        assert result.exit_code == 1
