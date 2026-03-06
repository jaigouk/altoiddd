"""Value objects for the Bounded Context Canvas.

BoundedContextCanvas, StrategicClassification, DomainRole, and
CommunicationMessage are frozen dataclasses following the ddd-crew v5
canvas format.
"""

from __future__ import annotations

import enum
from dataclasses import dataclass
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from src.domain.models.domain_values import SubdomainClassification


class DomainRole(enum.Enum):
    """Role a bounded context plays in the domain."""

    EXECUTION = "execution"
    ANALYSIS = "analysis"
    GATEWAY = "gateway"
    SPECIFICATION = "specification"
    DRAFT = "draft"


@dataclass(frozen=True)
class StrategicClassification:
    """Strategic classification of a bounded context.

    Attributes:
        domain: Core/Supporting/Generic subdomain classification.
        business_model: Business model contribution (e.g. "Revenue", "Compliance").
        evolution: Wardley evolution stage (e.g. "Genesis", "Custom", "Product", "Commodity").
    """

    domain: SubdomainClassification
    business_model: str
    evolution: str


@dataclass(frozen=True)
class CommunicationMessage:
    """A message flowing into or out of a bounded context.

    Attributes:
        message: Message name (e.g. "PlaceOrder", "OrderPlaced").
        message_type: "Command", "Query", or "Event".
        counterpart: Sender (inbound) or receiver (outbound) context name.
    """

    message: str
    message_type: str
    counterpart: str


@dataclass(frozen=True)
class BoundedContextCanvas:
    """A Bounded Context Canvas following the ddd-crew v5 format.

    Attributes:
        context_name: Name of the bounded context.
        purpose: What this context does and why it exists.
        classification: Strategic classification (domain, business model, evolution).
        domain_roles: Roles this context plays.
        inbound_communication: Messages received by this context.
        outbound_communication: Messages sent by this context.
        ubiquitous_language: (term, definition) pairs for this context.
        business_decisions: Business rules and invariants.
        assumptions: Assumptions made about this context (Round 2).
        open_questions: Unresolved questions (Round 2).
    """

    context_name: str
    purpose: str
    classification: StrategicClassification
    domain_roles: tuple[DomainRole, ...]
    inbound_communication: tuple[CommunicationMessage, ...]
    outbound_communication: tuple[CommunicationMessage, ...]
    ubiquitous_language: tuple[tuple[str, str], ...]
    business_decisions: tuple[str, ...]
    assumptions: tuple[str, ...]
    open_questions: tuple[str, ...]
