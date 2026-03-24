package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/alto-cli/alto/internal/discovery/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// -- ActorType tests --

func TestNewActorType_Valid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected domain.ActorType
	}{
		{"person", "person", domain.ActorTypePerson},
		{"system", "system", domain.ActorTypeSystem},
		{"group", "group", domain.ActorTypeGroup},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			at, err := domain.NewActorType(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, at)
		})
	}
}

func TestNewActorType_Invalid(t *testing.T) {
	t.Parallel()
	_, err := domain.NewActorType("robot")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewActorType_Empty(t *testing.T) {
	t.Parallel()
	_, err := domain.NewActorType("")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestActorType_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		at       domain.ActorType
		expected string
	}{
		{"person", domain.ActorTypePerson, "person"},
		{"system", domain.ActorTypeSystem, "system"},
		{"group", domain.ActorTypeGroup, "group"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.at.String())
		})
	}
}

func TestActorType_TextRoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		at   domain.ActorType
	}{
		{"person", domain.ActorTypePerson},
		{"system", domain.ActorTypeSystem},
		{"group", domain.ActorTypeGroup},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := tt.at.MarshalText()
			require.NoError(t, err)

			var got domain.ActorType
			err = got.UnmarshalText(data)
			require.NoError(t, err)
			assert.Equal(t, tt.at, got)
		})
	}
}

func TestActorType_MarshalText_Invalid(t *testing.T) {
	t.Parallel()
	invalid := domain.ActorType("bad")
	_, err := invalid.MarshalText()
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestActorType_UnmarshalText_Invalid(t *testing.T) {
	t.Parallel()
	var at domain.ActorType
	err := at.UnmarshalText([]byte("nonsense"))
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestActorType_UnmarshalText_Empty(t *testing.T) {
	t.Parallel()
	var at domain.ActorType
	err := at.UnmarshalText([]byte(""))
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestAllActorTypes(t *testing.T) {
	t.Parallel()
	assert.Len(t, domain.AllActorTypes(), 3)
}

// -- StoryActor tests --

func TestNewStoryActor_Valid(t *testing.T) {
	t.Parallel()
	actor, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	assert.Equal(t, "Customer", actor.Name())
	assert.Equal(t, domain.ActorTypePerson, actor.Type())
	assert.Equal(t, vo.UserStated, actor.Trust())
	assert.Empty(t, actor.Source())
}

func TestNewStoryActor_EmptyName(t *testing.T) {
	t.Parallel()
	_, err := domain.NewStoryActor("", domain.ActorTypePerson, vo.UserStated, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewStoryActor_WhitespaceName(t *testing.T) {
	t.Parallel()
	_, err := domain.NewStoryActor("   ", domain.ActorTypePerson, vo.UserStated, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewStoryActor_InvalidType(t *testing.T) {
	t.Parallel()
	_, err := domain.NewStoryActor("Customer", domain.ActorType("invalid"), vo.UserStated, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewStoryActor_InvalidTrust(t *testing.T) {
	t.Parallel()
	_, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.TrustLevel(99), "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewStoryActor_AIResearchedWithoutSource(t *testing.T) {
	t.Parallel()
	_, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.AIResearched, "")
	require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
}

func TestNewStoryActor_AIResearchedWithSource(t *testing.T) {
	t.Parallel()
	actor, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.AIResearched, "ref")
	require.NoError(t, err)
	assert.Equal(t, "ref", actor.Source())
}

func TestNewStoryActor_UserStatedWithoutSource(t *testing.T) {
	t.Parallel()
	actor, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	assert.Empty(t, actor.Source())
}

func TestStoryActor_Equals_CaseInsensitive(t *testing.T) {
	t.Parallel()
	a, err := domain.NewStoryActor("customer", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	b, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	assert.True(t, a.Equals(b))
}

func TestStoryActor_Equals_DifferentName(t *testing.T) {
	t.Parallel()
	a, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	b, err := domain.NewStoryActor("Seller", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	assert.False(t, a.Equals(b))
}

func TestStoryActor_String(t *testing.T) {
	t.Parallel()
	actor, err := domain.NewStoryActor("Customer", domain.ActorTypePerson, vo.UserStated, "")
	require.NoError(t, err)
	assert.Equal(t, "Actor: Customer (person, user_stated)", actor.String())
}
