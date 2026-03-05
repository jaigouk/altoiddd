"""Shared test helpers for doc-review tests.

Used by both handler tests and CLI tests.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from pathlib import Path


def registry_toml(entries: list[dict[str, str | int]]) -> str:
    """Build a minimal [[docs]] TOML string from a list of dicts."""
    lines: list[str] = []
    for entry in entries:
        lines.append("[[docs]]")
        for k, v in entry.items():
            if isinstance(v, int):
                lines.append(f"{k} = {v}")
            else:
                lines.append(f'{k} = "{v}"')
        lines.append("")
    return "\n".join(lines)


def write_doc(
    project_dir: Path,
    rel_path: str,
    last_reviewed: str | None = None,
    extra_frontmatter: str = "",
    content_body: str = "# Doc\n",
) -> None:
    """Write a markdown doc with optional last_reviewed frontmatter."""
    doc_path = project_dir / rel_path
    doc_path.parent.mkdir(parents=True, exist_ok=True)
    if last_reviewed or extra_frontmatter:
        fm_lines: list[str] = []
        if last_reviewed:
            fm_lines.append(f"last_reviewed: {last_reviewed}")
        if extra_frontmatter:
            fm_lines.append(extra_frontmatter)
        content = f"---\n{chr(10).join(fm_lines)}\n---\n\n{content_body}"
    else:
        content = content_body
    doc_path.write_text(content)


def write_registry(
    project_dir: Path, entries: list[dict[str, Any]]
) -> None:
    """Write a doc-registry.toml to the project."""
    registry_dir = project_dir / ".alty" / "maintenance"
    registry_dir.mkdir(parents=True, exist_ok=True)
    (registry_dir / "doc-registry.toml").write_text(registry_toml(entries))
