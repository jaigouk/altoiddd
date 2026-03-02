"""Tests for DriftSignal and DriftReport value objects.

RED phase: defines the domain model contract for knowledge base drift
detection (Knowledge Base bounded context, alty-djb).
"""

from __future__ import annotations

import pytest

from src.domain.models.errors import InvariantViolationError

# ── DriftSignalType ────────────────────────────────────────────────


class TestDriftSignalType:
    def test_has_version_change(self) -> None:
        from src.domain.models.drift_detection import DriftSignalType

        assert DriftSignalType.VERSION_CHANGE.value == "version_change"

    def test_has_doc_code_mismatch(self) -> None:
        from src.domain.models.drift_detection import DriftSignalType

        assert DriftSignalType.DOC_CODE_MISMATCH.value == "doc_code_mismatch"

    def test_has_stale(self) -> None:
        from src.domain.models.drift_detection import DriftSignalType

        assert DriftSignalType.STALE.value == "stale"


# ── DriftSeverity ──────────────────────────────────────────────────


class TestDriftSeverity:
    def test_has_info(self) -> None:
        from src.domain.models.drift_detection import DriftSeverity

        assert DriftSeverity.INFO.value == "info"

    def test_has_warning(self) -> None:
        from src.domain.models.drift_detection import DriftSeverity

        assert DriftSeverity.WARNING.value == "warning"

    def test_has_error(self) -> None:
        from src.domain.models.drift_detection import DriftSeverity

        assert DriftSeverity.ERROR.value == "error"


# ── DriftSignal ────────────────────────────────────────────────────


class TestDriftSignal:
    def test_creates_with_all_fields(self) -> None:
        from src.domain.models.drift_detection import (
            DriftSeverity,
            DriftSignal,
            DriftSignalType,
        )

        signal = DriftSignal(
            entry_path="tools/claude-code/config-structure",
            signal_type=DriftSignalType.VERSION_CHANGE,
            description="Key 'rules/*.md' added in current but missing in v2.0",
            severity=DriftSeverity.WARNING,
        )
        assert signal.entry_path == "tools/claude-code/config-structure"
        assert signal.signal_type == DriftSignalType.VERSION_CHANGE
        assert signal.severity == DriftSeverity.WARNING

    def test_is_frozen(self) -> None:
        from src.domain.models.drift_detection import (
            DriftSeverity,
            DriftSignal,
            DriftSignalType,
        )

        signal = DriftSignal(
            entry_path="tools/cursor/rules-format",
            signal_type=DriftSignalType.STALE,
            description="Entry not reviewed in 120 days",
            severity=DriftSeverity.INFO,
        )
        with pytest.raises(AttributeError):
            signal.entry_path = "changed"  # type: ignore[misc]

    def test_empty_entry_path_raises(self) -> None:
        from src.domain.models.drift_detection import (
            DriftSeverity,
            DriftSignal,
            DriftSignalType,
        )

        with pytest.raises(InvariantViolationError):
            DriftSignal(
                entry_path="",
                signal_type=DriftSignalType.VERSION_CHANGE,
                description="Something changed",
                severity=DriftSeverity.WARNING,
            )

    def test_whitespace_entry_path_raises(self) -> None:
        from src.domain.models.drift_detection import (
            DriftSeverity,
            DriftSignal,
            DriftSignalType,
        )

        with pytest.raises(InvariantViolationError):
            DriftSignal(
                entry_path="   ",
                signal_type=DriftSignalType.VERSION_CHANGE,
                description="Something changed",
                severity=DriftSeverity.WARNING,
            )

    def test_empty_description_raises(self) -> None:
        from src.domain.models.drift_detection import (
            DriftSeverity,
            DriftSignal,
            DriftSignalType,
        )

        with pytest.raises(InvariantViolationError):
            DriftSignal(
                entry_path="tools/cursor/rules-format",
                signal_type=DriftSignalType.STALE,
                description="",
                severity=DriftSeverity.INFO,
            )

    def test_equality_by_value(self) -> None:
        from src.domain.models.drift_detection import (
            DriftSeverity,
            DriftSignal,
            DriftSignalType,
        )

        a = DriftSignal(
            entry_path="tools/cursor/rules-format",
            signal_type=DriftSignalType.STALE,
            description="Stale entry",
            severity=DriftSeverity.INFO,
        )
        b = DriftSignal(
            entry_path="tools/cursor/rules-format",
            signal_type=DriftSignalType.STALE,
            description="Stale entry",
            severity=DriftSeverity.INFO,
        )
        assert a == b

    def test_different_signals_not_equal(self) -> None:
        from src.domain.models.drift_detection import (
            DriftSeverity,
            DriftSignal,
            DriftSignalType,
        )

        a = DriftSignal(
            entry_path="tools/cursor/rules-format",
            signal_type=DriftSignalType.STALE,
            description="Stale entry",
            severity=DriftSeverity.INFO,
        )
        b = DriftSignal(
            entry_path="tools/cursor/rules-format",
            signal_type=DriftSignalType.VERSION_CHANGE,
            description="Changed entry",
            severity=DriftSeverity.WARNING,
        )
        assert a != b


# ── DriftReport ────────────────────────────────────────────────────


class TestDriftReport:
    def test_creates_with_signals(self) -> None:
        from src.domain.models.drift_detection import (
            DriftReport,
            DriftSeverity,
            DriftSignal,
            DriftSignalType,
        )

        signals = (
            DriftSignal(
                entry_path="tools/claude-code/config-structure",
                signal_type=DriftSignalType.VERSION_CHANGE,
                description="Key added",
                severity=DriftSeverity.WARNING,
            ),
        )
        report = DriftReport(signals=signals)
        assert report.total_count == 1

    def test_is_frozen(self) -> None:
        from src.domain.models.drift_detection import DriftReport

        report = DriftReport(signals=())
        with pytest.raises(AttributeError):
            report.signals = ()  # type: ignore[misc]

    def test_empty_report(self) -> None:
        from src.domain.models.drift_detection import DriftReport

        report = DriftReport(signals=())
        assert report.total_count == 0
        assert report.has_drift is False

    def test_has_drift_true(self) -> None:
        from src.domain.models.drift_detection import (
            DriftReport,
            DriftSeverity,
            DriftSignal,
            DriftSignalType,
        )

        report = DriftReport(
            signals=(
                DriftSignal(
                    entry_path="tools/cursor/rules-format",
                    signal_type=DriftSignalType.STALE,
                    description="Stale",
                    severity=DriftSeverity.INFO,
                ),
            )
        )
        assert report.has_drift is True

    def test_count_by_severity(self) -> None:
        from src.domain.models.drift_detection import (
            DriftReport,
            DriftSeverity,
            DriftSignal,
            DriftSignalType,
        )

        report = DriftReport(
            signals=(
                DriftSignal(
                    entry_path="a",
                    signal_type=DriftSignalType.STALE,
                    description="Stale 1",
                    severity=DriftSeverity.INFO,
                ),
                DriftSignal(
                    entry_path="b",
                    signal_type=DriftSignalType.VERSION_CHANGE,
                    description="Changed",
                    severity=DriftSeverity.WARNING,
                ),
                DriftSignal(
                    entry_path="c",
                    signal_type=DriftSignalType.DOC_CODE_MISMATCH,
                    description="Mismatch",
                    severity=DriftSeverity.ERROR,
                ),
                DriftSignal(
                    entry_path="d",
                    signal_type=DriftSignalType.STALE,
                    description="Stale 2",
                    severity=DriftSeverity.INFO,
                ),
            )
        )
        assert report.count_by_severity(DriftSeverity.INFO) == 2
        assert report.count_by_severity(DriftSeverity.WARNING) == 1
        assert report.count_by_severity(DriftSeverity.ERROR) == 1

    def test_count_by_type(self) -> None:
        from src.domain.models.drift_detection import (
            DriftReport,
            DriftSeverity,
            DriftSignal,
            DriftSignalType,
        )

        report = DriftReport(
            signals=(
                DriftSignal(
                    entry_path="a",
                    signal_type=DriftSignalType.STALE,
                    description="Stale",
                    severity=DriftSeverity.INFO,
                ),
                DriftSignal(
                    entry_path="b",
                    signal_type=DriftSignalType.VERSION_CHANGE,
                    description="Changed 1",
                    severity=DriftSeverity.WARNING,
                ),
                DriftSignal(
                    entry_path="c",
                    signal_type=DriftSignalType.VERSION_CHANGE,
                    description="Changed 2",
                    severity=DriftSeverity.WARNING,
                ),
            )
        )
        assert report.count_by_type(DriftSignalType.VERSION_CHANGE) == 2
        assert report.count_by_type(DriftSignalType.STALE) == 1
        assert report.count_by_type(DriftSignalType.DOC_CODE_MISMATCH) == 0
