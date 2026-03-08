// Package composition is the composition root that wires all dependencies.
// This is the ONLY place where concrete infrastructure adapters are instantiated
// and injected into application handlers. Handlers only know about port interfaces.
package composition

import (
	"fmt"

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
	"github.com/alty-cli/alty/internal/shared/infrastructure/eventbus"
	"github.com/alty-cli/alty/internal/shared/infrastructure/persistence"
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

	// --- Rescue ---
	RescueHandler *rescueapp.RescueHandler

	// --- Challenge ---
	ChallengeHandler *challengeapp.ChallengeHandler

	// --- Infrastructure ---
	EventBus *eventbus.Bus

	// --- Metadata ---
	Version string
}

// NewApp creates a fully wired App with all dependencies injected.
// Infrastructure adapters are created here and injected into handlers.
func NewApp() (*App, error) {
	// 1. Event bus
	bus := eventbus.NewBus()

	// 2. Shared infrastructure
	fileWriter := persistence.NewFilesystemFileWriter()

	// 3. Discovery infrastructure
	toolScanner := discoveryinfra.NewFilesystemToolScanner("")
	artifactRenderer := discoveryinfra.NewMarkdownArtifactRenderer()

	// 4. DocHealth infrastructure
	docScanner := dochealthinfra.NewFilesystemDocScanner()

	// 5. Fitness infrastructure
	gateRunner := fitnessinfra.NewSubprocessGateRunner("", nil)

	// 6. Knowledge infrastructure
	knowledgeReader := knowledgeinfra.NewFileKnowledgeReader(".alty/knowledge")

	// 7. Rescue infrastructure
	projectScanner := &rescueinfra.ProjectScanner{}
	gitOps := &rescueinfra.GitOpsAdapter{}

	// 8. Ticket infrastructure
	ticketReader := ticketinfra.NewBeadsTicketReader(".beads")

	// 9. Challenge infrastructure
	challenger := &challengeinfra.RuleBasedChallengerAdapter{}

	// 10. DocReview infrastructure (reuses the same scanner as DocHealth)
	docReviewAdapter := dochealthinfra.NewDocReviewAdapter(docScanner)

	// 11. Research infrastructure
	spikeFollowUpAdapter := researchinfra.NewSpikeFollowUpAdapter()

	// --- Wire handlers (using adapter bridges for interface mismatches) ---

	toolDetector := &bootstrapToolDetectorAdapter{scanner: toolScanner}

	fileChecker := &bootstrapinfra.OSFileChecker{}
	bootstrapHandler := bootstrapapp.NewBootstrapHandler(toolDetector, fileChecker)
	detectionHandler := discoveryapp.NewDetectionHandler(toolDetector)
	discoveryHandler := discoveryapp.NewDiscoveryHandler()
	artifactGenerationHandler := discoveryapp.NewArtifactGenerationHandler(artifactRenderer, fileWriter)
	fitnessGenerationHandler := fitnessapp.NewFitnessGenerationHandler(fileWriter)
	qualityGateHandler := fitnessapp.NewQualityGateHandler(gateRunner)
	ticketGenerationHandler := ticketapp.NewTicketGenerationHandler(fileWriter)
	ticketHealthHandler := ticketapp.NewTicketHealthHandler(&ticketReaderAdapter{reader: ticketReader})
	configGenerationHandler := ttapp.NewConfigGenerationHandler(fileWriter)
	personaHandler := ttapp.NewPersonaHandler(fileWriter)
	docHealthHandler := dochealthapp.NewDocHealthHandler(&docScannerAdapter{scanner: docScanner})
	docReviewHandler := dochealthapp.NewDocReviewHandler(docReviewAdapter)
	knowledgeLookupHandler := knowledgeapp.NewKnowledgeLookupHandler(knowledgeReader)
	spikeFollowUpHandler := researchapp.NewSpikeFollowUpHandler(spikeFollowUpAdapter)
	rescueHandler := rescueapp.NewRescueHandler(projectScanner, gitOps, fileWriter)
	challengeHandler := challengeapp.NewChallengeHandler(challenger)

	return &App{
		BootstrapHandler:          bootstrapHandler,
		DetectionHandler:          detectionHandler,
		DiscoveryHandler:          discoveryHandler,
		ArtifactGenerationHandler: artifactGenerationHandler,
		FitnessGenerationHandler:  fitnessGenerationHandler,
		QualityGateHandler:        qualityGateHandler,
		TicketGenerationHandler:   ticketGenerationHandler,
		TicketHealthHandler:       ticketHealthHandler,
		ConfigGenerationHandler:   configGenerationHandler,
		PersonaHandler:            personaHandler,
		DocHealthHandler:          docHealthHandler,
		DocReviewHandler:          docReviewHandler,
		SpikeFollowUpHandler:      spikeFollowUpHandler,
		KnowledgeLookupHandler:    knowledgeLookupHandler,
		RescueHandler:             rescueHandler,
		ChallengeHandler:          challengeHandler,
		EventBus:                  bus,
		Version:                   Version,
	}, nil
}

// Close shuts down the event bus and releases resources.
func (a *App) Close() error {
	if err := a.EventBus.Close(); err != nil {
		return fmt.Errorf("closing event bus: %w", err)
	}
	return nil
}
