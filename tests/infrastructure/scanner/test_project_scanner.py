"""Tests for ProjectScanner infrastructure adapter.

Uses tmp_path fixtures to create real filesystem structures for scanning.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.infrastructure.scanner.project_scanner import ProjectScanner

if TYPE_CHECKING:
    from pathlib import Path


class TestProjectScannerEmptyProject:
    def test_scan_empty_project(self, tmp_path: Path) -> None:
        scanner = ProjectScanner()
        scan = scanner.scan(tmp_path)

        assert scan.project_dir == tmp_path
        assert scan.existing_docs == ()
        assert scan.existing_configs == ()
        assert scan.existing_structure == ()
        assert scan.has_knowledge_dir is False
        assert scan.has_agents_md is False
        assert scan.has_git is False


class TestProjectScannerDocs:
    def test_scan_finds_prd(self, tmp_path: Path) -> None:
        (tmp_path / "docs").mkdir()
        (tmp_path / "docs" / "PRD.md").write_text("# PRD")
        scanner = ProjectScanner()
        scan = scanner.scan(tmp_path)
        assert "docs/PRD.md" in scan.existing_docs

    def test_scan_finds_ddd(self, tmp_path: Path) -> None:
        (tmp_path / "docs").mkdir()
        (tmp_path / "docs" / "DDD.md").write_text("# DDD")
        scanner = ProjectScanner()
        scan = scanner.scan(tmp_path)
        assert "docs/DDD.md" in scan.existing_docs

    def test_scan_finds_architecture(self, tmp_path: Path) -> None:
        (tmp_path / "docs").mkdir()
        (tmp_path / "docs" / "ARCHITECTURE.md").write_text("# Arch")
        scanner = ProjectScanner()
        scan = scanner.scan(tmp_path)
        assert "docs/ARCHITECTURE.md" in scan.existing_docs

    def test_scan_finds_agents_md(self, tmp_path: Path) -> None:
        (tmp_path / "AGENTS.md").write_text("# AGENTS")
        scanner = ProjectScanner()
        scan = scanner.scan(tmp_path)
        assert "AGENTS.md" in scan.existing_docs
        assert scan.has_agents_md is True


class TestProjectScannerConfigs:
    def test_scan_finds_claude_md(self, tmp_path: Path) -> None:
        (tmp_path / ".claude").mkdir()
        (tmp_path / ".claude" / "CLAUDE.md").write_text("# Claude")
        scanner = ProjectScanner()
        scan = scanner.scan(tmp_path)
        assert ".claude/CLAUDE.md" in scan.existing_configs

    def test_scan_finds_beads(self, tmp_path: Path) -> None:
        (tmp_path / ".beads").mkdir()
        (tmp_path / ".beads" / "issues.jsonl").write_text("{}")
        scanner = ProjectScanner()
        scan = scanner.scan(tmp_path)
        assert ".beads/issues.jsonl" in scan.existing_configs

    def test_scan_finds_pyproject(self, tmp_path: Path) -> None:
        (tmp_path / "pyproject.toml").write_text("[project]")
        scanner = ProjectScanner()
        scan = scanner.scan(tmp_path)
        assert "pyproject.toml" in scan.existing_configs


class TestProjectScannerStructure:
    def test_scan_finds_domain_dir(self, tmp_path: Path) -> None:
        (tmp_path / "src" / "domain").mkdir(parents=True)
        scanner = ProjectScanner()
        scan = scanner.scan(tmp_path)
        assert "src/domain/" in scan.existing_structure

    def test_scan_finds_application_dir(self, tmp_path: Path) -> None:
        (tmp_path / "src" / "application").mkdir(parents=True)
        scanner = ProjectScanner()
        scan = scanner.scan(tmp_path)
        assert "src/application/" in scan.existing_structure

    def test_scan_finds_infrastructure_dir(self, tmp_path: Path) -> None:
        (tmp_path / "src" / "infrastructure").mkdir(parents=True)
        scanner = ProjectScanner()
        scan = scanner.scan(tmp_path)
        assert "src/infrastructure/" in scan.existing_structure


class TestProjectScannerSpecialDirs:
    def test_scan_finds_knowledge_dir(self, tmp_path: Path) -> None:
        (tmp_path / ".alty" / "knowledge").mkdir(parents=True)
        scanner = ProjectScanner()
        scan = scanner.scan(tmp_path)
        assert scan.has_knowledge_dir is True

    def test_scan_finds_git_dir(self, tmp_path: Path) -> None:
        (tmp_path / ".git").mkdir()
        scanner = ProjectScanner()
        scan = scanner.scan(tmp_path)
        assert scan.has_git is True



class TestProjectScannerFullProject:
    def test_scan_full_project(self, tmp_path: Path) -> None:
        """A fully-seeded project should have all artifacts detected."""
        # Create docs
        (tmp_path / "docs").mkdir()
        (tmp_path / "docs" / "PRD.md").write_text("# PRD")
        (tmp_path / "docs" / "DDD.md").write_text("# DDD")
        (tmp_path / "docs" / "ARCHITECTURE.md").write_text("# Arch")
        (tmp_path / "AGENTS.md").write_text("# AGENTS")

        # Create configs
        (tmp_path / ".claude").mkdir()
        (tmp_path / ".claude" / "CLAUDE.md").write_text("# Claude")
        (tmp_path / ".beads").mkdir()
        (tmp_path / ".beads" / "issues.jsonl").write_text("{}")
        (tmp_path / "pyproject.toml").write_text("[project]")

        # Create structure
        (tmp_path / "src" / "domain").mkdir(parents=True)
        (tmp_path / "src" / "application").mkdir(parents=True)
        (tmp_path / "src" / "infrastructure").mkdir(parents=True)

        # Create special dirs
        (tmp_path / ".alty" / "knowledge").mkdir(parents=True)
        (tmp_path / ".git").mkdir()
        scanner = ProjectScanner()
        scan = scanner.scan(tmp_path)

        assert len(scan.existing_docs) == 4
        assert len(scan.existing_configs) == 3
        assert len(scan.existing_structure) == 3
        assert scan.has_knowledge_dir is True
        assert scan.has_agents_md is True
        assert scan.has_git is True

    def test_scan_result_is_frozen(self, tmp_path: Path) -> None:
        scanner = ProjectScanner()
        scan = scanner.scan(tmp_path)
        import pytest

        with pytest.raises(AttributeError):
            scan.has_git = True  # type: ignore[misc]
