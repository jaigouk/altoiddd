"""Tests for FitnessTestSuite aggregate root.

Covers contract generation, invariant enforcement, preview, approve,
and edge cases per k7m.19 ticket.
"""

from __future__ import annotations

import pytest

from src.domain.models.domain_values import BoundedContext, SubdomainClassification
from src.domain.models.fitness_values import ContractType

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _core_bc(name: str = "Orders") -> BoundedContext:
    return BoundedContext(
        name=name,
        responsibility=f"Manages {name}",
        classification=SubdomainClassification.CORE,
    )


def _supporting_bc(name: str = "Notifications") -> BoundedContext:
    return BoundedContext(
        name=name,
        responsibility=f"Manages {name}",
        classification=SubdomainClassification.SUPPORTING,
    )


def _generic_bc(name: str = "Logging") -> BoundedContext:
    return BoundedContext(
        name=name,
        responsibility=f"Manages {name}",
        classification=SubdomainClassification.GENERIC,
    )


# ---------------------------------------------------------------------------
# 1. Suite creation
# ---------------------------------------------------------------------------


class TestCreateSuite:
    def test_new_suite_is_empty(self) -> None:
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        assert suite.contracts == ()
        assert suite.arch_rules == ()
        assert suite.root_package == "myapp"

    def test_suite_has_unique_id(self) -> None:
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        a = FitnessTestSuite(root_package="myapp")
        b = FitnessTestSuite(root_package="myapp")
        assert a.suite_id != b.suite_id


# ---------------------------------------------------------------------------
# 2. Contract generation per strictness level
# ---------------------------------------------------------------------------


class TestGenerateContracts:
    def test_core_produces_all_four_types(self) -> None:
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc(),))

        types = {c.contract_type for c in suite.contracts}
        assert types == {
            ContractType.LAYERS,
            ContractType.FORBIDDEN,
            ContractType.INDEPENDENCE,
            ContractType.ACYCLIC_SIBLINGS,
        }

    def test_supporting_produces_two_types(self) -> None:
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_supporting_bc(),))

        types = {c.contract_type for c in suite.contracts}
        assert types == {ContractType.LAYERS, ContractType.FORBIDDEN}

    def test_generic_produces_one_type(self) -> None:
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_generic_bc(),))

        types = {c.contract_type for c in suite.contracts}
        assert types == {ContractType.FORBIDDEN}

    def test_mixed_classifications(self) -> None:
        """2 Core + 1 Supporting + 1 Generic produces correct contract counts."""
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(
            bounded_contexts=(
                _core_bc("Orders"),
                _core_bc("Payments"),
                _supporting_bc("Notifications"),
                _generic_bc("Logging"),
            )
        )
        # Core: 4 each = 8, Supporting: 2, Generic: 1 = 11 total
        assert len(suite.contracts) == 11

    def test_contracts_have_correct_context_name(self) -> None:
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc("Orders"),))

        for c in suite.contracts:
            assert c.context_name == "Orders"

    def test_layers_contract_module_order(self) -> None:
        """Layers contract specifies infrastructure > application > domain."""
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc("Orders"),))

        layers_contracts = [
            c for c in suite.contracts if c.contract_type == ContractType.LAYERS
        ]
        assert len(layers_contracts) == 1
        modules = layers_contracts[0].modules
        # infrastructure first (top), domain last (bottom)
        assert "infrastructure" in modules[0].lower()
        assert "domain" in modules[-1].lower()

    def test_forbidden_contract_prevents_domain_importing_infra(self) -> None:
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc("Orders"),))

        forbidden = [
            c for c in suite.contracts if c.contract_type == ContractType.FORBIDDEN
        ]
        assert len(forbidden) >= 1
        # Should prevent domain from importing infrastructure
        f = forbidden[0]
        assert "domain" in " ".join(f.modules).lower()
        assert "infrastructure" in " ".join(f.forbidden_modules).lower()

    def test_arch_rules_generated_for_each_bc(self) -> None:
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc("Orders"),))

        assert len(suite.arch_rules) >= 1
        for r in suite.arch_rules:
            assert r.context_name == "Orders"


# ---------------------------------------------------------------------------
# 3. Edge cases: single BC, no siblings
# ---------------------------------------------------------------------------


class TestSingleBcEdgeCases:
    def test_single_bc_still_generates_independence_contract(self) -> None:
        """With only one BC, independence contract is meaningless but still generated
        (the contract itself just won't have sibling modules to check against)."""
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc(),))

        # Independence and acyclic_siblings still generated per invariant,
        # but with empty sibling lists they are effectively no-ops
        types = {c.contract_type for c in suite.contracts}
        assert ContractType.INDEPENDENCE in types

    def test_empty_bounded_contexts_raises(self) -> None:
        """I-3: Should raise InvariantViolationError, not ValueError."""
        from src.domain.models.errors import InvariantViolationError
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        with pytest.raises(InvariantViolationError, match="No bounded contexts"):
            suite.generate_contracts(bounded_contexts=())

    def test_bc_without_classification_raises(self) -> None:
        """I-3: Should raise InvariantViolationError, not ValueError."""
        from src.domain.models.errors import InvariantViolationError
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        unclassified = BoundedContext(
            name="Unclassified",
            responsibility="test",
            classification=None,
        )
        suite = FitnessTestSuite(root_package="myapp")
        with pytest.raises(InvariantViolationError, match="no subdomain classification"):
            suite.generate_contracts(bounded_contexts=(unclassified,))


# ---------------------------------------------------------------------------
# 4. Invariant enforcement on approve()
# ---------------------------------------------------------------------------


class TestApproveInvariants:
    def test_approve_after_generate_succeeds(self) -> None:
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc(),))
        suite.approve()  # Should not raise

    def test_approve_without_generate_raises(self) -> None:
        """Cannot approve an empty suite."""
        from src.domain.models.errors import InvariantViolationError
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        with pytest.raises(InvariantViolationError, match="no contracts"):
            suite.approve()

    def test_approve_emits_fitness_tests_generated(self) -> None:
        from src.domain.events.fitness_events import FitnessTestsGenerated
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc(),))
        suite.approve()

        assert len(suite.events) == 1
        assert isinstance(suite.events[0], FitnessTestsGenerated)

    def test_events_are_tuple(self) -> None:
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc(),))
        suite.approve()
        assert isinstance(suite.events, tuple)

    def test_double_approve_raises(self) -> None:
        """Cannot approve twice."""
        from src.domain.models.errors import InvariantViolationError
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc(),))
        suite.approve()
        with pytest.raises(InvariantViolationError, match="already approved"):
            suite.approve()


# ---------------------------------------------------------------------------
# 5. Preview
# ---------------------------------------------------------------------------


class TestPreview:
    def test_preview_returns_string(self) -> None:
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc("Orders"),))
        preview = suite.preview()
        assert isinstance(preview, str)
        assert "Orders" in preview

    def test_preview_shows_contract_counts(self) -> None:
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(
            bounded_contexts=(
                _core_bc("Orders"),
                _supporting_bc("Notifications"),
            )
        )
        preview = suite.preview()
        assert "Orders" in preview
        assert "Notifications" in preview
        assert "STRICT" in preview or "strict" in preview.lower()
        assert "MODERATE" in preview or "moderate" in preview.lower()

    def test_preview_without_generate_raises(self) -> None:
        from src.domain.models.errors import InvariantViolationError
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        with pytest.raises(InvariantViolationError, match="No contracts"):
            suite.preview()


# ---------------------------------------------------------------------------
# 6. Rendering (TOML + pytestarch test files)
# ---------------------------------------------------------------------------


class TestRendering:
    def test_render_import_linter_toml(self) -> None:
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc("Orders"),))
        toml = suite.render_import_linter_toml()

        assert "[tool.importlinter]" in toml
        assert 'root_package = "myapp"' in toml
        assert "[[tool.importlinter.contracts]]" in toml
        assert "Orders" in toml

    def test_render_pytestarch_test_file(self) -> None:
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc("Orders"),))
        test_code = suite.render_pytestarch_tests()

        assert "from pytestarch" in test_code
        assert "def test_" in test_code
        assert "orders" in test_code.lower()

    def test_render_toml_without_contracts_raises(self) -> None:
        from src.domain.models.errors import InvariantViolationError
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        with pytest.raises(InvariantViolationError, match="No contracts"):
            suite.render_import_linter_toml()

    def test_toml_layers_correct_format(self) -> None:
        """Generated TOML matches import-linter's expected format."""
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc("Orders"),))
        toml = suite.render_import_linter_toml()

        # Must have type = "layers" for at least one contract
        assert 'type = "layers"' in toml

    def test_toml_forbidden_correct_format(self) -> None:
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc("Orders"),))
        toml = suite.render_import_linter_toml()

        assert 'type = "forbidden"' in toml

    def test_multiple_bcs_render_separate_contracts(self) -> None:
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(
            bounded_contexts=(
                _core_bc("Orders"),
                _supporting_bc("Notifications"),
            )
        )
        toml = suite.render_import_linter_toml()

        # Should have multiple [[tool.importlinter.contracts]] sections
        assert toml.count("[[tool.importlinter.contracts]]") >= 3  # 4 + 2 contracts


# ---------------------------------------------------------------------------
# 7. Invariant 5: no cross-context module references (C-1)
# ---------------------------------------------------------------------------


class TestInvariant5ModuleBoundaries:
    def test_generated_contracts_pass_boundary_check(self) -> None:
        """Contracts built by generate_contracts() must pass invariant 5."""
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(
            bounded_contexts=(
                _core_bc("Orders"),
                _supporting_bc("Notifications"),
            )
        )
        # approve() should succeed — all modules are within their BC
        suite.approve()
        assert len(suite.events) == 1

    def test_approve_validates_module_boundaries(self) -> None:
        """If a contract references a module outside its BC, approve() must reject."""
        from src.domain.models.errors import InvariantViolationError
        from src.domain.models.fitness_test_suite import FitnessTestSuite
        from src.domain.models.fitness_values import Contract, ContractType

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc("Orders"),))

        # Inject a bad contract that references another BC's module
        suite._contracts.append(
            Contract(
                name="cross-context violation",
                contract_type=ContractType.FORBIDDEN,
                context_name="Orders",
                modules=("myapp.payments.domain",),  # wrong BC!
                forbidden_modules=("myapp.orders.infrastructure",),
            )
        )

        with pytest.raises(InvariantViolationError, match=r"outside.*bounded context"):
            suite.approve()

    def test_forbidden_modules_also_checked(self) -> None:
        """forbidden_modules must also be within the contract's BC."""
        from src.domain.models.errors import InvariantViolationError
        from src.domain.models.fitness_test_suite import FitnessTestSuite
        from src.domain.models.fitness_values import Contract, ContractType

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc("Orders"),))

        suite._contracts.append(
            Contract(
                name="cross-context forbidden",
                contract_type=ContractType.FORBIDDEN,
                context_name="Orders",
                modules=("myapp.orders.domain",),
                forbidden_modules=("myapp.payments.infrastructure",),  # wrong BC!
            )
        )

        with pytest.raises(InvariantViolationError, match=r"outside.*bounded context"):
            suite.approve()


# ---------------------------------------------------------------------------
# 8. Approve event includes arch_rules (I-2)
# ---------------------------------------------------------------------------


class TestApproveEventPayload:
    def test_event_includes_arch_rules(self) -> None:
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc("Orders"),))
        suite.approve()

        event = suite.events[0]
        assert hasattr(event, "arch_rules")
        assert len(event.arch_rules) >= 1

    def test_event_contracts_and_rules_match_suite(self) -> None:
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc("Orders"),))
        suite.approve()

        event = suite.events[0]
        assert event.contracts == suite.contracts
        assert event.arch_rules == suite.arch_rules


# ---------------------------------------------------------------------------
# 9. _snake_case edge cases (I-5)
# ---------------------------------------------------------------------------


class TestSnakeCase:
    def test_pascal_case(self) -> None:
        from src.domain.models.fitness_test_suite import _snake_case

        assert _snake_case("OrderManagement") == "order_management"

    def test_space_separated(self) -> None:
        from src.domain.models.fitness_test_suite import _snake_case

        assert _snake_case("Order Management") == "order_management"

    def test_hyphenated_no_double_underscores(self) -> None:
        """I-5: Hyphens should produce single underscores, not double."""
        from src.domain.models.fitness_test_suite import _snake_case

        assert _snake_case("Architecture-Testing") == "architecture_testing"
        assert "__" not in _snake_case("Architecture-Testing")

    def test_consecutive_uppercase(self) -> None:
        """I-5: 'ABCTest' should become 'abc_test', not 'a_b_c_test'."""
        from src.domain.models.fitness_test_suite import _snake_case

        result = _snake_case("ABCTest")
        assert result == "abc_test"

    def test_already_snake_case(self) -> None:
        from src.domain.models.fitness_test_suite import _snake_case

        assert _snake_case("order_management") == "order_management"


# ---------------------------------------------------------------------------
# 10. generate_contracts() after approve() must be rejected (I-8)
# ---------------------------------------------------------------------------


class TestGenerateAfterApprove:
    def test_generate_after_approve_raises(self) -> None:
        """I-8: An approved suite is finalized — no regeneration allowed."""
        from src.domain.models.errors import InvariantViolationError
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc(),))
        suite.approve()

        with pytest.raises(InvariantViolationError, match="approved"):
            suite.generate_contracts(bounded_contexts=(_core_bc(),))


# ---------------------------------------------------------------------------
# 11. Precondition errors use InvariantViolationError consistently (N-1)
# ---------------------------------------------------------------------------


class TestPreconditionErrorConsistency:
    def test_preview_without_contracts_raises_invariant_error(self) -> None:
        from src.domain.models.errors import InvariantViolationError
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        with pytest.raises(InvariantViolationError, match="No contracts"):
            suite.preview()

    def test_render_toml_without_contracts_raises_invariant_error(self) -> None:
        from src.domain.models.errors import InvariantViolationError
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        with pytest.raises(InvariantViolationError, match="No contracts"):
            suite.render_import_linter_toml()

    def test_render_pytestarch_without_contracts_raises_invariant_error(self) -> None:
        """N-2: Missing coverage for render_pytestarch_tests() precondition."""
        from src.domain.models.errors import InvariantViolationError
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        with pytest.raises(InvariantViolationError, match="No contracts"):
            suite.render_pytestarch_tests()


# ---------------------------------------------------------------------------
# 12. TOML renders all forbidden_modules entries (I-10)
# ---------------------------------------------------------------------------


class TestTomlMultiForbidden:
    def test_forbidden_renders_all_source_modules(self) -> None:
        """I-10: All source_modules entries should be rendered, not just [0]."""
        from src.domain.models.fitness_test_suite import FitnessTestSuite
        from src.domain.models.fitness_values import Contract, ContractType

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc("Orders"),))

        # Inject a forbidden contract with multiple source modules
        suite._contracts.append(
            Contract(
                name="multi source",
                contract_type=ContractType.FORBIDDEN,
                context_name="Orders",
                modules=(
                    "myapp.orders.domain",
                    "myapp.orders.application",
                ),
                forbidden_modules=("myapp.orders.infrastructure",),
            )
        )

        toml = suite.render_import_linter_toml()
        assert "myapp.orders.domain" in toml
        assert "myapp.orders.application" in toml

    def test_forbidden_renders_all_forbidden_modules(self) -> None:
        """I-10: All forbidden_modules entries should be rendered, not just [0]."""
        from src.domain.models.fitness_test_suite import FitnessTestSuite
        from src.domain.models.fitness_values import Contract, ContractType

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc("Orders"),))

        suite._contracts.append(
            Contract(
                name="multi forbidden",
                contract_type=ContractType.FORBIDDEN,
                context_name="Orders",
                modules=("myapp.orders.domain",),
                forbidden_modules=(
                    "myapp.orders.infrastructure",
                    "myapp.orders.external",
                ),
            )
        )

        toml = suite.render_import_linter_toml()
        assert "myapp.orders.infrastructure" in toml
        assert "myapp.orders.external" in toml


# ---------------------------------------------------------------------------
# 13. Additional edge cases
# ---------------------------------------------------------------------------


class TestAdditionalEdgeCases:
    def test_generate_clears_previous_contracts(self) -> None:
        """Calling generate_contracts() a second time replaces, not appends."""
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc("Orders"),))
        first_count = len(suite.contracts)

        suite.generate_contracts(bounded_contexts=(_generic_bc("Logging"),))
        # Should have ONLY logging contracts, not orders + logging
        assert len(suite.contracts) < first_count
        for c in suite.contracts:
            assert c.context_name == "Logging"

    def test_arch_rules_cleared_on_regenerate(self) -> None:
        """Regeneration also replaces arch rules."""
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc("Orders"),))
        assert all(r.context_name == "Orders" for r in suite.arch_rules)

        suite.generate_contracts(bounded_contexts=(_generic_bc("Logging"),))
        assert all(r.context_name == "Logging" for r in suite.arch_rules)

    def test_pytestarch_test_has_correct_function_names(self) -> None:
        """Generated test function names use snake_case of BC name."""
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(
            bounded_contexts=(_core_bc("OrderManagement"),)
        )
        test_code = suite.render_pytestarch_tests()
        assert "def test_order_management_domain_isolation" in test_code

    def test_event_carries_suite_id(self) -> None:
        """The emitted event's suite_id matches the suite."""
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=(_core_bc(),))
        suite.approve()
        assert suite.events[0].suite_id == suite.suite_id

    def test_event_root_package_matches(self) -> None:
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="custom_pkg")
        suite.generate_contracts(bounded_contexts=(_core_bc(),))
        suite.approve()
        assert suite.events[0].root_package == "custom_pkg"

    def test_module_prefix_uses_snake_case_context_name(self) -> None:
        """Module paths should use snake_case, not PascalCase."""
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(
            bounded_contexts=(_core_bc("OrderManagement"),)
        )
        for c in suite.contracts:
            for mod in c.modules:
                assert "OrderManagement" not in mod
                assert "order_management" in mod

    def test_many_bcs_produce_correct_total_contracts(self) -> None:
        """5 Core BCs should produce 5*4=20 contracts + 5 arch rules."""
        from src.domain.models.fitness_test_suite import FitnessTestSuite

        bcs = tuple(
            _core_bc(f"Context{i}") for i in range(5)
        )
        suite = FitnessTestSuite(root_package="myapp")
        suite.generate_contracts(bounded_contexts=bcs)
        assert len(suite.contracts) == 20
        assert len(suite.arch_rules) == 5
