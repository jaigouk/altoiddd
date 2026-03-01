"""FitnessTestSuite aggregate root for the Architecture Testing bounded context.

Generates import-linter contracts and pytestarch rules from bounded contexts
and their subdomain classifications. Enforces invariants per DDD.md Section 5.

Invariants:
1. Every bounded context must have at least one contract.
2. Core subdomains must have all 4 contract types.
3. Supporting subdomains must have layers + forbidden.
4. Generic subdomains must have at least forbidden.
5. No contract may reference modules outside its bounded context.
"""

from __future__ import annotations

import uuid
from typing import TYPE_CHECKING

from src.domain.models.errors import InvariantViolationError
from src.domain.models.fitness_values import (
    ArchRule,
    Contract,
    ContractStrictness,
    ContractType,
)

if TYPE_CHECKING:
    from src.domain.events.fitness_events import FitnessTestsGenerated
    from src.domain.models.domain_values import BoundedContext


def _snake_case(name: str) -> str:
    """Convert a PascalCase, space-separated, or hyphenated name to snake_case.

    Handles consecutive uppercase gracefully: ``ABCTest`` → ``abc_test``.
    """
    result: list[str] = []
    for i, ch in enumerate(name):
        if ch.isupper():
            if result and result[-1] != "_":
                # Only insert underscore before an uppercase char when:
                # - previous char was lowercase, OR
                # - next char is lowercase (end of acronym like ABCTest → abc_test)
                prev_lower = name[i - 1].islower() if i > 0 else False
                next_lower = name[i + 1].islower() if i + 1 < len(name) else False
                if prev_lower or (next_lower and i > 0 and name[i - 1].isupper()):
                    result.append("_")
            result.append(ch.lower())
        elif ch in (" ", "-"):
            if result and result[-1] != "_":
                result.append("_")
        else:
            result.append(ch)
    return "".join(result)


class FitnessTestSuite:
    """Aggregate root: generates and manages architecture fitness tests.

    Attributes:
        suite_id: Unique identifier for this suite.
        root_package: The root Python package being tested.
    """

    def __init__(self, root_package: str) -> None:
        self.suite_id: str = str(uuid.uuid4())
        self.root_package: str = root_package
        self._contracts: list[Contract] = []
        self._arch_rules: list[ArchRule] = []
        self._events: list[FitnessTestsGenerated] = []
        self._approved: bool = False

    # -- Properties -----------------------------------------------------------

    @property
    def contracts(self) -> tuple[Contract, ...]:
        """All generated contracts (defensive copy)."""
        return tuple(self._contracts)

    @property
    def arch_rules(self) -> tuple[ArchRule, ...]:
        """All generated pytestarch rules (defensive copy)."""
        return tuple(self._arch_rules)

    @property
    def events(self) -> tuple[FitnessTestsGenerated, ...]:
        """Domain events produced by this aggregate (defensive copy)."""
        return tuple(self._events)

    # -- Commands -------------------------------------------------------------

    def generate_contracts(
        self,
        bounded_contexts: tuple[BoundedContext, ...],
    ) -> None:
        """Generate contracts for each bounded context based on classification.

        Args:
            bounded_contexts: The classified bounded contexts.

        Raises:
            InvariantViolationError: If suite is already approved, no bounded
                contexts provided, or any BC lacks classification.
        """
        if self._approved:
            msg = "Cannot regenerate contracts on an approved suite"
            raise InvariantViolationError(msg)

        if not bounded_contexts:
            msg = "No bounded contexts to generate fitness tests for"
            raise InvariantViolationError(msg)

        self._contracts.clear()
        self._arch_rules.clear()

        for bc in bounded_contexts:
            if bc.classification is None:
                msg = f"Bounded context '{bc.name}' has no subdomain classification"
                raise InvariantViolationError(msg)

            strictness = ContractStrictness.from_classification(bc.classification)
            required_types = strictness.required_contract_types()
            module_prefix = f"{self.root_package}.{_snake_case(bc.name)}"

            for ct in required_types:
                contract = self._build_contract(ct, bc.name, module_prefix)
                self._contracts.append(contract)

            # Generate pytestarch rule for each BC
            self._arch_rules.append(
                ArchRule(
                    name=f"{bc.name} domain isolation",
                    assertion=(
                        f"modules in {module_prefix}.domain should not "
                        f"import from {module_prefix}.infrastructure"
                    ),
                    context_name=bc.name,
                )
            )

    def preview(self) -> str:
        """Return a human-readable preview of generated contracts.

        Raises:
            InvariantViolationError: If no contracts have been generated.
        """
        if not self._contracts:
            msg = "No contracts generated yet — call generate_contracts() first"
            raise InvariantViolationError(msg)

        lines: list[str] = [
            f"Fitness Test Suite: {self.root_package}",
            f"Total contracts: {len(self._contracts)}",
            f"Total arch rules: {len(self._arch_rules)}",
            "",
        ]

        # Group by context
        contexts_seen: dict[str, list[Contract]] = {}
        for c in self._contracts:
            contexts_seen.setdefault(c.context_name, []).append(c)

        for ctx_name, contracts in contexts_seen.items():
            strictness = self._infer_strictness(contracts)
            types = [c.contract_type.value for c in contracts]
            lines.append(f"  {ctx_name} ({strictness.value.upper()}):")
            lines.append(f"    Contracts: {', '.join(types)}")
            lines.append("")

        return "\n".join(lines)

    def approve(self) -> None:
        """Approve the suite, emitting FitnessTestsGenerated.

        Raises:
            InvariantViolationError: If suite has no contracts, is already
                approved, or any contract references modules outside its BC.
        """
        if self._approved:
            msg = "Suite already approved"
            raise InvariantViolationError(msg)

        if not self._contracts:
            msg = "Cannot approve suite with no contracts"
            raise InvariantViolationError(msg)

        self._validate_module_boundaries()

        self._approved = True

        from src.domain.events.fitness_events import FitnessTestsGenerated

        self._events.append(
            FitnessTestsGenerated(
                suite_id=self.suite_id,
                root_package=self.root_package,
                contracts=tuple(self._contracts),
                arch_rules=tuple(self._arch_rules),
            )
        )

    # -- Rendering ------------------------------------------------------------

    def render_import_linter_toml(self) -> str:
        """Render import-linter TOML configuration.

        Returns:
            Valid TOML string for [tool.importlinter] section.

        Raises:
            InvariantViolationError: If no contracts generated.
        """
        if not self._contracts:
            msg = "No contracts generated yet — call generate_contracts() first"
            raise InvariantViolationError(msg)

        lines: list[str] = [
            "[tool.importlinter]",
            f'root_package = "{self.root_package}"',
            "",
        ]

        for c in self._contracts:
            lines.append("[[tool.importlinter.contracts]]")
            lines.append(f'name = "{c.name}"')
            lines.append(f'type = "{c.contract_type.value}"')

            if c.contract_type == ContractType.LAYERS:
                lines.append("layers = [")
                lines.extend(f'  "{m}",' for m in c.modules)
                lines.append("]")
            elif c.contract_type == ContractType.FORBIDDEN:
                lines.append("source_modules = [")
                lines.extend(f'  "{m}",' for m in c.modules)
                lines.append("]")
                if c.forbidden_modules:
                    lines.append("forbidden_modules = [")
                    lines.extend(f'  "{m}",' for m in c.forbidden_modules)
                    lines.append("]")
            elif c.contract_type == ContractType.INDEPENDENCE:
                lines.append("modules = [")
                lines.extend(f'  "{m}",' for m in c.modules)
                lines.append("]")
            elif c.contract_type == ContractType.ACYCLIC_SIBLINGS:
                lines.append(f'source_module = "{c.modules[0]}"')

            lines.append("")

        return "\n".join(lines)

    def render_pytestarch_tests(self) -> str:
        """Render pytestarch test file content.

        Returns:
            Valid Python test file using pytestarch.

        Raises:
            InvariantViolationError: If no contracts generated.
        """
        if not self._contracts:
            msg = "No contracts generated yet — call generate_contracts() first"
            raise InvariantViolationError(msg)

        lines: list[str] = [
            '"""Auto-generated architecture fitness tests.',
            "",
            "Generated by alty from bounded context map.",
            '"""',
            "",
            "import pytest",
            "from pytestarch import get_evaluable_architecture, Rule",
            "",
            "",
            '@pytest.fixture(scope="session")',
            "def evaluable():",
            f'    return get_evaluable_architecture(".", "{self.root_package}")',
            "",
        ]

        for rule in self._arch_rules:
            fn_name = f"test_{_snake_case(rule.name)}"
            lines.append("")
            lines.append(f"def {fn_name}(evaluable):")
            lines.append(f'    """{rule.assertion}"""')

            # Extract module names from assertion
            ctx_module = _snake_case(rule.context_name)
            domain_mod = f"{self.root_package}.{ctx_module}.domain"
            infra_mod = f"{self.root_package}.{ctx_module}.infrastructure"

            lines.append("    rule = (")
            lines.append("        Rule()")
            lines.append("        .modules_that()")
            lines.append(f'        .are_named("{infra_mod}")')
            lines.append("        .should_not()")
            lines.append("        .be_imported_by_modules_that()")
            lines.append(f'        .are_named("{domain_mod}")')
            lines.append("    )")
            lines.append("    rule.assert_applies(evaluable)")
            lines.append("")

        return "\n".join(lines)

    # -- Private helpers ------------------------------------------------------

    def _validate_module_boundaries(self) -> None:
        """Invariant 5: no contract may reference modules outside its BC.

        Raises:
            InvariantViolationError: If any contract references a module
                outside the expected prefix for its bounded context.
        """
        for contract in self._contracts:
            expected_prefix = f"{self.root_package}.{_snake_case(contract.context_name)}"
            for module in contract.modules:
                if not module.startswith(expected_prefix):
                    msg = (
                        f"Contract '{contract.name}' references module "
                        f"'{module}' outside its bounded context "
                        f"'{contract.context_name}' (expected prefix: "
                        f"'{expected_prefix}')"
                    )
                    raise InvariantViolationError(msg)
            for module in contract.forbidden_modules:
                if not module.startswith(expected_prefix):
                    msg = (
                        f"Contract '{contract.name}' references forbidden "
                        f"module '{module}' outside its bounded context "
                        f"'{contract.context_name}' (expected prefix: "
                        f"'{expected_prefix}')"
                    )
                    raise InvariantViolationError(msg)

    def _build_contract(
        self,
        contract_type: ContractType,
        context_name: str,
        module_prefix: str,
    ) -> Contract:
        """Build a single contract for a bounded context."""
        if contract_type == ContractType.LAYERS:
            return Contract(
                name=f"{context_name} DDD layer contract",
                contract_type=ContractType.LAYERS,
                context_name=context_name,
                modules=(
                    f"{module_prefix}.infrastructure",
                    f"{module_prefix}.application",
                    f"{module_prefix}.domain",
                ),
            )

        if contract_type == ContractType.FORBIDDEN:
            return Contract(
                name=f"{context_name} domain isolation",
                contract_type=ContractType.FORBIDDEN,
                context_name=context_name,
                modules=(f"{module_prefix}.domain",),
                forbidden_modules=(f"{module_prefix}.infrastructure",),
            )

        if contract_type == ContractType.INDEPENDENCE:
            return Contract(
                name=f"{context_name} independence",
                contract_type=ContractType.INDEPENDENCE,
                context_name=context_name,
                modules=(module_prefix,),
            )

        # ACYCLIC_SIBLINGS
        return Contract(
            name=f"{context_name} acyclic siblings",
            contract_type=ContractType.ACYCLIC_SIBLINGS,
            context_name=context_name,
            modules=(module_prefix,),
        )

    @staticmethod
    def _infer_strictness(contracts: list[Contract]) -> ContractStrictness:
        """Infer strictness level from contract types present."""
        types = {c.contract_type for c in contracts}
        if len(types) >= 4:
            return ContractStrictness.STRICT
        if ContractType.LAYERS in types:
            return ContractStrictness.MODERATE
        return ContractStrictness.MINIMAL
