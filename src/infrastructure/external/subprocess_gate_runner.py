"""SubprocessGateRunner -- infrastructure adapter for GateRunnerProtocol.

Executes quality gate commands as subprocesses and returns structured
GateResult objects. Commands are read from a StackProfile so the runner
works for any supported stack. GenericProfile (empty commands) produces
graceful skip results.
"""

from __future__ import annotations

import subprocess
import time
from pathlib import Path
from typing import TYPE_CHECKING, ClassVar

from src.domain.models.quality_gate import GateResult, QualityGate

if TYPE_CHECKING:
    from src.domain.models.stack_profile import StackProfile


class SubprocessGateRunner:
    """Subprocess-based implementation of GateRunnerProtocol.

    Reads gate commands from a StackProfile and executes them as
    subprocesses, capturing output, exit code, and duration.

    Attributes:
        _project_dir: The project directory to run commands against.
        _commands: Gate-to-command mapping from the stack profile.
    """

    _TIMEOUT_SECONDS: ClassVar[int] = 300

    def __init__(
        self,
        project_dir: Path | None = None,
        profile: StackProfile | None = None,
    ) -> None:
        """Initialize with an optional project directory and stack profile.

        Args:
            project_dir: Directory to run commands in. Defaults to cwd.
            profile: Stack profile providing gate commands. Defaults to
                PythonUvProfile for backward compatibility.
        """
        self._project_dir = project_dir if project_dir is not None else Path.cwd()
        if profile is None:
            from src.domain.models.stack_profile import PythonUvProfile

            profile = PythonUvProfile()
        self._commands = profile.quality_gate_commands

    def run(self, gate: QualityGate) -> GateResult:
        """Execute a single quality gate as a subprocess.

        Args:
            gate: The quality gate to run.

        Returns:
            GateResult with pass/fail, combined stdout+stderr, and duration.
        """
        # Skip gracefully if profile has no command for this gate
        if gate not in self._commands:
            return GateResult(
                gate=gate,
                passed=True,
                output=f"Skipped: no {gate.value} command for this stack",
                duration_ms=0,
            )

        # FITNESS gate: skip gracefully if tests/architecture/ does not exist
        if gate == QualityGate.FITNESS:
            arch_dir = self._project_dir / "tests" / "architecture"
            if not arch_dir.exists():
                return GateResult(
                    gate=gate,
                    passed=True,
                    output="Skipped: tests/architecture/ directory not found",
                    duration_ms=0,
                )

        cmd = self._commands[gate]
        start = time.monotonic()

        try:
            result = subprocess.run(  # noqa: S603
                cmd,
                cwd=str(self._project_dir),
                capture_output=True,
                text=True,
                timeout=self._TIMEOUT_SECONDS,
            )
            duration_ms = int((time.monotonic() - start) * 1000)
            output = result.stdout + result.stderr
            passed = result.returncode == 0

        except FileNotFoundError:
            duration_ms = int((time.monotonic() - start) * 1000)
            output = f"Command not found: {cmd[0]}"
            passed = False

        except subprocess.TimeoutExpired:
            duration_ms = int((time.monotonic() - start) * 1000)
            output = f"Timed out after {self._TIMEOUT_SECONDS}s"
            passed = False

        return GateResult(
            gate=gate,
            passed=passed,
            output=output,
            duration_ms=duration_ms,
        )
