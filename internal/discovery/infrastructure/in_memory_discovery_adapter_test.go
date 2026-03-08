package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/discovery/application"
	discoverydomain "github.com/alty-cli/alty/internal/discovery/domain"
	"github.com/alty-cli/alty/internal/discovery/infrastructure"
	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
	"github.com/alty-cli/alty/internal/shared/infrastructure/persistence"
)

// Compile-time interface check.
var _ application.Discovery = (*infrastructure.InMemoryDiscoveryAdapter)(nil)

const testREADME = "A test project idea in 4-5 sentences."

var allQuestionIDs = []string{"Q1", "Q2", "Q3", "Q4", "Q5", "Q6", "Q7", "Q8", "Q9", "Q10"}

func makeAdapter() *infrastructure.InMemoryDiscoveryAdapter {
	store := persistence.NewSessionStore(60 * time.Second)
	return infrastructure.NewInMemoryDiscoveryAdapter(store)
}

func startWithPersona(t *testing.T, adapter *infrastructure.InMemoryDiscoveryAdapter, choice string) string {
	t.Helper()
	session, err := adapter.StartSession(testREADME)
	require.NoError(t, err)
	_, err = adapter.DetectPersona(session.SessionID(), choice)
	require.NoError(t, err)
	return session.SessionID()
}

func answerAllQuestions(t *testing.T, adapter *infrastructure.InMemoryDiscoveryAdapter, sessionID string) {
	t.Helper()
	for _, qid := range allQuestionIDs {
		session, err := adapter.AnswerQuestion(sessionID, qid, "Answer for "+qid)
		require.NoError(t, err)
		if session.Status() == discoverydomain.StatusPlaybackPending {
			_, err := adapter.ConfirmPlayback(sessionID, true)
			require.NoError(t, err)
		}
	}
}

// -- Happy Path --

func TestHappyPath_StartSession(t *testing.T) {
	t.Parallel()
	adapter := makeAdapter()
	session, err := adapter.StartSession(testREADME)
	require.NoError(t, err)
	assert.NotEmpty(t, session.SessionID())
	assert.Equal(t, discoverydomain.StatusCreated, session.Status())
	assert.Equal(t, testREADME, session.ReadmeContent())
}

func TestHappyPath_DetectPersona(t *testing.T) {
	t.Parallel()
	adapter := makeAdapter()
	session, err := adapter.StartSession(testREADME)
	require.NoError(t, err)
	updated, err := adapter.DetectPersona(session.SessionID(), "1")
	require.NoError(t, err)
	assert.Equal(t, discoverydomain.StatusPersonaDetected, updated.Status())
}

func TestHappyPath_AnswerQuestion(t *testing.T) {
	t.Parallel()
	adapter := makeAdapter()
	sid := startWithPersona(t, adapter, "1")
	updated, err := adapter.AnswerQuestion(sid, "Q1", "Users and admins")
	require.NoError(t, err)
	assert.Len(t, updated.Answers(), 1)
	assert.Equal(t, "Q1", updated.Answers()[0].QuestionID())
}

func TestHappyPath_SkipQuestion(t *testing.T) {
	t.Parallel()
	adapter := makeAdapter()
	sid := startWithPersona(t, adapter, "1")
	_, err := adapter.AnswerQuestion(sid, "Q1", "Actors")
	require.NoError(t, err)
	updated, err := adapter.SkipQuestion(sid, "Q2", "Not relevant")
	require.NoError(t, err)
	answerIDs := make(map[string]bool)
	for _, a := range updated.Answers() {
		answerIDs[a.QuestionID()] = true
	}
	assert.False(t, answerIDs["Q2"])
}

func TestHappyPath_ConfirmPlayback(t *testing.T) {
	t.Parallel()
	adapter := makeAdapter()
	sid := startWithPersona(t, adapter, "1")
	_, err := adapter.AnswerQuestion(sid, "Q1", "Actors")
	require.NoError(t, err)
	_, err = adapter.AnswerQuestion(sid, "Q2", "Entities")
	require.NoError(t, err)
	session, err := adapter.AnswerQuestion(sid, "Q3", "Use case")
	require.NoError(t, err)
	assert.Equal(t, discoverydomain.StatusPlaybackPending, session.Status())
	updated, err := adapter.ConfirmPlayback(sid, true)
	require.NoError(t, err)
	assert.Equal(t, discoverydomain.StatusAnswering, updated.Status())
}

func TestHappyPath_Complete(t *testing.T) {
	t.Parallel()
	adapter := makeAdapter()
	sid := startWithPersona(t, adapter, "1")
	answerAllQuestions(t, adapter, sid)
	completed, err := adapter.Complete(sid)
	require.NoError(t, err)
	assert.Equal(t, discoverydomain.StatusCompleted, completed.Status())
	assert.Len(t, completed.Events(), 1)
}

func TestHappyPath_GetSession(t *testing.T) {
	t.Parallel()
	adapter := makeAdapter()
	session, err := adapter.StartSession(testREADME)
	require.NoError(t, err)
	// GetSession is an adapter-only method (not on the Discovery port)
	retrieved, err := adapter.GetSession(context.TODO(), session.SessionID())
	require.NoError(t, err)
	assert.Equal(t, session.SessionID(), retrieved.SessionID())
}

func TestHappyPath_MultipleSessions(t *testing.T) {
	t.Parallel()
	adapter := makeAdapter()
	s1, err := adapter.StartSession("Project A.")
	require.NoError(t, err)
	s2, err := adapter.StartSession("Project B.")
	require.NoError(t, err)
	assert.NotEqual(t, s1.SessionID(), s2.SessionID())
}

// -- Error Propagation --

func TestError_GetSessionNotFound(t *testing.T) {
	t.Parallel()
	adapter := makeAdapter()
	_, err := adapter.GetSession(context.TODO(), "nonexistent-id")
	require.ErrorIs(t, err, domainerrors.ErrNotFound)
}

func TestError_DetectPersonaInvalidChoice(t *testing.T) {
	t.Parallel()
	adapter := makeAdapter()
	session, err := adapter.StartSession(testREADME)
	require.NoError(t, err)
	_, err = adapter.DetectPersona(session.SessionID(), "5")
	require.Error(t, err)
}

func TestError_AnswerEmptyResponse(t *testing.T) {
	t.Parallel()
	adapter := makeAdapter()
	sid := startWithPersona(t, adapter, "1")
	_, err := adapter.AnswerQuestion(sid, "Q1", "")
	require.Error(t, err)
}

func TestError_AnswerUnknownQuestion(t *testing.T) {
	t.Parallel()
	adapter := makeAdapter()
	sid := startWithPersona(t, adapter, "1")
	_, err := adapter.AnswerQuestion(sid, "Q99", "Some answer")
	require.Error(t, err)
}

func TestError_CompleteWithoutEnoughAnswers(t *testing.T) {
	t.Parallel()
	adapter := makeAdapter()
	sid := startWithPersona(t, adapter, "1")
	_, err := adapter.AnswerQuestion(sid, "Q1", "Actors")
	require.NoError(t, err)
	_, err = adapter.AnswerQuestion(sid, "Q2", "Entities")
	require.NoError(t, err)
	_, err = adapter.AnswerQuestion(sid, "Q3", "Use case")
	require.NoError(t, err)
	_, err = adapter.ConfirmPlayback(sid, true)
	require.NoError(t, err)
	_, err = adapter.Complete(sid)
	require.Error(t, err)
}
