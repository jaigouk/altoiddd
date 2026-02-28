"""Domain events for the Bootstrap bounded context.

BootstrapCompleted is emitted when a bootstrap session finishes executing
all planned file actions.
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from pathlib import Path


@dataclass(frozen=True)
class BootstrapCompleted:
    """Emitted when a bootstrap session completes execution.

    Attributes:
        session_id: Unique identifier of the completed session.
        project_dir: The directory that was bootstrapped.
    """

    session_id: str
    project_dir: Path
