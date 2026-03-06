"""Tests for AnthropicLLMClient adapter.

All tests mock the Anthropic SDK — no real API calls.
Verifies structured output parsing, text completion, error handling, and timeout.
"""

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from src.domain.models.errors import LLMUnavailableError
from src.infrastructure.external.anthropic_llm_client import AnthropicLLMClient
from src.infrastructure.external.llm_client import LLMClient, LLMResponse


class TestAnthropicLLMClientProtocol:
    def test_implements_llm_client_protocol(self) -> None:
        client = AnthropicLLMClient(api_key="sk-test")
        assert isinstance(client, LLMClient)


class TestAnthropicStructuredOutput:
    async def test_structured_output_returns_parsed_json(self) -> None:
        client = AnthropicLLMClient(api_key="sk-test")
        mock_response = MagicMock()
        mock_response.content = [
            MagicMock(text='{"name": "test", "value": 42}')
        ]
        mock_response.model = "claude-sonnet-4-20250514"
        mock_response.usage = MagicMock(input_tokens=10, output_tokens=20)

        with patch.object(
            client, "_client", create=True
        ) as mock_client:
            mock_client.messages = AsyncMock()
            mock_client.messages.create = AsyncMock(return_value=mock_response)

            result = await client.structured_output(
                prompt="test prompt",
                schema={"type": "object", "properties": {"name": {"type": "string"}}},
            )

        assert isinstance(result, LLMResponse)
        assert '"name"' in result.content or "test" in result.content

    async def test_structured_output_includes_model_used(self) -> None:
        client = AnthropicLLMClient(api_key="sk-test", model="claude-sonnet-4-20250514")
        mock_response = MagicMock()
        mock_response.content = [MagicMock(text='{"key": "val"}')]
        mock_response.model = "claude-sonnet-4-20250514"
        mock_response.usage = MagicMock(input_tokens=5, output_tokens=10)

        with patch.object(
            client, "_client", create=True
        ) as mock_client:
            mock_client.messages = AsyncMock()
            mock_client.messages.create = AsyncMock(return_value=mock_response)

            result = await client.structured_output(
                prompt="test", schema={"type": "object"}
            )

        assert result.model_used == "claude-sonnet-4-20250514"


class TestAnthropicTextCompletion:
    async def test_text_completion_returns_string(self) -> None:
        client = AnthropicLLMClient(api_key="sk-test")
        mock_response = MagicMock()
        mock_response.content = [MagicMock(text="Hello, world!")]
        mock_response.model = "claude-sonnet-4-20250514"
        mock_response.usage = MagicMock(input_tokens=3, output_tokens=5)

        with patch.object(
            client, "_client", create=True
        ) as mock_client:
            mock_client.messages = AsyncMock()
            mock_client.messages.create = AsyncMock(return_value=mock_response)

            result = await client.text_completion(prompt="Say hello")

        assert isinstance(result, LLMResponse)
        assert result.content == "Hello, world!"

    async def test_text_completion_reports_token_usage(self) -> None:
        client = AnthropicLLMClient(api_key="sk-test")
        mock_response = MagicMock()
        mock_response.content = [MagicMock(text="response")]
        mock_response.model = "claude-sonnet-4-20250514"
        mock_response.usage = MagicMock(input_tokens=10, output_tokens=20)

        with patch.object(
            client, "_client", create=True
        ) as mock_client:
            mock_client.messages = AsyncMock()
            mock_client.messages.create = AsyncMock(return_value=mock_response)

            result = await client.text_completion(prompt="test")

        assert result.usage_tokens == 30


class TestAnthropicErrorHandling:
    @staticmethod
    def _make_api_connection_error() -> Exception:
        import anthropic
        import httpx

        req = httpx.Request("POST", "https://api.anthropic.com")
        return anthropic.APIConnectionError(request=req)

    @staticmethod
    def _make_status_error(status_code: int, cls_name: str) -> Exception:
        import anthropic
        import httpx

        req = httpx.Request("POST", "https://api.anthropic.com")
        resp = httpx.Response(status_code, request=req, text="error")
        cls = getattr(anthropic, cls_name)
        exc: Exception = cls(response=resp, body=None, message="error")
        return exc

    async def test_structured_output_error_raises_llm_unavailable(self) -> None:
        """APIConnectionError during structured_output → LLMUnavailableError."""
        client = AnthropicLLMClient(api_key="sk-test")

        with patch.object(
            client, "_client", create=True
        ) as mock_client:
            mock_client.messages = AsyncMock()
            mock_client.messages.create = AsyncMock(
                side_effect=self._make_api_connection_error()
            )

            with pytest.raises(LLMUnavailableError):
                await client.structured_output(
                    prompt="test", schema={"type": "object"}
                )

    async def test_handles_api_error_raises_llm_unavailable(self) -> None:
        """AuthenticationError from Anthropic SDK → LLMUnavailableError."""
        client = AnthropicLLMClient(api_key="bad-key")

        with patch.object(
            client, "_client", create=True
        ) as mock_client:
            mock_client.messages = AsyncMock()
            mock_client.messages.create = AsyncMock(
                side_effect=self._make_status_error(401, "AuthenticationError")
            )

            with pytest.raises(LLMUnavailableError):
                await client.text_completion(prompt="test")

    async def test_handles_rate_limit_raises_llm_unavailable(self) -> None:
        """RateLimitError → LLMUnavailableError."""
        client = AnthropicLLMClient(api_key="sk-test")

        with patch.object(
            client, "_client", create=True
        ) as mock_client:
            mock_client.messages = AsyncMock()
            mock_client.messages.create = AsyncMock(
                side_effect=self._make_status_error(429, "RateLimitError")
            )

            with pytest.raises(LLMUnavailableError):
                await client.text_completion(prompt="test")

    async def test_handles_network_error_raises_llm_unavailable(self) -> None:
        """APIConnectionError → LLMUnavailableError."""
        client = AnthropicLLMClient(api_key="sk-test")

        with patch.object(
            client, "_client", create=True
        ) as mock_client:
            mock_client.messages = AsyncMock()
            mock_client.messages.create = AsyncMock(
                side_effect=self._make_api_connection_error()
            )

            with pytest.raises(LLMUnavailableError):
                await client.text_completion(prompt="test")


class TestAnthropicTimeout:
    def test_timeout_configurable(self) -> None:
        client = AnthropicLLMClient(api_key="sk-test", timeout=120.0)
        assert client._timeout == 120.0

    def test_timeout_default(self) -> None:
        client = AnthropicLLMClient(api_key="sk-test")
        assert client._timeout == 30.0


class TestAnthropicInitFallback:
    def test_client_none_when_anthropic_not_installed(self) -> None:
        """If anthropic SDK is not importable, _client is None."""
        with patch.dict("sys.modules", {"anthropic": None}):
            client = AnthropicLLMClient(api_key="sk-test")
        assert client._client is None

    async def test_structured_output_raises_when_sdk_missing(self) -> None:
        """structured_output with _client=None → LLMUnavailableError, not AttributeError."""
        with patch.dict("sys.modules", {"anthropic": None}):
            client = AnthropicLLMClient(api_key="sk-test")
        assert client._client is None

        with pytest.raises(LLMUnavailableError, match="not installed"):
            await client.structured_output(prompt="test", schema={"type": "object"})

    async def test_text_completion_raises_when_sdk_missing(self) -> None:
        """text_completion with _client=None → LLMUnavailableError, not AttributeError."""
        with patch.dict("sys.modules", {"anthropic": None}):
            client = AnthropicLLMClient(api_key="sk-test")
        assert client._client is None

        with pytest.raises(LLMUnavailableError, match="not installed"):
            await client.text_completion(prompt="test")


class TestAnthropicExceptionNarrowing:
    async def test_programming_error_not_swallowed(self) -> None:
        """TypeError/KeyError should propagate, not become LLMUnavailableError."""
        client = AnthropicLLMClient(api_key="sk-test")

        with patch.object(client, "_client", create=True) as mock_client:
            mock_client.messages = AsyncMock()
            mock_client.messages.create = AsyncMock(
                side_effect=TypeError("unexpected keyword argument")
            )

            with pytest.raises(TypeError):
                await client.text_completion(prompt="test")
