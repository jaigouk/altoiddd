"""Anthropic LLM-powered challenger adapter.

Implements ChallengerPort by sending the DomainModel summary to an LLM
via structured_output and parsing the response into Challenge VOs.
Falls back to rule-based generation on LLMUnavailableError or parse failure.
"""

from __future__ import annotations

import json
import logging
from typing import TYPE_CHECKING, Any

from src.domain.models.challenge import Challenge, ChallengeType
from src.domain.models.errors import InvariantViolationError, LLMUnavailableError
from src.domain.services.challenger_service import ChallengerService

if TYPE_CHECKING:
    from src.domain.models.domain_model import DomainModel
    from src.infrastructure.external.llm_client import LLMClient

logger = logging.getLogger(__name__)

_CHALLENGE_SCHEMA: dict[str, Any] = {
    "type": "object",
    "properties": {
        "challenges": {
            "type": "array",
            "items": {
                "type": "object",
                "properties": {
                    "challenge_type": {
                        "type": "string",
                        "enum": [ct.value for ct in ChallengeType],
                    },
                    "question_text": {"type": "string"},
                    "context_name": {"type": "string"},
                    "source_reference": {"type": "string"},
                    "evidence": {"type": "string"},
                },
                "required": [
                    "challenge_type",
                    "question_text",
                    "context_name",
                    "source_reference",
                ],
            },
        }
    },
    "required": ["challenges"],
}


class AnthropicChallengerAdapter:
    """ChallengerPort adapter that uses LLM for challenge generation."""

    def __init__(self, llm_client: LLMClient) -> None:
        self._llm = llm_client

    async def generate_challenges(
        self,
        model: DomainModel,
        max_per_type: int = 5,
    ) -> tuple[Challenge, ...]:
        """Generate challenges via LLM, falling back to rule-based on failure.

        Args:
            model: The DomainModel to analyze.
            max_per_type: Maximum challenges per type.

        Returns:
            Tuple of Challenge VOs.
        """
        try:
            return await self._llm_generate(model, max_per_type)
        except (
            LLMUnavailableError,
            InvariantViolationError,
            ValueError,
            KeyError,
            json.JSONDecodeError,
        ):
            logger.info("LLM challenge generation failed, falling back to rule-based")
            return ChallengerService.generate(model, max_per_type)

    async def _llm_generate(
        self,
        model: DomainModel,
        max_per_type: int,
    ) -> tuple[Challenge, ...]:
        """Call LLM and parse structured output into Challenge VOs."""
        prompt = self._build_prompt(model, max_per_type)
        response = await self._llm.structured_output(prompt, _CHALLENGE_SCHEMA)

        data = json.loads(response.content)
        raw_challenges = data["challenges"]

        challenges = [
            Challenge(
                challenge_type=ChallengeType(item["challenge_type"]),
                question_text=item["question_text"],
                context_name=item["context_name"],
                source_reference=item["source_reference"],
                evidence=item.get("evidence", ""),
            )
            for item in raw_challenges
        ]
        return tuple(challenges)

    @staticmethod
    def _build_prompt(model: DomainModel, max_per_type: int) -> str:
        """Build a prompt summarizing the DomainModel for the LLM."""
        parts = ["Analyze this domain model and generate challenges:\n"]

        parts.append("## Bounded Contexts")
        for ctx in model.bounded_contexts:
            classification = ctx.classification.value if ctx.classification else "unclassified"
            parts.append(f"- {ctx.name} ({classification}): {ctx.responsibility}")

        parts.append("\n## Aggregates")
        for agg in model.aggregate_designs:
            inv_count = len(agg.invariants)
            parts.append(
                f"- {agg.name} in {agg.context_name}: "
                f"root={agg.root_entity}, invariants={inv_count}"
            )

        parts.append("\n## Domain Stories")
        parts.extend(
            f"- {story.name}: {' → '.join(story.steps)}"
            for story in model.domain_stories
        )

        parts.append("\n## Ubiquitous Language")
        parts.extend(
            f"- {entry.term} ({entry.context_name}): {entry.definition}"
            for entry in model.ubiquitous_language.terms
        )

        parts.append(
            f"\nGenerate up to {max_per_type} challenges per type. "
            f"Types: {', '.join(ct.value for ct in ChallengeType)}. "
            f"Each challenge must be a QUESTION (never state facts). "
            f"Cite source_reference for every challenge."
        )
        return "\n".join(parts)
