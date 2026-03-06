"""Tests for ChallengeHandler domain research integration.

Verifies that research() delegates to DomainResearchPort, returns None
when no port is configured, and that existing tests remain unaffected
(backward compatibility with optional keyword argument).
"""

from __future__ import annotations

from unittest.mock import AsyncMock

import pytest

from src.application.commands.challenge_handler import ChallengeHandler
from src.domain.models.domain_model import DomainModel
from src.domain.models.research import ResearchBriefing


class TestChallengeHandlerResearch:
    @pytest.mark.asyncio
    async def test_research_delegates_to_port(self) -> None:
        expected = ResearchBriefing(
            findings=(),
            no_data_areas=("Sales",),
            summary="No data",
        )
        research_port = AsyncMock()
        research_port.research.return_value = expected
        challenger = AsyncMock()
        challenger.generate_challenges.return_value = ()

        handler = ChallengeHandler(
            challenger=challenger,
            domain_research=research_port,
        )
        model = DomainModel()

        result = await handler.research(model)

        research_port.research.assert_awaited_once_with(model)
        assert result is expected

    @pytest.mark.asyncio
    async def test_research_returns_none_when_no_port(self) -> None:
        challenger = AsyncMock()
        challenger.generate_challenges.return_value = ()

        handler = ChallengeHandler(challenger=challenger)
        model = DomainModel()

        result = await handler.research(model)

        assert result is None

    @pytest.mark.asyncio
    async def test_backward_compat_existing_constructor(self) -> None:
        """Existing code using ChallengeHandler(challenger=port) still works."""
        challenger = AsyncMock()
        challenger.generate_challenges.return_value = ()

        handler = ChallengeHandler(challenger=challenger)

        # Should work without domain_research
        model = DomainModel()
        await handler.generate_challenges(model)
        challenger.generate_challenges.assert_awaited_once()
