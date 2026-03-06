"""Stateless domain service for rule-based challenge generation.

ChallengerService inspects a DomainModel aggregate and produces typed
Challenge value objects that probe for gaps: ambiguous language, missing
invariants, unexamined failure modes, and questionable boundaries.

All methods are static — pure input→output with no side effects.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.domain.models.challenge import Challenge, ChallengeType
from src.domain.models.domain_values import SubdomainClassification

if TYPE_CHECKING:
    from src.domain.models.domain_model import DomainModel


class ChallengerService:
    """Stateless domain service: generates challenges from a DomainModel."""

    @staticmethod
    def generate(
        model: DomainModel,
        max_per_type: int = 5,
    ) -> tuple[Challenge, ...]:
        """Generate typed challenges by inspecting the domain model.

        Args:
            model: The DomainModel aggregate to inspect.
            max_per_type: Maximum challenges per ChallengeType.

        Returns:
            Tuple of Challenge VOs, limited to max_per_type per category.
        """
        challenges: list[Challenge] = []
        challenges.extend(ChallengerService._language_challenges(model, max_per_type))
        challenges.extend(ChallengerService._invariant_challenges(model, max_per_type))
        challenges.extend(ChallengerService._failure_mode_challenges(model, max_per_type))
        challenges.extend(ChallengerService._boundary_challenges(model, max_per_type))
        return tuple(challenges)

    @staticmethod
    def _language_challenges(
        model: DomainModel,
        max_count: int,
    ) -> list[Challenge]:
        """Find ambiguous terms used across contexts without clear definitions."""
        ambiguous = model.ubiquitous_language.find_ambiguous_terms()
        challenges: list[Challenge] = []
        for term in ambiguous[:max_count]:
            # Find which contexts use this term
            entries = [
                e
                for e in model.ubiquitous_language.terms
                if e.term.lower() == term.lower()
            ]
            context_names = sorted({e.context_name for e in entries})
            if len(context_names) < 2:
                continue
            target_context = context_names[0]
            challenges.append(
                Challenge(
                    challenge_type=ChallengeType.LANGUAGE,
                    question_text=(
                        f"The term '{term}' appears in {', '.join(context_names)}. "
                        f"Does it mean the same thing in each context, or should each "
                        f"context have its own definition?"
                    ),
                    context_name=target_context,
                    source_reference=f"UL glossary: '{term}' in {', '.join(context_names)}",
                )
            )
        return challenges

    @staticmethod
    def _invariant_challenges(
        model: DomainModel,
        max_count: int,
    ) -> list[Challenge]:
        """Find Core/Supporting aggregates with no invariants listed."""
        challenges: list[Challenge] = []
        core_supporting = {
            ctx.name
            for ctx in model.bounded_contexts
            if ctx.classification
            in (SubdomainClassification.CORE, SubdomainClassification.SUPPORTING)
        }
        for agg in model.aggregate_designs:
            if len(challenges) >= max_count:
                break
            if agg.context_name not in core_supporting:
                continue
            if not agg.invariants:
                challenges.append(
                    Challenge(
                        challenge_type=ChallengeType.INVARIANT,
                        question_text=(
                            f"Aggregate '{agg.name}' in {agg.context_name} has no "
                            f"invariants listed. What business rules must this "
                            f"aggregate protect?"
                        ),
                        context_name=agg.context_name,
                        source_reference=f"Aggregate design: {agg.name}",
                    )
                )
        return challenges

    @staticmethod
    def _failure_mode_challenges(
        model: DomainModel,
        max_count: int,
    ) -> list[Challenge]:
        """Probe Core domain story steps for unexamined failure paths."""
        core_context_names = {
            ctx.name
            for ctx in model.bounded_contexts
            if ctx.classification == SubdomainClassification.CORE
        }
        if not core_context_names:
            return []

        challenges: list[Challenge] = []
        for story in model.domain_stories:
            if len(challenges) >= max_count:
                break
            for step in story.steps:
                if len(challenges) >= max_count:
                    break
                # Pick the first Core context for attribution
                target_context = sorted(core_context_names)[0]
                challenges.append(
                    Challenge(
                        challenge_type=ChallengeType.FAILURE_MODE,
                        question_text=(
                            f"In story '{story.name}', what happens if this step "
                            f"fails: '{step}'?"
                        ),
                        context_name=target_context,
                        source_reference=f"Domain story: {story.name}",
                    )
                )
        return challenges

    @staticmethod
    def _boundary_challenges(
        model: DomainModel,
        max_count: int,
    ) -> list[Challenge]:
        """Question context boundaries when >= 2 contexts exist."""
        contexts = model.bounded_contexts
        if len(contexts) < 2:
            return []

        challenges: list[Challenge] = []
        for rel in model.context_relationships:
            if len(challenges) >= max_count:
                break
            # Find the downstream context
            downstream_ctx = next(
                (c for c in contexts if c.name == rel.downstream),
                None,
            )
            if downstream_ctx is None:
                continue
            challenges.append(
                Challenge(
                    challenge_type=ChallengeType.BOUNDARY,
                    question_text=(
                        f"Context '{rel.downstream}' depends on '{rel.upstream}' "
                        f"via {rel.integration_pattern}. Could "
                        f"'{rel.downstream}' own this data directly instead?"
                    ),
                    context_name=rel.downstream,
                    source_reference=(
                        f"Context map: {rel.upstream} → {rel.downstream}"
                    ),
                )
            )
        return challenges
