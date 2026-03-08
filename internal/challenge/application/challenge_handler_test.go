package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/challenge/application"
	challengedomain "github.com/alty-cli/alty/internal/challenge/domain"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
)

// ---------------------------------------------------------------------------
// Fake challenger
// ---------------------------------------------------------------------------

type fakeChallenger struct {
	challenges []challengedomain.Challenge
	called     int
	lastMax    int
}

func (f *fakeChallenger) GenerateChallenges(
	_ context.Context,
	_ *ddd.DomainModel,
	maxPerType int,
) ([]challengedomain.Challenge, error) {
	f.called++
	f.lastMax = maxPerType
	return f.challenges, nil
}

func makeChallenge(q, ctx string) challengedomain.Challenge {
	c, _ := challengedomain.NewChallenge(
		challengedomain.ChallengeLanguage,
		q, ctx, "test reference", "",
	)
	return c
}

// ---------------------------------------------------------------------------
// Tests — Generate
// ---------------------------------------------------------------------------

func TestChallengeHandler_Generate(t *testing.T) {
	t.Parallel()

	t.Run("delegates to port", func(t *testing.T) {
		t.Parallel()
		expected := []challengedomain.Challenge{makeChallenge("Is this right?", "Sales")}
		fake := &fakeChallenger{challenges: expected}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")

		result, err := handler.GenerateChallenges(context.Background(), model, 5)

		require.NoError(t, err)
		assert.Equal(t, 1, fake.called)
		assert.Equal(t, 5, fake.lastMax)
		assert.Equal(t, len(expected), len(result))
	})

	t.Run("custom max per type", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: nil}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")

		_, err := handler.GenerateChallenges(context.Background(), model, 3)

		require.NoError(t, err)
		assert.Equal(t, 3, fake.lastMax)
	})

	t.Run("stores challenges internally", func(t *testing.T) {
		t.Parallel()
		challenges := []challengedomain.Challenge{
			makeChallenge("Q1", "Sales"),
			makeChallenge("Q2", "Orders"),
		}
		fake := &fakeChallenger{challenges: challenges}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")

		_, err := handler.GenerateChallenges(context.Background(), model, 5)
		require.NoError(t, err)

		iteration := handler.Complete()
		assert.Equal(t, 2, len(iteration.Challenges()))
	})
}

// ---------------------------------------------------------------------------
// Tests — Record Response
// ---------------------------------------------------------------------------

func TestChallengeHandler_RecordResponse(t *testing.T) {
	t.Parallel()

	t.Run("stores response", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: []challengedomain.Challenge{makeChallenge("Q", "Ctx")}}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")
		_, _ = handler.GenerateChallenges(context.Background(), model, 5)

		response := challengedomain.NewChallengeResponse("c1", "Good point", true, nil)
		handler.RecordResponse(response)

		iteration := handler.Complete()
		assert.Equal(t, 1, len(iteration.Responses()))
		assert.True(t, iteration.Responses()[0].Accepted())
	})

	t.Run("multiple responses", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: []challengedomain.Challenge{makeChallenge("Q", "Ctx")}}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")
		_, _ = handler.GenerateChallenges(context.Background(), model, 5)

		handler.RecordResponse(challengedomain.NewChallengeResponse("c1", "Yes", true, nil))
		handler.RecordResponse(challengedomain.NewChallengeResponse("c2", "No", false, nil))

		iteration := handler.Complete()
		assert.Equal(t, 2, len(iteration.Responses()))
	})
}

// ---------------------------------------------------------------------------
// Tests — Complete
// ---------------------------------------------------------------------------

func TestChallengeHandler_Complete(t *testing.T) {
	t.Parallel()

	t.Run("returns challenge iteration", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: []challengedomain.Challenge{makeChallenge("Q", "Ctx")}}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")
		_, _ = handler.GenerateChallenges(context.Background(), model, 5)

		iteration := handler.Complete()
		assert.NotNil(t, iteration)
	})

	t.Run("convergence delta counts accepted updates", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: []challengedomain.Challenge{makeChallenge("Q", "Ctx")}}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")
		_, _ = handler.GenerateChallenges(context.Background(), model, 5)

		handler.RecordResponse(challengedomain.NewChallengeResponse(
			"c1", "Yes", true, []string{"Add invariant", "Add term"},
		))
		handler.RecordResponse(challengedomain.NewChallengeResponse(
			"c2", "No", false, nil,
		))
		handler.RecordResponse(challengedomain.NewChallengeResponse(
			"c3", "Yes", true, []string{"Add story"},
		))

		iteration := handler.Complete()
		assert.Equal(t, 3, iteration.ConvergenceDelta())
	})

	t.Run("convergence delta zero when all rejected", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: []challengedomain.Challenge{makeChallenge("Q", "Ctx")}}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")
		_, _ = handler.GenerateChallenges(context.Background(), model, 5)

		handler.RecordResponse(challengedomain.NewChallengeResponse("c1", "No", false, nil))

		iteration := handler.Complete()
		assert.Equal(t, 0, iteration.ConvergenceDelta())
	})

	t.Run("complete with no responses", func(t *testing.T) {
		t.Parallel()
		fake := &fakeChallenger{challenges: []challengedomain.Challenge{makeChallenge("Q", "Ctx")}}
		handler := application.NewChallengeHandler(fake)
		model := ddd.NewDomainModel("test")
		_, _ = handler.GenerateChallenges(context.Background(), model, 5)

		iteration := handler.Complete()
		assert.Equal(t, 0, len(iteration.Responses()))
		assert.Equal(t, 0, iteration.ConvergenceDelta())
	})
}
