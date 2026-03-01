"""DomainModel aggregate root for the Domain Model bounded context.

Manages the complete set of DDD artifacts for a project: domain stories,
ubiquitous language, bounded contexts, context relationships, and aggregate
designs. Enforces 4 invariants on finalize().

Invariants (from DDD.md §5):
1. Every UL term must appear in at least one DomainStory (word boundary match).
2. Every BoundedContext must have a SubdomainClassification.
3. Every Core subdomain must have at least one AggregateDesign.
4. Ambiguous terms must have per-context definitions.
"""

from __future__ import annotations

import re
import uuid
from typing import TYPE_CHECKING

from src.domain.models.domain_values import (
    AggregateDesign,
    BoundedContext,
    ContextRelationship,
    DomainStory,
    SubdomainClassification,
)
from src.domain.models.errors import DuplicateStoryError, InvariantViolationError
from src.domain.models.ubiquitous_language import UbiquitousLanguage

if TYPE_CHECKING:
    from src.domain.events.domain_events import DomainModelGenerated


class DomainModel:
    """Aggregate root: the complete DDD artifact set for a project.

    Attributes:
        model_id: Unique identifier for this domain model.
    """

    def __init__(self) -> None:
        self.model_id: str = str(uuid.uuid4())
        self._stories: list[DomainStory] = []
        self._language: UbiquitousLanguage = UbiquitousLanguage()
        self._contexts: list[BoundedContext] = []
        self._relationships: list[ContextRelationship] = []
        self._aggregates: list[AggregateDesign] = []
        self._events: list[DomainModelGenerated] = []

    # -- Properties -----------------------------------------------------------

    @property
    def domain_stories(self) -> tuple[DomainStory, ...]:
        """All domain stories (defensive copy)."""
        return tuple(self._stories)

    @property
    def ubiquitous_language(self) -> UbiquitousLanguage:
        """The ubiquitous language glossary."""
        return self._language

    @property
    def bounded_contexts(self) -> tuple[BoundedContext, ...]:
        """All bounded contexts (defensive copy)."""
        return tuple(self._contexts)

    @property
    def context_relationships(self) -> tuple[ContextRelationship, ...]:
        """All context relationships (defensive copy)."""
        return tuple(self._relationships)

    @property
    def aggregate_designs(self) -> tuple[AggregateDesign, ...]:
        """All aggregate designs (defensive copy)."""
        return tuple(self._aggregates)

    @property
    def events(self) -> tuple[DomainModelGenerated, ...]:
        """Domain events produced by this aggregate (defensive copy)."""
        return tuple(self._events)

    # -- Commands -------------------------------------------------------------

    def add_domain_story(self, story: DomainStory) -> None:
        """Add a business process narrative.

        Args:
            story: The domain story to add.

        Raises:
            DuplicateStoryError: If a story with the same name already exists.
            ValueError: If story name is empty.
        """
        if not story.name.strip():
            msg = "Story name cannot be empty"
            raise ValueError(msg)

        existing_names = {s.name.lower() for s in self._stories}
        if story.name.lower() in existing_names:
            msg = f"Domain story '{story.name}' already exists"
            raise DuplicateStoryError(msg)

        self._stories.append(story)

    def add_term(
        self,
        term: str,
        definition: str,
        context_name: str,
        source_question_ids: tuple[str, ...] = (),
    ) -> None:
        """Add a term to the ubiquitous language glossary.

        Args:
            term: The domain term.
            definition: What it means in this context.
            context_name: Which bounded context it belongs to.
            source_question_ids: Which questions surfaced this term.
        """
        self._language.add_term(term, definition, context_name, source_question_ids)

    def reassign_terms_to_context(
        self,
        from_context: str,
        to_context: str,
    ) -> None:
        """Move all terms from one context to another.

        Args:
            from_context: Source context name (case-insensitive match).
            to_context: Target context name.

        Raises:
            ValueError: If either context name is empty.
        """
        if not from_context.strip():
            msg = "Source context name cannot be empty"
            raise ValueError(msg)
        if not to_context.strip():
            msg = "Target context name cannot be empty"
            raise ValueError(msg)

        from src.domain.models.ubiquitous_language import TermEntry

        from_lower = from_context.lower()
        new_terms: list[TermEntry] = []
        for t in self._language._terms:
            if t.context_name.lower() == from_lower:
                new_terms.append(
                    TermEntry(
                        term=t.term,
                        definition=t.definition,
                        context_name=to_context,
                        source_question_ids=t.source_question_ids,
                    )
                )
            else:
                new_terms.append(t)
        self._language._terms = new_terms

    def add_bounded_context(self, context: BoundedContext) -> None:
        """Add a bounded context.

        Args:
            context: The bounded context to add.

        Raises:
            ValueError: If context name is empty or duplicate.
        """
        if not context.name.strip():
            msg = "Context name cannot be empty"
            raise ValueError(msg)

        existing_names = {c.name.lower() for c in self._contexts}
        if context.name.lower() in existing_names:
            msg = f"Bounded context '{context.name}' already exists"
            raise ValueError(msg)

        self._contexts.append(context)

    def classify_subdomain(
        self,
        context_name: str,
        classification: SubdomainClassification,
        rationale: str = "",
    ) -> None:
        """Classify a bounded context's subdomain type.

        Replaces the context entry with an updated classification.

        Args:
            context_name: Name of the bounded context to classify.
            classification: Core, Supporting, or Generic.
            rationale: Why this classification was chosen.

        Raises:
            ValueError: If the context does not exist.
        """
        for i, ctx in enumerate(self._contexts):
            if ctx.name.lower() == context_name.lower():
                self._contexts[i] = BoundedContext(
                    name=ctx.name,
                    responsibility=ctx.responsibility,
                    key_domain_objects=ctx.key_domain_objects,
                    classification=classification,
                    classification_rationale=rationale,
                )
                return

        msg = f"Bounded context '{context_name}' not found"
        raise ValueError(msg)

    def add_context_relationship(self, relationship: ContextRelationship) -> None:
        """Add a relationship between two bounded contexts.

        Args:
            relationship: The context relationship to add.

        Raises:
            ValueError: If upstream or downstream context name is empty.
        """
        if not relationship.upstream.strip() or not relationship.downstream.strip():
            msg = "Relationship upstream and downstream cannot be empty"
            raise ValueError(msg)
        self._relationships.append(relationship)

    def design_aggregate(self, aggregate: AggregateDesign) -> None:
        """Add an aggregate design for a bounded context.

        Args:
            aggregate: The aggregate design to add.

        Raises:
            ValueError: If aggregate name or context name is empty,
                or if a duplicate exists in the same context.
        """
        if not aggregate.name.strip():
            msg = "Aggregate name cannot be empty"
            raise ValueError(msg)
        if not aggregate.context_name.strip():
            msg = "Aggregate context name cannot be empty"
            raise ValueError(msg)

        existing = {(a.name.lower(), a.context_name.lower()) for a in self._aggregates}
        key = (aggregate.name.lower(), aggregate.context_name.lower())
        if key in existing:
            msg = (
                f"Aggregate design '{aggregate.name}' already exists"
                f" in context '{aggregate.context_name}'"
            )
            raise ValueError(msg)

        self._aggregates.append(aggregate)

    def finalize(self) -> None:
        """Validate all invariants and emit DomainModelGenerated.

        Raises:
            InvariantViolationError: If any invariant is violated.
        """
        self._check_terms_in_stories()
        self._check_context_classifications()
        self._check_core_aggregates()
        self._check_ambiguous_terms()

        from src.domain.events.domain_events import DomainModelGenerated

        self._events.append(
            DomainModelGenerated(
                model_id=self.model_id,
                domain_stories=tuple(self._stories),
                ubiquitous_language=self._language.terms,
                bounded_contexts=tuple(self._contexts),
                context_relationships=tuple(self._relationships),
                aggregate_designs=tuple(self._aggregates),
            )
        )

    # -- Invariant checks (private) -------------------------------------------

    def _check_terms_in_stories(self) -> None:
        """Invariant 1: Every UL term must appear in at least one DomainStory.

        Uses word boundary matching to prevent false positives from
        substring matches (e.g. "Or" should not match "Order").
        """
        story_text_parts: list[str] = []
        for story in self._stories:
            story_text_parts.append(story.name.lower())
            story_text_parts.extend(a.lower() for a in story.actors)
            story_text_parts.append(story.trigger.lower())
            story_text_parts.extend(s.lower() for s in story.steps)
            story_text_parts.extend(o.lower() for o in story.observations)
        story_text = " ".join(story_text_parts)

        for entry in self._language.terms:
            pattern = r"\b" + re.escape(entry.term.lower()) + r"\b"
            if not re.search(pattern, story_text):
                msg = f"Term '{entry.term}' not found in any domain story"
                raise InvariantViolationError(msg)

    def _check_context_classifications(self) -> None:
        """Invariant 2: Every BoundedContext must have a SubdomainClassification."""
        for ctx in self._contexts:
            if ctx.classification is None:
                msg = f"BoundedContext '{ctx.name}' has no classification"
                raise InvariantViolationError(msg)

    def _check_core_aggregates(self) -> None:
        """Invariant 3: Every Core subdomain must have at least one AggregateDesign."""
        core_contexts = {
            ctx.name
            for ctx in self._contexts
            if ctx.classification == SubdomainClassification.CORE
        }
        contexts_with_aggregates = {a.context_name for a in self._aggregates}

        for core_name in core_contexts:
            if core_name not in contexts_with_aggregates:
                msg = f"Core subdomain '{core_name}' has no aggregate design"
                raise InvariantViolationError(msg)

    def _check_ambiguous_terms(self) -> None:
        """Invariant 4: Ambiguous terms must have per-context definitions."""
        ambiguous = self._language.find_ambiguous_terms()
        for term in ambiguous:
            if not self._language.has_per_context_definitions(term):
                msg = f"Ambiguous term '{term}' needs per-context definitions"
                raise InvariantViolationError(msg)
