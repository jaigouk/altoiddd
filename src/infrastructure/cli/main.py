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
def init() -> None:
    """Bootstrap a new project from a README idea."""
    typer.echo("alty init: not yet implemented")


@app.command()
def guide() -> None:
    """Run the 10-question guided DDD discovery flow."""
    typer.echo("alty guide: not yet implemented")


@app.command()
def detect() -> None:
    """Scan for installed AI coding tools and global settings."""
    typer.echo("alty detect: not yet implemented")


@app.command()
def check() -> None:
    """Run quality gates (lint, types, tests, fitness)."""
    typer.echo("alty check: not yet implemented")


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
