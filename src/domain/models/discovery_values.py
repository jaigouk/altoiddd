"""Value objects for the Guided Discovery bounded context.

Persona, Register, and QuestionPhase are enums representing the user's
role, language register, and discovery flow phase. Answer and Playback
are frozen dataclass value objects capturing responses and confirmation state.
"""

from __future__ import annotations

import enum
from dataclasses import dataclass


class Persona(enum.Enum):
    """The user's self-identified role during discovery.

    DEVELOPER:     Technical lead / developer (register: TECHNICAL).
    PRODUCT_OWNER: Product owner / business stakeholder.
    DOMAIN_EXPERT: Subject-matter expert with deep domain knowledge.
    MIXED:         Unsure or wearing multiple hats.
    """

    DEVELOPER = "developer"
    PRODUCT_OWNER = "product_owner"
    DOMAIN_EXPERT = "domain_expert"
    MIXED = "mixed"


class Register(enum.Enum):
    """Language register for question phrasing.

    TECHNICAL:     Uses DDD/engineering terminology.
    NON_TECHNICAL: Uses plain-language business terminology.
    """

    TECHNICAL = "technical"
    NON_TECHNICAL = "non_technical"


class QuestionPhase(enum.Enum):
    """Phases of the 10-question DDD discovery flow.

    SEED:       AI reads README, extracts initial candidates (0 questions).
    ACTORS:     Who are the actors and entities (Q1-Q2).
    STORY:      Primary use case, failure mode, other workflows (Q3-Q5).
    EVENTS:     Domain events, policies, read models (Q6-Q8).
    BOUNDARIES: Bounded contexts and subdomain classification (Q9-Q10).
    """

    SEED = "seed"
    ACTORS = "actors"
    STORY = "story"
    EVENTS = "events"
    BOUNDARIES = "boundaries"


@dataclass(frozen=True)
class Answer:
    """A user's response to a single discovery question.

    Attributes:
        question_id: Identifier of the question (e.g. "Q1").
        response_text: The user's free-text answer.
    """

    question_id: str
    response_text: str


@dataclass(frozen=True)
class Playback:
    """A playback summary shown to the user for confirmation.

    Attributes:
        summary_text: The generated summary of recent answers.
        confirmed: Whether the user confirmed the playback.
        corrections: User-provided corrections (empty if confirmed).
    """

    summary_text: str
    confirmed: bool = False
    corrections: str = ""
