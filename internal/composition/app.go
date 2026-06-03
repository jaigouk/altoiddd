// Package composition is the composition root that wires all dependencies.
// This is the ONLY place where concrete infrastructure adapters are instantiated
// and injected into application handlers. Handlers only know about port interfaces.
package composition

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	bootstrapapp "github.com/alto-cli/alto/internal/bootstrap/application"
	bootstrapinfra "github.com/alto-cli/alto/internal/bootstrap/infrastructure"
	challengeapp "github.com/alto-cli/alto/internal/challenge/application"
	challengeinfra "github.com/alto-cli/alto/internal/challenge/infrastructure"
	discoveryapp "github.com/alto-cli/alto/internal/discovery/application"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	discoveryinfra "github.com/alto-cli/alto/internal/discovery/infrastructure"
	dochealthapp "github.com/alto-cli/alto/internal/dochealth/application"
	dochealthdomain "github.com/alto-cli/alto/internal/dochealth/domain"
	dochealthinfra "github.com/alto-cli/alto/internal/dochealth/infrastructure"
	docimportapp "github.com/alto-cli/alto/internal/docimport/application"
	docimportinfra "github.com/alto-cli/alto/internal/docimport/infrastructure"
	fitnessapp "github.com/alto-cli/alto/internal/fitness/application"
	fitnessinfra "github.com/alto-cli/alto/internal/fitness/infrastructure"
	knowledgeapp "github.com/alto-cli/alto/internal/knowledge/application"
	knowledgeinfra "github.com/alto-cli/alto/internal/knowledge/infrastructure"
	rescueapp "github.com/alto-cli/alto/internal/rescue/application"
	rescueinfra "github.com/alto-cli/alto/internal/rescue/infrastructure"
	researchapp "github.com/alto-cli/alto/internal/research/application"
	researchinfra "github.com/alto-cli/alto/internal/research/infrastructure"
	shareddomain "github.com/alto-cli/alto/internal/shared/domain"
	"github.com/alto-cli/alto/internal/shared/domain/valueobjects"
	"github.com/alto-cli/alto/internal/shared/infrastructure/eventbus"
	"github.com/alto-cli/alto/internal/shared/infrastructure/llm"
	"github.com/alto-cli/alto/internal/shared/infrastructure/persistence"
	"github.com/alto-cli/alto/internal/shared/infrastructure/stack"
	ticketapp "github.com/alto-cli/alto/internal/ticket/application"
	ticketinfra "github.com/alto-cli/alto/internal/ticket/infrastructure"
	ttapp "github.com/alto-cli/alto/internal/tooltranslation/application"
	ttdomain "github.com/alto-cli/alto/internal/tooltranslation/domain"
	ttinfra "github.com/alto-cli/alto/internal/tooltranslation/infrastructure"
)

// Version is the application version. Set via ldflags at build time.
var Version = "dev"

// App holds all wired dependencies. Downstream adapters (CLI, MCP) access
// handlers through this struct. This is the single place where the dependency
// graph is assembled.
type App struct {
	// --- Bootstrap ---
	BootstrapHandler *bootstrapapp.BootstrapHandler
	ProjectDetector  bootstrapapp.ProjectDetector
	GitCommitter     bootstrapapp.GitCommitter

	// --- DocImport ---
	DocImportHandler *docimportapp.DocImportHandler

	// --- Discovery ---
	DetectionHandler          *discoveryapp.DetectionHandler
	DiscoveryHandler          *discoveryapp.DiscoveryHandler
	ArtifactGenerationHandler *discoveryapp.ArtifactGenerationHandler
	DocInferenceHandler       *discoveryapp.DocInferenceHandler
	BoundaryDetector          discoveryapp.BoundaryDetector
	DomainResearcher          discoveryapp.DomainResearcher
	GlossaryExportHandler     *discoveryapp.GlossaryExportHandler
	ContextMapExportHandler   *discoveryapp.ContextMapExportHandler
	PlantUMLExportHandler     *discoveryapp.PlantUMLExportHandler
	EgnExportHandler          *discoveryapp.EgnExportHandler
	DiscoveryReportHandler    *discoveryapp.DiscoveryReportHandler

	// --- Fitness ---
	FitnessGenerationHandler *fitnessapp.FitnessGenerationHandler
	QualityGateHandler       *fitnessapp.QualityGateHandler

	// --- Ticket ---
	TicketGenerationHandler *ticketapp.TicketGenerationHandler
	TicketHealthHandler     *ticketapp.TicketHealthHandler
	TicketVerifyHandler     *ticketapp.TicketVerifyHandler
	RippleHandler           *ticketapp.RippleHandler

	// --- ToolTranslation ---
	ConfigGenerationHandler        *ttapp.ConfigGenerationHandler
	PersonaHandler                 *ttapp.PersonaHandler
	WorkflowAssetGenerationHandler *ttapp.WorkflowAssetGenerationHandler

	// --- DocHealth ---
	DocHealthHandler      *dochealthapp.DocHealthHandler
	DocReviewHandler      *dochealthapp.DocReviewHandler
	ScaffoldHealthHandler *dochealthapp.ScaffoldHealthHandler

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
	LLMClient           llm.Client
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

	// 3. Stack detection (shared by discovery + fitness)
	stackProfile := stack.DetectProfile("")
	// README fallback: if no manifest-based stack was detected (GenericProfile),
	// try parsing README.md for stack-language signals. This handles new-project
	// bootstrap (PRD Scenario 1: docs/PRD.md:29) where source files do not yet
	// exist but the README already describes the intended stack. Errors reading
	// the README are intentionally swallowed — we retain GenericProfile silently.
	if _, isGeneric := stackProfile.(valueobjects.GenericProfile); isGeneric {
		if readme, err := os.ReadFile("README.md"); err == nil {
			if lang := stack.ExtractLanguageFromText(string(readme)); lang != "" {
				ts := valueobjects.NewTechStack(lang, "")
				stackProfile = discoverydomain.ResolveProfile(&ts)
			}
		}
	}

	// 4. Discovery infrastructure
	toolScanner := discoveryinfra.NewFilesystemToolScanner("")
	artifactRenderer := discoveryinfra.NewMarkdownArtifactRenderer(stackProfile)
	sessionRepo := discoveryinfra.NewFileSystemSessionRepository("alto-scaffold")
	storyParser := &discoveryinfra.StoryYAMLParser{}
	glossaryParser := &discoveryinfra.GlossaryYAMLParser{}
	contextMapParser := &discoveryinfra.ContextMapYAMLParser{}

	// 5. DocHealth infrastructure
	docScanner := dochealthinfra.NewFilesystemDocScanner()

	// 6. Fitness infrastructure
	gateRunner := fitnessinfra.NewSubprocessGateRunner("", stackProfile)

	// 6. Knowledge infrastructure
	knowledgeReader := knowledgeinfra.NewFileKnowledgeReader("alto-scaffold/knowledge")
	driftDetector := knowledgeinfra.NewDriftDetectionAdapter(".")

	// 7. Rescue infrastructure
	projectScanner := &rescueinfra.ProjectScanner{}
	gitOps := &rescueinfra.GitOpsAdapter{}
	testRunner := &rescueinfra.TestRunnerAdapter{}

	// 8. Ticket infrastructure
	ticketReader := ticketinfra.NewBeadsTicketReader(".beads")
	ticketContentReader := ticketinfra.NewBeadsTicketContentReader(".beads")
	commandRunner := ticketinfra.NewShellCommandRunner()
	beadsGraphReader := ticketinfra.NewBeadsGraphReader()
	beadsLabelWriter := ticketinfra.NewBeadsLabelWriter()
	beadsCommentWriter := ticketinfra.NewBeadsCommentWriter()

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

	// --- LLM credential detection ---
	homeDir, _ := os.UserHomeDir()
	credDetector := llm.NewCredentialDetector(os.Getenv, os.ReadFile, homeDir)
	detectedCreds := credDetector.Detect()

	var llmConfig llm.Config
	if len(detectedCreds) > 0 {
		best := detectedCreds[0]
		llmConfig = llm.NewConfig(best.Provider(), best.Model(), best.APIKey(), best.BaseURL(), 30.0)
	} else {
		llmConfig = llm.DefaultConfig()
	}
	llmClient := llm.Factory{}.Create(llmConfig)

	// --- Boundary detection (hybrid: algorithmic + LLM) ---
	algorithmicDetector := discoveryinfra.NewAlgorithmicDetector()
	var llmBoundaryDetector *discoveryinfra.LLMBoundaryDetector
	if llmClient != nil {
		llmBoundaryDetector = discoveryinfra.NewLLMBoundaryDetector(llmClient)
	}
	hybridDetector := discoveryinfra.NewHybridBoundaryDetector(algorithmicDetector, llmBoundaryDetector)

	// --- Domain researcher (web search + LLM extraction) ---
	domainResearcher := discoveryinfra.NewWebSearchDomainResearcher(llmClient, &http.Client{Timeout: 10 * time.Second})

	// --- Wire handlers (using adapter bridges for interface mismatches) ---

	toolDetector := &bootstrapToolDetectorAdapter{scanner: toolScanner}
	discoveryDetector := &discoveryToolDetectorAdapter{scanner: toolScanner}

	fileChecker := &bootstrapinfra.OSFileChecker{}
	contentProvider := &bootstrapinfra.ContentProviderAdapter{}
	projectDetector := &bootstrapinfra.FileSystemProjectDetector{}
	// DocImport infrastructure
	docParser := docimportinfra.NewMarkdownDocParser()
	docImportHandler := docimportapp.NewDocImportHandler(docParser)

	gitCommitter := &bootstrapinfra.GitCommitterAdapter{}
	// bootstrapHandler is constructed AFTER workflowAssetGenerationHandler
	// below so it can be wired with WithWorkflowAssetGenerator. See
	// "Bootstrap handler relocation" comment further down.
	detectionHandler := discoveryapp.NewDetectionHandler(discoveryDetector)
	discoveryHandler := discoveryapp.NewDiscoveryHandler(publisher, discoveryapp.WithSessionRepository(sessionRepo))
	artifactGenerationHandler := discoveryapp.NewArtifactGenerationHandler(artifactRenderer, fileWriter, publisher)
	glossaryExportHandler := discoveryapp.NewGlossaryExportHandler(storyParser, glossaryParser)
	contextMapExportHandler := discoveryapp.NewContextMapExportHandler(contextMapParser)
	plantUMLExporter := &discoveryinfra.PlantUMLExporter{}
	plantUMLExportHandler := discoveryapp.NewPlantUMLExportHandler(storyParser, plantUMLExporter, fileWriter)
	egnExporter := &discoveryinfra.EgnExporter{}
	egnExportHandler := discoveryapp.NewEgnExportHandler(storyParser, egnExporter, fileWriter)
	discoveryReportHandler := discoveryapp.NewDiscoveryReportHandler(storyParser, glossaryParser, contextMapParser, fileWriter)

	// DocInference: doc reader + LLM reader + regex fallback
	fsDocReader := discoveryinfra.NewFilesystemDocReader()
	llmDocReader := discoveryinfra.NewLLMDocReaderAdapter(llmClient)
	regexFallback := &regexImporterAdapter{handler: docImportHandler}
	docInferenceHandler := discoveryapp.NewDocInferenceHandler(fsDocReader, llmDocReader, regexFallback)
	fitnessGenerationHandler := fitnessapp.NewFitnessGenerationHandler(fileWriter, publisher)
	qualityGateHandler := fitnessapp.NewQualityGateHandler(gateRunner)
	ticketGenerationHandler := ticketapp.NewTicketGenerationHandler(fileWriter, publisher)
	portScannerBridge := &portScannerBridge{scanner: fitnessinfra.CodebasePortScanner{}}
	ticketGenerationHandler.SetPortScanner(portScannerBridge, "")
	ticketHealthHandler := ticketapp.NewTicketHealthHandler(&ticketReaderAdapter{reader: ticketReader})
	ticketVerifyHandler := ticketapp.NewTicketVerifyHandler(ticketContentReader, commandRunner)
	rippleHandler := ticketapp.NewRippleHandler(beadsGraphReader, beadsLabelWriter, beadsCommentWriter)
	configGenerationHandler := ttapp.NewConfigGenerationHandler(fileWriter, publisher)
	personaHandler := ttapp.NewPersonaHandler(fileWriter)
	openCodeCommandAdapter := ttinfra.NewOpenCodeCommandAdapter()
	workflowAssetRegistry := map[ttdomain.SupportedTool]ttapp.WorkflowAssetGeneration{
		ttdomain.ToolOpenCode: openCodeCommandAdapter,
	}
	workflowAssetGenerationHandler := ttapp.NewWorkflowAssetGenerationHandler(workflowAssetRegistry)
	// Bootstrap handler relocation: constructed here so the
	// WithWorkflowAssetGenerator option can reference the already-built
	// WAG handler. The CLI's `alto init --with-scaffold --primary-tool=opencode`
	// flow dispatches into WAG after the embed write succeeds.
	// The bootstrapWorkflowAssetAdapter bridges from the bootstrap-local
	// port to the tooltranslation handler — bootstrap must not depend on
	// tooltranslation directly (arch-go boundary).
	embedScaffoldWriter := bootstrapinfra.NewEmbedScaffoldWriter()
	beadsHookWriter := bootstrapinfra.NewFilesystemBeadsHookWriter()
	workflowAssetBridge := &bootstrapWorkflowAssetAdapter{handler: workflowAssetGenerationHandler}
	bootstrapHandler := bootstrapapp.NewBootstrapHandler(
		toolDetector, fileChecker, publisher, fileWriter, contentProvider,
		bootstrapapp.WithGitCommitter(gitCommitter),
		bootstrapapp.WithScaffoldWriter(embedScaffoldWriter),
		bootstrapapp.WithWorkflowAssetGenerator(workflowAssetBridge),
		bootstrapapp.WithBeadsHookWriter(beadsHookWriter),
	)
	docHealthHandler := dochealthapp.NewDocHealthHandler(&docScannerAdapter{scanner: docScanner})
	docReviewHandler := dochealthapp.NewDocReviewHandler(docReviewAdapter)
	scaffoldWalker := dochealthinfra.NewFilesystemScaffoldWalker()
	scaffoldParams, paramsErr := dochealthdomain.NewScaffoldParams(30, nil)
	if paramsErr != nil {
		cancelSub()
		_ = bus.Close()
		return nil, fmt.Errorf("constructing scaffold params: %w", paramsErr)
	}
	scaffoldRules := dochealthinfra.DefaultScaffoldRules(scaffoldParams)
	scaffoldHealthHandler := dochealthapp.NewScaffoldHealthHandler(scaffoldWalker, scaffoldRules)
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
		BootstrapHandler:               bootstrapHandler,
		ProjectDetector:                projectDetector,
		GitCommitter:                   gitCommitter,
		DocImportHandler:               docImportHandler,
		DetectionHandler:               detectionHandler,
		DiscoveryHandler:               discoveryHandler,
		ArtifactGenerationHandler:      artifactGenerationHandler,
		DocInferenceHandler:            docInferenceHandler,
		BoundaryDetector:               hybridDetector,
		DomainResearcher:               domainResearcher,
		GlossaryExportHandler:          glossaryExportHandler,
		ContextMapExportHandler:        contextMapExportHandler,
		PlantUMLExportHandler:          plantUMLExportHandler,
		EgnExportHandler:               egnExportHandler,
		DiscoveryReportHandler:         discoveryReportHandler,
		FitnessGenerationHandler:       fitnessGenerationHandler,
		QualityGateHandler:             qualityGateHandler,
		TicketGenerationHandler:        ticketGenerationHandler,
		TicketHealthHandler:            ticketHealthHandler,
		TicketVerifyHandler:            ticketVerifyHandler,
		RippleHandler:                  rippleHandler,
		ConfigGenerationHandler:        configGenerationHandler,
		PersonaHandler:                 personaHandler,
		WorkflowAssetGenerationHandler: workflowAssetGenerationHandler,
		DocHealthHandler:               docHealthHandler,
		DocReviewHandler:               docReviewHandler,
		ScaffoldHealthHandler:          scaffoldHealthHandler,
		SpikeFollowUpHandler:           spikeFollowUpHandler,
		KnowledgeLookupHandler:         knowledgeLookupHandler,
		DriftDetectionHandler:          driftDetectionHandler,
		RescueHandler:                  rescueHandler,
		GapQueryHandler:                gapQueryHandler,
		ChallengeHandler:               challengeHandler,
		VersionHandler:                 versionHandler,
		LLMClient:                      llmClient,
		EventBus:                       bus,
		Subscriber:                     subscriber,
		WorkflowCoordinator:            coordinator,
		Version:                        Version,
		cancelEvents:                   cancelSub,
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
