"""Tests for Domain Model bounded context value objects."""

from __future__ import annotations

import pytest

from src.domain.models.domain_values import (
    AggregateDesign,
    BoundedContext,
    ContextRelationship,
    DomainStory,
    SubdomainClassification,
)


class TestSubdomainClassification:
    def test_core_value(self) -> None:
        assert SubdomainClassification.CORE.value == "core"

    def test_supporting_value(self) -> None:
        assert SubdomainClassification.SUPPORTING.value == "supporting"

    def test_generic_value(self) -> None:
        assert SubdomainClassification.GENERIC.value == "generic"

    def test_all_members(self) -> None:
        members = {m.value for m in SubdomainClassification}
        assert members == {"core", "supporting", "generic"}


class TestDomainStory:
    def test_create_minimal(self) -> None:
        story = DomainStory(
            name="Checkout",
            actors=("Customer",),
            trigger="Customer clicks checkout",
            steps=("Customer reviews cart",),
        )
        assert story.name == "Checkout"
        assert story.actors == ("Customer",)
        assert story.observations == ()

    def test_frozen(self) -> None:
        story = DomainStory(name="Test", actors=("A",), trigger="T", steps=("S",))
        with pytest.raises(AttributeError):
            story.name = "Changed"  # type: ignore[misc]

    def test_with_observations(self) -> None:
        story = DomainStory(
            name="Test",
            actors=("A",),
            trigger="T",
            steps=("S",),
            observations=("Surprising finding",),
        )
        assert story.observations == ("Surprising finding",)


class TestBoundedContext:
    def test_create_without_classification(self) -> None:
        ctx = BoundedContext(name="Orders", responsibility="Manages orders")
        assert ctx.classification is None
        assert ctx.classification_rationale == ""

    def test_create_with_classification(self) -> None:
        ctx = BoundedContext(
            name="Orders",
            responsibility="Manages orders",
            classification=SubdomainClassification.CORE,
            classification_rationale="Competitive advantage",
        )
        assert ctx.classification == SubdomainClassification.CORE

    def test_frozen(self) -> None:
        ctx = BoundedContext(name="X", responsibility="Y")
        with pytest.raises(AttributeError):
            ctx.name = "Z"  # type: ignore[misc]


class TestContextRelationship:
    def test_create(self) -> None:
        rel = ContextRelationship(
            upstream="Orders",
            downstream="Shipping",
            integration_pattern="Domain Events",
        )
        assert rel.upstream == "Orders"
        assert rel.downstream == "Shipping"

    def test_frozen(self) -> None:
        rel = ContextRelationship(upstream="A", downstream="B", integration_pattern="Events")
        with pytest.raises(AttributeError):
            rel.upstream = "C"  # type: ignore[misc]


class TestAggregateDesign:
    def test_create_minimal(self) -> None:
        agg = AggregateDesign(
            name="OrderAggregate",
            context_name="Orders",
            root_entity="Order",
        )
        assert agg.name == "OrderAggregate"
        assert agg.contained_objects == ()
        assert agg.invariants == ()

    def test_create_full(self) -> None:
        agg = AggregateDesign(
            name="OrderAggregate",
            context_name="Orders",
            root_entity="Order",
            contained_objects=("OrderLine", "ShippingAddress"),
            invariants=("Total cannot be negative",),
            commands=("place_order", "cancel_order"),
            domain_events=("OrderPlaced", "OrderCancelled"),
        )
        assert len(agg.contained_objects) == 2
        assert len(agg.commands) == 2

    def test_frozen(self) -> None:
        agg = AggregateDesign(name="X", context_name="Y", root_entity="Z")
        with pytest.raises(AttributeError):
            agg.name = "W"  # type: ignore[misc]
