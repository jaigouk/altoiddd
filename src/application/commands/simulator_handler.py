"""SimulatorHandler — orchestrates the Round 3 simulation lifecycle.

Thin orchestrator that delegates scenario generation and tracing to a
SimulatorPort, records verdicts, and produces a SimulationTally summary.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.domain.models.scenario import ScenarioResult, SimulationTally

if TYPE_CHECKING:
    from src.application.ports.simulator_port import SimulatorPort
    from src.domain.models.domain_model import DomainModel
    from src.domain.models.scenario import Scenario, ScenarioTrace


class SimulatorHandler:
    """Orchestrates generate → trace → verdict → tally cycle for Round 3."""

    def __init__(self, simulator: SimulatorPort) -> None:
        self._simulator = simulator
        self._verdicts: list[ScenarioResult] = []

    async def generate_scenarios(
        self,
        model: DomainModel,
        max_scenarios: int = 5,
    ) -> tuple[Scenario, ...]:
        """Delegate scenario generation to the port.

        Args:
            model: The DomainModel to analyze.
            max_scenarios: Maximum scenarios to return.

        Returns:
            Tuple of generated Scenario VOs.
        """
        return await self._simulator.generate_scenarios(model, max_scenarios)

    async def trace_scenario(
        self,
        scenario: Scenario,
        model: DomainModel,
    ) -> ScenarioTrace:
        """Delegate scenario tracing to the port.

        Args:
            scenario: The scenario to trace.
            model: The DomainModel to trace against.

        Returns:
            ScenarioTrace with steps and gaps found.
        """
        return await self._simulator.trace_scenario(scenario, model)

    def record_verdict(self, result: ScenarioResult) -> None:
        """Record a verdict for a traced scenario.

        Args:
            result: The verdict for this scenario.
        """
        self._verdicts.append(result)

    def tally(self) -> SimulationTally:
        """Produce a summary of all recorded verdicts.

        Returns:
            SimulationTally with tested, gaps, and deferred counts.
        """
        tested = len(self._verdicts)
        gaps = sum(1 for v in self._verdicts if v == ScenarioResult.MODEL_INCOMPLETE)
        deferred = sum(1 for v in self._verdicts if v == ScenarioResult.OUT_OF_SCOPE)
        return SimulationTally(tested=tested, gaps=gaps, deferred=deferred)
