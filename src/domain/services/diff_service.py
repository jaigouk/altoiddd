"""Stateless domain service for computing diffs between DomainModel snapshots.

DiffService compares two DomainModel objects field-by-field and produces
an ArtifactDiff with typed DiffEntry items and a ConvergenceMetric.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.domain.models.artifact_diff import (
    ArtifactDiff,
    ConvergenceMetric,
    DiffEntry,
    DiffType,
)
from src.domain.models.research import TrustLevel

if TYPE_CHECKING:
    from src.domain.models.domain_model import DomainModel
    from src.domain.models.domain_values import AggregateDesign


class DiffService:
    """Stateless domain service: computes diff between two DomainModel snapshots."""

    @staticmethod
    def compute(
        before: DomainModel,
        after: DomainModel,
        from_version: int = 1,
        to_version: int = 2,
    ) -> ArtifactDiff:
        """Compare two DomainModel objects and return an ArtifactDiff.

        Args:
            before: The earlier DomainModel snapshot.
            after: The later DomainModel snapshot.
            from_version: Version number of the earlier snapshot.
            to_version: Version number of the later snapshot.

        Returns:
            ArtifactDiff with entries and convergence metric.
        """
        entries: list[DiffEntry] = []

        ctx_entries = DiffService._diff_bounded_contexts(before, after)
        entries.extend(ctx_entries)

        story_entries = DiffService._diff_domain_stories(before, after)
        entries.extend(story_entries)

        term_entries = DiffService._diff_ubiquitous_language(before, after)
        entries.extend(term_entries)

        agg_entries = DiffService._diff_aggregate_designs(before, after)
        entries.extend(agg_entries)

        convergence = ConvergenceMetric(
            invariants_delta=sum(
                1
                for e in agg_entries
                if e.diff_type in (DiffType.ADDED, DiffType.MODIFIED, DiffType.REMOVED)
            ),
            terms_delta=sum(
                1
                for e in term_entries
                if e.diff_type
                in (DiffType.ADDED, DiffType.MODIFIED, DiffType.REMOVED, DiffType.DISAMBIGUATED)
            ),
            stories_delta=sum(
                1
                for e in story_entries
                if e.diff_type in (DiffType.ADDED, DiffType.MODIFIED, DiffType.REMOVED)
            ),
            canvases_delta=sum(
                1
                for e in ctx_entries
                if e.diff_type in (DiffType.ADDED, DiffType.MODIFIED, DiffType.REMOVED)
            ),
        )

        return ArtifactDiff(
            from_version=from_version,
            to_version=to_version,
            entries=tuple(entries),
            convergence=convergence,
        )

    @staticmethod
    def _diff_bounded_contexts(
        before: DomainModel, after: DomainModel
    ) -> list[DiffEntry]:
        before_by_name = {c.name.lower(): c for c in before.bounded_contexts}
        after_by_name = {c.name.lower(): c for c in after.bounded_contexts}

        entries: list[DiffEntry] = []

        for name_lower, ctx in after_by_name.items():
            if name_lower not in before_by_name:
                entries.append(
                    DiffEntry(
                        diff_type=DiffType.ADDED,
                        section="Bounded Contexts",
                        description=f"Added bounded context '{ctx.name}'",
                        provenance=TrustLevel.AI_INFERRED,
                    )
                )
            else:
                old = before_by_name[name_lower]
                if old.classification != ctx.classification:
                    entries.append(
                        DiffEntry(
                            diff_type=DiffType.MODIFIED,
                            section="Bounded Contexts",
                            description=(
                                f"Classification of '{ctx.name}' changed "
                                f"from {old.classification} to {ctx.classification}"
                            ),
                            provenance=TrustLevel.AI_INFERRED,
                        )
                    )

        for name_lower, ctx in before_by_name.items():
            if name_lower not in after_by_name:
                entries.append(
                    DiffEntry(
                        diff_type=DiffType.REMOVED,
                        section="Bounded Contexts",
                        description=f"Removed bounded context '{ctx.name}'",
                        provenance=TrustLevel.AI_INFERRED,
                    )
                )

        return entries

    @staticmethod
    def _diff_domain_stories(
        before: DomainModel, after: DomainModel
    ) -> list[DiffEntry]:
        before_by_name = {s.name.lower(): s for s in before.domain_stories}
        after_by_name = {s.name.lower(): s for s in after.domain_stories}

        entries: list[DiffEntry] = []

        for name_lower, story in after_by_name.items():
            if name_lower not in before_by_name:
                entries.append(
                    DiffEntry(
                        diff_type=DiffType.ADDED,
                        section="Domain Stories",
                        description=f"Added domain story '{story.name}'",
                        provenance=TrustLevel.AI_INFERRED,
                    )
                )
            else:
                old = before_by_name[name_lower]
                if old.steps != story.steps or old.actors != story.actors:
                    entries.append(
                        DiffEntry(
                            diff_type=DiffType.MODIFIED,
                            section="Domain Stories",
                            description=f"Modified domain story '{story.name}'",
                            provenance=TrustLevel.AI_INFERRED,
                        )
                    )

        for name_lower, story in before_by_name.items():
            if name_lower not in after_by_name:
                entries.append(
                    DiffEntry(
                        diff_type=DiffType.REMOVED,
                        section="Domain Stories",
                        description=f"Removed domain story '{story.name}'",
                        provenance=TrustLevel.AI_INFERRED,
                    )
                )

        return entries

    @staticmethod
    def _diff_ubiquitous_language(
        before: DomainModel, after: DomainModel
    ) -> list[DiffEntry]:
        before_terms = before.ubiquitous_language.terms
        after_terms = after.ubiquitous_language.terms

        # Group terms by normalized name
        before_by_term: dict[str, list[str]] = {}
        for t in before_terms:
            before_by_term.setdefault(t.term.lower(), []).append(t.context_name)

        after_by_term: dict[str, list[str]] = {}
        for t in after_terms:
            after_by_term.setdefault(t.term.lower(), []).append(t.context_name)

        entries: list[DiffEntry] = []

        for term_lower, contexts in after_by_term.items():
            original_name = next(
                t.term for t in after_terms if t.term.lower() == term_lower
            )
            if term_lower not in before_by_term:
                entries.append(
                    DiffEntry(
                        diff_type=DiffType.ADDED,
                        section="Ubiquitous Language",
                        description=f"Added term '{original_name}'",
                        provenance=TrustLevel.AI_INFERRED,
                    )
                )
            else:
                old_contexts = set(before_by_term[term_lower])
                new_contexts = set(contexts)
                # Term existed in one context, now split into multiple
                if len(old_contexts) == 1 and len(new_contexts) > 1:
                    entries.append(
                        DiffEntry(
                            diff_type=DiffType.DISAMBIGUATED,
                            section="Ubiquitous Language",
                            description=(
                                f"Term '{original_name}' disambiguated across "
                                f"contexts: {', '.join(sorted(new_contexts))}"
                            ),
                            provenance=TrustLevel.AI_INFERRED,
                        )
                    )
                elif old_contexts != new_contexts:
                    entries.append(
                        DiffEntry(
                            diff_type=DiffType.MODIFIED,
                            section="Ubiquitous Language",
                            description=f"Modified term '{original_name}'",
                            provenance=TrustLevel.AI_INFERRED,
                        )
                    )

        for term_lower in before_by_term:
            if term_lower not in after_by_term:
                original_name = next(
                    t.term for t in before_terms if t.term.lower() == term_lower
                )
                entries.append(
                    DiffEntry(
                        diff_type=DiffType.REMOVED,
                        section="Ubiquitous Language",
                        description=f"Removed term '{original_name}'",
                        provenance=TrustLevel.AI_INFERRED,
                    )
                )

        return entries

    @staticmethod
    def _diff_aggregate_designs(
        before: DomainModel, after: DomainModel
    ) -> list[DiffEntry]:
        def key(a: AggregateDesign) -> tuple[str, str]:
            return (a.name.lower(), a.context_name.lower())

        before_by_key = {key(a): a for a in before.aggregate_designs}
        after_by_key = {key(a): a for a in after.aggregate_designs}

        entries: list[DiffEntry] = []

        for k, agg in after_by_key.items():
            if k not in before_by_key:
                entries.append(
                    DiffEntry(
                        diff_type=DiffType.ADDED,
                        section="Aggregate Designs",
                        description=f"Added aggregate '{agg.name}' in {agg.context_name}",
                        provenance=TrustLevel.AI_INFERRED,
                    )
                )
            else:
                old = before_by_key[k]
                if old.invariants != agg.invariants:
                    entries.append(
                        DiffEntry(
                            diff_type=DiffType.MODIFIED,
                            section="Aggregate Designs",
                            description=(
                                f"Modified aggregate '{agg.name}' invariants "
                                f"in {agg.context_name}"
                            ),
                            provenance=TrustLevel.AI_INFERRED,
                        )
                    )

        for k, agg in before_by_key.items():
            if k not in after_by_key:
                entries.append(
                    DiffEntry(
                        diff_type=DiffType.REMOVED,
                        section="Aggregate Designs",
                        description=f"Removed aggregate '{agg.name}' from {agg.context_name}",
                        provenance=TrustLevel.AI_INFERRED,
                    )
                )

        return entries
