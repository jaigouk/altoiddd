"""Tests for Question entity and catalog.

Verifies the 10-question catalog, MVP question IDs, and dual-register text.
"""

from __future__ import annotations

from src.domain.models.discovery_values import QuestionPhase
from src.domain.models.question import Question


class TestQuestionEntity:
    def test_question_has_required_fields(self):
        q = Question(
            id="Q1",
            phase=QuestionPhase.ACTORS,
            technical_text="Who are the actors in the system?",
            non_technical_text="Who will use this product?",
            produces=("actors",),
        )
        assert q.id == "Q1"
        assert q.phase == QuestionPhase.ACTORS
        assert q.technical_text == "Who are the actors in the system?"
        assert q.non_technical_text == "Who will use this product?"
        assert q.produces == ("actors",)


class TestQuestionCatalog:
    def test_catalog_has_ten_questions(self):
        assert len(Question.CATALOG) == 10

    def test_catalog_ids_are_q1_through_q10(self):
        ids = [q.id for q in Question.CATALOG]
        assert ids == [f"Q{i}" for i in range(1, 11)]

    def test_catalog_phase_assignments(self):
        """Verify each question is in the correct phase."""
        phase_map = {q.id: q.phase for q in Question.CATALOG}
        assert phase_map["Q1"] == QuestionPhase.ACTORS
        assert phase_map["Q2"] == QuestionPhase.ACTORS
        assert phase_map["Q3"] == QuestionPhase.STORY
        assert phase_map["Q4"] == QuestionPhase.STORY
        assert phase_map["Q5"] == QuestionPhase.STORY
        assert phase_map["Q6"] == QuestionPhase.EVENTS
        assert phase_map["Q7"] == QuestionPhase.EVENTS
        assert phase_map["Q8"] == QuestionPhase.EVENTS
        assert phase_map["Q9"] == QuestionPhase.BOUNDARIES
        assert phase_map["Q10"] == QuestionPhase.BOUNDARIES

    def test_each_question_has_both_register_texts(self):
        for q in Question.CATALOG:
            assert q.technical_text, f"{q.id} missing technical_text"
            assert q.non_technical_text, f"{q.id} missing non_technical_text"

    def test_each_question_has_produces(self):
        for q in Question.CATALOG:
            assert len(q.produces) > 0, f"{q.id} has empty produces"

    def test_dual_register_texts_differ(self):
        """Technical and non-technical texts should differ for every question."""
        for q in Question.CATALOG:
            assert q.technical_text != q.non_technical_text, f"{q.id} has identical register texts"


class TestMVPQuestionIds:
    def test_mvp_ids_contains_five_questions(self):
        assert len(Question.MVP_QUESTION_IDS) == 5

    def test_mvp_ids_are_correct(self):
        assert frozenset({"Q1", "Q3", "Q4", "Q9", "Q10"}) == Question.MVP_QUESTION_IDS

    def test_mvp_ids_are_subset_of_catalog(self):
        catalog_ids = {q.id for q in Question.CATALOG}
        assert Question.MVP_QUESTION_IDS.issubset(catalog_ids)
