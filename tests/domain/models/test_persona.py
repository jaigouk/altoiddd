"""Tests for persona domain models in the Tool Translation bounded context.

Covers PersonaType enum, Register enum, PersonaDefinition value object,
PERSONA_REGISTRY constant, and SUPPORTED_TOOLS / TOOL_TARGET_PATHS constants.
"""

from __future__ import annotations

import pytest

from src.domain.models.errors import InvariantViolationError

# ---------------------------------------------------------------------------
# 1. PersonaType enum
# ---------------------------------------------------------------------------


class TestPersonaType:
    def test_persona_type_has_five_values(self) -> None:
        from src.domain.models.persona import PersonaType

        assert len(PersonaType) == 5

    def test_persona_type_values(self) -> None:
        from src.domain.models.persona import PersonaType

        expected = {
            "solo_developer",
            "team_lead",
            "ai_tool_switcher",
            "product_owner",
            "domain_expert",
        }
        actual = {member.value for member in PersonaType}
        assert actual == expected


# ---------------------------------------------------------------------------
# 2. Register enum
# ---------------------------------------------------------------------------


class TestRegister:
    def test_register_has_two_values(self) -> None:
        from src.domain.models.persona import Register

        assert len(Register) == 2

    def test_register_values(self) -> None:
        from src.domain.models.persona import Register

        assert Register.TECHNICAL.value == "technical"
        assert Register.NON_TECHNICAL.value == "non_technical"


# ---------------------------------------------------------------------------
# 3. PersonaDefinition value object
# ---------------------------------------------------------------------------


class TestPersonaDefinition:
    def test_persona_definition_frozen(self) -> None:
        from src.domain.models.persona import PersonaDefinition, PersonaType, Register

        defn = PersonaDefinition(
            name="Test",
            persona_type=PersonaType.SOLO_DEVELOPER,
            register=Register.TECHNICAL,
            description="A test persona",
            instructions_template="# Test\n\nInstructions here.",
        )

        with pytest.raises(AttributeError):
            defn.name = "Changed"  # type: ignore[misc]

    def test_persona_definition_fields(self) -> None:
        from src.domain.models.persona import PersonaDefinition, PersonaType, Register

        defn = PersonaDefinition(
            name="Solo Developer",
            persona_type=PersonaType.SOLO_DEVELOPER,
            register=Register.TECHNICAL,
            description="Full-stack developer",
            instructions_template="# Solo\n\nInstructions.",
        )

        assert defn.name == "Solo Developer"
        assert defn.persona_type == PersonaType.SOLO_DEVELOPER
        assert defn.register == Register.TECHNICAL
        assert defn.description == "Full-stack developer"
        assert defn.instructions_template == "# Solo\n\nInstructions."

    def test_persona_definition_rejects_empty_name(self) -> None:
        from src.domain.models.persona import PersonaDefinition, PersonaType, Register

        with pytest.raises(InvariantViolationError, match="name cannot be empty"):
            PersonaDefinition(
                name="",
                persona_type=PersonaType.SOLO_DEVELOPER,
                register=Register.TECHNICAL,
                description="A test persona",
                instructions_template="# Test\n\nInstructions.",
            )

    def test_persona_definition_rejects_whitespace_name(self) -> None:
        from src.domain.models.persona import PersonaDefinition, PersonaType, Register

        with pytest.raises(InvariantViolationError, match="name cannot be empty"):
            PersonaDefinition(
                name="   ",
                persona_type=PersonaType.SOLO_DEVELOPER,
                register=Register.TECHNICAL,
                description="A test persona",
                instructions_template="# Test\n\nInstructions.",
            )

    def test_persona_definition_rejects_empty_template(self) -> None:
        from src.domain.models.persona import PersonaDefinition, PersonaType, Register

        with pytest.raises(InvariantViolationError, match="instructions template cannot be empty"):
            PersonaDefinition(
                name="Test",
                persona_type=PersonaType.SOLO_DEVELOPER,
                register=Register.TECHNICAL,
                description="A test persona",
                instructions_template="",
            )

    def test_persona_definition_rejects_whitespace_template(self) -> None:
        from src.domain.models.persona import PersonaDefinition, PersonaType, Register

        with pytest.raises(InvariantViolationError, match="instructions template cannot be empty"):
            PersonaDefinition(
                name="Test",
                persona_type=PersonaType.SOLO_DEVELOPER,
                register=Register.TECHNICAL,
                description="A test persona",
                instructions_template="   ",
            )


# ---------------------------------------------------------------------------
# 4. Register assignments in PERSONA_REGISTRY
# ---------------------------------------------------------------------------


class TestRegistryRegisters:
    def test_technical_register_for_solo_developer(self) -> None:
        from src.domain.models.persona import PERSONA_REGISTRY, PersonaType, Register

        assert PERSONA_REGISTRY[PersonaType.SOLO_DEVELOPER].register == Register.TECHNICAL

    def test_technical_register_for_team_lead(self) -> None:
        from src.domain.models.persona import PERSONA_REGISTRY, PersonaType, Register

        assert PERSONA_REGISTRY[PersonaType.TEAM_LEAD].register == Register.TECHNICAL

    def test_technical_register_for_ai_tool_switcher(self) -> None:
        from src.domain.models.persona import PERSONA_REGISTRY, PersonaType, Register

        assert PERSONA_REGISTRY[PersonaType.AI_TOOL_SWITCHER].register == Register.TECHNICAL

    def test_non_technical_register_for_product_owner(self) -> None:
        from src.domain.models.persona import PERSONA_REGISTRY, PersonaType, Register

        assert PERSONA_REGISTRY[PersonaType.PRODUCT_OWNER].register == Register.NON_TECHNICAL

    def test_non_technical_register_for_domain_expert(self) -> None:
        from src.domain.models.persona import PERSONA_REGISTRY, PersonaType, Register

        assert PERSONA_REGISTRY[PersonaType.DOMAIN_EXPERT].register == Register.NON_TECHNICAL


# ---------------------------------------------------------------------------
# 5. PERSONA_REGISTRY completeness
# ---------------------------------------------------------------------------


class TestPersonaRegistry:
    def test_persona_registry_has_five_entries(self) -> None:
        from src.domain.models.persona import PERSONA_REGISTRY

        assert len(PERSONA_REGISTRY) == 5

    def test_persona_registry_keys_match_persona_type(self) -> None:
        from src.domain.models.persona import PERSONA_REGISTRY, PersonaType

        assert set(PERSONA_REGISTRY.keys()) == set(PersonaType)

    def test_each_registry_entry_has_matching_persona_type(self) -> None:
        from src.domain.models.persona import PERSONA_REGISTRY

        for key, defn in PERSONA_REGISTRY.items():
            assert defn.persona_type == key


# ---------------------------------------------------------------------------
# 6. SUPPORTED_TOOLS and TOOL_TARGET_PATHS
# ---------------------------------------------------------------------------


class TestSupportedToolsAndPaths:
    def test_supported_tools_list(self) -> None:
        from src.domain.models.persona import SUPPORTED_TOOLS

        assert SUPPORTED_TOOLS == ("claude-code", "cursor", "roo-code", "opencode")

    def test_tool_target_paths(self) -> None:
        from src.domain.models.persona import TOOL_TARGET_PATHS

        assert "claude-code" in TOOL_TARGET_PATHS
        assert "cursor" in TOOL_TARGET_PATHS
        assert "roo-code" in TOOL_TARGET_PATHS
        assert "opencode" in TOOL_TARGET_PATHS

    def test_tool_target_paths_contain_name_placeholder(self) -> None:
        from src.domain.models.persona import TOOL_TARGET_PATHS

        for tool, path in TOOL_TARGET_PATHS.items():
            assert "{name}" in path, f"Path for {tool} must contain {{name}} placeholder"
