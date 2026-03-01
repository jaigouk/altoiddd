"""Application command handler for ticket pipeline generation.

TicketGenerationHandler builds a TicketPlan from a DomainModel's bounded
contexts and aggregate designs, generates dependency-ordered tickets with
complexity-budget-driven detail levels, and writes them via FileWriterPort.

Supports a preview-before-write workflow: build_preview() renders content
for user review, approve_and_write() commits approved content to disk.
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import TYPE_CHECKING

from src.domain.models.ticket_plan import TicketPlan

if TYPE_CHECKING:
    from pathlib import Path

    from src.application.ports.file_writer_port import FileWriterPort
    from src.domain.models.domain_model import DomainModel


@dataclass
class TicketPreview:
    """Generated ticket plan ready for user review before writing.

    Attributes:
        plan: The TicketPlan aggregate.
        summary: Human-readable preview summary.
    """

    plan: TicketPlan
    summary: str


class TicketGenerationHandler:
    """Orchestrates ticket pipeline generation from a DomainModel.

    Reads bounded contexts and aggregate designs from a finalized DomainModel,
    builds a TicketPlan with dependency-ordered tickets at appropriate detail
    levels, and writes ticket files via FileWriterPort.
    """

    def __init__(self, writer: FileWriterPort) -> None:
        self._writer = writer

    def build_preview(self, model: DomainModel) -> TicketPreview:
        """Build a ticket plan and render for preview without writing.

        Args:
            model: A finalized DomainModel with classified bounded contexts.

        Returns:
            TicketPreview with plan and summary.

        Raises:
            InvariantViolationError: If no bounded contexts in the model.
        """
        plan = TicketPlan()
        plan.generate_plan(model)

        return TicketPreview(
            plan=plan,
            summary=plan.preview(),
        )

    def approve_and_write(
        self,
        preview: TicketPreview,
        output_dir: Path,
        approved_ids: tuple[str, ...] | None = None,
    ) -> None:
        """Approve the plan (emitting domain event) and write tickets to disk.

        This is the only way to finalize tickets -- enforcing the
        preview-before-action pattern per ARCHITECTURE.md Design Principle 3.

        Args:
            preview: The TicketPreview from build_preview().
            output_dir: Directory where ticket files will be written.
            approved_ids: If provided, only approve these ticket IDs.
                          If None, all tickets are approved.
        """
        preview.plan.approve(approved_ids=approved_ids)

        # Determine which tickets to write
        event = preview.plan.events[-1]
        approved_set = set(event.approved_ticket_ids)

        for ticket in preview.plan.tickets:
            if ticket.ticket_id in approved_set:
                ticket_path = output_dir / "tickets" / f"{ticket.ticket_id}.md"
                self._writer.write_file(ticket_path, ticket.description)

        # Write plan summary
        summary_path = output_dir / "tickets" / "PLAN_SUMMARY.md"
        self._writer.write_file(summary_path, preview.summary)
