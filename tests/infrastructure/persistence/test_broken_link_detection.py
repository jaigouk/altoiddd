"""Tests for broken markdown link detection in doc-health (alty-2j7.12).

Covers link extraction, external/anchor skip, relative path resolution,
multiple broken links, empty targets, and anchor fragment stripping.
"""

from __future__ import annotations

from pathlib import Path  # noqa: TC003 — used at runtime for tmp_path

from src.infrastructure.persistence.filesystem_doc_scanner import FilesystemDocScanner

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _write_doc(path: Path, content: str) -> None:
    """Write a markdown file with given content."""
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content)


def _write_registry(project_dir: Path, doc_path: str) -> None:
    """Write a minimal doc-registry.toml with one entry."""
    registry_dir = project_dir / ".alty" / "maintenance"
    registry_dir.mkdir(parents=True, exist_ok=True)
    (registry_dir / "doc-registry.toml").write_text(
        f'[[docs]]\npath = "{doc_path}"\n'
    )


# ---------------------------------------------------------------------------
# Broken link detection
# ---------------------------------------------------------------------------


class TestBrokenLinkDetection:
    """Scanner detects broken internal markdown links."""

    def test_detects_broken_internal_link(self, tmp_path: Path) -> None:
        """A link to a nonexistent file produces broken_links entry."""
        doc = tmp_path / "docs" / "guide.md"
        _write_doc(doc, "---\nlast_reviewed: 2026-01-01\n---\n\nSee [setup](setup.md).\n")
        _write_registry(tmp_path, "docs/guide.md")

        scanner = FilesystemDocScanner()
        entries = scanner.load_registry(tmp_path / ".alty" / "maintenance" / "doc-registry.toml")
        statuses = scanner.scan_registered(entries, tmp_path)

        guide_status = next(s for s in statuses if s.path == "docs/guide.md")
        assert len(guide_status.broken_links) == 1
        assert guide_status.broken_links[0].target == "setup.md"
        assert guide_status.broken_links[0].link_text == "setup"

    def test_skips_external_urls(self, tmp_path: Path) -> None:
        """External URLs (https://) are not checked."""
        doc = tmp_path / "docs" / "guide.md"
        _write_doc(
            doc,
            "---\nlast_reviewed: 2026-01-01\n---\n\n"
            "See [example](https://example.com).\n",
        )
        _write_registry(tmp_path, "docs/guide.md")

        scanner = FilesystemDocScanner()
        entries = scanner.load_registry(tmp_path / ".alty" / "maintenance" / "doc-registry.toml")
        statuses = scanner.scan_registered(entries, tmp_path)

        guide_status = next(s for s in statuses if s.path == "docs/guide.md")
        assert guide_status.broken_links == ()

    def test_skips_anchor_links(self, tmp_path: Path) -> None:
        """Anchor-only links (#section) are not checked."""
        doc = tmp_path / "docs" / "guide.md"
        _write_doc(
            doc,
            "---\nlast_reviewed: 2026-01-01\n---\n\nSee [below](#details).\n",
        )
        _write_registry(tmp_path, "docs/guide.md")

        scanner = FilesystemDocScanner()
        entries = scanner.load_registry(tmp_path / ".alty" / "maintenance" / "doc-registry.toml")
        statuses = scanner.scan_registered(entries, tmp_path)

        guide_status = next(s for s in statuses if s.path == "docs/guide.md")
        assert guide_status.broken_links == ()

    def test_resolves_relative_paths(self, tmp_path: Path) -> None:
        """Relative paths resolve from the doc's directory."""
        doc = tmp_path / "docs" / "sub" / "guide.md"
        target = tmp_path / "docs" / "other" / "ref.md"
        _write_doc(doc, "---\nlast_reviewed: 2026-01-01\n---\n\nSee [ref](../other/ref.md).\n")
        _write_doc(target, "# Reference\n")
        _write_registry(tmp_path, "docs/sub/guide.md")

        scanner = FilesystemDocScanner()
        entries = scanner.load_registry(tmp_path / ".alty" / "maintenance" / "doc-registry.toml")
        statuses = scanner.scan_registered(entries, tmp_path)

        guide_status = next(s for s in statuses if s.path == "docs/sub/guide.md")
        assert guide_status.broken_links == ()

    def test_multiple_broken_links_all_reported(self, tmp_path: Path) -> None:
        """All broken links in a doc are reported."""
        doc = tmp_path / "docs" / "guide.md"
        _write_doc(
            doc,
            "---\nlast_reviewed: 2026-01-01\n---\n\n"
            "See [a](a.md), [b](b.md), and [c](c.md).\n",
        )
        _write_registry(tmp_path, "docs/guide.md")

        scanner = FilesystemDocScanner()
        entries = scanner.load_registry(tmp_path / ".alty" / "maintenance" / "doc-registry.toml")
        statuses = scanner.scan_registered(entries, tmp_path)

        guide_status = next(s for s in statuses if s.path == "docs/guide.md")
        assert len(guide_status.broken_links) == 3
        targets = {bl.target for bl in guide_status.broken_links}
        assert targets == {"a.md", "b.md", "c.md"}

    def test_no_broken_links_stays_clean(self, tmp_path: Path) -> None:
        """Doc with only valid links has empty broken_links."""
        doc = tmp_path / "docs" / "guide.md"
        target = tmp_path / "docs" / "real.md"
        _write_doc(doc, "---\nlast_reviewed: 2026-01-01\n---\n\nSee [real](real.md).\n")
        _write_doc(target, "# Real doc\n")
        _write_registry(tmp_path, "docs/guide.md")

        scanner = FilesystemDocScanner()
        entries = scanner.load_registry(tmp_path / ".alty" / "maintenance" / "doc-registry.toml")
        statuses = scanner.scan_registered(entries, tmp_path)

        guide_status = next(s for s in statuses if s.path == "docs/guide.md")
        assert guide_status.broken_links == ()

    def test_empty_link_target_reported(self, tmp_path: Path) -> None:
        """Empty link target [text]() is reported as broken."""
        doc = tmp_path / "docs" / "guide.md"
        _write_doc(doc, "---\nlast_reviewed: 2026-01-01\n---\n\nSee [empty]().\n")
        _write_registry(tmp_path, "docs/guide.md")

        scanner = FilesystemDocScanner()
        entries = scanner.load_registry(tmp_path / ".alty" / "maintenance" / "doc-registry.toml")
        statuses = scanner.scan_registered(entries, tmp_path)

        guide_status = next(s for s in statuses if s.path == "docs/guide.md")
        assert len(guide_status.broken_links) == 1
        assert guide_status.broken_links[0].reason == "empty target"

    def test_link_with_anchor_strips_fragment(self, tmp_path: Path) -> None:
        """Link with anchor [text](file.md#section) checks file.md exists."""
        doc = tmp_path / "docs" / "guide.md"
        target = tmp_path / "docs" / "real.md"
        _write_doc(doc, "---\nlast_reviewed: 2026-01-01\n---\n\nSee [sec](real.md#intro).\n")
        _write_doc(target, "# Real\n")
        _write_registry(tmp_path, "docs/guide.md")

        scanner = FilesystemDocScanner()
        entries = scanner.load_registry(tmp_path / ".alty" / "maintenance" / "doc-registry.toml")
        statuses = scanner.scan_registered(entries, tmp_path)

        guide_status = next(s for s in statuses if s.path == "docs/guide.md")
        assert guide_status.broken_links == ()


class TestFencedCodeBlocksSkipped:
    """Links inside fenced code blocks are not checked."""

    def test_link_inside_backtick_fence_skipped(self, tmp_path: Path) -> None:
        """Links inside ``` fences are ignored."""
        doc = tmp_path / "docs" / "guide.md"
        content = (
            "---\nlast_reviewed: 2026-01-01\n---\n\n"
            "```\n"
            "See [example](nonexistent.md) for reference\n"
            "```\n"
        )
        _write_doc(doc, content)
        _write_registry(tmp_path, "docs/guide.md")

        scanner = FilesystemDocScanner()
        entries = scanner.load_registry(
            tmp_path / ".alty" / "maintenance" / "doc-registry.toml",
        )
        statuses = scanner.scan_registered(entries, tmp_path)

        guide_status = next(s for s in statuses if s.path == "docs/guide.md")
        assert guide_status.broken_links == ()

    def test_link_inside_tilde_fence_skipped(self, tmp_path: Path) -> None:
        """Links inside ~~~ fences are ignored."""
        doc = tmp_path / "docs" / "guide.md"
        content = (
            "---\nlast_reviewed: 2026-01-01\n---\n\n"
            "~~~\n"
            "See [example](nonexistent.md)\n"
            "~~~\n"
        )
        _write_doc(doc, content)
        _write_registry(tmp_path, "docs/guide.md")

        scanner = FilesystemDocScanner()
        entries = scanner.load_registry(
            tmp_path / ".alty" / "maintenance" / "doc-registry.toml",
        )
        statuses = scanner.scan_registered(entries, tmp_path)

        guide_status = next(s for s in statuses if s.path == "docs/guide.md")
        assert guide_status.broken_links == ()

    def test_link_after_fence_close_still_checked(self, tmp_path: Path) -> None:
        """Links after a closed fence ARE checked."""
        doc = tmp_path / "docs" / "guide.md"
        content = (
            "---\nlast_reviewed: 2026-01-01\n---\n\n"
            "```\ncode\n```\n\n"
            "See [broken](missing.md).\n"
        )
        _write_doc(doc, content)
        _write_registry(tmp_path, "docs/guide.md")

        scanner = FilesystemDocScanner()
        entries = scanner.load_registry(
            tmp_path / ".alty" / "maintenance" / "doc-registry.toml",
        )
        statuses = scanner.scan_registered(entries, tmp_path)

        guide_status = next(s for s in statuses if s.path == "docs/guide.md")
        assert len(guide_status.broken_links) == 1


class TestEdgeCases:
    """Additional edge cases for link detection."""

    def test_image_link_not_checked(self, tmp_path: Path) -> None:
        """Image links ![alt](path) are skipped."""
        doc = tmp_path / "docs" / "guide.md"
        _write_doc(
            doc,
            "---\nlast_reviewed: 2026-01-01\n---\n\n![logo](nonexistent.png)\n",
        )
        _write_registry(tmp_path, "docs/guide.md")

        scanner = FilesystemDocScanner()
        entries = scanner.load_registry(
            tmp_path / ".alty" / "maintenance" / "doc-registry.toml",
        )
        statuses = scanner.scan_registered(entries, tmp_path)

        guide_status = next(s for s in statuses if s.path == "docs/guide.md")
        assert guide_status.broken_links == ()

    def test_malformed_link_no_paren_ignored(self, tmp_path: Path) -> None:
        """Malformed [text] without (target) is not matched."""
        doc = tmp_path / "docs" / "guide.md"
        _write_doc(
            doc,
            "---\nlast_reviewed: 2026-01-01\n---\n\nSee [text] for details.\n",
        )
        _write_registry(tmp_path, "docs/guide.md")

        scanner = FilesystemDocScanner()
        entries = scanner.load_registry(
            tmp_path / ".alty" / "maintenance" / "doc-registry.toml",
        )
        statuses = scanner.scan_registered(entries, tmp_path)

        guide_status = next(s for s in statuses if s.path == "docs/guide.md")
        assert guide_status.broken_links == ()

    def test_directory_target_not_reported(self, tmp_path: Path) -> None:
        """Link to existing directory is not reported as broken."""
        doc = tmp_path / "docs" / "guide.md"
        subdir = tmp_path / "docs" / "subdir"
        subdir.mkdir(parents=True)
        _write_doc(
            doc,
            "---\nlast_reviewed: 2026-01-01\n---\n\nSee [sub](subdir).\n",
        )
        _write_registry(tmp_path, "docs/guide.md")

        scanner = FilesystemDocScanner()
        entries = scanner.load_registry(
            tmp_path / ".alty" / "maintenance" / "doc-registry.toml",
        )
        statuses = scanner.scan_registered(entries, tmp_path)

        guide_status = next(s for s in statuses if s.path == "docs/guide.md")
        assert guide_status.broken_links == ()


class TestBrokenLinksInUnregistered:
    """Broken links also detected in unregistered docs."""

    def test_unregistered_doc_broken_links(self, tmp_path: Path) -> None:
        """Unregistered doc with broken link has non-empty broken_links."""
        docs_dir = tmp_path / "docs"
        doc = docs_dir / "orphan.md"
        _write_doc(doc, "---\nlast_reviewed: 2026-01-01\n---\n\nSee [gone](gone.md).\n")

        scanner = FilesystemDocScanner()
        statuses = scanner.scan_unregistered(docs_dir, frozenset())

        orphan = next(s for s in statuses if "orphan.md" in s.path)
        assert len(orphan.broken_links) == 1
        assert orphan.broken_links[0].target == "gone.md"
