"""Port for the Customer Simulator bounded context.

Defines the interface for generating scenarios and tracing them
against a DomainModel to find gaps.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from src.domain.models.domain_model import DomainModel
    from src.domain.models.scenario import Scenario, ScenarioTrace


@runtime_checkable
class SimulatorPort(Protocol):
    """Interface for scenario generation and tracing in Round 3 simulation.

    Adapters implement this to generate scenarios from a DomainModel,
    either via rule-based heuristics (local) or LLM-powered analysis.
    """

    async def generate_scenarios(
        self,
        model: DomainModel,
        max_scenarios: int = 5,
    ) -> tuple[Scenario, ...]:
        """Generate typed scenarios by analyzing the domain model.

        Args:
            model: The DomainModel aggregate to inspect.
            max_scenarios: Maximum scenarios to return.

        Returns:
            Tuple of Scenario VOs.
        """
        ...

    async def trace_scenario(
        self,
        scenario: Scenario,
        model: DomainModel,
    ) -> ScenarioTrace:
        """Trace a scenario against the model to find coverage and gaps.

        Args:
            scenario: The scenario to trace.
            model: The DomainModel to trace against.

        Returns:
            ScenarioTrace with steps and gaps found.
        """
        ...
