"""Tests for the knowledge base seed content bundle.

Verifies that the MVP 12-file seed bundle at
``src/infrastructure/persistence/seed/knowledge/`` contains all required
files with valid structure, non-empty content, and proper metadata.
"""

from __future__ import annotations

import tomllib
from pathlib import Path

import pytest

# ---------------------------------------------------------------------------
# Constants
# ---------------------------------------------------------------------------

SEED_ROOT = (
    Path(__file__).resolve().parents[3]
    / "src"
    / "infrastructure"
    / "persistence"
    / "seed"
    / "knowledge"
)

EXPECTED_SEED_FILES: list[str] = [
    "_index.toml",
    "ddd/tactical-patterns.md",
    "ddd/strategic-patterns.md",
    "conventions/tdd.md",
    "conventions/solid.md",
    "conventions/quality-gates.md",
    "tools/claude-code/_meta.toml",
    "tools/claude-code/current/config-structure.toml",
    "tools/claude-code/current/agent-format.toml",
    "tools/claude-code/current/global-paths.toml",
    "cross-tool/concept-mapping.toml",
    "maintenance/doc-registry.toml",
]

MARKDOWN_FILES: list[str] = [f for f in EXPECTED_SEED_FILES if f.endswith(".md")]
TOML_FILES: list[str] = [f for f in EXPECTED_SEED_FILES if f.endswith(".toml")]

# Minimum content length in characters for seed files
MIN_CONTENT_LENGTH = 100


# ---------------------------------------------------------------------------
# Tests: File Presence
# ---------------------------------------------------------------------------


class TestSeedFilesExist:
    """All 12 expected seed files must be present in the package."""

    @pytest.mark.parametrize("relative_path", EXPECTED_SEED_FILES)
    def test_seed_file_exists(self, relative_path: str) -> None:
        full_path = SEED_ROOT / relative_path
        assert full_path.exists(), f"Missing seed file: {relative_path}"

    def test_seed_file_count(self) -> None:
        """Exactly 12 seed files are expected."""
        assert len(EXPECTED_SEED_FILES) == 12


# ---------------------------------------------------------------------------
# Tests: Valid Structure
# ---------------------------------------------------------------------------


class TestSeedFilesHaveValidStructure:
    """TOML files must parse without error; markdown files must have content."""

    @pytest.mark.parametrize("relative_path", TOML_FILES)
    def test_toml_files_parse(self, relative_path: str) -> None:
        full_path = SEED_ROOT / relative_path
        content = full_path.read_text()
        parsed = tomllib.loads(content)
        assert isinstance(parsed, dict), f"{relative_path} did not parse to a dict"

    @pytest.mark.parametrize("relative_path", MARKDOWN_FILES)
    def test_markdown_files_have_content(self, relative_path: str) -> None:
        full_path = SEED_ROOT / relative_path
        content = full_path.read_text()
        assert len(content.strip()) > 0, f"{relative_path} is empty"


# ---------------------------------------------------------------------------
# Tests: DDD Topics Complete
# ---------------------------------------------------------------------------


class TestSeedDddTopicsComplete:
    """ddd/ directory must contain tactical-patterns and strategic-patterns."""

    def test_tactical_patterns_present(self) -> None:
        assert (SEED_ROOT / "ddd" / "tactical-patterns.md").exists()

    def test_strategic_patterns_present(self) -> None:
        assert (SEED_ROOT / "ddd" / "strategic-patterns.md").exists()


# ---------------------------------------------------------------------------
# Tests: Tools Complete
# ---------------------------------------------------------------------------


class TestSeedToolsComplete:
    """tools/ must have claude-code with _meta.toml and current/ subtopics."""

    def test_claude_code_meta_present(self) -> None:
        assert (SEED_ROOT / "tools" / "claude-code" / "_meta.toml").exists()

    def test_claude_code_current_dir_exists(self) -> None:
        current_dir = SEED_ROOT / "tools" / "claude-code" / "current"
        assert current_dir.is_dir(), "tools/claude-code/current/ directory missing"

    def test_config_structure_present(self) -> None:
        assert (SEED_ROOT / "tools" / "claude-code" / "current" / "config-structure.toml").exists()

    def test_agent_format_present(self) -> None:
        assert (SEED_ROOT / "tools" / "claude-code" / "current" / "agent-format.toml").exists()

    def test_global_paths_present(self) -> None:
        assert (SEED_ROOT / "tools" / "claude-code" / "current" / "global-paths.toml").exists()


# ---------------------------------------------------------------------------
# Tests: Conventions Complete
# ---------------------------------------------------------------------------


class TestSeedConventionsComplete:
    """conventions/ must have tdd, solid, and quality-gates."""

    def test_tdd_present(self) -> None:
        assert (SEED_ROOT / "conventions" / "tdd.md").exists()

    def test_solid_present(self) -> None:
        assert (SEED_ROOT / "conventions" / "solid.md").exists()

    def test_quality_gates_present(self) -> None:
        assert (SEED_ROOT / "conventions" / "quality-gates.md").exists()


# ---------------------------------------------------------------------------
# Tests: Content Non-Empty
# ---------------------------------------------------------------------------


class TestSeedContentNonempty:
    """Each seed file must have more than 100 characters of content."""

    @pytest.mark.parametrize("relative_path", EXPECTED_SEED_FILES)
    def test_file_has_minimum_content(self, relative_path: str) -> None:
        full_path = SEED_ROOT / relative_path
        content = full_path.read_text()
        assert len(content) > MIN_CONTENT_LENGTH, (
            f"{relative_path} has only {len(content)} chars, minimum is {MIN_CONTENT_LENGTH}"
        )


# ---------------------------------------------------------------------------
# Tests: Markdown Frontmatter
# ---------------------------------------------------------------------------


class TestSeedMarkdownHasFrontmatter:
    """Markdown files must have YAML frontmatter with last_verified, confidence."""

    @pytest.mark.parametrize("relative_path", MARKDOWN_FILES)
    def test_has_yaml_frontmatter(self, relative_path: str) -> None:
        full_path = SEED_ROOT / relative_path
        content = full_path.read_text()
        assert content.startswith("---"), (
            f"{relative_path} must start with YAML frontmatter delimiter '---'"
        )
        # Must have closing delimiter
        second_delimiter = content.index("---", 3)
        assert second_delimiter > 3, f"{relative_path} must have closing '---' for frontmatter"

    @pytest.mark.parametrize("relative_path", MARKDOWN_FILES)
    def test_frontmatter_has_last_verified(self, relative_path: str) -> None:
        full_path = SEED_ROOT / relative_path
        content = full_path.read_text()
        frontmatter = _extract_frontmatter(content)
        assert "last_verified" in frontmatter, (
            f"{relative_path} frontmatter must contain 'last_verified'"
        )

    @pytest.mark.parametrize("relative_path", MARKDOWN_FILES)
    def test_frontmatter_has_confidence(self, relative_path: str) -> None:
        full_path = SEED_ROOT / relative_path
        content = full_path.read_text()
        frontmatter = _extract_frontmatter(content)
        assert "confidence" in frontmatter, (
            f"{relative_path} frontmatter must contain 'confidence'"
        )

    @pytest.mark.parametrize("relative_path", MARKDOWN_FILES)
    def test_frontmatter_has_next_review_date(self, relative_path: str) -> None:
        full_path = SEED_ROOT / relative_path
        content = full_path.read_text()
        frontmatter = _extract_frontmatter(content)
        assert "next_review_date" in frontmatter, (
            f"{relative_path} frontmatter must contain 'next_review_date'"
        )


# ---------------------------------------------------------------------------
# Tests: Index TOML Categories
# ---------------------------------------------------------------------------


class TestIndexTomlListsCategories:
    """_index.toml must have a [categories] section listing 4+ categories."""

    def test_index_has_categories_section(self) -> None:
        content = (SEED_ROOT / "_index.toml").read_text()
        parsed = tomllib.loads(content)
        assert "categories" in parsed, "_index.toml must have a [categories] section"

    def test_index_has_at_least_four_categories(self) -> None:
        content = (SEED_ROOT / "_index.toml").read_text()
        parsed = tomllib.loads(content)
        categories = parsed["categories"]
        assert len(categories) >= 4, (
            f"_index.toml [categories] has {len(categories)} entries, need >= 4"
        )

    def test_index_lists_ddd_category(self) -> None:
        content = (SEED_ROOT / "_index.toml").read_text()
        parsed = tomllib.loads(content)
        assert "ddd" in parsed["categories"]

    def test_index_lists_tools_category(self) -> None:
        content = (SEED_ROOT / "_index.toml").read_text()
        parsed = tomllib.loads(content)
        assert "tools" in parsed["categories"]

    def test_index_lists_conventions_category(self) -> None:
        content = (SEED_ROOT / "_index.toml").read_text()
        parsed = tomllib.loads(content)
        assert "conventions" in parsed["categories"]

    def test_index_lists_maintenance_category(self) -> None:
        content = (SEED_ROOT / "_index.toml").read_text()
        parsed = tomllib.loads(content)
        assert "maintenance" in parsed["categories"]


# ---------------------------------------------------------------------------
# Tests: Concept Mapping Entries
# ---------------------------------------------------------------------------


class TestConceptMappingHasEntries:
    """cross-tool/concept-mapping.toml must have tool entries."""

    def test_concept_mapping_has_concept_map(self) -> None:
        content = (SEED_ROOT / "cross-tool" / "concept-mapping.toml").read_text()
        parsed = tomllib.loads(content)
        assert "concept_map" in parsed, "concept-mapping.toml must have a [concept_map] section"

    def test_concept_mapping_has_at_least_one_concept(self) -> None:
        content = (SEED_ROOT / "cross-tool" / "concept-mapping.toml").read_text()
        parsed = tomllib.loads(content)
        concept_map = parsed["concept_map"]
        # Exclude the top-level 'description' key; remaining are concept entries
        concepts = [k for k in concept_map if k != "description"]
        assert len(concepts) >= 1, "concept-mapping.toml must have at least one concept entry"

    def test_concept_entries_reference_tools(self) -> None:
        content = (SEED_ROOT / "cross-tool" / "concept-mapping.toml").read_text()
        parsed = tomllib.loads(content)
        concept_map = parsed["concept_map"]
        concepts = {k: v for k, v in concept_map.items() if k != "description"}
        for name, entry in concepts.items():
            assert isinstance(entry, dict), (
                f"concept_map.{name} must be a table, got {type(entry).__name__}"
            )
            # At least one tool reference in the concept
            tool_keys = {"claude_code", "cursor", "roo_code", "opencode"}
            found = tool_keys & set(entry.keys())
            assert len(found) >= 1, (
                f"concept_map.{name} must reference at least one tool, "
                f"found keys: {list(entry.keys())}"
            )


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _extract_frontmatter(content: str) -> str:
    """Extract YAML frontmatter from markdown content."""
    if not content.startswith("---"):
        return ""
    end = content.index("---", 3)
    return content[3:end]
