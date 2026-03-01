"""Domain events for the Domain Model bounded context.

DomainModelGenerated is emitted when a DomainModel aggregate passes all
invariant checks via finalize(). It carries the complete set of DDD
artifacts for downstream consumers (Architecture Testing, Ticket Pipeline,
Tool Translation).
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from src.domain.models.domain_values import (
        AggregateDesign,
        BoundedContext,
        ContextRelationship,
        DomainStory,
    )
    from src.domain.models.ubiquitous_language import TermEntry


@dataclass(frozen=True)
class DomainModelGenerated:
    """Emitted when a DomainModel passes all invariant checks.

    Attributes:
        model_id: Unique identifier of the domain model.
        domain_stories: All captured domain stories.
        ubiquitous_language: All glossary terms.
        bounded_contexts: All identified bounded contexts with classifications.
        context_relationships: All relationships between contexts.
        aggregate_designs: Aggregate designs for Core subdomains.
    """

    model_id: str
    domain_stories: tuple[DomainStory, ...]
    ubiquitous_language: tuple[TermEntry, ...]
    bounded_contexts: tuple[BoundedContext, ...]
    context_relationships: tuple[ContextRelationship, ...]
    aggregate_designs: tuple[AggregateDesign, ...]
