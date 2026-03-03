"""Domain service: resolve TechStack to StackProfile.

Stateless factory that maps a TechStack value object to the appropriate
StackProfile strategy. No I/O, no side effects.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.domain.models.stack_profile import GenericProfile, PythonUvProfile

if TYPE_CHECKING:
    from src.domain.models.stack_profile import StackProfile
    from src.domain.models.tech_stack import TechStack


def resolve_profile(tech_stack: TechStack | None) -> StackProfile:
    """Map a TechStack to the corresponding StackProfile.

    Args:
        tech_stack: The tech stack from a discovery session, or None for
            old sessions that predate tech stack detection.

    Returns:
        PythonUvProfile if language is "python", GenericProfile otherwise.
    """
    if tech_stack is not None and tech_stack.language == "python":
        return PythonUvProfile()
    return GenericProfile()
