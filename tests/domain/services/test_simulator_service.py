"""Tests for SimulatorService — rule-based scenario generation and tracing.

Verifies that the stateless service inspects a DomainModel and generates
typed scenarios for happy paths, failure modes, edge cases, and cross-context
flows; and traces scenarios against the model to find gaps.
"""

from __future__ import annotations

from src.domain.models.domain_model import DomainModel
from src.domain.models.domain_values import (
    AggregateDesign,
    BoundedContext,
    ContextRelationship,
    DomainStory,
    SubdomainClassification,
)
from src.domain.models.scenario import Scenario, ScenarioType
from src.domain.services.simulator_service import SimulatorService


def _make_rich_model() -> DomainModel:
    """Create a DomainModel with enough data to trigger all scenario types."""
    model = DomainModel()

    model.add_bounded_context(BoundedContext(name="Sales", responsibility="Orders"))
    model.classify_subdomain("Sales", SubdomainClassification.CORE)
    model.add_bounded_context(BoundedContext(name="Shipping", responsibility="Deliveries"))
    model.classify_subdomain("Shipping", SubdomainClassification.SUPPORTING)

    model.design_aggregate(
        AggregateDesign(
            name="OrderAggregate",
            context_name="Sales",
            root_entity="Order",
            invariants=("Total must be positive",),
            commands=("PlaceOrder", "CancelOrder"),
            domain_events=("OrderPlaced", "OrderCancelled"),
        )
    )

    model.add_domain_story(
        DomainStory(
            name="Checkout Flow",
            actors=("Customer",),
            trigger="Customer clicks checkout",
            steps=(
                "Customer reviews order",
                "System validates payment",
                "System creates shipment",
            ),
        )
    )

    model.add_term("Order", "A customer purchase", "Sales")
    model.add_term("Shipment", "A delivery package", "Shipping")

    model.add_context_relationship(ContextRelationship("Sales", "Shipping", "Domain Events"))

    return model


def _make_empty_model() -> DomainModel:
    return DomainModel()


def _make_generic_only_model() -> DomainModel:
    model = DomainModel()
    model.add_bounded_context(BoundedContext(name="Auth", responsibility="Authentication"))
    model.classify_subdomain("Auth", SubdomainClassification.GENERIC)
    model.add_domain_story(
        DomainStory(
            name="Login",
            actors=("User",),
            trigger="User enters credentials",
            steps=("System verifies credentials",),
        )
    )
    model.add_term("User", "An authenticated person", "Auth")
    return model


class TestSimulatorServiceGenerate:
    def test_generates_happy_path_scenarios_for_core_stories(self) -> None:
        model = _make_rich_model()
        scenarios = SimulatorService.generate(model)
        happy = [s for s in scenarios if s.scenario_type == ScenarioType.HAPPY_PATH]
        assert len(happy) >= 1

    def test_generates_failure_mode_scenarios(self) -> None:
        model = _make_rich_model()
        scenarios = SimulatorService.generate(model)
        failures = [s for s in scenarios if s.scenario_type == ScenarioType.FAILURE_MODE]
        assert len(failures) >= 1

    def test_generates_edge_case_scenarios_from_invariants(self) -> None:
        model = _make_rich_model()
        scenarios = SimulatorService.generate(model)
        edges = [s for s in scenarios if s.scenario_type == ScenarioType.EDGE_CASE]
        assert len(edges) >= 1

    def test_generates_cross_context_scenarios_for_relationships(self) -> None:
        model = _make_rich_model()
        scenarios = SimulatorService.generate(model)
        cross = [s for s in scenarios if s.scenario_type == ScenarioType.CROSS_CONTEXT]
        assert len(cross) >= 1

    def test_empty_model_returns_no_scenarios(self) -> None:
        model = _make_empty_model()
        scenarios = SimulatorService.generate(model)
        assert scenarios == ()

    def test_max_scenarios_respected(self) -> None:
        model = _make_rich_model()
        scenarios = SimulatorService.generate(model, max_scenarios=2)
        assert len(scenarios) <= 2

    def test_scenarios_only_reference_model_contexts(self) -> None:
        model = _make_rich_model()
        scenarios = SimulatorService.generate(model)
        context_names = {ctx.name for ctx in model.bounded_contexts}
        for s in scenarios:
            for cn in s.context_names:
                assert cn in context_names, f"Scenario references unknown context: {cn}"

    def test_generic_only_model_no_happy_path(self) -> None:
        """Generic subdomains don't get Core-focused happy path scenarios."""
        model = _make_generic_only_model()
        scenarios = SimulatorService.generate(model)
        happy = [s for s in scenarios if s.scenario_type == ScenarioType.HAPPY_PATH]
        assert len(happy) == 0


class TestSimulatorServiceTrace:
    def test_trace_returns_steps_matching_scenario_contexts(self) -> None:
        model = _make_rich_model()
        scenario = Scenario(
            scenario_type=ScenarioType.HAPPY_PATH,
            narrative_text="Customer places an order",
            context_names=("Sales",),
            derived_from="Checkout Flow",
        )
        trace = SimulatorService.trace(scenario, model)
        assert len(trace.steps) >= 1
        for step in trace.steps:
            assert step.context in {ctx.name for ctx in model.bounded_contexts}

    def test_trace_finds_gaps_for_missing_coverage(self) -> None:
        """A scenario referencing a context with no aggregates should find gaps."""
        model = DomainModel()
        model.add_bounded_context(BoundedContext(name="Billing", responsibility="Payments"))
        model.classify_subdomain("Billing", SubdomainClassification.CORE)
        model.add_domain_story(
            DomainStory(
                name="Pay Invoice",
                actors=("Customer",),
                trigger="Invoice due",
                steps=("Customer pays invoice",),
            )
        )
        model.add_term("Invoice", "A bill for payment", "Billing")
        scenario = Scenario(
            scenario_type=ScenarioType.HAPPY_PATH,
            narrative_text="Customer pays an invoice",
            context_names=("Billing",),
            derived_from="Pay Invoice",
        )
        trace = SimulatorService.trace(scenario, model)
        # With no aggregates in Billing, there should be gap(s)
        assert len(trace.gaps_found) >= 1

    def test_trace_covers_steps_when_aggregate_exists(self) -> None:
        model = _make_rich_model()
        scenario = Scenario(
            scenario_type=ScenarioType.HAPPY_PATH,
            narrative_text="Customer places an order via checkout",
            context_names=("Sales",),
            derived_from="Checkout Flow",
        )
        trace = SimulatorService.trace(scenario, model)
        covered = [s for s in trace.steps if s.model_covers is True]
        assert len(covered) >= 1
