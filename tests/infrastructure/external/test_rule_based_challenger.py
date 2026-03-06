"""Tests for RuleBasedChallengerAdapter — local fallback using ChallengerService.

Verifies protocol compliance and delegation to the domain service.
"""

from __future__ import annotations

import pytest

from src.application.ports.challenger_port import ChallengerPort
from src.domain.models.domain_model import DomainModel
from src.domain.models.domain_values import (
    AggregateDesign,
    BoundedContext,
    DomainStory,
    SubdomainClassification,
)
from src.infrastructure.external.rule_based_challenger_adapter import (
    RuleBasedChallengerAdapter,
)


def _make_model_with_gaps() -> DomainModel:
    """Create a model that triggers at least one challenge."""
    model = DomainModel()
    model.add_bounded_context(BoundedContext(name="Sales", responsibility="Orders"))
    model.classify_subdomain("Sales", SubdomainClassification.CORE)
    model.design_aggregate(
        AggregateDesign(
            name="OrderAggregate",
            context_name="Sales",
            root_entity="Order",
            invariants=(),  # no invariants
        )
    )
    model.add_domain_story(
        DomainStory(
            name="Place Order",
            actors=("Customer",),
            trigger="Customer submits",
            steps=("System creates order",),
        )
    )
    model.add_term("Order", "A purchase", "Sales")
    return model


class TestRuleBasedChallengerProtocol:
    def test_satisfies_challenger_port(self) -> None:
        adapter = RuleBasedChallengerAdapter()
        assert isinstance(adapter, ChallengerPort)


class TestRuleBasedChallengerDelegation:
    @pytest.mark.asyncio
    async def test_generates_challenges_from_model(self) -> None:
        adapter = RuleBasedChallengerAdapter()
        model = _make_model_with_gaps()
        challenges = await adapter.generate_challenges(model)
        assert len(challenges) >= 1

    @pytest.mark.asyncio
    async def test_respects_max_per_type(self) -> None:
        adapter = RuleBasedChallengerAdapter()
        model = _make_model_with_gaps()
        challenges = await adapter.generate_challenges(model, max_per_type=1)
        by_type: dict[str, int] = {}
        for c in challenges:
            by_type[c.challenge_type.value] = by_type.get(c.challenge_type.value, 0) + 1
        for type_name, count in by_type.items():
            assert count <= 1, f"{type_name} has {count} challenges (max 1)"

    @pytest.mark.asyncio
    async def test_empty_model_returns_no_challenges(self) -> None:
        adapter = RuleBasedChallengerAdapter()
        model = DomainModel()
        challenges = await adapter.generate_challenges(model)
        assert challenges == ()
