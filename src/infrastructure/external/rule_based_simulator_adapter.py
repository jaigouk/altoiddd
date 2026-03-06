"""Rule-based simulator adapter — local fallback using SimulatorService.

Implements SimulatorPort by delegating to the stateless domain service.
No LLM required — pure heuristic scenario generation and tracing.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.domain.services.simulator_service import SimulatorService

if TYPE_CHECKING:
    from src.domain.models.domain_model import DomainModel
    from src.domain.models.scenario import Scenario, ScenarioTrace


class RuleBasedSimulatorAdapter:
    """SimulatorPort adapter that uses rule-based heuristics (no LLM)."""

    async def generate_scenarios(
        self,
        model: DomainModel,
        max_scenarios: int = 5,
    ) -> tuple[Scenario, ...]:
        """Generate scenarios using SimulatorService domain service.

        Args:
            model: The DomainModel to inspect.
            max_scenarios: Maximum scenarios to return.

        Returns:
            Tuple of Scenario VOs.
        """
        return SimulatorService.generate(model, max_scenarios)

    async def trace_scenario(
        self,
        scenario: Scenario,
        model: DomainModel,
    ) -> ScenarioTrace:
        """Trace a scenario using SimulatorService domain service.

        Args:
            scenario: The scenario to trace.
            model: The DomainModel to trace against.

        Returns:
            ScenarioTrace with steps and gaps found.
        """
        return SimulatorService.trace(scenario, model)
