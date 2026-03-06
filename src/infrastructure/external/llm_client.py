"""LLM client protocol and value objects for provider-agnostic LLM access."""

from __future__ import annotations

import enum
from dataclasses import dataclass
from typing import Any, Protocol, runtime_checkable


class LLMProvider(enum.Enum):
    """Supported LLM provider backends."""

    ANTHROPIC = "anthropic"
    OLLAMA = "ollama"
    VERTEXAI = "vertexai"
    NONE = "none"


@dataclass(frozen=True)
class LLMConfig:
    """Configuration for LLM client creation."""

    provider: LLMProvider = LLMProvider.NONE
    model: str = ""
    api_key: str = ""
    timeout: float = 30.0

    def __repr__(self) -> str:
        masked = "***" if self.api_key else ""
        return (
            f"LLMConfig(provider={self.provider!r}, model={self.model!r}, "
            f"api_key={masked!r}, timeout={self.timeout!r})"
        )


@dataclass(frozen=True)
class LLMResponse:
    """Response from an LLM call."""

    content: str
    model_used: str
    usage_tokens: int


@runtime_checkable
class LLMClient(Protocol):
    """Provider-agnostic LLM client interface."""

    async def structured_output(
        self, prompt: str, schema: dict[str, Any]
    ) -> LLMResponse: ...

    async def text_completion(self, prompt: str) -> LLMResponse: ...
