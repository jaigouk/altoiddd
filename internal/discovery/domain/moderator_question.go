package domain

import (
	"fmt"
	"strings"

	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// NarrationPhase represents a phase in the domain storytelling narration process.
type NarrationPhase string

// NarrationPhase constants.
const (
	NarrationPhaseOpening   NarrationPhase = "opening"
	NarrationPhaseNarration NarrationPhase = "narration"
	NarrationPhaseDeepening NarrationPhase = "deepening"
	NarrationPhaseClosing   NarrationPhase = "closing"
)

var validNarrationPhases = map[NarrationPhase]struct{}{
	NarrationPhaseOpening:   {},
	NarrationPhaseNarration: {},
	NarrationPhaseDeepening: {},
	NarrationPhaseClosing:   {},
}

// NewNarrationPhase creates a NarrationPhase from a string, returning an error if invalid.
func NewNarrationPhase(s string) (NarrationPhase, error) {
	np := NarrationPhase(s)
	if err := np.Validate(); err != nil {
		return "", err
	}

	return np, nil
}

// AllNarrationPhases returns all valid NarrationPhase values.
func AllNarrationPhases() []NarrationPhase {
	return []NarrationPhase{
		NarrationPhaseOpening,
		NarrationPhaseNarration,
		NarrationPhaseDeepening,
		NarrationPhaseClosing,
	}
}

// String returns the string representation of a NarrationPhase.
func (p NarrationPhase) String() string {
	return string(p)
}

// Validate checks whether the NarrationPhase holds a valid value.
func (p NarrationPhase) Validate() error {
	if _, ok := validNarrationPhases[p]; !ok {
		return fmt.Errorf("invalid narration phase %q: %w", string(p), domainerrors.ErrInvariantViolation)
	}

	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (p NarrationPhase) MarshalText() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("marshaling narration phase: %w", err)
	}

	return []byte(p), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (p *NarrationPhase) UnmarshalText(data []byte) error {
	parsed, err := NewNarrationPhase(string(data))
	if err != nil {
		return err
	}

	*p = parsed

	return nil
}

// ModeratorElicits represents the type of information a moderator question elicits.
type ModeratorElicits string

// ModeratorElicits constants.
const (
	ModeratorElicitsActor      ModeratorElicits = "actor"
	ModeratorElicitsSentence   ModeratorElicits = "sentence"
	ModeratorElicitsAnnotation ModeratorElicits = "annotation"
	ModeratorElicitsDone       ModeratorElicits = "done"
)

var validModeratorElicits = map[ModeratorElicits]struct{}{
	ModeratorElicitsActor:      {},
	ModeratorElicitsSentence:   {},
	ModeratorElicitsAnnotation: {},
	ModeratorElicitsDone:       {},
}

// NewModeratorElicits creates a ModeratorElicits from a string, returning an error if invalid.
func NewModeratorElicits(s string) (ModeratorElicits, error) {
	me := ModeratorElicits(s)
	if err := me.Validate(); err != nil {
		return "", err
	}

	return me, nil
}

// AllModeratorElicits returns all valid ModeratorElicits values.
func AllModeratorElicits() []ModeratorElicits {
	return []ModeratorElicits{
		ModeratorElicitsActor,
		ModeratorElicitsSentence,
		ModeratorElicitsAnnotation,
		ModeratorElicitsDone,
	}
}

// String returns the string representation of a ModeratorElicits.
func (e ModeratorElicits) String() string {
	return string(e)
}

// Validate checks whether the ModeratorElicits holds a valid value.
func (e ModeratorElicits) Validate() error {
	if _, ok := validModeratorElicits[e]; !ok {
		return fmt.Errorf("invalid moderator elicits %q: %w", string(e), domainerrors.ErrInvariantViolation)
	}

	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (e ModeratorElicits) MarshalText() ([]byte, error) {
	if err := e.Validate(); err != nil {
		return nil, fmt.Errorf("marshaling moderator elicits: %w", err)
	}

	return []byte(e), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (e *ModeratorElicits) UnmarshalText(data []byte) error {
	parsed, err := NewModeratorElicits(string(data))
	if err != nil {
		return err
	}

	*e = parsed

	return nil
}

// ModeratorQuestion is a value object representing a question the moderator asks
// during domain storytelling narration.
type ModeratorQuestion struct {
	id               string
	phase            NarrationPhase
	technicalText    string
	nonTechnicalText string
	elicits          ModeratorElicits
}

// NewModeratorQuestion creates a ModeratorQuestion value object.
// Validates: id non-empty after trim, phase valid, technicalText non-empty after trim,
// nonTechnicalText non-empty after trim, elicits valid.
func NewModeratorQuestion(
	id string,
	phase NarrationPhase,
	technicalText, nonTechnicalText string,
	elicits ModeratorElicits,
) (ModeratorQuestion, error) {
	if strings.TrimSpace(id) == "" {
		return ModeratorQuestion{}, fmt.Errorf("moderator question id must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	if err := phase.Validate(); err != nil {
		return ModeratorQuestion{}, fmt.Errorf("moderator question phase: %w", err)
	}

	if strings.TrimSpace(technicalText) == "" {
		return ModeratorQuestion{}, fmt.Errorf("moderator question technical text must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	if strings.TrimSpace(nonTechnicalText) == "" {
		return ModeratorQuestion{}, fmt.Errorf("moderator question non-technical text must not be empty: %w", domainerrors.ErrInvariantViolation)
	}

	if err := elicits.Validate(); err != nil {
		return ModeratorQuestion{}, fmt.Errorf("moderator question elicits: %w", err)
	}

	return ModeratorQuestion{
		id:               id,
		phase:            phase,
		technicalText:    technicalText,
		nonTechnicalText: nonTechnicalText,
		elicits:          elicits,
	}, nil
}

// ID returns the question identifier.
func (q ModeratorQuestion) ID() string { return q.id }

// Phase returns the narration phase this question belongs to.
func (q ModeratorQuestion) Phase() NarrationPhase { return q.phase }

// TechnicalText returns the DDD/engineering phrasing.
func (q ModeratorQuestion) TechnicalText() string { return q.technicalText }

// NonTechnicalText returns the plain-language phrasing.
func (q ModeratorQuestion) NonTechnicalText() string { return q.nonTechnicalText }

// Elicits returns the type of information this question elicits.
func (q ModeratorQuestion) Elicits() ModeratorElicits { return q.elicits }

// Text returns the appropriate text based on register.
func (q ModeratorQuestion) Text(register DiscoveryRegister) string {
	if register == RegisterTechnical {
		return q.technicalText
	}

	return q.nonTechnicalText
}

// String returns the question ID.
func (q ModeratorQuestion) String() string { return q.id }

// mustNewModeratorQuestion creates a ModeratorQuestion or panics.
// Acceptable for init-time programmer errors, like regexp.MustCompile.
func mustNewModeratorQuestion(
	id string,
	phase NarrationPhase,
	technicalText, nonTechnicalText string,
	elicits ModeratorElicits,
) ModeratorQuestion {
	q, err := NewModeratorQuestion(id, phase, technicalText, nonTechnicalText, elicits)
	if err != nil {
		panic(fmt.Sprintf("invalid moderator question %q: %v", id, err))
	}

	return q
}

// moderatorQuestionBank is the package-level singleton question bank.
//
//nolint:gochecknoglobals // singleton bank populated at init, read-only after
var moderatorQuestionBank = []ModeratorQuestion{
	// Opening phase — elicits: actor
	mustNewModeratorQuestion(
		"MQ-O1", NarrationPhaseOpening,
		"Who is the primary actor that initiates this workflow?",
		"Who starts this process? What's their role?",
		ModeratorElicitsActor,
	),
	mustNewModeratorQuestion(
		"MQ-O2", NarrationPhaseOpening,
		"What event or trigger causes this actor to begin?",
		"What makes this process start? What happens first?",
		ModeratorElicitsActor,
	),
	mustNewModeratorQuestion(
		"MQ-O3", NarrationPhaseOpening,
		"Are there other actors involved in this workflow?",
		"Does anyone else participate in this process?",
		ModeratorElicitsActor,
	),

	// Narration phase — elicits: sentence
	mustNewModeratorQuestion(
		"MQ-N1", NarrationPhaseNarration,
		"What does {actor} do first with which work object?",
		"What does {actor} do first?",
		ModeratorElicitsSentence,
	),
	mustNewModeratorQuestion(
		"MQ-N2", NarrationPhaseNarration,
		"What happens next in the workflow sequence?",
		"What happens next?",
		ModeratorElicitsSentence,
	),
	mustNewModeratorQuestion(
		"MQ-N3", NarrationPhaseNarration,
		"Who performs this step? (press Enter to keep '{last_actor}')",
		"Who does this? (press Enter to keep '{last_actor}')",
		ModeratorElicitsSentence,
	),
	mustNewModeratorQuestion(
		"MQ-N4", NarrationPhaseNarration,
		"What work object is produced or consumed in this step?",
		"What document, form, or thing is involved here?",
		ModeratorElicitsSentence,
	),

	// Deepening phase — elicits: annotation
	mustNewModeratorQuestion(
		"MQ-D1", NarrationPhaseDeepening,
		"Are there any invariants or business rules that constrain this step?",
		"Are there any rules that must always be true here?",
		ModeratorElicitsAnnotation,
	),
	mustNewModeratorQuestion(
		"MQ-D2", NarrationPhaseDeepening,
		"What happens when this step fails? Is there an exception path?",
		"What goes wrong sometimes? What happens then?",
		ModeratorElicitsAnnotation,
	),
	mustNewModeratorQuestion(
		"MQ-D3", NarrationPhaseDeepening,
		"Are there time constraints or SLAs on this activity?",
		"Does this need to happen within a certain time?",
		ModeratorElicitsAnnotation,
	),

	// Closing phase — elicits: done
	mustNewModeratorQuestion(
		"MQ-C1", NarrationPhaseClosing,
		"Is this story complete? Have we captured the full workflow?",
		"Does this cover everything? Are we missing anything?",
		ModeratorElicitsDone,
	),
	mustNewModeratorQuestion(
		"MQ-C2", NarrationPhaseClosing,
		"Are there alternative scenarios that should be separate stories?",
		"Are there other ways this could go that we should capture separately?",
		ModeratorElicitsDone,
	),
	mustNewModeratorQuestion(
		"MQ-C3", NarrationPhaseClosing,
		"Would you like to review the captured story before we proceed?",
		"Let me read back what we've captured. Ready?",
		ModeratorElicitsDone,
	),
}

// moderatorQuestionIndex is a lookup map by ID, built once at package init.
//
//nolint:gochecknoglobals // singleton index, read-only after init
var moderatorQuestionIndex map[string]ModeratorQuestion

//nolint:gochecknoinits // singleton bank index built once
func init() {
	moderatorQuestionIndex = make(map[string]ModeratorQuestion, len(moderatorQuestionBank))
	for _, q := range moderatorQuestionBank {
		moderatorQuestionIndex[q.id] = q
	}
}

// ModeratorQuestionBank returns a defensive copy of all moderator questions.
func ModeratorQuestionBank() []ModeratorQuestion {
	out := make([]ModeratorQuestion, len(moderatorQuestionBank))
	copy(out, moderatorQuestionBank)

	return out
}

// QuestionsByPhase returns all moderator questions for a given narration phase.
// Returns a defensive copy. Returns an empty slice for unknown phases.
func QuestionsByPhase(phase NarrationPhase) []ModeratorQuestion {
	var result []ModeratorQuestion

	for _, q := range moderatorQuestionBank {
		if q.phase == phase {
			result = append(result, q)
		}
	}

	if result == nil {
		return []ModeratorQuestion{}
	}

	return result
}

// ModeratorQuestionByID looks up a moderator question by its ID.
func ModeratorQuestionByID(id string) (ModeratorQuestion, bool) {
	q, ok := moderatorQuestionIndex[id]
	return q, ok
}
