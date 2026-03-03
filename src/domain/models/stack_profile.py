"""Stack profile protocol and concrete implementations.

A StackProfile provides stack-specific knowledge for the generation pipeline.
Each profile knows its language toolchain, quality gate commands, source layout,
and how to derive a root package name from a project name.

Supported profiles:
- PythonUvProfile: Python + uv (ruff, mypy, pytest, import-linter/pytestarch)
- GenericProfile: Fallback for unknown stacks, skips stack-specific stages
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from src.domain.models.quality_gate import QualityGate


@runtime_checkable
class StackProfile(Protocol):
    """Strategy interface providing stack-specific knowledge for the generation pipeline."""

    @property
    def stack_id(self) -> str: ...

    @property
    def file_glob(self) -> str: ...

    @property
    def project_manifest(self) -> str: ...

    @property
    def source_layout(self) -> tuple[str, ...]: ...

    @property
    def quality_gate_commands(self) -> dict[QualityGate, list[str]]: ...

    @property
    def quality_gate_display(self) -> str: ...

    @property
    def fitness_available(self) -> bool: ...

    def to_root_package(self, project_name: str) -> str: ...


class PythonUvProfile:
    """Python + uv stack profile with full quality gate pipeline."""

    @property
    def stack_id(self) -> str:
        return "python-uv"

    @property
    def file_glob(self) -> str:
        return "**/*.py"

    @property
    def project_manifest(self) -> str:
        return "pyproject.toml"

    @property
    def source_layout(self) -> tuple[str, ...]:
        return ("src/domain/", "src/application/", "src/infrastructure/")

    @property
    def quality_gate_commands(self) -> dict[QualityGate, list[str]]:
        from src.domain.models.quality_gate import QualityGate

        return {
            QualityGate.LINT: ["uv", "run", "ruff", "check", "."],
            QualityGate.TYPES: ["uv", "run", "mypy", "."],
            QualityGate.TESTS: ["uv", "run", "pytest"],
            QualityGate.FITNESS: ["uv", "run", "pytest", "tests/architecture/"],
        }

    @property
    def quality_gate_display(self) -> str:
        return (
            "## Quality Gates\n"
            "\n"
            "```bash\n"
            "uv run ruff check .              # Lint\n"
            "uv run mypy .                    # Type check\n"
            "uv run pytest                    # Tests\n"
            "```\n"
        )

    @property
    def fitness_available(self) -> bool:
        return True

    def to_root_package(self, project_name: str) -> str:
        return project_name.replace("-", "_")


class GenericProfile:
    """Fallback profile for unknown stacks. Skips stack-specific stages."""

    @property
    def stack_id(self) -> str:
        return "generic"

    @property
    def file_glob(self) -> str:
        return "*"

    @property
    def project_manifest(self) -> str:
        return ""

    @property
    def source_layout(self) -> tuple[str, ...]:
        return ()

    @property
    def quality_gate_commands(self) -> dict[QualityGate, list[str]]:
        return {}

    @property
    def quality_gate_display(self) -> str:
        return ""

    @property
    def fitness_available(self) -> bool:
        return False

    def to_root_package(self, project_name: str) -> str:
        return project_name
