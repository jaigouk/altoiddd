"""Tests for DomainModelGenerated event."""

from __future__ import annotations

import pytest

from src.domain.events.domain_events import DomainModelGenerated
from src.domain.models.domain_values import (
    AggregateDesign,
    BoundedContext,
    DomainStory,
    SubdomainClassification,
)
from src.domain.models.ubiquitous_language import TermEntry


class TestDomainModelGenerated:
    def test_create(self) -> None:
        event = DomainModelGenerated(
            model_id="test-123",
            domain_stories=(
                DomainStory(
                    name="Test", actors=("A",), trigger="T", steps=("S",)
                ),
            ),
            ubiquitous_language=(
                TermEntry(term="Order", definition="A purchase", context_name="Sales"),
            ),
            bounded_contexts=(
                BoundedContext(
                    name="Sales",
                    responsibility="Orders",
                    classification=SubdomainClassification.CORE,
                ),
            ),
            context_relationships=(),
            aggregate_designs=(
                AggregateDesign(
                    name="OrderAgg", context_name="Sales", root_entity="Order"
                ),
            ),
        )
        assert event.model_id == "test-123"
        assert len(event.domain_stories) == 1
        assert len(event.bounded_contexts) == 1

    def test_frozen(self) -> None:
        event = DomainModelGenerated(
            model_id="test",
            domain_stories=(),
            ubiquitous_language=(),
            bounded_contexts=(),
            context_relationships=(),
            aggregate_designs=(),
        )
        with pytest.raises(AttributeError):
            event.model_id = "changed"  # type: ignore[misc]

    def test_empty_event(self) -> None:
        event = DomainModelGenerated(
            model_id="empty",
            domain_stories=(),
            ubiquitous_language=(),
            bounded_contexts=(),
            context_relationships=(),
            aggregate_designs=(),
        )
        assert event.domain_stories == ()
        assert event.aggregate_designs == ()
