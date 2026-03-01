"""Tests for FitnessGenerationHandler.

Covers the application-layer orchestration: building a FitnessTestSuite
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


def _make_model_with_contexts(
    contexts: list[tuple[str, SubdomainClassification]],
) -> DomainModel:
    """Build a minimal valid DomainModel with the given classified contexts."""
    model = DomainModel()

    # Add stories that mention the context names (word boundary invariant)
    all_names = [name for name, _ in contexts]
    model.add_domain_story(
        DomainStory(
            name="Test flow",
            actors=("User",),
            trigger="User starts",
            steps=tuple(f"User manages {name}" for name in all_names),
        )
    )

    # Add terms and contexts
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

    # Core contexts need aggregate designs
    for name, classification in contexts:
        if classification == SubdomainClassification.CORE:
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
# 1. Build preview from DomainModel
# ---------------------------------------------------------------------------


class TestBuildPreview:
    def test_build_preview_returns_preview(self) -> None:
        from src.application.commands.fitness_generation_handler import (
            FitnessGenerationHandler,
            FitnessPreview,
        )

        model = _make_model_with_contexts(
            [("Orders", SubdomainClassification.CORE)]
        )
        writer = FakeFileWriter()
        handler = FitnessGenerationHandler(writer=writer)

        preview = handler.build_preview(model, root_package="myapp")

        assert isinstance(preview, FitnessPreview)
        assert preview.toml_content  # non-empty
        assert preview.test_content  # non-empty
        assert preview.summary  # non-empty

    def test_build_preview_does_not_write(self) -> None:
        from src.application.commands.fitness_generation_handler import (
            FitnessGenerationHandler,
        )

        model = _make_model_with_contexts(
            [("Orders", SubdomainClassification.CORE)]
        )
        writer = FakeFileWriter()
        handler = FitnessGenerationHandler(writer=writer)
        handler.build_preview(model, root_package="myapp")

        assert writer.written == {}

    def test_preview_contains_all_bc_names(self) -> None:
        from src.application.commands.fitness_generation_handler import (
            FitnessGenerationHandler,
        )

        model = _make_model_with_contexts(
            [
                ("Orders", SubdomainClassification.CORE),
                ("Notifications", SubdomainClassification.SUPPORTING),
            ]
        )
        writer = FakeFileWriter()
        handler = FitnessGenerationHandler(writer=writer)
        preview = handler.build_preview(model, root_package="myapp")

        assert "Orders" in preview.summary
        assert "Notifications" in preview.summary

    def test_preview_model_with_no_contexts_raises(self) -> None:
        from src.application.commands.fitness_generation_handler import (
            FitnessGenerationHandler,
        )

        model = DomainModel()  # empty, not finalized
        writer = FakeFileWriter()
        handler = FitnessGenerationHandler(writer=writer)

        with pytest.raises(ValueError, match="No bounded contexts"):
            handler.build_preview(model, root_package="myapp")


# ---------------------------------------------------------------------------
# 2. Write files from preview
# ---------------------------------------------------------------------------


class TestWriteFiles:
    def test_write_creates_toml_and_test_files(self) -> None:
        from src.application.commands.fitness_generation_handler import (
            FitnessGenerationHandler,
        )

        model = _make_model_with_contexts(
            [("Orders", SubdomainClassification.CORE)]
        )
        writer = FakeFileWriter()
        handler = FitnessGenerationHandler(writer=writer)
        preview = handler.build_preview(model, root_package="myapp")

        handler.write_files(preview, output_dir=Path("/project"))

        # Should write at least 2 files: TOML config and test file
        assert len(writer.written) >= 2
        paths = list(writer.written.keys())
        assert any("pyproject" in p or "importlinter" in p for p in paths)
        assert any("test_" in p for p in paths)

    def test_write_content_matches_preview(self) -> None:
        from src.application.commands.fitness_generation_handler import (
            FitnessGenerationHandler,
        )

        model = _make_model_with_contexts(
            [("Orders", SubdomainClassification.CORE)]
        )
        writer = FakeFileWriter()
        handler = FitnessGenerationHandler(writer=writer)
        preview = handler.build_preview(model, root_package="myapp")
        handler.write_files(preview, output_dir=Path("/project"))

        # Written TOML content should match preview
        toml_paths = [p for p in writer.written if "importlinter" in p or "pyproject" in p]
        assert len(toml_paths) >= 1
        assert writer.written[toml_paths[0]] == preview.toml_content


# ---------------------------------------------------------------------------
# 3. Approve and write (C-2: handler must call approve, emit domain event)
# ---------------------------------------------------------------------------


class TestApproveAndWrite:
    def test_approve_and_write_calls_approve(self) -> None:
        """C-2: Handler must call suite.approve() so FitnessTestsGenerated fires."""
        from src.application.commands.fitness_generation_handler import (
            FitnessGenerationHandler,
        )

        model = _make_model_with_contexts(
            [("Orders", SubdomainClassification.CORE)]
        )
        writer = FakeFileWriter()
        handler = FitnessGenerationHandler(writer=writer)
        preview = handler.build_preview(model, root_package="myapp")

        handler.approve_and_write(preview, output_dir=Path("/project"))

        assert preview.suite.events  # event was emitted
        assert len(writer.written) >= 2

    def test_approve_and_write_emits_fitness_tests_generated(self) -> None:
        from src.application.commands.fitness_generation_handler import (
            FitnessGenerationHandler,
        )
        from src.domain.events.fitness_events import FitnessTestsGenerated

        model = _make_model_with_contexts(
            [("Orders", SubdomainClassification.CORE)]
        )
        writer = FakeFileWriter()
        handler = FitnessGenerationHandler(writer=writer)
        preview = handler.build_preview(model, root_package="myapp")

        handler.approve_and_write(preview, output_dir=Path("/project"))

        assert isinstance(preview.suite.events[0], FitnessTestsGenerated)

    def test_no_generate_convenience_method(self) -> None:
        """I-1: generate() convenience removed — enforce preview-before-action."""
        from src.application.commands.fitness_generation_handler import (
            FitnessGenerationHandler,
        )

        assert not hasattr(FitnessGenerationHandler, "generate")


# ---------------------------------------------------------------------------
# 4. Strictness mapping from DomainModel contexts
# ---------------------------------------------------------------------------


class TestStrictnessMapping:
    def test_core_gets_strict(self) -> None:
        from src.application.commands.fitness_generation_handler import (
            FitnessGenerationHandler,
        )

        model = _make_model_with_contexts(
            [("Orders", SubdomainClassification.CORE)]
        )
        writer = FakeFileWriter()
        handler = FitnessGenerationHandler(writer=writer)
        preview = handler.build_preview(model, root_package="myapp")

        assert "STRICT" in preview.summary or "strict" in preview.summary.lower()

    def test_generic_gets_minimal(self) -> None:
        from src.application.commands.fitness_generation_handler import (
            FitnessGenerationHandler,
        )

        model = _make_model_with_contexts(
            [("Logging", SubdomainClassification.GENERIC)]
        )
        writer = FakeFileWriter()
        handler = FitnessGenerationHandler(writer=writer)
        preview = handler.build_preview(model, root_package="myapp")

        # Minimal = only forbidden contracts
        assert "[tool.importlinter]" in preview.toml_content
        assert 'type = "forbidden"' in preview.toml_content
        # Should NOT have layers for generic
        assert 'type = "layers"' not in preview.toml_content


# ---------------------------------------------------------------------------
# 5. FitnessGenerationPort updated (I-4, I-9)
# ---------------------------------------------------------------------------


class TestPortSignature:
    def test_port_has_generate_method_accepting_domain_model(self) -> None:
        """I-4: Port signature should accept DomainModel, not str."""
        import inspect

        from src.application.ports.fitness_generation_port import FitnessGenerationPort

        sig = inspect.signature(FitnessGenerationPort.generate)
        params = list(sig.parameters.keys())
        assert "model" in params
        assert "root_package" in params

    def test_port_generate_accepts_output_dir(self) -> None:
        """I-9: Port must include output_dir parameter."""
        import inspect

        from src.application.ports.fitness_generation_port import FitnessGenerationPort

        sig = inspect.signature(FitnessGenerationPort.generate)
        params = list(sig.parameters.keys())
        assert "output_dir" in params


# ---------------------------------------------------------------------------
# 6. Edge cases
# ---------------------------------------------------------------------------


class TestEdgeCases:
    def test_approve_and_write_twice_raises(self) -> None:
        """Cannot approve the same preview twice."""
        from src.application.commands.fitness_generation_handler import (
            FitnessGenerationHandler,
        )
        from src.domain.models.errors import InvariantViolationError

        model = _make_model_with_contexts(
            [("Orders", SubdomainClassification.CORE)]
        )
        writer = FakeFileWriter()
        handler = FitnessGenerationHandler(writer=writer)
        preview = handler.build_preview(model, root_package="myapp")

        handler.approve_and_write(preview, output_dir=Path("/project"))

        with pytest.raises(InvariantViolationError, match="already approved"):
            handler.approve_and_write(preview, output_dir=Path("/project"))

    def test_all_three_classifications_in_one_model(self) -> None:
        """Full model with Core + Supporting + Generic contexts."""
        from src.application.commands.fitness_generation_handler import (
            FitnessGenerationHandler,
        )

        model = _make_model_with_contexts(
            [
                ("Orders", SubdomainClassification.CORE),
                ("Notifications", SubdomainClassification.SUPPORTING),
                ("Logging", SubdomainClassification.GENERIC),
            ]
        )
        writer = FakeFileWriter()
        handler = FitnessGenerationHandler(writer=writer)
        preview = handler.build_preview(model, root_package="myapp")

        # All 3 contexts should be in the summary
        assert "Orders" in preview.summary
        assert "Notifications" in preview.summary
        assert "Logging" in preview.summary

        # TOML should have contracts for all
        assert "orders" in preview.toml_content
        assert "notifications" in preview.toml_content
        assert "logging" in preview.toml_content

    def test_supporting_gets_moderate(self) -> None:
        from src.application.commands.fitness_generation_handler import (
            FitnessGenerationHandler,
        )

        model = _make_model_with_contexts(
            [("Notifications", SubdomainClassification.SUPPORTING)]
        )
        writer = FakeFileWriter()
        handler = FitnessGenerationHandler(writer=writer)
        preview = handler.build_preview(model, root_package="myapp")

        assert 'type = "layers"' in preview.toml_content
        assert 'type = "forbidden"' in preview.toml_content
        # Should NOT have independence or acyclic_siblings
        assert 'type = "independence"' not in preview.toml_content
        assert 'type = "acyclic_siblings"' not in preview.toml_content

    def test_write_files_uses_correct_paths(self) -> None:
        from src.application.commands.fitness_generation_handler import (
            FitnessGenerationHandler,
        )

        model = _make_model_with_contexts(
            [("Orders", SubdomainClassification.CORE)]
        )
        writer = FakeFileWriter()
        handler = FitnessGenerationHandler(writer=writer)
        preview = handler.build_preview(model, root_package="myapp")
        handler.write_files(preview, output_dir=Path("/project"))

        paths = list(writer.written.keys())
        assert "/project/importlinter.toml" in paths
        assert (
            "/project/tests/architecture/test_fitness.py" in paths
        )
