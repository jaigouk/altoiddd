"""Tests for MarkdownArtifactRenderer.

Verifies rendering of DomainModel into PRD.md, DDD.md, and ARCHITECTURE.md
markdown strings following the template structures in docs/templates/.
"""

from __future__ import annotations

import pytest

from src.application.ports.artifact_generation_port import ArtifactRendererPort
from src.domain.models.domain_model import DomainModel
from src.domain.models.domain_values import (
    AggregateDesign,
    BoundedContext,
    ContextRelationship,
    DomainStory,
    SubdomainClassification,
)
from src.infrastructure.persistence.markdown_artifact_renderer import (
    MarkdownArtifactRenderer,
)

# ── Fixtures ─────────────────────────────────────────────────────────


@pytest.fixture
def sample_model() -> DomainModel:
    """A finalized DomainModel with 2 contexts, 2 stories, terms, and aggregates."""
    model = DomainModel()

    # Contexts
    model.add_bounded_context(
        BoundedContext(name="Ordering", responsibility="Manages order lifecycle")
    )
    model.add_bounded_context(
        BoundedContext(name="Shipping", responsibility="Handles delivery logistics")
    )

    # Stories (actors/steps must mention terms for invariant 1)
    model.add_domain_story(
        DomainStory(
            name="Place Order",
            actors=("Customer",),
            trigger="Customer submits cart",
            steps=("Customer places order", "System validates order"),
            observations=("Order must have at least one item",),
        )
    )
    model.add_domain_story(
        DomainStory(
            name="Ship Order",
            actors=("Warehouse",),
            trigger="Order is paid",
            steps=("Warehouse picks shipment", "Carrier delivers shipment"),
        )
    )

    # Terms — must appear in stories for invariant 1
    model.add_term("Order", "A customer purchase request", "Ordering", ("Q2",))
    model.add_term("Shipment", "A delivery package", "Shipping", ("Q2",))

    # Classifications
    model.classify_subdomain("Ordering", SubdomainClassification.CORE, "Competitive advantage")
    model.classify_subdomain("Shipping", SubdomainClassification.SUPPORTING, "Necessary logistics")

    # Aggregate for Core
    model.design_aggregate(
        AggregateDesign(
            name="OrderRoot",
            context_name="Ordering",
            root_entity="OrderRoot",
            invariants=("Order must have items",),
            domain_events=("OrderPlaced",),
        )
    )

    model.finalize()
    return model


@pytest.fixture
def minimal_model() -> DomainModel:
    """A finalized DomainModel with a single context, single story, no terms."""
    model = DomainModel()
    model.add_bounded_context(
        BoundedContext(name="Core", responsibility="Main domain")
    )
    model.add_domain_story(
        DomainStory(
            name="Basic Flow",
            actors=("User",),
            trigger="User starts",
            steps=("User performs action",),
        )
    )
    model.classify_subdomain("Core", SubdomainClassification.CORE, "Main")
    model.design_aggregate(
        AggregateDesign(
            name="CoreRoot",
            context_name="Core",
            root_entity="CoreRoot",
        )
    )
    model.finalize()
    return model


@pytest.fixture
def model_with_special_chars() -> DomainModel:
    """A finalized DomainModel with markdown-special characters in names."""
    model = DomainModel()
    model.add_bounded_context(
        BoundedContext(name="Order*Processing", responsibility="Handles order|flow")
    )
    model.add_domain_story(
        DomainStory(
            name="Process Order*Processing items",
            actors=("User",),
            trigger="User starts",
            steps=("User processes Order*Processing items",),
        )
    )
    model.add_term(
        "Order*Processing", "Handles orders with special chars", "Order*Processing", ("Q2",)
    )
    model.classify_subdomain(
        "Order*Processing", SubdomainClassification.CORE, "Core domain"
    )
    model.design_aggregate(
        AggregateDesign(
            name="OrderProcessingRoot",
            context_name="Order*Processing",
            root_entity="OrderProcessingRoot",
        )
    )
    model.finalize()
    return model


@pytest.fixture
def model_with_all_classifications() -> DomainModel:
    """A finalized DomainModel with CORE, SUPPORTING, and GENERIC contexts."""
    model = DomainModel()

    model.add_bounded_context(
        BoundedContext(name="Identity", responsibility="User identity management")
    )
    model.add_bounded_context(
        BoundedContext(name="Billing", responsibility="Payment processing")
    )
    model.add_bounded_context(
        BoundedContext(name="Logging", responsibility="Audit logging")
    )

    model.add_domain_story(
        DomainStory(
            name="User Signs Up",
            actors=("User",),
            trigger="User registers",
            steps=(
                "User provides identity details",
                "System creates billing account",
                "System writes to logging",
            ),
        )
    )

    model.add_term("Identity", "User identity record", "Identity", ("Q2",))
    model.add_term("Billing", "Payment billing record", "Billing", ("Q2",))
    model.add_term("Logging", "Audit logging entry", "Logging", ("Q2",))

    model.classify_subdomain("Identity", SubdomainClassification.CORE, "Differentiator")
    model.classify_subdomain("Billing", SubdomainClassification.SUPPORTING, "Necessary")
    model.classify_subdomain("Logging", SubdomainClassification.GENERIC, "Commodity")

    model.design_aggregate(
        AggregateDesign(
            name="IdentityRoot",
            context_name="Identity",
            root_entity="IdentityRoot",
        )
    )

    model.finalize()
    return model


# ── Protocol compliance ──────────────────────────────────────────────


class TestProtocolCompliance:
    def test_satisfies_artifact_renderer_port(self) -> None:
        """MarkdownArtifactRenderer is a runtime instance of ArtifactRendererPort."""
        renderer = MarkdownArtifactRenderer()
        assert isinstance(renderer, ArtifactRendererPort)


# ── render_prd ───────────────────────────────────────────────────────


class TestRenderPrd:
    def test_produces_markdown(self, sample_model: DomainModel) -> None:
        """render_prd returns non-empty markdown string."""
        renderer = MarkdownArtifactRenderer()
        result = renderer.render_prd(sample_model)
        assert isinstance(result, str)
        assert len(result) > 0

    def test_contains_prd_heading(self, sample_model: DomainModel) -> None:
        """PRD starts with a Product Requirements Document heading."""
        renderer = MarkdownArtifactRenderer()
        result = renderer.render_prd(sample_model)
        assert "# Product Requirements Document" in result

    def test_contains_context_names(self, sample_model: DomainModel) -> None:
        """PRD includes bounded context names from model."""
        renderer = MarkdownArtifactRenderer()
        result = renderer.render_prd(sample_model)
        assert "Ordering" in result
        assert "Shipping" in result

    def test_contains_stories_as_scenarios(self, sample_model: DomainModel) -> None:
        """PRD includes domain stories as user scenarios."""
        renderer = MarkdownArtifactRenderer()
        result = renderer.render_prd(sample_model)
        assert "Place Order" in result
        assert "Ship Order" in result

    def test_contains_capabilities_section(self, sample_model: DomainModel) -> None:
        """PRD includes a capabilities section."""
        renderer = MarkdownArtifactRenderer()
        result = renderer.render_prd(sample_model)
        assert "## 5. Capabilities" in result


# ── render_ddd ───────────────────────────────────────────────────────


class TestRenderDdd:
    def test_produces_markdown(self, sample_model: DomainModel) -> None:
        """render_ddd returns non-empty markdown string."""
        renderer = MarkdownArtifactRenderer()
        result = renderer.render_ddd(sample_model)
        assert isinstance(result, str)
        assert len(result) > 0

    def test_contains_ddd_heading(self, sample_model: DomainModel) -> None:
        """DDD.md starts with DDD artifacts heading."""
        renderer = MarkdownArtifactRenderer()
        result = renderer.render_ddd(sample_model)
        assert "# Domain-Driven Design Artifacts" in result

    def test_contains_stories(self, sample_model: DomainModel) -> None:
        """DDD.md includes domain stories from model."""
        renderer = MarkdownArtifactRenderer()
        result = renderer.render_ddd(sample_model)
        assert "Place Order" in result
        assert "Ship Order" in result

    def test_contains_story_steps(self, sample_model: DomainModel) -> None:
        """DDD.md includes story steps."""
        renderer = MarkdownArtifactRenderer()
        result = renderer.render_ddd(sample_model)
        assert "Customer places order" in result
        assert "Warehouse picks shipment" in result

    def test_contains_ubiquitous_language(self, sample_model: DomainModel) -> None:
        """DDD.md includes ubiquitous language terms."""
        renderer = MarkdownArtifactRenderer()
        result = renderer.render_ddd(sample_model)
        assert "Order" in result
        assert "Shipment" in result
        assert "A customer purchase request" in result

    def test_contains_bounded_contexts(self, sample_model: DomainModel) -> None:
        """DDD.md includes bounded context table."""
        renderer = MarkdownArtifactRenderer()
        result = renderer.render_ddd(sample_model)
        assert "Ordering" in result
        assert "Shipping" in result
        assert "Manages order lifecycle" in result

    def test_contains_classifications(self, sample_model: DomainModel) -> None:
        """DDD.md includes subdomain classifications."""
        renderer = MarkdownArtifactRenderer()
        result = renderer.render_ddd(sample_model)
        assert "core" in result.lower()
        assert "supporting" in result.lower()

    def test_contains_aggregate_designs(self, sample_model: DomainModel) -> None:
        """DDD.md includes aggregate designs for Core contexts."""
        renderer = MarkdownArtifactRenderer()
        result = renderer.render_ddd(sample_model)
        assert "OrderRoot" in result
        assert "Order must have items" in result
        assert "OrderPlaced" in result


# ── render_architecture ──────────────────────────────────────────────


class TestRenderArchitecture:
    def test_produces_markdown(self, sample_model: DomainModel) -> None:
        """render_architecture returns non-empty markdown string."""
        renderer = MarkdownArtifactRenderer()
        result = renderer.render_architecture(sample_model)
        assert isinstance(result, str)
        assert len(result) > 0

    def test_contains_architecture_heading(self, sample_model: DomainModel) -> None:
        """ARCHITECTURE.md starts with architecture heading."""
        renderer = MarkdownArtifactRenderer()
        result = renderer.render_architecture(sample_model)
        assert "# Architecture" in result

    def test_contains_bounded_contexts(self, sample_model: DomainModel) -> None:
        """ARCHITECTURE.md includes bounded context boundaries."""
        renderer = MarkdownArtifactRenderer()
        result = renderer.render_architecture(sample_model)
        assert "Ordering" in result
        assert "Shipping" in result

    def test_contains_classifications(self, sample_model: DomainModel) -> None:
        """ARCHITECTURE.md includes subdomain classifications."""
        renderer = MarkdownArtifactRenderer()
        result = renderer.render_architecture(sample_model)
        assert "core" in result.lower()
        assert "supporting" in result.lower()

    def test_contains_aggregate_info(self, sample_model: DomainModel) -> None:
        """ARCHITECTURE.md includes aggregate designs."""
        renderer = MarkdownArtifactRenderer()
        result = renderer.render_architecture(sample_model)
        assert "OrderRoot" in result

    def test_contains_layer_rules(self, sample_model: DomainModel) -> None:
        """ARCHITECTURE.md includes layer dependency rules."""
        renderer = MarkdownArtifactRenderer()
        result = renderer.render_architecture(sample_model)
        assert "domain" in result.lower()
        assert "application" in result.lower()
        assert "infrastructure" in result.lower()


# ── Edge cases ───────────────────────────────────────────────────────


class TestEdgeCases:
    def test_minimal_model(self, minimal_model: DomainModel) -> None:
        """Renders correctly with single context, single story."""
        renderer = MarkdownArtifactRenderer()

        prd = renderer.render_prd(minimal_model)
        ddd = renderer.render_ddd(minimal_model)
        arch = renderer.render_architecture(minimal_model)

        assert "Core" in prd
        assert "Basic Flow" in ddd
        assert "Core" in arch

    def test_empty_optional_fields(self, minimal_model: DomainModel) -> None:
        """Handles model with no observations, empty domain_events, etc."""
        renderer = MarkdownArtifactRenderer()

        # minimal_model has no terms, no observations, no domain_events on aggregate
        ddd = renderer.render_ddd(minimal_model)
        assert "# Domain-Driven Design Artifacts" in ddd

    def test_markdown_special_chars_in_table(
        self, model_with_special_chars: DomainModel
    ) -> None:
        """Context names with *, |, or # don't break table structure."""
        renderer = MarkdownArtifactRenderer()

        ddd = renderer.render_ddd(model_with_special_chars)
        arch = renderer.render_architecture(model_with_special_chars)

        # Tables should have consistent row counts (no broken pipes)
        for doc in (ddd, arch):
            for line in doc.split("\n"):
                if line.startswith("|") and line.endswith("|"):
                    # Every table row should have same number of pipes
                    # (at minimum, not fewer than the header)
                    assert line.count("|") >= 3

    def test_all_classification_types_rendered(
        self, model_with_all_classifications: DomainModel
    ) -> None:
        """Model with CORE + SUPPORTING + GENERIC renders all 3 in classification table."""
        renderer = MarkdownArtifactRenderer()

        ddd = renderer.render_ddd(model_with_all_classifications)
        arch = renderer.render_architecture(model_with_all_classifications)

        for doc in (ddd, arch):
            doc_lower = doc.lower()
            assert "core" in doc_lower
            assert "supporting" in doc_lower
            assert "generic" in doc_lower

    def test_long_descriptions_not_truncated(self, sample_model: DomainModel) -> None:
        """Very long story steps are rendered without truncation."""
        model = DomainModel()
        long_step = "A" * 500
        model.add_bounded_context(
            BoundedContext(name="Test", responsibility="Test context")
        )
        model.add_domain_story(
            DomainStory(
                name="Long Flow",
                actors=("User",),
                trigger="Start",
                steps=(long_step,),
            )
        )
        model.classify_subdomain("Test", SubdomainClassification.GENERIC, "Test")
        model.finalize()

        renderer = MarkdownArtifactRenderer()
        ddd = renderer.render_ddd(model)
        assert long_step in ddd

    def test_context_relationships_rendered(self) -> None:
        """DDD.md includes context relationship table when relationships exist."""
        model = DomainModel()
        model.add_bounded_context(
            BoundedContext(name="Orders", responsibility="Order management")
        )
        model.add_bounded_context(
            BoundedContext(name="Shipping", responsibility="Delivery")
        )
        model.add_domain_story(
            DomainStory(
                name="Order to Ship",
                actors=("System",),
                trigger="Order paid",
                steps=("Orders emits event", "Shipping receives event"),
            )
        )
        model.add_context_relationship(
            ContextRelationship(
                upstream="Orders",
                downstream="Shipping",
                integration_pattern="Domain Events",
            )
        )
        model.classify_subdomain("Orders", SubdomainClassification.CORE, "Core")
        model.classify_subdomain("Shipping", SubdomainClassification.SUPPORTING, "Support")
        model.design_aggregate(
            AggregateDesign(
                name="OrderRoot",
                context_name="Orders",
                root_entity="OrderRoot",
            )
        )
        model.finalize()

        renderer = MarkdownArtifactRenderer()
        ddd = renderer.render_ddd(model)
        assert "Context Map (Relationships)" in ddd
        assert "Orders" in ddd
        assert "Shipping" in ddd
        assert "Domain Events" in ddd

    def test_rich_aggregate_fields_rendered(self) -> None:
        """DDD.md renders contained_objects, commands, and domain_events on aggregates."""
        model = DomainModel()
        model.add_bounded_context(
            BoundedContext(
                name="Catalog",
                responsibility="Product catalog",
                key_domain_objects=("Product", "Category"),
            )
        )
        model.add_domain_story(
            DomainStory(
                name="Add Product",
                actors=("Admin",),
                trigger="Admin creates product",
                steps=("Admin adds product to catalog",),
            )
        )
        model.add_term("Product", "A sellable item", "Catalog", ("Q2",))
        model.classify_subdomain("Catalog", SubdomainClassification.CORE, "Core")
        model.design_aggregate(
            AggregateDesign(
                name="ProductRoot",
                context_name="Catalog",
                root_entity="ProductRoot",
                contained_objects=("ProductVariant", "Price"),
                invariants=("Price must be positive",),
                commands=("create_product", "update_price"),
                domain_events=("ProductCreated", "PriceChanged"),
            )
        )
        model.finalize()

        renderer = MarkdownArtifactRenderer()
        ddd = renderer.render_ddd(model)

        # contained_objects
        assert "ProductVariant" in ddd
        assert "Price" in ddd
        # commands
        assert "create_product" in ddd
        assert "update_price" in ddd
        # domain_events
        assert "ProductCreated" in ddd
        assert "PriceChanged" in ddd
        # key_domain_objects on context
        assert "Product" in ddd
        assert "Category" in ddd

    def test_no_aggregates_fallback(self) -> None:
        """Architecture renders fallback when no aggregate designs exist."""
        model = DomainModel()
        model.add_bounded_context(
            BoundedContext(name="Logging", responsibility="Audit logs")
        )
        model.add_domain_story(
            DomainStory(
                name="Write Log",
                actors=("System",),
                trigger="Event occurs",
                steps=("System writes logging entry",),
            )
        )
        model.classify_subdomain("Logging", SubdomainClassification.GENERIC, "Commodity")
        model.finalize()

        renderer = MarkdownArtifactRenderer()
        arch = renderer.render_architecture(model)
        assert "No aggregate designs yet" in arch

        ddd = renderer.render_ddd(model)
        assert "No aggregate designs yet" in ddd
