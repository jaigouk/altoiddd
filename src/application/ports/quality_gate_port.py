"""Ports for quality gate operations.

GateRunnerProtocol: low-level interface for running a single quality gate.
QualityGatePort: high-level interface for checking multiple gates at once.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from src.domain.models.quality_gate import GateResult, QualityGate, QualityReport


@runtime_checkable
class GateRunnerProtocol(Protocol):
    """Interface for running a single quality gate command.

    Infrastructure adapters implement this to execute lint, type-check,
    test, or fitness commands and return structured results.
    """

    def run(self, gate: QualityGate) -> GateResult:
        """Execute a single quality gate and return its result.

        Args:
            gate: The quality gate to run.

        Returns:
            GateResult with pass/fail, output, and duration.
        """
        ...


@runtime_checkable
class QualityGatePort(Protocol):
    """Interface for running multiple quality gate checks.

    Higher-level port that orchestrates running a set of gates
    and returns an aggregated report.
    """

    def check(self, gates: tuple[QualityGate, ...] | None = None) -> QualityReport:
        """Run the specified quality gates (or all if None).

        Args:
            gates: Tuple of gates to run. None means all gates.

        Returns:
            QualityReport aggregating all individual results.
        """
        ...
