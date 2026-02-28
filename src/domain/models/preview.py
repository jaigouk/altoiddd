"""Value objects for the bootstrap preview.

FileAction, GlobalSettingConflict, ConflictResolution, and Preview are
immutable value objects that represent what alty *will* do before it does it.
The preview-before-action invariant guarantees users see every change first.

Design decisions:
- FileActionType has no OVERWRITE variant. Alty never silently overwrites.
- Conflict rename suffix is ``_alty``.
- All value objects are frozen dataclasses (stdlib only, no Pydantic).
"""

from __future__ import annotations

import enum
from dataclasses import dataclass
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from pathlib import Path


class FileActionType(enum.Enum):
    """The kind of file operation alty plans to perform.

    CREATE:          Write a new file that does not exist yet.
    SKIP:            Leave an existing file untouched.
    CONFLICT_RENAME: Rename an existing file with the ``_alty`` suffix
                     before writing the template version.
    """

    CREATE = "create"
    SKIP = "skip"
    CONFLICT_RENAME = "conflict_rename"


@dataclass(frozen=True)
class FileAction:
    """A single planned file operation.

    Attributes:
        path: Relative path from the project root.
        action_type: What alty will do with this file.
        reason: Human-readable explanation (e.g. "already exists").
        renamed_path: Only set for CONFLICT_RENAME actions.
    """

    path: Path
    action_type: FileActionType
    reason: str = ""
    renamed_path: Path | None = None


@dataclass(frozen=True)
class GlobalSettingConflict:
    """A conflict between a global tool setting and the local project config.

    Attributes:
        tool: Tool identifier (e.g. "cursor", "claude").
        global_path: Absolute path to the global settings file.
        global_value: Current global value as a string.
        local_value: Value alty wants to set locally.
    """

    tool: str
    global_path: Path
    global_value: str
    local_value: str


class ConflictResolution(enum.Enum):
    """How to resolve a GlobalSettingConflict.

    KEEP_GLOBAL:            Leave the global setting and skip local config.
    UPDATE_GLOBAL:          Update the global setting to the new value.
    SET_LOCAL_WITH_WARNING: Write local config with a warning comment.
    """

    KEEP_GLOBAL = "keep_global"
    UPDATE_GLOBAL = "update_global"
    SET_LOCAL_WITH_WARNING = "set_local_with_warning"


@dataclass(frozen=True)
class Preview:
    """Immutable snapshot of all planned bootstrap actions.

    Attributes:
        file_actions: Tuple of planned file operations.
        conflicts: Tuple of detected global-setting conflicts (may be empty).
    """

    file_actions: tuple[FileAction, ...]
    conflicts: tuple[GlobalSettingConflict, ...] = ()
    conflict_descriptions: tuple[str, ...] = ()
