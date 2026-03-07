package domain

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// DiscoveryStatus represents the state of a discovery session.
type DiscoveryStatus string

// DiscoveryStatus constants.
const (
	StatusCreated         DiscoveryStatus = "created"
	StatusPersonaDetected DiscoveryStatus = "persona_detected"
	StatusAnswering       DiscoveryStatus = "answering"
	StatusPlaybackPending DiscoveryStatus = "playback_pending"
	StatusCompleted       DiscoveryStatus = "completed"
	StatusCancelled       DiscoveryStatus = "cancelled"
	StatusRound1Complete  DiscoveryStatus = "round_1_complete"
	StatusChallenging     DiscoveryStatus = "challenging"
	StatusRound2Complete  DiscoveryStatus = "round_2_complete"
	StatusSimulating      DiscoveryStatus = "simulating"
)

// Ordered question phases for enforcement.
var questionPhases = []QuestionPhase{PhaseActors, PhaseStory, PhaseEvents, PhaseBoundaries}

// Persona choices mapping.
var personaChoices = map[string]struct {
	persona  DiscoveryPersona
	register DiscoveryRegister
}{
	"1": {PersonaDeveloper, RegisterTechnical},
	"2": {PersonaProductOwner, RegisterNonTechnical},
	"3": {PersonaDomainExpert, RegisterNonTechnical},
	"4": {PersonaMixed, RegisterNonTechnical},
}

const playbackInterval = 3

// DiscoverySession is the aggregate root for the 10-question DDD discovery flow.
type DiscoverySession struct {
	register                 *DiscoveryRegister
	skipped                  map[string]bool
	round                    *DiscoveryRound
	mode                     *DiscoveryMode
	techStack                *vo.TechStack
	persona                  *DiscoveryPersona
	sessionID                string
	readmeContent            string
	status                   DiscoveryStatus
	events                   []DiscoveryCompletedEvent
	playbackConfirmations    []Playback
	answers                  []Answer
	answersSinceLastPlayback int
}

// NewDiscoverySession creates a new session in CREATED state.
func NewDiscoverySession(readmeContent string) *DiscoverySession {
	return &DiscoverySession{
		sessionID:     uuid.New().String(),
		readmeContent: readmeContent,
		status:        StatusCreated,
		skipped:       make(map[string]bool),
	}
}

// -- Properties --

// SessionID returns the unique session identifier.
func (s *DiscoverySession) SessionID() string { return s.sessionID }

// ReadmeContent returns the raw README text.
func (s *DiscoverySession) ReadmeContent() string { return s.readmeContent }

// Status returns the current session state.
func (s *DiscoverySession) Status() DiscoveryStatus { return s.status }

// Persona returns the detected persona and whether it's set.
func (s *DiscoverySession) Persona() (DiscoveryPersona, bool) {
	if s.persona == nil {
		return "", false
	}
	return *s.persona, true
}

// Register returns the language register and whether it's set.
func (s *DiscoverySession) Register() (DiscoveryRegister, bool) {
	if s.register == nil {
		return "", false
	}
	return *s.register, true
}

// TechStack returns the tech stack, or nil if not set.
func (s *DiscoverySession) TechStack() *vo.TechStack { return s.techStack }

// Answers returns a defensive copy of all answers.
func (s *DiscoverySession) Answers() []Answer {
	out := make([]Answer, len(s.answers))
	copy(out, s.answers)
	return out
}

// PlaybackConfirmations returns a defensive copy of all playback confirmations.
func (s *DiscoverySession) PlaybackConfirmations() []Playback {
	out := make([]Playback, len(s.playbackConfirmations))
	copy(out, s.playbackConfirmations)
	return out
}

// Mode returns the discovery mode. Defaults to EXPRESS if not set.
func (s *DiscoverySession) Mode() DiscoveryMode {
	if s.mode == nil {
		return ModeExpress
	}
	return *s.mode
}

// Events returns a defensive copy of domain events.
func (s *DiscoverySession) Events() []DiscoveryCompletedEvent {
	out := make([]DiscoveryCompletedEvent, len(s.events))
	copy(out, s.events)
	return out
}

// CurrentPhase returns the current discovery phase based on answered/skipped questions.
func (s *DiscoverySession) CurrentPhase() QuestionPhase {
	if len(s.answers) == 0 && len(s.skipped) == 0 {
		return PhaseSeed
	}

	allHandled := make(map[string]bool)
	for _, a := range s.answers {
		allHandled[a.QuestionID()] = true
	}
	for id := range s.skipped {
		allHandled[id] = true
	}

	catalog := QuestionCatalog()

	// Check from last phase backward
	for i := len(questionPhases) - 1; i >= 0; i-- {
		phase := questionPhases[i]
		allDone := true
		for _, q := range catalog {
			if q.Phase() == phase && !allHandled[q.ID()] {
				allDone = false
				break
			}
		}
		if allDone {
			if i+1 < len(questionPhases) {
				return questionPhases[i+1]
			}
			return phase
		}
	}

	// Find first incomplete phase
	for _, phase := range questionPhases {
		for _, q := range catalog {
			if q.Phase() == phase && !allHandled[q.ID()] {
				return phase
			}
		}
	}

	return questionPhases[len(questionPhases)-1]
}

// -- Commands --

// SetTechStack sets the tech stack for this session.
func (s *DiscoverySession) SetTechStack(ts *vo.TechStack) error {
	if s.status != StatusCreated && s.status != StatusPersonaDetected {
		return fmt.Errorf("can only set tech stack in CREATED or PERSONA_DETECTED state, currently %s: %w",
			s.status, domainerrors.ErrInvariantViolation)
	}
	s.techStack = ts
	return nil
}

// SetMode sets the discovery mode. Only allowed once, in CREATED state.
func (s *DiscoverySession) SetMode(mode DiscoveryMode) error {
	if s.status != StatusCreated {
		return fmt.Errorf("can only set mode in CREATED state, currently %s: %w",
			s.status, domainerrors.ErrInvariantViolation)
	}
	if s.mode != nil {
		return fmt.Errorf("discovery mode has already been set: %w", domainerrors.ErrInvariantViolation)
	}
	s.mode = &mode
	return nil
}

// DetectPersona sets the user persona and language register from a choice string.
func (s *DiscoverySession) DetectPersona(choice string) error {
	if s.status != StatusCreated {
		return fmt.Errorf("can only detect persona in CREATED state, currently %s: %w",
			s.status, domainerrors.ErrInvariantViolation)
	}
	pc, ok := personaChoices[choice]
	if !ok {
		return fmt.Errorf("invalid persona choice '%s': must be '1', '2', '3', or '4'", choice)
	}
	s.persona = &pc.persona
	s.register = &pc.register
	s.status = StatusPersonaDetected
	return nil
}

// AnswerQuestion records an answer to a discovery question.
func (s *DiscoverySession) AnswerQuestion(questionID, response string) error {
	if s.status == StatusCreated {
		return fmt.Errorf("cannot answer questions before persona is detected: %w", domainerrors.ErrInvariantViolation)
	}
	if s.status == StatusPlaybackPending {
		return fmt.Errorf("must confirm playback before answering more questions: %w", domainerrors.ErrInvariantViolation)
	}
	if s.status != StatusPersonaDetected && s.status != StatusAnswering {
		return fmt.Errorf("cannot answer in %s state: %w", s.status, domainerrors.ErrInvariantViolation)
	}
	if strings.TrimSpace(response) == "" {
		return fmt.Errorf("answer cannot be empty")
	}

	// Check duplicates
	for _, a := range s.answers {
		if a.QuestionID() == questionID {
			return fmt.Errorf("question '%s' already answered: %w", questionID, domainerrors.ErrInvariantViolation)
		}
	}

	// Lookup question
	qByID := QuestionByID()
	question, ok := qByID[questionID]
	if !ok {
		return fmt.Errorf("unknown question '%s'", questionID)
	}

	// Enforce phase order
	if err := s.enforcePhaseOrder(question); err != nil {
		return err
	}

	s.answers = append(s.answers, NewAnswer(questionID, response))
	s.answersSinceLastPlayback++
	s.status = StatusAnswering

	if s.answersSinceLastPlayback >= playbackInterval {
		s.status = StatusPlaybackPending
	}
	return nil
}

// SkipQuestion skips a question with an explicit reason.
func (s *DiscoverySession) SkipQuestion(questionID, reason string) error {
	if s.status == StatusPlaybackPending {
		return fmt.Errorf("must confirm playback before skipping questions: %w", domainerrors.ErrInvariantViolation)
	}
	if s.status != StatusPersonaDetected && s.status != StatusAnswering {
		return fmt.Errorf("cannot skip questions in %s state: %w", s.status, domainerrors.ErrInvariantViolation)
	}
	qByID := QuestionByID()
	if _, ok := qByID[questionID]; !ok {
		return fmt.Errorf("unknown question '%s'", questionID)
	}
	if strings.TrimSpace(reason) == "" {
		return fmt.Errorf("skip reason cannot be empty")
	}
	s.skipped[questionID] = true
	return nil
}

// ConfirmPlayback confirms or rejects a playback summary.
func (s *DiscoverySession) ConfirmPlayback(confirmed bool, corrections string) error {
	if s.status != StatusPlaybackPending {
		return fmt.Errorf("can only confirm playback in PLAYBACK_PENDING state, currently %s: %w",
			s.status, domainerrors.ErrInvariantViolation)
	}
	summaryText := fmt.Sprintf("Playback %d", len(s.playbackConfirmations)+1)
	s.playbackConfirmations = append(s.playbackConfirmations, NewPlayback(summaryText, confirmed, corrections))
	s.answersSinceLastPlayback = 0
	s.status = StatusAnswering
	return nil
}

// Complete completes the discovery session.
func (s *DiscoverySession) Complete() error {
	if s.status != StatusAnswering {
		return fmt.Errorf("can only complete from ANSWERING state, currently %s: %w",
			s.status, domainerrors.ErrInvariantViolation)
	}

	// Check MVP questions
	answeredIDs := make(map[string]bool)
	for _, a := range s.answers {
		answeredIDs[a.QuestionID()] = true
	}
	mvpIDs := MVPQuestionIDs()
	var missing []string
	for id := range mvpIDs {
		if !answeredIDs[id] {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("cannot complete: MVP questions not answered: %v: %w",
			missing, domainerrors.ErrInvariantViolation)
	}

	if s.Mode() == ModeDeep {
		s.status = StatusRound1Complete
		round := RoundDiscovery
		s.round = &round
		return nil
	}

	s.status = StatusCompleted
	round := RoundDiscovery
	s.round = &round
	s.emitCompletedEvent()
	return nil
}

// StartChallenge transitions to CHALLENGING. Only from ROUND_1_COMPLETE in DEEP mode.
func (s *DiscoverySession) StartChallenge() error {
	if s.status != StatusRound1Complete {
		return fmt.Errorf("can only start challenge from ROUND_1_COMPLETE state, currently %s: %w",
			s.status, domainerrors.ErrInvariantViolation)
	}
	if s.Mode() != ModeDeep {
		return fmt.Errorf("start_challenge() is only available in DEEP mode: %w", domainerrors.ErrInvariantViolation)
	}
	s.status = StatusChallenging
	round := RoundChallenge
	s.round = &round
	return nil
}

// CompleteChallenge transitions to ROUND_2_COMPLETE from CHALLENGING.
func (s *DiscoverySession) CompleteChallenge() error {
	if s.status != StatusChallenging {
		return fmt.Errorf("can only complete challenge from CHALLENGING state, currently %s: %w",
			s.status, domainerrors.ErrInvariantViolation)
	}
	s.status = StatusRound2Complete
	return nil
}

// StartSimulate transitions to SIMULATING from ROUND_2_COMPLETE in DEEP mode.
func (s *DiscoverySession) StartSimulate() error {
	if s.status != StatusRound2Complete {
		return fmt.Errorf("can only start simulation from ROUND_2_COMPLETE state, currently %s: %w",
			s.status, domainerrors.ErrInvariantViolation)
	}
	if s.Mode() != ModeDeep {
		return fmt.Errorf("start_simulate() is only available in DEEP mode: %w", domainerrors.ErrInvariantViolation)
	}
	s.status = StatusSimulating
	round := RoundSimulate
	s.round = &round
	return nil
}

// CompleteSimulation transitions to COMPLETED from SIMULATING.
func (s *DiscoverySession) CompleteSimulation() error {
	if s.status != StatusSimulating {
		return fmt.Errorf("can only complete simulation from SIMULATING state, currently %s: %w",
			s.status, domainerrors.ErrInvariantViolation)
	}
	s.status = StatusCompleted
	s.emitCompletedEvent()
	return nil
}

func (s *DiscoverySession) emitCompletedEvent() {
	s.events = append(s.events, NewDiscoveryCompletedEvent(
		s.sessionID,
		*s.persona,
		*s.register,
		s.answers,
		s.playbackConfirmations,
		s.techStack,
	))
}

func (s *DiscoverySession) enforcePhaseOrder(question Question) error {
	targetIdx := -1
	for i, p := range questionPhases {
		if p == question.Phase() {
			targetIdx = i
			break
		}
	}
	if targetIdx < 0 {
		return nil // SEED phase always allowed
	}

	allHandled := make(map[string]bool)
	for _, a := range s.answers {
		allHandled[a.QuestionID()] = true
	}
	for id := range s.skipped {
		allHandled[id] = true
	}

	catalog := QuestionCatalog()
	for i := 0; i < targetIdx; i++ {
		earlierPhase := questionPhases[i]
		for _, q := range catalog {
			if q.Phase() == earlierPhase && !allHandled[q.ID()] {
				return fmt.Errorf("cannot answer %s (%s phase) before completing %s phase (question %s not answered or skipped): %w",
					question.ID(), question.Phase(), earlierPhase, q.ID(), domainerrors.ErrInvariantViolation)
			}
		}
	}
	return nil
}

// -- Serialization --

// ToSnapshot serializes session state to a map.
func (s *DiscoverySession) ToSnapshot() map[string]interface{} {
	answers := make([]map[string]string, len(s.answers))
	for i, a := range s.answers {
		answers[i] = map[string]string{
			"question_id":   a.QuestionID(),
			"response_text": a.ResponseText(),
		}
	}

	skipped := make([]string, 0, len(s.skipped))
	for id := range s.skipped {
		skipped = append(skipped, id)
	}

	playbacks := make([]map[string]interface{}, len(s.playbackConfirmations))
	for i, p := range s.playbackConfirmations {
		playbacks[i] = map[string]interface{}{
			"summary_text": p.SummaryText(),
			"confirmed":    p.Confirmed(),
			"corrections":  p.Corrections(),
		}
	}

	var personaVal, registerVal interface{}
	if s.persona != nil {
		personaVal = string(*s.persona)
	}
	if s.register != nil {
		registerVal = string(*s.register)
	}

	var modeVal, roundVal interface{}
	if s.mode != nil {
		modeVal = string(*s.mode)
	}
	if s.round != nil {
		roundVal = string(*s.round)
	}

	var techStackVal interface{}
	if s.techStack != nil {
		techStackVal = map[string]string{
			"language":        s.techStack.Language(),
			"package_manager": s.techStack.PackageManager(),
		}
	}

	return map[string]interface{}{
		"session_id":                  s.sessionID,
		"readme_content":              s.readmeContent,
		"status":                      string(s.status),
		"persona":                     personaVal,
		"register":                    registerVal,
		"answers":                     answers,
		"skipped":                     skipped,
		"playback_confirmations":      playbacks,
		"answers_since_last_playback": s.answersSinceLastPlayback,
		"mode":                        modeVal,
		"round":                       roundVal,
		"tech_stack":                  techStackVal,
	}
}

// FromSnapshot reconstructs a DiscoverySession from a snapshot map.
func FromSnapshot(data map[string]interface{}) (*DiscoverySession, error) {
	// Validate required keys
	required := []string{
		"session_id", "readme_content", "status", "persona", "register",
		"answers", "skipped", "playback_confirmations", "answers_since_last_playback",
	}
	for _, key := range required {
		if _, ok := data[key]; !ok {
			return nil, fmt.Errorf("snapshot missing required field: %s", key)
		}
	}

	// Parse status
	statusStr, _ := data["status"].(string)
	status := DiscoveryStatus(statusStr)
	switch status {
	case StatusCreated, StatusPersonaDetected, StatusAnswering, StatusPlaybackPending,
		StatusCompleted, StatusCancelled, StatusRound1Complete, StatusChallenging,
		StatusRound2Complete, StatusSimulating:
		// valid
	default:
		return nil, fmt.Errorf("invalid status: %q", statusStr)
	}

	// Parse persona
	var persona *DiscoveryPersona
	if pVal := data["persona"]; pVal != nil {
		pStr, ok := pVal.(string)
		if !ok {
			return nil, fmt.Errorf("invalid persona type")
		}
		p, err := ParseDiscoveryPersona(pStr)
		if err != nil {
			return nil, err
		}
		persona = &p
	}

	// Parse register
	var register *DiscoveryRegister
	if rVal := data["register"]; rVal != nil {
		rStr, ok := rVal.(string)
		if !ok {
			return nil, fmt.Errorf("invalid register type")
		}
		r, err := ParseDiscoveryRegister(rStr)
		if err != nil {
			return nil, err
		}
		register = &r
	}

	// Cross-validate status vs persona
	if status == StatusCreated && persona != nil {
		return nil, fmt.Errorf("CREATED state must have persona=nil")
	}
	if status != StatusCreated && persona == nil {
		return nil, fmt.Errorf("%s state requires a persona", status)
	}

	// Parse answers - handle both direct map and JSON-decoded formats
	var answers []Answer
	switch raw := data["answers"].(type) {
	case []interface{}:
		answers = make([]Answer, len(raw))
		for i, item := range raw {
			m, ok := item.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid answer format")
			}
			qid, _ := m["question_id"].(string)
			rt, _ := m["response_text"].(string)
			answers[i] = NewAnswer(qid, rt)
		}
	case []map[string]string:
		answers = make([]Answer, len(raw))
		for i, m := range raw {
			answers[i] = NewAnswer(m["question_id"], m["response_text"])
		}
	default:
		return nil, fmt.Errorf("answers must be a list")
	}

	// Parse skipped
	skipped := make(map[string]bool)
	switch raw := data["skipped"].(type) {
	case []interface{}:
		for _, item := range raw {
			s, _ := item.(string)
			skipped[s] = true
		}
	case []string:
		for _, s := range raw {
			skipped[s] = true
		}
	default:
		return nil, fmt.Errorf("skipped must be a list")
	}

	// Parse playback confirmations
	var playbacks []Playback
	switch raw := data["playback_confirmations"].(type) {
	case []interface{}:
		playbacks = make([]Playback, len(raw))
		for i, item := range raw {
			m, ok := item.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid playback format")
			}
			st, _ := m["summary_text"].(string)
			confirmed, _ := m["confirmed"].(bool)
			corrections, _ := m["corrections"].(string)
			playbacks[i] = NewPlayback(st, confirmed, corrections)
		}
	case []map[string]interface{}:
		playbacks = make([]Playback, len(raw))
		for i, m := range raw {
			st, _ := m["summary_text"].(string)
			confirmed, _ := m["confirmed"].(bool)
			corrections, _ := m["corrections"].(string)
			playbacks[i] = NewPlayback(st, confirmed, corrections)
		}
	default:
		return nil, fmt.Errorf("playback_confirmations must be a list")
	}

	// Parse counter - handle both int and float64 (from JSON)
	var counter int
	switch v := data["answers_since_last_playback"].(type) {
	case float64:
		counter = int(v)
	case int:
		counter = v
	default:
		return nil, fmt.Errorf("answers_since_last_playback must be a number")
	}
	if counter < 0 {
		return nil, fmt.Errorf("answers_since_last_playback must be a non-negative integer")
	}
	if counter > playbackInterval {
		return nil, fmt.Errorf("answers_since_last_playback (%d) exceeds playback interval (%d)", counter, playbackInterval)
	}

	// Cross-validate counter
	if status == StatusPlaybackPending && counter != playbackInterval {
		return nil, fmt.Errorf("PLAYBACK_PENDING state requires counter=%d, got %d", playbackInterval, counter)
	}
	if status == StatusAnswering && counter >= playbackInterval {
		return nil, fmt.Errorf("ANSWERING state requires counter < %d, got %d", playbackInterval, counter)
	}

	// Parse mode
	var mode *DiscoveryMode
	if mVal, exists := data["mode"]; exists && mVal != nil {
		mStr, ok := mVal.(string)
		if !ok {
			return nil, fmt.Errorf("invalid mode type")
		}
		m, err := ParseDiscoveryMode(mStr)
		if err != nil {
			return nil, err
		}
		mode = &m
	}

	// Parse round
	var round *DiscoveryRound
	if rVal, exists := data["round"]; exists && rVal != nil {
		rStr, ok := rVal.(string)
		if !ok {
			return nil, fmt.Errorf("invalid round type")
		}
		r, err := ParseDiscoveryRound(rStr)
		if err != nil {
			return nil, err
		}
		round = &r
	}

	// Parse tech stack
	var techStack *vo.TechStack
	if tsVal, exists := data["tech_stack"]; exists && tsVal != nil {
		switch tsMap := tsVal.(type) {
		case map[string]interface{}:
			lang, _ := tsMap["language"].(string)
			pm, _ := tsMap["package_manager"].(string)
			ts := vo.NewTechStack(lang, pm)
			techStack = &ts
		case map[string]string:
			ts := vo.NewTechStack(tsMap["language"], tsMap["package_manager"])
			techStack = &ts
		default:
			return nil, fmt.Errorf("invalid tech_stack format")
		}
	}

	return &DiscoverySession{
		sessionID:                toString(data["session_id"]),
		readmeContent:            toString(data["readme_content"]),
		status:                   status,
		persona:                  persona,
		register:                 register,
		answers:                  answers,
		skipped:                  skipped,
		playbackConfirmations:    playbacks,
		answersSinceLastPlayback: counter,
		techStack:                techStack,
		mode:                     mode,
		round:                    round,
	}, nil
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
