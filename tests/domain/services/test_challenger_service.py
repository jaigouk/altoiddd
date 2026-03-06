"""Tests for ChallengerService — rule-based challenge generation.

Verifies that the stateless service inspects a DomainModel and generates
typed challenges for language ambiguity, missing invariants, failure modes,
and boundary disputes.
"""

from __future__ import annotations

from src.domain.models.challenge import ChallengeType
from src.domain.models.domain_model import DomainModel
from src.domain.models.domain_values import (
    AggregateDesign,
    BoundedContext,
    ContextRelationship,
    DomainStory,
    SubdomainClassification,
)
from src.domain.services.challenger_service import ChallengerService


def _make_rich_model() -> DomainModel:
    """Create a DomainModel with enough data to trigger all challenge types."""
    model = DomainModel()

    # Two contexts — enables boundary challenges
    model.add_bounded_context(BoundedContext(name="Sales", responsibility="Orders"))
    model.classify_subdomain("Sales", SubdomainClassification.CORE)
    model.add_bounded_context(BoundedContext(name="Shipping", responsibility="Deliveries"))
    model.classify_subdomain("Shipping", SubdomainClassification.SUPPORTING)

    # Aggregate with no invariants — triggers invariant challenge
    model.design_aggregate(
        AggregateDesign(
            name="OrderAggregate",
            context_name="Sales",
            root_entity="Order",
            invariants=(),  # deliberately empty
        )
    )

    # Domain story — triggers failure mode challenges
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

    # Terms — "Order" appears only in Sales, "Shipment" only in Shipping
    model.add_term("Order", "A customer purchase", "Sales")
    model.add_term("Shipment", "A delivery package", "Shipping")

    # Relationship
    model.add_context_relationship(ContextRelationship("Sales", "Shipping", "Domain Events"))

    return model


def _make_empty_model() -> DomainModel:
    """Create an empty DomainModel."""
    return DomainModel()


def _make_single_context_model() -> DomainModel:
    """Create a model with only one bounded context."""
    model = DomainModel()
    model.add_bounded_context(BoundedContext(name="Sales", responsibility="Orders"))
    model.classify_subdomain("Sales", SubdomainClassification.CORE)
    model.design_aggregate(
        AggregateDesign(
            name="OrderAggregate",
            context_name="Sales",
            root_entity="Order",
            invariants=("Total must be positive",),
        )
    )
    model.add_domain_story(
        DomainStory(
            name="Place Order",
            actors=("Customer",),
            trigger="Customer submits order",
            steps=("System creates order",),
        )
    )
    model.add_term("Order", "A purchase", "Sales")
    return model


def _make_generic_only_model() -> DomainModel:
    """Create a model with only Generic subdomains."""
    model = DomainModel()
    model.add_bounded_context(BoundedContext(name="Auth", responsibility="Authentication"))
    model.classify_subdomain("Auth", SubdomainClassification.GENERIC)
    model.add_domain_story(
        DomainStory(
            name="Login",
            actors=("User",),
            trigger="User enters credentials",
            steps=("System verifies credentials",),
        )
    )
    model.add_term("User", "An authenticated person", "Auth")
    return model


class TestChallengerServiceGeneration:
    def test_generates_language_challenges_for_ambiguous_terms(self) -> None:
        """Terms appearing in multiple contexts without per-context defs trigger challenges."""
        model = DomainModel()
        model.add_bounded_context(BoundedContext(name="Sales", responsibility="Orders"))
        model.classify_subdomain("Sales", SubdomainClassification.CORE)
        model.add_bounded_context(BoundedContext(name="Shipping", responsibility="Deliveries"))
        model.classify_subdomain("Shipping", SubdomainClassification.SUPPORTING)
        model.design_aggregate(
            AggregateDesign(
                name="OrderAggregate", context_name="Sales", root_entity="Order"
            )
        )
        model.add_domain_story(
            DomainStory(
                name="Ship Order",
                actors=("Warehouse",),
                trigger="Order confirmed",
                steps=("Create shipment for order",),
            )
        )
        # Same term in two contexts — ambiguous
        model.add_term("Order", "A purchase", "Sales")
        model.add_term("Order", "A shipping request", "Shipping")

        challenges = ChallengerService.generate(model)
        language = [c for c in challenges if c.challenge_type == ChallengeType.LANGUAGE]
        assert len(language) >= 1
        assert any("order" in c.question_text.lower() for c in language)

    def test_generates_invariant_challenges_for_empty_aggregates(self) -> None:
        """Core aggregates with no invariants trigger challenges."""
        model = _make_rich_model()
        challenges = ChallengerService.generate(model)
        invariant = [c for c in challenges if c.challenge_type == ChallengeType.INVARIANT]
        assert len(invariant) >= 1
        assert any(
            "orderaggregate" in c.question_text.lower()
            or "order" in c.question_text.lower()
            for c in invariant
        )

    def test_generates_failure_mode_challenges_for_core_stories(self) -> None:
        """Each Core domain story step gets probed for failure modes."""
        model = _make_rich_model()
        challenges = ChallengerService.generate(model)
        failure = [c for c in challenges if c.challenge_type == ChallengeType.FAILURE_MODE]
        assert len(failure) >= 1

    def test_generates_boundary_challenges_for_multiple_contexts(self) -> None:
        """When >= 2 contexts exist, boundary challenges are generated."""
        model = _make_rich_model()
        challenges = ChallengerService.generate(model)
        boundary = [c for c in challenges if c.challenge_type == ChallengeType.BOUNDARY]
        assert len(boundary) >= 1

    def test_single_context_no_boundary_challenges(self) -> None:
        """Need >= 2 contexts for boundary challenges."""
        model = _make_single_context_model()
        challenges = ChallengerService.generate(model)
        boundary = [c for c in challenges if c.challenge_type == ChallengeType.BOUNDARY]
        assert len(boundary) == 0

    def test_skips_generic_subdomains_for_invariant_challenges(self) -> None:
        """Generic subdomains don't get invariant challenges."""
        model = _make_generic_only_model()
        challenges = ChallengerService.generate(model)
        invariant = [c for c in challenges if c.challenge_type == ChallengeType.INVARIANT]
        assert len(invariant) == 0

    def test_every_challenge_has_source_reference(self) -> None:
        """Every challenge must cite evidence."""
        model = _make_rich_model()
        challenges = ChallengerService.generate(model)
        assert len(challenges) > 0
        for c in challenges:
            assert c.source_reference.strip(), f"Challenge missing source_reference: {c}"

    def test_max_challenges_per_type_respected(self) -> None:
        """Each type is limited to max_per_type challenges."""
        model = _make_rich_model()
        challenges = ChallengerService.generate(model, max_per_type=2)
        by_type: dict[ChallengeType, int] = {}
        for c in challenges:
            by_type[c.challenge_type] = by_type.get(c.challenge_type, 0) + 1
        for ct, count in by_type.items():
            assert count <= 2, f"{ct.value} has {count} challenges (max 2)"

    def test_empty_model_returns_no_challenges(self) -> None:
        """An empty model produces no challenges."""
        model = _make_empty_model()
        challenges = ChallengerService.generate(model)
        assert challenges == ()

    def test_challenge_context_name_matches_bounded_context(self) -> None:
        """Each challenge's context_name should reference a real bounded context."""
        model = _make_rich_model()
        challenges = ChallengerService.generate(model)
        context_names = {ctx.name for ctx in model.bounded_contexts}
        for c in challenges:
            assert c.context_name in context_names, (
                f"Challenge references unknown context: {c.context_name}"
            )
