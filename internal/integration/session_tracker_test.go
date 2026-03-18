package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/composition"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	shareddomain "github.com/alto-cli/alto/internal/shared/domain"
	"github.com/alto-cli/alto/internal/shared/domain/events"
	"github.com/alto-cli/alto/internal/shared/infrastructure/eventbus"
	ticketdomain "github.com/alto-cli/alto/internal/ticket/domain"
)

// ---------------------------------------------------------------------------
// Tier 2 Subscriber Integration Tests
// ---------------------------------------------------------------------------

func TestIntegration_DiscoveryCompleted_MarksArtifactReady(t *testing.T) {
	t.Parallel()

	app, err := composition.NewApp()
	require.NoError(t, err)
	defer app.Close()

	sessionID := "test-session-discovery"

	// Publish DiscoveryCompletedEvent
	evt := discoverydomain.NewDiscoveryCompletedEvent(
		sessionID,
		discoverydomain.PersonaDeveloper,
		discoverydomain.RegisterTechnical,
		nil, nil, nil,
	)

	publisher := eventbus.NewPublisher(app.EventBus)
	err = publisher.Publish(context.Background(), evt)
	require.NoError(t, err)

	// Give subscriber time to process (blocking publish, but goroutine dispatch)
	time.Sleep(50 * time.Millisecond)

	// Verify artifact_generation is ready
	actions := app.WorkflowCoordinator.ReadyActions(sessionID)
	require.NotEmpty(t, actions, "should have ready actions after DiscoveryCompleted")

	names := extractActionNames(actions)
	assert.Contains(t, names, "artifact_generation",
		"artifact_generation should be ready after DiscoveryCompleted")
}

func TestIntegration_DomainModelGenerated_MarksFitnessTicketsConfigsReady(t *testing.T) {
	t.Parallel()

	app, err := composition.NewApp()
	require.NoError(t, err)
	defer app.Close()

	modelID := "test-model-generated"

	// Publish DomainModelGenerated event
	evt := events.NewDomainModelGenerated(modelID, nil, nil, nil, nil, nil)

	publisher := eventbus.NewPublisher(app.EventBus)
	err = publisher.Publish(context.Background(), evt)
	require.NoError(t, err)

	// Give subscriber time to process
	time.Sleep(50 * time.Millisecond)

	// Verify fitness, tickets, configs are ready
	actions := app.WorkflowCoordinator.ReadyActions(modelID)
	require.NotEmpty(t, actions, "should have ready actions after DomainModelGenerated")

	names := extractActionNames(actions)
	assert.Contains(t, names, "fitness", "fitness should be ready")
	assert.Contains(t, names, "tickets", "tickets should be ready")
	assert.Contains(t, names, "configs", "configs should be ready")
}

func TestIntegration_TicketPlanApproved_MarksRippleReviewReady(t *testing.T) {
	t.Parallel()

	app, err := composition.NewApp()
	require.NoError(t, err)
	defer app.Close()

	planID := "test-plan-approved"

	// Publish TicketPlanApproved event
	evt := ticketdomain.NewTicketPlanApproved(planID, []string{"ticket-1"}, nil)

	publisher := eventbus.NewPublisher(app.EventBus)
	err = publisher.Publish(context.Background(), evt)
	require.NoError(t, err)

	// Give subscriber time to process
	time.Sleep(50 * time.Millisecond)

	// Verify ripple_review is ready
	actions := app.WorkflowCoordinator.ReadyActions(planID)
	require.NotEmpty(t, actions, "should have ready actions after TicketPlanApproved")

	names := extractActionNames(actions)
	assert.Contains(t, names, "ripple_review", "ripple_review should be ready")
}

func TestIntegration_SessionTracker_MultipleSessions_Independent(t *testing.T) {
	t.Parallel()

	app, err := composition.NewApp()
	require.NoError(t, err)
	defer app.Close()

	publisher := eventbus.NewPublisher(app.EventBus)

	// Session 1: DiscoveryCompleted
	err = publisher.Publish(context.Background(),
		discoverydomain.NewDiscoveryCompletedEvent(
			"session-1",
			discoverydomain.PersonaDeveloper,
			discoverydomain.RegisterTechnical,
			nil, nil, nil,
		))
	require.NoError(t, err)

	// Session 2: DomainModelGenerated
	err = publisher.Publish(context.Background(),
		events.NewDomainModelGenerated("session-2", nil, nil, nil, nil, nil))
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	// Verify sessions are independent
	actions1 := app.WorkflowCoordinator.ReadyActions("session-1")
	actions2 := app.WorkflowCoordinator.ReadyActions("session-2")

	names1 := extractActionNames(actions1)
	names2 := extractActionNames(actions2)

	assert.Contains(t, names1, "artifact_generation")
	assert.NotContains(t, names1, "fitness")

	assert.Contains(t, names2, "fitness")
	assert.NotContains(t, names2, "artifact_generation")
}

// extractActionNames returns a slice of action names from ReadyActions.
func extractActionNames(actions []shareddomain.ReadyAction) []string {
	names := make([]string, 0, len(actions))
	for _, a := range actions {
		names = append(names, a.Name())
	}
	return names
}
