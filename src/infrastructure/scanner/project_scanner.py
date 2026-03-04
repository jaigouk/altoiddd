"""ProjectScanner -- infrastructure adapter for ProjectScanPort.

Scans an existing project directory and returns a frozen ProjectScan
value object describing its current state.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.domain.models.gap_analysis import ProjectScan

if TYPE_CHECKING:
    from pathlib import Path

    from src.domain.models.stack_profile import StackProfile


# Documentation files to check for
_DOC_TARGETS: tuple[str, ...] = (
    "docs/PRD.md",
    "docs/DDD.md",
    "docs/ARCHITECTURE.md",
    "AGENTS.md",
)

# Alty-universal configuration files (always checked)
_ALTY_CONFIG_TARGETS: tuple[str, ...] = (
    ".claude/CLAUDE.md",
    ".beads/issues.jsonl",
)

# Default structure targets (used when no profile provided)
_DEFAULT_STRUCTURE_TARGETS: tuple[str, ...] = (
    "src/domain/",
    "src/application/",
    "src/infrastructure/",
)

# Default manifest (used when no profile provided)
_DEFAULT_MANIFEST: str = "pyproject.toml"


class ProjectScanner:
    """Filesystem-based implementation of ProjectScanPort.

    Scans for known documentation, configuration, and directory structure
    within a project directory.
    """

    def scan(
        self,
        project_dir: Path,
        profile: StackProfile | None = None,
    ) -> ProjectScan:
        """Scan a project directory and return a frozen snapshot.

        Args:
            project_dir: The project directory to scan.
            profile: Stack profile providing structure and manifest targets.
                When None, uses default Python targets for backward compat.

        Returns:
            A ProjectScan value object describing the current state.
        """
        existing_docs = [
            doc_path for doc_path in _DOC_TARGETS if (project_dir / doc_path).exists()
        ]

        # Build config targets: alty-universal + profile manifest
        manifest = profile.project_manifest if profile is not None else _DEFAULT_MANIFEST
        config_targets = list(_ALTY_CONFIG_TARGETS)
        if manifest:
            config_targets.append(manifest)

        existing_configs = [
            config_path for config_path in config_targets if (project_dir / config_path).exists()
        ]

        # Build structure targets from profile
        structure_targets = (
            profile.source_layout if profile is not None else _DEFAULT_STRUCTURE_TARGETS
        )

        existing_structure = [
            structure_path
            for structure_path in structure_targets
            if (project_dir / structure_path).is_dir()
        ]

        has_knowledge_dir = (project_dir / ".alty" / "knowledge").is_dir()
        has_agents_md = (project_dir / "AGENTS.md").is_file()
        has_git = (project_dir / ".git").exists()
        has_alty_config = (project_dir / ".alty" / "config.toml").is_file()
        has_maintenance_dir = (project_dir / ".alty" / "maintenance").is_dir()

        return ProjectScan(
            project_dir=project_dir,
            existing_docs=tuple(existing_docs),
            existing_configs=tuple(existing_configs),
            existing_structure=tuple(existing_structure),
            has_knowledge_dir=has_knowledge_dir,
            has_agents_md=has_agents_md,
            has_git=has_git,
            has_alty_config=has_alty_config,
            has_maintenance_dir=has_maintenance_dir,
        )
