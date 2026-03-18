package application

import (
	"context"
	"fmt"
	"sync"

	"github.com/alto-cli/alto/internal/discovery/domain"
	sharedapp "github.com/alto-cli/alto/internal/shared/application"
)

// DiscoveryHandler orchestrates the discovery session lifecycle.
type DiscoveryHandler struct {
	publisher   sharedapp.EventPublisher
	sessionRepo SessionRepository
	mu          sync.Mutex
	sessions    map[string]*domain.DiscoverySession
}

// HandlerOption configures optional dependencies for DiscoveryHandler.
type HandlerOption func(*DiscoveryHandler)

// WithSessionRepository injects an optional SessionRepository for persistence.
func WithSessionRepository(repo SessionRepository) HandlerOption {
	return func(h *DiscoveryHandler) {
		h.sessionRepo = repo
	}
}

// NewDiscoveryHandler creates a new DiscoveryHandler.
func NewDiscoveryHandler(publisher sharedapp.EventPublisher, opts ...HandlerOption) *DiscoveryHandler {
	h := &DiscoveryHandler{
		publisher: publisher,
		sessions:  make(map[string]*domain.DiscoverySession),
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// StartSession starts a new discovery session from README content.
func (h *DiscoveryHandler) StartSession(readmeContent string) (*domain.DiscoverySession, error) {
	session := domain.NewDiscoverySession(readmeContent)
	h.mu.Lock()
	h.sessions[session.SessionID()] = session
	h.mu.Unlock()
	return session, nil
}

// DetectPersona detects user persona for the given session.
func (h *DiscoveryHandler) DetectPersona(sessionID, choice string) (*domain.DiscoverySession, error) {
	session, err := h.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	if err := session.DetectPersona(choice); err != nil {
		return nil, fmt.Errorf("detect persona: %w", err)
	}
	return session, nil
}

// AnswerQuestion submits an answer to a discovery question.
func (h *DiscoveryHandler) AnswerQuestion(sessionID, questionID, answer string) (*domain.DiscoverySession, error) {
	session, err := h.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	if err := session.AnswerQuestion(questionID, answer); err != nil {
		return nil, fmt.Errorf("answer question %s: %w", questionID, err)
	}
	if err := h.persistSession(session); err != nil {
		return nil, fmt.Errorf("persisting session after answer: %w", err)
	}
	return session, nil
}

// SkipQuestion skips a question with an explicit reason.
func (h *DiscoveryHandler) SkipQuestion(sessionID, questionID, reason string) (*domain.DiscoverySession, error) {
	session, err := h.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	if err := session.SkipQuestion(questionID, reason); err != nil {
		return nil, fmt.Errorf("skip question %s: %w", questionID, err)
	}
	if err := h.persistSession(session); err != nil {
		return nil, fmt.Errorf("persisting session after skip: %w", err)
	}
	return session, nil
}

// ConfirmPlayback confirms or rejects a playback summary.
func (h *DiscoveryHandler) ConfirmPlayback(sessionID string, confirmed bool) (*domain.DiscoverySession, error) {
	session, err := h.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	if err := session.ConfirmPlayback(confirmed, ""); err != nil {
		return nil, fmt.Errorf("confirm playback: %w", err)
	}
	return session, nil
}

// Complete completes the discovery session.
func (h *DiscoveryHandler) Complete(sessionID string) (*domain.DiscoverySession, error) {
	session, err := h.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	if err := session.Complete(); err != nil {
		return nil, fmt.Errorf("complete session: %w", err)
	}
	for _, event := range session.Events() {
		_ = h.publisher.Publish(context.Background(), event)
	}
	return session, nil
}

// LoadOrGetSession retrieves a session from memory first, then falls back to the
// SessionRepository if configured. Loaded sessions are cached in memory for
// subsequent calls.
func (h *DiscoveryHandler) LoadOrGetSession(sessionID string) (*domain.DiscoverySession, error) {
	// 1. Try in-memory first
	h.mu.Lock()
	session, ok := h.sessions[sessionID]
	h.mu.Unlock()
	if ok {
		return session, nil
	}

	// 2. Try repository if available
	if h.sessionRepo == nil {
		return nil, fmt.Errorf("no active discovery session with id '%s'", sessionID)
	}

	session, err := h.sessionRepo.Load(context.TODO(), sessionID)
	if err != nil {
		return nil, fmt.Errorf("loading session '%s': %w", sessionID, err)
	}

	// 3. Cache in memory for subsequent calls
	h.mu.Lock()
	h.sessions[session.SessionID()] = session
	h.mu.Unlock()

	return session, nil
}

// GetSession retrieves an active discovery session by ID.
func (h *DiscoveryHandler) GetSession(sessionID string) (*domain.DiscoverySession, error) {
	h.mu.Lock()
	session, ok := h.sessions[sessionID]
	h.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("no active discovery session with id '%s'", sessionID)
	}
	return session, nil
}

// persistSession saves the session if a SessionRepository is configured.
// Returns nil when no repository is set (nil-safe).
// Uses context.TODO() because the Discovery interface deliberately omits context
// for synchronous CLI operations.
func (h *DiscoveryHandler) persistSession(session *domain.DiscoverySession) error {
	if h.sessionRepo == nil {
		return nil
	}
	if err := h.sessionRepo.Save(context.TODO(), session); err != nil {
		return fmt.Errorf("saving session: %w", err)
	}
	return nil
}

// ClassifySubdomain classifies a bounded context using the Khononov decision tree.
func (h *DiscoveryHandler) ClassifySubdomain(sessionID, contextName string, buyYes, complexRules, competitorThreat bool) (*domain.ClassificationResult, error) {
	session, err := h.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	tree := domain.NewClassificationDecisionTree()
	result := tree.Classify(buyYes, complexRules, competitorThreat)
	if err := session.ClassifyBoundedContext(contextName, result); err != nil {
		return nil, fmt.Errorf("classify bounded context %s: %w", contextName, err)
	}
	return &result, nil
}
