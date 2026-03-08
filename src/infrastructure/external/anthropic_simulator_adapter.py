"""Anthropic LLM-powered simulator adapter.

Implements SimulatorPort by sending the DomainModel summary to an LLM
via structured_output and parsing the response into Scenario/ScenarioTrace VOs.
Falls back to rule-based generation on LLMUnavailableError or parse failure.
"""

from __future__ import annotations

import json
import logging
from typing import TYPE_CHECKING, Any

from src.domain.models.errors import InvariantViolationError, LLMUnavailableError
from src.domain.models.scenario import (
    GapType,
    ModelGap,
    Scenario,
    ScenarioTrace,
    ScenarioType,
    TraceStep,
)
from src.domain.services.simulator_service import SimulatorService

if TYPE_CHECKING:
    from src.domain.models.domain_model import DomainModel
    from src.infrastructure.external.llm_client import LLMClient

logger = logging.getLogger(__name__)

_SCENARIO_SCHEMA: dict[str, Any] = {
    "type": "object",
    "properties": {
        "scenarios": {
            "type": "array",
            "items": {
                "type": "object",
                "properties": {
                    "scenario_type": {
                        "type": "string",
                        "enum": [st.value for st in ScenarioType],
                    },
                    "narrative_text": {"type": "string"},
                    "context_names": {
                        "type": "array",
                        "items": {"type": "string"},
                    },
                    "derived_from": {"type": "string"},
                },
                "required": [
                    "scenario_type",
                    "narrative_text",
                    "context_names",
                    "derived_from",
                ],
            },
        }
    },
    "required": ["scenarios"],
}

_TRACE_SCHEMA: dict[str, Any] = {
    "type": "object",
    "properties": {
        "steps": {
            "type": "array",
            "items": {
                "type": "object",
                "properties": {
                    "context": {"type": "string"},
                    "command_or_event": {"type": "string"},
                    "description": {"type": "string"},
                    "model_covers": {"type": ["boolean", "null"]},
                },
                "required": ["context", "command_or_event", "description", "model_covers"],
            },
        },
        "gaps": {
            "type": "array",
            "items": {
                "type": "object",
                "properties": {
                    "description": {"type": "string"},
                    "gap_type": {
                        "type": "string",
                        "enum": [gt.value for gt in GapType],
                    },
                },
                "required": ["description", "gap_type"],
            },
        },
    },
    "required": ["steps", "gaps"],
}


class AnthropicSimulatorAdapter:
    """SimulatorPort adapter that uses LLM for scenario generation and tracing."""

    def __init__(self, llm_client: LLMClient) -> None:
        self._llm = llm_client

    async def generate_scenarios(
        self,
        model: DomainModel,
        max_scenarios: int = 5,
    ) -> tuple[Scenario, ...]:
        """Generate scenarios via LLM, falling back to rule-based on failure.

        Args:
            model: The DomainModel to analyze.
            max_scenarios: Maximum scenarios to return.

        Returns:
            Tuple of Scenario VOs.
        """
        try:
            return await self._llm_generate(model, max_scenarios)
        except (
            LLMUnavailableError,
            InvariantViolationError,
            ValueError,
            KeyError,
            json.JSONDecodeError,
        ):
            logger.info("LLM scenario generation failed, falling back to rule-based")
            return SimulatorService.generate(model, max_scenarios)

    async def trace_scenario(
        self,
        scenario: Scenario,
        model: DomainModel,
    ) -> ScenarioTrace:
        """Trace a scenario via LLM, falling back to rule-based on failure.

        Args:
            scenario: The scenario to trace.
            model: The DomainModel to trace against.

        Returns:
            ScenarioTrace with steps and gaps found.
        """
        try:
            return await self._llm_trace(scenario, model)
        except (
            LLMUnavailableError,
            InvariantViolationError,
            ValueError,
            KeyError,
            json.JSONDecodeError,
        ):
            logger.info("LLM scenario tracing failed, falling back to rule-based")
            return SimulatorService.trace(scenario, model)

    async def _llm_generate(
        self,
        model: DomainModel,
        max_scenarios: int,
    ) -> tuple[Scenario, ...]:
        """Call LLM and parse structured output into Scenario VOs."""
        prompt = self._build_generate_prompt(model, max_scenarios)
        response = await self._llm.structured_output(prompt, _SCENARIO_SCHEMA)

        data = json.loads(response.content)
        raw_scenarios = data["scenarios"]

        scenarios = [
            Scenario(
                scenario_type=ScenarioType(item["scenario_type"]),
                narrative_text=item["narrative_text"],
                context_names=tuple(item["context_names"]),
                derived_from=item["derived_from"],
            )
            for item in raw_scenarios
        ]
        return tuple(scenarios)

    async def _llm_trace(
        self,
        scenario: Scenario,
        model: DomainModel,
    ) -> ScenarioTrace:
        """Call LLM and parse structured output into ScenarioTrace."""
        prompt = self._build_trace_prompt(scenario, model)
        response = await self._llm.structured_output(prompt, _TRACE_SCHEMA)

        data = json.loads(response.content)

        steps = tuple(
            TraceStep(
                context=s["context"],
                command_or_event=s["command_or_event"],
                description=s["description"],
                model_covers=s["model_covers"],
            )
            for s in data["steps"]
        )
        gaps = tuple(
            ModelGap(
                description=g["description"],
                gap_type=GapType(g["gap_type"]),
            )
            for g in data["gaps"]
        )
        return ScenarioTrace(scenario=scenario, steps=steps, gaps_found=gaps)

    @staticmethod
    def _build_generate_prompt(model: DomainModel, max_scenarios: int) -> str:
        """Build a prompt summarizing the DomainModel for scenario generation."""
        parts = ["Analyze this domain model and generate customer scenarios:\n"]

        parts.append("## Bounded Contexts")
        for ctx in model.bounded_contexts:
            classification = ctx.classification.value if ctx.classification else "unclassified"
            parts.append(f"- {ctx.name} ({classification}): {ctx.responsibility}")

        parts.append("\n## Aggregates")
        parts.extend(
            f"- {agg.name} in {agg.context_name}: "
            f"root={agg.root_entity}, invariants={len(agg.invariants)}, "
            f"commands={list(agg.commands)}, events={list(agg.domain_events)}"
            for agg in model.aggregate_designs
        )

        parts.append("\n## Domain Stories")
        parts.extend(
            f"- {story.name}: {' → '.join(story.steps)}"
            for story in model.domain_stories
        )

        parts.append(
            f"\nGenerate up to {max_scenarios} customer scenarios. "
            f"Types: {', '.join(st.value for st in ScenarioType)}. "
            f"Each scenario must be derived from model elements. "
            f"Only reference existing bounded contexts."
        )
        return "\n".join(parts)

    @staticmethod
    def _build_trace_prompt(scenario: Scenario, model: DomainModel) -> str:
        """Build a prompt for tracing a scenario against the model."""
        parts = ["Trace this scenario against the domain model:\n"]
        parts.append(f"Scenario: {scenario.narrative_text}")
        parts.append(f"Type: {scenario.scenario_type.value}")
        parts.append(f"Contexts: {', '.join(scenario.context_names)}")

        parts.append("\n## Model Aggregates")
        parts.extend(
            f"- {agg.name} ({agg.context_name}): "
            f"commands={list(agg.commands)}, events={list(agg.domain_events)}"
            for agg in model.aggregate_designs
        )

        parts.append(
            "\nFor each step, indicate if the model covers it (true/false/null). "
            "List any gaps found."
        )
        return "\n".join(parts)
