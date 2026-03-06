"""Tests for NoOpLLMClient.

Verifies that all methods raise LLMUnavailableError with a descriptive message.
"""

from __future__ import annotations

import pytest

from src.domain.models.errors import LLMUnavailableError
from src.infrastructure.external.llm_client import LLMClient
from src.infrastructure.external.noop_llm_client import NoOpLLMClient


class TestNoOpLLMClientProtocol:
    def test_implements_llm_client_protocol(self) -> None:
        client = NoOpLLMClient()
        assert isinstance(client, LLMClient)


class TestNoOpStructuredOutput:
    async def test_noop_structured_output_raises_llm_unavailable(self) -> None:
        client = NoOpLLMClient()
        with pytest.raises(LLMUnavailableError, match="not configured"):
            await client.structured_output(
                prompt="test", schema={"type": "object"}
            )


class TestNoOpTextCompletion:
    async def test_noop_text_completion_raises_llm_unavailable(self) -> None:
        client = NoOpLLMClient()
        with pytest.raises(LLMUnavailableError, match="not configured"):
            await client.text_completion(prompt="test")
