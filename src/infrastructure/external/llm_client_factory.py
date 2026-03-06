"""Factory for creating LLM client instances based on configuration."""

from __future__ import annotations

from src.infrastructure.external.llm_client import LLMClient, LLMConfig, LLMProvider
from src.infrastructure.external.noop_llm_client import NoOpLLMClient


class LLMClientFactory:
    """Creates the appropriate LLMClient based on LLMConfig."""

    @staticmethod
    def create(config: LLMConfig) -> LLMClient:
        if config.provider == LLMProvider.NONE:
            return NoOpLLMClient()

        if config.provider == LLMProvider.ANTHROPIC:
            if not config.api_key:
                return NoOpLLMClient()
            try:
                import anthropic  # noqa: F401

                from src.infrastructure.external.anthropic_llm_client import (
                    AnthropicLLMClient,
                )

                return AnthropicLLMClient(
                    api_key=config.api_key,
                    model=config.model,
                    timeout=config.timeout,
                )
            except ImportError:
                return NoOpLLMClient()

        return NoOpLLMClient()
