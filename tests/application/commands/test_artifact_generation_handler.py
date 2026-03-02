"""Tests for ArtifactGenerationHandler."""

from __future__ import annotations

from pathlib import Path
from unittest.mock import MagicMock

import pytest

from src.application.commands.artifact_generation_handler import (
    ArtifactGenerationHandler,
)
from src.domain.events.discovery_events import DiscoveryCompleted
from src.domain.models.discovery_values import Answer, Persona, Playback, Register
from src.domain.models.domain_values import SubdomainClassification


def _make_discovery_event(
    answers: tuple[Answer, ...] | None = None,
) -> DiscoveryCompleted:
    """Create a DiscoveryCompleted event with standard or custom answers."""
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
                    "Sales is core competitive advantage, Inventory is supporting necessary"
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


class TestBuildFromDiscoveryCompleted:
    def test_generates_domain_model(self) -> None:
        renderer = MagicMock()
        renderer.render_prd.return_value = "# PRD"
        renderer.render_ddd.return_value = "# DDD"
        renderer.render_architecture.return_value = "# ARCH"
        writer = MagicMock()

        handler = ArtifactGenerationHandler(renderer=renderer, writer=writer)
        event = _make_discovery_event()
        model = handler.generate(event, Path("/tmp/test"))

        assert len(model.events) == 1  # DomainModelGenerated emitted.
        assert len(model.domain_stories) >= 1
        assert len(model.bounded_contexts) == 2

    def test_writes_three_files(self) -> None:
        renderer = MagicMock()
        renderer.render_prd.return_value = "# PRD"
        renderer.render_ddd.return_value = "# DDD"
        renderer.render_architecture.return_value = "# ARCH"
        writer = MagicMock()

        handler = ArtifactGenerationHandler(renderer=renderer, writer=writer)
        event = _make_discovery_event()
        handler.generate(event, Path("/tmp/test"))

        assert writer.write_file.call_count == 3
        written_paths = [call[0][0] for call in writer.write_file.call_args_list]
        assert Path("/tmp/test/PRD.md") in written_paths
        assert Path("/tmp/test/DDD.md") in written_paths
        assert Path("/tmp/test/ARCHITECTURE.md") in written_paths

    def test_renderer_called_with_model(self) -> None:
        renderer = MagicMock()
        renderer.render_prd.return_value = ""
        renderer.render_ddd.return_value = ""
        renderer.render_architecture.return_value = ""
        writer = MagicMock()

        handler = ArtifactGenerationHandler(renderer=renderer, writer=writer)
        event = _make_discovery_event()
        model = handler.generate(event, Path("/tmp/test"))

        renderer.render_prd.assert_called_once_with(model)
        renderer.render_ddd.assert_called_once_with(model)
        renderer.render_architecture.assert_called_once_with(model)


class TestEmptyAnswers:
    def test_no_answers_raises(self) -> None:
        renderer = MagicMock()
        writer = MagicMock()
        handler = ArtifactGenerationHandler(renderer=renderer, writer=writer)
        event = _make_discovery_event(answers=())
        with pytest.raises(ValueError, match="No substantive answers"):
            handler.generate(event, Path("/tmp/test"))


class TestQuickMode:
    def test_mvp_questions_only(self) -> None:
        """Test with only MVP questions (Q1, Q3, Q4, Q9, Q10)."""
        renderer = MagicMock()
        renderer.render_prd.return_value = ""
        renderer.render_ddd.return_value = ""
        renderer.render_architecture.return_value = ""
        writer = MagicMock()

        handler = ArtifactGenerationHandler(renderer=renderer, writer=writer)
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
        model = handler.generate(event, Path("/tmp/test"))

        assert len(model.events) == 1
        assert len(model.bounded_contexts) == 1


class TestSplitAnswer:
    def test_comma_separated(self) -> None:
        result = ArtifactGenerationHandler._split_answer("Order, Product, Customer")
        assert result == ["Order", "Product", "Customer"]

    def test_newline_separated(self) -> None:
        result = ArtifactGenerationHandler._split_answer("1. Order\n2. Product")
        assert result == ["Order", "Product"]

    def test_single_item(self) -> None:
        result = ArtifactGenerationHandler._split_answer("Order")
        assert result == ["Order"]

    def test_empty_string(self) -> None:
        result = ArtifactGenerationHandler._split_answer("")
        assert result == []

    def test_whitespace_only(self) -> None:
        result = ArtifactGenerationHandler._split_answer("   ")
        assert result == []


# =========================================================================
# New tests from code review fixes
# =========================================================================


class TestBuildPreview:
    """H2: build_preview returns rendered content WITHOUT writing files."""

    def test_returns_preview_without_writing(self) -> None:
        renderer = MagicMock()
        renderer.render_prd.return_value = "# PRD"
        renderer.render_ddd.return_value = "# DDD"
        renderer.render_architecture.return_value = "# ARCH"
        writer = MagicMock()

        handler = ArtifactGenerationHandler(renderer=renderer, writer=writer)
        event = _make_discovery_event()
        preview = handler.build_preview(event)

        assert preview.prd_content == "# PRD"
        assert preview.ddd_content == "# DDD"
        assert preview.architecture_content == "# ARCH"
        writer.write_file.assert_not_called()

    def test_preview_model_is_finalized(self) -> None:
        renderer = MagicMock()
        renderer.render_prd.return_value = ""
        renderer.render_ddd.return_value = ""
        renderer.render_architecture.return_value = ""
        writer = MagicMock()

        handler = ArtifactGenerationHandler(renderer=renderer, writer=writer)
        event = _make_discovery_event()
        preview = handler.build_preview(event)

        assert len(preview.model.events) >= 1

    def test_preview_empty_answers_raises(self) -> None:
        renderer = MagicMock()
        writer = MagicMock()
        handler = ArtifactGenerationHandler(renderer=renderer, writer=writer)
        event = _make_discovery_event(answers=())
        with pytest.raises(ValueError, match="No substantive answers"):
            handler.build_preview(event)

    def test_preview_renderer_called_with_model(self) -> None:
        renderer = MagicMock()
        renderer.render_prd.return_value = ""
        renderer.render_ddd.return_value = ""
        renderer.render_architecture.return_value = ""
        writer = MagicMock()

        handler = ArtifactGenerationHandler(renderer=renderer, writer=writer)
        event = _make_discovery_event()
        preview = handler.build_preview(event)

        renderer.render_prd.assert_called_once_with(preview.model)
        renderer.render_ddd.assert_called_once_with(preview.model)
        renderer.render_architecture.assert_called_once_with(preview.model)


class TestWriteArtifacts:
    """H2: write_artifacts writes a previously built preview to disk."""

    def test_writes_three_files_from_preview(self) -> None:
        renderer = MagicMock()
        renderer.render_prd.return_value = "# PRD"
        renderer.render_ddd.return_value = "# DDD"
        renderer.render_architecture.return_value = "# ARCH"
        writer = MagicMock()

        handler = ArtifactGenerationHandler(renderer=renderer, writer=writer)
        event = _make_discovery_event()
        preview = handler.build_preview(event)
        handler.write_artifacts(preview, Path("/tmp/test"))

        assert writer.write_file.call_count == 3
        written = [c[0][0] for c in writer.write_file.call_args_list]
        assert Path("/tmp/test/PRD.md") in written
        assert Path("/tmp/test/DDD.md") in written
        assert Path("/tmp/test/ARCHITECTURE.md") in written

    def test_writes_preview_content_not_re_rendered(self) -> None:
        renderer = MagicMock()
        renderer.render_prd.return_value = "PRD body"
        renderer.render_ddd.return_value = "DDD body"
        renderer.render_architecture.return_value = "ARCH body"
        writer = MagicMock()

        handler = ArtifactGenerationHandler(renderer=renderer, writer=writer)
        event = _make_discovery_event()
        preview = handler.build_preview(event)
        handler.write_artifacts(preview, Path("/tmp/out"))

        contents = {c[0][0].name: c[0][1] for c in writer.write_file.call_args_list}
        assert contents["PRD.md"] == "PRD body"
        assert contents["DDD.md"] == "DDD body"
        assert contents["ARCHITECTURE.md"] == "ARCH body"

    def test_write_does_not_re_render(self) -> None:
        renderer = MagicMock()
        renderer.render_prd.return_value = ""
        renderer.render_ddd.return_value = ""
        renderer.render_architecture.return_value = ""
        writer = MagicMock()

        handler = ArtifactGenerationHandler(renderer=renderer, writer=writer)
        event = _make_discovery_event()
        preview = handler.build_preview(event)
        renderer.reset_mock()

        handler.write_artifacts(preview, Path("/tmp/out"))
        renderer.render_prd.assert_not_called()
        renderer.render_ddd.assert_not_called()
        renderer.render_architecture.assert_not_called()


class TestGenerateConvenience:
    """generate() backward-compatible convenience (preview + write)."""

    def test_generate_still_works(self) -> None:
        renderer = MagicMock()
        renderer.render_prd.return_value = "# PRD"
        renderer.render_ddd.return_value = "# DDD"
        renderer.render_architecture.return_value = "# ARCH"
        writer = MagicMock()

        handler = ArtifactGenerationHandler(renderer=renderer, writer=writer)
        event = _make_discovery_event()
        model = handler.generate(event, Path("/tmp/test"))

        assert len(model.events) >= 1
        assert writer.write_file.call_count == 3


class TestNoDefaultContext:
    """M5: Terms assigned to real context names, never 'Default'."""

    def test_terms_use_real_context_name(self) -> None:
        renderer = MagicMock()
        renderer.render_prd.return_value = ""
        renderer.render_ddd.return_value = ""
        renderer.render_architecture.return_value = ""
        writer = MagicMock()

        handler = ArtifactGenerationHandler(renderer=renderer, writer=writer)
        event = _make_discovery_event()
        preview = handler.build_preview(event)

        for term in preview.model.ubiquitous_language.terms:
            assert term.context_name != "Default", (
                f"Term '{term.term}' assigned to 'Default' context"
            )


class TestNoArtificialRelationships:
    """M4: Handler does not synthesize fake relationships."""

    def test_no_manufactured_relationships(self) -> None:
        renderer = MagicMock()
        renderer.render_prd.return_value = ""
        renderer.render_ddd.return_value = ""
        renderer.render_architecture.return_value = ""
        writer = MagicMock()

        handler = ArtifactGenerationHandler(renderer=renderer, writer=writer)
        event = _make_discovery_event()
        preview = handler.build_preview(event)

        assert preview.model.context_relationships == ()


class TestDeadCodeRemoved:
    """M1: _QUESTION_ARTIFACT_MAP constant is removed."""

    def test_no_question_artifact_map(self) -> None:
        from src.application.commands import artifact_generation_handler as mod

        assert not hasattr(mod, "_QUESTION_ARTIFACT_MAP")


# =========================================================================
# Fix 1: Silent SUPPORTING default (alty-5o4)
# =========================================================================


class TestNoSilentSupportingDefault:
    """_extract_classifications must NOT silently default to SUPPORTING.

    When Q10 is empty or keywords don't resolve for a context, the context
    should remain unclassified (None) so finalize() invariant 2 catches it.
    """

    def test_empty_q10_leaves_contexts_unclassified(self) -> None:
        """Empty Q10 → contexts stay None → finalize raises invariant 2."""
        from src.domain.models.errors import InvariantViolationError

        renderer = MagicMock()
        renderer.render_prd.return_value = ""
        renderer.render_ddd.return_value = ""
        renderer.render_architecture.return_value = ""
        writer = MagicMock()

        handler = ArtifactGenerationHandler(renderer=renderer, writer=writer)
        answers = (
            Answer(question_id="Q1", response_text="Customer"),
            Answer(
                question_id="Q3",
                response_text="Customer places order",
            ),
            Answer(
                question_id="Q4",
                response_text="Order must have at least one item",
            ),
            Answer(question_id="Q9", response_text="Sales, Inventory"),
            Answer(question_id="Q10", response_text=""),
        )
        event = _make_discovery_event(answers=answers)
        with pytest.raises(InvariantViolationError, match="has no classification"):
            handler.build_preview(event)

    def test_missing_q10_leaves_contexts_unclassified(self) -> None:
        """No Q10 answer at all → same as empty Q10."""
        from src.domain.models.errors import InvariantViolationError

        renderer = MagicMock()
        renderer.render_prd.return_value = ""
        renderer.render_ddd.return_value = ""
        renderer.render_architecture.return_value = ""
        writer = MagicMock()

        handler = ArtifactGenerationHandler(renderer=renderer, writer=writer)
        answers = (
            Answer(question_id="Q1", response_text="Customer"),
            Answer(
                question_id="Q3",
                response_text="Customer places order",
            ),
            Answer(
                question_id="Q4",
                response_text="Order must have at least one item",
            ),
            Answer(question_id="Q9", response_text="Sales"),
        )
        event = _make_discovery_event(answers=answers)
        with pytest.raises(InvariantViolationError, match="has no classification"):
            handler.build_preview(event)

    def test_unresolvable_keywords_leaves_context_unclassified(self) -> None:
        """Q10 has text but no matching keywords for a context → stays None."""
        from src.domain.models.errors import InvariantViolationError

        renderer = MagicMock()
        renderer.render_prd.return_value = ""
        renderer.render_ddd.return_value = ""
        renderer.render_architecture.return_value = ""
        writer = MagicMock()

        handler = ArtifactGenerationHandler(renderer=renderer, writer=writer)
        answers = (
            Answer(question_id="Q1", response_text="Customer"),
            Answer(
                question_id="Q3",
                response_text="Customer places order",
            ),
            Answer(
                question_id="Q4",
                response_text="Order must have at least one item",
            ),
            Answer(question_id="Q9", response_text="Sales, Billing"),
            Answer(
                question_id="Q10",
                response_text="Sales is core competitive advantage",
            ),
        )
        event = _make_discovery_event(answers=answers)
        # Sales resolves to CORE, but Billing has no matching keywords → stays None
        with pytest.raises(InvariantViolationError, match="has no classification"):
            handler.build_preview(event)

    def test_resolved_context_still_gets_classified(self) -> None:
        """Contexts that DO match keywords should still be classified."""
        renderer = MagicMock()
        renderer.render_prd.return_value = ""
        renderer.render_ddd.return_value = ""
        renderer.render_architecture.return_value = ""
        writer = MagicMock()

        handler = ArtifactGenerationHandler(renderer=renderer, writer=writer)
        answers = (
            Answer(question_id="Q1", response_text="Customer"),
            Answer(
                question_id="Q3",
                response_text="Customer places order",
            ),
            Answer(
                question_id="Q4",
                response_text="Order must have at least one item",
            ),
            Answer(question_id="Q9", response_text="Sales"),
            Answer(
                question_id="Q10",
                response_text="Sales is core competitive advantage",
            ),
        )
        event = _make_discovery_event(answers=answers)
        preview = handler.build_preview(event)
        ctx = preview.model.bounded_contexts[0]
        assert ctx.classification == SubdomainClassification.CORE
