"""Value objects for the Quality Gate bounded context.

QualityGate enumerates the gates (lint, types, tests, fitness).
GateResult captures the outcome of running a single gate.
QualityReport aggregates all gate results into a pass/fail verdict.
"""

from __future__ import annotations

import enum
from dataclasses import dataclass


class QualityGate(enum.Enum):
    """Quality gates that can be executed against a project."""

    LINT = "lint"
    TYPES = "types"
    TESTS = "tests"
    FITNESS = "fitness"


@dataclass(frozen=True)
class GateResult:
    """Outcome of running a single quality gate.

    Attributes:
        gate: Which quality gate was run.
        passed: Whether the gate passed.
        output: Captured stdout + stderr from the gate command.
        duration_ms: Wall-clock time in milliseconds.
    """

    gate: QualityGate
    passed: bool
    output: str
    duration_ms: int


@dataclass(frozen=True)
class QualityReport:
    """Aggregate result of running one or more quality gates.

    Attributes:
        results: Tuple of individual gate results.
    """

    results: tuple[GateResult, ...]

    @property
    def passed(self) -> bool:
        """True when every gate in the report passed."""
        return all(r.passed for r in self.results)
