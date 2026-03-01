"""SubprocessGateRunner -- infrastructure adapter for GateRunnerProtocol.

Executes quality gate commands (ruff, mypy, pytest) as subprocesses and
returns structured GateResult objects. Handles missing tools, timeouts,
and missing fitness test directories gracefully.
"""

from __future__ import annotations

import subprocess
import time
from pathlib import Path
from typing import ClassVar

from src.domain.models.quality_gate import GateResult, QualityGate


class SubprocessGateRunner:
    """Subprocess-based implementation of GateRunnerProtocol.

    Maps each QualityGate enum value to a shell command, executes it,
    and captures output, exit code, and duration.

    Attributes:
        _project_dir: The project directory to run commands against.
    """

    _GATE_COMMANDS: ClassVar[dict[QualityGate, list[str]]] = {
        QualityGate.LINT: ["uv", "run", "ruff", "check", "."],
        QualityGate.TYPES: ["uv", "run", "mypy", "."],
        QualityGate.TESTS: ["uv", "run", "pytest"],
        QualityGate.FITNESS: ["uv", "run", "pytest", "tests/architecture/"],
    }

    _TIMEOUT_SECONDS: ClassVar[int] = 300

    def __init__(self, project_dir: Path | None = None) -> None:
        """Initialize with an optional project directory.

        Args:
            project_dir: Directory to run commands in. Defaults to cwd.
        """
        self._project_dir = project_dir if project_dir is not None else Path.cwd()

    def run(self, gate: QualityGate) -> GateResult:
        """Execute a single quality gate as a subprocess.

        Args:
            gate: The quality gate to run.

        Returns:
            GateResult with pass/fail, combined stdout+stderr, and duration.
        """
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

        cmd = self._GATE_COMMANDS[gate]
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
