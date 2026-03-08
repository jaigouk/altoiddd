"""Tests for DiffService — computes ArtifactDiff between two DomainModel snapshots.

RED phase: 9 tests covering bounded contexts, domain stories, ubiquitous
language, aggregate designs, disambiguated terms, and convergence metric.
"""

from __future__ import annotations

from src.domain.models.artifact_diff import DiffType
from src.domain.models.domain_model import DomainModel
from src.domain.models.domain_values import (
    AggregateDesign,
    BoundedContext,
    DomainStory,
    SubdomainClassification,
)
from src.domain.models.research import TrustLevel
from src.domain.services.diff_service import DiffService


def _make_model() -> DomainModel:
    """Create a minimal valid DomainModel for testing."""
    return DomainModel()


class TestDiffServiceBoundedContexts:
    """Compare bounded contexts between two models."""

    def test_added_context(self) -> None:
        before = _make_model()
        after = _make_model()
        after.add_bounded_context(BoundedContext(name="Billing", responsibility="Payments"))

        diff = DiffService.compute(before, after)
        added = [e for e in diff.entries if e.diff_type == DiffType.ADDED]
        assert len(added) == 1
        assert "Billing" in added[0].description
        assert added[0].section == "Bounded Contexts"
        assert added[0].provenance == TrustLevel.AI_INFERRED

    def test_removed_context(self) -> None:
        before = _make_model()
        before.add_bounded_context(BoundedContext(name="Billing", responsibility="Payments"))
        after = _make_model()

        diff = DiffService.compute(before, after)
        removed = [e for e in diff.entries if e.diff_type == DiffType.REMOVED]
        assert len(removed) == 1
        assert "Billing" in removed[0].description

    def test_modified_context_classification(self) -> None:
        before = _make_model()
        before.add_bounded_context(
            BoundedContext(
                name="Billing",
                responsibility="Payments",
                classification=SubdomainClassification.SUPPORTING,
            )
        )
        after = _make_model()
        after.add_bounded_context(
            BoundedContext(
                name="Billing",
                responsibility="Payments",
                classification=SubdomainClassification.CORE,
            )
        )

        diff = DiffService.compute(before, after)
        modified = [e for e in diff.entries if e.diff_type == DiffType.MODIFIED]
        assert len(modified) == 1
        assert "Billing" in modified[0].description


class TestDiffServiceDomainStories:
    """Compare domain stories between two models."""

    def test_added_story(self) -> None:
        before = _make_model()
        after = _make_model()
        after.add_domain_story(
            DomainStory(
                name="Place Order",
                actors=("Customer",),
                trigger="Customer clicks buy",
                steps=("Customer submits order",),
            )
        )

        diff = DiffService.compute(before, after)
        added = [
            e
            for e in diff.entries
            if e.diff_type == DiffType.ADDED and e.section == "Domain Stories"
        ]
        assert len(added) == 1
        assert "Place Order" in added[0].description

    def test_modified_story_steps(self) -> None:
        before = _make_model()
        before.add_domain_story(
            DomainStory(
                name="Place Order",
                actors=("Customer",),
                trigger="Customer clicks buy",
                steps=("Customer submits order",),
            )
        )
        after = _make_model()
        after.add_domain_story(
            DomainStory(
                name="Place Order",
                actors=("Customer",),
                trigger="Customer clicks buy",
                steps=("Customer submits order", "System validates payment"),
            )
        )

        diff = DiffService.compute(before, after)
        modified = [
            e
            for e in diff.entries
            if e.diff_type == DiffType.MODIFIED and e.section == "Domain Stories"
        ]
        assert len(modified) == 1
        assert "Place Order" in modified[0].description


class TestDiffServiceUbiquitousLanguage:
    """Compare UL terms between two models."""

    def test_added_term(self) -> None:
        before = _make_model()
        after = _make_model()
        after.add_term("Order", "A purchase request", "Sales")

        diff = DiffService.compute(before, after)
        added = [
            e
            for e in diff.entries
            if e.diff_type == DiffType.ADDED and e.section == "Ubiquitous Language"
        ]
        assert len(added) == 1
        assert "Order" in added[0].description

    def test_disambiguated_term(self) -> None:
        """Term in one context before, split into per-context definitions after."""
        before = _make_model()
        before.add_term("Policy", "A business rule", "Sales")

        after = _make_model()
        after.add_term("Policy", "A sales rule", "Sales")
        after.add_term("Policy", "An insurance contract", "Insurance")

        diff = DiffService.compute(before, after)
        disambiguated = [
            e
            for e in diff.entries
            if e.diff_type == DiffType.DISAMBIGUATED
            and e.section == "Ubiquitous Language"
        ]
        assert len(disambiguated) == 1
        assert "Policy" in disambiguated[0].description


class TestDiffServiceAggregateDesigns:
    """Compare aggregate designs between two models."""

    def test_added_aggregate(self) -> None:
        before = _make_model()
        after = _make_model()
        after.design_aggregate(
            AggregateDesign(
                name="OrderAggregate",
                context_name="Sales",
                root_entity="Order",
            )
        )

        diff = DiffService.compute(before, after)
        added = [
            e
            for e in diff.entries
            if e.diff_type == DiffType.ADDED and e.section == "Aggregate Designs"
        ]
        assert len(added) == 1
        assert "OrderAggregate" in added[0].description

    def test_modified_aggregate_invariants(self) -> None:
        before = _make_model()
        before.design_aggregate(
            AggregateDesign(
                name="OrderAggregate",
                context_name="Sales",
                root_entity="Order",
                invariants=("total >= 0",),
            )
        )
        after = _make_model()
        after.design_aggregate(
            AggregateDesign(
                name="OrderAggregate",
                context_name="Sales",
                root_entity="Order",
                invariants=("total >= 0", "at least one line item"),
            )
        )

        diff = DiffService.compute(before, after)
        modified = [
            e
            for e in diff.entries
            if e.diff_type == DiffType.MODIFIED and e.section == "Aggregate Designs"
        ]
        assert len(modified) == 1
        assert "OrderAggregate" in modified[0].description


class TestDiffServiceConvergence:
    """ConvergenceMetric is computed correctly."""

    def test_convergence_counts(self) -> None:
        before = _make_model()
        before.add_bounded_context(BoundedContext(name="Billing", responsibility="Payments"))
        before.add_term("Order", "A purchase request", "Sales")

        after = _make_model()
        after.add_bounded_context(
            BoundedContext(name="Billing", responsibility="Payments and invoicing")
        )
        after.add_term("Order", "A purchase request", "Sales")
        after.add_term("Invoice", "A billing document", "Billing")
        after.add_domain_story(
            DomainStory(
                name="Bill Customer",
                actors=("Billing Agent",),
                trigger="Order completed",
                steps=("Agent creates invoice",),
            )
        )

        diff = DiffService.compute(before, after)
        assert diff.convergence.terms_delta == 1  # Invoice added
        assert diff.convergence.stories_delta == 1  # Bill Customer added


class TestDiffServiceVersionParams:
    """DiffService.compute() respects from_version and to_version."""

    def test_custom_version_numbers(self) -> None:
        before = _make_model()
        after = _make_model()
        after.add_bounded_context(BoundedContext(name="Billing", responsibility="Pay"))

        diff = DiffService.compute(before, after, from_version=3, to_version=7)
        assert diff.from_version == 3
        assert diff.to_version == 7

    def test_default_version_numbers(self) -> None:
        diff = DiffService.compute(_make_model(), _make_model())
        assert diff.from_version == 1
        assert diff.to_version == 2


class TestDiffServiceRemovedAggregate:
    """Compare removed aggregate designs between two models."""

    def test_removed_aggregate(self) -> None:
        before = _make_model()
        before.design_aggregate(
            AggregateDesign(
                name="OrderAggregate",
                context_name="Sales",
                root_entity="Order",
            )
        )
        after = _make_model()

        diff = DiffService.compute(before, after)
        removed = [
            e
            for e in diff.entries
            if e.diff_type == DiffType.REMOVED and e.section == "Aggregate Designs"
        ]
        assert len(removed) == 1
        assert "OrderAggregate" in removed[0].description


class TestDiffServiceRemovedStoryAndTerm:
    """Compare removed stories and terms."""

    def test_removed_story(self) -> None:
        before = _make_model()
        before.add_domain_story(
            DomainStory(
                name="Checkout",
                actors=("Customer",),
                trigger="Click buy",
                steps=("Submit order",),
            )
        )
        after = _make_model()

        diff = DiffService.compute(before, after)
        removed = [
            e
            for e in diff.entries
            if e.diff_type == DiffType.REMOVED and e.section == "Domain Stories"
        ]
        assert len(removed) == 1
        assert "Checkout" in removed[0].description

    def test_removed_term(self) -> None:
        before = _make_model()
        before.add_term("Order", "A purchase", "Sales")
        after = _make_model()

        diff = DiffService.compute(before, after)
        removed = [
            e
            for e in diff.entries
            if e.diff_type == DiffType.REMOVED and e.section == "Ubiquitous Language"
        ]
        assert len(removed) == 1
        assert "Order" in removed[0].description

    def test_modified_story_actors(self) -> None:
        before = _make_model()
        before.add_domain_story(
            DomainStory(
                name="Checkout",
                actors=("Customer",),
                trigger="Click buy",
                steps=("Submit order",),
            )
        )
        after = _make_model()
        after.add_domain_story(
            DomainStory(
                name="Checkout",
                actors=("Customer", "Cashier"),
                trigger="Click buy",
                steps=("Submit order",),
            )
        )

        diff = DiffService.compute(before, after)
        modified = [
            e
            for e in diff.entries
            if e.diff_type == DiffType.MODIFIED and e.section == "Domain Stories"
        ]
        assert len(modified) == 1
        assert "Checkout" in modified[0].description


class TestDiffServiceIdenticalModels:
    """Identical models produce empty diff."""

    def test_no_entries_for_identical(self) -> None:
        model1 = _make_model()
        model1.add_bounded_context(BoundedContext(name="Sales", responsibility="Orders"))
        model1.add_term("Order", "A purchase", "Sales")

        model2 = _make_model()
        model2.add_bounded_context(BoundedContext(name="Sales", responsibility="Orders"))
        model2.add_term("Order", "A purchase", "Sales")

        diff = DiffService.compute(model1, model2)
        assert len(diff.entries) == 0
        assert diff.convergence.canvases_delta == 0
        assert diff.convergence.terms_delta == 0
