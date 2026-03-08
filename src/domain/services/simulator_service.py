"""Stateless domain service for rule-based scenario generation and tracing.

SimulatorService inspects a DomainModel aggregate and produces typed
Scenario value objects for customer simulation (Round 3), then traces
scenarios against the model to find gaps.

All methods are static — pure input→output with no side effects.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.domain.models.domain_values import SubdomainClassification
from src.domain.models.scenario import (
    GapType,
    ModelGap,
    Scenario,
    ScenarioTrace,
    ScenarioType,
    TraceStep,
)

if TYPE_CHECKING:
    from src.domain.models.domain_model import DomainModel


class SimulatorService:
    """Stateless domain service: generates scenarios and traces them against a DomainModel."""

    @staticmethod
    def generate(
        model: DomainModel,
        max_scenarios: int = 5,
    ) -> tuple[Scenario, ...]:
        """Generate typed scenarios by inspecting the domain model.

        Args:
            model: The DomainModel aggregate to inspect.
            max_scenarios: Maximum total scenarios to return.

        Returns:
            Tuple of Scenario VOs, capped at max_scenarios.
        """
        scenarios: list[Scenario] = []
        scenarios.extend(SimulatorService._happy_path_scenarios(model))
        scenarios.extend(SimulatorService._failure_mode_scenarios(model))
        scenarios.extend(SimulatorService._edge_case_scenarios(model))
        scenarios.extend(SimulatorService._cross_context_scenarios(model))
        return tuple(scenarios[:max_scenarios])

    @staticmethod
    def trace(
        scenario: Scenario,
        model: DomainModel,
    ) -> ScenarioTrace:
        """Trace a scenario against the domain model to find coverage and gaps.

        Args:
            scenario: The scenario to trace.
            model: The DomainModel to trace against.

        Returns:
            ScenarioTrace with steps and gaps found.
        """
        steps: list[TraceStep] = []
        gaps: list[ModelGap] = []

        context_aggregates = {
            ctx.name: [a for a in model.aggregate_designs if a.context_name == ctx.name]
            for ctx in model.bounded_contexts
        }

        for ctx_name in scenario.context_names:
            aggs = context_aggregates.get(ctx_name, [])
            if not aggs:
                gaps.append(
                    ModelGap(
                        description=f"No aggregates defined in context '{ctx_name}'",
                        gap_type=GapType.MISSING_AGGREGATE,
                    )
                )
                steps.append(
                    TraceStep(
                        context=ctx_name,
                        command_or_event="(none)",
                        description=f"No aggregate found in {ctx_name}",
                        model_covers=False,
                    )
                )
                continue

            for agg in aggs:
                steps.extend(
                    TraceStep(
                        context=ctx_name,
                        command_or_event=cmd,
                        description=f"Command '{cmd}' on {agg.name}",
                        model_covers=True,
                    )
                    for cmd in agg.commands
                )
                steps.extend(
                    TraceStep(
                        context=ctx_name,
                        command_or_event=evt,
                        description=f"Event '{evt}' from {agg.name}",
                        model_covers=True,
                    )
                    for evt in agg.domain_events
                )
                if not agg.commands and not agg.domain_events:
                    gaps.append(
                        ModelGap(
                            description=(
                                f"Aggregate '{agg.name}' in {ctx_name} has no "
                                f"commands or events"
                            ),
                            gap_type=GapType.SILENT_ON_FAILURE,
                        )
                    )
                    steps.append(
                        TraceStep(
                            context=ctx_name,
                            command_or_event=agg.name,
                            description=f"Aggregate '{agg.name}' has no commands/events",
                            model_covers=False,
                        )
                    )

        return ScenarioTrace(
            scenario=scenario,
            steps=tuple(steps),
            gaps_found=tuple(gaps),
        )

    @staticmethod
    def _happy_path_scenarios(model: DomainModel) -> list[Scenario]:
        """One happy path per domain story touching a Core context."""
        core_names = {
            ctx.name
            for ctx in model.bounded_contexts
            if ctx.classification == SubdomainClassification.CORE
        }
        if not core_names:
            return []

        return [
            Scenario(
                scenario_type=ScenarioType.HAPPY_PATH,
                narrative_text=(
                    f"Happy path: {story.name} completes successfully"
                ),
                context_names=tuple(sorted(core_names)),
                derived_from=f"Domain story: {story.name}",
            )
            for story in model.domain_stories
        ]

    @staticmethod
    def _failure_mode_scenarios(model: DomainModel) -> list[Scenario]:
        """One failure mode per domain story workflow."""
        core_names = {
            ctx.name
            for ctx in model.bounded_contexts
            if ctx.classification == SubdomainClassification.CORE
        }
        if not core_names:
            return []

        scenarios: list[Scenario] = []
        for story in model.domain_stories:
            if story.steps:
                last_step = story.steps[-1]
                scenarios.append(
                    Scenario(
                        scenario_type=ScenarioType.FAILURE_MODE,
                        narrative_text=(
                            f"Failure mode: what if '{last_step}' fails in {story.name}?"
                        ),
                        context_names=tuple(sorted(core_names)),
                        derived_from=f"Domain story: {story.name}",
                    )
                )
        return scenarios

    @staticmethod
    def _edge_case_scenarios(model: DomainModel) -> list[Scenario]:
        """Edge cases from invariants (boundary probing)."""
        scenarios: list[Scenario] = []
        for agg in model.aggregate_designs:
            scenarios.extend(
                Scenario(
                    scenario_type=ScenarioType.EDGE_CASE,
                    narrative_text=(
                        f"Edge case: probe boundary of '{invariant}' "
                        f"on {agg.name}"
                    ),
                    context_names=(agg.context_name,),
                    derived_from=f"Invariant: {invariant}",
                )
                for invariant in agg.invariants
            )
        return scenarios

    @staticmethod
    def _cross_context_scenarios(model: DomainModel) -> list[Scenario]:
        """One cross-context scenario per context relationship."""
        return [
            Scenario(
                scenario_type=ScenarioType.CROSS_CONTEXT,
                narrative_text=(
                    f"Cross-context: data flows from {rel.upstream} to "
                    f"{rel.downstream} via {rel.integration_pattern}"
                ),
                context_names=(rel.upstream, rel.downstream),
                derived_from=(
                    f"Context map: {rel.upstream} → {rel.downstream}"
                ),
            )
            for rel in model.context_relationships
        ]
