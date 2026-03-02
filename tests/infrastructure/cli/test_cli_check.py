"""Tests for the alty check CLI command.

Verifies that `alty check` wires to QualityGateHandler via composition root,
formats the QualityReport correctly, and returns appropriate exit codes.
"""

from __future__ import annotations

from unittest.mock import patch

from typer.testing import CliRunner

from src.domain.models.quality_gate import GateResult, QualityGate, QualityReport
from src.infrastructure.cli.main import app

runner = CliRunner()


def _make_result(gate: QualityGate, *, passed: bool, duration_ms: int = 100) -> GateResult:
    """Helper to create a GateResult with sensible defaults."""
    output = "All checks passed!" if passed else f"{gate.value} failed"
    return GateResult(gate=gate, passed=passed, output=output, duration_ms=duration_ms)


def _all_pass_report() -> QualityReport:
    return QualityReport(
        results=tuple(_make_result(g, passed=True) for g in QualityGate),
    )


def _one_fail_report() -> QualityReport:
    results = [
        _make_result(QualityGate.LINT, passed=True),
        _make_result(QualityGate.TYPES, passed=False),
        _make_result(QualityGate.TESTS, passed=True),
        _make_result(QualityGate.FITNESS, passed=True),
    ]
    return QualityReport(results=tuple(results))


def _all_fail_report() -> QualityReport:
    return QualityReport(
        results=tuple(_make_result(g, passed=False) for g in QualityGate),
    )


# ---------------------------------------------------------------------------
# Exit codes
# ---------------------------------------------------------------------------


class TestCheckExitCodes:
    """alty check returns 0 on all-pass, 1 on any failure."""

    @patch("src.infrastructure.composition.create_app")
    def test_exit_zero_when_all_pass(self, mock_create_app):
        mock_create_app.return_value.quality_gate.check.return_value = _all_pass_report()
        result = runner.invoke(app, ["check"])
        assert result.exit_code == 0

    @patch("src.infrastructure.composition.create_app")
    def test_exit_one_when_any_fail(self, mock_create_app):
        mock_create_app.return_value.quality_gate.check.return_value = _one_fail_report()
        result = runner.invoke(app, ["check"])
        assert result.exit_code == 1

    @patch("src.infrastructure.composition.create_app")
    def test_exit_one_when_all_fail(self, mock_create_app):
        mock_create_app.return_value.quality_gate.check.return_value = _all_fail_report()
        result = runner.invoke(app, ["check"])
        assert result.exit_code == 1


# ---------------------------------------------------------------------------
# Output formatting
# ---------------------------------------------------------------------------


class TestCheckOutputFormatting:
    """alty check displays a formatted report with per-gate results."""

    @patch("src.infrastructure.composition.create_app")
    def test_displays_each_gate_name(self, mock_create_app):
        mock_create_app.return_value.quality_gate.check.return_value = _all_pass_report()
        result = runner.invoke(app, ["check"])
        for gate in QualityGate:
            assert gate.value in result.output

    @patch("src.infrastructure.composition.create_app")
    def test_displays_pass_indicator(self, mock_create_app):
        mock_create_app.return_value.quality_gate.check.return_value = _all_pass_report()
        result = runner.invoke(app, ["check"])
        assert "pass" in result.output.lower() or "PASS" in result.output

    @patch("src.infrastructure.composition.create_app")
    def test_displays_fail_indicator(self, mock_create_app):
        mock_create_app.return_value.quality_gate.check.return_value = _one_fail_report()
        result = runner.invoke(app, ["check"])
        assert "fail" in result.output.lower() or "FAIL" in result.output

    @patch("src.infrastructure.composition.create_app")
    def test_displays_duration(self, mock_create_app):
        report = QualityReport(
            results=(
                _make_result(QualityGate.LINT, passed=True, duration_ms=1234),
                _make_result(QualityGate.TYPES, passed=True, duration_ms=567),
                _make_result(QualityGate.TESTS, passed=True, duration_ms=890),
                _make_result(QualityGate.FITNESS, passed=True, duration_ms=0),
            ),
        )
        mock_create_app.return_value.quality_gate.check.return_value = report
        result = runner.invoke(app, ["check"])
        # Duration should appear somewhere in the output
        assert "1234" in result.output or "1.2" in result.output

    @patch("src.infrastructure.composition.create_app")
    def test_displays_summary_on_all_pass(self, mock_create_app):
        mock_create_app.return_value.quality_gate.check.return_value = _all_pass_report()
        result = runner.invoke(app, ["check"])
        output_lower = result.output.lower()
        assert "pass" in output_lower

    @patch("src.infrastructure.composition.create_app")
    def test_displays_summary_on_failure(self, mock_create_app):
        mock_create_app.return_value.quality_gate.check.return_value = _one_fail_report()
        result = runner.invoke(app, ["check"])
        output_lower = result.output.lower()
        assert "fail" in output_lower


# ---------------------------------------------------------------------------
# Handler wiring
# ---------------------------------------------------------------------------


class TestCheckWiring:
    """alty check uses composition root, not standalone construction."""

    @patch("src.infrastructure.composition.create_app")
    def test_calls_create_app(self, mock_create_app):
        mock_create_app.return_value.quality_gate.check.return_value = _all_pass_report()
        runner.invoke(app, ["check"])
        mock_create_app.assert_called_once()

    @patch("src.infrastructure.composition.create_app")
    def test_calls_quality_gate_check(self, mock_create_app):
        mock_create_app.return_value.quality_gate.check.return_value = _all_pass_report()
        runner.invoke(app, ["check"])
        mock_create_app.return_value.quality_gate.check.assert_called_once()

    @patch("src.infrastructure.composition.create_app")
    def test_default_runs_all_gates(self, mock_create_app):
        """Without --gate flag, check() should pass gates=None (all gates)."""
        mock_create_app.return_value.quality_gate.check.return_value = _all_pass_report()
        runner.invoke(app, ["check"])
        mock_create_app.return_value.quality_gate.check.assert_called_once_with(gates=None)


# ---------------------------------------------------------------------------
# --gate flag
# ---------------------------------------------------------------------------


class TestCheckGateFlag:
    """alty check --gate <name> runs only the specified gate."""

    @patch("src.infrastructure.composition.create_app")
    def test_gate_lint_runs_only_lint(self, mock_create_app):
        report = QualityReport(results=(_make_result(QualityGate.LINT, passed=True),))
        mock_create_app.return_value.quality_gate.check.return_value = report
        result = runner.invoke(app, ["check", "--gate", "lint"])
        assert result.exit_code == 0
        mock_create_app.return_value.quality_gate.check.assert_called_once_with(
            gates=(QualityGate.LINT,),
        )

    @patch("src.infrastructure.composition.create_app")
    def test_gate_types_runs_only_types(self, mock_create_app):
        report = QualityReport(results=(_make_result(QualityGate.TYPES, passed=True),))
        mock_create_app.return_value.quality_gate.check.return_value = report
        result = runner.invoke(app, ["check", "--gate", "types"])
        assert result.exit_code == 0
        mock_create_app.return_value.quality_gate.check.assert_called_once_with(
            gates=(QualityGate.TYPES,),
        )

    @patch("src.infrastructure.composition.create_app")
    def test_invalid_gate_exits_with_error(self, mock_create_app):
        result = runner.invoke(app, ["check", "--gate", "bogus"])
        assert result.exit_code == 1
        assert "bogus" in result.output.lower() or "invalid" in result.output.lower()


# ---------------------------------------------------------------------------
# Edge cases
# ---------------------------------------------------------------------------


class TestCheckEdgeCases:
    """Edge cases: skipped fitness, mixed results, no output."""

    @patch("src.infrastructure.composition.create_app")
    def test_skipped_fitness_shows_in_output(self, mock_create_app):
        """Fitness gate can be 'skipped' — still passed=True, output says 'Skipped'."""
        results = (
            _make_result(QualityGate.LINT, passed=True),
            _make_result(QualityGate.TYPES, passed=True),
            _make_result(QualityGate.TESTS, passed=True),
            GateResult(
                gate=QualityGate.FITNESS,
                passed=True,
                output="Skipped: tests/architecture/ directory not found",
                duration_ms=0,
            ),
        )
        mock_create_app.return_value.quality_gate.check.return_value = QualityReport(
            results=results,
        )
        result = runner.invoke(app, ["check"])
        assert result.exit_code == 0
        assert "fitness" in result.output.lower()

    @patch("src.infrastructure.composition.create_app")
    def test_mixed_pass_fail_shows_both(self, mock_create_app):
        mock_create_app.return_value.quality_gate.check.return_value = _one_fail_report()
        result = runner.invoke(app, ["check"])
        # Should show both passing and failing gates
        assert "lint" in result.output.lower()
        assert "types" in result.output.lower()

    @patch("src.infrastructure.composition.create_app")
    def test_single_gate_failure_details(self, mock_create_app):
        """When a gate fails, the output text from the runner should be visible."""
        fail_result = GateResult(
            gate=QualityGate.LINT,
            passed=False,
            output="src/main.py:1:1: E302 expected 2 blank lines",
            duration_ms=50,
        )
        mock_create_app.return_value.quality_gate.check.return_value = QualityReport(
            results=(fail_result,),
        )
        result = runner.invoke(app, ["check", "--gate", "lint"])
        assert result.exit_code == 1
        # The error detail from the runner should appear in output
        assert "E302" in result.output or "blank lines" in result.output
