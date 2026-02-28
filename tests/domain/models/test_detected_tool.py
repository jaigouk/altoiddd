"""Tests for DetectedTool value object.

Verifies immutability, equality, and field defaults.
"""

from __future__ import annotations

from dataclasses import FrozenInstanceError
from pathlib import Path

import pytest

from src.domain.models.detected_tool import DetectedTool


class TestDetectedToolCreation:
    def test_create_with_all_fields(self):
        tool = DetectedTool(
            name="claude-code",
            config_path=Path.home() / ".claude",
            version="1.2.3",
        )
        assert tool.name == "claude-code"
        assert tool.config_path == Path.home() / ".claude"
        assert tool.version == "1.2.3"

    def test_create_with_defaults(self):
        tool = DetectedTool(name="cursor")
        assert tool.name == "cursor"
        assert tool.config_path is None
        assert tool.version is None

    def test_create_with_config_path_only(self):
        tool = DetectedTool(name="roo-code", config_path=Path("/home/user/.roo"))
        assert tool.name == "roo-code"
        assert tool.config_path == Path("/home/user/.roo")
        assert tool.version is None


class TestDetectedToolImmutability:
    def test_cannot_mutate_name(self):
        tool = DetectedTool(name="claude-code")
        with pytest.raises(FrozenInstanceError):
            tool.name = "cursor"  # type: ignore[misc]

    def test_cannot_mutate_config_path(self):
        tool = DetectedTool(name="claude-code", config_path=Path("/a"))
        with pytest.raises(FrozenInstanceError):
            tool.config_path = Path("/b")  # type: ignore[misc]

    def test_cannot_mutate_version(self):
        tool = DetectedTool(name="claude-code", version="1.0")
        with pytest.raises(FrozenInstanceError):
            tool.version = "2.0"  # type: ignore[misc]


class TestDetectedToolEquality:
    def test_equal_tools_are_equal(self):
        t1 = DetectedTool(name="claude-code", config_path=Path("/a"), version="1.0")
        t2 = DetectedTool(name="claude-code", config_path=Path("/a"), version="1.0")
        assert t1 == t2

    def test_different_names_are_not_equal(self):
        t1 = DetectedTool(name="claude-code")
        t2 = DetectedTool(name="cursor")
        assert t1 != t2

    def test_different_versions_are_not_equal(self):
        t1 = DetectedTool(name="claude-code", version="1.0")
        t2 = DetectedTool(name="claude-code", version="2.0")
        assert t1 != t2
