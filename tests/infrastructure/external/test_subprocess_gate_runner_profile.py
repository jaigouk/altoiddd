"""Tests for SubprocessGateRunner profile integration.

Verifies that the runner reads commands from StackProfile instead
of hardcoded _GATE_COMMANDS, and that GenericProfile (empty commands)
produces graceful skip results.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.domain.models.quality_gate import GateResult, QualityGate

if TYPE_CHECKING:
    from pathlib import Path
from src.domain.models.stack_profile import GenericProfile, PythonUvProfile
from src.infrastructure.external.subprocess_gate_runner import SubprocessGateRunner


class TestRunnerUsesProfileCommands:
    """Runner reads commands from profile, not hardcoded dict."""

    def test_lint_command_from_python_profile(self, tmp_path: Path) -> None:
        """LINT gate uses profile.quality_gate_commands[LINT]."""
        profile = PythonUvProfile()
        runner = SubprocessGateRunner(project_dir=tmp_path, profile=profile)

        result = runner.run(QualityGate.LINT)

        assert isinstance(result, GateResult)
        assert result.gate == QualityGate.LINT
        # Command will fail (no pyproject.toml) but it should attempt to run
        assert result.duration_ms >= 0

    def test_types_command_from_python_profile(self, tmp_path: Path) -> None:
        """TYPES gate uses profile.quality_gate_commands[TYPES]."""
        profile = PythonUvProfile()
        runner = SubprocessGateRunner(project_dir=tmp_path, profile=profile)

        result = runner.run(QualityGate.TYPES)

        assert isinstance(result, GateResult)
        assert result.gate == QualityGate.TYPES

    def test_tests_command_from_python_profile(self, tmp_path: Path) -> None:
        """TESTS gate uses profile.quality_gate_commands[TESTS]."""
        profile = PythonUvProfile()
        runner = SubprocessGateRunner(project_dir=tmp_path, profile=profile)

        result = runner.run(QualityGate.TESTS)

        assert isinstance(result, GateResult)
        assert result.gate == QualityGate.TESTS


class TestRunnerGenericProfileSkips:
    """GenericProfile (empty commands) produces skip results."""

    def test_lint_skips_with_generic_profile(self, tmp_path: Path) -> None:
        """LINT gate returns skip result when profile has no commands."""
        profile = GenericProfile()
        runner = SubprocessGateRunner(project_dir=tmp_path, profile=profile)

        result = runner.run(QualityGate.LINT)

        assert isinstance(result, GateResult)
        assert result.gate == QualityGate.LINT
        assert result.passed is True
        assert "skip" in result.output.lower()
        assert result.duration_ms == 0

    def test_types_skips_with_generic_profile(self, tmp_path: Path) -> None:
        """TYPES gate returns skip result when profile has no commands."""
        profile = GenericProfile()
        runner = SubprocessGateRunner(project_dir=tmp_path, profile=profile)

        result = runner.run(QualityGate.TYPES)

        assert result.passed is True
        assert "skip" in result.output.lower()

    def test_tests_skips_with_generic_profile(self, tmp_path: Path) -> None:
        """TESTS gate returns skip result when profile has no commands."""
        profile = GenericProfile()
        runner = SubprocessGateRunner(project_dir=tmp_path, profile=profile)

        result = runner.run(QualityGate.TESTS)

        assert result.passed is True
        assert "skip" in result.output.lower()

    def test_fitness_skips_with_generic_profile(self, tmp_path: Path) -> None:
        """FITNESS gate returns skip result when profile has no commands."""
        profile = GenericProfile()
        runner = SubprocessGateRunner(project_dir=tmp_path, profile=profile)

        result = runner.run(QualityGate.FITNESS)

        assert result.passed is True
        assert "skip" in result.output.lower()


class TestRunnerBackwardCompat:
    """Runner without profile still works (uses default PythonUvProfile)."""

    def test_no_profile_uses_default(self, tmp_path: Path) -> None:
        """Runner without profile arg still runs gates."""
        runner = SubprocessGateRunner(project_dir=tmp_path)

        result = runner.run(QualityGate.FITNESS)

        # Should still skip gracefully when arch dir missing
        assert result.gate == QualityGate.FITNESS
        assert isinstance(result, GateResult)
