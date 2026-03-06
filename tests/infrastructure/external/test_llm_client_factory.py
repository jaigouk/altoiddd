"""Tests for LLMClientFactory.

Verifies factory routing: Anthropic when configured, NoOp for graceful degradation.
"""

from __future__ import annotations

from unittest.mock import patch

from src.infrastructure.external.llm_client import LLMConfig, LLMProvider
from src.infrastructure.external.llm_client_factory import LLMClientFactory
from src.infrastructure.external.noop_llm_client import NoOpLLMClient


class TestLLMClientFactoryRouting:
    def test_factory_creates_anthropic_when_configured(self) -> None:
        """ANTHROPIC provider + api_key + SDK installed → AnthropicLLMClient."""
        from src.infrastructure.external.anthropic_llm_client import (
            AnthropicLLMClient,
        )

        config = LLMConfig(
            provider=LLMProvider.ANTHROPIC,
            model="claude-sonnet-4-20250514",
            api_key="sk-test-key",
        )
        client = LLMClientFactory.create(config)
        assert isinstance(client, AnthropicLLMClient)

    def test_factory_creates_noop_for_none_provider(self) -> None:
        """NONE provider → NoOpLLMClient."""
        config = LLMConfig(provider=LLMProvider.NONE)
        client = LLMClientFactory.create(config)
        assert isinstance(client, NoOpLLMClient)

    def test_factory_creates_noop_when_no_api_key(self) -> None:
        """ANTHROPIC provider but empty api_key → graceful degradation to NoOp."""
        config = LLMConfig(
            provider=LLMProvider.ANTHROPIC,
            model="claude-sonnet-4-20250514",
            api_key="",
        )
        client = LLMClientFactory.create(config)
        assert isinstance(client, NoOpLLMClient)

    def test_factory_creates_noop_when_anthropic_not_installed(self) -> None:
        """ANTHROPIC provider + api_key but SDK missing → graceful degradation."""
        config = LLMConfig(
            provider=LLMProvider.ANTHROPIC,
            model="claude-sonnet-4-20250514",
            api_key="sk-test-key",
        )
        with patch.dict("sys.modules", {"anthropic": None}):
            client = LLMClientFactory.create(config)
        assert isinstance(client, NoOpLLMClient)

    def test_factory_defaults_to_none_when_no_config(self) -> None:
        """Default LLMConfig (no args) → NoOpLLMClient."""
        config = LLMConfig()
        client = LLMClientFactory.create(config)
        assert isinstance(client, NoOpLLMClient)

    def test_factory_creates_noop_for_unsupported_provider(self) -> None:
        """Unsupported provider (OLLAMA, VERTEXAI) → NoOpLLMClient fallback."""
        config = LLMConfig(provider=LLMProvider.OLLAMA, api_key="key")
        client = LLMClientFactory.create(config)
        assert isinstance(client, NoOpLLMClient)
