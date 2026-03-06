"""Challenge domain value objects for the AI Challenger (Round 2).

ChallengeType classifies what a challenge probes: language ambiguity,
missing invariants, failure modes, boundary disputes, aggregate gaps,
or communication patterns.

Challenge, ChallengeResponse, and ChallengeIteration are frozen VOs
that capture the challenge round lifecycle.
"""

from __future__ import annotations

import enum
from dataclasses import dataclass

from src.domain.models.errors import InvariantViolationError


class ChallengeType(enum.Enum):
    """Classification of what a challenge probes in the domain model.

    LANGUAGE:       Ambiguous terms used across contexts without per-context definitions.
    INVARIANT:      Missing business rules on aggregates.
    FAILURE_MODE:   Unexamined failure paths in domain stories.
    BOUNDARY:       Questionable bounded context responsibilities.
    AGGREGATE:      Aggregate design gaps (missing entities, wrong root).
    COMMUNICATION:  Unclear inter-context integration patterns.
    """

    LANGUAGE = "language"
    INVARIANT = "invariant"
    FAILURE_MODE = "failure_mode"
    BOUNDARY = "boundary"
    AGGREGATE = "aggregate"
    COMMUNICATION = "communication"


@dataclass(frozen=True)
class Challenge:
    """A typed question that probes the domain model for gaps.

    Follows CHALLENGE-AS-QUESTION pattern: always a question, never a fact.

    Attributes:
        challenge_type: What category of gap this probes.
        question_text: The challenge question (must not be empty).
        context_name: Which bounded context this targets (must not be empty).
        source_reference: Evidence or citation backing this challenge (must not be empty).
        evidence: Optional supporting detail.
    """

    challenge_type: ChallengeType
    question_text: str
    context_name: str
    source_reference: str
    evidence: str = ""

    def __post_init__(self) -> None:
        if not self.question_text.strip():
            msg = "Challenge question_text cannot be empty"
            raise InvariantViolationError(msg)
        if not self.context_name.strip():
            msg = "Challenge context_name cannot be empty"
            raise InvariantViolationError(msg)
        if not self.source_reference.strip():
            msg = "Challenge source_reference cannot be empty"
            raise InvariantViolationError(msg)


@dataclass(frozen=True)
class ChallengeResponse:
    """User's response to a single challenge.

    Attributes:
        challenge_id: Identifier linking back to the challenge.
        user_response: What the user said.
        accepted: Whether the user accepted the challenge's premise.
        artifact_updates: DDD.md changes prompted by this response.
    """

    challenge_id: str
    user_response: str
    accepted: bool
    artifact_updates: tuple[str, ...] = ()


@dataclass(frozen=True)
class ChallengeIteration:
    """A complete challenge round: all challenges posed and responses received.

    Attributes:
        challenges: All challenges generated for this iteration.
        responses: All user responses collected.
        convergence_delta: Count of model changes (invariants, terms, stories added).
    """

    challenges: tuple[Challenge, ...]
    responses: tuple[ChallengeResponse, ...]
    convergence_delta: int
