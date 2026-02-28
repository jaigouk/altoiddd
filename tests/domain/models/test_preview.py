"""Tests for Preview value objects.

Verifies FileAction, GlobalSettingConflict, ConflictResolution, and Preview
are immutable frozen dataclasses with the correct invariants.
"""

from __future__ import annotations

from dataclasses import FrozenInstanceError
from pathlib import Path

import pytest

from src.domain.models.preview import (
    ConflictResolution,
    FileAction,
    FileActionType,
    GlobalSettingConflict,
    Preview,
)

# ── FileAction ───────────────────────────────────────────────


class TestFileAction:
    def test_create_action(self):
        action = FileAction(path=Path("docs/PRD.md"), action_type=FileActionType.CREATE)
        assert action.path == Path("docs/PRD.md")
        assert action.action_type == FileActionType.CREATE
        assert action.reason == ""
        assert action.renamed_path is None

    def test_skip_action_for_existing_file(self):
        action = FileAction(
            path=Path("README.md"),
            action_type=FileActionType.SKIP,
            reason="already exists",
        )
        assert action.action_type == FileActionType.SKIP
        assert action.reason == "already exists"

    def test_conflict_rename_uses_alty_suffix(self):
        original = Path(".claude/CLAUDE.md")
        renamed = Path(".claude/CLAUDE_alty.md")
        action = FileAction(
            path=original,
            action_type=FileActionType.CONFLICT_RENAME,
            reason="existing file conflicts with template",
            renamed_path=renamed,
        )
        assert action.action_type == FileActionType.CONFLICT_RENAME
        assert action.renamed_path == renamed
        assert "_alty" in str(action.renamed_path)

    def test_file_action_is_immutable(self):
        action = FileAction(path=Path("a.txt"), action_type=FileActionType.CREATE)
        with pytest.raises(FrozenInstanceError):
            action.path = Path("b.txt")  # type: ignore[misc]


# ── Preview ──────────────────────────────────────────────────


class TestPreview:
    def test_preview_is_immutable(self):
        preview = Preview(
            file_actions=(FileAction(path=Path("a.txt"), action_type=FileActionType.CREATE),),
        )
        with pytest.raises(FrozenInstanceError):
            preview.file_actions = ()  # type: ignore[misc]

    def test_preview_contains_file_actions(self):
        actions = (
            FileAction(path=Path("docs/PRD.md"), action_type=FileActionType.CREATE),
            FileAction(
                path=Path("README.md"),
                action_type=FileActionType.SKIP,
                reason="already exists",
            ),
        )
        preview = Preview(file_actions=actions)
        assert len(preview.file_actions) == 2
        assert preview.file_actions[0].action_type == FileActionType.CREATE
        assert preview.file_actions[1].action_type == FileActionType.SKIP

    def test_preview_never_has_overwrite_action(self):
        """FileActionType has no OVERWRITE variant -- verify at the enum level."""
        action_names = [a.value for a in FileActionType]
        assert "overwrite" not in action_names

    def test_preview_default_conflict_descriptions_empty(self):
        preview = Preview(
            file_actions=(FileAction(path=Path("a.txt"), action_type=FileActionType.CREATE),),
        )
        assert preview.conflict_descriptions == ()

    def test_preview_with_conflict_descriptions(self):
        preview = Preview(
            file_actions=(FileAction(path=Path("a.txt"), action_type=FileActionType.CREATE),),
            conflict_descriptions=("Global cursor setting overrides local",),
        )
        assert len(preview.conflict_descriptions) == 1
        assert "cursor" in preview.conflict_descriptions[0]


# ── GlobalSettingConflict ────────────────────────────────────


class TestGlobalSettingConflict:
    def test_conflict_has_tool_and_paths(self):
        conflict = GlobalSettingConflict(
            tool="cursor",
            global_path=Path.home() / ".cursor" / "settings.json",
            global_value='{"theme": "dark"}',
            local_value='{"theme": "light"}',
        )
        assert conflict.tool == "cursor"
        assert conflict.global_path.name == "settings.json"
        assert conflict.global_value != conflict.local_value

    def test_conflict_resolution_options(self):
        assert ConflictResolution.KEEP_GLOBAL.value == "keep_global"
        assert ConflictResolution.UPDATE_GLOBAL.value == "update_global"
        assert ConflictResolution.SET_LOCAL_WITH_WARNING.value == "set_local_with_warning"
