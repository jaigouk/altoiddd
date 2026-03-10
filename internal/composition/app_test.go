package composition

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// App structure
// ---------------------------------------------------------------------------

func TestNewApp_ReturnsNonNilApp(t *testing.T) {
	t.Parallel()
	app, err := NewApp()
	require.NoError(t, err)
	assert.NotNil(t, app)
}

func TestNewApp_AllHandlersAreWired(t *testing.T) {
	t.Parallel()
	app, err := NewApp()
	require.NoError(t, err)

	assert.NotNil(t, app.BootstrapHandler, "BootstrapHandler")
	assert.NotNil(t, app.DetectionHandler, "DetectionHandler")
	assert.NotNil(t, app.DiscoveryHandler, "DiscoveryHandler")
	assert.NotNil(t, app.ArtifactGenerationHandler, "ArtifactGenerationHandler")
	assert.NotNil(t, app.FitnessGenerationHandler, "FitnessGenerationHandler")
	assert.NotNil(t, app.QualityGateHandler, "QualityGateHandler")
	assert.NotNil(t, app.TicketGenerationHandler, "TicketGenerationHandler")
	assert.NotNil(t, app.TicketHealthHandler, "TicketHealthHandler")
	assert.NotNil(t, app.ConfigGenerationHandler, "ConfigGenerationHandler")
	assert.NotNil(t, app.PersonaHandler, "PersonaHandler")
	assert.NotNil(t, app.DocHealthHandler, "DocHealthHandler")
	assert.NotNil(t, app.KnowledgeLookupHandler, "KnowledgeLookupHandler")
	assert.NotNil(t, app.RescueHandler, "RescueHandler")
	assert.NotNil(t, app.ChallengeHandler, "ChallengeHandler")
}

func TestNewApp_EventBusIsWired(t *testing.T) {
	t.Parallel()
	app, err := NewApp()
	require.NoError(t, err)
	defer app.Close()

	assert.NotNil(t, app.EventBus)
	assert.NotNil(t, app.Subscriber, "Subscriber")
}

func TestNewApp_WorkflowCoordinatorIsWired(t *testing.T) {
	t.Parallel()
	app, err := NewApp()
	require.NoError(t, err)
	defer app.Close()

	assert.NotNil(t, app.WorkflowCoordinator, "WorkflowCoordinator")
}

func TestNewApp_Close_NoError(t *testing.T) {
	t.Parallel()
	app, err := NewApp()
	require.NoError(t, err)

	err = app.Close()
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Independence
// ---------------------------------------------------------------------------

func TestNewApp_MultipleCalls_ReturnIndependentApps(t *testing.T) {
	t.Parallel()
	app1, err1 := NewApp()
	require.NoError(t, err1)
	defer app1.Close()

	app2, err2 := NewApp()
	require.NoError(t, err2)
	defer app2.Close()

	assert.NotSame(t, app1, app2)
	assert.NotSame(t, app1.BootstrapHandler, app2.BootstrapHandler)
	assert.NotSame(t, app1.DiscoveryHandler, app2.DiscoveryHandler)
}

// ---------------------------------------------------------------------------
// Version
// ---------------------------------------------------------------------------

func TestApp_Version(t *testing.T) {
	t.Parallel()
	app, err := NewApp()
	require.NoError(t, err)

	assert.NotEmpty(t, app.Version)
}
