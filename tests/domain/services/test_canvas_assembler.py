"""Tests for CanvasAssembler domain service.

RED phase: all tests must FAIL because the module does not exist yet.
"""

from __future__ import annotations

from src.domain.models.domain_model import DomainModel
from src.domain.models.domain_values import (
    AggregateDesign,
    BoundedContext,
    ContextRelationship,
    DomainStory,
    SubdomainClassification,
)


def _make_finalized_model(
    *,
    contexts: tuple[BoundedContext, ...] = (),
    relationships: tuple[ContextRelationship, ...] = (),
    aggregates: tuple[AggregateDesign, ...] = (),
    stories: tuple[DomainStory, ...] | None = None,
    terms: tuple[tuple[str, str, str], ...] = (),
) -> DomainModel:
    """Build a finalized DomainModel with given artifacts.

    Args:
        contexts: Bounded contexts to add (must all have classification).
        relationships: Context relationships to add.
        aggregates: Aggregate designs (needed for Core contexts).
        stories: Domain stories. If None, a default story mentioning all terms is created.
        terms: (term, definition, context_name) tuples for UL.
    """
    model = DomainModel()

    # Add contexts.
    for ctx in contexts:
        model.add_bounded_context(ctx)

    # Add relationships.
    for rel in relationships:
        model.add_context_relationship(rel)

    # Build default story that mentions all terms if not provided.
    if stories is None:
        all_terms = [t[0] for t in terms] if terms else ["placeholder"]
        steps_text = ", ".join(all_terms)
        stories = (
            DomainStory(
                name="Default Story",
                actors=("Actor",),
                trigger=f"System processes {steps_text}",
                steps=(f"Actor manages {steps_text}",),
            ),
        )

    for story in stories:
        model.add_domain_story(story)

    # Add terms.
    for term, definition, ctx_name in terms:
        model.add_term(term, definition, ctx_name)

    # Classify contexts (required by invariant 2).
    for ctx in contexts:
        if ctx.classification is not None:
            model.classify_subdomain(ctx.name, ctx.classification)

    # Add aggregates.
    for agg in aggregates:
        model.design_aggregate(agg)

    model.finalize()
    return model


class TestAssembleEmptyModel:
    """assemble() with no bounded contexts."""

    def test_empty_model_returns_empty_tuple(self) -> None:
        from src.domain.services.canvas_assembler import CanvasAssembler

        model = _make_finalized_model()
        result = CanvasAssembler.assemble(model)
        assert result == ()


class TestAssembleSingleContext:
    """assemble() with one bounded context."""

    def test_single_core_context(self) -> None:
        from src.domain.models.bounded_context_canvas import DomainRole
        from src.domain.services.canvas_assembler import CanvasAssembler

        model = _make_finalized_model(
            contexts=(
                BoundedContext(
                    name="Sales",
                    responsibility="Manages order lifecycle",
                    classification=SubdomainClassification.CORE,
                ),
            ),
            aggregates=(
                AggregateDesign(
                    name="SalesRoot",
                    context_name="Sales",
                    root_entity="SalesRoot",
                ),
            ),
            terms=(("Order", "A purchase request", "Sales"),),
        )
        canvases = CanvasAssembler.assemble(model)
        assert len(canvases) == 1
        canvas = canvases[0]
        assert canvas.context_name == "Sales"
        assert canvas.purpose == "Manages order lifecycle"
        assert DomainRole.EXECUTION in canvas.domain_roles

    def test_single_supporting_context(self) -> None:
        from src.domain.models.bounded_context_canvas import DomainRole
        from src.domain.services.canvas_assembler import CanvasAssembler

        model = _make_finalized_model(
            contexts=(
                BoundedContext(
                    name="Notifications",
                    responsibility="Sends alerts",
                    classification=SubdomainClassification.SUPPORTING,
                ),
            ),
        )
        canvases = CanvasAssembler.assemble(model)
        assert len(canvases) == 1
        assert DomainRole.SPECIFICATION in canvases[0].domain_roles

    def test_single_generic_context(self) -> None:
        from src.domain.models.bounded_context_canvas import DomainRole
        from src.domain.services.canvas_assembler import CanvasAssembler

        model = _make_finalized_model(
            contexts=(
                BoundedContext(
                    name="Logging",
                    responsibility="Records events",
                    classification=SubdomainClassification.GENERIC,
                ),
            ),
        )
        canvases = CanvasAssembler.assemble(model)
        assert len(canvases) == 1
        assert DomainRole.GATEWAY in canvases[0].domain_roles


class TestAssembleMissingClassification:
    """assemble() when classification is None (fallback to GENERIC)."""

    def test_none_classification_falls_back_to_generic(self) -> None:
        from src.domain.services.canvas_assembler import CanvasAssembler

        # Build a model WITHOUT calling finalize (invariant 2 would reject None).
        model = DomainModel()
        model.add_bounded_context(
            BoundedContext(name="Orphan", responsibility="Unclassified context")
        )
        model.add_domain_story(
            DomainStory(
                name="Orphan Story",
                actors=("Actor",),
                trigger="System starts",
                steps=("Actor uses Orphan",),
            )
        )
        # Don't finalize — we test assembler directly on unfinalized model.
        canvases = CanvasAssembler.assemble(model)
        assert len(canvases) == 1
        assert canvases[0].classification.domain == SubdomainClassification.GENERIC
        assert canvases[0].classification.business_model == "unclassified"


class TestAssembleMultipleContexts:
    """assemble() with multiple bounded contexts."""

    def test_two_contexts_produce_two_canvases(self) -> None:
        from src.domain.services.canvas_assembler import CanvasAssembler

        model = _make_finalized_model(
            contexts=(
                BoundedContext(
                    name="Sales",
                    responsibility="Order management",
                    classification=SubdomainClassification.CORE,
                ),
                BoundedContext(
                    name="Inventory",
                    responsibility="Stock tracking",
                    classification=SubdomainClassification.SUPPORTING,
                ),
            ),
            aggregates=(
                AggregateDesign(
                    name="SalesRoot",
                    context_name="Sales",
                    root_entity="SalesRoot",
                ),
            ),
            terms=(
                ("Order", "A purchase request", "Sales"),
                ("Stock", "Available inventory", "Inventory"),
            ),
        )
        canvases = CanvasAssembler.assemble(model)
        assert len(canvases) == 2
        names = {c.context_name for c in canvases}
        assert names == {"Sales", "Inventory"}


class TestAssembleCommunication:
    """assemble() maps context relationships to inbound/outbound communication."""

    def test_relationship_maps_to_communication(self) -> None:
        from src.domain.services.canvas_assembler import CanvasAssembler

        model = _make_finalized_model(
            contexts=(
                BoundedContext(
                    name="Sales",
                    responsibility="Orders",
                    classification=SubdomainClassification.CORE,
                ),
                BoundedContext(
                    name="Shipping",
                    responsibility="Delivery",
                    classification=SubdomainClassification.SUPPORTING,
                ),
            ),
            relationships=(
                ContextRelationship(
                    upstream="Sales",
                    downstream="Shipping",
                    integration_pattern="Domain Events",
                ),
            ),
            aggregates=(
                AggregateDesign(
                    name="SalesRoot",
                    context_name="Sales",
                    root_entity="SalesRoot",
                ),
            ),
            terms=(("Order", "A purchase", "Sales"),),
        )
        canvases = CanvasAssembler.assemble(model)
        sales_canvas = next(c for c in canvases if c.context_name == "Sales")
        shipping_canvas = next(c for c in canvases if c.context_name == "Shipping")

        # Sales is upstream → outbound communication to Shipping.
        assert len(sales_canvas.outbound_communication) >= 1
        assert any(
            m.counterpart == "Shipping" for m in sales_canvas.outbound_communication
        )

        # Shipping is downstream → inbound communication from Sales.
        assert len(shipping_canvas.inbound_communication) >= 1
        assert any(
            m.counterpart == "Sales" for m in shipping_canvas.inbound_communication
        )


class TestAssembleUbiquitousLanguage:
    """assemble() filters UL terms per context."""

    def test_terms_filtered_by_context(self) -> None:
        from src.domain.services.canvas_assembler import CanvasAssembler

        model = _make_finalized_model(
            contexts=(
                BoundedContext(
                    name="Sales",
                    responsibility="Orders",
                    classification=SubdomainClassification.CORE,
                ),
                BoundedContext(
                    name="Inventory",
                    responsibility="Stock",
                    classification=SubdomainClassification.SUPPORTING,
                ),
            ),
            aggregates=(
                AggregateDesign(
                    name="SalesRoot",
                    context_name="Sales",
                    root_entity="SalesRoot",
                ),
            ),
            terms=(
                ("Order", "A purchase request", "Sales"),
                ("Stock", "Available items", "Inventory"),
            ),
        )
        canvases = CanvasAssembler.assemble(model)
        sales_canvas = next(c for c in canvases if c.context_name == "Sales")
        inv_canvas = next(c for c in canvases if c.context_name == "Inventory")

        sales_terms = dict(sales_canvas.ubiquitous_language)
        inv_terms = dict(inv_canvas.ubiquitous_language)

        assert "Order" in sales_terms
        assert "Stock" not in sales_terms
        assert "Stock" in inv_terms
        assert "Order" not in inv_terms

    def test_empty_ul_for_context(self) -> None:
        from src.domain.services.canvas_assembler import CanvasAssembler

        model = _make_finalized_model(
            contexts=(
                BoundedContext(
                    name="Logging",
                    responsibility="Records events",
                    classification=SubdomainClassification.GENERIC,
                ),
            ),
        )
        canvases = CanvasAssembler.assemble(model)
        assert canvases[0].ubiquitous_language == ()


class TestAssembleBusinessDecisions:
    """assemble() extracts invariants as business decisions."""

    def test_aggregate_invariants_become_business_decisions(self) -> None:
        from src.domain.services.canvas_assembler import CanvasAssembler

        model = _make_finalized_model(
            contexts=(
                BoundedContext(
                    name="Sales",
                    responsibility="Orders",
                    classification=SubdomainClassification.CORE,
                ),
            ),
            aggregates=(
                AggregateDesign(
                    name="SalesRoot",
                    context_name="Sales",
                    root_entity="SalesRoot",
                    invariants=("Order must have items", "Payment must be positive"),
                ),
            ),
            terms=(("Order", "A purchase", "Sales"),),
        )
        canvases = CanvasAssembler.assemble(model)
        canvas = canvases[0]
        assert "Order must have items" in canvas.business_decisions
        assert "Payment must be positive" in canvas.business_decisions


class TestAssembleAssumptionsAndQuestions:
    """assemble() sets assumptions and open_questions to empty (Round 2)."""

    def test_assumptions_empty(self) -> None:
        from src.domain.services.canvas_assembler import CanvasAssembler

        model = _make_finalized_model(
            contexts=(
                BoundedContext(
                    name="Sales",
                    responsibility="Orders",
                    classification=SubdomainClassification.CORE,
                ),
            ),
            aggregates=(
                AggregateDesign(
                    name="SalesRoot",
                    context_name="Sales",
                    root_entity="SalesRoot",
                ),
            ),
        )
        canvases = CanvasAssembler.assemble(model)
        assert canvases[0].assumptions == ()
        assert canvases[0].open_questions == ()


class TestRenderMarkdownEmpty:
    """render_markdown() with empty canvases."""

    def test_empty_canvases_returns_empty_string(self) -> None:
        from src.domain.services.canvas_assembler import CanvasAssembler

        result = CanvasAssembler.render_markdown(())
        assert result == ""


class TestRenderMarkdownSingleCanvas:
    """render_markdown() with one canvas."""

    def test_contains_context_name_heading(self) -> None:
        from src.domain.models.bounded_context_canvas import (
            BoundedContextCanvas,
            CommunicationMessage,
            DomainRole,
            StrategicClassification,
        )
        from src.domain.models.domain_values import SubdomainClassification
        from src.domain.services.canvas_assembler import CanvasAssembler

        canvas = BoundedContextCanvas(
            context_name="Sales",
            purpose="Manages orders",
            classification=StrategicClassification(
                domain=SubdomainClassification.CORE,
                business_model="Revenue",
                evolution="Custom",
            ),
            domain_roles=(DomainRole.EXECUTION,),
            inbound_communication=(
                CommunicationMessage(
                    message="PlaceOrder",
                    message_type="Command",
                    counterpart="API Gateway",
                ),
            ),
            outbound_communication=(
                CommunicationMessage(
                    message="OrderPlaced",
                    message_type="Event",
                    counterpart="Fulfillment",
                ),
            ),
            ubiquitous_language=(("Order", "A purchase request"),),
            business_decisions=("Order must have items",),
            assumptions=(),
            open_questions=(),
        )
        md = CanvasAssembler.render_markdown((canvas,))
        assert "# Bounded Context Canvas: Sales" in md

    def test_contains_purpose_section(self) -> None:
        from src.domain.models.bounded_context_canvas import (
            BoundedContextCanvas,
            DomainRole,
            StrategicClassification,
        )
        from src.domain.models.domain_values import SubdomainClassification
        from src.domain.services.canvas_assembler import CanvasAssembler

        canvas = BoundedContextCanvas(
            context_name="Sales",
            purpose="Manages orders",
            classification=StrategicClassification(
                domain=SubdomainClassification.CORE,
                business_model="Revenue",
                evolution="Custom",
            ),
            domain_roles=(DomainRole.EXECUTION,),
            inbound_communication=(),
            outbound_communication=(),
            ubiquitous_language=(),
            business_decisions=(),
            assumptions=(),
            open_questions=(),
        )
        md = CanvasAssembler.render_markdown((canvas,))
        assert "## Purpose" in md
        assert "Manages orders" in md

    def test_contains_strategic_classification_table(self) -> None:
        from src.domain.models.bounded_context_canvas import (
            BoundedContextCanvas,
            DomainRole,
            StrategicClassification,
        )
        from src.domain.models.domain_values import SubdomainClassification
        from src.domain.services.canvas_assembler import CanvasAssembler

        canvas = BoundedContextCanvas(
            context_name="Sales",
            purpose="Manages orders",
            classification=StrategicClassification(
                domain=SubdomainClassification.CORE,
                business_model="Revenue",
                evolution="Custom",
            ),
            domain_roles=(DomainRole.EXECUTION,),
            inbound_communication=(),
            outbound_communication=(),
            ubiquitous_language=(),
            business_decisions=(),
            assumptions=(),
            open_questions=(),
        )
        md = CanvasAssembler.render_markdown((canvas,))
        assert "## Strategic Classification" in md
        assert "core" in md.lower()
        assert "Revenue" in md
        assert "Custom" in md

    def test_contains_domain_roles_checklist(self) -> None:
        from src.domain.models.bounded_context_canvas import (
            BoundedContextCanvas,
            DomainRole,
            StrategicClassification,
        )
        from src.domain.models.domain_values import SubdomainClassification
        from src.domain.services.canvas_assembler import CanvasAssembler

        canvas = BoundedContextCanvas(
            context_name="Sales",
            purpose="Manages orders",
            classification=StrategicClassification(
                domain=SubdomainClassification.CORE,
                business_model="Revenue",
                evolution="Custom",
            ),
            domain_roles=(DomainRole.EXECUTION, DomainRole.ANALYSIS),
            inbound_communication=(),
            outbound_communication=(),
            ubiquitous_language=(),
            business_decisions=(),
            assumptions=(),
            open_questions=(),
        )
        md = CanvasAssembler.render_markdown((canvas,))
        assert "## Domain Roles" in md
        assert "execution" in md.lower()
        assert "analysis" in md.lower()

    def test_contains_communication_tables(self) -> None:
        from src.domain.models.bounded_context_canvas import (
            BoundedContextCanvas,
            CommunicationMessage,
            DomainRole,
            StrategicClassification,
        )
        from src.domain.models.domain_values import SubdomainClassification
        from src.domain.services.canvas_assembler import CanvasAssembler

        canvas = BoundedContextCanvas(
            context_name="Sales",
            purpose="Manages orders",
            classification=StrategicClassification(
                domain=SubdomainClassification.CORE,
                business_model="Revenue",
                evolution="Custom",
            ),
            domain_roles=(DomainRole.EXECUTION,),
            inbound_communication=(
                CommunicationMessage(
                    message="PlaceOrder",
                    message_type="Command",
                    counterpart="API Gateway",
                ),
            ),
            outbound_communication=(
                CommunicationMessage(
                    message="OrderPlaced",
                    message_type="Event",
                    counterpart="Fulfillment",
                ),
            ),
            ubiquitous_language=(),
            business_decisions=(),
            assumptions=(),
            open_questions=(),
        )
        md = CanvasAssembler.render_markdown((canvas,))
        assert "## Inbound Communication" in md
        assert "PlaceOrder" in md
        assert "## Outbound Communication" in md
        assert "OrderPlaced" in md

    def test_contains_ubiquitous_language_table(self) -> None:
        from src.domain.models.bounded_context_canvas import (
            BoundedContextCanvas,
            DomainRole,
            StrategicClassification,
        )
        from src.domain.models.domain_values import SubdomainClassification
        from src.domain.services.canvas_assembler import CanvasAssembler

        canvas = BoundedContextCanvas(
            context_name="Sales",
            purpose="Manages orders",
            classification=StrategicClassification(
                domain=SubdomainClassification.CORE,
                business_model="Revenue",
                evolution="Custom",
            ),
            domain_roles=(DomainRole.EXECUTION,),
            inbound_communication=(),
            outbound_communication=(),
            ubiquitous_language=(("Order", "A purchase request"),),
            business_decisions=(),
            assumptions=(),
            open_questions=(),
        )
        md = CanvasAssembler.render_markdown((canvas,))
        assert "## Ubiquitous Language" in md
        assert "Order" in md
        assert "A purchase request" in md

    def test_contains_business_decisions(self) -> None:
        from src.domain.models.bounded_context_canvas import (
            BoundedContextCanvas,
            DomainRole,
            StrategicClassification,
        )
        from src.domain.models.domain_values import SubdomainClassification
        from src.domain.services.canvas_assembler import CanvasAssembler

        canvas = BoundedContextCanvas(
            context_name="Sales",
            purpose="Manages orders",
            classification=StrategicClassification(
                domain=SubdomainClassification.CORE,
                business_model="Revenue",
                evolution="Custom",
            ),
            domain_roles=(DomainRole.EXECUTION,),
            inbound_communication=(),
            outbound_communication=(),
            ubiquitous_language=(),
            business_decisions=("Order must have items",),
            assumptions=(),
            open_questions=(),
        )
        md = CanvasAssembler.render_markdown((canvas,))
        assert "## Business Decisions" in md
        assert "Order must have items" in md

    def test_special_chars_in_markdown(self) -> None:
        """Special characters in context name render properly."""
        from src.domain.models.bounded_context_canvas import (
            BoundedContextCanvas,
            DomainRole,
            StrategicClassification,
        )
        from src.domain.models.domain_values import SubdomainClassification
        from src.domain.services.canvas_assembler import CanvasAssembler

        canvas = BoundedContextCanvas(
            context_name='Auth & Identity "Service"',
            purpose="Handles auth",
            classification=StrategicClassification(
                domain=SubdomainClassification.SUPPORTING,
                business_model="Compliance",
                evolution="Product",
            ),
            domain_roles=(DomainRole.SPECIFICATION,),
            inbound_communication=(),
            outbound_communication=(),
            ubiquitous_language=(),
            business_decisions=(),
            assumptions=(),
            open_questions=(),
        )
        md = CanvasAssembler.render_markdown((canvas,))
        assert 'Auth & Identity "Service"' in md
