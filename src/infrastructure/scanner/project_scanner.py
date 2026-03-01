"""ProjectScanner -- infrastructure adapter for ProjectScanPort.

Scans an existing project directory and returns a frozen ProjectScan
value object describing its current state.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.domain.models.gap_analysis import ProjectScan

if TYPE_CHECKING:
    from pathlib import Path


# Documentation files to check for
_DOC_TARGETS: tuple[str, ...] = (
    "docs/PRD.md",
    "docs/DDD.md",
    "docs/ARCHITECTURE.md",
    "AGENTS.md",
)

# Configuration files to check for
_CONFIG_TARGETS: tuple[str, ...] = (
    ".claude/CLAUDE.md",
    ".beads/issues.jsonl",
    "pyproject.toml",
)

# Directory structure to check for
_STRUCTURE_TARGETS: tuple[str, ...] = (
    "src/domain/",
    "src/application/",
    "src/infrastructure/",
)


class ProjectScanner:
    """Filesystem-based implementation of ProjectScanPort.

    Scans for known documentation, configuration, and directory structure
    within a project directory.
    """

    def scan(self, project_dir: Path) -> ProjectScan:
        """Scan a project directory and return a frozen snapshot.

        Args:
            project_dir: The project directory to scan.

        Returns:
            A ProjectScan value object describing the current state.
        """
        existing_docs = [
            doc_path for doc_path in _DOC_TARGETS if (project_dir / doc_path).exists()
        ]

        existing_configs = [
            config_path for config_path in _CONFIG_TARGETS if (project_dir / config_path).exists()
        ]

        existing_structure = [
            structure_path
            for structure_path in _STRUCTURE_TARGETS
            if (project_dir / structure_path).is_dir()
        ]

        has_knowledge_dir = (project_dir / ".alty" / "knowledge").is_dir()
        has_agents_md = (project_dir / "AGENTS.md").is_file()
        has_git = (project_dir / ".git").exists()
        has_tests = (project_dir / "tests").is_dir()

        return ProjectScan(
            project_dir=project_dir,
            existing_docs=tuple(existing_docs),
            existing_configs=tuple(existing_configs),
            existing_structure=tuple(existing_structure),
            has_knowledge_dir=has_knowledge_dir,
            has_agents_md=has_agents_md,
            has_git=has_git,
            has_tests=has_tests,
        )
