package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alty-cli/alty/internal/ticket/domain"
)

func TestTemplateType_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		tt       domain.TemplateType
		expected string
	}{
		{"epic", domain.TemplateEpic, "epic"},
		{"task", domain.TemplateTask, "task"},
		{"spike", domain.TemplateSpike, "spike"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, tc.tt.String())
		})
	}
}

func TestTemplateType_IsValid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		tt    domain.TemplateType
		valid bool
	}{
		{"epic is valid", domain.TemplateEpic, true},
		{"task is valid", domain.TemplateTask, true},
		{"spike is valid", domain.TemplateSpike, true},
		{"unknown is invalid", domain.TemplateType("unknown"), false},
		{"empty is invalid", domain.TemplateType(""), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.valid, tc.tt.IsValid())
		})
	}
}

func TestGeneratedEpic_TemplateType(t *testing.T) {
	t.Parallel()

	// All epics should have TemplateEpic
	epic := domain.NewGeneratedEpic(
		"epic-1", "Order Epic", "Implement Order context",
		"Order", "core",
	)

	assert.Equal(t, domain.TemplateEpic, epic.TemplateType())
}

func TestGeneratedTicket_TemplateType_Task(t *testing.T) {
	t.Parallel()

	// Regular tickets (from aggregates) should have TemplateTask
	ticket := domain.NewGeneratedTicket(
		"ticket-1", "Implement Order aggregate",
		"Description", "full",
		"epic-1", "Order", "Order",
		nil, 0,
	)

	assert.Equal(t, domain.TemplateTask, ticket.TemplateType())
}

func TestGeneratedTicket_TemplateType_Spike(t *testing.T) {
	t.Parallel()

	// Spike tickets should have TemplateSpike
	ticket := domain.NewGeneratedSpikeTicket(
		"spike-1", "Spike: Research payment gateway",
		"Investigate options", "epic-1", "Payment",
	)

	assert.Equal(t, domain.TemplateSpike, ticket.TemplateType())
}
