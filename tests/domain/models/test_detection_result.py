"""Tests for DetectionResult value object and ConflictSeverity enum.

Verifies immutability, severity classification, and field access.
"""

from __future__ import annotations

from dataclasses import FrozenInstanceError
from pathlib import Path

import pytest

from src.domain.models.detected_tool import DetectedTool
from src.domain.models.detection_result import ConflictSeverity, DetectionResult


class TestConflictSeverity:
    def test_compatible_value(self):
        assert ConflictSeverity.COMPATIBLE.value == "compatible"

    def test_warning_value(self):
        assert ConflictSeverity.WARNING.value == "warning"

    def test_conflict_value(self):
        assert ConflictSeverity.CONFLICT.value == "conflict"

    def test_all_severities(self):
        assert len(ConflictSeverity) == 3


class TestDetectionResultCreation:
    def test_create_with_tools_and_conflicts(self):
        tools = (
            DetectedTool(name="claude-code", config_path=Path("/home/user/.claude")),
            DetectedTool(name="cursor"),
        )
        conflicts = ("Global cursor setting overrides local",)
        severity_map = {"Global cursor setting overrides local": ConflictSeverity.WARNING}

        result = DetectionResult(
            detected_tools=tools,
            conflicts=conflicts,
            severity_map=severity_map,
        )

        assert len(result.detected_tools) == 2
        assert result.detected_tools[0].name == "claude-code"
        assert len(result.conflicts) == 1
        assert result.severity_map["Global cursor setting overrides local"] == (
            ConflictSeverity.WARNING
        )

    def test_create_empty_result(self):
        result = DetectionResult(
            detected_tools=(),
            conflicts=(),
            severity_map={},
        )
        assert len(result.detected_tools) == 0
        assert len(result.conflicts) == 0
        assert len(result.severity_map) == 0

    def test_create_with_multiple_severities(self):
        severity_map = {
            "Same value in both": ConflictSeverity.COMPATIBLE,
            "Cursor SQLite detected": ConflictSeverity.WARNING,
            "Contradicting settings": ConflictSeverity.CONFLICT,
        }
        result = DetectionResult(
            detected_tools=(),
            conflicts=(
                "Same value in both",
                "Cursor SQLite detected",
                "Contradicting settings",
            ),
            severity_map=severity_map,
        )
        assert result.severity_map["Same value in both"] == ConflictSeverity.COMPATIBLE
        assert result.severity_map["Cursor SQLite detected"] == ConflictSeverity.WARNING
        assert result.severity_map["Contradicting settings"] == ConflictSeverity.CONFLICT


class TestDetectionResultImmutability:
    def test_cannot_mutate_detected_tools(self):
        result = DetectionResult(
            detected_tools=(DetectedTool(name="claude-code"),),
            conflicts=(),
            severity_map={},
        )
        with pytest.raises(FrozenInstanceError):
            result.detected_tools = ()  # type: ignore[misc]

    def test_cannot_mutate_conflicts(self):
        result = DetectionResult(
            detected_tools=(),
            conflicts=("a conflict",),
            severity_map={},
        )
        with pytest.raises(FrozenInstanceError):
            result.conflicts = ()  # type: ignore[misc]

    def test_cannot_mutate_severity_map(self):
        result = DetectionResult(
            detected_tools=(),
            conflicts=("a conflict",),
            severity_map={"a conflict": ConflictSeverity.WARNING},
        )
        with pytest.raises(TypeError):
            result.severity_map["new_key"] = ConflictSeverity.CONFLICT  # type: ignore[index]
