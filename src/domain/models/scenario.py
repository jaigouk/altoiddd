"""Scenario domain value objects for the Customer Simulator (Round 3).

ScenarioType classifies what a scenario probes: happy paths, failure modes,
edge cases, cross-context flows, or scale concerns.

Scenario, TraceStep, ModelGap, ScenarioTrace, and SimulationTally are frozen
VOs that capture the simulation lifecycle.
"""

from __future__ import annotations

import enum
from dataclasses import dataclass

from src.domain.models.errors import InvariantViolationError


class ScenarioType(enum.Enum):
    """Classification of what a scenario probes in the domain model."""

    HAPPY_PATH = "happy_path"
    FAILURE_MODE = "failure_mode"
    EDGE_CASE = "edge_case"
    CROSS_CONTEXT = "cross_context"
    SCALE = "scale"


class GapType(enum.Enum):
    """Classification of gaps found when tracing a scenario against the model."""

    MISSING_EVENT = "missing_event"
    MISSING_POLICY = "missing_policy"
    SILENT_ON_FAILURE = "silent_on_failure"
    MISSING_AGGREGATE = "missing_aggregate"
    AMBIGUOUS_OWNERSHIP = "ambiguous_ownership"


class ScenarioResult(enum.Enum):
    """Verdict for a traced scenario."""

    VALIDATED = "validated"
    MODEL_INCOMPLETE = "model_incomplete"
    OUT_OF_SCOPE = "out_of_scope"


@dataclass(frozen=True)
class Scenario:
    """A typed scenario that probes the domain model via customer simulation.

    Attributes:
        scenario_type: What category of scenario this represents.
        narrative_text: The scenario narrative (must not be empty).
        context_names: Which bounded contexts are involved (must not be empty).
        derived_from: Reference to the model element this was derived from.
    """

    scenario_type: ScenarioType
    narrative_text: str
    context_names: tuple[str, ...]
    derived_from: str

    def __post_init__(self) -> None:
        if not self.narrative_text.strip():
            msg = "Scenario narrative_text cannot be empty"
            raise InvariantViolationError(msg)
        if not self.context_names:
            msg = "Scenario context_names cannot be empty"
            raise InvariantViolationError(msg)
        if not self.derived_from.strip():
            msg = "Scenario derived_from cannot be empty"
            raise InvariantViolationError(msg)


@dataclass(frozen=True)
class TraceStep:
    """A single step in a scenario trace against the domain model.

    Attributes:
        context: Which bounded context this step occurs in.
        command_or_event: The command or event being exercised.
        description: Human-readable description of what happens.
        model_covers: True if covered, False if gap, None if unknown.
    """

    context: str
    command_or_event: str
    description: str
    model_covers: bool | None

    def __post_init__(self) -> None:
        if not self.context.strip():
            msg = "TraceStep context cannot be empty"
            raise InvariantViolationError(msg)
        if not self.command_or_event.strip():
            msg = "TraceStep command_or_event cannot be empty"
            raise InvariantViolationError(msg)


@dataclass(frozen=True)
class ModelGap:
    """A gap found when tracing a scenario against the domain model.

    Attributes:
        description: What is missing or ambiguous.
        gap_type: Classification of the gap.
    """

    description: str
    gap_type: GapType

    def __post_init__(self) -> None:
        if not self.description.strip():
            msg = "ModelGap description cannot be empty"
            raise InvariantViolationError(msg)


@dataclass(frozen=True)
class ScenarioTrace:
    """Result of tracing a scenario against the domain model.

    Attributes:
        scenario: The scenario that was traced.
        steps: Ordered trace steps.
        gaps_found: Gaps discovered during tracing.
    """

    scenario: Scenario
    steps: tuple[TraceStep, ...]
    gaps_found: tuple[ModelGap, ...]


@dataclass(frozen=True)
class SimulationTally:
    """Summary counts for a simulation run.

    Attributes:
        tested: Number of scenarios tested.
        gaps: Number of gaps found.
        deferred: Number of scenarios deferred (out of scope).
    """

    tested: int
    gaps: int
    deferred: int

    def __post_init__(self) -> None:
        if self.tested < 0:
            msg = "SimulationTally tested cannot be negative"
            raise InvariantViolationError(msg)
        if self.gaps < 0:
            msg = "SimulationTally gaps cannot be negative"
            raise InvariantViolationError(msg)
        if self.deferred < 0:
            msg = "SimulationTally deferred cannot be negative"
            raise InvariantViolationError(msg)
