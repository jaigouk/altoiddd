"""Tests for discovery domain events.

Verifies DiscoveryCompleted event immutability and field presence.
"""

from __future__ import annotations

import dataclasses

import pytest

from src.domain.events.discovery_events import DiscoveryCompleted
from src.domain.models.discovery_values import Answer, Persona, Playback, Register


class TestDiscoveryCompleted:
    def test_event_creation(self):
        event = DiscoveryCompleted(
            session_id="abc-123",
            persona=Persona.DEVELOPER,
            register=Register.TECHNICAL,
            answers=(Answer(question_id="Q1", response_text="Users"),),
            playback_confirmations=(Playback(summary_text="Sum", confirmed=True),),
        )
        assert event.session_id == "abc-123"
        assert event.persona == Persona.DEVELOPER
        assert event.register == Register.TECHNICAL
        assert len(event.answers) == 1
        assert len(event.playback_confirmations) == 1

    def test_event_is_frozen(self):
        event = DiscoveryCompleted(
            session_id="abc-123",
            persona=Persona.DEVELOPER,
            register=Register.TECHNICAL,
            answers=(),
            playback_confirmations=(),
        )
        with pytest.raises(dataclasses.FrozenInstanceError):
            event.session_id = "other"  # type: ignore[misc]

    def test_event_equality(self):
        kwargs = {
            "session_id": "abc-123",
            "persona": Persona.DEVELOPER,
            "register": Register.TECHNICAL,
            "answers": (Answer(question_id="Q1", response_text="Users"),),
            "playback_confirmations": (),
        }
        e1 = DiscoveryCompleted(**kwargs)  # type: ignore[arg-type]
        e2 = DiscoveryCompleted(**kwargs)  # type: ignore[arg-type]
        assert e1 == e2

    def test_event_has_all_required_fields(self):
        fields = {f.name for f in dataclasses.fields(DiscoveryCompleted)}
        assert fields == {
            "session_id",
            "persona",
            "register",
            "answers",
            "playback_confirmations",
            "tech_stack",
        }
