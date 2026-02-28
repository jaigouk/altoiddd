"""Question entity and catalog for the 10-question DDD discovery flow.

Each Question holds dual-register text (technical and non-technical) and
declares what domain artifacts it produces. The CATALOG class constant
contains all 10 questions; MVP_QUESTION_IDS identifies the minimum viable
subset for quick mode.
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import ClassVar

from src.domain.models.discovery_values import QuestionPhase


@dataclass(frozen=True)
class Question:
    """A single discovery question with dual-register phrasing.

    Attributes:
        id: Question identifier (Q1-Q10).
        phase: Which discovery phase this question belongs to.
        technical_text: DDD/engineering phrasing of the question.
        non_technical_text: Plain-language phrasing of the question.
        produces: Domain artifacts this question helps discover.
    """

    id: str
    phase: QuestionPhase
    technical_text: str
    non_technical_text: str
    produces: tuple[str, ...]

    # Class-level constants assigned after class definition (self-referencing).
    CATALOG: ClassVar[tuple[Question, ...]]
    MVP_QUESTION_IDS: ClassVar[frozenset[str]]


# Assign class-level constants after the class is fully defined.
Question.CATALOG = (
    Question(
        id="Q1",
        phase=QuestionPhase.ACTORS,
        technical_text=(
            "Who are the actors (users, external systems) that interact with your system?"
        ),
        non_technical_text=("Who will use this product, and what other systems does it talk to?"),
        produces=("actors", "external_systems"),
    ),
    Question(
        id="Q2",
        phase=QuestionPhase.ACTORS,
        technical_text="What are the core entities (nouns) in your domain?",
        non_technical_text="What are the main things or concepts your product deals with?",
        produces=("entities", "value_objects"),
    ),
    Question(
        id="Q3",
        phase=QuestionPhase.STORY,
        technical_text=(
            "Describe the primary use case as a domain story: "
            "actor -> command -> event -> outcome."
        ),
        non_technical_text=("Walk me through the most important thing a user does, step by step."),
        produces=("commands", "events", "domain_story"),
    ),
    Question(
        id="Q4",
        phase=QuestionPhase.STORY,
        technical_text="What is the most critical failure mode? What invariants must hold?",
        non_technical_text=(
            "What could go wrong that would be a serious problem? What rules must never be broken?"
        ),
        produces=("invariants", "failure_modes"),
    ),
    Question(
        id="Q5",
        phase=QuestionPhase.STORY,
        technical_text="What other workflows or use cases exist beyond the primary one?",
        non_technical_text="What else can users do with the product besides the main thing?",
        produces=("secondary_stories", "commands"),
    ),
    Question(
        id="Q6",
        phase=QuestionPhase.EVENTS,
        technical_text="What domain events are published when state changes occur?",
        non_technical_text=(
            "What important things happen in the system that other parts need to know about?"
        ),
        produces=("domain_events",),
    ),
    Question(
        id="Q7",
        phase=QuestionPhase.EVENTS,
        technical_text="What policies (event -> command reactions) exist in the system?",
        non_technical_text="When something happens, what automatic actions should follow?",
        produces=("policies", "reactions"),
    ),
    Question(
        id="Q8",
        phase=QuestionPhase.EVENTS,
        technical_text="What read models or projections does the system need?",
        non_technical_text="What views or reports do users need to see?",
        produces=("read_models", "projections"),
    ),
    Question(
        id="Q9",
        phase=QuestionPhase.BOUNDARIES,
        technical_text="How would you partition the domain into bounded contexts?",
        non_technical_text=(
            "If you split the product into independent teams, what would each team own?"
        ),
        produces=("bounded_contexts",),
    ),
    Question(
        id="Q10",
        phase=QuestionPhase.BOUNDARIES,
        technical_text=(
            "Classify each context: core (competitive advantage), "
            "supporting (necessary but not differentiating), or generic (commodity)."
        ),
        non_technical_text=(
            "Which parts are your secret sauce, which are necessary plumbing, "
            "and which are off-the-shelf?"
        ),
        produces=("subdomain_classification",),
    ),
)

Question.MVP_QUESTION_IDS = frozenset({"Q1", "Q3", "Q4", "Q9", "Q10"})
