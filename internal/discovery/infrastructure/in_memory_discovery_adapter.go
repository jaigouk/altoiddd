package infrastructure

import (
	"context"
	"fmt"

	discoverydomain "github.com/alty-cli/alty/internal/discovery/domain"
	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
	"github.com/alty-cli/alty/internal/shared/infrastructure/persistence"
)

// InMemoryDiscoveryAdapter is an in-memory adapter for guided discovery sessions.
// Wraps SessionStore and DiscoverySession aggregate, delegating all state
// transitions to the domain model.
type InMemoryDiscoveryAdapter struct {
	store *persistence.SessionStore
}

// NewInMemoryDiscoveryAdapter creates a new InMemoryDiscoveryAdapter.
func NewInMemoryDiscoveryAdapter(store *persistence.SessionStore) *InMemoryDiscoveryAdapter {
	return &InMemoryDiscoveryAdapter{store: store}
}

func (a *InMemoryDiscoveryAdapter) getSession(sessionID string) (*discoverydomain.DiscoverySession, error) {
	val, err := a.store.Get(sessionID)
	if err != nil {
		return nil, fmt.Errorf("session '%s': %w", sessionID, domainerrors.ErrNotFound)
	}
	session, ok := val.(*discoverydomain.DiscoverySession)
	if !ok {
		return nil, fmt.Errorf("session '%s' has unexpected type: %w", sessionID, domainerrors.ErrNotFound)
	}
	return session, nil
}

// --- Discovery port methods (no context.Context) ---

// StartSession starts a new guided discovery session from README content.
func (a *InMemoryDiscoveryAdapter) StartSession(readmeContent string) (*discoverydomain.DiscoverySession, error) {
	session := discoverydomain.NewDiscoverySession(readmeContent)
	a.store.Put(session.SessionID(), session)
	return session, nil
}

// DetectPersona detects the user persona based on their choice.
func (a *InMemoryDiscoveryAdapter) DetectPersona(sessionID string, choice string) (*discoverydomain.DiscoverySession, error) {
	session, err := a.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	if err := session.DetectPersona(choice); err != nil {
		return nil, err
	}
	a.store.Put(sessionID, session)
	return session, nil
}

// AnswerQuestion submits an answer to a discovery question.
func (a *InMemoryDiscoveryAdapter) AnswerQuestion(sessionID string, questionID string, answer string) (*discoverydomain.DiscoverySession, error) {
	session, err := a.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	if err := session.AnswerQuestion(questionID, answer); err != nil {
		return nil, err
	}
	a.store.Put(sessionID, session)
	return session, nil
}

// SkipQuestion skips a question with an explicit reason.
func (a *InMemoryDiscoveryAdapter) SkipQuestion(sessionID string, questionID string, reason string) (*discoverydomain.DiscoverySession, error) {
	session, err := a.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	if err := session.SkipQuestion(questionID, reason); err != nil {
		return nil, err
	}
	a.store.Put(sessionID, session)
	return session, nil
}

// ConfirmPlayback confirms or rejects the playback summary.
func (a *InMemoryDiscoveryAdapter) ConfirmPlayback(sessionID string, confirmed bool) (*discoverydomain.DiscoverySession, error) {
	session, err := a.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	if err := session.ConfirmPlayback(confirmed, ""); err != nil {
		return nil, err
	}
	a.store.Put(sessionID, session)
	return session, nil
}

// Complete completes the discovery session and produces domain artifacts.
func (a *InMemoryDiscoveryAdapter) Complete(sessionID string) (*discoverydomain.DiscoverySession, error) {
	session, err := a.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	if err := session.Complete(); err != nil {
		return nil, err
	}
	a.store.Put(sessionID, session)
	return session, nil
}

// --- Adapter-only methods (not part of Discovery port) ---

// GetSession retrieves a discovery session by ID.
func (a *InMemoryDiscoveryAdapter) GetSession(_ context.Context, sessionID string) (*discoverydomain.DiscoverySession, error) {
	return a.getSession(sessionID)
}

// SetTechStack sets the tech stack on a discovery session.
func (a *InMemoryDiscoveryAdapter) SetTechStack(_ context.Context, sessionID string, techStack vo.TechStack) (*discoverydomain.DiscoverySession, error) {
	session, err := a.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	if err := session.SetTechStack(&techStack); err != nil {
		return nil, err
	}
	a.store.Put(sessionID, session)
	return session, nil
}

// SetMode sets the discovery mode on a session.
func (a *InMemoryDiscoveryAdapter) SetMode(_ context.Context, sessionID string, mode discoverydomain.DiscoveryMode) (*discoverydomain.DiscoverySession, error) {
	session, err := a.getSession(sessionID)
	if err != nil {
		return nil, err
	}
	if err := session.SetMode(mode); err != nil {
		return nil, err
	}
	a.store.Put(sessionID, session)
	return session, nil
}
