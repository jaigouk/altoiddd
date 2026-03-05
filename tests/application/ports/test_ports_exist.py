"""Tests for application layer port Protocols.

Verifies all 17 port Protocols exist, are importable, are runtime-checkable
Protocol subclasses, expose the required methods, and have no external
dependencies.
"""

from __future__ import annotations

import ast
import inspect
from pathlib import Path
from typing import Protocol, runtime_checkable

import pytest

# ---------------------------------------------------------------------------
# 1. All 17 Protocols exist and are importable from the ports package
# ---------------------------------------------------------------------------

ALL_PORT_NAMES = [
    "ArtifactRendererPort",
    "BootstrapPort",
    "ConfigGenerationPort",
    "DiscoveryPort",
    "DocHealthPort",
    "DocReviewPort",
    "DriftDetectionPort",
    "FileWriterPort",
    "FitnessGenerationPort",
    "GateRunnerProtocol",
    "KnowledgeLookupPort",
    "PersonaPort",
    "QualityGatePort",
    "RescuePort",
    "TicketGenerationPort",
    "TicketHealthPort",
    "ToolDetectionPort",
]


@pytest.mark.parametrize("name", ALL_PORT_NAMES)
def test_protocol_importable_from_package(name):
    """Each Protocol is re-exported from src.application.ports."""
    from src.application import ports

    assert hasattr(ports, name), f"{name} not found in src.application.ports"


def test_all_dunder_exports():
    """__all__ in the ports package lists exactly the 17 Protocols."""
    from src.application import ports

    assert hasattr(ports, "__all__"), "ports package has no __all__"
    assert sorted(ports.__all__) == sorted(ALL_PORT_NAMES)


# ---------------------------------------------------------------------------
# 2. Each is a typing.Protocol subclass
# ---------------------------------------------------------------------------


@pytest.mark.parametrize("name", ALL_PORT_NAMES)
def test_is_protocol_subclass(name):
    """Each exported name is a Protocol."""
    from src.application import ports

    cls = getattr(ports, name)
    assert issubclass(cls, Protocol), f"{name} is not a Protocol subclass"  # type: ignore[arg-type]


@pytest.mark.parametrize("name", ALL_PORT_NAMES)
def test_is_runtime_checkable(name):
    """Each Protocol is decorated with @runtime_checkable."""
    from src.application import ports

    cls = getattr(ports, name)
    assert getattr(cls, "__protocol_attrs__", None) is not None or runtime_checkable(
        type("_Dummy", (Protocol,), {})  # type: ignore[arg-type]
    )
    # The canonical check: runtime_checkable sets _is_runtime_protocol
    assert getattr(cls, "_is_runtime_protocol", False), f"{name} is not @runtime_checkable"


# ---------------------------------------------------------------------------
# 3. Key methods exist on each Protocol
# ---------------------------------------------------------------------------

EXPECTED_METHODS: dict[str, list[str]] = {
    "BootstrapPort": ["preview", "confirm", "execute"],
    "RescuePort": ["analyze", "plan", "execute"],
    "DiscoveryPort": [
        "start_session",
        "detect_persona",
        "answer_question",
        "confirm_playback",
        "complete",
    ],
    "ArtifactRendererPort": ["render_prd", "render_ddd", "render_architecture"],
    "FileWriterPort": ["write_file"],
    "FitnessGenerationPort": ["generate"],
    "TicketGenerationPort": ["generate"],
    "ConfigGenerationPort": ["generate"],
    "ToolDetectionPort": ["detect", "scan_conflicts"],
    "GateRunnerProtocol": ["run"],
    "QualityGatePort": ["check"],
    "KnowledgeLookupPort": ["lookup", "list_tools", "list_versions", "list_topics"],
    "DocHealthPort": ["check", "check_knowledge"],
    "DocReviewPort": ["reviewable_docs", "mark_reviewed", "mark_all_reviewed"],
    "TicketHealthPort": ["report"],
    "PersonaPort": ["list_personas", "generate"],
    "DriftDetectionPort": ["detect"],
}


@pytest.mark.parametrize(
    ("port_name", "method_name"),
    [(port, method) for port, methods in EXPECTED_METHODS.items() for method in methods],
)
def test_method_exists(port_name, method_name):
    """Each Protocol declares its required methods."""
    from src.application import ports

    cls = getattr(ports, port_name)
    assert hasattr(cls, method_name), f"{port_name} is missing method '{method_name}'"
    member = getattr(cls, method_name)
    assert callable(member), f"{port_name}.{method_name} is not callable"


# ---------------------------------------------------------------------------
# 4. No external dependencies in port files
# ---------------------------------------------------------------------------

FORBIDDEN_IMPORTS = {"typer", "mcp", "requests", "sqlalchemy", "fastapi"}

PORT_DIR = Path(__file__).resolve().parents[3] / "src" / "application" / "ports"


def _collect_imports(filepath: Path) -> set[str]:
    """Parse a Python file and return top-level imported module names."""
    source = filepath.read_text()
    tree = ast.parse(source, filename=str(filepath))
    modules: set[str] = set()
    for node in ast.walk(tree):
        if isinstance(node, ast.Import):
            for alias in node.names:
                modules.add(alias.name.split(".")[0])
        elif isinstance(node, ast.ImportFrom) and node.module:
            modules.add(node.module.split(".")[0])
    return modules


def test_no_forbidden_imports_in_port_files():
    """Port files must not import external framework packages."""
    port_files = list(PORT_DIR.glob("*.py"))
    assert len(port_files) > 1, "Expected port .py files in src/application/ports/"

    violations: list[str] = []
    for pf in port_files:
        if pf.name == "__init__.py":
            continue
        imported = _collect_imports(pf)
        bad = imported & FORBIDDEN_IMPORTS
        if bad:
            violations.append(f"{pf.name}: {bad}")

    assert violations == [], f"Forbidden imports found: {violations}"


# ---------------------------------------------------------------------------
# 5. BootstrapPort has: preview, confirm, execute
# ---------------------------------------------------------------------------


def test_bootstrap_port_method_signatures():
    """BootstrapPort methods have correct parameter names."""
    from src.application.ports import BootstrapPort

    sig_preview = inspect.signature(BootstrapPort.preview)
    assert "project_dir" in sig_preview.parameters

    sig_confirm = inspect.signature(BootstrapPort.confirm)
    assert "session_id" in sig_confirm.parameters

    sig_execute = inspect.signature(BootstrapPort.execute)
    assert "session_id" in sig_execute.parameters


# ---------------------------------------------------------------------------
# 6. DiscoveryPort has: start_session, detect_persona, answer_question,
#    confirm_playback, complete
# ---------------------------------------------------------------------------


def test_discovery_port_method_signatures():
    """DiscoveryPort methods have correct parameter names."""
    from src.application.ports import DiscoveryPort

    sig = inspect.signature(DiscoveryPort.start_session)
    assert "readme_content" in sig.parameters

    sig = inspect.signature(DiscoveryPort.detect_persona)
    assert "session_id" in sig.parameters
    assert "choice" in sig.parameters

    sig = inspect.signature(DiscoveryPort.answer_question)
    assert "session_id" in sig.parameters
    assert "answer" in sig.parameters

    sig = inspect.signature(DiscoveryPort.confirm_playback)
    assert "session_id" in sig.parameters
    assert "confirmed" in sig.parameters

    sig = inspect.signature(DiscoveryPort.complete)
    assert "session_id" in sig.parameters


# ---------------------------------------------------------------------------
# 7. KnowledgeLookupPort has: lookup, list_tools, list_versions, list_topics
# ---------------------------------------------------------------------------


def test_knowledge_lookup_port_method_signatures():
    """KnowledgeLookupPort methods have correct parameter names."""
    from src.application.ports import KnowledgeLookupPort

    sig = inspect.signature(KnowledgeLookupPort.lookup)
    assert "category" in sig.parameters
    assert "topic" in sig.parameters
    assert "version" in sig.parameters

    sig = inspect.signature(KnowledgeLookupPort.list_tools)
    # Only self — no other params
    non_self = [p for p in sig.parameters if p != "self"]
    assert non_self == []

    sig = inspect.signature(KnowledgeLookupPort.list_versions)
    assert "tool" in sig.parameters

    sig = inspect.signature(KnowledgeLookupPort.list_topics)
    assert "category" in sig.parameters
    assert "tool" in sig.parameters


# ---------------------------------------------------------------------------
# 8. ToolDetectionPort has: detect, scan_conflicts
# ---------------------------------------------------------------------------


def test_tool_detection_port_method_signatures():
    """ToolDetectionPort methods have correct parameter names."""
    from src.application.ports import ToolDetectionPort

    sig = inspect.signature(ToolDetectionPort.detect)
    assert "project_dir" in sig.parameters

    sig = inspect.signature(ToolDetectionPort.scan_conflicts)
    assert "project_dir" in sig.parameters


# ---------------------------------------------------------------------------
# 9. DocReviewPort has: reviewable_docs, mark_reviewed, mark_all_reviewed
# ---------------------------------------------------------------------------


def test_doc_review_port_method_signatures():
    """DocReviewPort methods have correct parameter names."""
    from src.application.ports import DocReviewPort

    sig = inspect.signature(DocReviewPort.reviewable_docs)
    assert "project_dir" in sig.parameters

    sig = inspect.signature(DocReviewPort.mark_reviewed)
    assert "doc_path" in sig.parameters
    assert "project_dir" in sig.parameters

    sig = inspect.signature(DocReviewPort.mark_all_reviewed)
    assert "project_dir" in sig.parameters
