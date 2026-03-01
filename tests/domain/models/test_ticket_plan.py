"""Tests for TicketPlan aggregate root.

Covers plan generation from DomainModel, detail level mapping,
dependency ordering, preview, promote, and approve workflows.
"""

from __future__ import annotations

import pytest

from src.domain.models.domain_model import DomainModel
from src.domain.models.domain_values import (
    AggregateDesign,
    BoundedContext,
    DomainStory,
    SubdomainClassification,
)
from src.domain.models.errors import InvariantViolationError
from src.domain.models.ticket_plan import TicketPlan
from src.domain.models.ticket_values import TicketDetailLevel

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _make_model(
    contexts: list[tuple[str, SubdomainClassification]],
    aggregates: dict[str, list[str]] | None = None,
) -> DomainModel:
    """Build a minimal valid DomainModel with classified contexts and aggregates.

    Args:
        contexts: List of (context_name, classification) pairs.
        aggregates: Optional dict of context_name -> list of aggregate names.
                    Core contexts get a default aggregate if not provided.
    """
    model = DomainModel()

    all_names = [name for name, _ in contexts]
    model.add_domain_story(
        DomainStory(
            name="Test flow",
            actors=("User",),
            trigger="User starts",
            steps=tuple(f"User manages {name}" for name in all_names),
        )
    )

    for name, classification in contexts:
        model.add_term(
            term=name,
            definition=f"{name} domain",
            context_name=name,
        )
        model.add_bounded_context(
            BoundedContext(
                name=name,
                responsibility=f"Manages {name}",
                classification=classification,
            )
        )

    aggregates = aggregates or {}
    for name, classification in contexts:
        if name in aggregates:
            for agg_name in aggregates[name]:
                model.design_aggregate(
                    AggregateDesign(
                        name=agg_name,
                        context_name=name,
                        root_entity=agg_name,
                        invariants=("must be valid",),
                        commands=("Create",),
                        domain_events=("Created",),
                    )
                )
        elif classification == SubdomainClassification.CORE:
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

    model.finalize()
    return model


# ---------------------------------------------------------------------------
# 1. Empty plan
# ---------------------------------------------------------------------------


class TestEmptyPlan:
    def test_empty_plan_has_no_epics(self):
        plan = TicketPlan()
        assert plan.epics == ()
        assert plan.tickets == ()
        assert plan.dependency_order is None
        assert plan.events == ()


# ---------------------------------------------------------------------------
# 2. One epic per BC (INV1)
# ---------------------------------------------------------------------------


class TestOneEpicPerBC:
    def test_generate_one_epic_per_bc(self):
        model = _make_model([
            ("Orders", SubdomainClassification.CORE),
            ("Shipping", SubdomainClassification.SUPPORTING),
        ], aggregates={"Shipping": ["ShipmentRoot"]})

        plan = TicketPlan()
        plan.generate_plan(model)

        assert len(plan.epics) == 2
        epic_names = {e.bounded_context_name for e in plan.epics}
        assert epic_names == {"Orders", "Shipping"}


# ---------------------------------------------------------------------------
# 3. Detail level mapping (INV4)
# ---------------------------------------------------------------------------


class TestDetailLevelMapping:
    def test_core_tickets_full_detail(self):
        model = _make_model([("Orders", SubdomainClassification.CORE)])
        plan = TicketPlan()
        plan.generate_plan(model)

        for ticket in plan.tickets:
            if ticket.bounded_context_name == "Orders":
                assert ticket.detail_level == TicketDetailLevel.FULL

    def test_supporting_standard_detail(self):
        model = _make_model(
            [("Shipping", SubdomainClassification.SUPPORTING)],
            aggregates={"Shipping": ["ShipmentRoot"]},
        )
        plan = TicketPlan()
        plan.generate_plan(model)

        for ticket in plan.tickets:
            if ticket.bounded_context_name == "Shipping":
                assert ticket.detail_level == TicketDetailLevel.STANDARD

    def test_generic_stub_detail(self):
        model = _make_model([("Logging", SubdomainClassification.GENERIC)])
        plan = TicketPlan()
        plan.generate_plan(model)

        for ticket in plan.tickets:
            if ticket.bounded_context_name == "Logging":
                assert ticket.detail_level == TicketDetailLevel.STUB


# ---------------------------------------------------------------------------
# 4. Preview
# ---------------------------------------------------------------------------


class TestPreview:
    def test_preview_shows_summary(self):
        model = _make_model([
            ("Orders", SubdomainClassification.CORE),
            ("Logging", SubdomainClassification.GENERIC),
        ])
        plan = TicketPlan()
        plan.generate_plan(model)

        summary = plan.preview()
        assert "Epics: 2" in summary
        assert "FULL=" in summary
        assert "STUB=" in summary
        assert "Orders" in summary
        assert "Logging" in summary

    def test_preview_before_generate_raises(self):
        plan = TicketPlan()
        with pytest.raises(InvariantViolationError, match="No plan generated"):
            plan.preview()


# ---------------------------------------------------------------------------
# 5. Promote stub
# ---------------------------------------------------------------------------


class TestPromoteStub:
    def test_promote_stub_to_full(self):
        model = _make_model([("Logging", SubdomainClassification.GENERIC)])
        plan = TicketPlan()
        plan.generate_plan(model)

        stub_ticket = plan.tickets[0]
        assert stub_ticket.detail_level == TicketDetailLevel.STUB

        plan.promote_stub(stub_ticket.ticket_id)
        promoted = plan.tickets[0]
        assert promoted.detail_level == TicketDetailLevel.FULL
        assert "## Goal" in promoted.description
        assert "## SOLID Mapping" in promoted.description

    def test_promote_non_stub_raises(self):
        model = _make_model([("Orders", SubdomainClassification.CORE)])
        plan = TicketPlan()
        plan.generate_plan(model)

        full_ticket = plan.tickets[0]
        assert full_ticket.detail_level == TicketDetailLevel.FULL

        with pytest.raises(InvariantViolationError, match="not STUB"):
            plan.promote_stub(full_ticket.ticket_id)

    def test_promote_unknown_ticket_raises(self):
        model = _make_model([("Logging", SubdomainClassification.GENERIC)])
        plan = TicketPlan()
        plan.generate_plan(model)

        with pytest.raises(InvariantViolationError, match="not found"):
            plan.promote_stub("nonexistent-id")


# ---------------------------------------------------------------------------
# 6. Approve
# ---------------------------------------------------------------------------


class TestApprove:
    def test_approve_all_emits_event(self):
        model = _make_model([("Orders", SubdomainClassification.CORE)])
        plan = TicketPlan()
        plan.generate_plan(model)

        plan.approve()

        assert len(plan.events) == 1
        event = plan.events[0]
        assert event.plan_id == plan.plan_id
        assert len(event.approved_ticket_ids) == len(plan.tickets)
        assert event.dismissed_ticket_ids == ()

    def test_approve_subset(self):
        model = _make_model([
            ("Orders", SubdomainClassification.CORE),
            ("Logging", SubdomainClassification.GENERIC),
        ])
        plan = TicketPlan()
        plan.generate_plan(model)

        first_ticket_id = plan.tickets[0].ticket_id
        plan.approve(approved_ids=(first_ticket_id,))

        event = plan.events[0]
        assert event.approved_ticket_ids == (first_ticket_id,)
        assert len(event.dismissed_ticket_ids) == len(plan.tickets) - 1

    def test_approve_twice_raises(self):
        model = _make_model([("Orders", SubdomainClassification.CORE)])
        plan = TicketPlan()
        plan.generate_plan(model)

        plan.approve()

        with pytest.raises(InvariantViolationError, match="already approved"):
            plan.approve()

    def test_approve_empty_plan_raises(self):
        plan = TicketPlan()
        with pytest.raises(InvariantViolationError, match="no tickets"):
            plan.approve()


# ---------------------------------------------------------------------------
# 7. Error cases
# ---------------------------------------------------------------------------


class TestErrorCases:
    def test_empty_model_raises(self):
        model = DomainModel()
        plan = TicketPlan()

        with pytest.raises(InvariantViolationError, match="No bounded contexts"):
            plan.generate_plan(model)

    def test_no_classification_raises(self):
        model = DomainModel()
        model.add_domain_story(
            DomainStory(
                name="Test flow",
                actors=("User",),
                trigger="User starts",
                steps=("User manages Orders",),
            )
        )
        model.add_term(term="Orders", definition="Orders domain", context_name="Orders")
        model.add_bounded_context(
            BoundedContext(
                name="Orders",
                responsibility="Manages Orders",
                classification=None,
            )
        )
        # Skip finalize since we need the unclassified BC
        plan = TicketPlan()

        with pytest.raises(InvariantViolationError, match="no subdomain classification"):
            plan.generate_plan(model)

    def test_regenerate_after_approve_raises(self):
        model = _make_model([("Orders", SubdomainClassification.CORE)])
        plan = TicketPlan()
        plan.generate_plan(model)
        plan.approve()

        with pytest.raises(InvariantViolationError, match="Cannot regenerate"):
            plan.generate_plan(model)
