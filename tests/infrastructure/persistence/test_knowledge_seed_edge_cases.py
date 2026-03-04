"""Edge case tests for knowledge base seed content (2j7.5 QA).

BICEP analysis uncovered:
- Boundary: no upper bound test for content length (<2000 chars per AC)
- Cross-check: _index.toml category naming vs actual directory names
- Boundary: TOML files should also respect size bounds
"""

from __future__ import annotations

import tomllib
from pathlib import Path

import pytest

SEED_ROOT = (
    Path(__file__).resolve().parents[3]
    / "src"
    / "infrastructure"
    / "persistence"
    / "seed"
    / "knowledge"
)

# Markdown seed files (content files, not metadata)
MARKDOWN_SEED_FILES: list[str] = [
    "ddd/tactical-patterns.md",
    "ddd/strategic-patterns.md",
    "conventions/tdd.md",
    "conventions/solid.md",
    "conventions/quality-gates.md",
]

# All seed files
ALL_SEED_FILES: list[str] = [
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

MAX_CONTENT_LENGTH = 2000


# ---------------------------------------------------------------------------
# Boundary: Content length upper bound
# ---------------------------------------------------------------------------


class TestSeedContentUpperBound:
    """AC edge case: each file must be <2000 chars to stay concise."""

    @pytest.mark.parametrize("relative_path", ALL_SEED_FILES)
    def test_seed_file_under_max_length(self, relative_path: str) -> None:
        full_path = SEED_ROOT / relative_path
        content = full_path.read_text()
        assert len(content) <= MAX_CONTENT_LENGTH, (
            f"{relative_path} has {len(content)} chars,"
            f" maximum is {MAX_CONTENT_LENGTH}"
        )


# ---------------------------------------------------------------------------
# Cross-check: Index categories match directory structure
# ---------------------------------------------------------------------------


class TestIndexCategoryConsistency:
    """Index categories should be verifiable against actual directory layout."""

    def test_index_entry_categories_are_defined(self) -> None:
        """Every entry's category must appear in [categories]."""
        content = (SEED_ROOT / "_index.toml").read_text()
        parsed = tomllib.loads(content)
        defined = set(parsed.get("categories", {}).keys())
        for entry in parsed.get("entries", []):
            cat = entry.get("category", "")
            assert cat in defined, (
                f"Entry category '{cat}' not in defined categories {defined}"
            )

    def test_index_tool_entries_have_tool_field(self) -> None:
        """Tool entries should specify which tool they describe."""
        content = (SEED_ROOT / "_index.toml").read_text()
        parsed = tomllib.loads(content)
        for entry in parsed.get("entries", []):
            if entry.get("category") == "tools":
                assert "tool" in entry, (
                    f"Tools entry missing 'tool' field: {entry}"
                )

    def test_index_entries_topics_match_files(self) -> None:
        """Topics referenced in index entries should have corresponding files."""
        content = (SEED_ROOT / "_index.toml").read_text()
        parsed = tomllib.loads(content)
        for entry in parsed.get("entries", []):
            cat = entry.get("category", "")
            topics = entry.get("topics", [])
            tool = entry.get("tool")
            for topic in topics:
                # Build the expected path based on category
                if cat == "tools" and tool:
                    expected_dir = SEED_ROOT / cat / tool / "current"
                elif cat == "cross_tool":
                    expected_dir = SEED_ROOT / "cross-tool"
                else:
                    expected_dir = SEED_ROOT / cat
                # Check for .md or .toml variant
                md_path = expected_dir / f"{topic}.md"
                toml_path = expected_dir / f"{topic}.toml"
                assert md_path.exists() or toml_path.exists(), (
                    f"Topic '{topic}' for category '{cat}' not found as"
                    f" {md_path} or {toml_path}"
                )


# ---------------------------------------------------------------------------
# Boundary: Frontmatter date format
# ---------------------------------------------------------------------------


class TestSeedFrontmatterDateFormat:
    """Frontmatter dates should be valid ISO format."""

    @pytest.mark.parametrize("relative_path", MARKDOWN_SEED_FILES)
    def test_frontmatter_last_verified_is_iso_date(self, relative_path: str) -> None:
        """last_verified must be a valid ISO date string (YYYY-MM-DD)."""
        from datetime import date

        full_path = SEED_ROOT / relative_path
        content = full_path.read_text()
        # Extract frontmatter
        fm_start = content.index("---") + 3
        fm_end = content.index("---", fm_start)
        frontmatter = content[fm_start:fm_end]
        # Find last_verified value
        for line in frontmatter.strip().splitlines():
            if line.strip().startswith("last_verified"):
                _, _, value = line.partition(":")
                raw = value.strip().strip('"').strip("'")
                # Must parse as ISO date
                parsed = date.fromisoformat(raw)
                assert isinstance(parsed, date)
                return
        pytest.fail(f"{relative_path} missing last_verified in frontmatter")


# ---------------------------------------------------------------------------
# Cross-check: DDD terms match docs/DDD.md glossary
# ---------------------------------------------------------------------------


class TestDddSeedMatchesGlossary:
    """DDD seed content should reference terms from the ubiquitous language."""

    def test_tactical_patterns_mentions_key_terms(self) -> None:
        """tactical-patterns.md should mention Value Object, Entity, Aggregate."""
        content = (SEED_ROOT / "ddd" / "tactical-patterns.md").read_text().lower()
        assert "value object" in content or "value objects" in content
        assert "entit" in content  # entity or entities
        assert "aggregate" in content

    def test_strategic_patterns_mentions_key_terms(self) -> None:
        """strategic-patterns.md should mention Bounded Context, Ubiquitous Language."""
        content = (SEED_ROOT / "ddd" / "strategic-patterns.md").read_text().lower()
        assert "bounded context" in content
        assert "ubiquitous language" in content
