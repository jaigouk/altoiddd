"""Tests for UbiquitousLanguage entity."""

from __future__ import annotations

import pytest

from src.domain.models.ubiquitous_language import UbiquitousLanguage


class TestAddTerm:
    def test_add_single_term(self) -> None:
        ul = UbiquitousLanguage()
        ul.add_term("Order", "A customer purchase", "Sales")
        assert len(ul.terms) == 1
        assert ul.terms[0].term == "Order"

    def test_add_multiple_terms(self) -> None:
        ul = UbiquitousLanguage()
        ul.add_term("Order", "A purchase", "Sales")
        ul.add_term("Product", "An item for sale", "Catalog")
        assert len(ul.terms) == 2

    def test_empty_term_raises(self) -> None:
        ul = UbiquitousLanguage()
        with pytest.raises(ValueError, match="Term cannot be empty"):
            ul.add_term("", "Definition", "Context")

    def test_whitespace_term_raises(self) -> None:
        ul = UbiquitousLanguage()
        with pytest.raises(ValueError, match="Term cannot be empty"):
            ul.add_term("   ", "Definition", "Context")

    def test_empty_definition_raises(self) -> None:
        ul = UbiquitousLanguage()
        with pytest.raises(ValueError, match="Definition cannot be empty"):
            ul.add_term("Order", "", "Context")

    def test_term_stripped(self) -> None:
        ul = UbiquitousLanguage()
        ul.add_term("  Order  ", "A purchase", "Sales")
        assert ul.terms[0].term == "Order"

    def test_source_question_ids(self) -> None:
        ul = UbiquitousLanguage()
        ul.add_term("Order", "A purchase", "Sales", source_question_ids=("Q1", "Q2"))
        assert ul.terms[0].source_question_ids == ("Q1", "Q2")


class TestGetTermsForContext:
    def test_filter_by_context(self) -> None:
        ul = UbiquitousLanguage()
        ul.add_term("Order", "A purchase", "Sales")
        ul.add_term("Product", "An item", "Catalog")
        ul.add_term("Invoice", "A bill", "Sales")

        sales_terms = ul.get_terms_for_context("Sales")
        assert len(sales_terms) == 2
        assert {t.term for t in sales_terms} == {"Order", "Invoice"}

    def test_no_terms_in_context(self) -> None:
        ul = UbiquitousLanguage()
        ul.add_term("Order", "A purchase", "Sales")
        assert ul.get_terms_for_context("Unknown") == ()


class TestFindAmbiguousTerms:
    def test_no_ambiguity(self) -> None:
        ul = UbiquitousLanguage()
        ul.add_term("Order", "A purchase", "Sales")
        ul.add_term("Product", "An item", "Catalog")
        assert ul.find_ambiguous_terms() == ()

    def test_detects_ambiguity(self) -> None:
        ul = UbiquitousLanguage()
        ul.add_term("Config", "App settings", "Bootstrap")
        ul.add_term("Config", "Tool configuration", "Tool Translation")
        assert ul.find_ambiguous_terms() == ("config",)

    def test_case_insensitive(self) -> None:
        ul = UbiquitousLanguage()
        ul.add_term("Order", "Sales order", "Sales")
        ul.add_term("order", "Work order", "Manufacturing")
        assert ul.find_ambiguous_terms() == ("order",)


class TestHasPerContextDefinitions:
    def test_ambiguous_with_definitions(self) -> None:
        ul = UbiquitousLanguage()
        ul.add_term("Config", "App settings", "Bootstrap")
        ul.add_term("Config", "Tool config", "Tool Translation")
        assert ul.has_per_context_definitions("Config") is True

    def test_ambiguous_missing_definition(self) -> None:
        ul = UbiquitousLanguage()
        ul.add_term("Config", "App settings", "Bootstrap")
        # Bypass add_term validation to simulate a term with empty definition.
        from src.domain.models.ubiquitous_language import TermEntry

        ul._terms.append(TermEntry(term="Config", definition="", context_name="Tool Translation"))
        assert ul.has_per_context_definitions("Config") is False


class TestAllTermNames:
    def test_returns_frozenset(self) -> None:
        ul = UbiquitousLanguage()
        ul.add_term("Order", "A purchase", "Sales")
        ul.add_term("Product", "An item", "Catalog")
        assert ul.all_term_names == frozenset({"order", "product"})

    def test_empty(self) -> None:
        ul = UbiquitousLanguage()
        assert ul.all_term_names == frozenset()


class TestTermsProperty:
    def test_defensive_copy(self) -> None:
        ul = UbiquitousLanguage()
        ul.add_term("Order", "A purchase", "Sales")
        terms1 = ul.terms
        terms2 = ul.terms
        assert terms1 == terms2
        assert terms1 is not terms2  # Different tuple instances.
