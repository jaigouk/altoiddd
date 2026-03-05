"""TicketPlan aggregate root for the Ticket Pipeline bounded context.

Generates dependency-ordered beads tickets from a DomainModel's bounded
contexts and aggregate designs. Detail level is driven by subdomain
classification (complexity budget).

Invariants:
1. One epic per bounded context -- no duplicates.
2. Dependency order is topologically sorted (foundation first).
3. Circular dependencies are detected and rejected.
4. Core subdomains produce FULL-detail tickets.
5. No cross-BC dependency references (tickets only depend within their BC
   or on tickets in upstream BCs per context relationships).
"""

from __future__ import annotations

import uuid
from collections import deque
from typing import TYPE_CHECKING

from src.domain.models.domain_values import AggregateDesign
from src.domain.models.errors import InvariantViolationError
from src.domain.models.ticket_values import (
    DependencyOrder,
    GeneratedEpic,
    GeneratedTicket,
    TicketDetailLevel,
    classify_tier,
    tier_to_detail_level,
)
from src.domain.services.ticket_detail_renderer import TicketDetailRenderer

if TYPE_CHECKING:
    from src.domain.events.ticket_events import TicketPlanApproved
    from src.domain.models.domain_model import DomainModel
    from src.domain.models.domain_values import SubdomainClassification
    from src.domain.models.stack_profile import StackProfile


class TicketPlan:
    """Aggregate root: generates and manages a ticket plan from a DomainModel.

    Attributes:
        plan_id: Unique identifier for this plan.
    """

    def __init__(self) -> None:
        self.plan_id: str = str(uuid.uuid4())
        self._epics: list[GeneratedEpic] = []
        self._tickets: list[GeneratedTicket] = []
        self._dependency_order: DependencyOrder | None = None
        self._events: list[TicketPlanApproved] = []
        self._approved: bool = False
        self._profile: StackProfile | None = None

    # -- Properties -----------------------------------------------------------

    @property
    def epics(self) -> tuple[GeneratedEpic, ...]:
        """All generated epics (defensive copy)."""
        return tuple(self._epics)

    @property
    def tickets(self) -> tuple[GeneratedTicket, ...]:
        """All generated tickets (defensive copy)."""
        return tuple(self._tickets)

    @property
    def dependency_order(self) -> DependencyOrder | None:
        """Topologically sorted ticket execution order, or None if not computed."""
        return self._dependency_order

    @property
    def events(self) -> tuple[TicketPlanApproved, ...]:
        """Domain events produced by this aggregate (defensive copy)."""
        return tuple(self._events)

    # -- Commands -------------------------------------------------------------

    def generate_plan(self, model: DomainModel, profile: StackProfile | None = None) -> None:
        """Generate epics and tickets from a finalized DomainModel.

        Creates one epic per bounded context, then generates tickets per
        aggregate with detail level driven by subdomain classification.
        Generic BCs with no aggregates get a single stub ticket.

        Args:
            model: A finalized DomainModel with classified bounded contexts.
            profile: Stack profile for quality gate rendering. Defaults to
                PythonUvProfile for backward compatibility.

        Raises:
            InvariantViolationError: If model has no bounded contexts,
                any BC lacks classification, or a cycle is detected.
        """
        if self._approved:
            msg = "Cannot regenerate plan on an approved TicketPlan"
            raise InvariantViolationError(msg)

        if profile is None:
            from src.domain.models.stack_profile import PythonUvProfile

            profile = PythonUvProfile()
        self._profile = profile

        contexts = model.bounded_contexts
        if not contexts:
            msg = "No bounded contexts to generate tickets for"
            raise InvariantViolationError(msg)

        self._epics.clear()
        self._tickets.clear()

        # Build a lookup of aggregates by context name
        aggregates_by_context: dict[str, list[AggregateDesign]] = {}
        for agg in model.aggregate_designs:
            aggregates_by_context.setdefault(agg.context_name, []).append(agg)

        # Build upstream lookup from context relationships
        upstream_contexts: dict[str, set[str]] = {}
        for rel in model.context_relationships:
            upstream_contexts.setdefault(rel.downstream, set()).add(rel.upstream)

        # Track epic IDs by context name for cross-referencing
        epic_id_by_context: dict[str, str] = {}

        for bc in contexts:
            if bc.classification is None:
                msg = f"Bounded context '{bc.name}' has no subdomain classification"
                raise InvariantViolationError(msg)

            epic_id = str(uuid.uuid4())
            epic_id_by_context[bc.name] = epic_id

            self._epics.append(
                GeneratedEpic(
                    epic_id=epic_id,
                    title=f"{bc.name} Epic",
                    description=(
                        f"Implement the {bc.name} bounded context "
                        f"({bc.classification.value} subdomain)."
                    ),
                    bounded_context_name=bc.name,
                    classification=bc.classification,
                )
            )

            detail_level = TicketDetailLevel.from_classification(bc.classification)
            context_aggregates = aggregates_by_context.get(bc.name, [])

            if not context_aggregates:
                # Generic/Supporting BCs with no aggregates get a stub ticket
                stub_agg = AggregateDesign(
                    name=bc.name,
                    context_name=bc.name,
                    root_entity=bc.name,
                )
                stub_description = TicketDetailRenderer.render(
                    stub_agg, TicketDetailLevel.STUB, profile
                )
                self._tickets.append(
                    GeneratedTicket(
                        ticket_id=str(uuid.uuid4()),
                        title=f"Integrate {bc.name} boundary",
                        description=stub_description,
                        detail_level=TicketDetailLevel.STUB,
                        epic_id=epic_id,
                        bounded_context_name=bc.name,
                        aggregate_name=bc.name,
                    )
                )
            else:
                for agg in context_aggregates:
                    description = TicketDetailRenderer.render(agg, detail_level, profile)
                    self._tickets.append(
                        GeneratedTicket(
                            ticket_id=str(uuid.uuid4()),
                            title=f"Implement {agg.name} aggregate",
                            description=description,
                            detail_level=detail_level,
                            epic_id=epic_id,
                            bounded_context_name=bc.name,
                            aggregate_name=agg.name,
                        )
                    )

        # Assign cross-BC dependencies based on context relationships
        self._assign_cross_bc_dependencies(upstream_contexts, epic_id_by_context)

        # Compute topological order (INV2, INV3)
        self._dependency_order = self._compute_dependency_order()

        # Two-tier generation: compute depths and reclassify (alty-2j7.11)
        # classification is guaranteed non-None (validated at line 127)
        classification_by_context: dict[str, SubdomainClassification] = {
            bc.name: bc.classification  # type: ignore[misc]
            for bc in contexts
        }
        self._reclassify_by_depth(classification_by_context, profile)

    def preview(self) -> str:
        """Return a human-readable preview of the generated plan.

        Returns:
            Summary showing epic count and ticket counts by detail level.

        Raises:
            InvariantViolationError: If no plan has been generated.
        """
        if not self._epics:
            msg = "No plan generated yet -- call generate_plan() first"
            raise InvariantViolationError(msg)

        full_count = sum(1 for t in self._tickets if t.detail_level == TicketDetailLevel.FULL)
        standard_count = sum(
            1 for t in self._tickets if t.detail_level == TicketDetailLevel.STANDARD
        )
        stub_count = sum(1 for t in self._tickets if t.detail_level == TicketDetailLevel.STUB)

        lines: list[str] = [
            f"Ticket Plan: {self.plan_id}",
            f"Epics: {len(self._epics)}",
            f"Tickets: {len(self._tickets)} "
            f"(FULL={full_count}, STANDARD={standard_count}, STUB={stub_count})",
            "",
        ]

        for epic in self._epics:
            epic_tickets = [t for t in self._tickets if t.epic_id == epic.epic_id]
            lines.append(f"  {epic.title} ({epic.classification.value.upper()}):")
            lines.extend(
                f"    - [{ticket.detail_level.value}] {ticket.title}" for ticket in epic_tickets
            )
            lines.append("")

        return "\n".join(lines)

    def promote_stub(self, ticket_id: str, profile: StackProfile | None = None) -> None:
        """Promote a STUB ticket to FULL detail.

        Args:
            ticket_id: ID of the stub ticket to promote.
            profile: Stack profile for quality gate rendering. Uses the
                profile from generate_plan() if not provided.

        Raises:
            InvariantViolationError: If ticket not found or not a STUB.
        """
        for i, ticket in enumerate(self._tickets):
            if ticket.ticket_id == ticket_id:
                if ticket.detail_level != TicketDetailLevel.STUB:
                    msg = (
                        f"Ticket '{ticket_id}' is {ticket.detail_level.value}, "
                        f"not STUB -- cannot promote"
                    )
                    raise InvariantViolationError(msg)

                # Find or create aggregate design for re-rendering
                resolved_profile = profile or self._profile
                if resolved_profile is None:
                    from src.domain.models.stack_profile import PythonUvProfile

                    resolved_profile = PythonUvProfile()
                agg = AggregateDesign(
                    name=ticket.aggregate_name,
                    context_name=ticket.bounded_context_name,
                    root_entity=ticket.aggregate_name,
                )
                new_description = TicketDetailRenderer.render(
                    agg, TicketDetailLevel.FULL, resolved_profile
                )
                self._tickets[i] = GeneratedTicket(
                    ticket_id=ticket.ticket_id,
                    title=ticket.title,
                    description=new_description,
                    detail_level=TicketDetailLevel.FULL,
                    epic_id=ticket.epic_id,
                    bounded_context_name=ticket.bounded_context_name,
                    aggregate_name=ticket.aggregate_name,
                    dependencies=ticket.dependencies,
                    depth=ticket.depth,
                )
                return

        msg = f"Ticket '{ticket_id}' not found"
        raise InvariantViolationError(msg)

    def approve(
        self,
        approved_ids: tuple[str, ...] | None = None,
    ) -> None:
        """Approve the plan (all or a subset), emitting TicketPlanApproved.

        Args:
            approved_ids: If provided, only these ticket IDs are approved.
                          If None, all tickets are approved.

        Raises:
            InvariantViolationError: If plan already approved, has no tickets,
                or approved_ids contains unknown ticket IDs.
        """
        if self._approved:
            msg = "Plan already approved"
            raise InvariantViolationError(msg)

        if not self._tickets:
            msg = "Cannot approve plan with no tickets"
            raise InvariantViolationError(msg)

        all_ids = {t.ticket_id for t in self._tickets}

        if approved_ids is None:
            final_approved = tuple(t.ticket_id for t in self._tickets)
            final_dismissed: tuple[str, ...] = ()
        else:
            unknown = set(approved_ids) - all_ids
            if unknown:
                msg = f"Unknown ticket IDs: {', '.join(sorted(unknown))}"
                raise InvariantViolationError(msg)
            final_approved = approved_ids
            final_dismissed = tuple(tid for tid in all_ids if tid not in set(approved_ids))

        self._approved = True

        from src.domain.events.ticket_events import TicketPlanApproved

        self._events.append(
            TicketPlanApproved(
                plan_id=self.plan_id,
                approved_ticket_ids=final_approved,
                dismissed_ticket_ids=final_dismissed,
            )
        )

    def promotion_eligible_ids(self, resolved_ids: frozenset[str]) -> frozenset[str]:
        """Return IDs of STUB tickets whose dependencies are all resolved.

        Args:
            resolved_ids: Ticket IDs that have been completed/resolved.

        Returns:
            Frozenset of ticket IDs eligible for promotion.
        """
        eligible: set[str] = set()
        for ticket in self._tickets:
            if ticket.detail_level != TicketDetailLevel.STUB:
                continue
            if not ticket.dependencies:
                continue
            if all(dep_id in resolved_ids for dep_id in ticket.dependencies):
                eligible.add(ticket.ticket_id)
        return frozenset(eligible)

    # -- Private helpers ------------------------------------------------------

    def _compute_ticket_depths(self) -> dict[str, int]:
        """Compute depth for each ticket using topological order.

        Depth 0 = tickets with no in-plan dependencies (roots).
        Depth N = 1 + max(dependency depths).

        Requires ``_dependency_order`` to be computed first.
        """
        if self._dependency_order is None:
            msg = "Dependency order must be computed before depths"
            raise InvariantViolationError(msg)

        all_ids = {t.ticket_id for t in self._tickets}
        deps_map = {
            t.ticket_id: [d for d in t.dependencies if d in all_ids]
            for t in self._tickets
        }
        depths: dict[str, int] = {}

        for tid in self._dependency_order.ordered_ids:
            dep_depths = [depths[d] for d in deps_map[tid] if d in depths]
            depths[tid] = (max(dep_depths) + 1) if dep_depths else 0

        return depths

    def _reclassify_by_depth(
        self,
        classification_by_context: dict[str, SubdomainClassification],
        profile: StackProfile,
    ) -> None:
        """Reclassify ticket detail levels using depth-based two-tier rules.

        Near-term tickets (depth ≤2) keep their classification-based detail.
        Far-term tickets (depth >2) are downgraded to STUB.
        Core tickets are always FULL regardless of depth.
        """
        depths = self._compute_ticket_depths()
        updated: list[GeneratedTicket] = []

        for ticket in self._tickets:
            depth = depths.get(ticket.ticket_id, 0)
            classification = classification_by_context[ticket.bounded_context_name]
            tier = classify_tier(depth, classification)
            new_level = tier_to_detail_level(tier, classification)

            if new_level != ticket.detail_level:
                # Re-render description at new detail level
                agg = AggregateDesign(
                    name=ticket.aggregate_name,
                    context_name=ticket.bounded_context_name,
                    root_entity=ticket.aggregate_name,
                )
                new_description = TicketDetailRenderer.render(agg, new_level, profile)
                updated.append(
                    GeneratedTicket(
                        ticket_id=ticket.ticket_id,
                        title=ticket.title,
                        description=new_description,
                        detail_level=new_level,
                        epic_id=ticket.epic_id,
                        bounded_context_name=ticket.bounded_context_name,
                        aggregate_name=ticket.aggregate_name,
                        dependencies=ticket.dependencies,
                        depth=depth,
                    )
                )
            else:
                # Keep same ticket but add depth
                updated.append(
                    GeneratedTicket(
                        ticket_id=ticket.ticket_id,
                        title=ticket.title,
                        description=ticket.description,
                        detail_level=ticket.detail_level,
                        epic_id=ticket.epic_id,
                        bounded_context_name=ticket.bounded_context_name,
                        aggregate_name=ticket.aggregate_name,
                        dependencies=ticket.dependencies,
                        depth=depth,
                    )
                )

        self._tickets = updated

    def _assign_cross_bc_dependencies(
        self,
        upstream_contexts: dict[str, set[str]],
        epic_id_by_context: dict[str, str],
    ) -> None:
        """Assign dependencies to tickets based on context relationships.

        Tickets in downstream BCs depend on tickets in upstream BCs.
        """
        # Build a map of context name -> list of ticket IDs
        tickets_by_context: dict[str, list[str]] = {}
        for ticket in self._tickets:
            tickets_by_context.setdefault(ticket.bounded_context_name, []).append(ticket.ticket_id)

        # For each downstream context, make its tickets depend on upstream tickets
        updated_tickets: list[GeneratedTicket] = []
        for ticket in self._tickets:
            ctx = ticket.bounded_context_name
            upstream_names = upstream_contexts.get(ctx, set())
            dep_ids: list[str] = list(ticket.dependencies)
            for upstream_name in upstream_names:
                upstream_ticket_ids = tickets_by_context.get(upstream_name, [])
                dep_ids.extend(upstream_ticket_ids)
            if dep_ids != list(ticket.dependencies):
                updated_tickets.append(
                    GeneratedTicket(
                        ticket_id=ticket.ticket_id,
                        title=ticket.title,
                        description=ticket.description,
                        detail_level=ticket.detail_level,
                        epic_id=ticket.epic_id,
                        bounded_context_name=ticket.bounded_context_name,
                        aggregate_name=ticket.aggregate_name,
                        dependencies=tuple(dep_ids),
                    )
                )
            else:
                updated_tickets.append(ticket)
        self._tickets = updated_tickets

    def _compute_dependency_order(self) -> DependencyOrder:
        """Compute topological sort of tickets using Kahn's algorithm.

        Returns:
            DependencyOrder with tickets in dependency-safe execution order.

        Raises:
            InvariantViolationError: If a dependency cycle is detected.
        """
        # Build adjacency and in-degree
        all_ids = {t.ticket_id for t in self._tickets}
        in_degree: dict[str, int] = {tid: 0 for tid in all_ids}
        dependents: dict[str, list[str]] = {tid: [] for tid in all_ids}

        for ticket in self._tickets:
            for dep_id in ticket.dependencies:
                if dep_id in all_ids:
                    in_degree[ticket.ticket_id] += 1
                    dependents[dep_id].append(ticket.ticket_id)

        # Kahn's algorithm
        queue: deque[str] = deque()
        for tid, degree in in_degree.items():
            if degree == 0:
                queue.append(tid)

        ordered: list[str] = []
        while queue:
            current = queue.popleft()
            ordered.append(current)
            for dependent in dependents[current]:
                in_degree[dependent] -= 1
                if in_degree[dependent] == 0:
                    queue.append(dependent)

        if len(ordered) != len(all_ids):
            msg = "Circular dependency detected in ticket plan"
            raise InvariantViolationError(msg)

        return DependencyOrder(ordered_ids=tuple(ordered))
