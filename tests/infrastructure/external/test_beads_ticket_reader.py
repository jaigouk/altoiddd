"""Tests for BeadsTicketReader anticorruption layer.

Verifies JSONL parsing, filtering, and graceful error handling.
"""

from __future__ import annotations

import json
from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from pathlib import Path

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _write_issues_jsonl(beads_dir: Path, issues: list[dict[str, Any]]) -> None:
    """Write a list of issue dicts to .beads/issues.jsonl."""
    beads_dir.mkdir(parents=True, exist_ok=True)
    jsonl_path = beads_dir / "issues.jsonl"
    with jsonl_path.open("w") as f:
        for issue in issues:
            f.write(json.dumps(issue) + "\n")


def _open_issue(
    ticket_id: str = "k7m.25",
    title: str = "Ticket Health",
    **overrides: object,
) -> dict[str, Any]:
    """Build an open issue dict matching beads JSONL format."""
    base: dict[str, Any] = {
        "id": ticket_id,
        "title": title,
        "description": "Some description",
        "status": "open",
        "priority": "P1",
        "issue_type": "task",
        "owner": "",
        "created_at": "2026-03-01",
        "created_by": "agent",
        "updated_at": "2026-03-01",
    }
    base.update(overrides)
    return base


def _closed_issue(ticket_id: str = "k7m.19", title: str = "Fitness") -> dict[str, Any]:
    base = _open_issue(ticket_id=ticket_id, title=title)
    base["status"] = "closed"
    base["closed_at"] = "2026-03-01"
    base["close_reason"] = "done"
    return base


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------


class TestBeadsTicketReaderOpenTickets:
    def test_reader_reads_open_tickets(self, tmp_path: Path) -> None:
        from src.infrastructure.external.beads_ticket_reader import BeadsTicketReader

        beads_dir = tmp_path / ".beads"
        _write_issues_jsonl(
            beads_dir,
            [
                _open_issue("k7m.25", "Ticket Health"),
                _open_issue("k7m.20", "Ticket Gen"),
            ],
        )

        reader = BeadsTicketReader(beads_dir=beads_dir)
        tickets = reader.read_open_tickets()

        assert len(tickets) == 2
        ids = {t.ticket_id for t in tickets}
        assert ids == {"k7m.25", "k7m.20"}

    def test_reader_filters_closed_tickets(self, tmp_path: Path) -> None:
        from src.infrastructure.external.beads_ticket_reader import BeadsTicketReader

        beads_dir = tmp_path / ".beads"
        _write_issues_jsonl(
            beads_dir,
            [
                _open_issue("k7m.25", "Open ticket"),
                _closed_issue("k7m.19", "Closed ticket"),
            ],
        )

        reader = BeadsTicketReader(beads_dir=beads_dir)
        tickets = reader.read_open_tickets()

        assert len(tickets) == 1
        assert tickets[0].ticket_id == "k7m.25"

    def test_reader_handles_missing_dir(self, tmp_path: Path) -> None:
        from src.infrastructure.external.beads_ticket_reader import BeadsTicketReader

        beads_dir = tmp_path / ".beads"  # does not exist
        reader = BeadsTicketReader(beads_dir=beads_dir)
        tickets = reader.read_open_tickets()

        assert tickets == ()

    def test_reader_handles_corrupted_lines(self, tmp_path: Path) -> None:
        from src.infrastructure.external.beads_ticket_reader import BeadsTicketReader

        beads_dir = tmp_path / ".beads"
        beads_dir.mkdir(parents=True)
        jsonl_path = beads_dir / "issues.jsonl"
        with jsonl_path.open("w") as f:
            f.write(json.dumps(_open_issue("k7m.25", "Good")) + "\n")
            f.write("this is not valid json\n")
            f.write(json.dumps(_open_issue("k7m.20", "Also good")) + "\n")

        reader = BeadsTicketReader(beads_dir=beads_dir)
        tickets = reader.read_open_tickets()

        # Should skip the corrupted line and return the two valid tickets
        assert len(tickets) == 2

    def test_reader_extracts_title(self, tmp_path: Path) -> None:
        from src.infrastructure.external.beads_ticket_reader import BeadsTicketReader

        beads_dir = tmp_path / ".beads"
        _write_issues_jsonl(
            beads_dir,
            [_open_issue("k7m.25", "My Title")],
        )

        reader = BeadsTicketReader(beads_dir=beads_dir)
        tickets = reader.read_open_tickets()

        assert tickets[0].title == "My Title"


class TestBeadsTicketReaderFlags:
    def test_reader_reads_flags_empty(self, tmp_path: Path) -> None:
        """When no comments exist, read_flags returns empty tuple."""
        from src.infrastructure.external.beads_ticket_reader import BeadsTicketReader

        beads_dir = tmp_path / ".beads"
        beads_dir.mkdir(parents=True)

        reader = BeadsTicketReader(beads_dir=beads_dir)
        flags = reader.read_flags("k7m.25")

        assert flags == ()

    def test_reader_reads_flags_from_comments(self, tmp_path: Path) -> None:
        """When interactions.jsonl has ripple comments, parse them as flags."""
        from src.infrastructure.external.beads_ticket_reader import BeadsTicketReader

        beads_dir = tmp_path / ".beads"
        beads_dir.mkdir(parents=True)

        # Write an interactions JSONL with a ripple comment
        interactions_path = beads_dir / "interactions.jsonl"
        interaction = {
            "issue_id": "k7m.25",
            "type": "comment",
            "body": (
                "**Ripple context diff from `k7m.19`:**\n"
                "Implemented fitness test generation aggregate"
            ),
            "created_at": "2026-03-01T10:00:00",
            "created_by": "agent",
        }
        with interactions_path.open("w") as f:
            f.write(json.dumps(interaction) + "\n")

        reader = BeadsTicketReader(beads_dir=beads_dir)
        flags = reader.read_flags("k7m.25")

        assert len(flags) == 1
        assert "fitness" in flags[0].context_diff.summary.lower()

    def test_reader_reads_flags_filters_by_ticket(self, tmp_path: Path) -> None:
        """Only return flags for the requested ticket_id."""
        from src.infrastructure.external.beads_ticket_reader import BeadsTicketReader

        beads_dir = tmp_path / ".beads"
        beads_dir.mkdir(parents=True)

        interactions_path = beads_dir / "interactions.jsonl"
        with interactions_path.open("w") as f:
            f.write(
                json.dumps(
                    {
                        "issue_id": "k7m.25",
                        "type": "comment",
                        "body": ("**Ripple context diff from `k7m.19`:**\nChange for 25"),
                        "created_at": "2026-03-01T10:00:00",
                        "created_by": "agent",
                    }
                )
                + "\n"
            )
            f.write(
                json.dumps(
                    {
                        "issue_id": "k7m.20",
                        "type": "comment",
                        "body": ("**Ripple context diff from `k7m.18`:**\nChange for 20"),
                        "created_at": "2026-03-01T11:00:00",
                        "created_by": "agent",
                    }
                )
                + "\n"
            )

        reader = BeadsTicketReader(beads_dir=beads_dir)
        flags = reader.read_flags("k7m.25")

        assert len(flags) == 1
        assert "25" in flags[0].context_diff.summary
