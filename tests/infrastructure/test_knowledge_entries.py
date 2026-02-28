"""Structural validation tests for .alty/knowledge/ entries."""

from __future__ import annotations

import tomllib
from pathlib import Path

import pytest

KNOWLEDGE_ROOT = Path(__file__).resolve().parents[2] / ".alty" / "knowledge"
TOOLS_DIR = KNOWLEDGE_ROOT / "tools"
CROSS_TOOL_DIR = KNOWLEDGE_ROOT / "cross-tool"
DDD_DIR = KNOWLEDGE_ROOT / "ddd"
CONVENTIONS_DIR = KNOWLEDGE_ROOT / "conventions"

REQUIRED_META_FIELDS = {
    "last_verified",
    "verified_against",
    "changelog_url",
    "source_urls",
    "confidence",
    "deprecated",
    "deprecated_since",
    "superseded_by",
    "next_review_date",
    "schema_version",
}

EXPECTED_TOOLS = ["claude-code", "cursor", "roo-code", "opencode"]

EXPECTED_TOPICS: dict[str, list[str]] = {
    "claude-code": [
        "config-structure",
        "agent-format",
        "settings-format",
        "rules-format",
        "commands-format",
        "mcp-config",
        "global-paths",
        "gitignore-patterns",
    ],
    "cursor": [
        "config-structure",
        "rules-format",
        "agents-md-support",
        "global-paths",
        "gitignore-patterns",
    ],
    "roo-code": [
        "config-structure",
        "mode-format",
        "rules-format",
        "global-paths",
        "gitignore-patterns",
    ],
    "opencode": [
        "config-structure",
        "agent-format",
        "mode-format",
        "rules-format",
        "opencode-json-schema",
        "global-paths",
        "gitignore-patterns",
    ],
}

EXPECTED_CROSS_TOOL = ["agents-md", "concept-mapping", "generation-matrix"]
EXPECTED_DDD = ["tactical-patterns", "strategic-patterns", "event-storming", "domain-storytelling"]
EXPECTED_CONVENTIONS = ["tdd", "solid", "quality-gates"]


def _all_toml_files() -> list[Path]:
    """Collect all .toml files under knowledge/tools/*/current/."""
    return sorted(TOOLS_DIR.glob("*/current/*.toml"))


def _all_cross_tool_toml_files() -> list[Path]:
    """Collect all .toml files under knowledge/cross-tool/."""
    return sorted(CROSS_TOOL_DIR.glob("*.toml"))


def _toml_id(p: Path) -> str:
    return str(p.relative_to(KNOWLEDGE_ROOT))


class TestTomlParsing:
    """All TOML files must parse without error."""

    @pytest.mark.parametrize("toml_file", _all_toml_files(), ids=_toml_id)
    def test_tool_toml_parses(self, toml_file: Path) -> None:
        data = tomllib.loads(toml_file.read_text())
        assert isinstance(data, dict)

    @pytest.mark.parametrize("toml_file", _all_cross_tool_toml_files(), ids=_toml_id)
    def test_cross_tool_toml_parses(self, toml_file: Path) -> None:
        data = tomllib.loads(toml_file.read_text())
        assert isinstance(data, dict)

    def test_index_toml_parses(self) -> None:
        index_path = KNOWLEDGE_ROOT / "_index.toml"
        assert index_path.exists(), "_index.toml missing"
        data = tomllib.loads(index_path.read_text())
        assert "_meta" in data

    @pytest.mark.parametrize("tool", EXPECTED_TOOLS)
    def test_meta_toml_parses(self, tool: str) -> None:
        meta_path = TOOLS_DIR / tool / "_meta.toml"
        assert meta_path.exists(), f"{tool}/_meta.toml missing"
        data = tomllib.loads(meta_path.read_text())
        assert "tool" in data
        assert "versions" in data


class TestMetaSections:
    """Every tool TOML entry must have [_meta] with required fields."""

    @pytest.mark.parametrize("toml_file", _all_toml_files(), ids=_toml_id)
    def test_tool_entry_has_meta(self, toml_file: Path) -> None:
        data = tomllib.loads(toml_file.read_text())
        assert "_meta" in data, f"{toml_file.name} missing [_meta] section"
        meta = data["_meta"]
        missing = REQUIRED_META_FIELDS - set(meta.keys())
        assert not missing, f"{toml_file.name} [_meta] missing fields: {missing}"

    @pytest.mark.parametrize("toml_file", _all_cross_tool_toml_files(), ids=_toml_id)
    def test_cross_tool_entry_has_meta(self, toml_file: Path) -> None:
        data = tomllib.loads(toml_file.read_text())
        assert "_meta" in data, f"{toml_file.name} missing [_meta] section"
        meta = data["_meta"]
        missing = REQUIRED_META_FIELDS - set(meta.keys())
        assert not missing, f"{toml_file.name} [_meta] missing fields: {missing}"

    @pytest.mark.parametrize("toml_file", _all_toml_files(), ids=_toml_id)
    def test_confidence_is_valid(self, toml_file: Path) -> None:
        data = tomllib.loads(toml_file.read_text())
        confidence = data["_meta"]["confidence"]
        assert confidence in ("high", "medium", "low"), f"Invalid confidence: {confidence}"

    @pytest.mark.parametrize("toml_file", _all_toml_files(), ids=_toml_id)
    def test_schema_version_is_int(self, toml_file: Path) -> None:
        data = tomllib.loads(toml_file.read_text())
        assert isinstance(data["_meta"]["schema_version"], int)


class TestCompleteness:
    """All expected files exist."""

    @pytest.mark.parametrize("tool", EXPECTED_TOOLS)
    def test_meta_toml_exists(self, tool: str) -> None:
        assert (TOOLS_DIR / tool / "_meta.toml").exists()

    @pytest.mark.parametrize(
        ("tool", "topic"),
        [(t, tp) for t in EXPECTED_TOOLS for tp in EXPECTED_TOPICS[t]],
        ids=[f"{t}/{tp}" for t in EXPECTED_TOOLS for tp in EXPECTED_TOPICS[t]],
    )
    def test_tool_entry_exists(self, tool: str, topic: str) -> None:
        path = TOOLS_DIR / tool / "current" / f"{topic}.toml"
        assert path.exists(), f"Missing: {path.relative_to(KNOWLEDGE_ROOT)}"

    @pytest.mark.parametrize("topic", EXPECTED_CROSS_TOOL)
    def test_cross_tool_entry_exists(self, topic: str) -> None:
        assert (CROSS_TOOL_DIR / f"{topic}.toml").exists()

    @pytest.mark.parametrize("topic", EXPECTED_DDD)
    def test_ddd_reference_exists(self, topic: str) -> None:
        assert (DDD_DIR / f"{topic}.md").exists()

    @pytest.mark.parametrize("topic", EXPECTED_CONVENTIONS)
    def test_convention_reference_exists(self, topic: str) -> None:
        assert (CONVENTIONS_DIR / f"{topic}.md").exists()

    def test_index_toml_exists(self) -> None:
        assert (KNOWLEDGE_ROOT / "_index.toml").exists()


class TestIndexCompleteness:
    """_index.toml must reference all entries."""

    def test_index_references_all_tools(self) -> None:
        data = tomllib.loads((KNOWLEDGE_ROOT / "_index.toml").read_text())
        entries = data.get("entries", [])
        tool_entries = [e for e in entries if e.get("category") == "tools"]
        indexed_tools = {e["tool"] for e in tool_entries}
        assert indexed_tools == set(EXPECTED_TOOLS)

    def test_index_references_all_tool_topics(self) -> None:
        data = tomllib.loads((KNOWLEDGE_ROOT / "_index.toml").read_text())
        entries = data.get("entries", [])
        for entry in entries:
            if entry.get("category") != "tools":
                continue
            tool = entry["tool"]
            indexed_topics = set(entry["topics"])
            expected = set(EXPECTED_TOPICS[tool])
            assert indexed_topics == expected, (
                f"{tool}: indexed={indexed_topics}, expected={expected}"
            )

    def test_index_references_cross_tool(self) -> None:
        data = tomllib.loads((KNOWLEDGE_ROOT / "_index.toml").read_text())
        entries = data.get("entries", [])
        cross_entries = [e for e in entries if e.get("category") == "cross_tool"]
        assert len(cross_entries) == 1
        assert set(cross_entries[0]["topics"]) == set(EXPECTED_CROSS_TOOL)

    def test_index_references_ddd(self) -> None:
        data = tomllib.loads((KNOWLEDGE_ROOT / "_index.toml").read_text())
        entries = data.get("entries", [])
        ddd_entries = [e for e in entries if e.get("category") == "ddd"]
        assert len(ddd_entries) == 1
        assert set(ddd_entries[0]["topics"]) == set(EXPECTED_DDD)

    def test_index_references_conventions(self) -> None:
        data = tomllib.loads((KNOWLEDGE_ROOT / "_index.toml").read_text())
        entries = data.get("entries", [])
        conv_entries = [e for e in entries if e.get("category") == "conventions"]
        assert len(conv_entries) == 1
        assert set(conv_entries[0]["topics"]) == set(EXPECTED_CONVENTIONS)


class TestNoOrphanFiles:
    """No TOML files exist that aren't tracked in the index."""

    def test_no_orphan_tool_entries(self) -> None:
        data = tomllib.loads((KNOWLEDGE_ROOT / "_index.toml").read_text())
        entries = data.get("entries", [])
        indexed: set[str] = set()
        for entry in entries:
            if entry.get("category") != "tools":
                continue
            tool = entry["tool"]
            for topic in entry["topics"]:
                indexed.add(f"{tool}/current/{topic}.toml")

        actual = {str(p.relative_to(TOOLS_DIR)) for p in TOOLS_DIR.glob("*/current/*.toml")}
        orphans = actual - indexed
        assert not orphans, f"Orphan files not in _index.toml: {orphans}"
