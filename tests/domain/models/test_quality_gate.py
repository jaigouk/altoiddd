"""Tests for QualityGate, GateResult, and QualityReport value objects.

Covers enum membership, frozen dataclass invariants, and report
pass/fail aggregation logic.
"""

from __future__ import annotations

import pytest

from src.domain.models.quality_gate import GateResult, QualityGate, QualityReport


class TestQualityGateEnum:
    def test_has_four_members(self):
        assert len(QualityGate) == 4

    def test_member_values(self):
        assert QualityGate.LINT.value == "lint"
        assert QualityGate.TYPES.value == "types"
        assert QualityGate.TESTS.value == "tests"
        assert QualityGate.FITNESS.value == "fitness"


class TestGateResult:
    def test_is_frozen(self):
        result = GateResult(
            gate=QualityGate.LINT,
            passed=True,
            output="ok",
            duration_ms=42,
        )
        with pytest.raises(AttributeError):
            result.passed = False  # type: ignore[misc]

    def test_fields_accessible(self):
        result = GateResult(
            gate=QualityGate.TYPES,
            passed=False,
            output="error on line 5",
            duration_ms=100,
        )
        assert result.gate == QualityGate.TYPES
        assert result.passed is False
        assert result.output == "error on line 5"
        assert result.duration_ms == 100


class TestQualityReport:
    def test_passed_true_when_all_pass(self):
        results = (
            GateResult(gate=QualityGate.LINT, passed=True, output="", duration_ms=10),
            GateResult(gate=QualityGate.TYPES, passed=True, output="", duration_ms=20),
        )
        report = QualityReport(results=results)
        assert report.passed is True

    def test_passed_false_when_any_fails(self):
        results = (
            GateResult(gate=QualityGate.LINT, passed=True, output="", duration_ms=10),
            GateResult(gate=QualityGate.TESTS, passed=False, output="1 failed", duration_ms=50),
        )
        report = QualityReport(results=results)
        assert report.passed is False

    def test_passed_true_for_empty_results(self):
        report = QualityReport(results=())
        assert report.passed is True

    def test_is_frozen(self):
        report = QualityReport(results=())
        with pytest.raises(AttributeError):
            report.results = ()  # type: ignore[misc]
