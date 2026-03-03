"""DiscoverySession aggregate root.

Manages the lifecycle of a guided DDD discovery session through a strict
state machine: CREATED -> PERSONA_DETECTED -> ANSWERING -> PLAYBACK_PENDING
-> COMPLETED (or CANCELLED).

Core invariants:
1. Cannot answer questions before persona is detected.
2. Phase order enforced (ACTORS -> STORY -> EVENTS -> BOUNDARIES).
3. Playback triggered after every 3 answers (must confirm before continuing).
4. Skip requires a non-empty reason.
5. Complete requires minimum 5 MVP questions answered.
"""

from __future__ import annotations

import enum
import uuid
from typing import TYPE_CHECKING

from src.domain.models.discovery_values import (
    Answer,
    Persona,
    Playback,
    QuestionPhase,
    Register,
)
from src.domain.models.errors import InvariantViolationError
from src.domain.models.question import Question
from src.domain.models.tech_stack import TechStack

if TYPE_CHECKING:
    from src.domain.events.discovery_events import DiscoveryCompleted

# Ordered phases for enforcement (SEED is implicit, not a question phase).
_QUESTION_PHASES: tuple[QuestionPhase, ...] = (
    QuestionPhase.ACTORS,
    QuestionPhase.STORY,
    QuestionPhase.EVENTS,
    QuestionPhase.BOUNDARIES,
)

_PERSONA_CHOICES: dict[str, tuple[Persona, Register]] = {
    "1": (Persona.DEVELOPER, Register.TECHNICAL),
    "2": (Persona.PRODUCT_OWNER, Register.NON_TECHNICAL),
    "3": (Persona.DOMAIN_EXPERT, Register.NON_TECHNICAL),
    "4": (Persona.MIXED, Register.NON_TECHNICAL),
}

# Map question IDs to their Question for fast lookup.
_QUESTION_BY_ID: dict[str, Question] = {q.id: q for q in Question.CATALOG}

# Number of answers between playback checkpoints.
_PLAYBACK_INTERVAL = 3


class DiscoveryStatus(enum.Enum):
    """States in the discovery session lifecycle."""

    CREATED = "created"
    PERSONA_DETECTED = "persona_detected"
    ANSWERING = "answering"
    PLAYBACK_PENDING = "playback_pending"
    COMPLETED = "completed"
    CANCELLED = "cancelled"


class DiscoverySession:
    """Aggregate root for the 10-question DDD discovery flow.

    Attributes:
        session_id: Unique identifier for this session.
        readme_content: The raw README text used to seed discovery.
    """

    def __init__(self, readme_content: str) -> None:
        self.session_id: str = str(uuid.uuid4())
        self.readme_content: str = readme_content
        self._status: DiscoveryStatus = DiscoveryStatus.CREATED
        self._persona: Persona | None = None
        self._register: Register | None = None
        self._answers: list[Answer] = []
        self._skipped: set[str] = set()
        self._playback_confirmations: list[Playback] = []
        self._answers_since_last_playback: int = 0
        self._tech_stack: TechStack | None = None
        self._events: list[DiscoveryCompleted] = []

    # -- Properties -----------------------------------------------------------

    @property
    def status(self) -> DiscoveryStatus:
        """Current session state."""
        return self._status

    @property
    def persona(self) -> Persona | None:
        """The detected user persona, or None if not yet detected."""
        return self._persona

    @property
    def register(self) -> Register | None:
        """The language register, or None if not yet determined."""
        return self._register

    @property
    def tech_stack(self) -> TechStack | None:
        """The user's chosen tech stack, or None if not yet set."""
        return self._tech_stack

    @property
    def answers(self) -> tuple[Answer, ...]:
        """All answers collected so far (defensive copy as tuple)."""
        return tuple(self._answers)

    @property
    def playback_confirmations(self) -> tuple[Playback, ...]:
        """All playback confirmations (defensive copy as tuple)."""
        return tuple(self._playback_confirmations)

    @property
    def current_phase(self) -> QuestionPhase:
        """The current discovery phase based on answered/skipped questions."""
        if not self._answers and not self._skipped:
            return QuestionPhase.SEED

        # Find the phase of the last answered or skipped question.
        all_handled = {a.question_id for a in self._answers} | self._skipped
        for phase in reversed(_QUESTION_PHASES):
            phase_questions = [q for q in Question.CATALOG if q.phase == phase]
            if all(q.id in all_handled for q in phase_questions):
                # All questions in this phase done; advance to the next phase.
                idx = _QUESTION_PHASES.index(phase)
                if idx + 1 < len(_QUESTION_PHASES):
                    return _QUESTION_PHASES[idx + 1]
                return phase  # Already at the last phase.

        # Some questions answered but no phase fully complete; return phase of
        # the earliest unanswered question.
        for phase in _QUESTION_PHASES:
            phase_questions = [q for q in Question.CATALOG if q.phase == phase]
            if not all(q.id in all_handled for q in phase_questions):
                return phase

        return _QUESTION_PHASES[-1]  # pragma: no cover

    @property
    def events(self) -> list[DiscoveryCompleted]:
        """Domain events produced by this aggregate (defensive copy)."""
        return list(self._events)

    # -- Commands -------------------------------------------------------------

    def set_tech_stack(self, tech_stack: TechStack) -> None:
        """Set the tech stack for this session.

        Allowed in CREATED or PERSONA_DETECTED states (pre-flight step before questions).

        Raises:
            InvariantViolationError: If not in CREATED or PERSONA_DETECTED state.
        """
        if self._status not in (DiscoveryStatus.CREATED, DiscoveryStatus.PERSONA_DETECTED):
            msg = (
                f"Can only set tech stack in CREATED or PERSONA_DETECTED state, "
                f"currently {self._status.value}"
            )
            raise InvariantViolationError(msg)
        self._tech_stack = tech_stack

    def detect_persona(self, choice: str) -> None:
        """Set the user persona and language register from a choice string.

        Args:
            choice: "1" for Developer/TECHNICAL, "2"-"4" for NON_TECHNICAL.

        Raises:
            InvariantViolationError: If not in CREATED state.
            ValueError: If choice is not "1"-"4".
        """
        if self._status != DiscoveryStatus.CREATED:
            msg = f"Can only detect persona in CREATED state, currently {self._status.value}"
            raise InvariantViolationError(msg)
        if choice not in _PERSONA_CHOICES:
            msg = f"Invalid persona choice '{choice}': must be '1', '2', '3', or '4'"
            raise ValueError(msg)
        self._persona, self._register = _PERSONA_CHOICES[choice]
        self._status = DiscoveryStatus.PERSONA_DETECTED

    def answer_question(self, question_id: str, response: str) -> None:
        """Record an answer to a discovery question.

        Args:
            question_id: The question ID (e.g. "Q1").
            response: The user's free-text response.

        Raises:
            InvariantViolationError: If persona not set, phase order violated,
                duplicate answer, or playback pending.
            ValueError: If response is empty or whitespace.
        """
        # Invariant 1: persona must be set.
        if self._status == DiscoveryStatus.CREATED:
            msg = "Cannot answer questions before persona is detected"
            raise InvariantViolationError(msg)

        # Invariant 3: cannot answer during playback.
        if self._status == DiscoveryStatus.PLAYBACK_PENDING:
            msg = "Must confirm playback before answering more questions"
            raise InvariantViolationError(msg)

        if self._status not in (DiscoveryStatus.PERSONA_DETECTED, DiscoveryStatus.ANSWERING):
            msg = f"Cannot answer in {self._status.value} state"
            raise InvariantViolationError(msg)

        # Validate response is non-empty.
        if not response.strip():
            msg = "Answer cannot be empty"
            raise ValueError(msg)

        # Check for duplicates.
        answered_ids = {a.question_id for a in self._answers}
        if question_id in answered_ids:
            msg = f"Question '{question_id}' already answered"
            raise InvariantViolationError(msg)

        # Invariant 2: phase order enforced.
        question = _QUESTION_BY_ID.get(question_id)
        if question is None:
            msg = f"Unknown question '{question_id}'"
            raise ValueError(msg)
        self._enforce_phase_order(question)

        # Record the answer.
        self._answers.append(Answer(question_id=question_id, response_text=response))
        self._answers_since_last_playback += 1
        self._status = DiscoveryStatus.ANSWERING

        # Invariant 3: trigger playback after every 3 answers.
        if self._answers_since_last_playback >= _PLAYBACK_INTERVAL:
            self._status = DiscoveryStatus.PLAYBACK_PENDING

    def skip_question(self, question_id: str, reason: str) -> None:
        """Skip a question with an explicit reason.

        Args:
            question_id: The question to skip.
            reason: Why it was skipped (must be non-empty).

        Raises:
            InvariantViolationError: If not in PERSONA_DETECTED or ANSWERING state,
                or if playback is pending.
            ValueError: If question_id is unknown or reason is empty.
        """
        if self._status == DiscoveryStatus.PLAYBACK_PENDING:
            msg = "Must confirm playback before skipping questions"
            raise InvariantViolationError(msg)
        if self._status not in (
            DiscoveryStatus.PERSONA_DETECTED,
            DiscoveryStatus.ANSWERING,
        ):
            msg = f"Cannot skip questions in {self._status.value} state"
            raise InvariantViolationError(msg)
        if question_id not in _QUESTION_BY_ID:
            msg = f"Unknown question '{question_id}'"
            raise ValueError(msg)
        if not reason.strip():
            msg = "Skip reason cannot be empty"
            raise ValueError(msg)
        self._skipped.add(question_id)

    def confirm_playback(self, confirmed: bool, corrections: str = "") -> None:
        """Confirm or reject a playback summary.

        Args:
            confirmed: True if the user accepts the summary.
            corrections: User corrections if not confirmed.

        Raises:
            InvariantViolationError: If not in PLAYBACK_PENDING state.
        """
        if self._status != DiscoveryStatus.PLAYBACK_PENDING:
            msg = (
                f"Can only confirm playback in PLAYBACK_PENDING state, "
                f"currently {self._status.value}"
            )
            raise InvariantViolationError(msg)
        self._playback_confirmations.append(
            Playback(
                summary_text=f"Playback {len(self._playback_confirmations) + 1}",
                confirmed=confirmed,
                corrections=corrections,
            )
        )
        self._answers_since_last_playback = 0
        self._status = DiscoveryStatus.ANSWERING

    def complete(self) -> None:
        """Complete the discovery session.

        Validates that minimum MVP questions have been answered and emits
        a DiscoveryCompleted event.

        Raises:
            InvariantViolationError: If not in ANSWERING state or insufficient
                MVP questions answered.
        """
        from src.domain.events.discovery_events import DiscoveryCompleted

        if self._status != DiscoveryStatus.ANSWERING:
            msg = f"Can only complete from ANSWERING state, currently {self._status.value}"
            raise InvariantViolationError(msg)

        # Invariant 5: minimum MVP questions answered.
        answered_ids = {a.question_id for a in self._answers}
        mvp_answered = answered_ids & Question.MVP_QUESTION_IDS
        if len(mvp_answered) < len(Question.MVP_QUESTION_IDS):
            missing = Question.MVP_QUESTION_IDS - mvp_answered
            msg = f"Cannot complete: MVP questions not answered: {sorted(missing)}"
            raise InvariantViolationError(msg)

        self._status = DiscoveryStatus.COMPLETED

        assert self._persona is not None  # Guaranteed by state machine.
        assert self._register is not None

        self._events.append(
            DiscoveryCompleted(
                session_id=self.session_id,
                persona=self._persona,
                register=self._register,
                answers=tuple(self._answers),
                playback_confirmations=tuple(self._playback_confirmations),
                tech_stack=self._tech_stack,
            )
        )

    # -- Serialization --------------------------------------------------------

    def to_snapshot(self) -> dict[str, object]:
        """Serialize session state to a JSON-serializable dict.

        Domain events (_events) are intentionally excluded. Events are
        transient side-effects dispatched by the application layer, not
        part of persisted state.
        """
        return {
            "session_id": self.session_id,
            "readme_content": self.readme_content,
            "status": self._status.value,
            "persona": self._persona.value if self._persona else None,
            "register": self._register.value if self._register else None,
            "answers": [
                {"question_id": a.question_id, "response_text": a.response_text}
                for a in self._answers
            ],
            "skipped": sorted(self._skipped),
            "playback_confirmations": [
                {
                    "summary_text": p.summary_text,
                    "confirmed": p.confirmed,
                    "corrections": p.corrections,
                }
                for p in self._playback_confirmations
            ],
            "answers_since_last_playback": self._answers_since_last_playback,
            "tech_stack": (
                {
                    "language": self._tech_stack.language,
                    "package_manager": self._tech_stack.package_manager,
                }
                if self._tech_stack
                else None
            ),
        }

    @classmethod
    def from_snapshot(cls, data: dict[str, object]) -> DiscoverySession:
        """Reconstruct a DiscoverySession from a snapshot dict.

        Raises:
            ValueError: If required fields are missing or contain invalid values.
        """
        cls._validate_snapshot_keys(data)
        status, persona, register, answers, skipped, playbacks, counter = (
            cls._parse_snapshot_fields(data)
        )
        cls._validate_snapshot_consistency(status, persona, counter)

        # Construct session bypassing __init__ to restore internal state
        session = object.__new__(cls)
        session.session_id = str(data["session_id"])
        session.readme_content = str(data["readme_content"])
        session._status = status
        session._persona = persona
        session._register = register
        session._answers = answers
        session._skipped = skipped
        session._playback_confirmations = playbacks
        session._answers_since_last_playback = counter
        session._events = []

        tech_stack_raw = data.get("tech_stack")
        if isinstance(tech_stack_raw, dict):
            session._tech_stack = TechStack(
                language=tech_stack_raw["language"],
                package_manager=tech_stack_raw["package_manager"],
            )
        else:
            session._tech_stack = None

        return session

    @staticmethod
    def _validate_snapshot_keys(data: dict[str, object]) -> None:
        """Verify all required keys are present."""
        required = {
            "session_id",
            "readme_content",
            "status",
            "persona",
            "register",
            "answers",
            "skipped",
            "playback_confirmations",
            "answers_since_last_playback",
        }
        missing = required - set(data.keys())
        if missing:
            msg = f"Snapshot missing required fields: {sorted(missing)}"
            raise ValueError(msg)

    @staticmethod
    def _parse_snapshot_fields(
        data: dict[str, object],
    ) -> tuple[
        DiscoveryStatus,
        Persona | None,
        Register | None,
        list[Answer],
        set[str],
        list[Playback],
        int,
    ]:
        """Parse and validate individual snapshot fields."""
        status = DiscoveryStatus(data["status"])

        persona_raw = data["persona"]
        persona = Persona(persona_raw) if persona_raw is not None else None

        register_raw = data["register"]
        register = Register(register_raw) if register_raw is not None else None

        answers_raw = data["answers"]
        if not isinstance(answers_raw, list):
            msg = "answers must be a list"
            raise ValueError(msg)
        answers = [
            Answer(question_id=a["question_id"], response_text=a["response_text"])
            for a in answers_raw
        ]

        skipped_raw = data["skipped"]
        if not isinstance(skipped_raw, list):
            msg = "skipped must be a list"
            raise ValueError(msg)
        skipped = set(skipped_raw)

        pb_raw = data["playback_confirmations"]
        if not isinstance(pb_raw, list):
            msg = "playback_confirmations must be a list"
            raise ValueError(msg)
        playbacks = [
            Playback(
                summary_text=p["summary_text"],
                confirmed=p["confirmed"],
                corrections=p.get("corrections", ""),
            )
            for p in pb_raw
        ]

        counter = data["answers_since_last_playback"]
        if not isinstance(counter, int) or counter < 0:
            msg = "answers_since_last_playback must be a non-negative integer"
            raise ValueError(msg)
        if counter > _PLAYBACK_INTERVAL:
            msg = (
                f"answers_since_last_playback ({counter}) "
                f"exceeds playback interval ({_PLAYBACK_INTERVAL})"
            )
            raise ValueError(msg)

        return status, persona, register, answers, skipped, playbacks, counter

    @staticmethod
    def _validate_snapshot_consistency(
        status: DiscoveryStatus,
        persona: Persona | None,
        counter: int,
    ) -> None:
        """Cross-validate status against persona and playback counter."""
        if status == DiscoveryStatus.CREATED and persona is not None:
            msg = "CREATED state must have persona=None"
            raise ValueError(msg)
        if status != DiscoveryStatus.CREATED and persona is None:
            msg = f"{status.value} state requires a persona"
            raise ValueError(msg)
        if status == DiscoveryStatus.PLAYBACK_PENDING and counter != _PLAYBACK_INTERVAL:
            msg = f"PLAYBACK_PENDING state requires counter={_PLAYBACK_INTERVAL}, got {counter}"
            raise ValueError(msg)
        if status == DiscoveryStatus.ANSWERING and counter >= _PLAYBACK_INTERVAL:
            msg = f"ANSWERING state requires counter < {_PLAYBACK_INTERVAL}, got {counter}"
            raise ValueError(msg)

    # -- Private helpers ------------------------------------------------------

    def _enforce_phase_order(self, question: Question) -> None:
        """Ensure the question belongs to the current or an earlier active phase.

        Raises:
            InvariantViolationError: If the question's phase has not been
                reached yet.
        """
        if question.phase not in _QUESTION_PHASES:
            return  # SEED phase questions (if any) are always allowed.

        target_idx = _QUESTION_PHASES.index(question.phase)
        all_handled = {a.question_id for a in self._answers} | self._skipped

        # Check that all earlier phases are fully handled.
        for i in range(target_idx):
            earlier_phase = _QUESTION_PHASES[i]
            phase_questions = [q for q in Question.CATALOG if q.phase == earlier_phase]
            for pq in phase_questions:
                if pq.id not in all_handled:
                    msg = (
                        f"Cannot answer {question.id} ({question.phase.value} phase) "
                        f"before completing {earlier_phase.value} phase "
                        f"(question {pq.id} not answered or skipped)"
                    )
                    raise InvariantViolationError(msg)
