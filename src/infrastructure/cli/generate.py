"""Generate subcommand group: alty generate <subcommand>.

Subcommands:
    artifacts — Generate PRD, DDD.md, ARCHITECTURE.md
    fitness   — Generate import-linter + pytestarch tests
    tickets   — Generate beads epics + tasks from DDD
    configs   — Generate tool-specific configs (.claude/, .cursor/, etc.)

Reference: ARCHITECTURE.md §6.1
"""

from __future__ import annotations

import typer

app = typer.Typer(
    name="generate",
    help="Generate project artifacts, fitness tests, tickets, and configs.",
    no_args_is_help=True,
)


@app.command()
def artifacts() -> None:
    """Generate PRD, DDD.md, and ARCHITECTURE.md from domain model."""
    typer.echo("alty generate artifacts: not yet implemented")


@app.command()
def fitness() -> None:
    """Generate architecture fitness tests (import-linter + pytestarch)."""
    typer.echo("alty generate fitness: not yet implemented")


@app.command()
def tickets() -> None:
    """Generate beads epics and tasks from DDD stories."""
    typer.echo("alty generate tickets: not yet implemented")


@app.command()
def configs() -> None:
    """Generate tool-specific configs (.claude/, .cursor/, etc.)."""
    typer.echo("alty generate configs: not yet implemented")
