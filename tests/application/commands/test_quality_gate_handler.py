"""Tests for QualityGateHandler.

Covers the application-layer orchestration: running all gates, running
specific gates, continuing after failure, and result aggregation.
"""

from __future__ import annotations

from src.domain.models.quality_gate import GateResult, QualityGate, QualityReport

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


class FakeGateRunner:
    """In-memory gate runner that records which gates were called."""

    def __init__(self, fail_gates: set[QualityGate] | None = None) -> None:
        self.called: list[QualityGate] = []
        self._fail_gates = fail_gates or set()

    def run(self, gate: QualityGate) -> GateResult:
        self.called.append(gate)
        passed = gate not in self._fail_gates
        return GateResult(
            gate=gate,
            passed=passed,
            output=f"{'ok' if passed else 'fail'}: {gate.value}",
            duration_ms=10,
        )


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------


class TestRunAllGates:
    def test_runs_all_four_gates_when_none_specified(self):
        from src.application.commands.quality_gate_handler import QualityGateHandler

        runner = FakeGateRunner()
        handler = QualityGateHandler(runner=runner)

        report = handler.check()

        assert len(runner.called) == 4
        assert set(runner.called) == {
            QualityGate.LINT,
            QualityGate.TYPES,
            QualityGate.TESTS,
            QualityGate.FITNESS,
        }
        assert isinstance(report, QualityReport)
        assert report.passed is True

    def test_returns_report_with_all_results(self):
        from src.application.commands.quality_gate_handler import QualityGateHandler

        runner = FakeGateRunner()
        handler = QualityGateHandler(runner=runner)

        report = handler.check()

        assert len(report.results) == 4
        gates_in_report = {r.gate for r in report.results}
        assert gates_in_report == {
            QualityGate.LINT,
            QualityGate.TYPES,
            QualityGate.TESTS,
            QualityGate.FITNESS,
        }


class TestRunSpecificGates:
    def test_runs_only_requested_gates(self):
        from src.application.commands.quality_gate_handler import QualityGateHandler

        runner = FakeGateRunner()
        handler = QualityGateHandler(runner=runner)

        report = handler.check(gates=(QualityGate.LINT, QualityGate.TYPES))

        assert len(runner.called) == 2
        assert runner.called == [QualityGate.LINT, QualityGate.TYPES]
        assert len(report.results) == 2

    def test_single_gate(self):
        from src.application.commands.quality_gate_handler import QualityGateHandler

        runner = FakeGateRunner()
        handler = QualityGateHandler(runner=runner)

        report = handler.check(gates=(QualityGate.TESTS,))

        assert len(runner.called) == 1
        assert runner.called[0] == QualityGate.TESTS
        assert report.passed is True


class TestContinuesAfterFailure:
    def test_continues_after_first_gate_fails(self):
        from src.application.commands.quality_gate_handler import QualityGateHandler

        runner = FakeGateRunner(fail_gates={QualityGate.LINT})
        handler = QualityGateHandler(runner=runner)

        report = handler.check()

        # All four gates should still have been called
        assert len(runner.called) == 4
        assert report.passed is False
        # Lint failed, others passed
        lint_result = next(r for r in report.results if r.gate == QualityGate.LINT)
        assert lint_result.passed is False
        tests_result = next(r for r in report.results if r.gate == QualityGate.TESTS)
        assert tests_result.passed is True

    def test_multiple_failures_still_runs_all(self):
        from src.application.commands.quality_gate_handler import QualityGateHandler

        runner = FakeGateRunner(fail_gates={QualityGate.LINT, QualityGate.TYPES})
        handler = QualityGateHandler(runner=runner)

        report = handler.check()

        assert len(runner.called) == 4
        assert report.passed is False
        failed_gates = {r.gate for r in report.results if not r.passed}
        assert failed_gates == {QualityGate.LINT, QualityGate.TYPES}


class TestReportCorrectness:
    def test_report_results_match_runner_output(self):
        from src.application.commands.quality_gate_handler import QualityGateHandler

        runner = FakeGateRunner(fail_gates={QualityGate.FITNESS})
        handler = QualityGateHandler(runner=runner)

        report = handler.check(gates=(QualityGate.LINT, QualityGate.FITNESS))

        assert len(report.results) == 2
        lint_result = report.results[0]
        fitness_result = report.results[1]
        assert lint_result.gate == QualityGate.LINT
        assert lint_result.passed is True
        assert fitness_result.gate == QualityGate.FITNESS
        assert fitness_result.passed is False

    def test_empty_gates_tuple_returns_empty_report(self):
        from src.application.commands.quality_gate_handler import QualityGateHandler

        runner = FakeGateRunner()
        handler = QualityGateHandler(runner=runner)

        report = handler.check(gates=())

        assert len(report.results) == 0
        assert report.passed is True
        assert runner.called == []
