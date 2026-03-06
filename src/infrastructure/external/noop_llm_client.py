"""No-op LLM client for graceful degradation when no provider is configured."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

from src.domain.models.errors import LLMUnavailableError

if TYPE_CHECKING:
    from src.infrastructure.external.llm_client import LLMResponse


class NoOpLLMClient:
    """LLMClient that always raises LLMUnavailableError."""

    async def structured_output(
        self, prompt: str, schema: dict[str, Any]
    ) -> LLMResponse:
        raise LLMUnavailableError("LLM service not configured")

    async def text_completion(self, prompt: str) -> LLMResponse:
        raise LLMUnavailableError("LLM service not configured")
