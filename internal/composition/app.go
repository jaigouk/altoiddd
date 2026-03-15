// Package composition is the composition root that wires all dependencies.
// This is the ONLY place where concrete infrastructure adapters are instantiated
// and injected into application handlers. Handlers only know about port interfaces.
package composition

import (
	"context"
	"fmt"
	"log/slog"

	bootstrapapp "github.com/alty-cli/alty/internal/bootstrap/application"
	bootstrapinfra "github.com/alty-cli/alty/internal/bootstrap/infrastructure"
	challengeapp "github.com/alty-cli/alty/internal/challenge/application"
	challengeinfra "github.com/alty-cli/alty/internal/challenge/infrastructure"
	discoveryapp "github.com/alty-cli/alty/internal/discovery/application"
	discoveryinfra "github.com/alty-cli/alty/internal/discovery/infrastructure"
	dochealthapp "github.com/alty-cli/alty/internal/dochealth/application"
	dochealthinfra "github.com/alty-cli/alty/internal/dochealth/infrastructure"
	fitnessapp "github.com/alty-cli/alty/internal/fitness/application"
	fitnessinfra "github.com/alty-cli/alty/internal/fitness/infrastructure"
	knowledgeapp "github.com/alty-cli/alty/internal/knowledge/application"
	knowledgeinfra "github.com/alty-cli/alty/internal/knowledge/infrastructure"
	rescueapp "github.com/alty-cli/alty/internal/rescue/application"
	rescueinfra "github.com/alty-cli/alty/internal/rescue/infrastructure"
	researchapp "github.com/alty-cli/alty/internal/research/application"
	researchinfra "github.com/alty-cli/alty/internal/research/infrastructure"
	shareddomain "github.com/alty-cli/alty/internal/shared/domain"
	"github.com/alty-cli/alty/internal/shared/domain/valueobjects"
	"github.com/alty-cli/alty/internal/shared/infrastructure/eventbus"
	"github.com/alty-cli/alty/internal/shared/infrastructure/persistence"
	"github.com/alty-cli/alty/internal/shared/infrastructure/stack"
	ticketapp "github.com/alty-cli/alty/internal/ticket/application"
	ticketinfra "github.com/alty-cli/alty/internal/ticket/infrastructure"
	ttapp "github.com/alty-cli/alty/internal/tooltranslation/application"
)

// Version is the application version. Set via ldflags at build time.
var Version = "dev"

// App holds all wired dependencies. Downstream adapters (CLI, MCP) access
// handlers through this struct. This is the single place where the dependency
// graph is assembled.
type App struct {
	// --- Bootstrap ---
	BootstrapHandler *bootstrapapp.BootstrapHandler

	// --- Discovery ---
	DetectionHandler          *discoveryapp.DetectionHandler
	DiscoveryHandler          *discoveryapp.DiscoveryHandler
	ArtifactGenerationHandler *discoveryapp.ArtifactGenerationHandler

	// --- Fitness ---
	FitnessGenerationHandler *fitnessapp.FitnessGenerationHandler
	QualityGateHandler       *fitnessapp.QualityGateHandler

	// --- Ticket ---
	TicketGenerationHandler *ticketapp.TicketGenerationHandler
	TicketHealthHandler     *ticketapp.TicketHealthHandler
	TicketVerifyHandler     *ticketapp.TicketVerifyHandler

	// --- ToolTranslation ---
	ConfigGenerationHandler *ttapp.ConfigGenerationHandler
	PersonaHandler          *ttapp.PersonaHandler

	// --- DocHealth ---
	DocHealthHandler *dochealthapp.DocHealthHandler
	DocReviewHandler *dochealthapp.DocReviewHandler

	// --- Research ---
	SpikeFollowUpHandler *researchapp.SpikeFollowUpHandler

	// --- Knowledge ---
	KnowledgeLookupHandler *knowledgeapp.KnowledgeLookupHandler
	DriftDetectionHandler  *knowledgeapp.DriftDetectionHandler

	// --- Rescue ---
	RescueHandler   *rescueapp.RescueHandler
	GapQueryHandler *rescueapp.GapQueryHandler

	// --- Challenge ---
	ChallengeHandler *challengeapp.ChallengeHandler
	VersionHandler   *challengeapp.VersionHandler

	// --- Infrastructure ---
	EventBus            *eventbus.Bus
	Subscriber          *eventbus.Subscriber
	WorkflowCoordinator *shareddomain.WorkflowCoordinator

	// --- Metadata ---
	Version string

	// cancelEvents cancels the subscriber context, signaling listener goroutines to exit.
	cancelEvents context.CancelFunc
}

// NewApp creates a fully wired App with all dependencies injected.
// Infrastructure adapters are created here and injected into handlers.
func NewApp() (*App, error) {
	// 1. Event bus
	bus := eventbus.NewBus()

	// 2. Shared infrastructure
	fileReader := persistence.NewFilesystemFileReader()
	innerWriter := persistence.NewFilesystemFileWriter()
	fileWriter := persistence.NewConflictDetectingFileWriter(innerWriter, valueobjects.ConflictStrategyRename)

	// 3. Discovery infrastructure
	toolScanner := discoveryinfra.NewFilesystemToolScanner("")
	artifactRenderer := discoveryinfra.NewMarkdownArtifactRenderer()
	sessionRepo := discoveryinfra.NewFileSystemSessionRepository(".alty")

	// 4. DocHealth infrastructure
	docScanner := dochealthinfra.NewFilesystemDocScanner()

	// 5. Fitness infrastructure
	stackProfile := stack.DetectProfile("")
	gateRunner := fitnessinfra.NewSubprocessGateRunner("", stackProfile)

	// 6. Knowledge infrastructure
	knowledgeReader := knowledgeinfra.NewFileKnowledgeReader(".alty/knowledge")
	driftDetector := knowledgeinfra.NewDriftDetectionAdapter(".")

	// 7. Rescue infrastructure
	projectScanner := &rescueinfra.ProjectScanner{}
	gitOps := &rescueinfra.GitOpsAdapter{}
	testRunner := &rescueinfra.TestRunnerAdapter{}

	// 8. Ticket infrastructure
	ticketReader := ticketinfra.NewBeadsTicketReader(".beads")
	ticketContentReader := ticketinfra.NewBeadsTicketContentReader(".beads")
	commandRunner := ticketinfra.NewShellCommandRunner()

	// 9. Challenge infrastructure
	challenger := &challengeinfra.RuleBasedChallengerAdapter{}

	// 10. DocReview infrastructure (reuses the same scanner as DocHealth)
	docReviewAdapter := dochealthinfra.NewDocReviewAdapter(docScanner)

	// 11. Research infrastructure
	spikeFollowUpAdapter := researchinfra.NewSpikeFollowUpAdapter()

	// 12. Workflow coordination (Tier 2 readiness)
	coordinator := shareddomain.NewWorkflowCoordinator()

	// --- Event publisher + subscriber ---
	publisher := eventbus.NewPublisher(bus)

	subscriber, err := wireEventSubscribers(bus, slog.Default(), coordinator)
	if err != nil {
		_ = bus.Close()
		return nil, fmt.Errorf("wiring event subscribers: %w", err)
	}

	subCtx, cancelSub := context.WithCancel(context.Background())
	if err := subscriber.Start(subCtx); err != nil {
		cancelSub()
		_ = bus.Close()
		return nil, fmt.Errorf("starting event subscriber: %w", err)
	}

	// --- Wire handlers (using adapter bridges for interface mismatches) ---

	toolDetector := &bootstrapToolDetectorAdapter{scanner: toolScanner}
	discoveryDetector := &discoveryToolDetectorAdapter{scanner: toolScanner}

	fileChecker := &bootstrapinfra.OSFileChecker{}
	contentProvider := &bootstrapinfra.ContentProviderAdapter{}
	bootstrapHandler := bootstrapapp.NewBootstrapHandler(toolDetector, fileChecker, publisher, fileWriter, contentProvider)
	detectionHandler := discoveryapp.NewDetectionHandler(discoveryDetector)
	discoveryHandler := discoveryapp.NewDiscoveryHandler(publisher, discoveryapp.WithSessionRepository(sessionRepo))
	artifactGenerationHandler := discoveryapp.NewArtifactGenerationHandler(artifactRenderer, fileWriter, publisher)
	fitnessGenerationHandler := fitnessapp.NewFitnessGenerationHandler(fileWriter, publisher)
	qualityGateHandler := fitnessapp.NewQualityGateHandler(gateRunner)
	ticketGenerationHandler := ticketapp.NewTicketGenerationHandler(fileWriter, publisher)
	ticketHealthHandler := ticketapp.NewTicketHealthHandler(&ticketReaderAdapter{reader: ticketReader})
	ticketVerifyHandler := ticketapp.NewTicketVerifyHandler(ticketContentReader, commandRunner)
	configGenerationHandler := ttapp.NewConfigGenerationHandler(fileWriter, publisher)
	personaHandler := ttapp.NewPersonaHandler(fileWriter)
	docHealthHandler := dochealthapp.NewDocHealthHandler(&docScannerAdapter{scanner: docScanner})
	docReviewHandler := dochealthapp.NewDocReviewHandler(docReviewAdapter)
	knowledgeLookupHandler := knowledgeapp.NewKnowledgeLookupHandler(knowledgeReader)
	driftDetectionHandler := knowledgeapp.NewDriftDetectionHandler(driftDetector)
	spikeFollowUpHandler := researchapp.NewSpikeFollowUpHandler(spikeFollowUpAdapter)
	dirCreator := persistence.NewFilesystemDirCreator()
	rescueHandler := rescueapp.NewRescueHandler(projectScanner, gitOps, fileWriter, publisher, testRunner, dirCreator)
	gapQueryHandler := rescueapp.NewGapQueryHandler(projectScanner, &stackProfileDetectorAdapter{})

	challengeHandler := challengeapp.NewChallengeHandler(challenger)
	versionParser := challengeinfra.NewYAMLFrontmatterParser()
	versionHandler := challengeapp.NewVersionHandler(fileReader, fileWriter, versionParser)

	return &App{
		BootstrapHandler:          bootstrapHandler,
		DetectionHandler:          detectionHandler,
		DiscoveryHandler:          discoveryHandler,
		ArtifactGenerationHandler: artifactGenerationHandler,
		FitnessGenerationHandler:  fitnessGenerationHandler,
		QualityGateHandler:        qualityGateHandler,
		TicketGenerationHandler:   ticketGenerationHandler,
		TicketHealthHandler:       ticketHealthHandler,
		TicketVerifyHandler:       ticketVerifyHandler,
		ConfigGenerationHandler:   configGenerationHandler,
		PersonaHandler:            personaHandler,
		DocHealthHandler:          docHealthHandler,
		DocReviewHandler:          docReviewHandler,
		SpikeFollowUpHandler:      spikeFollowUpHandler,
		KnowledgeLookupHandler:    knowledgeLookupHandler,
		DriftDetectionHandler:     driftDetectionHandler,
		RescueHandler:             rescueHandler,
		GapQueryHandler:           gapQueryHandler,
		ChallengeHandler:          challengeHandler,
		VersionHandler:            versionHandler,
		EventBus:                  bus,
		Subscriber:                subscriber,
		WorkflowCoordinator:       coordinator,
		Version:                   Version,
		cancelEvents:              cancelSub,
	}, nil
}

// Close shuts down the event subscriber and bus in correct order:
// 1. Cancel subscriber context (signals goroutines to exit)
// 2. Wait for subscriber goroutines to finish
// 3. Close the event bus
func (a *App) Close() error {
	a.cancelEvents()
	a.Subscriber.Wait()
	if err := a.EventBus.Close(); err != nil {
		return fmt.Errorf("closing event bus: %w", err)
	}
	return nil
}
