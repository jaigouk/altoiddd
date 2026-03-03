"""Tests for ProjectScanner profile integration.

Verifies that the scanner reads structure and config targets from
StackProfile instead of hardcoded constants.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.domain.models.stack_profile import GenericProfile, PythonUvProfile
from src.infrastructure.scanner.project_scanner import ProjectScanner

if TYPE_CHECKING:
    from pathlib import Path


# ---------------------------------------------------------------------------
# 1. Scanner uses profile structure targets
# ---------------------------------------------------------------------------


class TestScannerUsesProfileStructure:
    def test_scanner_uses_python_profile_structure(self, tmp_path: Path) -> None:
        """Scanner checks profile.source_layout for structure targets."""
        (tmp_path / "src" / "domain").mkdir(parents=True)
        profile = PythonUvProfile()
        scanner = ProjectScanner()

        scan = scanner.scan(tmp_path, profile=profile)

        assert "src/domain/" in scan.existing_structure

    def test_scanner_generic_profile_empty_structure(self, tmp_path: Path) -> None:
        """GenericProfile (empty source_layout) → no structure targets scanned."""
        (tmp_path / "src" / "domain").mkdir(parents=True)
        profile = GenericProfile()
        scanner = ProjectScanner()

        scan = scanner.scan(tmp_path, profile=profile)

        assert scan.existing_structure == ()


# ---------------------------------------------------------------------------
# 2. Scanner uses profile manifest in config targets
# ---------------------------------------------------------------------------


class TestScannerUsesProfileManifest:
    def test_scanner_includes_python_manifest(self, tmp_path: Path) -> None:
        """Scanner includes profile.project_manifest in config targets."""
        (tmp_path / "pyproject.toml").write_text("[project]")
        profile = PythonUvProfile()
        scanner = ProjectScanner()

        scan = scanner.scan(tmp_path, profile=profile)

        assert "pyproject.toml" in scan.existing_configs

    def test_scanner_generic_profile_no_manifest(self, tmp_path: Path) -> None:
        """GenericProfile (empty manifest) → no manifest in config targets."""
        (tmp_path / "pyproject.toml").write_text("[project]")
        profile = GenericProfile()
        scanner = ProjectScanner()

        scan = scanner.scan(tmp_path, profile=profile)

        # pyproject.toml exists on disk but generic profile doesn't check for it
        assert "pyproject.toml" not in scan.existing_configs


# ---------------------------------------------------------------------------
# 3. Alty-universal configs always checked
# ---------------------------------------------------------------------------


class TestScannerAltyUniversal:
    def test_claude_md_always_checked(self, tmp_path: Path) -> None:
        """Alty-universal configs checked regardless of profile."""
        (tmp_path / ".claude").mkdir()
        (tmp_path / ".claude" / "CLAUDE.md").write_text("# Claude")
        profile = GenericProfile()
        scanner = ProjectScanner()

        scan = scanner.scan(tmp_path, profile=profile)

        assert ".claude/CLAUDE.md" in scan.existing_configs

    def test_beads_always_checked(self, tmp_path: Path) -> None:
        """.beads/issues.jsonl always checked regardless of profile."""
        (tmp_path / ".beads").mkdir()
        (tmp_path / ".beads" / "issues.jsonl").write_text("{}")
        profile = GenericProfile()
        scanner = ProjectScanner()

        scan = scanner.scan(tmp_path, profile=profile)

        assert ".beads/issues.jsonl" in scan.existing_configs


# ---------------------------------------------------------------------------
# 4. Backward compatibility (no profile)
# ---------------------------------------------------------------------------


class TestScannerBackwardCompat:
    def test_no_profile_still_works(self, tmp_path: Path) -> None:
        """Scanner without profile still works (default behavior)."""
        scanner = ProjectScanner()

        scan = scanner.scan(tmp_path)

        assert scan.project_dir == tmp_path

    def test_no_profile_checks_pyproject(self, tmp_path: Path) -> None:
        """Without profile, default behavior checks pyproject.toml."""
        (tmp_path / "pyproject.toml").write_text("[project]")
        scanner = ProjectScanner()

        scan = scanner.scan(tmp_path)

        assert "pyproject.toml" in scan.existing_configs

    def test_no_profile_checks_python_structure(self, tmp_path: Path) -> None:
        """Without profile, default behavior checks Python structure."""
        (tmp_path / "src" / "domain").mkdir(parents=True)
        scanner = ProjectScanner()

        scan = scanner.scan(tmp_path)

        assert "src/domain/" in scan.existing_structure
