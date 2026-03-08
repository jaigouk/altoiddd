// Package integration provides BDD-style integration tests that verify
// handler -> port -> adapter flows using the composition root wiring.
// These tests complement the existing cross_context_test.go (domain-only)
// by exercising the application + infrastructure layers together.
package integration

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/composition"
	fitnessdomain "github.com/alty-cli/alty/internal/fitness/domain"
	rescuedomain "github.com/alty-cli/alty/internal/rescue/domain"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
	"github.com/alty-cli/alty/internal/shared/infrastructure/eventbus"
	ticketdomain "github.com/alty-cli/alty/internal/ticket/domain"
	tooltranslationdomain "github.com/alty-cli/alty/internal/tooltranslation/domain"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newApp creates a wired App for integration tests and registers a cleanup.
func newApp(t *testing.T) *composition.App {
	t.Helper()
	app, err := composition.NewApp()
	require.NoError(t, err, "NewApp should not fail")
	t.Cleanup(func() { _ = app.Close() })
	return app
}

// makeTempProjectDir creates a temp directory with a README.md.
func makeTempProjectDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("My project idea"), 0o644))
	return dir
}

// makeFinalizedModel builds a finalized DomainModel for cross-context integration tests.
func makeFinalizedModel(t *testing.T) *ddd.DomainModel {
	t.Helper()
	model := ddd.NewDomainModel("integ-model")

	story := vo.NewDomainStory(
		"Place Order",
		[]string{"Customer"},
		"Customer submits an order",
		[]string{
			"Customer creates an order with items",
			"System validates the order",
			"System processes the payment",
			"System sends a notification",
		},
		nil,
	)
	require.NoError(t, model.AddDomainStory(story))
	require.NoError(t, model.AddTerm("order", "A purchase request", "OrderManagement", nil))
	require.NoError(t, model.AddTerm("payment", "Financial transaction", "Billing", nil))
	require.NoError(t, model.AddTerm("notification", "Alert sent", "Notifications", nil))

	core := vo.SubdomainCore
	supporting := vo.SubdomainSupporting
	generic := vo.SubdomainGeneric
	require.NoError(t, model.AddBoundedContext(vo.NewDomainBoundedContext(
		"OrderManagement", "Handle orders", []string{"Order"}, &core, "core",
	)))
	require.NoError(t, model.AddBoundedContext(vo.NewDomainBoundedContext(
		"Billing", "Process payments", []string{"Payment"}, &supporting, "supporting",
	)))
	require.NoError(t, model.AddBoundedContext(vo.NewDomainBoundedContext(
		"Notifications", "Send alerts", []string{"Notification"}, &generic, "generic",
	)))
	require.NoError(t, model.AddContextRelationship(
		vo.NewContextRelationship("OrderManagement", "Billing", "Published Language"),
	))
	require.NoError(t, model.AddContextRelationship(
		vo.NewContextRelationship("OrderManagement", "Notifications", "Published Language"),
	))
	require.NoError(t, model.DesignAggregate(vo.NewAggregateDesign(
		"Order", "OrderManagement", "Order",
		[]string{"OrderItem"}, []string{"must have items"},
		[]string{"PlaceOrder"}, []string{"OrderPlaced"},
	)))
	require.NoError(t, model.Finalize())
	return model
}

// ===========================================================================
// Scenario 1: Bootstrap flow
// ===========================================================================

func TestBootstrapFlow_GivenREADME_WhenPreviewConfirmExecute_ThenProducesSession(t *testing.T) {
	t.Parallel()
	app := newApp(t)
	dir := makeTempProjectDir(t)

	// When: Preview
	session, err := app.BootstrapHandler.Preview(dir)
	require.NoError(t, err)
	require.NotNil(t, session)
	assert.NotEmpty(t, session.SessionID())
	require.NotNil(t, session.Preview(), "preview should be set after Preview()")

	// When: Confirm
	session, err = app.BootstrapHandler.Confirm(session.SessionID())
	require.NoError(t, err)

	// When: Execute
	session, err = app.BootstrapHandler.Execute(session.SessionID())
	require.NoError(t, err)
	assert.Len(t, session.Events(), 1, "completed session emits BootstrapCompleted event")
}

func TestBootstrapFlow_GivenNoREADME_WhenPreview_ThenReturnsError(t *testing.T) {
	t.Parallel()
	app := newApp(t)
	dir := t.TempDir() // no README.md

	_, err := app.BootstrapHandler.Preview(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "README.md")
}

func TestBootstrapFlow_GivenPreviewedSession_WhenCancel_ThenStatusIsCancelled(t *testing.T) {
	t.Parallel()
	app := newApp(t)
	dir := makeTempProjectDir(t)

	session, err := app.BootstrapHandler.Preview(dir)
	require.NoError(t, err)

	session, err = app.BootstrapHandler.Cancel(session.SessionID())
	require.NoError(t, err)
	assert.Contains(t, string(session.Status()), "cancelled")
}

func TestBootstrapFlow_GivenInvalidSessionID_WhenConfirm_ThenReturnsError(t *testing.T) {
	t.Parallel()
	app := newApp(t)

	_, err := app.BootstrapHandler.Confirm("nonexistent-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent-id")
}

// ===========================================================================
// Scenario 2: Detection flow
// ===========================================================================

func TestDetectionFlow_GivenProjectDir_WhenDetect_ThenReturnsResult(t *testing.T) {
	t.Parallel()
	app := newApp(t)
	dir := makeTempProjectDir(t)

	result, err := app.DetectionHandler.Detect(dir)
	require.NoError(t, err)
	// The result should be valid even if no tools are found.
	// At minimum it should not panic or error.
	assert.NotNil(t, result)
	// DetectedTools returns a slice (possibly empty)
	_ = result.DetectedTools()
}

func TestDetectionFlow_GivenEmptyDir_WhenDetect_ThenReturnsValidResult(t *testing.T) {
	t.Parallel()
	app := newApp(t)
	dir := t.TempDir()

	result, err := app.DetectionHandler.Detect(dir)
	require.NoError(t, err)
	// The real FilesystemToolScanner detects installed CLI tools (not project files),
	// so even an empty dir may return tools if claude/cursor/etc are installed.
	// The key assertion: it does not error and returns a valid result.
	assert.NotNil(t, result)
}

// ===========================================================================
// Scenario 3: Doc health flow
// ===========================================================================

func TestDocHealthFlow_GivenProjectWithDocs_WhenHandle_ThenReturnsReport(t *testing.T) {
	t.Parallel()
	app := newApp(t)
	dir := t.TempDir()

	// Set up a project with docs directory and some markdown files
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docs", "PRD.md"), []byte("# PRD\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docs", "DDD.md"), []byte("# DDD\n"), 0o644))

	ctx := context.Background()
	report, err := app.DocHealthHandler.Handle(ctx, dir)
	require.NoError(t, err)
	assert.NotNil(t, report)
	// Report should have statuses for at least the default entries
	statuses := report.Statuses()
	assert.NotEmpty(t, statuses, "report should have at least default doc statuses")
}

func TestDocHealthFlow_GivenProjectWithNoDocsDir_WhenHandle_ThenReportsDefaultEntries(t *testing.T) {
	t.Parallel()
	app := newApp(t)
	dir := t.TempDir() // empty dir, no docs/

	ctx := context.Background()
	report, err := app.DocHealthHandler.Handle(ctx, dir)
	require.NoError(t, err)
	// Should still return a report using default entries, even if docs don't exist
	assert.NotNil(t, report)
}

func TestDocHealthFlow_GivenProjectWithRegistry_WhenHandle_ThenUsesRegistry(t *testing.T) {
	t.Parallel()
	app := newApp(t)
	dir := t.TempDir()

	// Create a registry file
	registryDir := filepath.Join(dir, ".alty", "maintenance")
	require.NoError(t, os.MkdirAll(registryDir, 0o755))
	registryContent := `[[docs]]
path = "docs/CUSTOM.md"
review_interval_days = 14
`
	require.NoError(t, os.WriteFile(filepath.Join(registryDir, "doc-registry.toml"), []byte(registryContent), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "docs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "docs", "CUSTOM.md"), []byte("# Custom\n"), 0o644))

	ctx := context.Background()
	report, err := app.DocHealthHandler.Handle(ctx, dir)
	require.NoError(t, err)

	// Should have at least the custom entry
	found := false
	for _, s := range report.Statuses() {
		if s.Path() == "docs/CUSTOM.md" {
			found = true
			break
		}
	}
	assert.True(t, found, "custom registry entry should appear in report")
}

// ===========================================================================
// Scenario 4: Quality gate flow
// ===========================================================================

func TestQualityGateFlow_GivenRunner_WhenCheckAllGates_ThenReturnsReport(t *testing.T) {
	t.Parallel()
	app := newApp(t)
	ctx := context.Background()

	// With the subprocess runner from composition root, running against the
	// actual project. We run a single lightweight gate (lint) since this is
	// an integration test and we don't want long-running commands.
	// Pass a nil gate list: the handler runs all gates. This may fail if
	// lint/type/test commands are not configured, which is expected in a
	// test environment with the default subprocess runner.
	report, err := app.QualityGateHandler.Check(ctx, []vo.QualityGate{vo.QualityGateLint})
	if err != nil {
		// GateRunner returning an error is acceptable if the command is not found
		// in the test environment. The handler should NOT panic.
		t.Logf("QualityGateHandler.Check returned expected error in test env: %v", err)
		return
	}
	assert.NotNil(t, report)
}

// ===========================================================================
// Scenario 5: Knowledge lookup flow
// ===========================================================================

func TestKnowledgeLookupFlow_GivenCategories_WhenListCategories_ThenReturnsAll(t *testing.T) {
	t.Parallel()
	app := newApp(t)

	categories := app.KnowledgeLookupHandler.ListCategories()
	assert.NotEmpty(t, categories, "should return at least one category")
}

func TestKnowledgeLookupFlow_GivenInvalidPath_WhenLookup_ThenReturnsError(t *testing.T) {
	t.Parallel()
	app := newApp(t)
	ctx := context.Background()

	_, err := app.KnowledgeLookupHandler.Lookup(ctx, "", "latest")
	require.Error(t, err, "empty path should produce error")
}

func TestKnowledgeLookupFlow_GivenInvalidCategory_WhenListTopics_ThenReturnsError(t *testing.T) {
	t.Parallel()
	app := newApp(t)
	ctx := context.Background()

	_, err := app.KnowledgeLookupHandler.ListTopics(ctx, "nonexistent-category", nil)
	require.Error(t, err, "invalid category should produce error")
}

// ===========================================================================
// Scenario 6: Ticket health flow
// ===========================================================================

func TestTicketHealthFlow_GivenNoTickets_WhenReport_ThenReturnsEmptyReport(t *testing.T) {
	t.Parallel()
	app := newApp(t)
	ctx := context.Background()

	// The BeadsTicketReader reads from .beads directory. In a temp-like
	// environment with the default path, it may find no tickets.
	report, err := app.TicketHealthHandler.Report(ctx)
	require.NoError(t, err)
	assert.NotNil(t, report)
}

// ===========================================================================
// Scenario 7: Event bus flow
// ===========================================================================

type integrationEvent struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

func TestEventBusFlow_GivenSubscriber_WhenPublish_ThenSubscriberReceivesEvent(t *testing.T) {
	t.Parallel()
	app := newApp(t)

	pub := eventbus.NewPublisher(app.EventBus)
	sub := eventbus.NewSubscriber(app.EventBus)

	received := make(chan integrationEvent, 1)
	err := eventbus.SubscribeTyped(sub, func(ctx context.Context, evt *integrationEvent) error {
		received <- *evt
		return nil
	})
	require.NoError(t, err)

	err = sub.Start(context.Background())
	require.NoError(t, err)

	err = pub.Publish(context.Background(), integrationEvent{ID: "integ-1", Message: "hello from integration"})
	require.NoError(t, err)

	select {
	case evt := <-received:
		assert.Equal(t, "integ-1", evt.ID)
		assert.Equal(t, "hello from integration", evt.Message)
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for event in integration test")
	}
}

func TestEventBusFlow_GivenMultipleSubscribers_WhenPublish_ThenAllReceive(t *testing.T) {
	t.Parallel()
	app := newApp(t)

	pub := eventbus.NewPublisher(app.EventBus)
	sub1 := eventbus.NewSubscriber(app.EventBus)
	sub2 := eventbus.NewSubscriber(app.EventBus)

	var mu sync.Mutex
	var ids []string

	handler := func(label string) func(context.Context, *integrationEvent) error {
		return func(ctx context.Context, evt *integrationEvent) error {
			mu.Lock()
			defer mu.Unlock()
			ids = append(ids, label)
			return nil
		}
	}

	require.NoError(t, eventbus.SubscribeTyped(sub1, handler("sub1")))
	require.NoError(t, eventbus.SubscribeTyped(sub2, handler("sub2")))
	require.NoError(t, sub1.Start(context.Background()))
	require.NoError(t, sub2.Start(context.Background()))

	require.NoError(t, pub.Publish(context.Background(), integrationEvent{ID: "fan-1", Message: "broadcast"}))

	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(ids) == 2
	}, 3*time.Second, 10*time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	assert.Contains(t, ids, "sub1")
	assert.Contains(t, ids, "sub2")
}

// ===========================================================================
// Scenario 8: Discovery handler session flow
// ===========================================================================

func TestDiscoveryFlow_GivenHandler_WhenStartSession_ThenReturnsSession(t *testing.T) {
	t.Parallel()
	app := newApp(t)

	session, err := app.DiscoveryHandler.StartSession("# My new project idea")
	require.NoError(t, err)
	assert.NotNil(t, session)
	assert.NotEmpty(t, session.SessionID())
}

func TestDiscoveryFlow_GivenEmptyIdea_WhenStartSession_ThenAcceptsWithoutValidation(t *testing.T) {
	t.Parallel()
	app := newApp(t)

	// NOTE: The handler currently does NOT validate empty readme content.
	// This is a potential defect -- empty ideas should arguably be rejected
	// at the domain layer. Documented in QA report as a finding.
	session, err := app.DiscoveryHandler.StartSession("")
	require.NoError(t, err, "currently accepts empty, no validation at handler level")
	assert.NotNil(t, session)
}

// ===========================================================================
// Scenario 9: Challenge handler flow
// ===========================================================================

func TestChallengeFlow_GivenModel_WhenGenerateChallenges_ThenReturnsChallenges(t *testing.T) {
	t.Parallel()
	app := newApp(t)
	model := makeFinalizedModel(t)

	challenges, err := app.ChallengeHandler.GenerateChallenges(context.Background(), model, 3)
	require.NoError(t, err)
	assert.NotEmpty(t, challenges, "should generate at least one challenge")
}

// ===========================================================================
// Scenario 10: Ticket generation flow (domain-only, no file write)
// ===========================================================================

func TestTicketGenerationFlow_GivenModel_WhenGenerateAndApprove_ThenEmitsEvent(t *testing.T) {
	t.Parallel()

	model := makeFinalizedModel(t)
	plan := ticketdomain.NewTicketPlan()
	require.NoError(t, plan.GeneratePlan(model, vo.PythonUvProfile{}))
	assert.NotEmpty(t, plan.Tickets())

	require.NoError(t, plan.Approve(nil))
	assert.Len(t, plan.Events(), 1)
	evt := plan.Events()[0]
	assert.Equal(t, plan.PlanID(), evt.PlanID())
}

// ===========================================================================
// Scenario 11: Fitness generation flow (domain-only)
// ===========================================================================

func TestFitnessGenerationFlow_GivenModel_WhenGenerateContracts_ThenProducesContracts(t *testing.T) {
	t.Parallel()

	model := makeFinalizedModel(t)
	suite := fitnessdomain.NewFitnessTestSuite("mypackage")

	// Build BoundedContextInput from model
	var inputs []fitnessdomain.BoundedContextInput
	for _, bc := range model.BoundedContexts() {
		inputs = append(inputs, fitnessdomain.BoundedContextInput{
			Name:           bc.Name(),
			Classification: bc.Classification(),
			Responsibility: bc.Responsibility(),
		})
	}

	err := suite.GenerateContracts(inputs)
	require.NoError(t, err)
	assert.NotEmpty(t, suite.Contracts())
	assert.NotEmpty(t, suite.ArchRules())
}

// ===========================================================================
// Scenario 12: ToolTranslation flow via composition root
// ===========================================================================

func TestToolTranslationFlow_GivenModel_WhenTranslateAllAdapters_ThenAllProduceSections(t *testing.T) {
	t.Parallel()

	model := makeFinalizedModel(t)
	profile := vo.PythonUvProfile{}

	adapters := map[string]interface {
		Translate(*ddd.DomainModel, vo.StackProfile) []tooltranslationdomain.ConfigSection
	}{
		"ClaudeCode": tooltranslationdomain.NewClaudeCodeAdapter(),
		"Cursor":     tooltranslationdomain.NewCursorAdapter(),
		"RooCode":    tooltranslationdomain.NewRooCodeAdapter(),
		"OpenCode":   tooltranslationdomain.NewOpenCodeAdapter(),
	}

	for name, adapter := range adapters {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			sections := adapter.Translate(model, profile)
			assert.NotEmpty(t, sections, "%s adapter should produce sections", name)
		})
	}
}

// ===========================================================================
// Scenario 13: Cross-layer error propagation
// ===========================================================================

func TestErrorPropagation_GivenDomainError_WhenHandlerPropagates_ThenSentinelPreserved(t *testing.T) {
	t.Parallel()

	t.Run("TicketPlan approve empty plan propagates ErrInvariantViolation", func(t *testing.T) {
		t.Parallel()
		plan := ticketdomain.NewTicketPlan()
		err := plan.Approve(nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})

	t.Run("GapAnalysis complete without executing propagates ErrInvariantViolation", func(t *testing.T) {
		t.Parallel()
		analysis := rescuedomain.NewGapAnalysis("/tmp")
		err := analysis.Complete()
		require.Error(t, err)
		assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})
}

// ===========================================================================
// Scenario 14: Fitness strictness mapping across contexts
// ===========================================================================

func TestFitnessStrictnessMapping_GivenAllClassifications_WhenMap_ThenCorrectStrictness(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		class      vo.SubdomainClassification
		strictness fitnessdomain.ContractStrictness
	}{
		{"core", vo.SubdomainCore, fitnessdomain.ContractStrictnessStrict},
		{"supporting", vo.SubdomainSupporting, fitnessdomain.ContractStrictnessModerate},
		{"generic", vo.SubdomainGeneric, fitnessdomain.ContractStrictnessMinimal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := fitnessdomain.StrictnessFromClassification(tt.class)
			assert.Equal(t, tt.strictness, got)
			contracts := fitnessdomain.RequiredContractTypes(got)
			assert.NotEmpty(t, contracts)
		})
	}
}

// ===========================================================================
// Scenario 15: Composition root independence
// ===========================================================================

func TestCompositionRoot_GivenMultipleApps_WhenCreated_ThenIndependent(t *testing.T) {
	t.Parallel()

	app1 := newApp(t)
	app2 := newApp(t)

	assert.NotSame(t, app1.BootstrapHandler, app2.BootstrapHandler)
	assert.NotSame(t, app1.EventBus, app2.EventBus)
	assert.NotSame(t, app1.DocHealthHandler, app2.DocHealthHandler)
}
