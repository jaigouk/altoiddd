"""Tests for BeadsTicketReader bd CLI integration (2j7.10).

Tests label enrichment via ``bd query label=review_needed`` and
comment parsing via ``bd comments <id>``.
"""

from __future__ import annotations

import json
from typing import TYPE_CHECKING, Any
from unittest.mock import patch

_SUBPROCESS_RUN = (
    "src.infrastructure.external.beads_ticket_reader.subprocess.run"
)

if TYPE_CHECKING:
    from pathlib import Path
    from subprocess import CompletedProcess


def _write_issues_jsonl(beads_dir: Path, issues: list[dict[str, Any]]) -> None:
    """Write a list of issue dicts to .beads/issues.jsonl."""
    beads_dir.mkdir(parents=True, exist_ok=True)
    with (beads_dir / "issues.jsonl").open("w") as f:
        for issue in issues:
            f.write(json.dumps(issue) + "\n")


def _open_issue(ticket_id: str = "alty-2j7.8", title: str = "Doc review") -> dict[str, Any]:
    return {
        "id": ticket_id,
        "title": title,
        "description": "",
        "status": "open",
        "priority": "P2",
        "issue_type": "task",
        "owner": "",
        "created_at": "2026-03-01",
        "created_by": "agent",
        "updated_at": "2026-03-01",
    }


BD_QUERY_FLAGGED_OUTPUT = """\
Found 2 issues:
○ alty-2j7.8 [● P2] [task] [review_needed] - alty doc-review command
○ alty-2j7.11 [● P2] [task] [review_needed] - Two-tier ticket generation
"""

BD_QUERY_NO_RESULTS = """\
No issues found matching query: label=review_needed
"""

BD_COMMENTS_RIPPLE = """\
Comments on alty-2j7.8:

[Jaigouk Kim] at 2026-03-04 21:11
  **Ripple review needed** -- `alty-2j7.5` (Knowledge base seed content) was closed.

  **What changed:** Created 12 MVP seed files for .alty/knowledge/ with DDD patterns.

  **Review checklist:**
  - [ ] Read the description -- does it still match the new context?

[Jaigouk Kim] at 2026-03-04 21:11
  **Ripple review needed** -- `alty-2j7.7` (doc-health command) was closed.

  **What changed:** Wired alty doc-health CLI to DocHealthHandler and FilesystemDocScanner.

  **Review checklist:**
  - [ ] Read the description -- does it still match the new context?
"""

BD_COMMENTS_EMPTY = """\
Comments on alty-2j7.11:

(no comments)
"""

BD_COMMENTS_NO_RIPPLE = """\
Comments on alty-2j7.8:

[Jaigouk Kim] at 2026-03-04 19:44
  **Reviewed:** 2026-03-04
  **Triggered by:** `alty-2j7.1`, `alty-2j7.2`
  **Verdict:** unchanged
  **Changes:** No description updates needed.
"""


def _mock_run(returncode: int = 0, stdout: str = "", stderr: str = "") -> CompletedProcess[str]:
    from subprocess import CompletedProcess

    return CompletedProcess(args=[], returncode=returncode, stdout=stdout, stderr=stderr)


class TestBeadsReaderLabelEnrichment:
    """read_open_tickets enriches labels via bd query."""

    def test_enriches_flagged_tickets_with_review_needed(self, tmp_path: Path) -> None:
        from src.infrastructure.external.beads_ticket_reader import BeadsTicketReader

        beads_dir = tmp_path / ".beads"
        _write_issues_jsonl(
            beads_dir,
            [
                _open_issue("alty-2j7.8", "Doc review"),
                _open_issue("alty-2j7.11", "Two-tier"),
                _open_issue("alty-2j7.12", "Broken links"),
            ],
        )

        with patch(_SUBPROCESS_RUN, return_value=_mock_run(stdout=BD_QUERY_FLAGGED_OUTPUT)):
            reader = BeadsTicketReader(beads_dir=beads_dir)
            tickets = reader.read_open_tickets()

        flagged = {t.ticket_id for t in tickets if "review_needed" in t.labels}
        assert flagged == {"alty-2j7.8", "alty-2j7.11"}

        unflagged = {t.ticket_id for t in tickets if "review_needed" not in t.labels}
        assert unflagged == {"alty-2j7.12"}

    def test_no_flagged_tickets_returns_empty_labels(self, tmp_path: Path) -> None:
        from src.infrastructure.external.beads_ticket_reader import BeadsTicketReader

        beads_dir = tmp_path / ".beads"
        _write_issues_jsonl(beads_dir, [_open_issue("alty-2j7.8", "Doc review")])

        with patch(_SUBPROCESS_RUN, return_value=_mock_run(stdout=BD_QUERY_NO_RESULTS)):
            reader = BeadsTicketReader(beads_dir=beads_dir)
            tickets = reader.read_open_tickets()

        assert all(t.labels == () for t in tickets)

    def test_bd_not_found_returns_empty_labels(self, tmp_path: Path) -> None:
        from src.infrastructure.external.beads_ticket_reader import BeadsTicketReader

        beads_dir = tmp_path / ".beads"
        _write_issues_jsonl(beads_dir, [_open_issue("alty-2j7.8", "Doc review")])

        with patch(_SUBPROCESS_RUN, side_effect=FileNotFoundError):
            reader = BeadsTicketReader(beads_dir=beads_dir)
            tickets = reader.read_open_tickets()

        assert len(tickets) == 1
        assert tickets[0].labels == ()

    def test_bd_timeout_returns_empty_labels(self, tmp_path: Path) -> None:
        import subprocess

        from src.infrastructure.external.beads_ticket_reader import BeadsTicketReader

        beads_dir = tmp_path / ".beads"
        _write_issues_jsonl(beads_dir, [_open_issue("alty-2j7.8", "Doc review")])

        with patch(_SUBPROCESS_RUN, side_effect=subprocess.TimeoutExpired(cmd="bd", timeout=10)):
            reader = BeadsTicketReader(beads_dir=beads_dir)
            tickets = reader.read_open_tickets()

        assert len(tickets) == 1
        assert tickets[0].labels == ()


class TestBeadsReaderCommentParsing:
    """read_flags parses ripple comments from bd comments output."""

    def test_parses_ripple_comments(self, tmp_path: Path) -> None:
        from src.infrastructure.external.beads_ticket_reader import BeadsTicketReader

        beads_dir = tmp_path / ".beads"
        beads_dir.mkdir(parents=True)

        with patch(_SUBPROCESS_RUN, return_value=_mock_run(stdout=BD_COMMENTS_RIPPLE)):
            reader = BeadsTicketReader(beads_dir=beads_dir)
            flags = reader.read_flags("alty-2j7.8")

        assert len(flags) == 2
        trigger_ids = {f.context_diff.triggering_ticket_id for f in flags}
        assert trigger_ids == {"alty-2j7.5", "alty-2j7.7"}

    def test_extracts_what_changed_as_summary(self, tmp_path: Path) -> None:
        from src.infrastructure.external.beads_ticket_reader import BeadsTicketReader

        beads_dir = tmp_path / ".beads"
        beads_dir.mkdir(parents=True)

        with patch(_SUBPROCESS_RUN, return_value=_mock_run(stdout=BD_COMMENTS_RIPPLE)):
            reader = BeadsTicketReader(beads_dir=beads_dir)
            flags = reader.read_flags("alty-2j7.8")

        summaries = [f.context_diff.summary for f in flags]
        assert any("12 MVP seed files" in s for s in summaries)
        assert any("doc-health CLI" in s.lower() or "DocHealthHandler" in s for s in summaries)

    def test_no_comments_returns_empty(self, tmp_path: Path) -> None:
        from src.infrastructure.external.beads_ticket_reader import BeadsTicketReader

        beads_dir = tmp_path / ".beads"
        beads_dir.mkdir(parents=True)

        with patch(_SUBPROCESS_RUN, return_value=_mock_run(stdout=BD_COMMENTS_EMPTY)):
            reader = BeadsTicketReader(beads_dir=beads_dir)
            flags = reader.read_flags("alty-2j7.11")

        assert flags == ()

    def test_non_ripple_comments_skipped(self, tmp_path: Path) -> None:
        from src.infrastructure.external.beads_ticket_reader import BeadsTicketReader

        beads_dir = tmp_path / ".beads"
        beads_dir.mkdir(parents=True)

        with patch(_SUBPROCESS_RUN, return_value=_mock_run(stdout=BD_COMMENTS_NO_RIPPLE)):
            reader = BeadsTicketReader(beads_dir=beads_dir)
            flags = reader.read_flags("alty-2j7.8")

        assert flags == ()

    def test_bd_not_found_returns_empty(self, tmp_path: Path) -> None:
        from src.infrastructure.external.beads_ticket_reader import BeadsTicketReader

        beads_dir = tmp_path / ".beads"
        beads_dir.mkdir(parents=True)

        with patch(_SUBPROCESS_RUN, side_effect=FileNotFoundError):
            reader = BeadsTicketReader(beads_dir=beads_dir)
            flags = reader.read_flags("alty-2j7.8")

        assert flags == ()
