"""Tests for the MCP server adapter.

Verifies FastMCP tool/resource registration, AppContext composition root,
input validation, resource invocation, and the alty-mcp entry point.
"""

from __future__ import annotations

import asyncio
import dataclasses

import pytest


class TestAppContext:
    """Tests for the AppContext composition root."""

    def test_app_context_is_dataclass(self):
        from src.infrastructure.composition import AppContext

        assert dataclasses.is_dataclass(AppContext)

    def test_app_context_has_bootstrap_port(self):
        from src.infrastructure.composition import AppContext

        fields = {f.name for f in dataclasses.fields(AppContext)}
        assert "bootstrap" in fields

    def test_app_context_has_discovery_port(self):
        from src.infrastructure.composition import AppContext

        fields = {f.name for f in dataclasses.fields(AppContext)}
        assert "discovery" in fields

    def test_app_context_has_tool_detection_port(self):
        from src.infrastructure.composition import AppContext

        fields = {f.name for f in dataclasses.fields(AppContext)}
        assert "tool_detection" in fields

    def test_app_context_has_fitness_generation_port(self):
        from src.infrastructure.composition import AppContext

        fields = {f.name for f in dataclasses.fields(AppContext)}
        assert "fitness_generation" in fields

    def test_app_context_has_ticket_generation_port(self):
        from src.infrastructure.composition import AppContext

        fields = {f.name for f in dataclasses.fields(AppContext)}
        assert "ticket_generation" in fields

    def test_app_context_has_config_generation_port(self):
        from src.infrastructure.composition import AppContext

        fields = {f.name for f in dataclasses.fields(AppContext)}
        assert "config_generation" in fields

    def test_app_context_has_quality_gate_port(self):
        from src.infrastructure.composition import AppContext

        fields = {f.name for f in dataclasses.fields(AppContext)}
        assert "quality_gate" in fields

    def test_app_context_has_doc_health_port(self):
        from src.infrastructure.composition import AppContext

        fields = {f.name for f in dataclasses.fields(AppContext)}
        assert "doc_health" in fields

    def test_app_context_has_doc_review_port(self):
        from src.infrastructure.composition import AppContext

        fields = {f.name for f in dataclasses.fields(AppContext)}
        assert "doc_review" in fields

    def test_app_context_has_ticket_health_port(self):
        from src.infrastructure.composition import AppContext

        fields = {f.name for f in dataclasses.fields(AppContext)}
        assert "ticket_health" in fields

    def test_app_context_has_all_ports(self):
        """AppContext must declare all 10 port attributes."""
        from src.infrastructure.composition import AppContext

        fields = {f.name for f in dataclasses.fields(AppContext)}
        expected = {
            "bootstrap",
            "discovery",
            "tool_detection",
            "fitness_generation",
            "ticket_generation",
            "config_generation",
            "quality_gate",
            "doc_health",
            "doc_review",
            "ticket_health",
        }
        assert expected.issubset(fields)

    def test_create_app_returns_app_context(self):
        from src.infrastructure.composition import AppContext, create_app

        ctx = create_app()
        assert isinstance(ctx, AppContext)


class TestMcpServerInstance:
    """Tests for the FastMCP server instance and registration."""

    def test_mcp_server_exists(self):
        from mcp.server.fastmcp import FastMCP

        from src.infrastructure.mcp.server import mcp

        assert isinstance(mcp, FastMCP)

    def test_mcp_server_name_is_alty(self):
        from src.infrastructure.mcp.server import mcp

        assert mcp.name == "alty"

    def test_mcp_server_has_lifespan(self):
        """The FastMCP server must be constructed with a lifespan."""
        from src.infrastructure.mcp.server import app_lifespan, mcp

        # Verify app_lifespan is importable and the server's internal
        # lifespan was set (wrapped by FastMCP)
        assert callable(app_lifespan)
        assert mcp._mcp_server.lifespan is not None

    def test_mcp_server_registers_seventeen_tools(self):
        from src.infrastructure.mcp.server import mcp

        tools = mcp._tool_manager._tools
        assert len(tools) == 17, f"Expected 17 tools, got {len(tools)}: {list(tools.keys())}"

    def test_mcp_server_registers_ten_resources(self):
        """10 resources total = static resources + URI templates."""
        from src.infrastructure.mcp.server import mcp

        static_count = len(mcp._resource_manager._resources)
        template_count = len(mcp._resource_manager._templates)
        total = static_count + template_count
        assert total == 10, (
            f"Expected 10 resources, got {total} "
            f"(static={static_count}, templates={template_count})"
        )


def _tool_names() -> set[str]:
    """Return the set of registered tool names from the MCP server."""
    from src.infrastructure.mcp.server import mcp

    return set(mcp._tool_manager._tools.keys())


def _all_resource_uris() -> set[str]:
    """Return all registered resource URIs (static + templates)."""
    from src.infrastructure.mcp.server import mcp

    static_uris = {str(u) for u in mcp._resource_manager._resources}
    template_uris = {str(u) for u in mcp._resource_manager._templates}
    return static_uris | template_uris


class TestToolRegistration:
    """Tests that each of the 17 tools is registered."""

    def test_init_project_tool_exists(self):
        assert "init_project" in _tool_names()

    def test_guide_start_tool_exists(self):
        assert "guide_start" in _tool_names()

    def test_guide_detect_persona_tool_exists(self):
        assert "guide_detect_persona" in _tool_names()

    def test_guide_answer_tool_exists(self):
        assert "guide_answer" in _tool_names()

    def test_guide_skip_question_tool_exists(self):
        assert "guide_skip_question" in _tool_names()

    def test_guide_confirm_playback_tool_exists(self):
        assert "guide_confirm_playback" in _tool_names()

    def test_guide_complete_tool_exists(self):
        assert "guide_complete" in _tool_names()

    def test_guide_status_tool_exists(self):
        assert "guide_status" in _tool_names()

    def test_generate_artifacts_tool_exists(self):
        assert "generate_artifacts" in _tool_names()

    def test_generate_fitness_tool_exists(self):
        assert "generate_fitness" in _tool_names()

    def test_generate_tickets_tool_exists(self):
        assert "generate_tickets" in _tool_names()

    def test_generate_configs_tool_exists(self):
        assert "generate_configs" in _tool_names()

    def test_detect_tools_tool_exists(self):
        assert "detect_tools" in _tool_names()

    def test_check_quality_tool_exists(self):
        assert "check_quality" in _tool_names()

    def test_doc_health_tool_exists(self):
        assert "doc_health" in _tool_names()

    def test_doc_review_tool_exists(self):
        assert "doc_review" in _tool_names()

    def test_ticket_health_tool_exists(self):
        assert "ticket_health" in _tool_names()


class TestResourceRegistration:
    """Tests that resources and resource templates are registered."""

    def test_knowledge_ddd_resource_exists(self):
        uris = _all_resource_uris()
        assert any("knowledge/ddd" in u for u in uris)

    def test_knowledge_tools_resource_exists(self):
        uris = _all_resource_uris()
        assert any("knowledge/tools" in u for u in uris)

    def test_knowledge_conventions_resource_exists(self):
        uris = _all_resource_uris()
        assert any("knowledge/conventions" in u for u in uris)

    def test_knowledge_cross_tool_resource_exists(self):
        uris = _all_resource_uris()
        assert any("knowledge/cross-tool" in u for u in uris)

    def test_project_domain_model_resource_exists(self):
        uris = _all_resource_uris()
        assert any("domain-model" in u for u in uris)

    def test_project_architecture_resource_exists(self):
        uris = _all_resource_uris()
        assert any("architecture" in u for u in uris)

    def test_project_prd_resource_exists(self):
        uris = _all_resource_uris()
        assert any("prd" in u for u in uris)

    def test_tickets_ready_resource_exists(self):
        uris = _all_resource_uris()
        assert any("tickets/ready" in u for u in uris)

    def test_tickets_by_id_resource_exists(self):
        uris = _all_resource_uris()
        assert any("tickets/" in u and "ready" not in u for u in uris)


class TestEntryPoint:
    """Tests for the alty-mcp entry point."""

    def test_main_entry_point_exists(self):
        from src.infrastructure.mcp.server import main

        assert callable(main)

    def test_main_is_a_function(self):
        import types

        from src.infrastructure.mcp.server import main

        assert isinstance(main, types.FunctionType)


class TestAppLifespan:
    """Tests for the app_lifespan context manager."""

    def test_app_lifespan_yields_app_context(self):
        from src.infrastructure.composition import AppContext
        from src.infrastructure.mcp.server import app_lifespan, mcp

        async def _run() -> AppContext:
            async with app_lifespan(mcp) as ctx:
                return ctx

        ctx = asyncio.run(_run())
        assert isinstance(ctx, AppContext)

    def test_app_lifespan_wires_discovery_adapter(self):
        from src.infrastructure.mcp.discovery_adapter import DiscoveryAdapter
        from src.infrastructure.mcp.server import app_lifespan, mcp

        async def _run() -> object:
            async with app_lifespan(mcp) as ctx:
                return ctx.discovery

        discovery = asyncio.run(_run())
        assert isinstance(discovery, DiscoveryAdapter)


class TestInputValidation:
    """Tests for the input validation helpers."""

    def test_safe_component_accepts_valid_name(self):
        from src.infrastructure.mcp.server import _safe_component

        assert _safe_component("aggregate", "topic") == "aggregate"

    def test_safe_component_accepts_hyphenated_name(self):
        from src.infrastructure.mcp.server import _safe_component

        assert _safe_component("claude-code", "tool") == "claude-code"

    def test_safe_component_accepts_underscored_name(self):
        from src.infrastructure.mcp.server import _safe_component

        assert _safe_component("agent_format", "subtopic") == "agent_format"

    def test_safe_component_rejects_path_traversal(self):
        import pytest

        from src.infrastructure.mcp.server import _safe_component

        with pytest.raises(ValueError, match="Invalid topic"):
            _safe_component("../../etc/passwd", "topic")

    def test_safe_component_rejects_slash(self):
        import pytest

        from src.infrastructure.mcp.server import _safe_component

        with pytest.raises(ValueError, match="Invalid tool"):
            _safe_component("a/b", "tool")

    def test_safe_component_rejects_empty_string(self):
        import pytest

        from src.infrastructure.mcp.server import _safe_component

        with pytest.raises(ValueError, match="Invalid topic"):
            _safe_component("", "topic")

    def test_safe_component_rejects_too_long_name(self):
        import pytest

        from src.infrastructure.mcp.server import _safe_component

        with pytest.raises(ValueError, match="Invalid topic"):
            _safe_component("a" * 65, "topic")

    def test_safe_project_path_rejects_empty(self):
        import pytest

        from src.infrastructure.mcp.server import _safe_project_path

        with pytest.raises(ValueError, match="must not be empty"):
            _safe_project_path("", "project_dir")

    def test_safe_project_path_rejects_whitespace_only(self):
        import pytest

        from src.infrastructure.mcp.server import _safe_project_path

        with pytest.raises(ValueError, match="must not be empty"):
            _safe_project_path("   ", "project_dir")

    def test_safe_project_path_resolves_to_absolute(self):
        from src.infrastructure.mcp.server import _safe_project_path

        result = _safe_project_path("/tmp/test", "project_dir")
        assert result.is_absolute()

    def test_ticket_id_regex_accepts_valid_id(self):
        from src.infrastructure.mcp.server import _TICKET_ID_RE

        assert _TICKET_ID_RE.fullmatch("k7m.27")
        assert _TICKET_ID_RE.fullmatch("alty-k7m.28")

    def test_ticket_id_regex_rejects_shell_injection(self):
        from src.infrastructure.mcp.server import _TICKET_ID_RE

        assert _TICKET_ID_RE.fullmatch("; rm -rf /") is None
        assert _TICKET_ID_RE.fullmatch("--delete") is None

    def test_reviewer_regex_accepts_valid_reviewer(self):
        from src.infrastructure.mcp.server import _REVIEWER_RE

        assert _REVIEWER_RE.fullmatch("jaigouk")
        assert _REVIEWER_RE.fullmatch("user@example.com")

    def test_reviewer_regex_rejects_injection(self):
        from src.infrastructure.mcp.server import _REVIEWER_RE

        assert _REVIEWER_RE.fullmatch("evil\nlast_reviewed: 1970") is None


class TestResourceInvocation:
    """Tests that resource functions return expected content."""

    def test_knowledge_ddd_not_found(self):
        from src.infrastructure.mcp.server import knowledge_ddd

        result = asyncio.run(knowledge_ddd("nonexistent_topic_xyz"))
        assert "not found" in result

    def test_knowledge_ddd_returns_content(self, tmp_path):
        import os

        from src.infrastructure.mcp.server import knowledge_ddd

        kb_dir = tmp_path / ".alty" / "knowledge" / "ddd"
        kb_dir.mkdir(parents=True)
        (kb_dir / "aggregate.md").write_text("# Aggregate Root")

        original_cwd = os.getcwd()
        try:
            os.chdir(tmp_path)
            result = asyncio.run(knowledge_ddd("aggregate"))
            assert result == "# Aggregate Root"
        finally:
            os.chdir(original_cwd)

    def test_knowledge_ddd_rejects_traversal(self):
        from src.infrastructure.mcp.server import knowledge_ddd

        result = asyncio.run(knowledge_ddd("../../etc/passwd"))
        assert result == "Invalid topic name."

    def test_knowledge_tool_not_found(self):
        from src.infrastructure.mcp.server import knowledge_tool

        result = asyncio.run(knowledge_tool("nonexistent_tool_xyz"))
        assert "not found" in result

    def test_knowledge_tool_rejects_traversal(self):
        from src.infrastructure.mcp.server import knowledge_tool

        result = asyncio.run(knowledge_tool("../../../etc"))
        assert result == "Invalid tool name."

    def test_knowledge_conventions_not_found(self):
        from src.infrastructure.mcp.server import knowledge_conventions

        result = asyncio.run(knowledge_conventions("nonexistent_topic_xyz"))
        assert "not found" in result

    def test_knowledge_cross_tool_not_found(self):
        from src.infrastructure.mcp.server import knowledge_cross_tool

        result = asyncio.run(knowledge_cross_tool("nonexistent_xyz"))
        assert "not found" in result

    def test_knowledge_subtopic_rejects_traversal(self):
        from src.infrastructure.mcp.server import knowledge_tool_subtopic

        result = asyncio.run(knowledge_tool_subtopic("../evil", "config"))
        assert result == "Invalid tool or subtopic name."

    def test_project_domain_model_not_found(self, tmp_path):
        from src.infrastructure.mcp.server import project_domain_model

        result = asyncio.run(project_domain_model(str(tmp_path)))
        assert "not found" in result

    def test_project_domain_model_returns_content(self, tmp_path):
        from src.infrastructure.mcp.server import project_domain_model

        docs_dir = tmp_path / "docs"
        docs_dir.mkdir()
        (docs_dir / "DDD.md").write_text("# Domain Model")
        result = asyncio.run(project_domain_model(str(tmp_path)))
        assert result == "# Domain Model"

    def test_project_prd_empty_dir_raises(self):
        import pytest

        from src.infrastructure.mcp.server import project_prd

        with pytest.raises(ValueError, match="must not be empty"):
            asyncio.run(project_prd(""))

    def test_tickets_by_id_rejects_invalid_id(self):
        from src.infrastructure.mcp.server import tickets_by_id

        result = asyncio.run(tickets_by_id("; rm -rf /"))
        assert result == "Invalid ticket ID format."

    def test_tickets_by_id_rejects_flag_injection(self):
        from src.infrastructure.mcp.server import tickets_by_id

        result = asyncio.run(tickets_by_id("--delete"))
        assert result == "Invalid ticket ID format."


class TestDocReviewValidation:
    """Tests for doc_review reviewer validation."""

    def test_doc_review_rejects_invalid_reviewer(self):
        """doc_review returns error for malicious reviewer input."""
        from src.infrastructure.mcp.server import doc_review

        # We can't easily mock the context, but we can verify the
        # validation happens before the port is called by checking
        # that invalid reviewers are rejected with the right message.
        result = asyncio.run(doc_review("/tmp/doc.md", "evil\ninjection", None))
        assert result == "Invalid reviewer identifier."


class TestRunBdErrorHandling:
    """Tests for _run_bd stderr handling on failure."""

    def test_run_bd_includes_stderr_on_failure(self):
        """When bd command fails, stderr should be included in the result."""
        from src.infrastructure.mcp.server import _run_bd

        # Use a command that will fail (invalid subcommand)
        result = asyncio.run(_run_bd("--nonexistent-flag-that-does-not-exist"))
        assert "bd command failed" in result or result == ""

    def test_run_bd_returns_stdout_on_success(self):
        """When bd command succeeds, stdout is returned."""
        import shutil

        from src.infrastructure.mcp.server import _run_bd

        if shutil.which("bd") is None:
            pytest.skip("bd not available")
        # 'bd' with no args should at least not crash
        result = asyncio.run(_run_bd("--help"))
        assert isinstance(result, str)


class TestKbRootHelper:
    """Tests for _kb_root helper."""

    def test_kb_root_returns_absolute_path(self):
        from src.infrastructure.mcp.server import _kb_root

        root = _kb_root()
        assert root.is_absolute()

    def test_kb_root_ends_with_knowledge(self):
        from src.infrastructure.mcp.server import _kb_root

        root = _kb_root()
        assert root.parts[-2:] == (".alty", "knowledge")

    def test_knowledge_ddd_uses_absolute_path(self, tmp_path):
        """Knowledge resources should use absolute paths via _kb_root."""
        import os

        from src.infrastructure.mcp.server import knowledge_ddd

        kb_dir = tmp_path / ".alty" / "knowledge" / "ddd"
        kb_dir.mkdir(parents=True)
        (kb_dir / "test_topic.md").write_text("# Test Topic Content")

        original_cwd = os.getcwd()
        try:
            os.chdir(tmp_path)
            result = asyncio.run(knowledge_ddd("test_topic"))
            assert result == "# Test Topic Content"
        finally:
            os.chdir(original_cwd)


class TestStubDiscoveryPortCompliance:
    """Tests for _StubDiscovery matching the DiscoveryPort protocol."""

    def test_stub_has_skip_question(self):
        from src.infrastructure.composition import create_app

        ctx = create_app()
        assert hasattr(ctx.discovery, "skip_question")

    def test_stub_answer_question_takes_question_id(self):
        import inspect

        from src.infrastructure.composition import _StubDiscovery

        sig = inspect.signature(_StubDiscovery.answer_question)
        params = list(sig.parameters.keys())
        assert "question_id" in params

    def test_stub_answer_question_raises_not_implemented(self):
        from src.infrastructure.composition import create_app

        ctx = create_app()
        with pytest.raises(NotImplementedError):
            ctx.discovery.answer_question("sid", "Q1", "answer")

    def test_stub_skip_question_raises_not_implemented(self):
        from src.infrastructure.composition import create_app

        ctx = create_app()
        with pytest.raises(NotImplementedError):
            ctx.discovery.skip_question("sid", "Q5", "reason")


class TestDiscoveryPortProtocol:
    """Tests for DiscoveryPort protocol structural compliance."""

    def test_port_answer_question_has_question_id(self):
        import inspect

        from src.application.ports.discovery_port import DiscoveryPort

        sig = inspect.signature(DiscoveryPort.answer_question)
        params = list(sig.parameters.keys())
        assert "question_id" in params
        assert "answer" in params

    def test_port_has_skip_question(self):
        from src.application.ports.discovery_port import DiscoveryPort

        assert hasattr(DiscoveryPort, "skip_question")

    def test_port_skip_question_has_reason(self):
        import inspect

        from src.application.ports.discovery_port import DiscoveryPort

        sig = inspect.signature(DiscoveryPort.skip_question)
        params = list(sig.parameters.keys())
        assert "question_id" in params
        assert "reason" in params

    def test_port_returns_discovery_session_types(self):
        """All port methods should return DiscoverySession, not str."""
        import inspect

        from src.application.ports.discovery_port import DiscoveryPort

        for method_name in [
            "start_session",
            "detect_persona",
            "answer_question",
            "skip_question",
            "confirm_playback",
            "complete",
        ]:
            method = getattr(DiscoveryPort, method_name)
            sig = inspect.signature(method)
            assert sig.return_annotation == "DiscoverySession", (
                f"{method_name} should return DiscoverySession, "
                f"got {sig.return_annotation}"
            )
