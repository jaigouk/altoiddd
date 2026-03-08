"""Tests for SimulatorHandler — orchestrates the Round 3 simulation lifecycle.

Verifies that the handler delegates to SimulatorPort for generation/tracing,
records verdicts, and produces a SimulationTally.
"""

from __future__ import annotations

import pytest

from src.application.commands.simulator_handler import SimulatorHandler
from src.application.ports.simulator_port import SimulatorPort
from src.domain.models.domain_model import DomainModel
from src.domain.models.domain_values import (
    AggregateDesign,
    BoundedContext,
    DomainStory,
    SubdomainClassification,
)
from src.domain.models.scenario import (
    GapType,
    ModelGap,
    Scenario,
    ScenarioResult,
    ScenarioTrace,
    ScenarioType,
    TraceStep,
)


def _make_model() -> DomainModel:
    model = DomainModel()
    model.add_bounded_context(BoundedContext(name="Sales", responsibility="Orders"))
    model.classify_subdomain("Sales", SubdomainClassification.CORE)
    model.design_aggregate(
        AggregateDesign(
            name="OrderAggregate",
            context_name="Sales",
            root_entity="Order",
            commands=("PlaceOrder",),
            domain_events=("OrderPlaced",),
        )
    )
    model.add_domain_story(
        DomainStory(
            name="Checkout",
            actors=("Customer",),
            trigger="Customer clicks checkout",
            steps=("System creates order",),
        )
    )
    model.add_term("Order", "A purchase", "Sales")
    return model


_SCENARIO = Scenario(
    scenario_type=ScenarioType.HAPPY_PATH,
    narrative_text="Customer places an order",
    context_names=("Sales",),
    derived_from="Checkout Flow",
)

_TRACE = ScenarioTrace(
    scenario=_SCENARIO,
    steps=(
        TraceStep(
            context="Sales",
            command_or_event="PlaceOrder",
            description="Command on OrderAggregate",
            model_covers=True,
        ),
    ),
    gaps_found=(),
)

_TRACE_WITH_GAPS = ScenarioTrace(
    scenario=_SCENARIO,
    steps=(),
    gaps_found=(
        ModelGap(description="Missing event", gap_type=GapType.MISSING_EVENT),
    ),
)


class FakeSimulatorAdapter:
    """Fake SimulatorPort for testing."""

    def __init__(
        self,
        scenarios: tuple[Scenario, ...] = (),
        trace: ScenarioTrace | None = None,
    ) -> None:
        self._scenarios = scenarios
        self._trace = trace or _TRACE

    async def generate_scenarios(
        self, model: DomainModel, max_scenarios: int = 5
    ) -> tuple[Scenario, ...]:
        return self._scenarios

    async def trace_scenario(
        self, scenario: Scenario, model: DomainModel
    ) -> ScenarioTrace:
        return self._trace


class TestSimulatorHandlerProtocol:
    def test_fake_adapter_satisfies_port(self) -> None:
        adapter = FakeSimulatorAdapter()
        assert isinstance(adapter, SimulatorPort)


class TestSimulatorHandlerGenerate:
    @pytest.mark.asyncio
    async def test_generate_delegates_to_port(self) -> None:
        adapter = FakeSimulatorAdapter(scenarios=(_SCENARIO,))
        handler = SimulatorHandler(simulator=adapter)
        model = _make_model()
        result = await handler.generate_scenarios(model, max_scenarios=5)
        assert len(result) == 1
        assert result[0] is _SCENARIO


class TestSimulatorHandlerTrace:
    @pytest.mark.asyncio
    async def test_trace_delegates_to_port(self) -> None:
        adapter = FakeSimulatorAdapter(trace=_TRACE)
        handler = SimulatorHandler(simulator=adapter)
        model = _make_model()
        result = await handler.trace_scenario(_SCENARIO, model)
        assert result is _TRACE


class TestSimulatorHandlerTally:
    def test_tally_counts_verdicts(self) -> None:
        adapter = FakeSimulatorAdapter()
        handler = SimulatorHandler(simulator=adapter)

        handler.record_verdict(ScenarioResult.VALIDATED)
        handler.record_verdict(ScenarioResult.MODEL_INCOMPLETE)
        handler.record_verdict(ScenarioResult.OUT_OF_SCOPE)

        tally = handler.tally()
        assert tally.tested == 3
        assert tally.gaps == 1
        assert tally.deferred == 1

    def test_tally_empty_when_no_verdicts(self) -> None:
        adapter = FakeSimulatorAdapter()
        handler = SimulatorHandler(simulator=adapter)
        tally = handler.tally()
        assert tally.tested == 0
        assert tally.gaps == 0
        assert tally.deferred == 0

    def test_tally_all_validated(self) -> None:
        adapter = FakeSimulatorAdapter()
        handler = SimulatorHandler(simulator=adapter)
        handler.record_verdict(ScenarioResult.VALIDATED)
        handler.record_verdict(ScenarioResult.VALIDATED)
        handler.record_verdict(ScenarioResult.VALIDATED)
        tally = handler.tally()
        assert tally.tested == 3
        assert tally.gaps == 0
        assert tally.deferred == 0

    def test_tally_all_incomplete(self) -> None:
        adapter = FakeSimulatorAdapter()
        handler = SimulatorHandler(simulator=adapter)
        handler.record_verdict(ScenarioResult.MODEL_INCOMPLETE)
        handler.record_verdict(ScenarioResult.MODEL_INCOMPLETE)
        tally = handler.tally()
        assert tally.tested == 2
        assert tally.gaps == 2
        assert tally.deferred == 0
