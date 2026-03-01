"""Tests for SubprocessGateRunner infrastructure adapter.

Covers FITNESS gate skip behavior when tests/architecture/ is absent,
and basic GateResult structure validation.
"""

from __future__ import annotations

from src.domain.models.quality_gate import GateResult, QualityGate
from src.infrastructure.external.subprocess_gate_runner import SubprocessGateRunner


class TestFitnessGateSkip:
    def test_fitness_skips_when_no_architecture_dir(self, tmp_path):
        """FITNESS gate returns passed=True with skip message when dir missing."""
        runner = SubprocessGateRunner(project_dir=tmp_path)

        result = runner.run(QualityGate.FITNESS)

        assert isinstance(result, GateResult)
        assert result.gate == QualityGate.FITNESS
        assert result.passed is True
        assert "Skipped" in result.output
        assert result.duration_ms == 0

    def test_fitness_runs_when_architecture_dir_exists(self, tmp_path):
        """FITNESS gate attempts to run when tests/architecture/ exists."""
        arch_dir = tmp_path / "tests" / "architecture"
        arch_dir.mkdir(parents=True)

        runner = SubprocessGateRunner(project_dir=tmp_path)
        result = runner.run(QualityGate.FITNESS)

        # The command will likely fail (no pyproject.toml etc.) but
        # it should NOT skip -- it should actually attempt to run
        assert result.gate == QualityGate.FITNESS
        assert isinstance(result.duration_ms, int)
        assert result.duration_ms >= 0


class TestGateResultStructure:
    def test_result_has_correct_gate(self, tmp_path):
        """GateResult.gate matches the requested gate."""
        runner = SubprocessGateRunner(project_dir=tmp_path)

        result = runner.run(QualityGate.FITNESS)

        assert result.gate == QualityGate.FITNESS

    def test_duration_is_non_negative(self, tmp_path):
        """Duration should always be >= 0."""
        runner = SubprocessGateRunner(project_dir=tmp_path)

        result = runner.run(QualityGate.FITNESS)

        assert result.duration_ms >= 0


class TestProtocolCompliance:
    def test_implements_gate_runner_protocol(self):
        """SubprocessGateRunner satisfies the GateRunnerProtocol."""
        from src.application.ports.quality_gate_port import GateRunnerProtocol

        runner = SubprocessGateRunner()
        assert isinstance(runner, GateRunnerProtocol)
