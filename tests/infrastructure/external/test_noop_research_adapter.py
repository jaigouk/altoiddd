"""Tests for NoOpResearchAdapter — empty briefing fallback.

Verifies protocol compliance, empty findings, bounded context names
as no_data_areas, and empty summary.
"""

from __future__ import annotations

import pytest

from src.application.ports.domain_research_port import DomainResearchPort
from src.domain.models.domain_model import DomainModel
from src.domain.models.domain_values import BoundedContext
from src.infrastructure.external.noop_research_adapter import NoOpResearchAdapter


class TestNoOpResearchAdapter:
    def test_satisfies_domain_research_port(self) -> None:
        adapter = NoOpResearchAdapter()
        assert isinstance(adapter, DomainResearchPort)

    @pytest.mark.asyncio
    async def test_returns_empty_briefing(self) -> None:
        adapter = NoOpResearchAdapter()
        model = DomainModel()
        briefing = await adapter.research(model)
        assert len(briefing.findings) == 0
        assert briefing.summary == ""

    @pytest.mark.asyncio
    async def test_bounded_context_names_as_no_data_areas(self) -> None:
        adapter = NoOpResearchAdapter()
        model = DomainModel()
        model.add_bounded_context(
            BoundedContext(name="Sales", responsibility="Sell things")
        )
        model.add_bounded_context(
            BoundedContext(name="Billing", responsibility="Charge money")
        )

        briefing = await adapter.research(model)

        assert set(briefing.no_data_areas) == {"Sales", "Billing"}

    @pytest.mark.asyncio
    async def test_empty_model_returns_empty_no_data(self) -> None:
        adapter = NoOpResearchAdapter()
        model = DomainModel()

        briefing = await adapter.research(model)

        assert briefing.no_data_areas == ()
