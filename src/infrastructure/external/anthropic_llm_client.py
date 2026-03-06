"""Anthropic LLM client adapter wrapping the Anthropic SDK."""

from __future__ import annotations

import json
from typing import Any

from src.domain.models.errors import LLMUnavailableError
from src.infrastructure.external.llm_client import LLMResponse


class AnthropicLLMClient:
    """LLMClient adapter backed by the Anthropic Messages API."""

    def __init__(
        self,
        api_key: str,
        model: str = "claude-sonnet-4-20250514",
        timeout: float = 30.0,
    ) -> None:
        self._model = model
        self._timeout = timeout
        try:
            from anthropic import AsyncAnthropic

            self._client: Any = AsyncAnthropic(api_key=api_key, timeout=timeout)
        except ImportError:
            self._client = None

    async def structured_output(
        self, prompt: str, schema: dict[str, Any]
    ) -> LLMResponse:
        if self._client is None:
            raise LLMUnavailableError("anthropic SDK not installed")
        schema_instruction = (
            "Respond with valid JSON matching this schema:\n"
            f"{json.dumps(schema)}"
        )
        try:
            response = await self._client.messages.create(
                model=self._model,
                max_tokens=4096,
                system=schema_instruction,
                messages=[{"role": "user", "content": prompt}],
            )
            return LLMResponse(
                content=response.content[0].text,
                model_used=response.model,
                usage_tokens=response.usage.input_tokens + response.usage.output_tokens,
            )
        except (OSError, ConnectionError, TimeoutError) as exc:
            raise LLMUnavailableError(str(exc)) from exc
        except Exception as exc:
            if "anthropic" in type(exc).__module__.lower():
                raise LLMUnavailableError(str(exc)) from exc
            raise

    async def text_completion(self, prompt: str) -> LLMResponse:
        if self._client is None:
            raise LLMUnavailableError("anthropic SDK not installed")
        try:
            response = await self._client.messages.create(
                model=self._model,
                max_tokens=4096,
                messages=[{"role": "user", "content": prompt}],
            )
            return LLMResponse(
                content=response.content[0].text,
                model_used=response.model,
                usage_tokens=response.usage.input_tokens + response.usage.output_tokens,
            )
        except (OSError, ConnectionError, TimeoutError) as exc:
            raise LLMUnavailableError(str(exc)) from exc
        except Exception as exc:
            if "anthropic" in type(exc).__module__.lower():
                raise LLMUnavailableError(str(exc)) from exc
            raise
