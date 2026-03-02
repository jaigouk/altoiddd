"""Tests for knowledge base version history directories."""

from __future__ import annotations

import tomllib
from pathlib import Path

KB_ROOT = Path(__file__).resolve().parents[2] / ".alty" / "knowledge" / "tools"


class TestClaudeCodeVersions:
    """Claude Code must have v2.1 (current alias) and v2.0 (pre-subagent) directories."""

    def test_v2_1_dir_exists(self) -> None:
        assert (KB_ROOT / "claude-code" / "v2.1").is_dir()

    def test_v2_0_dir_exists(self) -> None:
        assert (KB_ROOT / "claude-code" / "v2.0").is_dir()

    def test_v2_1_matches_current_entry_count(self) -> None:
        current = [
            f for f in (KB_ROOT / "claude-code" / "current").glob("*.toml")
            if f.name != "_meta.toml"
        ]
        v2_1 = [
            f for f in (KB_ROOT / "claude-code" / "v2.1").glob("*.toml")
            if f.name != "_meta.toml"
        ]
        assert len(v2_1) == len(current)

    def test_v2_1_has_same_filenames_as_current(self) -> None:
        current_names = {
            f.name for f in (KB_ROOT / "claude-code" / "current").glob("*.toml")
            if f.name != "_meta.toml"
        }
        v2_1_names = {
            f.name for f in (KB_ROOT / "claude-code" / "v2.1").glob("*.toml")
            if f.name != "_meta.toml"
        }
        assert v2_1_names == current_names

    def test_v2_0_has_no_rules_format(self) -> None:
        """v2.0 should NOT have rules-format.toml (rules/ dir added in v2.1)."""
        assert not (KB_ROOT / "claude-code" / "v2.0" / "rules-format.toml").exists()

    def test_v2_0_has_no_agent_format(self) -> None:
        """v2.0 should NOT have agent-format.toml (subagents added in v2.1)."""
        assert not (KB_ROOT / "claude-code" / "v2.0" / "agent-format.toml").exists()

    def test_v2_0_verified_against_reflects_version(self) -> None:
        """v2.0 entries must have verified_against reflecting v2.0 range."""
        toml_files = list((KB_ROOT / "claude-code" / "v2.0").glob("*.toml"))
        assert len(toml_files) > 0, "v2.0 directory has no TOML entries"
        for toml_file in toml_files:
            data = tomllib.loads(toml_file.read_text())
            meta = data.get("_meta", {})
            assert "2.0" in meta.get("verified_against", ""), (
                f"{toml_file.name} verified_against should reference v2.0"
            )

    def test_v2_0_all_entries_parse(self) -> None:
        for toml_file in (KB_ROOT / "claude-code" / "v2.0").glob("*.toml"):
            data = tomllib.loads(toml_file.read_text())
            assert isinstance(data, dict)

    def test_v2_1_all_entries_parse(self) -> None:
        for toml_file in (KB_ROOT / "claude-code" / "v2.1").glob("*.toml"):
            data = tomllib.loads(toml_file.read_text())
            assert isinstance(data, dict)


class TestCursorVersions:
    """Cursor must have v2.5 (current alias) and v2.4 (pre-AGENTS.md) directories."""

    def test_v2_5_dir_exists(self) -> None:
        assert (KB_ROOT / "cursor" / "v2.5").is_dir()

    def test_v2_4_dir_exists(self) -> None:
        assert (KB_ROOT / "cursor" / "v2.4").is_dir()

    def test_v2_5_matches_current_entry_count(self) -> None:
        current = [
            f for f in (KB_ROOT / "cursor" / "current").glob("*.toml")
            if f.name != "_meta.toml"
        ]
        v2_5 = [
            f for f in (KB_ROOT / "cursor" / "v2.5").glob("*.toml")
            if f.name != "_meta.toml"
        ]
        assert len(v2_5) == len(current)

    def test_v2_5_has_same_filenames_as_current(self) -> None:
        current_names = {
            f.name for f in (KB_ROOT / "cursor" / "current").glob("*.toml")
            if f.name != "_meta.toml"
        }
        v2_5_names = {
            f.name for f in (KB_ROOT / "cursor" / "v2.5").glob("*.toml")
            if f.name != "_meta.toml"
        }
        assert v2_5_names == current_names

    def test_v2_4_has_no_agents_md_support(self) -> None:
        """v2.4 should NOT have agents-md-support.toml (added in v2.5)."""
        assert not (KB_ROOT / "cursor" / "v2.4" / "agents-md-support.toml").exists()

    def test_v2_4_verified_against_reflects_version(self) -> None:
        """v2.4 entries must have verified_against reflecting v2.4 range."""
        toml_files = list((KB_ROOT / "cursor" / "v2.4").glob("*.toml"))
        assert len(toml_files) > 0, "v2.4 directory has no TOML entries"
        for toml_file in toml_files:
            data = tomllib.loads(toml_file.read_text())
            meta = data.get("_meta", {})
            assert "2.4" in meta.get("verified_against", ""), (
                f"{toml_file.name} verified_against should reference v2.4"
            )

    def test_v2_4_all_entries_parse(self) -> None:
        for toml_file in (KB_ROOT / "cursor" / "v2.4").glob("*.toml"):
            data = tomllib.loads(toml_file.read_text())
            assert isinstance(data, dict)

    def test_v2_5_all_entries_parse(self) -> None:
        for toml_file in (KB_ROOT / "cursor" / "v2.5").glob("*.toml"):
            data = tomllib.loads(toml_file.read_text())
            assert isinstance(data, dict)


class TestMetaTomlVersionTracking:
    """_meta.toml files must include version tracking sections."""

    def test_claude_code_meta_has_v2_1(self) -> None:
        data = tomllib.loads(
            (KB_ROOT / "claude-code" / "_meta.toml").read_text()
        )
        versions = data.get("versions", {})
        assert "v2_1" in versions

    def test_claude_code_meta_has_v2_0(self) -> None:
        data = tomllib.loads(
            (KB_ROOT / "claude-code" / "_meta.toml").read_text()
        )
        versions = data.get("versions", {})
        assert "v2_0" in versions

    def test_claude_code_meta_v2_1_is_alias(self) -> None:
        data = tomllib.loads(
            (KB_ROOT / "claude-code" / "_meta.toml").read_text()
        )
        v2_1 = data["versions"]["v2_1"]
        assert v2_1.get("alias") == "current"

    def test_claude_code_meta_v2_0_has_breaking_changes(self) -> None:
        data = tomllib.loads(
            (KB_ROOT / "claude-code" / "_meta.toml").read_text()
        )
        v2_0 = data["versions"]["v2_0"]
        assert "breaking_changes" in v2_0
        assert isinstance(v2_0["breaking_changes"], list)
        assert len(v2_0["breaking_changes"]) > 0

    def test_claude_code_tracked_includes_both_versions(self) -> None:
        data = tomllib.loads(
            (KB_ROOT / "claude-code" / "_meta.toml").read_text()
        )
        tracked = data["versions"]["tracked"]
        assert "v2.1" in tracked
        assert "v2.0" in tracked

    def test_cursor_meta_has_v2_5(self) -> None:
        data = tomllib.loads(
            (KB_ROOT / "cursor" / "_meta.toml").read_text()
        )
        versions = data.get("versions", {})
        assert "v2_5" in versions

    def test_cursor_meta_has_v2_4(self) -> None:
        data = tomllib.loads(
            (KB_ROOT / "cursor" / "_meta.toml").read_text()
        )
        versions = data.get("versions", {})
        assert "v2_4" in versions

    def test_cursor_meta_v2_5_is_alias(self) -> None:
        data = tomllib.loads(
            (KB_ROOT / "cursor" / "_meta.toml").read_text()
        )
        v2_5 = data["versions"]["v2_5"]
        assert v2_5.get("alias") == "current"

    def test_cursor_meta_v2_4_has_breaking_changes(self) -> None:
        data = tomllib.loads(
            (KB_ROOT / "cursor" / "_meta.toml").read_text()
        )
        v2_4 = data["versions"]["v2_4"]
        assert "breaking_changes" in v2_4
        assert isinstance(v2_4["breaking_changes"], list)
        assert len(v2_4["breaking_changes"]) > 0

    def test_cursor_tracked_includes_both_versions(self) -> None:
        data = tomllib.loads(
            (KB_ROOT / "cursor" / "_meta.toml").read_text()
        )
        tracked = data["versions"]["tracked"]
        assert "v2.5" in tracked
        assert "v2.4" in tracked


class TestV20ConfigStructureNoSubagentRefs:
    """v2.0 config-structure.toml must not reference v2.1 additions."""

    def test_no_rules_directory_ref(self) -> None:
        data = tomllib.loads(
            (KB_ROOT / "claude-code" / "v2.0" / "config-structure.toml").read_text()
        )
        project = data.get("project_structure", {})
        for key in project:
            assert "rules/" not in key, (
                f"v2.0 config-structure should not reference rules/ directory: {key}"
            )

    def test_no_skills_ref(self) -> None:
        data = tomllib.loads(
            (KB_ROOT / "claude-code" / "v2.0" / "config-structure.toml").read_text()
        )
        project = data.get("project_structure", {})
        for key in project:
            assert "skills/" not in key, (
                f"v2.0 config-structure should not reference skills/: {key}"
            )

    def test_no_agent_memory_ref(self) -> None:
        data = tomllib.loads(
            (KB_ROOT / "claude-code" / "v2.0" / "config-structure.toml").read_text()
        )
        project = data.get("project_structure", {})
        for key in project:
            assert "agent-memory" not in key, (
                f"v2.0 config-structure should not reference agent-memory: {key}"
            )


class TestV24CursorNoAgentsMd:
    """v2.4 Cursor config must not reference AGENTS.md."""

    def test_config_structure_no_agents_md(self) -> None:
        data = tomllib.loads(
            (KB_ROOT / "cursor" / "v2.4" / "config-structure.toml").read_text()
        )
        project = data.get("project_structure", {})
        for key in project:
            assert "AGENTS.md" not in key, (
                f"v2.4 config-structure should not reference AGENTS.md: {key}"
            )

    def test_config_precedence_no_agents_md(self) -> None:
        data = tomllib.loads(
            (KB_ROOT / "cursor" / "v2.4" / "config-structure.toml").read_text()
        )
        precedence = data.get("config_precedence", {})
        order = precedence.get("order", [])
        for item in order:
            assert "AGENTS.md" not in item, (
                f"v2.4 config precedence should not reference AGENTS.md: {item}"
            )
