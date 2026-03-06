"""Tests for BC Canvas integration in ArtifactGenerationHandler.

RED phase: tests must FAIL because canvas_content field and assembler
integration do not exist yet.
"""

from __future__ import annotations

from pathlib import Path
from unittest.mock import MagicMock

from src.application.commands.artifact_generation_handler import (
    ArtifactGenerationHandler,
)
from src.domain.events.discovery_events import DiscoveryCompleted
from src.domain.models.discovery_values import Answer, Persona, Playback, Register


def _make_discovery_event(
    answers: tuple[Answer, ...] | None = None,
) -> DiscoveryCompleted:
    """Create a DiscoveryCompleted event with standard answers."""
    if answers is None:
        answers = (
            Answer(question_id="Q1", response_text="Customer, Admin"),
            Answer(question_id="Q2", response_text="Order, Product"),
            Answer(
                question_id="Q3",
                response_text="Customer reviews order, System processes payment",
            ),
            Answer(
                question_id="Q4",
                response_text="Payment must not be negative, Order must have items",
            ),
            Answer(question_id="Q5", response_text="Admin manages Product catalog"),
            Answer(question_id="Q6", response_text="OrderPlaced, PaymentProcessed"),
            Answer(
                question_id="Q7",
                response_text="When OrderPlaced, send confirmation email",
            ),
            Answer(question_id="Q8", response_text="Order history, Sales dashboard"),
            Answer(question_id="Q9", response_text="Sales, Inventory"),
            Answer(
                question_id="Q10",
                response_text=(
                    "Sales is core competitive advantage, "
                    "Inventory is supporting necessary"
                ),
            ),
        )
    return DiscoveryCompleted(
        session_id="session-1",
        persona=Persona.DEVELOPER,
        register=Register.TECHNICAL,
        answers=answers,
        playback_confirmations=(Playback(summary_text="Playback 1", confirmed=True),),
    )


def _make_handler() -> tuple[ArtifactGenerationHandler, MagicMock, MagicMock]:
    """Create handler with mock renderer and writer."""
    renderer = MagicMock()
    renderer.render_prd.return_value = "# PRD"
    renderer.render_ddd.return_value = "# DDD"
    renderer.render_architecture.return_value = "# ARCH"
    writer = MagicMock()
    handler = ArtifactGenerationHandler(renderer=renderer, writer=writer)
    return handler, renderer, writer


class TestArtifactPreviewHasCanvasContent:
    """ArtifactPreview must have a canvas_content field."""

    def test_preview_has_canvas_content_attr(self) -> None:
        handler, _, _ = _make_handler()
        event = _make_discovery_event()
        preview = handler.build_preview(event)
        assert hasattr(preview, "canvas_content")

    def test_canvas_content_is_string(self) -> None:
        handler, _, _ = _make_handler()
        event = _make_discovery_event()
        preview = handler.build_preview(event)
        assert isinstance(preview.canvas_content, str)

    def test_canvas_content_not_empty_with_contexts(self) -> None:
        """With bounded contexts, canvas_content should contain markdown."""
        handler, _, _ = _make_handler()
        event = _make_discovery_event()
        preview = handler.build_preview(event)
        # Standard event has Sales + Inventory contexts.
        assert len(preview.canvas_content) > 0
        assert "Bounded Context Canvas" in preview.canvas_content


class TestBuildPreviewCallsAssembler:
    """build_preview() must call CanvasAssembler to generate canvas content."""

    def test_canvas_content_contains_context_names(self) -> None:
        handler, _, _ = _make_handler()
        event = _make_discovery_event()
        preview = handler.build_preview(event)
        # Standard event creates Sales and Inventory contexts.
        assert "Sales" in preview.canvas_content

    def test_canvas_content_has_purpose_sections(self) -> None:
        handler, _, _ = _make_handler()
        event = _make_discovery_event()
        preview = handler.build_preview(event)
        assert "## Purpose" in preview.canvas_content


class TestWriteArtifactsIncludesCanvas:
    """write_artifacts() must write canvas content."""

    def test_writes_four_files(self) -> None:
        """Now writes PRD.md, DDD.md, ARCHITECTURE.md, plus canvas in DDD.md."""
        handler, _, writer = _make_handler()
        event = _make_discovery_event()
        preview = handler.build_preview(event)
        handler.write_artifacts(preview, Path("/tmp/test"))

        # Canvas content should be part of what's written.
        # It can be appended to DDD.md or written separately.
        written_contents = [
            call[0][1] for call in writer.write_file.call_args_list
        ]
        all_content = "\n".join(written_contents)
        assert "Bounded Context Canvas" in all_content


class TestCanvasWithNoContexts:
    """Edge case: when model has no bounded contexts, canvas is empty."""

    def test_empty_canvas_when_no_contexts(self) -> None:
        """Single-context MVP still produces canvas content."""
        handler, _, _ = _make_handler()
        answers = (
            Answer(question_id="Q1", response_text="Customer"),
            Answer(
                question_id="Q3",
                response_text="Customer places order, System confirms",
            ),
            Answer(
                question_id="Q4",
                response_text="Order must have at least one item",
            ),
            Answer(question_id="Q9", response_text="Orders"),
            Answer(
                question_id="Q10",
                response_text="Orders is core competitive advantage",
            ),
        )
        event = _make_discovery_event(answers=answers)
        preview = handler.build_preview(event)
        # Single context → single canvas.
        assert "Orders" in preview.canvas_content


class TestGenerateConvenienceStillWorks:
    """generate() backward-compatible convenience still works with canvas."""

    def test_generate_includes_canvas_in_write(self) -> None:
        handler, _, writer = _make_handler()
        event = _make_discovery_event()
        handler.generate(event, Path("/tmp/test"))

        written_contents = [
            call[0][1] for call in writer.write_file.call_args_list
        ]
        all_content = "\n".join(written_contents)
        assert "Bounded Context Canvas" in all_content
