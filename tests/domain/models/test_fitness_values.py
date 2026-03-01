"""Tests for fitness function value objects.

Covers ContractStrictness enum, ContractType enum, Contract VO, and ArchRule VO.
"""

from __future__ import annotations

import pytest

from src.domain.models.fitness_values import (
    ArchRule,
    Contract,
    ContractStrictness,
    ContractType,
)

# ---------------------------------------------------------------------------
# ContractStrictness enum
# ---------------------------------------------------------------------------


class TestContractStrictness:
    def test_strict_value(self) -> None:
        assert ContractStrictness.STRICT.value == "strict"

    def test_moderate_value(self) -> None:
        assert ContractStrictness.MODERATE.value == "moderate"

    def test_minimal_value(self) -> None:
        assert ContractStrictness.MINIMAL.value == "minimal"

    def test_from_classification_core(self) -> None:
        from src.domain.models.domain_values import SubdomainClassification

        result = ContractStrictness.from_classification(SubdomainClassification.CORE)
        assert result == ContractStrictness.STRICT

    def test_from_classification_supporting(self) -> None:
        from src.domain.models.domain_values import SubdomainClassification

        result = ContractStrictness.from_classification(
            SubdomainClassification.SUPPORTING
        )
        assert result == ContractStrictness.MODERATE

    def test_from_classification_generic(self) -> None:
        from src.domain.models.domain_values import SubdomainClassification

        result = ContractStrictness.from_classification(SubdomainClassification.GENERIC)
        assert result == ContractStrictness.MINIMAL

    def test_required_contract_types_strict(self) -> None:
        types = ContractStrictness.STRICT.required_contract_types()
        assert set(types) == {
            ContractType.LAYERS,
            ContractType.FORBIDDEN,
            ContractType.INDEPENDENCE,
            ContractType.ACYCLIC_SIBLINGS,
        }

    def test_required_contract_types_moderate(self) -> None:
        types = ContractStrictness.MODERATE.required_contract_types()
        assert set(types) == {ContractType.LAYERS, ContractType.FORBIDDEN}

    def test_required_contract_types_minimal(self) -> None:
        types = ContractStrictness.MINIMAL.required_contract_types()
        assert set(types) == {ContractType.FORBIDDEN}


# ---------------------------------------------------------------------------
# ContractType enum
# ---------------------------------------------------------------------------


class TestContractType:
    def test_layers_value(self) -> None:
        assert ContractType.LAYERS.value == "layers"

    def test_forbidden_value(self) -> None:
        assert ContractType.FORBIDDEN.value == "forbidden"

    def test_independence_value(self) -> None:
        assert ContractType.INDEPENDENCE.value == "independence"

    def test_acyclic_siblings_value(self) -> None:
        assert ContractType.ACYCLIC_SIBLINGS.value == "acyclic_siblings"


# ---------------------------------------------------------------------------
# Contract value object (frozen)
# ---------------------------------------------------------------------------


class TestContract:
    def test_create_layers_contract(self) -> None:
        c = Contract(
            name="DDD layers",
            contract_type=ContractType.LAYERS,
            context_name="Orders",
            modules=("orders.infrastructure", "orders.application", "orders.domain"),
        )
        assert c.name == "DDD layers"
        assert c.contract_type == ContractType.LAYERS
        assert c.context_name == "Orders"
        assert len(c.modules) == 3

    def test_contract_is_frozen(self) -> None:
        c = Contract(
            name="test",
            contract_type=ContractType.FORBIDDEN,
            context_name="Ctx",
            modules=("a", "b"),
        )
        with pytest.raises(AttributeError):
            c.name = "changed"  # type: ignore[misc]

    def test_contract_equality(self) -> None:
        a = Contract(
            name="test",
            contract_type=ContractType.FORBIDDEN,
            context_name="Ctx",
            modules=("a",),
        )
        b = Contract(
            name="test",
            contract_type=ContractType.FORBIDDEN,
            context_name="Ctx",
            modules=("a",),
        )
        assert a == b

    def test_forbidden_source_and_target(self) -> None:
        """Forbidden contracts have optional source/target for clarity."""
        c = Contract(
            name="domain isolation",
            contract_type=ContractType.FORBIDDEN,
            context_name="Orders",
            modules=("orders.domain",),
            forbidden_modules=("orders.infrastructure",),
        )
        assert c.forbidden_modules == ("orders.infrastructure",)


# ---------------------------------------------------------------------------
# ArchRule value object (frozen)
# ---------------------------------------------------------------------------


class TestArchRule:
    def test_create_arch_rule(self) -> None:
        r = ArchRule(
            name="domain isolation",
            assertion="modules in orders.domain should not import from orders.infrastructure",
            context_name="Orders",
        )
        assert r.name == "domain isolation"
        assert r.context_name == "Orders"

    def test_arch_rule_is_frozen(self) -> None:
        r = ArchRule(
            name="test",
            assertion="some assertion",
            context_name="Ctx",
        )
        with pytest.raises(AttributeError):
            r.name = "changed"  # type: ignore[misc]

    def test_arch_rule_equality(self) -> None:
        a = ArchRule(name="test", assertion="x", context_name="Ctx")
        b = ArchRule(name="test", assertion="x", context_name="Ctx")
        assert a == b
