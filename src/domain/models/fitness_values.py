"""Value objects for the Architecture Testing bounded context.

ContractStrictness maps SubdomainClassification to treatment levels.
ContractType enumerates import-linter contract types.
Contract and ArchRule are frozen dataclasses representing generated
architecture fitness function artifacts.
"""

from __future__ import annotations

import enum
from dataclasses import dataclass
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from src.domain.models.domain_values import SubdomainClassification


class ContractType(enum.Enum):
    """Import-linter contract types used for architecture boundary enforcement."""

    LAYERS = "layers"
    FORBIDDEN = "forbidden"
    INDEPENDENCE = "independence"
    ACYCLIC_SIBLINGS = "acyclic_siblings"


class ContractStrictness(enum.Enum):
    """Treatment level mapped from SubdomainClassification.

    STRICT (Core):      All 4 contract types — maximum enforcement.
    MODERATE (Supporting): Layers + forbidden — standard enforcement.
    MINIMAL (Generic):  Forbidden only — ACL boundary enforcement.
    """

    STRICT = "strict"
    MODERATE = "moderate"
    MINIMAL = "minimal"

    @staticmethod
    def from_classification(classification: SubdomainClassification) -> ContractStrictness:
        """Map a SubdomainClassification to its ContractStrictness level."""
        from src.domain.models.domain_values import SubdomainClassification

        mapping = {
            SubdomainClassification.CORE: ContractStrictness.STRICT,
            SubdomainClassification.SUPPORTING: ContractStrictness.MODERATE,
            SubdomainClassification.GENERIC: ContractStrictness.MINIMAL,
        }
        return mapping[classification]

    def required_contract_types(self) -> tuple[ContractType, ...]:
        """Return the contract types required for this strictness level."""
        if self == ContractStrictness.STRICT:
            return (
                ContractType.LAYERS,
                ContractType.FORBIDDEN,
                ContractType.INDEPENDENCE,
                ContractType.ACYCLIC_SIBLINGS,
            )
        if self == ContractStrictness.MODERATE:
            return (ContractType.LAYERS, ContractType.FORBIDDEN)
        return (ContractType.FORBIDDEN,)


@dataclass(frozen=True)
class Contract:
    """An import-linter contract for architecture boundary enforcement.

    Attributes:
        name: Human-readable contract name.
        contract_type: Type of import-linter contract.
        context_name: Bounded context this contract belongs to.
        modules: Module paths for the contract (meaning depends on type).
        forbidden_modules: For FORBIDDEN contracts, the modules being blocked.
    """

    name: str
    contract_type: ContractType
    context_name: str
    modules: tuple[str, ...]
    forbidden_modules: tuple[str, ...] = ()


@dataclass(frozen=True)
class ArchRule:
    """A pytestarch rule for architecture boundary enforcement.

    Attributes:
        name: Human-readable rule name.
        assertion: The pytestarch assertion as a description string.
        context_name: Bounded context this rule belongs to.
    """

    name: str
    assertion: str
    context_name: str
