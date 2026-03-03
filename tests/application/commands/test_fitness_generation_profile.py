"""Tests for FitnessGenerationHandler profile integration.

Verifies that the handler accepts a StackProfile, returns None when
fitness_available is False (GenericProfile), and uses
profile.to_root_package() for root_package derivation.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

import pytest

if TYPE_CHECKING:
    from pathlib import Path

from src.domain.models.domain_model import DomainModel
from src.domain.models.domain_values import (
    AggregateDesign,
    BoundedContext,
    DomainStory,
    SubdomainClassification,
)
from src.domain.models.stack_profile import GenericProfile, PythonUvProfile


def _make_model() -> DomainModel:
    """Build a minimal valid DomainModel with one Core context."""
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
            classification=SubdomainClassification.CORE,
        )
    )
    model.design_aggregate(
        AggregateDesign(
            name="OrdersRoot",
            context_name="Orders",
            root_entity="OrdersRoot",
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
# 1. Returns None when fitness unavailable
# ---------------------------------------------------------------------------


class TestFitnessUnavailable:
    def test_build_preview_returns_none_for_generic_profile(self) -> None:
        """GenericProfile (fitness_available=False) → build_preview returns None."""
        from src.application.commands.fitness_generation_handler import (
            FitnessGenerationHandler,
        )

        model = _make_model()
        profile = GenericProfile()
        writer = FakeFileWriter()
        handler = FitnessGenerationHandler(writer=writer)

        result = handler.build_preview(model, root_package="myapp", profile=profile)

        assert result is None

    def test_build_preview_does_not_write_for_generic_profile(self) -> None:
        """GenericProfile → no files written."""
        from src.application.commands.fitness_generation_handler import (
            FitnessGenerationHandler,
        )

        model = _make_model()
        profile = GenericProfile()
        writer = FakeFileWriter()
        handler = FitnessGenerationHandler(writer=writer)

        handler.build_preview(model, root_package="myapp", profile=profile)

        assert writer.written == {}


# ---------------------------------------------------------------------------
# 2. Returns FitnessPreview when fitness available
# ---------------------------------------------------------------------------


class TestFitnessAvailable:
    def test_build_preview_returns_preview_for_python_profile(self) -> None:
        """PythonUvProfile (fitness_available=True) → returns FitnessPreview."""
        from src.application.commands.fitness_generation_handler import (
            FitnessGenerationHandler,
            FitnessPreview,
        )

        model = _make_model()
        profile = PythonUvProfile()
        writer = FakeFileWriter()
        handler = FitnessGenerationHandler(writer=writer)

        result = handler.build_preview(model, root_package="myapp", profile=profile)

        assert isinstance(result, FitnessPreview)
        assert result.toml_content
        assert result.test_content

    def test_python_output_identical_to_no_profile(self) -> None:
        """PythonUvProfile output matches backward-compat (no profile) output."""
        from src.application.commands.fitness_generation_handler import (
            FitnessGenerationHandler,
        )

        model = _make_model()
        profile = PythonUvProfile()
        writer = FakeFileWriter()
        handler = FitnessGenerationHandler(writer=writer)

        with_profile = handler.build_preview(model, root_package="myapp", profile=profile)
        without_profile = handler.build_preview(model, root_package="myapp")

        assert with_profile is not None
        assert without_profile is not None
        assert with_profile.toml_content == without_profile.toml_content
        assert with_profile.test_content == without_profile.test_content


# ---------------------------------------------------------------------------
# 3. Backward compatibility
# ---------------------------------------------------------------------------


class TestBackwardCompat:
    def test_no_profile_still_works(self) -> None:
        """Calling build_preview without profile still returns FitnessPreview."""
        from src.application.commands.fitness_generation_handler import (
            FitnessGenerationHandler,
            FitnessPreview,
        )

        model = _make_model()
        writer = FakeFileWriter()
        handler = FitnessGenerationHandler(writer=writer)

        result = handler.build_preview(model, root_package="myapp")

        assert isinstance(result, FitnessPreview)

    def test_empty_model_still_raises(self) -> None:
        """ValueError for empty model regardless of profile."""
        from src.application.commands.fitness_generation_handler import (
            FitnessGenerationHandler,
        )

        model = DomainModel()
        profile = PythonUvProfile()
        writer = FakeFileWriter()
        handler = FitnessGenerationHandler(writer=writer)

        with pytest.raises(ValueError, match="No bounded contexts"):
            handler.build_preview(model, root_package="myapp", profile=profile)
