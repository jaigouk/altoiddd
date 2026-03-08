package application_test

import (
	"context"

	"github.com/alty-cli/alty/internal/challenge/application"
	challengedomain "github.com/alty-cli/alty/internal/challenge/domain"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
)

// Compile-time interface satisfaction checks.
var _ application.Challenger = (*mockChallenger)(nil)

type mockChallenger struct{}

func (m *mockChallenger) GenerateChallenges(_ context.Context, _ *ddd.DomainModel, _ int) ([]challengedomain.Challenge, error) {
	return nil, nil
}
