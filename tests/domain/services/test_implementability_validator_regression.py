"""Regression tests for implementability validation.

These tests reproduce the specific 20c.5 scenario: a ticket description
that says "adapter performs iterative web search" without specifying
which port or library implements the search. The validator should catch
this as a CRITICAL unspecified dependency.
"""

from __future__ import annotations

from src.domain.models.ticket_implementability import FindingSeverity
from src.domain.services.implementability_validator import (
    ImplementabilityValidator,
)


class TestRegression20c5:
    """Regression: ticket 20c.5 had 'adapter performs web search' with no port."""

    def test_reproduces_20c5_unspecified_web_search(self) -> None:
        """Ticket with 'adapter performs iterative web search' -> CRITICAL."""
        from src.domain.models.ticket_values import (
            GeneratedTicket,
            TicketDetailLevel,
        )

        description = (
            "## Goal\n"
            "Implement domain research with RLM pattern.\n\n"
            "## DDD Alignment\n"
            "Bounded Context: Knowledge Base\n\n"
            "## Design\n"
            "### Invariants\n"
            "- Research findings must have sources\n\n"
            "The RLM research adapter performs iterative web search to "
            "gather domain intelligence. Results are synthesized via LLM.\n\n"
            "## SOLID Mapping\n"
            "- SRP: RlmResearchAdapter handles research only\n\n"
            "## TDD Workflow\n"
            "RED: test_research_returns_findings\n\n"
            "## Steps\n"
            "1. Create RlmResearchAdapter\n\n"
            "## Acceptance Criteria\n"
            "- [ ] Adapter returns ResearchBriefing\n"
            "- [ ] Findings have source attribution\n\n"
            "## Edge Cases\n"
            "- LLM unavailable -> graceful degradation\n"
        )

        ticket = GeneratedTicket(
            ticket_id="test-20c5",
            title="Domain Research Port and RLM Adapter",
            description=description,
            detail_level=TicketDetailLevel.FULL,
            epic_id="test-epic",
            bounded_context_name="Knowledge Base",
            aggregate_name="Research",
            dependencies=(),
            depth=0,
        )

        result = ImplementabilityValidator.validate(ticket)

        assert not result.is_valid
        critical = [
            f
            for f in result.findings
            if f.severity == FindingSeverity.CRITICAL
        ]
        assert len(critical) >= 1
        assert any("web search" in f.description.lower() for f in critical)
