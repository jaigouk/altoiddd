"""Tests for LLM port contracts: Protocol, enums, and value objects.

Verifies LLMProvider enum members, LLMConfig/LLMResponse immutability,
and LLMClient protocol compliance.
"""

from __future__ import annotations

import dataclasses

import pytest

from src.infrastructure.external.llm_client import (
    LLMClient,
    LLMConfig,
    LLMProvider,
    LLMResponse,
)


class TestLLMProviderEnum:
    def test_llm_provider_enum_has_four_members(self) -> None:
        assert len(LLMProvider) == 4

    def test_llm_provider_anthropic_value(self) -> None:
        assert LLMProvider.ANTHROPIC.value == "anthropic"

    def test_llm_provider_ollama_value(self) -> None:
        assert LLMProvider.OLLAMA.value == "ollama"

    def test_llm_provider_vertexai_value(self) -> None:
        assert LLMProvider.VERTEXAI.value == "vertexai"

    def test_llm_provider_none_value(self) -> None:
        assert LLMProvider.NONE.value == "none"


class TestLLMConfig:
    def test_llm_config_immutable(self) -> None:
        config = LLMConfig()
        with pytest.raises(dataclasses.FrozenInstanceError):
            config.provider = LLMProvider.ANTHROPIC  # type: ignore[misc]

    def test_llm_config_default_provider_none(self) -> None:
        config = LLMConfig()
        assert config.provider == LLMProvider.NONE

    def test_llm_config_default_model_empty(self) -> None:
        config = LLMConfig()
        assert config.model == ""

    def test_llm_config_default_api_key_empty(self) -> None:
        config = LLMConfig()
        assert config.api_key == ""

    def test_llm_config_default_timeout(self) -> None:
        config = LLMConfig()
        assert config.timeout == 30.0

    def test_llm_config_custom_values(self) -> None:
        config = LLMConfig(
            provider=LLMProvider.ANTHROPIC,
            model="claude-sonnet-4-20250514",
            api_key="sk-test",
            timeout=60.0,
        )
        assert config.provider == LLMProvider.ANTHROPIC
        assert config.model == "claude-sonnet-4-20250514"
        assert config.api_key == "sk-test"
        assert config.timeout == 60.0


class TestLLMResponse:
    def test_llm_response_immutable(self) -> None:
        resp = LLMResponse(content="hi", model_used="m", usage_tokens=10)
        with pytest.raises(dataclasses.FrozenInstanceError):
            resp.content = "changed"  # type: ignore[misc]

    def test_llm_response_fields(self) -> None:
        resp = LLMResponse(content="hello", model_used="claude", usage_tokens=42)
        assert resp.content == "hello"
        assert resp.model_used == "claude"
        assert resp.usage_tokens == 42


class TestLLMConfigRepr:
    def test_repr_masks_api_key_when_set(self) -> None:
        config = LLMConfig(
            provider=LLMProvider.ANTHROPIC,
            model="claude-sonnet-4-20250514",
            api_key="sk-secret-key-12345",
        )
        r = repr(config)
        assert "sk-secret-key-12345" not in r
        assert "***" in r

    def test_repr_shows_empty_when_no_api_key(self) -> None:
        config = LLMConfig(provider=LLMProvider.NONE)
        r = repr(config)
        assert "***" not in r

    def test_repr_includes_provider_and_model(self) -> None:
        config = LLMConfig(
            provider=LLMProvider.ANTHROPIC,
            model="claude-sonnet-4-20250514",
            api_key="sk-test",
        )
        r = repr(config)
        assert "ANTHROPIC" in r
        assert "claude-sonnet-4-20250514" in r


class TestLLMClientProtocol:
    def test_llm_client_is_runtime_checkable(self) -> None:
        assert isinstance(LLMClient, type(LLMClient))
        # Protocol must be runtime_checkable
        assert hasattr(LLMClient, "__protocol_attrs__") or hasattr(
            LLMClient, "__abstractmethods__"
        )
