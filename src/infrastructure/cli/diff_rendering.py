"""CLI rendering functions for artifact diffs.

Formats ArtifactDiff objects as plain text for terminal display.
Uses +/~/- markers for added/modified/removed entries.
"""

from __future__ import annotations

from collections import defaultdict

from src.domain.models.artifact_diff import ArtifactDiff, DiffEntry, DiffType

_MARKERS = {
    DiffType.ADDED: "+",
    DiffType.MODIFIED: "~",
    DiffType.REMOVED: "-",
    DiffType.DISAMBIGUATED: "~",
}


def format_diff(diff: ArtifactDiff, trend: str) -> str:
    """Format an ArtifactDiff as a plain text string.

    Args:
        diff: The ArtifactDiff to render.
        trend: Convergence trend label ('active refinement', 'stabilizing', 'converged').

    Returns:
        Formatted string suitable for typer.echo().
    """
    lines: list[str] = []
    lines.append(f"Diff: v{diff.from_version} -> v{diff.to_version}")
    lines.append("")

    if not diff.entries:
        lines.append("No changes detected.")
    else:
        grouped = _group_by_section(diff.entries)
        for section, entries in sorted(grouped.items()):
            lines.append(f"  {section}:")
            for entry in entries:
                marker = _MARKERS[entry.diff_type]
                lines.append(f"    {marker} {entry.description}")
            lines.append("")

    c = diff.convergence
    lines.append(
        f"Changes: terms={c.terms_delta}, stories={c.stories_delta}, "
        f"invariants={c.invariants_delta}, canvases={c.canvases_delta}"
    )
    lines.append(f"Trend: {trend}")

    return "\n".join(lines)


def _group_by_section(entries: tuple[DiffEntry, ...]) -> dict[str, list[DiffEntry]]:
    grouped: dict[str, list[DiffEntry]] = defaultdict(list)
    for entry in entries:
        grouped[entry.section].append(entry)
    return grouped
