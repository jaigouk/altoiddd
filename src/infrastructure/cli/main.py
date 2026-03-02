"""CLI entry point for alty.

Root Typer application with 9 top-level commands and 2 subcommand groups
(generate, persona). Each command is a stub that will be replaced by
application-layer handler calls in downstream tickets.

Entry point: alty = "src.infrastructure.cli.main:app"
Reference: ARCHITECTURE.md §6.1 CLI Command Tree
"""

from __future__ import annotations

from importlib.metadata import version

import typer

from src.infrastructure.cli import generate, persona

app = typer.Typer(
    name="alty",
    help="The AI architect for vibe coding — guided project bootstrapper.",
    no_args_is_help=True,
)

app.add_typer(generate.app, name="generate")
app.add_typer(persona.app, name="persona")


# ── Top-level commands ───────────────────────────────────────


@app.command()
def init(
    existing: bool = typer.Option(False, "--existing", help="Rescue an existing project."),
) -> None:
    """Bootstrap a new project from a README idea."""
    if existing:
        from pathlib import Path

        from src.application.commands.rescue_handler import RescueHandler
        from src.domain.models.errors import InvariantViolationError
        from src.infrastructure.git.git_ops_adapter import GitOpsAdapter
        from src.infrastructure.scanner.project_scanner import ProjectScanner

        scanner = ProjectScanner()
        git_ops = GitOpsAdapter()
        handler = RescueHandler(project_scan=scanner, git_ops=git_ops)
        project_dir = Path.cwd()
        try:
            analysis = handler.rescue(project_dir)
            if not analysis.gaps:
                typer.echo("No gaps found -- project already follows alty conventions.")
                return
            typer.echo(f"Found {len(analysis.gaps)} gap(s):")
            for gap in analysis.gaps:
                typer.echo(f"  [{gap.gap_type.value}] {gap.path}: {gap.description}")
        except InvariantViolationError as e:
            typer.echo(f"Rescue failed: {e}", err=True)
            raise typer.Exit(code=1) from None
    else:
        typer.echo("alty init: not yet implemented")


@app.command()
def guide() -> None:
    """Run the 10-question guided DDD discovery flow."""
    from src.application.commands.discovery_handler import DiscoveryHandler

    handler = DiscoveryHandler()
    session = handler.start_session(readme_content="")
    typer.echo(f"alty guide: session started ({session.session_id})")


@app.command()
def detect(
    project_dir: str = typer.Argument(
        ".",
        help="Project directory to scan (defaults to current directory).",
    ),
) -> None:
    """Scan for installed AI coding tools and global settings."""
    from pathlib import Path

    from src.application.commands.detection_handler import DetectionHandler
    from src.domain.models.detection_result import ConflictSeverity
    from src.infrastructure.external.filesystem_tool_scanner import FilesystemToolScanner

    resolved_dir = Path(project_dir).resolve()
    scanner = FilesystemToolScanner()
    handler = DetectionHandler(tool_detection=scanner)
    result = handler.detect(resolved_dir)

    if not result.detected_tools:
        typer.echo("No AI coding tools detected.")
        return

    typer.echo("Detected AI coding tools:")
    for tool in result.detected_tools:
        config_info = f" ({tool.config_path})" if tool.config_path else ""
        typer.echo(f"  - {tool.name}{config_info}")

    if result.conflicts:
        typer.echo("\nConfiguration conflicts:")
        for conflict in result.conflicts:
            severity = result.severity_map.get(conflict, ConflictSeverity.WARNING)
            typer.echo(f"  [{severity.value.upper()}] {conflict}")


@app.command()
def check(
    gate: str = typer.Option(
        "", "--gate", help="Run a specific gate: lint, types, tests, fitness."
    ),
) -> None:
    """Run quality gates (lint, types, tests, fitness)."""
    from src.domain.models.quality_gate import QualityGate
    from src.infrastructure.composition import create_app

    ctx = create_app()

    gates_arg: tuple[QualityGate, ...] | None = None
    if gate:
        try:
            gates_arg = (QualityGate(gate),)
        except ValueError:
            valid = ", ".join(g.value for g in QualityGate)
            typer.echo(f"Invalid gate: {gate!r}. Valid gates: {valid}", err=True)
            raise typer.Exit(code=1) from None

    report = ctx.quality_gate.check(gates=gates_arg)

    for result in report.results:
        status = "PASS" if result.passed else "FAIL"
        typer.echo(f"  [{status}] {result.gate.value} ({result.duration_ms}ms)")
        if not result.passed:
            for line in result.output.strip().splitlines():
                typer.echo(f"         {line}")

    if report.passed:
        typer.echo(f"\nAll {len(report.results)} quality gate(s) passed.")
    else:
        failed = sum(1 for r in report.results if not r.passed)
        typer.echo(f"\n{failed} quality gate(s) failed.", err=True)
        raise typer.Exit(code=1)


@app.command()
def kb(topic: str = typer.Argument("", help="Knowledge base topic to look up.")) -> None:
    """Look up a topic in the RLM knowledge base."""
    typer.echo(f"alty kb: not yet implemented (topic={topic!r})")


@app.command(name="doc-health")
def doc_health() -> None:
    """Check documentation freshness and health."""
    typer.echo("alty doc-health: not yet implemented")


@app.command(name="doc-review")
def doc_review() -> None:
    """Mark documentation as reviewed."""
    typer.echo("alty doc-review: not yet implemented")


@app.command(name="ticket-health")
def ticket_health() -> None:
    """Show ripple review report for tickets needing attention."""
    typer.echo("alty ticket-health: not yet implemented")


@app.command(name="version")
def version_cmd() -> None:
    """Show the alty version."""
    pkg_version = version("alty")
    typer.echo(f"alty {pkg_version}")
