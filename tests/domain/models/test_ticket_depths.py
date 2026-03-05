"""Tests for depth-based ticket reclassification (alty-2j7.11).

Integration tests verifying that TicketPlan.generate_plan() computes depths
and reclassifies tickets: near-term (depth ≤2) keeps classification-based
detail, far-term (depth >2) is downgraded to STUB.
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
from src.domain.models.ticket_plan import TicketPlan
from src.domain.models.ticket_values import TicketDetailLevel

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _make_chain_model(
    chain_length: int,
    classification: SubdomainClassification = SubdomainClassification.SUPPORTING,
) -> DomainModel:
    """Build a model with a chain of BCs: A→B→C→...

    Each BC has one aggregate. Context relationships create a dependency chain:
    BC_0 (upstream) → BC_1 (downstream) → BC_2 → ...

    This produces tickets at depths 0, 1, 2, ... (chain_length - 1).
    """
    model = DomainModel()
    names = [f"BC_{i}" for i in range(chain_length)]

    model.add_domain_story(
        DomainStory(
            name="Chain flow",
            actors=("User",),
            trigger="User starts",
            steps=tuple(f"User manages {name}" for name in names),
        )
    )

    for name in names:
        model.add_term(term=name, definition=f"{name} domain", context_name=name)
        model.add_bounded_context(
            BoundedContext(
                name=name,
                responsibility=f"Manages {name}",
                classification=classification,
            )
        )
        model.design_aggregate(
            AggregateDesign(
                name=f"{name}Root",
                context_name=name,
                root_entity=f"{name}Root",
                invariants=("must be valid",),
                commands=("Create",),
                domain_events=("Created",),
            )
        )

    # Create chain: BC_0 upstream of BC_1, BC_1 upstream of BC_2, etc.
    for i in range(chain_length - 1):
        model.add_context_relationship(
            ContextRelationship(
                upstream=names[i],
                downstream=names[i + 1],
                integration_pattern="Domain Events",
            )
        )

    model.finalize()
    return model


# ---------------------------------------------------------------------------
# Depth computation
# ---------------------------------------------------------------------------


class TestDepthComputation:
    """Verify tickets get correct depth values."""

    def test_root_tickets_have_depth_0(self):
        model = _make_chain_model(1)
        plan = TicketPlan()
        plan.generate_plan(model)

        assert all(t.depth == 0 for t in plan.tickets)

    def test_chain_depth_increments(self):
        model = _make_chain_model(4)
        plan = TicketPlan()
        plan.generate_plan(model)

        depth_by_context = {t.bounded_context_name: t.depth for t in plan.tickets}
        assert depth_by_context["BC_0"] == 0
        assert depth_by_context["BC_1"] == 1
        assert depth_by_context["BC_2"] == 2
        assert depth_by_context["BC_3"] == 3

    def test_diamond_dag_uses_max_depth(self):
        """Diamond: A→C, B→C. Both A and B at depth 0, C gets depth 1."""
        model = DomainModel()
        names = ["A", "B", "C"]
        model.add_domain_story(
            DomainStory(
                name="Diamond flow",
                actors=("User",),
                trigger="User starts",
                steps=tuple(f"User manages {n}" for n in names),
            )
        )
        for name in names:
            model.add_term(term=name, definition=f"{name} domain", context_name=name)
            model.add_bounded_context(
                BoundedContext(
                    name=name,
                    responsibility=f"Manages {name}",
                    classification=SubdomainClassification.SUPPORTING,
                )
            )
            model.design_aggregate(
                AggregateDesign(
                    name=f"{name}Root",
                    context_name=name,
                    root_entity=f"{name}Root",
                )
            )
        model.add_context_relationship(
            ContextRelationship(upstream="A", downstream="C", integration_pattern="ACL")
        )
        model.add_context_relationship(
            ContextRelationship(upstream="B", downstream="C", integration_pattern="ACL")
        )
        model.finalize()

        plan = TicketPlan()
        plan.generate_plan(model)

        depth_by_context = {t.bounded_context_name: t.depth for t in plan.tickets}
        assert depth_by_context["A"] == 0
        assert depth_by_context["B"] == 0
        assert depth_by_context["C"] == 1

    def test_orphan_ticket_depth_0(self):
        """Tickets with no dependencies get depth 0."""
        model = DomainModel()
        model.add_domain_story(
            DomainStory(
                name="Orphan flow",
                actors=("User",),
                trigger="User starts",
                steps=("User manages Orphan",),
            )
        )
        model.add_term(term="Orphan", definition="Orphan domain", context_name="Orphan")
        model.add_bounded_context(
            BoundedContext(
                name="Orphan",
                responsibility="Manages Orphan",
                classification=SubdomainClassification.SUPPORTING,
            )
        )
        model.design_aggregate(
            AggregateDesign(
                name="OrphanRoot",
                context_name="Orphan",
                root_entity="OrphanRoot",
            )
        )
        model.finalize()

        plan = TicketPlan()
        plan.generate_plan(model)

        assert plan.tickets[0].depth == 0


# ---------------------------------------------------------------------------
# Depth-based reclassification
# ---------------------------------------------------------------------------


class TestDepthReclassification:
    """Verify detail level changes based on depth."""

    def test_supporting_at_depth_2_stays_standard(self):
        """Supporting BC at depth 2 (boundary) keeps STANDARD."""
        model = _make_chain_model(3)  # depths 0, 1, 2
        plan = TicketPlan()
        plan.generate_plan(model)

        ticket_at_depth_2 = next(t for t in plan.tickets if t.depth == 2)
        assert ticket_at_depth_2.detail_level == TicketDetailLevel.STANDARD

    def test_supporting_at_depth_3_becomes_stub(self):
        """Supporting BC at depth 3 (past boundary) becomes STUB."""
        model = _make_chain_model(4)  # depths 0, 1, 2, 3
        plan = TicketPlan()
        plan.generate_plan(model)

        ticket_at_depth_3 = next(t for t in plan.tickets if t.depth == 3)
        assert ticket_at_depth_3.detail_level == TicketDetailLevel.STUB

    def test_core_at_depth_5_stays_full(self):
        """Core BC at any depth stays FULL (override)."""
        model = _make_chain_model(6, classification=SubdomainClassification.CORE)
        plan = TicketPlan()
        plan.generate_plan(model)

        ticket_at_depth_5 = next(t for t in plan.tickets if t.depth == 5)
        assert ticket_at_depth_5.detail_level == TicketDetailLevel.FULL

    def test_stub_description_enriched_at_depth_3(self):
        """Far-term stub tickets get the enriched stub template."""
        model = _make_chain_model(4)
        plan = TicketPlan()
        plan.generate_plan(model)

        ticket_at_depth_3 = next(t for t in plan.tickets if t.depth == 3)
        assert ticket_at_depth_3.detail_level == TicketDetailLevel.STUB
        assert "Stub ticket" in ticket_at_depth_3.description
        assert "## DDD Alignment" in ticket_at_depth_3.description


# ---------------------------------------------------------------------------
# Promotion eligibility
# ---------------------------------------------------------------------------


class TestPromotionEligibility:
    """Verify promotion_eligible_ids() identifies promotable stubs."""

    def test_stub_eligible_when_all_deps_resolved(self):
        model = _make_chain_model(4)  # depth 3 = STUB
        plan = TicketPlan()
        plan.generate_plan(model)

        stub = next(t for t in plan.tickets if t.detail_level == TicketDetailLevel.STUB)
        all_other_ids = frozenset(
            t.ticket_id for t in plan.tickets if t.ticket_id != stub.ticket_id
        )

        eligible = plan.promotion_eligible_ids(all_other_ids)
        assert stub.ticket_id in eligible

    def test_stub_not_eligible_when_deps_unresolved(self):
        model = _make_chain_model(4)
        plan = TicketPlan()
        plan.generate_plan(model)

        stub = next(t for t in plan.tickets if t.detail_level == TicketDetailLevel.STUB)
        # No resolved tickets
        eligible = plan.promotion_eligible_ids(frozenset())
        assert stub.ticket_id not in eligible

    def test_non_stub_not_eligible(self):
        model = _make_chain_model(2)  # depths 0, 1 — all STANDARD
        plan = TicketPlan()
        plan.generate_plan(model)

        all_ids = frozenset(t.ticket_id for t in plan.tickets)
        eligible = plan.promotion_eligible_ids(all_ids)
        assert len(eligible) == 0
