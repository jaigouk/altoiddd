// Package application defines ports for the Discovery bounded context.
package application

import (
	"context"

	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/shared/domain/ddd"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// --- Discovery Port ---

// Discovery manages the conversational flow of the 10-question DDD framework
// with persona detection, register selection, and playback confirmation loops.
// Methods omit context.Context because this runs synchronously in a CLI process.
type Discovery interface {
	// StartSession starts a new guided discovery session from README content.
	StartSession(readmeContent string) (*discoverydomain.DiscoverySession, error)

	// DetectPersona detects the user persona based on their self-identification choice.
	DetectPersona(sessionID string, choice string) (*discoverydomain.DiscoverySession, error)

	// AnswerQuestion submits an answer to a discovery question.
	AnswerQuestion(sessionID string, questionID string, answer string) (*discoverydomain.DiscoverySession, error)

	// SkipQuestion skips a question with an explicit reason.
	SkipQuestion(sessionID string, questionID string, reason string) (*discoverydomain.DiscoverySession, error)

	// ConfirmPlayback confirms or rejects the playback summary.
	ConfirmPlayback(sessionID string, confirmed bool) (*discoverydomain.DiscoverySession, error)

	// Complete completes the discovery session and produces domain artifacts.
	Complete(sessionID string) (*discoverydomain.DiscoverySession, error)
}

// Compile-time interface compliance check.
var _ Discovery = (*DiscoveryHandler)(nil)

// --- Session Repository Port ---

// SessionRepository persists and retrieves discovery sessions.
type SessionRepository interface {
	// Save persists a discovery session.
	Save(ctx context.Context, session *discoverydomain.DiscoverySession) error

	// Load retrieves a discovery session by ID.
	Load(ctx context.Context, sessionID string) (*discoverydomain.DiscoverySession, error)

	// Exists checks whether a persisted session exists.
	Exists(ctx context.Context, sessionID string) (bool, error)
}

// --- Artifact Renderer Port ---

// ArtifactRenderer renders a DomainModel into markdown documents (PRD, DDD.md, ARCHITECTURE.md).
type ArtifactRenderer interface {
	// RenderPRD renders the PRD markdown from a domain model.
	RenderPRD(ctx context.Context, model *ddd.DomainModel) (string, error)

	// RenderDDD renders the DDD.md markdown from a domain model.
	RenderDDD(ctx context.Context, model *ddd.DomainModel) (string, error)

	// RenderArchitecture renders the ARCHITECTURE.md markdown from a domain model.
	RenderArchitecture(ctx context.Context, model *ddd.DomainModel) (string, error)
}

// --- Prompter Port ---

// Prompter handles interactive CLI prompts for discovery flow.
type Prompter interface {
	// SelectPersona displays persona choices and returns the selected choice ("1"-"4").
	SelectPersona(ctx context.Context) (string, error)

	// AskQuestion displays a question and returns the user's answer.
	// Returns empty string if the user wants to skip.
	AskQuestion(ctx context.Context, question string) (string, error)

	// AskSkipReason prompts for a reason when skipping a question.
	AskSkipReason(ctx context.Context) (string, error)

	// ConfirmPlayback displays a summary and asks for confirmation.
	// Returns true if confirmed, false if user wants to review/edit.
	ConfirmPlayback(ctx context.Context, summary string) (bool, error)
}

// --- Doc Reader Port ---

// DocReader reads documentation files from a directory.
type DocReader interface {
	ReadDocs(ctx context.Context, docsDir string) (map[string]string, error)
}

// --- LLM Doc Reader Port ---

// LLMDocReader infers a domain model from document contents using an LLM.
type LLMDocReader interface {
	// InferModel takes document contents (filename->content map) and returns
	// an InferenceResult with a structured DomainModel.
	InferModel(ctx context.Context, docs map[string]string) (*discoverydomain.InferenceResult, error)
}

// --- Regex Importer Port ---

// RegexImporter imports a domain model from a docs directory using regex parsing.
// Used as fallback when LLM is unavailable.
type RegexImporter interface {
	Import(ctx context.Context, docDir string) (*ddd.DomainModel, error)
}

// --- Story Persistence Ports ---

// StoryReader reads a DomainStory from a file path.
type StoryReader interface {
	Read(ctx context.Context, path string) (*discoverydomain.DomainStory, error)
}

// StoryWriter writes a DomainStory to a file path.
type StoryWriter interface {
	Write(ctx context.Context, path string, story *discoverydomain.DomainStory) error
}

// --- Glossary Persistence Ports ---

// GlossaryReader reads ubiquitous language entries from a file path.
type GlossaryReader interface {
	Read(ctx context.Context, path string) ([]vo.UbiquitousLanguageEntry, error)
}

// GlossaryWriter writes ubiquitous language entries to a file path.
type GlossaryWriter interface {
	Write(ctx context.Context, path string, entries []vo.UbiquitousLanguageEntry) error
}

// --- Context Map Persistence Ports ---

// ContextMapReader reads a ContextMap from a file path.
type ContextMapReader interface {
	Read(ctx context.Context, path string) (*discoverydomain.ContextMap, error)
}

// ContextMapWriter writes a ContextMap to a file path.
type ContextMapWriter interface {
	Write(ctx context.Context, path string, cm *discoverydomain.ContextMap) error
}

// --- Storytelling Prompter Port ---

// StorytellingPrompter handles interactive CLI prompts for the Domain Storytelling
// discovery flow. Coexists with the legacy Prompter interface.
type StorytellingPrompter interface {
	// SelectMode prompts the user to choose between RAPID and THOROUGH discovery modes.
	SelectMode(ctx context.Context) (discoverydomain.DiscoveryMode, error)

	// ProposeStory presents a consultant-proposed story for user review and refinement.
	ProposeStory(ctx context.Context, proposed *discoverydomain.DomainStory) (*discoverydomain.DomainStory, error)

	// AskNarration asks a moderator question and returns the user's narration response.
	AskNarration(ctx context.Context, question string, context string) (string, error)

	// ConfirmSentence presents a structured sentence for confirmation.
	// Returns the (possibly edited) sentence, whether it was accepted, and any error.
	ConfirmSentence(ctx context.Context, sentence discoverydomain.StorySentence) (discoverydomain.StorySentence, bool, error)

	// AskChoice presents lettered options with an optional recommendation.
	// Returns the selected option key.
	AskChoice(ctx context.Context, prompt string, options []Choice, recommended string) (string, error)

	// DisplayStory renders a complete domain story for read-only display.
	DisplayStory(ctx context.Context, story *discoverydomain.DomainStory) error

	// SynthesisCheckpoint presents a synthesis summary for user confirmation.
	SynthesisCheckpoint(ctx context.Context, synthesis SynthesisSummary) (bool, error)

	// AskAnnotation prompts for a business rule annotation.
	// Returns (annotation text, sentence number, error). Empty text = done. Sentence 0 = story-wide.
	AskAnnotation(ctx context.Context) (string, int, error)
}

// Choice represents a selectable option in an AskChoice prompt.
type Choice struct {
	Key         string
	Label       string
	Description string
}

// SynthesisSummary contains the data shown during a synthesis checkpoint.
type SynthesisSummary struct {
	StoriesSoFar    []*discoverydomain.DomainStory
	ActorInventory  []discoverydomain.StoryActor
	ObjectInventory []discoverydomain.WorkObject
	BoundarySignals []discoverydomain.BoundarySignal
	GlossaryTerms   []string
}

// --- Tool Detection Port ---

// ToolDetection detects installed AI coding tools and scans for configuration conflicts.
type ToolDetection interface {
	// Detect detects installed AI coding tools in the project directory.
	Detect(ctx context.Context, projectDir string) ([]string, error)

	// ScanConflicts scans for global settings conflicts between detected tools.
	ScanConflicts(ctx context.Context, projectDir string) ([]discoverydomain.SettingsConflict, error)
}

// --- Boundary Detection Port ---

// BoundaryDetector detects bounded context boundaries from domain stories.
// Mode is accepted for future use by HybridBoundaryDetector (P3-3).
type BoundaryDetector interface {
	DetectBoundaries(ctx context.Context, stories []*discoverydomain.DomainStory, mode discoverydomain.DiscoveryMode) ([]discoverydomain.BoundedContextSketch, error)
}

// --- Boundary Prompter Port ---

// BoundaryPrompter handles CLI interaction for boundary detection results.
// Separate from StorytellingPrompter per ISP (not all callers need boundary UI).
type BoundaryPrompter interface {
	// DisplayBoundaryProposals presents sketches for per-sketch accept/reject.
	// Returns names of accepted sketches.
	DisplayBoundaryProposals(ctx context.Context, proposals []discoverydomain.BoundedContextSketch) ([]string, error)

	// AskMissingContext asks if the user sees a missing area.
	// Returns non-empty string if yes, empty if no.
	AskMissingContext(ctx context.Context) (string, error)
}
