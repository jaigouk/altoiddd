"""Tests for DomainModel aggregate root — all 5 invariants + edge cases."""

from __future__ import annotations

import pytest

from src.domain.models.bootstrap_session import InvariantViolationError
from src.domain.models.domain_model import DomainModel
from src.domain.models.domain_values import (
    AggregateDesign,
    BoundedContext,
    ContextRelationship,
    DomainStory,
    SubdomainClassification,
)
from src.domain.models.errors import DuplicateStoryError


def _make_valid_model() -> DomainModel:
    """Create a DomainModel that passes all 5 invariants."""
    model = DomainModel()

    # Story mentioning all terms.
    model.add_domain_story(
        DomainStory(
            name="Checkout Flow",
            actors=("Customer",),
            trigger="Customer clicks checkout",
            steps=(
                "Customer reviews order",
                "System validates payment",
                "System creates shipment",
            ),
        )
    )

    # Bounded contexts with classification.
    model.add_bounded_context(
        BoundedContext(name="Sales", responsibility="Manages orders")
    )
    model.classify_subdomain(
        "Sales", SubdomainClassification.CORE, "Competitive advantage"
    )

    model.add_bounded_context(
        BoundedContext(name="Shipping", responsibility="Manages shipments")
    )
    model.classify_subdomain(
        "Shipping", SubdomainClassification.SUPPORTING, "Necessary plumbing"
    )

    # Terms that appear in the story text.
    model.add_term("Order", "A customer purchase", "Sales", ("Q2",))
    model.add_term("Shipment", "A delivery package", "Shipping", ("Q2",))

    # Aggregate for Core subdomain.
    model.design_aggregate(
        AggregateDesign(
            name="OrderAggregate",
            context_name="Sales",
            root_entity="Order",
            invariants=("Total must be positive",),
        )
    )

    # Bidirectional relationship.
    model.add_context_relationship(
        ContextRelationship("Sales", "Shipping", "Domain Events")
    )
    model.add_context_relationship(
        ContextRelationship("Shipping", "Sales", "Query")
    )

    return model


class TestCreateDomainModel:
    def test_empty_model(self) -> None:
        model = DomainModel()
        assert model.domain_stories == ()
        assert model.bounded_contexts == ()
        assert model.aggregate_designs == ()
        assert model.events == ()

    def test_model_id_generated(self) -> None:
        model = DomainModel()
        assert model.model_id  # Non-empty UUID string.


class TestAddDomainStory:
    def test_add_story(self) -> None:
        model = DomainModel()
        story = DomainStory(
            name="Test", actors=("A",), trigger="T", steps=("S",)
        )
        model.add_domain_story(story)
        assert len(model.domain_stories) == 1
        assert model.domain_stories[0].name == "Test"

    def test_duplicate_story_raises(self) -> None:
        model = DomainModel()
        story = DomainStory(
            name="Test", actors=("A",), trigger="T", steps=("S",)
        )
        model.add_domain_story(story)
        with pytest.raises(DuplicateStoryError, match="'Test' already exists"):
            model.add_domain_story(story)

    def test_duplicate_case_insensitive(self) -> None:
        model = DomainModel()
        model.add_domain_story(
            DomainStory(name="Test", actors=("A",), trigger="T", steps=("S",))
        )
        with pytest.raises(DuplicateStoryError):
            model.add_domain_story(
                DomainStory(name="test", actors=("B",), trigger="T2", steps=("S2",))
            )

    def test_empty_name_raises(self) -> None:
        model = DomainModel()
        with pytest.raises(ValueError, match="Story name cannot be empty"):
            model.add_domain_story(
                DomainStory(name="", actors=("A",), trigger="T", steps=("S",))
            )


class TestAddBoundedContext:
    def test_add_context(self) -> None:
        model = DomainModel()
        ctx = BoundedContext(name="Sales", responsibility="Orders")
        model.add_bounded_context(ctx)
        assert len(model.bounded_contexts) == 1

    def test_duplicate_context_raises(self) -> None:
        model = DomainModel()
        model.add_bounded_context(
            BoundedContext(name="Sales", responsibility="Orders")
        )
        with pytest.raises(ValueError, match="'Sales' already exists"):
            model.add_bounded_context(
                BoundedContext(name="Sales", responsibility="Other")
            )

    def test_empty_name_raises(self) -> None:
        model = DomainModel()
        with pytest.raises(ValueError, match="Context name cannot be empty"):
            model.add_bounded_context(
                BoundedContext(name="", responsibility="X")
            )


class TestClassifySubdomain:
    def test_classify(self) -> None:
        model = DomainModel()
        model.add_bounded_context(
            BoundedContext(name="Sales", responsibility="Orders")
        )
        model.classify_subdomain("Sales", SubdomainClassification.CORE, "Key value")
        assert model.bounded_contexts[0].classification == SubdomainClassification.CORE

    def test_unknown_context_raises(self) -> None:
        model = DomainModel()
        with pytest.raises(ValueError, match="not found"):
            model.classify_subdomain("Missing", SubdomainClassification.CORE)


class TestDesignAggregate:
    def test_add_aggregate(self) -> None:
        model = DomainModel()
        agg = AggregateDesign(
            name="OrderAgg", context_name="Sales", root_entity="Order"
        )
        model.design_aggregate(agg)
        assert len(model.aggregate_designs) == 1

    def test_empty_name_raises(self) -> None:
        model = DomainModel()
        with pytest.raises(ValueError, match="Aggregate name cannot be empty"):
            model.design_aggregate(
                AggregateDesign(name="", context_name="Sales", root_entity="X")
            )

    def test_empty_context_raises(self) -> None:
        model = DomainModel()
        with pytest.raises(ValueError, match="context name cannot be empty"):
            model.design_aggregate(
                AggregateDesign(name="Agg", context_name="", root_entity="X")
            )


class TestAddContextRelationship:
    def test_add_relationship(self) -> None:
        model = DomainModel()
        rel = ContextRelationship("A", "B", "Events")
        model.add_context_relationship(rel)
        assert len(model.context_relationships) == 1

    def test_empty_upstream_raises(self) -> None:
        model = DomainModel()
        with pytest.raises(ValueError, match="cannot be empty"):
            model.add_context_relationship(
                ContextRelationship("", "B", "Events")
            )


class TestFinalizeInvariant1TermsInStories:
    """Invariant 1: Every UL term must appear in at least one DomainStory."""

    def test_term_in_story_passes(self) -> None:
        model = _make_valid_model()
        model.finalize()  # Should not raise.
        assert len(model.events) == 1

    def test_term_not_in_story_raises(self) -> None:
        model = _make_valid_model()
        model.add_term("Widget", "A widget thing", "Sales", ("Q2",))
        with pytest.raises(
            InvariantViolationError, match="Term 'Widget' not found in any domain story"
        ):
            model.finalize()


class TestFinalizeInvariant2ContextClassification:
    """Invariant 2: Every BoundedContext must have a SubdomainClassification."""

    def test_all_classified_passes(self) -> None:
        model = _make_valid_model()
        model.finalize()  # Should not raise.

    def test_unclassified_context_raises(self) -> None:
        model = _make_valid_model()
        model.add_bounded_context(
            BoundedContext(name="Unclassified", responsibility="Test")
        )
        with pytest.raises(
            InvariantViolationError, match="'Unclassified' has no classification"
        ):
            model.finalize()


class TestFinalizeInvariant3CoreAggregates:
    """Invariant 3: Every Core subdomain must have at least one AggregateDesign."""

    def test_core_with_aggregate_passes(self) -> None:
        model = _make_valid_model()
        model.finalize()  # Sales is Core and has OrderAggregate.

    def test_core_without_aggregate_raises(self) -> None:
        model = DomainModel()
        model.add_domain_story(
            DomainStory(
                name="Test Story",
                actors=("Actor",),
                trigger="Start",
                steps=("Actor does discovery",),
            )
        )
        model.add_bounded_context(
            BoundedContext(name="Discovery", responsibility="Guides")
        )
        model.classify_subdomain("Discovery", SubdomainClassification.CORE)
        model.add_term("Discovery", "The process", "Discovery")
        with pytest.raises(
            InvariantViolationError, match="Core subdomain 'Discovery' has no aggregate"
        ):
            model.finalize()


class TestRelationshipsAccepted:
    """Relationships are stored but not required to be bidirectional (M3)."""

    def test_bidirectional_passes(self) -> None:
        model = _make_valid_model()
        model.finalize()  # Sales↔Shipping bidirectional — still valid.

    def test_unidirectional_passes(self) -> None:
        model = _make_valid_model()
        model.add_bounded_context(
            BoundedContext(name="Billing", responsibility="Bills")
        )
        model.classify_subdomain("Billing", SubdomainClassification.GENERIC)
        model.add_context_relationship(
            ContextRelationship("Sales", "Billing", "Events")
        )
        model.finalize()  # One-way is now valid.


class TestFinalizeInvariant4AmbiguousTerms:
    """Invariant 4: Ambiguous terms must have per-context definitions."""

    def test_no_ambiguity_passes(self) -> None:
        model = _make_valid_model()
        model.finalize()

    def test_ambiguous_with_definitions_passes(self) -> None:
        model = _make_valid_model()
        # "Order" now appears in both contexts — but with definitions.
        model.add_term("Order", "Work order", "Shipping")
        # "Order" is already in Sales from _make_valid_model.
        model.finalize()  # Both contexts have definitions — OK.

    def test_ambiguous_without_definitions_raises(self) -> None:
        model = _make_valid_model()
        # Bypass add_term validation to simulate a term with empty definition.
        from src.domain.models.ubiquitous_language import TermEntry

        model._language._terms.append(
            TermEntry(term="Order", definition="", context_name="Shipping")
        )
        with pytest.raises(
            InvariantViolationError, match="Ambiguous term 'order' needs per-context"
        ):
            model.finalize()


class TestFinalizeEmitsEvent:
    def test_emits_domain_model_generated(self) -> None:
        model = _make_valid_model()
        model.finalize()
        assert len(model.events) == 1
        event = model.events[0]
        assert event.model_id == model.model_id
        assert len(event.domain_stories) == 1
        assert len(event.bounded_contexts) == 2
        assert len(event.aggregate_designs) == 1


class TestDefensiveCopies:
    def test_stories_defensive_copy(self) -> None:
        model = DomainModel()
        model.add_domain_story(
            DomainStory(name="S", actors=("A",), trigger="T", steps=("S",))
        )
        s1 = model.domain_stories
        s2 = model.domain_stories
        assert s1 is not s2

    def test_contexts_defensive_copy(self) -> None:
        model = DomainModel()
        model.add_bounded_context(
            BoundedContext(name="Ctx", responsibility="Test")
        )
        c1 = model.bounded_contexts
        c2 = model.bounded_contexts
        assert c1 is not c2

    def test_events_defensive_copy(self) -> None:
        model = _make_valid_model()
        model.finalize()
        e1 = model.events
        e2 = model.events
        assert e1 is not e2


# =========================================================================
# New tests from code review fixes
# =========================================================================


def _make_valid_model_without_relationships() -> DomainModel:
    """Create a DomainModel passing invariants 1-3 and 5, without relationships."""
    model = DomainModel()

    model.add_domain_story(
        DomainStory(
            name="Checkout Flow",
            actors=("Customer",),
            trigger="Customer clicks checkout",
            steps=(
                "Customer reviews order",
                "System validates payment",
                "System creates shipment",
            ),
        )
    )

    model.add_bounded_context(
        BoundedContext(name="Sales", responsibility="Manages orders")
    )
    model.classify_subdomain(
        "Sales", SubdomainClassification.CORE, "Competitive advantage"
    )

    model.add_bounded_context(
        BoundedContext(name="Shipping", responsibility="Manages shipments")
    )
    model.classify_subdomain(
        "Shipping", SubdomainClassification.SUPPORTING, "Necessary plumbing"
    )

    model.add_term("Order", "A customer purchase", "Sales", ("Q2",))
    model.add_term("Shipment", "A delivery package", "Shipping", ("Q2",))

    model.design_aggregate(
        AggregateDesign(
            name="OrderAggregate",
            context_name="Sales",
            root_entity="Order",
            invariants=("Total must be positive",),
        )
    )

    return model


class TestReassignTermsToContext:
    """H1: DomainModel.reassign_terms_to_context() replaces encapsulation hack."""

    def test_reassign_moves_all_matching_terms(self) -> None:
        model = DomainModel()
        model.add_term("Order", "A purchase", "Default")
        model.add_term("Product", "An item", "Default")
        model.reassign_terms_to_context("Default", "Sales")
        sales = model.ubiquitous_language.get_terms_for_context("Sales")
        assert len(sales) == 2

    def test_reassign_no_match_is_noop(self) -> None:
        model = DomainModel()
        model.add_term("Order", "A purchase", "Sales")
        model.reassign_terms_to_context("NonExistent", "Other")
        assert len(model.ubiquitous_language.get_terms_for_context("Sales")) == 1
        assert len(model.ubiquitous_language.get_terms_for_context("Other")) == 0

    def test_reassign_preserves_other_context_terms(self) -> None:
        model = DomainModel()
        model.add_term("Order", "A purchase", "Default")
        model.add_term("Product", "An item", "Catalog")
        model.reassign_terms_to_context("Default", "Sales")
        assert len(model.ubiquitous_language.get_terms_for_context("Sales")) == 1
        assert len(model.ubiquitous_language.get_terms_for_context("Catalog")) == 1

    def test_reassign_preserves_definition_and_source(self) -> None:
        model = DomainModel()
        model.add_term("Order", "A customer purchase", "Default", ("Q2",))
        model.reassign_terms_to_context("Default", "Sales")
        term = model.ubiquitous_language.get_terms_for_context("Sales")[0]
        assert term.definition == "A customer purchase"
        assert term.source_question_ids == ("Q2",)

    def test_reassign_empty_from_raises(self) -> None:
        model = DomainModel()
        with pytest.raises(ValueError, match="cannot be empty"):
            model.reassign_terms_to_context("", "Sales")

    def test_reassign_empty_to_raises(self) -> None:
        model = DomainModel()
        with pytest.raises(ValueError, match="cannot be empty"):
            model.reassign_terms_to_context("Default", "")

    def test_reassign_whitespace_from_raises(self) -> None:
        model = DomainModel()
        with pytest.raises(ValueError, match="cannot be empty"):
            model.reassign_terms_to_context("   ", "Sales")

    def test_reassign_whitespace_to_raises(self) -> None:
        model = DomainModel()
        with pytest.raises(ValueError, match="cannot be empty"):
            model.reassign_terms_to_context("Default", "   ")

    def test_reassign_case_insensitive_match(self) -> None:
        model = DomainModel()
        model.add_term("Order", "A purchase", "default")
        model.reassign_terms_to_context("Default", "Sales")
        assert len(model.ubiquitous_language.get_terms_for_context("Sales")) == 1
        assert len(model.ubiquitous_language.get_terms_for_context("default")) == 0


class TestRelaxedRelationships:
    """M3: Unidirectional relationships are valid (Conformist, ACL, etc.)."""

    def test_unidirectional_passes_finalize(self) -> None:
        model = _make_valid_model_without_relationships()
        model.add_context_relationship(
            ContextRelationship("Sales", "Shipping", "Domain Events")
        )
        model.finalize()  # Must NOT raise.
        assert len(model.events) >= 1

    def test_no_relationships_passes_finalize(self) -> None:
        model = _make_valid_model_without_relationships()
        model.finalize()  # Must NOT raise.
        assert len(model.events) >= 1

    def test_bidirectional_still_accepted(self) -> None:
        model = _make_valid_model()
        model.finalize()  # Bidirectional is still valid.


class TestEventsReturnsTuple:
    """L2: events property returns tuple for consistency with other properties."""

    def test_events_is_tuple_after_finalize(self) -> None:
        model = _make_valid_model()
        model.finalize()
        assert isinstance(model.events, tuple)

    def test_empty_events_is_tuple(self) -> None:
        model = DomainModel()
        assert isinstance(model.events, tuple)


class TestWordBoundaryTermMatching:
    """L3: Invariant 1 uses word boundary matching, not substring."""

    def test_short_term_not_substring_of_longer_word(self) -> None:
        model = DomainModel()
        model.add_domain_story(
            DomainStory(
                name="Flow",
                actors=("Operator",),
                trigger="Start",
                steps=("Operator processes work",),
            )
        )
        model.add_bounded_context(
            BoundedContext(name="Ops", responsibility="Operations")
        )
        model.classify_subdomain("Ops", SubdomainClassification.SUPPORTING)
        # "Or" should NOT match "Operator" with word boundary.
        model.add_term("Or", "Logical operator concept", "Ops")
        with pytest.raises(InvariantViolationError, match="Term 'Or'"):
            model.finalize()

    def test_exact_word_match_passes(self) -> None:
        model = DomainModel()
        model.add_domain_story(
            DomainStory(
                name="Checkout",
                actors=("Customer",),
                trigger="Customer clicks buy",
                steps=("Customer creates Order",),
            )
        )
        model.add_bounded_context(
            BoundedContext(name="Sales", responsibility="Handles orders")
        )
        model.classify_subdomain("Sales", SubdomainClassification.SUPPORTING)
        model.add_term("Order", "A purchase", "Sales")
        model.finalize()  # "Order" is a whole word in steps.

    def test_term_at_start_of_story_name(self) -> None:
        model = DomainModel()
        model.add_domain_story(
            DomainStory(
                name="Order Flow",
                actors=("User",),
                trigger="Start",
                steps=("User submits",),
            )
        )
        model.add_bounded_context(
            BoundedContext(name="Sales", responsibility="Orders")
        )
        model.classify_subdomain("Sales", SubdomainClassification.SUPPORTING)
        model.add_term("Order", "A purchase", "Sales")
        model.finalize()  # "Order" at start of story name.

    def test_multi_word_term_matching(self) -> None:
        model = DomainModel()
        model.add_domain_story(
            DomainStory(
                name="Payment",
                actors=("User",),
                trigger="User pays",
                steps=("System processes sales order",),
            )
        )
        model.add_bounded_context(
            BoundedContext(name="Sales", responsibility="Orders")
        )
        model.classify_subdomain("Sales", SubdomainClassification.SUPPORTING)
        model.add_term("Sales Order", "A customer order", "Sales")
        model.finalize()  # "sales order" appears as a phrase.

    def test_term_in_observations_matches(self) -> None:
        model = DomainModel()
        model.add_domain_story(
            DomainStory(
                name="Flow",
                actors=("User",),
                trigger="Start",
                steps=("User acts",),
                observations=("System creates Invoice",),
            )
        )
        model.add_bounded_context(
            BoundedContext(name="Billing", responsibility="Invoicing")
        )
        model.classify_subdomain("Billing", SubdomainClassification.SUPPORTING)
        model.add_term("Invoice", "A bill", "Billing")
        model.finalize()  # "Invoice" found in observations.

    def test_term_in_actor_list_matches(self) -> None:
        model = DomainModel()
        model.add_domain_story(
            DomainStory(
                name="Flow",
                actors=("Admin",),
                trigger="Start",
                steps=("Admin reviews",),
            )
        )
        model.add_bounded_context(
            BoundedContext(name="Mgmt", responsibility="Management")
        )
        model.classify_subdomain("Mgmt", SubdomainClassification.SUPPORTING)
        model.add_term("Admin", "An administrator", "Mgmt")
        model.finalize()  # "Admin" found in actors.


class TestDuplicateAggregateDesign:
    """L4: design_aggregate rejects duplicates within the same context."""

    def test_duplicate_name_same_context_raises(self) -> None:
        model = DomainModel()
        model.design_aggregate(
            AggregateDesign(
                name="OrderAgg", context_name="Sales", root_entity="Order"
            )
        )
        with pytest.raises(ValueError, match=r"'OrderAgg'.*already exists"):
            model.design_aggregate(
                AggregateDesign(
                    name="OrderAgg", context_name="Sales", root_entity="Other"
                )
            )

    def test_same_name_different_context_allowed(self) -> None:
        model = DomainModel()
        model.design_aggregate(
            AggregateDesign(
                name="RootAgg", context_name="Sales", root_entity="Order"
            )
        )
        model.design_aggregate(
            AggregateDesign(
                name="RootAgg", context_name="Shipping", root_entity="Shipment"
            )
        )
        assert len(model.aggregate_designs) == 2

    def test_duplicate_case_insensitive(self) -> None:
        model = DomainModel()
        model.design_aggregate(
            AggregateDesign(
                name="OrderAgg", context_name="Sales", root_entity="Order"
            )
        )
        with pytest.raises(ValueError, match="already exists"):
            model.design_aggregate(
                AggregateDesign(
                    name="orderagg", context_name="sales", root_entity="Other"
                )
            )
