package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/alto-cli/alto/internal/discovery/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// -- WorkObjectType tests --

func TestNewWorkObjectType_Valid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected domain.WorkObjectType
	}{
		{"document", "document", domain.WorkObjectTypeDocument},
		{"folder", "folder", domain.WorkObjectTypeFolder},
		{"call", "call", domain.WorkObjectTypeCall},
		{"email", "email", domain.WorkObjectTypeEmail},
		{"conversation", "conversation", domain.WorkObjectTypeConversation},
		{"info", "info", domain.WorkObjectTypeInfo},
		{"data", "data", domain.WorkObjectTypeData},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			wot, err := domain.NewWorkObjectType(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, wot)
		})
	}
}

func TestNewWorkObjectType_Invalid(t *testing.T) {
	t.Parallel()
	_, err := domain.NewWorkObjectType("spreadsheet")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewWorkObjectType_Empty(t *testing.T) {
	t.Parallel()
	_, err := domain.NewWorkObjectType("")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestWorkObjectType_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		wot      domain.WorkObjectType
		expected string
	}{
		{"document", domain.WorkObjectTypeDocument, "document"},
		{"folder", domain.WorkObjectTypeFolder, "folder"},
		{"call", domain.WorkObjectTypeCall, "call"},
		{"email", domain.WorkObjectTypeEmail, "email"},
		{"conversation", domain.WorkObjectTypeConversation, "conversation"},
		{"info", domain.WorkObjectTypeInfo, "info"},
		{"data", domain.WorkObjectTypeData, "data"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.wot.String())
		})
	}
}

func TestWorkObjectType_TextRoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		wot  domain.WorkObjectType
	}{
		{"document", domain.WorkObjectTypeDocument},
		{"folder", domain.WorkObjectTypeFolder},
		{"call", domain.WorkObjectTypeCall},
		{"email", domain.WorkObjectTypeEmail},
		{"conversation", domain.WorkObjectTypeConversation},
		{"info", domain.WorkObjectTypeInfo},
		{"data", domain.WorkObjectTypeData},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := tt.wot.MarshalText()
			require.NoError(t, err)

			var got domain.WorkObjectType
			err = got.UnmarshalText(data)
			require.NoError(t, err)
			assert.Equal(t, tt.wot, got)
		})
	}
}

func TestWorkObjectType_MarshalText_Invalid(t *testing.T) {
	t.Parallel()
	invalid := domain.WorkObjectType("bad")
	_, err := invalid.MarshalText()
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestWorkObjectType_UnmarshalText_Invalid(t *testing.T) {
	t.Parallel()
	var wot domain.WorkObjectType
	err := wot.UnmarshalText([]byte("nonsense"))
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestWorkObjectType_UnmarshalText_Empty(t *testing.T) {
	t.Parallel()
	var wot domain.WorkObjectType
	err := wot.UnmarshalText([]byte(""))
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestAllWorkObjectTypes(t *testing.T) {
	t.Parallel()
	assert.Len(t, domain.AllWorkObjectTypes(), 7)
}

// -- WorkObject tests --

func TestNewWorkObject_Valid(t *testing.T) {
	t.Parallel()
	wo, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	assert.Equal(t, "Order", wo.Name())
	assert.Equal(t, domain.WorkObjectTypeDocument, wo.Type())
	assert.Equal(t, vo.UserStated, wo.Trust())
	assert.Empty(t, wo.Source())
}

func TestNewWorkObject_EmptyName(t *testing.T) {
	t.Parallel()
	_, err := domain.NewWorkObject("", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewWorkObject_WhitespaceName(t *testing.T) {
	t.Parallel()
	_, err := domain.NewWorkObject("   ", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewWorkObject_InvalidType(t *testing.T) {
	t.Parallel()
	_, err := domain.NewWorkObject("Order", domain.WorkObjectType("invalid"), vo.UserStated, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewWorkObject_InvalidTrust(t *testing.T) {
	t.Parallel()
	_, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.TrustLevel(99), "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewWorkObject_AIResearchedWithoutSource(t *testing.T) {
	t.Parallel()
	_, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.AIResearched, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewWorkObject_AIResearchedWithSource(t *testing.T) {
	t.Parallel()
	wo, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.AIResearched, "ref")
	require.NoError(t, err)
	assert.Equal(t, "ref", wo.Source())
}

func TestNewWorkObject_UserStatedWithoutSource(t *testing.T) {
	t.Parallel()
	wo, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	assert.Empty(t, wo.Source())
}

func TestWorkObject_Equals_CaseInsensitive(t *testing.T) {
	t.Parallel()
	wo1, err := domain.NewWorkObject("order", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	wo2, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	assert.True(t, wo1.Equals(wo2))
}

func TestWorkObject_Equals_DifferentName(t *testing.T) {
	t.Parallel()
	wo1, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	wo2, err := domain.NewWorkObject("Invoice", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	assert.False(t, wo1.Equals(wo2))
}

func TestWorkObject_String(t *testing.T) {
	t.Parallel()
	wo, err := domain.NewWorkObject("Order", domain.WorkObjectTypeDocument, vo.UserStated, "")
	require.NoError(t, err)
	assert.Equal(t, "WorkObject: Order (document, user_stated)", wo.String())
}
