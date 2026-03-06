"""Codebase port scanner — AST-based Protocol extraction.

Scans a directory of Python files for Protocol class definitions
and extracts method signatures. Used by regression tests and
future ripple automation to verify port references in tickets.
"""

from __future__ import annotations

import ast
import logging
from dataclasses import dataclass
from pathlib import Path  # noqa: TC003 — used at runtime in scan()

logger = logging.getLogger(__name__)


@dataclass(frozen=True)
class MethodSignature:
    """A single method extracted from a Protocol class.

    Attributes:
        name: Method name (e.g., 'generate_challenges').
        parameters: Parameter names excluding 'self', mapped to
                    their annotation string (or '' if untyped).
    """

    name: str
    parameters: dict[str, str]


@dataclass(frozen=True)
class PortDefinition:
    """A Protocol class found in a port file.

    Attributes:
        name: Class name (e.g., 'ChallengerPort').
        file_path: Path to the source file.
        methods: Methods defined on the Protocol.
    """

    name: str
    file_path: Path
    methods: tuple[MethodSignature, ...]


class CodebasePortScanner:
    """Scans Python source files for Protocol definitions via AST."""

    @staticmethod
    def scan(ports_dir: Path) -> dict[str, PortDefinition]:
        """Scan a directory for Protocol class definitions.

        Args:
            ports_dir: Directory containing port .py files.

        Returns:
            Dict mapping class name to PortDefinition.
        """
        result: dict[str, PortDefinition] = {}

        if not ports_dir.is_dir():
            return result

        for py_file in sorted(ports_dir.glob("*.py")):
            if py_file.name == "__init__.py":
                continue
            try:
                tree = ast.parse(py_file.read_text())
            except SyntaxError:
                logger.debug("Skipping malformed file: %s", py_file)
                continue

            for node in ast.walk(tree):
                if not isinstance(node, ast.ClassDef):
                    continue
                if not _is_protocol(node):
                    continue

                methods = _extract_methods(node)
                port = PortDefinition(
                    name=node.name,
                    file_path=py_file,
                    methods=tuple(methods),
                )
                result[node.name] = port

        return result


def _is_protocol(cls: ast.ClassDef) -> bool:
    """Check if a class inherits from Protocol."""
    for base in cls.bases:
        if isinstance(base, ast.Name) and base.id == "Protocol":
            return True
        if isinstance(base, ast.Attribute) and base.attr == "Protocol":
            return True
    return False


def _extract_methods(cls: ast.ClassDef) -> list[MethodSignature]:
    """Extract method signatures from a class body."""
    methods: list[MethodSignature] = []
    for item in cls.body:
        if not isinstance(item, (ast.FunctionDef, ast.AsyncFunctionDef)):
            continue
        if item.name.startswith("_"):
            continue

        params: dict[str, str] = {}
        for arg in item.args.args:
            if arg.arg == "self":
                continue
            annotation = ""
            if arg.annotation:
                annotation = ast.unparse(arg.annotation)
            params[arg.arg] = annotation

        methods.append(MethodSignature(name=item.name, parameters=params))
    return methods
