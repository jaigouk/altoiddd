package application

import (
	"fmt"
	"sync"

	"github.com/alty-cli/alty/internal/discovery/domain"
)

// DiscoveryHandler orchestrates the discovery session lifecycle.
type DiscoveryHandler struct {
	mu       sync.Mutex
	sessions map[string]*domain.DiscoverySession
}

// NewDiscoveryHandler creates a new DiscoveryHandler.
func NewDiscoveryHandler() *DiscoveryHandler {
	return &DiscoveryHandler{
		sessions: make(map[string]*domain.DiscoverySession),
	}
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
