"""Application command handler for fitness function generation.

FitnessGenerationHandler builds a FitnessTestSuite from a DomainModel's
bounded contexts, generates import-linter contracts and pytestarch rules,
and writes them via the FileWriterPort.

Supports a preview-before-write workflow: build_preview() renders content
without writing, write_files() commits a preview to disk.
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import TYPE_CHECKING

from src.domain.models.fitness_test_suite import FitnessTestSuite

if TYPE_CHECKING:
    from pathlib import Path

    from src.application.ports.file_writer_port import FileWriterPort
    from src.domain.models.domain_model import DomainModel
    from src.domain.models.stack_profile import StackProfile


@dataclass
class FitnessPreview:
    """Rendered fitness test content ready for user review before writing.

    Attributes:
        suite: The FitnessTestSuite aggregate.
        toml_content: Rendered import-linter TOML configuration.
        test_content: Rendered pytestarch test file.
        summary: Human-readable preview summary.
    """

    suite: FitnessTestSuite
    toml_content: str
    test_content: str
    summary: str


class FitnessGenerationHandler:
    """Orchestrates fitness function generation from a DomainModel.

    Reads bounded contexts from a finalized DomainModel, builds a
    FitnessTestSuite, and writes import-linter TOML + pytestarch tests
    via the FileWriterPort.
    """

    def __init__(self, writer: FileWriterPort) -> None:
        self._writer = writer

    def build_preview(
        self,
        model: DomainModel,
        root_package: str,
        profile: StackProfile | None = None,
    ) -> FitnessPreview | None:
        """Build fitness tests and render for preview without writing.

        Args:
            model: A DomainModel with classified bounded contexts.
            root_package: The root Python package name.
            profile: Stack profile. When fitness_available is False,
                returns None instead of generating tests.

        Returns:
            FitnessPreview with rendered content and suite, or None
            if the profile indicates fitness tests are not available.

        Raises:
            ValueError: If no bounded contexts in the model.
        """
        if profile is not None and not profile.fitness_available:
            return None

        contexts = model.bounded_contexts
        if not contexts:
            msg = "No bounded contexts to generate fitness tests for"
            raise ValueError(msg)

        suite = FitnessTestSuite(root_package=root_package)
        suite.generate_contracts(bounded_contexts=contexts)

        return FitnessPreview(
            suite=suite,
            toml_content=suite.render_import_linter_toml(),
            test_content=suite.render_pytestarch_tests(),
            summary=suite.preview(),
        )

    def write_files(self, preview: FitnessPreview, output_dir: Path) -> None:
        """Write previously previewed fitness tests to disk.

        Args:
            preview: The FitnessPreview from build_preview().
            output_dir: Project root directory.
        """
        self._writer.write_file(
            output_dir / "importlinter.toml",
            preview.toml_content,
        )
        self._writer.write_file(
            output_dir / "tests" / "architecture" / "test_fitness.py",
            preview.test_content,
        )

    def approve_and_write(
        self,
        preview: FitnessPreview,
        output_dir: Path,
    ) -> None:
        """Approve the suite (emitting domain event) and write to disk.

        This is the only way to finalize fitness tests — enforcing the
        preview-before-action pattern per ARCHITECTURE.md Design Principle 3.

        Args:
            preview: The FitnessPreview from build_preview().
            output_dir: Project root directory.
        """
        preview.suite.approve()
        self.write_files(preview, output_dir)
