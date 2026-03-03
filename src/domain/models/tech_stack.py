"""TechStack value object for the stack-aware pipeline.

TechStack captures the user's chosen tech stack (language + package manager).
This is the data half — persisted in session.json. The strategy half
(StackProfile) is reconstructed from TechStack at load time via a factory.
"""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class TechStack:
    """Value object representing the user's chosen tech stack.

    This is the data half of the stack-aware pipeline. TechStack is persisted
    in session.json. The strategy half (StackProfile) is reconstructed from
    TechStack at load time via a factory (Layer 2, 5li.4).
    """

    language: str          # "python", "unknown"
    package_manager: str   # "uv", ""
