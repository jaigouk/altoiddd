"""Tests for KnowledgeDriftDetector adapter.

RED phase: defines the contract for filesystem-based drift detection
that compares current TOML entries against version history.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from pathlib import Path


def _make_toml(path: Path, content: str) -> None:
    """Write a TOML file, creating parent directories."""
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content)


def _make_knowledge_tree(root: Path) -> None:
    """Create a minimal knowledge tree with current and version history."""
    # _meta.toml for claude-code
    _make_toml(
        root / "tools" / "claude-code" / "_meta.toml",
        '[tool]\nname = "claude-code"\n\n'
        "[versions]\n"
        'current = "v2.1"\n'
        'tracked = ["v2.1", "v2.0"]\n',
    )

    # Current entry
    _make_toml(
        root / "tools" / "claude-code" / "current" / "config-structure.toml",
        "[_meta]\n"
        'last_verified = "2026-02-22"\n'
        'verified_against = "v2.1.15"\n'
        'confidence = "high"\n'
        'next_review_date = "2026-05-22"\n\n'
        "[project_structure]\n"
        '"CLAUDE.md" = "Project memory"\n'
        '"rules/*.md" = "Additional rules"\n'
        '"agents/*.md" = "Agent definitions"\n',
    )

    # v2.0 entry (fewer keys)
    _make_toml(
        root / "tools" / "claude-code" / "v2.0" / "config-structure.toml",
        "[_meta]\n"
        'last_verified = "2026-02-22"\n'
        'verified_against = "v2.0.x"\n'
        'confidence = "high"\n'
        'next_review_date = "2026-05-22"\n\n'
        "[project_structure]\n"
        '"CLAUDE.md" = "Project memory"\n',
    )

    # v2.1 entry (same as current)
    _make_toml(
        root / "tools" / "claude-code" / "v2.1" / "config-structure.toml",
        "[_meta]\n"
        'last_verified = "2026-02-22"\n'
        'verified_against = "v2.1.15"\n'
        'confidence = "high"\n'
        'next_review_date = "2026-05-22"\n\n'
        "[project_structure]\n"
        '"CLAUDE.md" = "Project memory"\n'
        '"rules/*.md" = "Additional rules"\n'
        '"agents/*.md" = "Agent definitions"\n',
    )


# ── Version-to-version drift ──────────────────────────────────────


class TestVersionDrift:
    def test_detects_keys_added_in_newer_version(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.knowledge_drift_detector import (
            KnowledgeDriftDetector,
        )

        _make_knowledge_tree(tmp_path)
        detector = KnowledgeDriftDetector(tmp_path)
        report = detector.detect()

        # current has rules/*.md and agents/*.md that v2.0 does not
        version_signals = [s for s in report.signals if s.signal_type.value == "version_change"]
        assert len(version_signals) >= 1
        descriptions = " ".join(s.description for s in version_signals)
        assert "rules" in descriptions.lower() or "agents" in descriptions.lower()

    def test_detects_section_removed_in_current(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.knowledge_drift_detector import (
            KnowledgeDriftDetector,
        )

        _make_toml(
            tmp_path / "tools" / "rm-tool" / "_meta.toml",
            '[tool]\nname = "rm-tool"\n\n'
            "[versions]\n"
            'current = "v2.0"\n'
            'tracked = ["v2.0", "v1.0"]\n',
        )
        # v1.0 has [deprecated_feature] section; current does not
        _make_toml(
            tmp_path / "tools" / "rm-tool" / "v1.0" / "config.toml",
            "[_meta]\n"
            'last_verified = "2026-03-01"\n'
            'next_review_date = "2027-03-01"\n\n'
            "[data]\nkey = 1\n\n"
            "[deprecated_feature]\nold_key = 1\n",
        )
        _make_toml(
            tmp_path / "tools" / "rm-tool" / "current" / "config.toml",
            "[_meta]\n"
            'last_verified = "2026-03-01"\n'
            'next_review_date = "2027-03-01"\n\n'
            "[data]\nkey = 1\n",
        )
        _make_toml(
            tmp_path / "tools" / "rm-tool" / "v2.0" / "config.toml",
            "[_meta]\n"
            'last_verified = "2026-03-01"\n'
            'next_review_date = "2027-03-01"\n\n'
            "[data]\nkey = 1\n",
        )

        detector = KnowledgeDriftDetector(tmp_path)
        report = detector.detect()

        version_signals = [s for s in report.signals if s.signal_type.value == "version_change"]
        assert len(version_signals) >= 1
        assert any("deprecated_feature" in s.description for s in version_signals)

    def test_no_drift_when_current_matches_latest_version(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.knowledge_drift_detector import (
            KnowledgeDriftDetector,
        )

        _make_knowledge_tree(tmp_path)
        # Only compare current vs v2.1 (should be identical)
        # Remove v2.0 so only v2.1 comparison happens
        import shutil

        shutil.rmtree(tmp_path / "tools" / "claude-code" / "v2.0")
        meta = tmp_path / "tools" / "claude-code" / "_meta.toml"
        meta.write_text(
            '[tool]\nname = "claude-code"\n\n[versions]\ncurrent = "v2.1"\ntracked = ["v2.1"]\n',
        )

        detector = KnowledgeDriftDetector(tmp_path)
        report = detector.detect()

        version_signals = [s for s in report.signals if s.signal_type.value == "version_change"]
        assert len(version_signals) == 0


# ── Staleness detection ───────────────────────────────────────────


class TestStalenessDetection:
    def test_detects_stale_entry_past_review_date(self, tmp_path: Path) -> None:
        from src.domain.models.drift_detection import DriftSignalType
        from src.infrastructure.persistence.knowledge_drift_detector import (
            KnowledgeDriftDetector,
        )

        _make_toml(
            tmp_path / "tools" / "stale-tool" / "_meta.toml",
            '[tool]\nname = "stale-tool"\n\n[versions]\ncurrent = "v1.0"\ntracked = ["v1.0"]\n',
        )
        _make_toml(
            tmp_path / "tools" / "stale-tool" / "current" / "config.toml",
            "[_meta]\n"
            'last_verified = "2025-01-01"\n'
            'next_review_date = "2025-06-01"\n'
            'confidence = "high"\n\n'
            "[data]\nkey = 1\n",
        )
        # v1.0 same as current
        _make_toml(
            tmp_path / "tools" / "stale-tool" / "v1.0" / "config.toml",
            "[_meta]\n"
            'last_verified = "2025-01-01"\n'
            'next_review_date = "2025-06-01"\n'
            'confidence = "high"\n\n'
            "[data]\nkey = 1\n",
        )

        detector = KnowledgeDriftDetector(tmp_path)
        report = detector.detect()

        stale_signals = [s for s in report.signals if s.signal_type == DriftSignalType.STALE]
        assert len(stale_signals) >= 1

    def test_no_staleness_when_review_date_in_future(self, tmp_path: Path) -> None:
        from src.domain.models.drift_detection import DriftSignalType
        from src.infrastructure.persistence.knowledge_drift_detector import (
            KnowledgeDriftDetector,
        )

        _make_toml(
            tmp_path / "tools" / "fresh-tool" / "_meta.toml",
            '[tool]\nname = "fresh-tool"\n\n[versions]\ncurrent = "v1.0"\ntracked = ["v1.0"]\n',
        )
        _make_toml(
            tmp_path / "tools" / "fresh-tool" / "current" / "config.toml",
            "[_meta]\n"
            'last_verified = "2026-03-01"\n'
            'next_review_date = "2027-03-01"\n'
            'confidence = "high"\n\n'
            "[data]\nkey = 1\n",
        )
        _make_toml(
            tmp_path / "tools" / "fresh-tool" / "v1.0" / "config.toml",
            "[_meta]\n"
            'last_verified = "2026-03-01"\n'
            'next_review_date = "2027-03-01"\n'
            'confidence = "high"\n\n'
            "[data]\nkey = 1\n",
        )

        detector = KnowledgeDriftDetector(tmp_path)
        report = detector.detect()

        stale_signals = [s for s in report.signals if s.signal_type == DriftSignalType.STALE]
        assert len(stale_signals) == 0


# ── Edge cases ────────────────────────────────────────────────────


class TestEdgeCases:
    def test_empty_knowledge_base(self, tmp_path: Path) -> None:
        """No .alty/knowledge/ directory → empty DriftReport."""
        from src.infrastructure.persistence.knowledge_drift_detector import (
            KnowledgeDriftDetector,
        )

        empty_dir = tmp_path / "empty"
        empty_dir.mkdir()
        detector = KnowledgeDriftDetector(empty_dir)
        report = detector.detect()
        assert report.total_count == 0
        assert report.has_drift is False

    def test_no_version_history(self, tmp_path: Path) -> None:
        """Tool entry exists but no version history → info signal."""
        from src.domain.models.drift_detection import DriftSeverity
        from src.infrastructure.persistence.knowledge_drift_detector import (
            KnowledgeDriftDetector,
        )

        _make_toml(
            tmp_path / "tools" / "new-tool" / "_meta.toml",
            '[tool]\nname = "new-tool"\n\n[versions]\ncurrent = "v1.0"\ntracked = []\n',
        )
        _make_toml(
            tmp_path / "tools" / "new-tool" / "current" / "config.toml",
            "[_meta]\n"
            'last_verified = "2026-03-01"\n'
            'next_review_date = "2027-03-01"\n\n'
            "[data]\nkey = 1\n",
        )

        detector = KnowledgeDriftDetector(tmp_path)
        report = detector.detect()

        info_signals = [s for s in report.signals if s.severity == DriftSeverity.INFO]
        assert len(info_signals) >= 1
        assert any("no version history" in s.description.lower() for s in info_signals)

    def test_missing_meta_toml_skips_tool(self, tmp_path: Path) -> None:
        """Tool directory without _meta.toml → skip gracefully."""
        from src.infrastructure.persistence.knowledge_drift_detector import (
            KnowledgeDriftDetector,
        )

        (tmp_path / "tools" / "bad-tool" / "current").mkdir(parents=True)
        _make_toml(
            tmp_path / "tools" / "bad-tool" / "current" / "config.toml",
            "[data]\nkey = 1\n",
        )
        # No _meta.toml

        detector = KnowledgeDriftDetector(tmp_path)
        report = detector.detect()  # Should not raise
        assert isinstance(report.total_count, int)

    def test_multiple_drift_signals_per_entry(self, tmp_path: Path) -> None:
        """Entry with version change + staleness → reports both signals."""
        from src.domain.models.drift_detection import DriftSignalType
        from src.infrastructure.persistence.knowledge_drift_detector import (
            KnowledgeDriftDetector,
        )

        _make_toml(
            tmp_path / "tools" / "multi-tool" / "_meta.toml",
            '[tool]\nname = "multi-tool"\n\n'
            "[versions]\n"
            'current = "v2.0"\n'
            'tracked = ["v2.0", "v1.0"]\n',
        )
        # Current: has extra key AND is stale
        _make_toml(
            tmp_path / "tools" / "multi-tool" / "current" / "config.toml",
            "[_meta]\n"
            'last_verified = "2024-01-01"\n'
            'next_review_date = "2024-06-01"\n\n'
            "[data]\nold_key = 1\nnew_key = 2\n",
        )
        # v1.0: fewer keys
        _make_toml(
            tmp_path / "tools" / "multi-tool" / "v1.0" / "config.toml",
            "[_meta]\n"
            'last_verified = "2024-01-01"\n'
            'next_review_date = "2024-06-01"\n\n'
            "[data]\nold_key = 1\n",
        )

        detector = KnowledgeDriftDetector(tmp_path)
        report = detector.detect()

        types = {s.signal_type for s in report.signals}
        assert DriftSignalType.VERSION_CHANGE in types
        assert DriftSignalType.STALE in types

    def test_version_entry_missing_from_disk(self, tmp_path: Path) -> None:
        """_meta.toml lists v1.0 but no v1.0 directory → warning signal."""
        from src.domain.models.drift_detection import DriftSeverity
        from src.infrastructure.persistence.knowledge_drift_detector import (
            KnowledgeDriftDetector,
        )

        _make_toml(
            tmp_path / "tools" / "gap-tool" / "_meta.toml",
            '[tool]\nname = "gap-tool"\n\n'
            "[versions]\n"
            'current = "v2.0"\n'
            'tracked = ["v2.0", "v1.0"]\n',
        )
        _make_toml(
            tmp_path / "tools" / "gap-tool" / "current" / "config.toml",
            "[_meta]\n"
            'last_verified = "2026-03-01"\n'
            'next_review_date = "2027-03-01"\n\n'
            "[data]\nkey = 1\n",
        )
        # v2.0 exists
        _make_toml(
            tmp_path / "tools" / "gap-tool" / "v2.0" / "config.toml",
            "[_meta]\n"
            'last_verified = "2026-03-01"\n'
            'next_review_date = "2027-03-01"\n\n'
            "[data]\nkey = 1\n",
        )
        # v1.0 directory does NOT exist

        detector = KnowledgeDriftDetector(tmp_path)
        report = detector.detect()

        warn_signals = [s for s in report.signals if s.severity == DriftSeverity.WARNING]
        assert any("v1.0" in s.description for s in warn_signals)

    def test_implements_drift_detection_port(self) -> None:
        """KnowledgeDriftDetector satisfies DriftDetectionPort protocol."""
        from pathlib import Path

        from src.application.ports.drift_detection_port import DriftDetectionPort
        from src.infrastructure.persistence.knowledge_drift_detector import (
            KnowledgeDriftDetector,
        )

        detector = KnowledgeDriftDetector(Path("/nonexistent"))
        assert isinstance(detector, DriftDetectionPort)

    def test_cross_tool_entries_not_scanned(self, tmp_path: Path) -> None:
        """Only tools/ entries are scanned for version drift."""
        from src.infrastructure.persistence.knowledge_drift_detector import (
            KnowledgeDriftDetector,
        )

        _make_toml(
            tmp_path / "cross-tool" / "concept-mapping.toml",
            "[_meta]\n"
            'last_verified = "2024-01-01"\n'
            'next_review_date = "2024-06-01"\n\n'
            "[data]\nkey = 1\n",
        )

        detector = KnowledgeDriftDetector(tmp_path)
        report = detector.detect()
        # cross-tool has no version history mechanism, so no version_change signals
        version_signals = [s for s in report.signals if s.signal_type.value == "version_change"]
        assert len(version_signals) == 0
