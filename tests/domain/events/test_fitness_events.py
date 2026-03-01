"""Tests for FitnessTestsGenerated domain event."""

from __future__ import annotations

from src.domain.events.fitness_events import FitnessTestsGenerated
from src.domain.models.fitness_values import ArchRule, Contract, ContractType


class TestFitnessTestsGenerated:
    def test_create_event(self) -> None:
        event = FitnessTestsGenerated(
            suite_id="test-id",
            root_package="myapp",
            contracts=(
                Contract(
                    name="test",
                    contract_type=ContractType.LAYERS,
                    context_name="Orders",
                    modules=("a", "b"),
                ),
            ),
            arch_rules=(
                ArchRule(
                    name="Orders domain isolation",
                    assertion=(
                        "modules in myapp.orders.domain should not "
                        "import from myapp.orders.infrastructure"
                    ),
                    context_name="Orders",
                ),
            ),
        )
        assert event.suite_id == "test-id"
        assert event.root_package == "myapp"
        assert len(event.contracts) == 1
        assert len(event.arch_rules) == 1

    def test_event_is_frozen(self) -> None:
        import pytest

        event = FitnessTestsGenerated(
            suite_id="test-id",
            root_package="myapp",
            contracts=(),
            arch_rules=(),
        )
        with pytest.raises(AttributeError):
            event.suite_id = "changed"  # type: ignore[misc]

    def test_event_includes_arch_rules_field(self) -> None:
        """I-2: Event must carry arch_rules alongside contracts."""
        event = FitnessTestsGenerated(
            suite_id="test-id",
            root_package="myapp",
            contracts=(),
            arch_rules=(
                ArchRule(
                    name="test rule",
                    assertion="test assertion",
                    context_name="TestCtx",
                ),
            ),
        )
        assert hasattr(event, "arch_rules")
        assert len(event.arch_rules) == 1
        assert event.arch_rules[0].context_name == "TestCtx"
