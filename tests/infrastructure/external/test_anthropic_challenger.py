"""Tests for AnthropicChallengerAdapter — LLM-powered challenge generation.

Verifies LLM delegation, structured output parsing, and fallback to
rule-based generation on LLMUnavailableError.
"""

from __future__ import annotations

from unittest.mock import AsyncMock

import pytest

from src.application.ports.challenger_port import ChallengerPort
from src.domain.models.challenge import ChallengeType
from src.domain.models.domain_model import DomainModel
from src.domain.models.domain_values import (
    AggregateDesign,
    BoundedContext,
    DomainStory,
    SubdomainClassification,
)
from src.domain.models.errors import LLMUnavailableError
from src.infrastructure.external.anthropic_challenger_adapter import (
    AnthropicChallengerAdapter,
)
from src.infrastructure.external.llm_client import LLMResponse


def _make_model() -> DomainModel:
    model = DomainModel()
    model.add_bounded_context(BoundedContext(name="Sales", responsibility="Orders"))
    model.classify_subdomain("Sales", SubdomainClassification.CORE)
    model.design_aggregate(
        AggregateDesign(
            name="OrderAggregate",
            context_name="Sales",
            root_entity="Order",
            invariants=(),
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


def _make_llm_response_json() -> str:
    """Return valid JSON for structured_output mock."""
    import json

    return json.dumps(
        {
            "challenges": [
                {
                    "challenge_type": "invariant",
                    "question_text": "What business rules protect OrderAggregate?",
                    "context_name": "Sales",
                    "source_reference": "Aggregate design: OrderAggregate",
                    "evidence": "",
                }
            ]
        }
    )


class TestAnthropicChallengerProtocol:
    def test_satisfies_challenger_port(self) -> None:
        llm = AsyncMock()
        adapter = AnthropicChallengerAdapter(llm_client=llm)
        assert isinstance(adapter, ChallengerPort)


class TestAnthropicChallengerLLMDelegation:
    @pytest.mark.asyncio
    async def test_calls_llm_structured_output(self) -> None:
        llm = AsyncMock()
        llm.structured_output.return_value = LLMResponse(
            content=_make_llm_response_json(),
            model_used="claude-sonnet-4-20250514",
            usage_tokens=100,
        )
        adapter = AnthropicChallengerAdapter(llm_client=llm)
        model = _make_model()

        challenges = await adapter.generate_challenges(model)

        llm.structured_output.assert_awaited_once()
        assert len(challenges) >= 1
        assert challenges[0].challenge_type == ChallengeType.INVARIANT

    @pytest.mark.asyncio
    async def test_parses_multiple_challenges(self) -> None:
        import json

        response_json = json.dumps(
            {
                "challenges": [
                    {
                        "challenge_type": "invariant",
                        "question_text": "What rules protect OrderAggregate?",
                        "context_name": "Sales",
                        "source_reference": "Aggregate: OrderAggregate",
                    },
                    {
                        "challenge_type": "failure_mode",
                        "question_text": "What if order creation fails?",
                        "context_name": "Sales",
                        "source_reference": "Story: Place Order",
                    },
                ]
            }
        )
        llm = AsyncMock()
        llm.structured_output.return_value = LLMResponse(
            content=response_json, model_used="m", usage_tokens=50
        )
        adapter = AnthropicChallengerAdapter(llm_client=llm)

        challenges = await adapter.generate_challenges(_make_model())

        assert len(challenges) == 2


class TestAnthropicChallengerFallback:
    @pytest.mark.asyncio
    async def test_falls_back_to_rule_based_on_llm_unavailable(self) -> None:
        llm = AsyncMock()
        llm.structured_output.side_effect = LLMUnavailableError("no key")
        adapter = AnthropicChallengerAdapter(llm_client=llm)
        model = _make_model()

        # Should not raise — falls back to rule-based
        challenges = await adapter.generate_challenges(model)

        # Rule-based should produce at least one challenge for this model
        assert len(challenges) >= 1

    @pytest.mark.asyncio
    async def test_falls_back_on_malformed_json(self) -> None:
        llm = AsyncMock()
        llm.structured_output.return_value = LLMResponse(
            content="not valid json", model_used="m", usage_tokens=10
        )
        adapter = AnthropicChallengerAdapter(llm_client=llm)

        challenges = await adapter.generate_challenges(_make_model())

        # Fallback to rule-based
        assert len(challenges) >= 1

    @pytest.mark.asyncio
    async def test_falls_back_on_missing_challenges_key(self) -> None:
        import json

        llm = AsyncMock()
        llm.structured_output.return_value = LLMResponse(
            content=json.dumps({"data": []}), model_used="m", usage_tokens=10
        )
        adapter = AnthropicChallengerAdapter(llm_client=llm)

        challenges = await adapter.generate_challenges(_make_model())

        assert len(challenges) >= 1
