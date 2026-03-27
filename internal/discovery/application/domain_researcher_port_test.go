package application_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/application"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
)

// --- mockDomainResearcher ---

type mockDomainResearcher struct {
	result       *discoverydomain.DomainResearchResult
	err          error
	receivedCtx  context.Context //nolint:containedctx // test-only mock captures context for assertions
	receivedDesc string
}

func (m *mockDomainResearcher) Research(ctx context.Context, domainDescription string) (*discoverydomain.DomainResearchResult, error) {
	m.receivedCtx = ctx
	m.receivedDesc = domainDescription

	return m.result, m.err
}

// Compile-time interface satisfaction check.
var _ application.DomainResearcher = (*mockDomainResearcher)(nil)

func TestDomainResearcherPortContract_ReturnsNilNilWhenUnavailable(t *testing.T) {
	t.Parallel()

	// Given: mock returning (nil, nil) — research infrastructure unavailable
	mock := &mockDomainResearcher{result: nil, err: nil}

	// When: Research is called
	result, err := mock.Research(context.Background(), "test domain")

	// Then: both result and error are nil (ADR-013 graceful degradation)
	assert.Nil(t, result)
	assert.NoError(t, err)
}

func TestDomainResearcherPortContract_ReturnsResultWhenAvailable(t *testing.T) {
	t.Parallel()

	// Given: mock returning a populated DomainResearchResult
	meta := discoverydomain.NewSearchMetadata([]string{"q1"}, 10, 5, time.Second)

	drr, err := discoverydomain.NewDomainResearchResult(
		"e-commerce",
		meta,
		nil, // actors
		nil, // entities
		nil, // workflows
		nil, // failureModes
		nil, // regulatory
		nil, // software
	)
	require.NoError(t, err)

	mock := &mockDomainResearcher{result: &drr, err: nil}

	// When: Research is called
	result, err := mock.Research(context.Background(), "e-commerce platform")

	// Then: result is populated and error is nil
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "e-commerce", result.Domain())
}

func TestDomainResearcherPortContract_ReturnsNilErrOnFailure(t *testing.T) {
	t.Parallel()

	// Given: mock returning (nil, error) — unexpected failure
	mock := &mockDomainResearcher{result: nil, err: errors.New("network timeout")}

	// When: Research is called
	result, err := mock.Research(context.Background(), "test domain")

	// Then: result is nil and error is returned
	assert.Nil(t, result)
	require.Error(t, err)
	assert.ErrorContains(t, err, "network timeout")
}

func TestDomainResearcherPortContract_EmptyDescriptionPassesThrough(t *testing.T) {
	t.Parallel()

	// Given: mock that records the input description
	mock := &mockDomainResearcher{result: nil, err: nil}

	// When: Research is called with empty description
	_, _ = mock.Research(context.Background(), "")

	// Then: mock received "" — port does NOT validate input
	assert.Empty(t, mock.receivedDesc)
}

// --- QA Edge Case Tests ---

func TestDomainResearcher_Contract_ContextCancellationIsPassedToImplementation(t *testing.T) {
	t.Parallel()

	// Given: mock that captures ctx
	mock := &mockDomainResearcher{result: nil, err: nil}

	// And: a context that is already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	// When: Research called with the cancelled context
	_, _ = mock.Research(ctx, "some domain")

	// Then: the mock received the cancelled context (ctx.Err() != nil)
	require.NotNil(t, mock.receivedCtx)
	require.Error(t, mock.receivedCtx.Err(), "implementation must receive the cancelled context")
	assert.ErrorIs(t, mock.receivedCtx.Err(), context.Canceled)
}

func TestDomainResearcher_Contract_ConcurrentCallsNoRace(t *testing.T) {
	t.Parallel()

	// Given: separate mock instances per goroutine (no shared mutable state)
	const goroutines = 2

	var wg sync.WaitGroup

	wg.Add(goroutines)

	// When: 2 goroutines call Research simultaneously on separate mocks
	for range goroutines {
		go func() {
			defer wg.Done()

			m := &mockDomainResearcher{result: nil, err: nil}
			_, _ = m.Research(context.Background(), "concurrent domain")
		}()
	}

	// Then: no data race (test passes with -race flag)
	wg.Wait()
}

func TestDomainResearcher_Contract_ErrorWithResultReturnsError(t *testing.T) {
	t.Parallel()

	// Given: mock returning BOTH a non-nil result AND an error
	// The interface does not prevent this combination; callers must handle it defensively.
	meta := discoverydomain.NewSearchMetadata([]string{"q1"}, 10, 5, time.Second)

	drr, err := discoverydomain.NewDomainResearchResult(
		"partial-domain",
		meta,
		nil, nil, nil, nil, nil, nil,
	)
	require.NoError(t, err)

	mock := &mockDomainResearcher{
		result: &drr,
		err:    errors.New("partial failure: some sources unreachable"),
	}

	// When: Research is called
	result, resErr := mock.Research(context.Background(), "partial-domain description")

	// Then: callers should check error first — error takes precedence
	require.Error(t, resErr)
	require.ErrorContains(t, resErr, "partial failure")

	// And: result is also non-nil (defensive callers must not assume result==nil when err!=nil)
	assert.NotNil(t, result)
}

func TestDomainResearcher_Contract_LargeDescriptionPassesThrough(t *testing.T) {
	t.Parallel()

	// Given: mock that records the input description
	mock := &mockDomainResearcher{result: nil, err: nil}

	// When: Research called with a 10,000-character description
	largeDesc := strings.Repeat("a", 10000)
	_, _ = mock.Research(context.Background(), largeDesc)

	// Then: mock received the full 10,000-char string — port does not truncate
	assert.Len(t, mock.receivedDesc, 10000)
	assert.Equal(t, largeDesc, mock.receivedDesc)
}
