"""Tests for TicketGenerationHandler.

Covers the application-layer orchestration: building a TicketPlan
from a DomainModel, preview-before-write, and file output.
"""

from __future__ import annotations

from pathlib import Path

import pytest

from src.domain.models.domain_model import DomainModel
from src.domain.models.domain_values import (
    AggregateDesign,
    BoundedContext,
    DomainStory,
    SubdomainClassification,
)

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _make_model(
    contexts: list[tuple[str, SubdomainClassification]],
    aggregates: dict[str, list[str]] | None = None,
) -> DomainModel:
    """Build a minimal valid DomainModel with classified contexts."""
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
                    )
                )
        elif classification == SubdomainClassification.CORE:
            model.design_aggregate(
                AggregateDesign(
                    name=f"{name}Root",
                    context_name=name,
                    root_entity=f"{name}Root",
                    invariants=("must be valid",),
                )
            )

    model.finalize()
    return model


class FakeFileWriter:
    """In-memory file writer for testing."""

    def __init__(self) -> None:
        self.written: dict[str, str] = {}

    def write_file(self, path: Path, content: str) -> None:
        self.written[str(path)] = content


# ---------------------------------------------------------------------------
# 1. Build preview
# ---------------------------------------------------------------------------


class TestBuildPreview:
    def test_build_preview_returns_preview(self):
        from src.application.commands.ticket_generation_handler import (
            TicketGenerationHandler,
            TicketPreview,
        )

        model = _make_model([("Orders", SubdomainClassification.CORE)])
        writer = FakeFileWriter()
        handler = TicketGenerationHandler(writer=writer)

        preview = handler.build_preview(model)

        assert isinstance(preview, TicketPreview)
        assert preview.plan is not None
        assert preview.summary  # non-empty

    def test_preview_does_not_write(self):
        from src.application.commands.ticket_generation_handler import (
            TicketGenerationHandler,
        )

        model = _make_model([("Orders", SubdomainClassification.CORE)])
        writer = FakeFileWriter()
        handler = TicketGenerationHandler(writer=writer)
        handler.build_preview(model)

        assert writer.written == {}

    def test_preview_contains_bc_names(self):
        from src.application.commands.ticket_generation_handler import (
            TicketGenerationHandler,
        )

        model = _make_model(
            [
                ("Orders", SubdomainClassification.CORE),
                ("Logging", SubdomainClassification.GENERIC),
            ]
        )
        writer = FakeFileWriter()
        handler = TicketGenerationHandler(writer=writer)
        preview = handler.build_preview(model)

        assert "Orders" in preview.summary
        assert "Logging" in preview.summary

    def test_preview_empty_model_raises(self):
        from src.application.commands.ticket_generation_handler import (
            TicketGenerationHandler,
        )
        from src.domain.models.errors import InvariantViolationError

        model = DomainModel()
        writer = FakeFileWriter()
        handler = TicketGenerationHandler(writer=writer)

        with pytest.raises(InvariantViolationError, match="No bounded contexts"):
            handler.build_preview(model)


# ---------------------------------------------------------------------------
# 2. Approve and write
# ---------------------------------------------------------------------------


class TestApproveAndWrite:
    def test_approve_writes_files(self):
        from src.application.commands.ticket_generation_handler import (
            TicketGenerationHandler,
        )

        model = _make_model([("Orders", SubdomainClassification.CORE)])
        writer = FakeFileWriter()
        handler = TicketGenerationHandler(writer=writer)
        preview = handler.build_preview(model)

        handler.approve_and_write(preview, output_dir=Path("/project"))

        # Should write ticket files + summary
        assert len(writer.written) >= 2
        paths = list(writer.written.keys())
        assert any("PLAN_SUMMARY" in p for p in paths)
        assert any("tickets" in p for p in paths)

    def test_approve_emits_event(self):
        from src.application.commands.ticket_generation_handler import (
            TicketGenerationHandler,
        )
        from src.domain.events.ticket_events import TicketPlanApproved

        model = _make_model([("Orders", SubdomainClassification.CORE)])
        writer = FakeFileWriter()
        handler = TicketGenerationHandler(writer=writer)
        preview = handler.build_preview(model)

        handler.approve_and_write(preview, output_dir=Path("/project"))

        assert len(preview.plan.events) == 1
        assert isinstance(preview.plan.events[0], TicketPlanApproved)

    def test_approve_twice_raises(self):
        from src.application.commands.ticket_generation_handler import (
            TicketGenerationHandler,
        )
        from src.domain.models.errors import InvariantViolationError

        model = _make_model([("Orders", SubdomainClassification.CORE)])
        writer = FakeFileWriter()
        handler = TicketGenerationHandler(writer=writer)
        preview = handler.build_preview(model)

        handler.approve_and_write(preview, output_dir=Path("/project"))

        with pytest.raises(InvariantViolationError, match="already approved"):
            handler.approve_and_write(preview, output_dir=Path("/project"))

    def test_approve_subset_only_writes_approved(self):
        from src.application.commands.ticket_generation_handler import (
            TicketGenerationHandler,
        )

        model = _make_model(
            [
                ("Orders", SubdomainClassification.CORE),
                ("Logging", SubdomainClassification.GENERIC),
            ]
        )
        writer = FakeFileWriter()
        handler = TicketGenerationHandler(writer=writer)
        preview = handler.build_preview(model)

        first_id = preview.plan.tickets[0].ticket_id
        handler.approve_and_write(
            preview,
            output_dir=Path("/project"),
            approved_ids=(first_id,),
        )

        # Summary + 1 approved ticket = 2 files
        assert len(writer.written) == 2

    def test_no_generate_method(self):
        """Enforce preview-before-action: no generate() convenience."""
        from src.application.commands.ticket_generation_handler import (
            TicketGenerationHandler,
        )

        assert not hasattr(TicketGenerationHandler, "generate")


# ---------------------------------------------------------------------------
# 3. Port signature
# ---------------------------------------------------------------------------


class TestPortSignature:
    def test_port_accepts_domain_model(self):
        import inspect

        from src.application.ports.ticket_generation_port import TicketGenerationPort

        sig = inspect.signature(TicketGenerationPort.generate)
        params = list(sig.parameters.keys())
        assert "model" in params
        assert "output_dir" in params
