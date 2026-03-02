"""KnowledgeDriftDetector -- filesystem-based drift detection adapter.

Implements DriftDetectionPort by scanning .alty/knowledge/tools/ entries,
comparing current TOML against version history, and checking staleness.
"""

from __future__ import annotations

import tomllib
from datetime import date
from typing import TYPE_CHECKING

from src.domain.models.drift_detection import (
    DriftReport,
    DriftSeverity,
    DriftSignal,
    DriftSignalType,
)

if TYPE_CHECKING:
    from pathlib import Path


class KnowledgeDriftDetector:
    """Scans knowledge entries for drift signals.

    Drift strategies:
    1. Version-to-version diff: compare current TOML sections/keys against
       previous version entries. Flag added/removed/changed keys.
    2. Staleness: flag entries whose next_review_date has passed.
    """

    def __init__(self, knowledge_dir: Path) -> None:
        self._root = knowledge_dir

    def detect(self) -> DriftReport:
        """Detect drift across all tool knowledge entries."""
        signals: list[DriftSignal] = []
        tools_dir = self._root / "tools"

        if not tools_dir.exists():
            return DriftReport(signals=())

        for tool_dir in sorted(tools_dir.iterdir()):
            if not tool_dir.is_dir():
                continue
            signals.extend(self._scan_tool(tool_dir))

        return DriftReport(signals=tuple(signals))

    def _scan_tool(self, tool_dir: Path) -> list[DriftSignal]:
        """Scan a single tool directory for drift."""
        signals: list[DriftSignal] = []
        meta_path = tool_dir / "_meta.toml"

        if not meta_path.exists():
            return signals

        meta = self._read_toml(meta_path)
        versions_section = meta.get("versions", {})
        if not isinstance(versions_section, dict):
            return signals

        tracked = versions_section.get("tracked", [])
        if not isinstance(tracked, list):
            tracked = []

        current_dir = tool_dir / "current"
        if not current_dir.exists():
            return signals

        tool_name = tool_dir.name

        # Scan each current entry
        for entry_path in sorted(current_dir.glob("*.toml")):
            entry_name = entry_path.stem
            rlm_path = f"tools/{tool_name}/{entry_name}"
            current_data = self._read_toml(entry_path)

            # Staleness check
            signals.extend(self._check_staleness(rlm_path, current_data))

            # Version comparison
            if not tracked:
                signals.append(
                    DriftSignal(
                        entry_path=rlm_path,
                        signal_type=DriftSignalType.VERSION_CHANGE,
                        description=(
                            f"No version history for {tool_name} — cannot compare for drift"
                        ),
                        severity=DriftSeverity.INFO,
                    )
                )
            else:
                signals.extend(
                    self._check_version_drift(
                        tool_dir, tool_name, entry_name, current_data, tracked
                    )
                )

        return signals

    def _check_staleness(self, rlm_path: str, data: dict[str, object]) -> list[DriftSignal]:
        """Check if an entry is past its next_review_date."""
        meta = data.get("_meta", {})
        if not isinstance(meta, dict):
            return []

        review_date_str = meta.get("next_review_date")
        if not review_date_str or not isinstance(review_date_str, str):
            return []

        try:
            review_date = date.fromisoformat(review_date_str)
        except ValueError:
            return []

        if date.today() > review_date:
            return [
                DriftSignal(
                    entry_path=rlm_path,
                    signal_type=DriftSignalType.STALE,
                    description=(
                        f"Entry past review date ({review_date_str}), "
                        f"last verified {meta.get('last_verified', 'unknown')}"
                    ),
                    severity=DriftSeverity.WARNING,
                )
            ]
        return []

    def _check_version_drift(
        self,
        tool_dir: Path,
        tool_name: str,
        entry_name: str,
        current_data: dict[str, object],
        tracked: list[object],
    ) -> list[DriftSignal]:
        """Compare current entry against each tracked version."""
        signals: list[DriftSignal] = []
        rlm_path = f"tools/{tool_name}/{entry_name}"

        for version in tracked:
            version_str = str(version)
            version_dir = tool_dir / version_str
            if not version_dir.exists():
                signals.append(
                    DriftSignal(
                        entry_path=rlm_path,
                        signal_type=DriftSignalType.VERSION_CHANGE,
                        description=(
                            f"Version {version_str} listed in _meta.toml but directory not found"
                        ),
                        severity=DriftSeverity.WARNING,
                    )
                )
                continue

            version_entry = version_dir / f"{entry_name}.toml"
            if not version_entry.exists():
                continue

            version_data = self._read_toml(version_entry)
            signals.extend(self._diff_entries(rlm_path, version_str, current_data, version_data))

        return signals

    def _diff_entries(
        self,
        rlm_path: str,
        version: str,
        current: dict[str, object],
        previous: dict[str, object],
    ) -> list[DriftSignal]:
        """Diff two TOML entries, ignoring _meta sections."""
        signals: list[DriftSignal] = []

        for section_name in current:
            if section_name == "_meta":
                continue
            current_section = current[section_name]
            previous_section = previous.get(section_name)

            if not isinstance(current_section, dict):
                continue

            if previous_section is None:
                signals.append(
                    DriftSignal(
                        entry_path=rlm_path,
                        signal_type=DriftSignalType.VERSION_CHANGE,
                        description=(
                            f"Section [{section_name}] added in current but missing in {version}"
                        ),
                        severity=DriftSeverity.WARNING,
                    )
                )
                continue

            if not isinstance(previous_section, dict):
                continue

            # Find added keys
            added = set(current_section) - set(previous_section)
            signals.extend(
                DriftSignal(
                    entry_path=rlm_path,
                    signal_type=DriftSignalType.VERSION_CHANGE,
                    description=(
                        f"Key '{key}' in [{section_name}] added in "
                        f"current but missing in {version}"
                    ),
                    severity=DriftSeverity.WARNING,
                )
                for key in sorted(added)
            )

            # Find removed keys
            removed = set(previous_section) - set(current_section)
            signals.extend(
                DriftSignal(
                    entry_path=rlm_path,
                    signal_type=DriftSignalType.VERSION_CHANGE,
                    description=(
                        f"Key '{key}' in [{section_name}] present in "
                        f"{version} but removed in current"
                    ),
                    severity=DriftSeverity.WARNING,
                )
                for key in sorted(removed)
            )

        # Detect sections removed from current (only in previous)
        for section_name in previous:
            if section_name == "_meta":
                continue
            if section_name not in current and isinstance(previous[section_name], dict):
                signals.append(
                    DriftSignal(
                        entry_path=rlm_path,
                        signal_type=DriftSignalType.VERSION_CHANGE,
                        description=(
                            f"Section [{section_name}] present in {version} but removed in current"
                        ),
                        severity=DriftSeverity.WARNING,
                    )
                )

        return signals

    @staticmethod
    def _read_toml(path: Path) -> dict[str, object]:
        """Read and parse a TOML file, returning empty dict on failure."""
        try:
            return tomllib.loads(path.read_text())
        except (tomllib.TOMLDecodeError, OSError):
            return {}
