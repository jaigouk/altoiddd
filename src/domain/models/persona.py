"""Domain models for the Tool Translation bounded context -- persona management.

PersonaType enumerates the agent persona archetypes.
Register classifies personas as technical or non-technical.
PersonaDefinition is a frozen value object describing one agent persona.
PERSONA_REGISTRY maps every PersonaType to its canonical PersonaDefinition.
SUPPORTED_TOOLS and TOOL_TARGET_PATHS define valid tool identifiers and
their target file path templates.
"""

from __future__ import annotations

import enum
from dataclasses import dataclass

from src.domain.models.errors import InvariantViolationError


class PersonaType(enum.Enum):
    """Agent persona archetypes supported by alty."""

    SOLO_DEVELOPER = "solo_developer"
    TEAM_LEAD = "team_lead"
    AI_TOOL_SWITCHER = "ai_tool_switcher"
    PRODUCT_OWNER = "product_owner"
    DOMAIN_EXPERT = "domain_expert"


class Register(enum.Enum):
    """Communication register for a persona -- technical or non-technical."""

    TECHNICAL = "technical"
    NON_TECHNICAL = "non_technical"


@dataclass(frozen=True)
class PersonaDefinition:
    """Immutable value object describing one agent persona.

    Attributes:
        name: Human-readable persona name.
        persona_type: The archetype this persona represents.
        register: Communication register (technical / non-technical).
        description: Short description of the persona's role.
        instructions_template: Template text for the persona's agent instructions.
    """

    name: str
    persona_type: PersonaType
    register: Register
    description: str
    instructions_template: str

    def __post_init__(self) -> None:
        if not self.name or not self.name.strip():
            raise InvariantViolationError("Persona name cannot be empty")
        if not self.instructions_template or not self.instructions_template.strip():
            raise InvariantViolationError("Persona instructions template cannot be empty")


PERSONA_REGISTRY: dict[PersonaType, PersonaDefinition] = {
    PersonaType.SOLO_DEVELOPER: PersonaDefinition(
        name="Solo Developer",
        persona_type=PersonaType.SOLO_DEVELOPER,
        register=Register.TECHNICAL,
        description="Full-stack developer following DDD + TDD + SOLID principles",
        instructions_template=(
            "# Solo Developer Agent\n\n"
            "You are a solo developer agent responsible for implementing features\n"
            "using DDD, TDD, and SOLID principles. Follow Red/Green/Refactor strictly.\n"
            "Write failing tests first, then minimal code to pass, then refactor."
        ),
    ),
    PersonaType.TEAM_LEAD: PersonaDefinition(
        name="Team Lead",
        persona_type=PersonaType.TEAM_LEAD,
        register=Register.TECHNICAL,
        description="Technical lead responsible for architecture and code quality",
        instructions_template=(
            "# Team Lead Agent\n\n"
            "You are a team lead agent responsible for architecture review\n"
            "and DDD compliance. Ensure bounded context boundaries are respected\n"
            "and code quality meets project standards."
        ),
    ),
    PersonaType.AI_TOOL_SWITCHER: PersonaDefinition(
        name="AI Tool Switcher",
        persona_type=PersonaType.AI_TOOL_SWITCHER,
        register=Register.TECHNICAL,
        description="Agent specialized in cross-tool configuration and migration",
        instructions_template=(
            "# AI Tool Switcher Agent\n\n"
            "You are an AI tool switcher agent responsible for cross-tool\n"
            "configuration mapping and config translation. Ensure consistent\n"
            "behavior across supported AI coding tools."
        ),
    ),
    PersonaType.PRODUCT_OWNER: PersonaDefinition(
        name="Product Owner",
        persona_type=PersonaType.PRODUCT_OWNER,
        register=Register.NON_TECHNICAL,
        description="Business-focused agent for requirements and prioritization",
        instructions_template=(
            "# Product Owner Agent\n\n"
            "You are a product owner agent focused on business requirements\n"
            "and prioritization. Use business language and ensure requirements\n"
            "are captured clearly for the development team."
        ),
    ),
    PersonaType.DOMAIN_EXPERT: PersonaDefinition(
        name="Domain Expert",
        persona_type=PersonaType.DOMAIN_EXPERT,
        register=Register.NON_TECHNICAL,
        description="Domain knowledge specialist for business rules and terminology",
        instructions_template=(
            "# Domain Expert Agent\n\n"
            "You are a domain expert agent responsible for business rules\n"
            "and domain terminology. Ensure ubiquitous language is used correctly\n"
            "and domain invariants are properly captured."
        ),
    ),
}


SUPPORTED_TOOLS: tuple[str, ...] = ("claude-code", "cursor", "roo-code", "opencode")

TOOL_TARGET_PATHS: dict[str, str] = {
    "claude-code": ".claude/agents/{name}.md",
    "cursor": ".cursor/rules/{name}.mdc",
    "roo-code": ".roo-code/modes/{name}.md",
    "opencode": ".opencode/agents/{name}.md",
}
