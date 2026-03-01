"""Value objects for the Domain Model bounded context.

SubdomainClassification, DomainStory, BoundedContext, ContextRelationship,
and AggregateDesign are frozen dataclasses representing DDD artifacts
produced by the artifact generation pipeline.
"""

from __future__ import annotations

import enum
from dataclasses import dataclass


class SubdomainClassification(enum.Enum):
    """Classification of a subdomain per Khononov's decision tree.

    CORE:       Competitive advantage — full DDD treatment.
    SUPPORTING: Necessary but not differentiating — simpler architecture.
    GENERIC:    Commodity — buy or use existing library.
    """

    CORE = "core"
    SUPPORTING = "supporting"
    GENERIC = "generic"


@dataclass(frozen=True)
class DomainStory:
    """A business process narrative using domain language.

    Attributes:
        name: Short descriptive name for the story.
        actors: Who is involved (domain terms, not "User").
        trigger: What starts this process.
        steps: Ordered steps as "[Actor] [verb] [work object]" sentences.
        observations: Notable findings during story capture.
    """

    name: str
    actors: tuple[str, ...]
    trigger: str
    steps: tuple[str, ...]
    observations: tuple[str, ...] = ()


@dataclass(frozen=True)
class BoundedContext:
    """An explicit boundary around a domain model.

    Attributes:
        name: Context name matching ubiquitous language.
        responsibility: What this context owns and manages.
        key_domain_objects: Entities, VOs, aggregates within this context.
        classification: Subdomain type (Core/Supporting/Generic), or None if unclassified.
        classification_rationale: Why this classification was chosen.
    """

    name: str
    responsibility: str
    key_domain_objects: tuple[str, ...] = ()
    classification: SubdomainClassification | None = None
    classification_rationale: str = ""


@dataclass(frozen=True)
class ContextRelationship:
    """A relationship between two bounded contexts.

    Attributes:
        upstream: The context that provides data/events.
        downstream: The context that consumes data/events.
        integration_pattern: How they communicate (e.g. "Domain Events",
            "Anticorruption Layer", "Shared Kernel").
    """

    upstream: str
    downstream: str
    integration_pattern: str


@dataclass(frozen=True)
class AggregateDesign:
    """Aggregate design for a Core subdomain bounded context.

    Attributes:
        name: Aggregate name.
        context_name: Which bounded context this belongs to.
        root_entity: The aggregate root entity name.
        contained_objects: Entities and VOs within this aggregate.
        invariants: Business rules this aggregate protects.
        commands: Actions this aggregate can perform.
        domain_events: Events this aggregate announces.
    """

    name: str
    context_name: str
    root_entity: str
    contained_objects: tuple[str, ...] = ()
    invariants: tuple[str, ...] = ()
    commands: tuple[str, ...] = ()
    domain_events: tuple[str, ...] = ()
