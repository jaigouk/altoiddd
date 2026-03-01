"""Tests for PersonaHandler command handler.

Covers listing personas, building previews, validation, target path
resolution, and the approve-and-write workflow with FakeFileWriter.
"""

from __future__ import annotations

from pathlib import Path

import pytest

from src.domain.models.errors import InvariantViolationError

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


class FakeFileWriter:
    """In-memory file writer for testing."""

    def __init__(self) -> None:
        self.written: dict[str, str] = {}

    def write_file(self, path: Path, content: str) -> None:
        self.written[str(path)] = content


# ---------------------------------------------------------------------------
# 1. list_personas
# ---------------------------------------------------------------------------


class TestListPersonas:
    def test_list_personas_returns_five(self) -> None:
        from src.application.commands.persona_handler import PersonaHandler

        writer = FakeFileWriter()
        handler = PersonaHandler(writer=writer)

        result = handler.list_personas()

        assert len(result) == 5

    def test_list_personas_correct_registers(self) -> None:
        from src.application.commands.persona_handler import PersonaHandler
        from src.domain.models.persona import Register

        writer = FakeFileWriter()
        handler = PersonaHandler(writer=writer)

        result = handler.list_personas()

        technical = [p for p in result if p.register == Register.TECHNICAL]
        non_technical = [p for p in result if p.register == Register.NON_TECHNICAL]

        assert len(technical) == 3
        assert len(non_technical) == 2

    def test_list_personas_returns_tuple(self) -> None:
        from src.application.commands.persona_handler import PersonaHandler

        writer = FakeFileWriter()
        handler = PersonaHandler(writer=writer)

        result = handler.list_personas()

        assert isinstance(result, tuple)


# ---------------------------------------------------------------------------
# 2. build_preview — valid inputs
# ---------------------------------------------------------------------------


class TestBuildPreviewValid:
    def test_build_preview_valid_persona(self) -> None:
        from src.application.commands.persona_handler import PersonaHandler, PersonaPreview

        writer = FakeFileWriter()
        handler = PersonaHandler(writer=writer)

        preview = handler.build_preview(persona_name="Solo Developer", tool="claude-code")

        assert isinstance(preview, PersonaPreview)
        assert preview.content  # non-empty
        assert preview.target_path  # non-empty
        assert preview.summary  # non-empty

    def test_build_preview_case_insensitive_name(self) -> None:
        from src.application.commands.persona_handler import PersonaHandler

        writer = FakeFileWriter()
        handler = PersonaHandler(writer=writer)

        preview = handler.build_preview(persona_name="solo developer", tool="claude-code")

        assert preview.persona.name == "Solo Developer"

    def test_build_preview_by_persona_type_value(self) -> None:
        from src.application.commands.persona_handler import PersonaHandler

        writer = FakeFileWriter()
        handler = PersonaHandler(writer=writer)

        preview = handler.build_preview(persona_name="solo_developer", tool="claude-code")

        assert preview.persona.name == "Solo Developer"

    def test_build_preview_does_not_write(self) -> None:
        from src.application.commands.persona_handler import PersonaHandler

        writer = FakeFileWriter()
        handler = PersonaHandler(writer=writer)

        handler.build_preview(persona_name="Solo Developer", tool="claude-code")

        assert writer.written == {}


# ---------------------------------------------------------------------------
# 3. build_preview — invalid inputs
# ---------------------------------------------------------------------------


class TestBuildPreviewInvalid:
    def test_build_preview_unknown_persona_raises(self) -> None:
        from src.application.commands.persona_handler import PersonaHandler

        writer = FakeFileWriter()
        handler = PersonaHandler(writer=writer)

        with pytest.raises(InvariantViolationError, match="Unknown persona"):
            handler.build_preview(persona_name="nonexistent", tool="claude-code")

    def test_build_preview_unknown_tool_raises(self) -> None:
        from src.application.commands.persona_handler import PersonaHandler

        writer = FakeFileWriter()
        handler = PersonaHandler(writer=writer)

        with pytest.raises(InvariantViolationError, match="Unsupported tool"):
            handler.build_preview(persona_name="Solo Developer", tool="unknown-tool")


# ---------------------------------------------------------------------------
# 4. build_preview — target paths per tool
# ---------------------------------------------------------------------------


class TestBuildPreviewTargetPaths:
    def test_build_preview_target_path_claude_code(self) -> None:
        from src.application.commands.persona_handler import PersonaHandler

        writer = FakeFileWriter()
        handler = PersonaHandler(writer=writer)

        preview = handler.build_preview(persona_name="Solo Developer", tool="claude-code")

        assert preview.target_path.startswith(".claude/agents/")
        assert preview.target_path.endswith(".md")

    def test_build_preview_target_path_cursor(self) -> None:
        from src.application.commands.persona_handler import PersonaHandler

        writer = FakeFileWriter()
        handler = PersonaHandler(writer=writer)

        preview = handler.build_preview(persona_name="Solo Developer", tool="cursor")

        assert preview.target_path.startswith(".cursor/rules/")
        assert preview.target_path.endswith(".mdc")

    def test_build_preview_target_path_roo_code(self) -> None:
        from src.application.commands.persona_handler import PersonaHandler

        writer = FakeFileWriter()
        handler = PersonaHandler(writer=writer)

        preview = handler.build_preview(persona_name="Solo Developer", tool="roo-code")

        assert preview.target_path.startswith(".roo-code/modes/")
        assert preview.target_path.endswith(".md")

    def test_build_preview_target_path_opencode(self) -> None:
        from src.application.commands.persona_handler import PersonaHandler

        writer = FakeFileWriter()
        handler = PersonaHandler(writer=writer)

        preview = handler.build_preview(persona_name="Solo Developer", tool="opencode")

        assert preview.target_path.startswith(".opencode/agents/")
        assert preview.target_path.endswith(".md")


# ---------------------------------------------------------------------------
# 5. approve_and_write
# ---------------------------------------------------------------------------


class TestApproveAndWrite:
    def test_approve_and_write_calls_writer(self) -> None:
        from src.application.commands.persona_handler import PersonaHandler

        writer = FakeFileWriter()
        handler = PersonaHandler(writer=writer)
        preview = handler.build_preview(persona_name="Solo Developer", tool="claude-code")

        handler.approve_and_write(preview, output_dir=Path("/project"))

        assert len(writer.written) == 1
        written_path = next(iter(writer.written.keys()))
        assert preview.target_path in written_path
        assert writer.written[written_path] == preview.content

    def test_approve_and_write_uses_output_dir(self) -> None:
        from src.application.commands.persona_handler import PersonaHandler

        writer = FakeFileWriter()
        handler = PersonaHandler(writer=writer)
        preview = handler.build_preview(persona_name="Team Lead", tool="cursor")

        handler.approve_and_write(preview, output_dir=Path("/my/project"))

        written_path = next(iter(writer.written.keys()))
        assert written_path.startswith("/my/project/")
