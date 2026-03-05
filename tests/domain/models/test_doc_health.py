"""Tests for document health domain models.

Covers DocHealthStatus enum, DocRegistryEntry and DocStatus value objects,
DocHealthReport aggregate, and the create_doc_status factory function.
"""

from __future__ import annotations

from datetime import date, timedelta

import pytest

from src.domain.models.doc_health import (
    BrokenLink,
    DocHealthReport,
    DocHealthStatus,
    DocRegistryEntry,
    DocStatus,
    create_doc_status,
)
from src.domain.models.errors import InvariantViolationError

# ---------------------------------------------------------------------------
# 1. DocHealthStatus enum
# ---------------------------------------------------------------------------


class TestDocHealthStatusEnum:
    def test_doc_health_status_enum_values(self) -> None:
        assert DocHealthStatus.OK.value == "ok"
        assert DocHealthStatus.STALE.value == "stale"
        assert DocHealthStatus.MISSING.value == "missing"
        assert DocHealthStatus.NO_FRONTMATTER.value == "no_frontmatter"


# ---------------------------------------------------------------------------
# 2. DocRegistryEntry value object
# ---------------------------------------------------------------------------


class TestDocRegistryEntry:
    def test_doc_registry_entry_defaults(self) -> None:
        entry = DocRegistryEntry(path="docs/PRD.md")
        assert entry.path == "docs/PRD.md"
        assert entry.owner is None
        assert entry.review_interval_days == 30

    def test_doc_registry_entry_custom_values(self) -> None:
        entry = DocRegistryEntry(path="docs/DDD.md", owner="team-lead", review_interval_days=14)
        assert entry.path == "docs/DDD.md"
        assert entry.owner == "team-lead"
        assert entry.review_interval_days == 14

    def test_doc_registry_entry_rejects_zero_interval(self) -> None:
        with pytest.raises(InvariantViolationError):
            DocRegistryEntry(path="docs/PRD.md", review_interval_days=0)

    def test_doc_registry_entry_rejects_negative_interval(self) -> None:
        with pytest.raises(InvariantViolationError):
            DocRegistryEntry(path="docs/PRD.md", review_interval_days=-5)

    def test_doc_registry_entry_is_frozen(self) -> None:
        entry = DocRegistryEntry(path="docs/PRD.md")
        with pytest.raises(AttributeError):
            entry.path = "other.md"  # type: ignore[misc]


# ---------------------------------------------------------------------------
# 3. create_doc_status factory
# ---------------------------------------------------------------------------


class TestCreateDocStatus:
    def test_create_doc_status_ok_within_interval(self) -> None:
        reviewed = date.today() - timedelta(days=10)
        status = create_doc_status(
            path="docs/PRD.md",
            exists=True,
            last_reviewed=reviewed,
            review_interval_days=30,
        )
        assert status.status == DocHealthStatus.OK
        assert status.last_reviewed == reviewed
        assert status.days_since == 10
        assert status.path == "docs/PRD.md"

    def test_create_doc_status_stale_beyond_interval(self) -> None:
        reviewed = date.today() - timedelta(days=45)
        status = create_doc_status(
            path="docs/PRD.md",
            exists=True,
            last_reviewed=reviewed,
            review_interval_days=30,
        )
        assert status.status == DocHealthStatus.STALE
        assert status.days_since == 45

    def test_create_doc_status_missing(self) -> None:
        status = create_doc_status(
            path="docs/MISSING.md",
            exists=False,
            last_reviewed=None,
        )
        assert status.status == DocHealthStatus.MISSING
        assert status.last_reviewed is None
        assert status.days_since is None

    def test_create_doc_status_no_frontmatter(self) -> None:
        status = create_doc_status(
            path="docs/NO_FM.md",
            exists=True,
            last_reviewed=None,
        )
        assert status.status == DocHealthStatus.NO_FRONTMATTER
        assert status.last_reviewed is None
        assert status.days_since is None

    def test_create_doc_status_exactly_at_interval_is_ok(self) -> None:
        reviewed = date.today() - timedelta(days=30)
        status = create_doc_status(
            path="docs/PRD.md",
            exists=True,
            last_reviewed=reviewed,
            review_interval_days=30,
        )
        assert status.status == DocHealthStatus.OK

    def test_create_doc_status_one_day_beyond_interval_is_stale(self) -> None:
        reviewed = date.today() - timedelta(days=31)
        status = create_doc_status(
            path="docs/PRD.md",
            exists=True,
            last_reviewed=reviewed,
            review_interval_days=30,
        )
        assert status.status == DocHealthStatus.STALE

    def test_create_doc_status_preserves_owner(self) -> None:
        status = create_doc_status(
            path="docs/PRD.md",
            exists=True,
            last_reviewed=date.today(),
            owner="team-lead",
        )
        assert status.owner == "team-lead"


# ---------------------------------------------------------------------------
# 4. DocHealthReport value object
# ---------------------------------------------------------------------------


class TestDocHealthReport:
    def test_doc_health_report_issue_count(self) -> None:
        statuses = (
            DocStatus(path="a.md", status=DocHealthStatus.OK),
            DocStatus(path="b.md", status=DocHealthStatus.STALE),
            DocStatus(path="c.md", status=DocHealthStatus.MISSING),
        )
        report = DocHealthReport(statuses=statuses)
        assert report.issue_count == 2

    def test_doc_health_report_total_checked(self) -> None:
        statuses = (
            DocStatus(path="a.md", status=DocHealthStatus.OK),
            DocStatus(path="b.md", status=DocHealthStatus.STALE),
        )
        report = DocHealthReport(statuses=statuses)
        assert report.total_checked == 2

    def test_doc_health_report_has_issues(self) -> None:
        statuses = (
            DocStatus(path="a.md", status=DocHealthStatus.OK),
            DocStatus(path="b.md", status=DocHealthStatus.STALE),
        )
        report = DocHealthReport(statuses=statuses)
        assert report.has_issues is True

    def test_doc_health_report_no_issues(self) -> None:
        statuses = (
            DocStatus(path="a.md", status=DocHealthStatus.OK),
            DocStatus(path="b.md", status=DocHealthStatus.OK),
        )
        report = DocHealthReport(statuses=statuses)
        assert report.has_issues is False

    def test_doc_health_report_empty(self) -> None:
        report = DocHealthReport(statuses=())
        assert report.issue_count == 0
        assert report.total_checked == 0
        assert report.has_issues is False

    def test_doc_health_report_is_frozen(self) -> None:
        report = DocHealthReport(statuses=())
        with pytest.raises(AttributeError):
            report.statuses = ()  # type: ignore[misc]

    def test_report_has_issues_with_broken_links(self) -> None:
        """Report has_issues is True when OK doc carries broken links."""
        bl = BrokenLink(line_number=5, link_text="ref", target="gone.md", reason="not found")
        statuses = (
            DocStatus(
                path="a.md",
                status=DocHealthStatus.OK,
                broken_links=(bl,),
            ),
        )
        report = DocHealthReport(statuses=statuses)
        assert report.has_issues is True
        assert report.issue_count == 1


# ---------------------------------------------------------------------------
# 5. BrokenLink value object
# ---------------------------------------------------------------------------


class TestBrokenLink:
    def test_broken_link_vo_immutable(self) -> None:
        bl = BrokenLink(line_number=10, link_text="ref", target="gone.md", reason="not found")
        with pytest.raises(AttributeError):
            bl.target = "other.md"  # type: ignore[misc]

    def test_doc_status_broken_links_default_empty(self) -> None:
        status = DocStatus(path="a.md", status=DocHealthStatus.OK)
        assert status.broken_links == ()

    def test_broken_link_rejects_zero_line_number(self) -> None:
        with pytest.raises(InvariantViolationError, match="line_number must be >= 1"):
            BrokenLink(line_number=0, link_text="x", target="x.md", reason="test")

    def test_broken_link_rejects_negative_line_number(self) -> None:
        with pytest.raises(InvariantViolationError, match="line_number must be >= 1"):
            BrokenLink(line_number=-1, link_text="x", target="x.md", reason="test")
