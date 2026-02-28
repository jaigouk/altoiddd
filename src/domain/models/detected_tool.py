"""DetectedTool value object.

Immutable representation of a detected AI coding tool, including
its name, optional configuration path, and optional version.
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from pathlib import Path


@dataclass(frozen=True)
class DetectedTool:
    """An AI coding tool detected on the system.

    Attributes:
        name: Tool identifier (e.g. "claude-code", "cursor", "roo-code").
        config_path: Relative path to the tool's config directory (relative to
            home), or None. Resolution to absolute is infrastructure's job.
        version: Detected version string, or None if unknown.
    """

    name: str
    config_path: Path | None = None
    version: str | None = None
