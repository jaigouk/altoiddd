"""Tests for StackProfile protocol, PythonUvProfile, and GenericProfile.

Verifies:
- Both concrete classes satisfy the StackProfile runtime_checkable Protocol
- PythonUvProfile returns correct Python+uv pipeline values
- GenericProfile returns safe empty/passthrough defaults
- Edge cases for to_root_package conversion
"""

from __future__ import annotations

from src.domain.models.quality_gate import QualityGate
from src.domain.models.stack_profile import (
    GenericProfile,
    PythonUvProfile,
    StackProfile,
)

# ---------------------------------------------------------------------------
# Protocol compliance
# ---------------------------------------------------------------------------


class TestProtocolCompliance:
    """Both concrete classes must satisfy isinstance(x, StackProfile)."""

    def test_python_uv_profile_satisfies_protocol(self) -> None:
        profile = PythonUvProfile()
        assert isinstance(profile, StackProfile)

    def test_generic_profile_satisfies_protocol(self) -> None:
        profile = GenericProfile()
        assert isinstance(profile, StackProfile)


# ---------------------------------------------------------------------------
# PythonUvProfile
# ---------------------------------------------------------------------------


class TestPythonUvProfile:
    """PythonUvProfile provides full Python+uv pipeline values."""

    def test_stack_id(self) -> None:
        assert PythonUvProfile().stack_id == "python-uv"

    def test_file_glob(self) -> None:
        assert PythonUvProfile().file_glob == "**/*.py"

    def test_project_manifest(self) -> None:
        assert PythonUvProfile().project_manifest == "pyproject.toml"

    def test_source_layout(self) -> None:
        assert PythonUvProfile().source_layout == (
            "src/domain/",
            "src/application/",
            "src/infrastructure/",
        )

    def test_quality_gate_commands_lint(self) -> None:
        cmds = PythonUvProfile().quality_gate_commands
        assert cmds[QualityGate.LINT] == ["uv", "run", "ruff", "check", "."]

    def test_quality_gate_commands_types(self) -> None:
        cmds = PythonUvProfile().quality_gate_commands
        assert cmds[QualityGate.TYPES] == ["uv", "run", "mypy", "."]

    def test_quality_gate_commands_tests(self) -> None:
        cmds = PythonUvProfile().quality_gate_commands
        assert cmds[QualityGate.TESTS] == ["uv", "run", "pytest"]

    def test_quality_gate_commands_fitness(self) -> None:
        cmds = PythonUvProfile().quality_gate_commands
        assert cmds[QualityGate.FITNESS] == [
            "uv", "run", "pytest", "tests/architecture/",
        ]

    def test_quality_gate_commands_covers_all_gates(self) -> None:
        """Every QualityGate enum member has a corresponding command."""
        cmds = PythonUvProfile().quality_gate_commands
        assert set(cmds.keys()) == set(QualityGate)

    def test_quality_gate_display(self) -> None:
        display = PythonUvProfile().quality_gate_display
        assert "uv run ruff check ." in display
        assert "uv run mypy ." in display
        assert "uv run pytest" in display
        assert display.startswith("## Quality Gates")

    def test_fitness_available(self) -> None:
        assert PythonUvProfile().fitness_available is True

    def test_to_root_package_hyphenated(self) -> None:
        assert PythonUvProfile().to_root_package("my-app") == "my_app"

    def test_to_root_package_multi_hyphen(self) -> None:
        assert PythonUvProfile().to_root_package("my-cool-app") == "my_cool_app"

    def test_to_root_package_already_underscored(self) -> None:
        assert PythonUvProfile().to_root_package("my_app") == "my_app"

    def test_to_root_package_empty(self) -> None:
        assert PythonUvProfile().to_root_package("") == ""


# ---------------------------------------------------------------------------
# GenericProfile
# ---------------------------------------------------------------------------


class TestGenericProfile:
    """GenericProfile provides safe defaults for unknown stacks."""

    def test_stack_id(self) -> None:
        assert GenericProfile().stack_id == "generic"

    def test_file_glob(self) -> None:
        assert GenericProfile().file_glob == "*"

    def test_project_manifest(self) -> None:
        assert GenericProfile().project_manifest == ""

    def test_source_layout(self) -> None:
        assert GenericProfile().source_layout == ()

    def test_quality_gate_commands_empty(self) -> None:
        assert GenericProfile().quality_gate_commands == {}

    def test_quality_gate_display_empty(self) -> None:
        assert GenericProfile().quality_gate_display == ""

    def test_fitness_not_available(self) -> None:
        assert GenericProfile().fitness_available is False

    def test_to_root_package_passthrough(self) -> None:
        assert GenericProfile().to_root_package("my-app") == "my-app"

    def test_to_root_package_passthrough_multi_hyphen(self) -> None:
        assert GenericProfile().to_root_package("my-cool-app") == "my-cool-app"

    def test_to_root_package_passthrough_underscored(self) -> None:
        assert GenericProfile().to_root_package("my_app") == "my_app"

    def test_to_root_package_passthrough_empty(self) -> None:
        assert GenericProfile().to_root_package("") == ""
