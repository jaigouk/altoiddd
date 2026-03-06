"""Tests for BoundedContextCanvas value objects.

RED phase: all tests must FAIL because the module does not exist yet.
"""

from __future__ import annotations

import pytest


class TestDomainRoleEnum:
    """DomainRole enum values."""

    def test_execution_value(self) -> None:
        from src.domain.models.bounded_context_canvas import DomainRole

        assert DomainRole.EXECUTION.value == "execution"

    def test_analysis_value(self) -> None:
        from src.domain.models.bounded_context_canvas import DomainRole

        assert DomainRole.ANALYSIS.value == "analysis"

    def test_gateway_value(self) -> None:
        from src.domain.models.bounded_context_canvas import DomainRole

        assert DomainRole.GATEWAY.value == "gateway"

    def test_specification_value(self) -> None:
        from src.domain.models.bounded_context_canvas import DomainRole

        assert DomainRole.SPECIFICATION.value == "specification"

    def test_draft_value(self) -> None:
        from src.domain.models.bounded_context_canvas import DomainRole

        assert DomainRole.DRAFT.value == "draft"

    def test_has_five_members(self) -> None:
        from src.domain.models.bounded_context_canvas import DomainRole

        assert len(DomainRole) == 5


class TestStrategicClassification:
    """StrategicClassification frozen dataclass."""

    def test_construction(self) -> None:
        from src.domain.models.bounded_context_canvas import StrategicClassification
        from src.domain.models.domain_values import SubdomainClassification

        sc = StrategicClassification(
            domain=SubdomainClassification.CORE,
            business_model="Revenue",
            evolution="Genesis",
        )
        assert sc.domain == SubdomainClassification.CORE
        assert sc.business_model == "Revenue"
        assert sc.evolution == "Genesis"

    def test_frozen(self) -> None:
        from src.domain.models.bounded_context_canvas import StrategicClassification
        from src.domain.models.domain_values import SubdomainClassification

        sc = StrategicClassification(
            domain=SubdomainClassification.CORE,
            business_model="Revenue",
            evolution="Genesis",
        )
        with pytest.raises(AttributeError):
            sc.business_model = "Engagement"  # type: ignore[misc]

    def test_equality(self) -> None:
        from src.domain.models.bounded_context_canvas import StrategicClassification
        from src.domain.models.domain_values import SubdomainClassification

        a = StrategicClassification(
            domain=SubdomainClassification.CORE,
            business_model="Revenue",
            evolution="Genesis",
        )
        b = StrategicClassification(
            domain=SubdomainClassification.CORE,
            business_model="Revenue",
            evolution="Genesis",
        )
        assert a == b


class TestCommunicationMessage:
    """CommunicationMessage frozen dataclass."""

    def test_construction(self) -> None:
        from src.domain.models.bounded_context_canvas import CommunicationMessage

        msg = CommunicationMessage(
            message="PlaceOrder",
            message_type="Command",
            counterpart="Sales",
        )
        assert msg.message == "PlaceOrder"
        assert msg.message_type == "Command"
        assert msg.counterpart == "Sales"

    def test_frozen(self) -> None:
        from src.domain.models.bounded_context_canvas import CommunicationMessage

        msg = CommunicationMessage(
            message="PlaceOrder",
            message_type="Command",
            counterpart="Sales",
        )
        with pytest.raises(AttributeError):
            msg.message = "CancelOrder"  # type: ignore[misc]


class TestBoundedContextCanvas:
    """BoundedContextCanvas frozen dataclass."""

    def test_full_construction(self) -> None:
        from src.domain.models.bounded_context_canvas import (
            BoundedContextCanvas,
            CommunicationMessage,
            DomainRole,
            StrategicClassification,
        )
        from src.domain.models.domain_values import SubdomainClassification

        canvas = BoundedContextCanvas(
            context_name="Sales",
            purpose="Manages order lifecycle",
            classification=StrategicClassification(
                domain=SubdomainClassification.CORE,
                business_model="Revenue",
                evolution="Custom",
            ),
            domain_roles=(DomainRole.EXECUTION,),
            inbound_communication=(
                CommunicationMessage(
                    message="PlaceOrder",
                    message_type="Command",
                    counterpart="API Gateway",
                ),
            ),
            outbound_communication=(
                CommunicationMessage(
                    message="OrderPlaced",
                    message_type="Event",
                    counterpart="Fulfillment",
                ),
            ),
            ubiquitous_language=(("Order", "A purchase request"),),
            business_decisions=("Order must have items",),
            assumptions=(),
            open_questions=(),
        )
        assert canvas.context_name == "Sales"
        assert canvas.purpose == "Manages order lifecycle"
        assert len(canvas.domain_roles) == 1
        assert canvas.domain_roles[0] == DomainRole.EXECUTION
        assert len(canvas.inbound_communication) == 1
        assert len(canvas.outbound_communication) == 1
        assert canvas.ubiquitous_language == (("Order", "A purchase request"),)
        assert canvas.business_decisions == ("Order must have items",)
        assert canvas.assumptions == ()
        assert canvas.open_questions == ()

    def test_frozen(self) -> None:
        from src.domain.models.bounded_context_canvas import (
            BoundedContextCanvas,
            DomainRole,
            StrategicClassification,
        )
        from src.domain.models.domain_values import SubdomainClassification

        canvas = BoundedContextCanvas(
            context_name="Sales",
            purpose="Manages order lifecycle",
            classification=StrategicClassification(
                domain=SubdomainClassification.CORE,
                business_model="Revenue",
                evolution="Custom",
            ),
            domain_roles=(DomainRole.EXECUTION,),
            inbound_communication=(),
            outbound_communication=(),
            ubiquitous_language=(),
            business_decisions=(),
            assumptions=(),
            open_questions=(),
        )
        with pytest.raises(AttributeError):
            canvas.context_name = "Other"  # type: ignore[misc]

    def test_empty_communications(self) -> None:
        from src.domain.models.bounded_context_canvas import (
            BoundedContextCanvas,
            DomainRole,
            StrategicClassification,
        )
        from src.domain.models.domain_values import SubdomainClassification

        canvas = BoundedContextCanvas(
            context_name="Logging",
            purpose="Records events",
            classification=StrategicClassification(
                domain=SubdomainClassification.GENERIC,
                business_model="unclassified",
                evolution="Commodity",
            ),
            domain_roles=(DomainRole.GATEWAY,),
            inbound_communication=(),
            outbound_communication=(),
            ubiquitous_language=(),
            business_decisions=(),
            assumptions=(),
            open_questions=(),
        )
        assert canvas.inbound_communication == ()
        assert canvas.outbound_communication == ()

    def test_special_chars_in_name(self) -> None:
        """Names with special characters should be accepted."""
        from src.domain.models.bounded_context_canvas import (
            BoundedContextCanvas,
            DomainRole,
            StrategicClassification,
        )
        from src.domain.models.domain_values import SubdomainClassification

        canvas = BoundedContextCanvas(
            context_name='Auth & Identity "Service"',
            purpose="Handles authentication",
            classification=StrategicClassification(
                domain=SubdomainClassification.SUPPORTING,
                business_model="Compliance",
                evolution="Product",
            ),
            domain_roles=(DomainRole.SPECIFICATION,),
            inbound_communication=(),
            outbound_communication=(),
            ubiquitous_language=(),
            business_decisions=(),
            assumptions=(),
            open_questions=(),
        )
        assert canvas.context_name == 'Auth & Identity "Service"'

    def test_very_long_purpose(self) -> None:
        """500+ char purpose should not be truncated."""
        from src.domain.models.bounded_context_canvas import (
            BoundedContextCanvas,
            DomainRole,
            StrategicClassification,
        )
        from src.domain.models.domain_values import SubdomainClassification

        long_purpose = "A" * 600
        canvas = BoundedContextCanvas(
            context_name="Verbose",
            purpose=long_purpose,
            classification=StrategicClassification(
                domain=SubdomainClassification.CORE,
                business_model="Revenue",
                evolution="Genesis",
            ),
            domain_roles=(DomainRole.EXECUTION,),
            inbound_communication=(),
            outbound_communication=(),
            ubiquitous_language=(),
            business_decisions=(),
            assumptions=(),
            open_questions=(),
        )
        assert len(canvas.purpose) == 600
