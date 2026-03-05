"""Tests for the composition root (create_app and AppContext).

Verifies that create_app() wires real adapters for implemented ports
and stubs for ports not yet implemented. Covers field count, Protocol
compliance, constructor injection, and stub behaviour.
"""

from __future__ import annotations

import dataclasses

import pytest

from src.infrastructure.composition import AppContext, create_app

# ---------------------------------------------------------------------------
# AppContext structure
# ---------------------------------------------------------------------------


class TestAppContextStructure:
    """Verify AppContext dataclass shape and field inventory."""

    def test_app_context_is_dataclass(self):
        assert dataclasses.is_dataclass(AppContext)

    def test_app_context_has_13_fields(self):
        """11 original ports + file_writer + artifact_renderer = 13."""
        fields = {f.name for f in dataclasses.fields(AppContext)}
        assert len(fields) == 13

    def test_app_context_has_file_writer_field(self):
        fields = {f.name for f in dataclasses.fields(AppContext)}
        assert "file_writer" in fields

    def test_app_context_has_artifact_renderer_field(self):
        fields = {f.name for f in dataclasses.fields(AppContext)}
        assert "artifact_renderer" in fields

    def test_app_context_has_all_expected_fields(self):
        """Every port that AppContext should expose is present."""
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
            "spike_follow_up",
            "file_writer",
            "artifact_renderer",
        }
        assert expected == fields


# ---------------------------------------------------------------------------
# Real adapter wiring
# ---------------------------------------------------------------------------


class TestRealAdapterWiring:
    """Verify create_app() returns real adapters for implemented ports."""

    def test_create_app_returns_app_context(self):
        ctx = create_app()
        assert isinstance(ctx, AppContext)

    def test_discovery_is_in_memory_adapter(self):
        from src.infrastructure.session.in_memory_discovery_adapter import (
            InMemoryDiscoveryAdapter,
        )

        ctx = create_app()
        assert isinstance(ctx.discovery, InMemoryDiscoveryAdapter)

    def test_quality_gate_is_real_handler(self):
        from src.application.commands.quality_gate_handler import QualityGateHandler

        ctx = create_app()
        assert isinstance(ctx.quality_gate, QualityGateHandler)

    def test_file_writer_is_filesystem_writer(self):
        from src.infrastructure.persistence.filesystem_file_writer import (
            FilesystemFileWriter,
        )

        ctx = create_app()
        assert isinstance(ctx.file_writer, FilesystemFileWriter)

    def test_artifact_renderer_is_markdown_renderer(self):
        from src.infrastructure.persistence.markdown_artifact_renderer import (
            MarkdownArtifactRenderer,
        )

        ctx = create_app()
        assert isinstance(ctx.artifact_renderer, MarkdownArtifactRenderer)


# ---------------------------------------------------------------------------
# Protocol compliance
# ---------------------------------------------------------------------------


class TestProtocolCompliance:
    """Verify each real adapter satisfies its Protocol via isinstance."""

    def test_discovery_satisfies_protocol(self):
        from src.application.ports.discovery_port import DiscoveryPort

        ctx = create_app()
        assert isinstance(ctx.discovery, DiscoveryPort)

    def test_quality_gate_satisfies_protocol(self):
        from src.application.ports.quality_gate_port import QualityGatePort

        ctx = create_app()
        assert isinstance(ctx.quality_gate, QualityGatePort)

    def test_file_writer_satisfies_protocol(self):
        from src.application.ports.file_writer_port import FileWriterPort

        ctx = create_app()
        assert isinstance(ctx.file_writer, FileWriterPort)

    def test_artifact_renderer_satisfies_protocol(self):
        from src.application.ports.artifact_generation_port import ArtifactRendererPort

        ctx = create_app()
        assert isinstance(ctx.artifact_renderer, ArtifactRendererPort)


# ---------------------------------------------------------------------------
# Constructor injection / internal wiring
# ---------------------------------------------------------------------------


class TestConstructorInjection:
    """Verify adapters receive correct dependencies internally."""

    def test_quality_gate_handler_has_subprocess_runner(self):
        """QualityGateHandler must be injected with SubprocessGateRunner."""
        from src.application.commands.quality_gate_handler import QualityGateHandler
        from src.infrastructure.external.subprocess_gate_runner import (
            SubprocessGateRunner,
        )

        ctx = create_app()
        handler = ctx.quality_gate
        assert isinstance(handler, QualityGateHandler)
        assert isinstance(handler._runner, SubprocessGateRunner)

    def test_discovery_adapter_has_session_store(self):
        """InMemoryDiscoveryAdapter must be backed by a SessionStore."""
        from src.infrastructure.session.in_memory_discovery_adapter import (
            InMemoryDiscoveryAdapter,
        )
        from src.infrastructure.session.session_store import SessionStore

        ctx = create_app()
        adapter = ctx.discovery
        assert isinstance(adapter, InMemoryDiscoveryAdapter)
        assert isinstance(adapter._store, SessionStore)

    def test_session_store_has_default_ttl(self):
        """SessionStore should have the default 30-minute TTL."""
        from src.infrastructure.session.in_memory_discovery_adapter import (
            InMemoryDiscoveryAdapter,
        )

        ctx = create_app()
        adapter = ctx.discovery
        assert isinstance(adapter, InMemoryDiscoveryAdapter)
        assert adapter._store.ttl_seconds == 1800


# ---------------------------------------------------------------------------
# Remaining stubs
# ---------------------------------------------------------------------------


class TestRemainingStubs:
    """Verify ports not yet wired still raise NotImplementedError."""

    def test_bootstrap_raises(self):
        from pathlib import Path

        ctx = create_app()
        with pytest.raises(NotImplementedError):
            ctx.bootstrap.preview(Path("."))

    def test_tool_detection_raises(self):
        from pathlib import Path

        ctx = create_app()
        with pytest.raises(NotImplementedError):
            ctx.tool_detection.detect(Path("."))

    def test_fitness_generation_raises(self):
        from pathlib import Path
        from typing import Any, cast

        ctx = create_app()
        with pytest.raises(NotImplementedError):
            ctx.fitness_generation.generate(cast("Any", object()), "pkg", Path("."))

    def test_ticket_generation_raises(self):
        from pathlib import Path
        from typing import Any, cast

        ctx = create_app()
        with pytest.raises(NotImplementedError):
            ctx.ticket_generation.generate(cast("Any", object()), Path("."))

    def test_config_generation_raises(self):
        from pathlib import Path
        from typing import Any, cast

        ctx = create_app()
        with pytest.raises(NotImplementedError):
            ctx.config_generation.generate(cast("Any", object()), (), Path("."))

    def test_doc_health_raises(self):
        from pathlib import Path

        ctx = create_app()
        with pytest.raises(NotImplementedError):
            ctx.doc_health.check(Path("."))

    def test_doc_review_raises(self):
        from pathlib import Path

        ctx = create_app()
        with pytest.raises(NotImplementedError):
            ctx.doc_review.mark_reviewed(Path("."), "reviewer")

    def test_ticket_health_is_real_adapter(self):
        from src.infrastructure.external.beads_ticket_health_adapter import (
            BeadsTicketHealthAdapter,
        )

        ctx = create_app()
        assert isinstance(ctx.ticket_health, BeadsTicketHealthAdapter)

    def test_spike_follow_up_raises(self):
        from pathlib import Path

        ctx = create_app()
        with pytest.raises(NotImplementedError):
            ctx.spike_follow_up.audit("spike-1", Path("."))


# ---------------------------------------------------------------------------
# Edge cases
# ---------------------------------------------------------------------------


class TestEdgeCases:
    """Edge cases for composition root."""

    def test_multiple_create_app_calls_return_independent_contexts(self):
        """Each call should produce a fresh, independent AppContext."""
        ctx1 = create_app()
        ctx2 = create_app()
        assert ctx1 is not ctx2
        assert ctx1.discovery is not ctx2.discovery

    def test_create_app_no_arguments(self):
        """create_app() takes no arguments — pure factory."""
        # Should not raise
        ctx = create_app()
        assert ctx is not None

    def test_real_adapters_are_not_stubs(self):
        """None of the 4 real adapters should be a private _Stub* class."""
        ctx = create_app()
        for attr_name in ("discovery", "quality_gate", "file_writer", "artifact_renderer"):
            adapter = getattr(ctx, attr_name)
            assert not type(adapter).__name__.startswith("_Stub"), (
                f"{attr_name} is still a stub: {type(adapter).__name__}"
            )
