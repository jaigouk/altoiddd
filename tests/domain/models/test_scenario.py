"""Tests for Scenario domain value objects (20c.6 Customer Simulator).

Verifies ScenarioType/GapType/ScenarioResult enums, Scenario/TraceStep/ModelGap/ScenarioTrace
immutability, and __post_init__ validation rules.
"""

from __future__ import annotations

import dataclasses

import pytest

from src.domain.models.errors import InvariantViolationError
from src.domain.models.scenario import (
    GapType,
    ModelGap,
    Scenario,
    ScenarioResult,
    ScenarioTrace,
    ScenarioType,
    SimulationTally,
    TraceStep,
)


class TestScenarioTypeEnum:
    def test_scenario_type_has_five_members(self) -> None:
        assert len(ScenarioType) == 5

    def test_happy_path_value(self) -> None:
        assert ScenarioType.HAPPY_PATH.value == "happy_path"

    def test_cross_context_value(self) -> None:
        assert ScenarioType.CROSS_CONTEXT.value == "cross_context"


class TestGapTypeEnum:
    def test_gap_type_has_five_members(self) -> None:
        assert len(GapType) == 5

    def test_missing_event_value(self) -> None:
        assert GapType.MISSING_EVENT.value == "missing_event"

    def test_ambiguous_ownership_value(self) -> None:
        assert GapType.AMBIGUOUS_OWNERSHIP.value == "ambiguous_ownership"


class TestScenarioResultEnum:
    def test_scenario_result_has_three_members(self) -> None:
        assert len(ScenarioResult) == 3

    def test_validated_value(self) -> None:
        assert ScenarioResult.VALIDATED.value == "validated"


class TestScenarioVO:
    def test_scenario_is_frozen(self) -> None:
        s = Scenario(
            scenario_type=ScenarioType.HAPPY_PATH,
            narrative_text="Customer places an order",
            context_names=("Sales",),
            derived_from="Checkout Flow",
        )
        with pytest.raises(dataclasses.FrozenInstanceError):
            s.narrative_text = "changed"  # type: ignore[misc]

    def test_scenario_requires_narrative_text(self) -> None:
        with pytest.raises(InvariantViolationError, match="narrative_text"):
            Scenario(
                scenario_type=ScenarioType.HAPPY_PATH,
                narrative_text="",
                context_names=("Sales",),
                derived_from="Checkout Flow",
            )

    def test_scenario_requires_context_names(self) -> None:
        with pytest.raises(InvariantViolationError, match="context_names"):
            Scenario(
                scenario_type=ScenarioType.HAPPY_PATH,
                narrative_text="Customer places an order",
                context_names=(),
                derived_from="Checkout Flow",
            )

    def test_scenario_requires_derived_from(self) -> None:
        with pytest.raises(InvariantViolationError, match="derived_from"):
            Scenario(
                scenario_type=ScenarioType.HAPPY_PATH,
                narrative_text="Customer places an order",
                context_names=("Sales",),
                derived_from="",
            )

    def test_scenario_whitespace_only_narrative_rejected(self) -> None:
        with pytest.raises(InvariantViolationError, match="narrative_text"):
            Scenario(
                scenario_type=ScenarioType.HAPPY_PATH,
                narrative_text="   ",
                context_names=("Sales",),
                derived_from="Checkout Flow",
            )


class TestTraceStepVO:
    def test_trace_step_is_frozen(self) -> None:
        ts = TraceStep(
            context="Sales",
            command_or_event="PlaceOrder",
            description="Customer places order",
            model_covers=True,
        )
        with pytest.raises(dataclasses.FrozenInstanceError):
            ts.context = "changed"  # type: ignore[misc]

    def test_trace_step_requires_context(self) -> None:
        with pytest.raises(InvariantViolationError, match="context"):
            TraceStep(
                context="",
                command_or_event="PlaceOrder",
                description="Customer places order",
                model_covers=True,
            )

    def test_trace_step_requires_command_or_event(self) -> None:
        with pytest.raises(InvariantViolationError, match="command_or_event"):
            TraceStep(
                context="Sales",
                command_or_event="",
                description="Customer places order",
                model_covers=True,
            )

    def test_trace_step_model_covers_none_means_unknown(self) -> None:
        ts = TraceStep(
            context="Sales",
            command_or_event="PlaceOrder",
            description="Unknown coverage",
            model_covers=None,
        )
        assert ts.model_covers is None


class TestModelGapVO:
    def test_model_gap_is_frozen(self) -> None:
        g = ModelGap(
            description="No event for payment failure",
            gap_type=GapType.MISSING_EVENT,
        )
        with pytest.raises(dataclasses.FrozenInstanceError):
            g.description = "changed"  # type: ignore[misc]

    def test_model_gap_requires_description(self) -> None:
        with pytest.raises(InvariantViolationError, match="description"):
            ModelGap(description="", gap_type=GapType.MISSING_EVENT)


class TestScenarioTraceVO:
    def test_scenario_trace_is_frozen(self) -> None:
        scenario = Scenario(
            scenario_type=ScenarioType.HAPPY_PATH,
            narrative_text="Customer places order",
            context_names=("Sales",),
            derived_from="Checkout Flow",
        )
        trace = ScenarioTrace(scenario=scenario, steps=(), gaps_found=())
        with pytest.raises(dataclasses.FrozenInstanceError):
            trace.steps = ()  # type: ignore[misc]

    def test_scenario_trace_holds_gaps(self) -> None:
        scenario = Scenario(
            scenario_type=ScenarioType.FAILURE_MODE,
            narrative_text="Payment fails mid-checkout",
            context_names=("Sales",),
            derived_from="Checkout Flow",
        )
        gap = ModelGap(
            description="No event for payment failure",
            gap_type=GapType.MISSING_EVENT,
        )
        trace = ScenarioTrace(
            scenario=scenario,
            steps=(),
            gaps_found=(gap,),
        )
        assert len(trace.gaps_found) == 1


class TestSimulationTallyVO:
    def test_tally_is_frozen(self) -> None:
        t = SimulationTally(tested=5, gaps=2, deferred=1)
        with pytest.raises(dataclasses.FrozenInstanceError):
            t.tested = 10  # type: ignore[misc]

    def test_tally_requires_non_negative_tested(self) -> None:
        with pytest.raises(InvariantViolationError, match="tested"):
            SimulationTally(tested=-1, gaps=0, deferred=0)

    def test_tally_requires_non_negative_gaps(self) -> None:
        with pytest.raises(InvariantViolationError, match="gaps"):
            SimulationTally(tested=0, gaps=-1, deferred=0)

    def test_tally_requires_non_negative_deferred(self) -> None:
        with pytest.raises(InvariantViolationError, match="deferred"):
            SimulationTally(tested=0, gaps=0, deferred=-1)
