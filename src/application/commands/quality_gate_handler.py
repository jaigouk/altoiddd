"""Application command handler for quality gate execution.

QualityGateHandler orchestrates running quality gates (lint, types, tests,
fitness) via a GateRunnerProtocol adapter and returns a QualityReport.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.domain.models.quality_gate import QualityGate, QualityReport

if TYPE_CHECKING:
    from src.application.ports.quality_gate_port import GateRunnerProtocol
    from src.domain.models.quality_gate import GateResult


class QualityGateHandler:
    """Orchestrates quality gate execution.

    Runs one or more quality gates via a GateRunnerProtocol adapter,
    collecting results into a QualityReport. Continues running
    remaining gates even when earlier gates fail.
    """

    _ALL_GATES: tuple[QualityGate, ...] = (
        QualityGate.LINT,
        QualityGate.TYPES,
        QualityGate.TESTS,
        QualityGate.FITNESS,
    )

    def __init__(self, runner: GateRunnerProtocol) -> None:
        """Initialize with a gate runner adapter.

        Args:
            runner: Infrastructure adapter that executes individual gates.
        """
        self._runner = runner

    def check(
        self,
        gates: tuple[QualityGate, ...] | None = None,
    ) -> QualityReport:
        """Run quality gates and return an aggregated report.

        Args:
            gates: Specific gates to run. None runs all four gates.

        Returns:
            QualityReport with results for each requested gate.
        """
        gates_to_run = gates if gates is not None else self._ALL_GATES
        results: list[GateResult] = []

        for gate in gates_to_run:
            result = self._runner.run(gate)
            results.append(result)

        return QualityReport(results=tuple(results))
