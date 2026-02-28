"""Persona subcommand group: alty persona <subcommand>.

Subcommands:
    list     — Show available agent personas
    generate — Generate persona config for detected tools

Reference: ARCHITECTURE.md §6.1
"""

from __future__ import annotations

import typer

app = typer.Typer(
    name="persona",
    help="Manage agent persona configurations.",
    no_args_is_help=True,
)


@app.command(name="list")
def list_personas() -> None:
    """Show available agent personas."""
    typer.echo("alty persona list: not yet implemented")


@app.command()
def generate() -> None:
    """Generate persona config for detected tools."""
    typer.echo("alty persona generate: not yet implemented")
