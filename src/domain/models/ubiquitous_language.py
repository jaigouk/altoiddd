"""UbiquitousLanguage entity for the Domain Model bounded context.

Manages the glossary of domain terms. Each term has a definition and
belongs to a bounded context. Terms appearing in multiple contexts with
different meanings are flagged as ambiguous.
"""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class TermEntry:
    """A single term in the ubiquitous language glossary.

    Attributes:
        term: The domain term (e.g. "Order", "Policy").
        definition: What it means in this context.
        context_name: Which bounded context this definition applies to.
        source_question_ids: Which discovery questions surfaced this term.
    """

    term: str
    definition: str
    context_name: str
    source_question_ids: tuple[str, ...] = ()


class UbiquitousLanguage:
    """Entity managing the shared vocabulary of a domain model.

    Terms are added with their definition and bounded context. The entity
    can detect ambiguous terms (same term, different contexts) and verify
    that all terms appear in domain stories.
    """

    def __init__(self) -> None:
        self._terms: list[TermEntry] = []

    @property
    def terms(self) -> tuple[TermEntry, ...]:
        """All term entries (defensive copy)."""
        return tuple(self._terms)

    def add_term(
        self,
        term: str,
        definition: str,
        context_name: str,
        source_question_ids: tuple[str, ...] = (),
    ) -> None:
        """Add a term to the glossary.

        Args:
            term: The domain term.
            definition: What it means.
            context_name: Which bounded context it belongs to.
            source_question_ids: Which questions surfaced this term.

        Raises:
            ValueError: If term or definition is empty.
        """
        if not term.strip():
            msg = "Term cannot be empty"
            raise ValueError(msg)
        if not definition.strip():
            msg = "Definition cannot be empty"
            raise ValueError(msg)

        self._terms.append(
            TermEntry(
                term=term.strip(),
                definition=definition.strip(),
                context_name=context_name,
                source_question_ids=source_question_ids,
            )
        )

    def get_terms_for_context(self, context_name: str) -> tuple[TermEntry, ...]:
        """Return all terms belonging to a specific bounded context."""
        return tuple(t for t in self._terms if t.context_name == context_name)

    def find_ambiguous_terms(self) -> tuple[str, ...]:
        """Find terms that appear in multiple bounded contexts.

        Returns:
            Tuple of term strings that appear in more than one context.
        """
        term_contexts: dict[str, set[str]] = {}
        for entry in self._terms:
            normalized = entry.term.lower()
            if normalized not in term_contexts:
                term_contexts[normalized] = set()
            term_contexts[normalized].add(entry.context_name)

        return tuple(
            term
            for term, contexts in sorted(term_contexts.items())
            if len(contexts) > 1
        )

    def has_per_context_definitions(self, term: str) -> bool:
        """Check if an ambiguous term has a definition in each context it appears in.

        Args:
            term: The term to check (case-insensitive).

        Returns:
            True if every context that uses this term has a definition for it.
        """
        normalized = term.lower()
        entries = [e for e in self._terms if e.term.lower() == normalized]
        contexts = {e.context_name for e in entries}

        # Each context must have at least one definition
        for ctx in contexts:
            ctx_entries = [e for e in entries if e.context_name == ctx]
            if not any(e.definition.strip() for e in ctx_entries):
                return False
        return True

    @property
    def all_term_names(self) -> frozenset[str]:
        """All unique term names (lowercased) for quick lookup."""
        return frozenset(e.term.lower() for e in self._terms)
