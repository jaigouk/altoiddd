package application_test

import (
	"context"

	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	"github.com/alto-cli/alto/internal/ticket/application"
	ticketdomain "github.com/alto-cli/alto/internal/ticket/domain"
)

// Compile-time interface satisfaction checks.
var (
	_ application.BeadsWriter      = (*mockBeadsWriter)(nil)
	_ application.TicketGeneration = (*mockTicketGeneration)(nil)
	_ application.TicketHealth     = (*mockTicketHealth)(nil)
)

// --- mockBeadsWriter ---

type mockBeadsWriter struct{}

func (m *mockBeadsWriter) WriteEpic(_ context.Context, _ ticketdomain.GeneratedEpic) (string, error) {
	return "", nil
}

func (m *mockBeadsWriter) WriteTicket(_ context.Context, _ ticketdomain.GeneratedTicket) (string, error) {
	return "", nil
}

func (m *mockBeadsWriter) SetDependency(_ context.Context, _ string, _ string) error {
	return nil
}

// --- mockTicketGeneration ---

type mockTicketGeneration struct{}

func (m *mockTicketGeneration) Generate(_ context.Context, _ *ddd.DomainModel, _ string) error {
	return nil
}

// --- mockTicketHealth ---

type mockTicketHealth struct{}

func (m *mockTicketHealth) Report(_ context.Context, _ string) (ticketdomain.TicketHealthReport, error) {
	return ticketdomain.TicketHealthReport{}, nil
}
