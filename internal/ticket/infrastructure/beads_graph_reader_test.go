package infrastructure_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/ticket/application"
	"github.com/alto-cli/alto/internal/ticket/infrastructure"
)

// Compile-time interface check.
var _ application.BeadsGraphReader = (*infrastructure.BeadsGraphReader)(nil)

// stubRunner returns canned bytes per (command, first-arg) key.
type stubRunner struct {
	responses map[string][]byte
	errors    map[string]error
	calls     []stubCall
}

type stubCall struct {
	args []string
}

func (s *stubRunner) run(_ context.Context, args ...string) ([]byte, error) {
	s.calls = append(s.calls, stubCall{args: append([]string(nil), args...)})
	key := args[0]
	if e, ok := s.errors[key]; ok {
		return nil, e
	}
	if b, ok := s.responses[key]; ok {
		return b, nil
	}
	return []byte("[]"), nil
}

func TestBeadsGraphReader_NewBeadsGraphReader_Defaults(t *testing.T) {
	t.Parallel()

	r := infrastructure.NewBeadsGraphReader()

	require.NotNil(t, r)
	assert.Equal(t, 5*time.Second, r.Timeout())
}

func TestBeadsGraphReader_ReadCloseContext_ParsesBDShowJSON(t *testing.T) {
	t.Parallel()

	stub := &stubRunner{
		responses: map[string][]byte{
			"show": []byte(`[{"id":"alty-cli-2f9","status":"closed","close_reason":"Subcommand shipped"}]`),
		},
	}
	r := infrastructure.NewBeadsGraphReaderWithRunner(stub.run)

	got, err := r.ReadCloseContext(context.Background(), "alty-cli-2f9")

	require.NoError(t, err)
	assert.Equal(t, "Subcommand shipped", got)
}

func TestBeadsGraphReader_ReadCloseContext_EmptyWhenAbsent(t *testing.T) {
	t.Parallel()

	stub := &stubRunner{
		responses: map[string][]byte{
			"show": []byte(`[{"id":"alty-cli-2f9","status":"open"}]`),
		},
	}
	r := infrastructure.NewBeadsGraphReaderWithRunner(stub.run)

	got, err := r.ReadCloseContext(context.Background(), "alty-cli-2f9")

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestBeadsGraphReader_ReadParent_ExtractsParentChildDependency(t *testing.T) {
	t.Parallel()

	stub := &stubRunner{
		responses: map[string][]byte{
			"show": []byte(`[{
				"id":"alty-cli-child",
				"status":"open",
				"dependencies":[
					{"id":"alty-cli-other","status":"open","dependency_type":"blocks"},
					{"id":"alty-cli-epic","status":"open","dependency_type":"parent-child"}
				]
			}]`),
		},
	}
	r := infrastructure.NewBeadsGraphReaderWithRunner(stub.run)

	got, err := r.ReadParent(context.Background(), "alty-cli-child")

	require.NoError(t, err)
	assert.Equal(t, "alty-cli-epic", got)
}

func TestBeadsGraphReader_ReadParent_EmptyWhenNoParent(t *testing.T) {
	t.Parallel()

	stub := &stubRunner{
		responses: map[string][]byte{
			"show": []byte(`[{"id":"alty-cli-2f9","status":"open","dependencies":[]}]`),
		},
	}
	r := infrastructure.NewBeadsGraphReaderWithRunner(stub.run)

	got, err := r.ReadParent(context.Background(), "alty-cli-2f9")

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestBeadsGraphReader_ReadSiblings_ExcludesSelf(t *testing.T) {
	t.Parallel()

	stub := &stubRunner{
		responses: map[string][]byte{
			"children": []byte(`[
				{"id":"alty-cli-bf7","status":"open"},
				{"id":"alty-cli-self","status":"open"},
				{"id":"alty-cli-zc9","status":"closed"}
			]`),
		},
	}
	r := infrastructure.NewBeadsGraphReaderWithRunner(stub.run)

	got, err := r.ReadSiblings(context.Background(), "alty-cli-epic", "alty-cli-self")

	require.NoError(t, err)
	require.Len(t, got, 2, "self excluded; closed still returned (handler filters)")
	assert.Equal(t, "alty-cli-bf7", got[0].ID)
	assert.Equal(t, "alty-cli-zc9", got[1].ID)
	assert.Equal(t, "closed", got[1].Status)
}

func TestBeadsGraphReader_ReadSiblings_NoParentReturnsNil(t *testing.T) {
	t.Parallel()

	stub := &stubRunner{}
	r := infrastructure.NewBeadsGraphReaderWithRunner(stub.run)

	got, err := r.ReadSiblings(context.Background(), "", "alty-cli-2f9")

	require.NoError(t, err)
	assert.Empty(t, got)
	assert.Empty(t, stub.calls, "should not invoke bd children when parent ID is empty")
}

func TestBeadsGraphReader_ReadSiblings_BeadsHasNoChildrenIsSoftFailure(t *testing.T) {
	t.Parallel()

	stub := &stubRunner{
		errors: map[string]error{
			"children": errors.New("exit 1: Issue 'alty-cli-epic' has no children"),
		},
	}
	r := infrastructure.NewBeadsGraphReaderWithRunner(stub.run)

	got, err := r.ReadSiblings(context.Background(), "alty-cli-epic", "self")

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestBeadsGraphReader_ReadDependents_ParsesDependentsArray(t *testing.T) {
	t.Parallel()

	stub := &stubRunner{
		responses: map[string][]byte{
			"show": []byte(`[{
				"id":"alty-cli-2f9",
				"status":"closed",
				"dependents":[
					{"id":"alty-cli-bf7","status":"open","dependency_type":"blocks"},
					{"id":"alty-cli-unrelated","status":"open","dependency_type":"related"}
				]
			}]`),
		},
	}
	r := infrastructure.NewBeadsGraphReaderWithRunner(stub.run)

	got, err := r.ReadDependents(context.Background(), "alty-cli-2f9")

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "alty-cli-bf7", got[0].ID)
	assert.Equal(t, "open", got[0].Status)
}

func TestBeadsGraphReader_ReadRelated_MergesBothDirections(t *testing.T) {
	t.Parallel()

	stub := &stubRunner{
		responses: map[string][]byte{
			"show": []byte(`[{
				"id":"alty-cli-2f9",
				"status":"closed",
				"dependencies":[
					{"id":"alty-cli-aaa","status":"open","dependency_type":"related"},
					{"id":"alty-cli-blocker","status":"open","dependency_type":"blocks"}
				],
				"dependents":[
					{"id":"alty-cli-bbb","status":"open","dependency_type":"related"},
					{"id":"alty-cli-blocked","status":"open","dependency_type":"blocks"}
				]
			}]`),
		},
	}
	r := infrastructure.NewBeadsGraphReaderWithRunner(stub.run)

	got, err := r.ReadRelated(context.Background(), "alty-cli-2f9")

	require.NoError(t, err)
	require.Len(t, got, 2)
	ids := []string{got[0].ID, got[1].ID}
	assert.Contains(t, ids, "alty-cli-aaa")
	assert.Contains(t, ids, "alty-cli-bbb")
}

func TestBeadsGraphReader_ShowRecord_RejectsEmptyTicketID(t *testing.T) {
	t.Parallel()

	stub := &stubRunner{}
	r := infrastructure.NewBeadsGraphReaderWithRunner(stub.run)

	_, err := r.ReadCloseContext(context.Background(), "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ticket ID")
}

func TestBeadsGraphReader_ShowRecord_WhenBDExitsNonZero_ReturnsWrappedError(t *testing.T) {
	t.Parallel()

	stub := &stubRunner{
		errors: map[string]error{"show": errors.New("bd boom")},
	}
	r := infrastructure.NewBeadsGraphReaderWithRunner(stub.run)

	_, err := r.ReadParent(context.Background(), "alty-cli-2f9")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "running bd show alty-cli-2f9")
}

func TestBeadsGraphReader_ShowRecord_WhenJSONInvalid_ReturnsParseError(t *testing.T) {
	t.Parallel()

	stub := &stubRunner{
		responses: map[string][]byte{"show": []byte("not json")},
	}
	r := infrastructure.NewBeadsGraphReaderWithRunner(stub.run)

	_, err := r.ReadParent(context.Background(), "alty-cli-2f9")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing bd show json")
}

func TestBeadsGraphReader_ShowRecord_WhenEmptyArray_ReturnsError(t *testing.T) {
	t.Parallel()

	stub := &stubRunner{
		responses: map[string][]byte{"show": []byte("[]")},
	}
	r := infrastructure.NewBeadsGraphReaderWithRunner(stub.run)

	_, err := r.ReadCloseContext(context.Background(), "alty-cli-2f9")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no records")
}

func TestBeadsGraphReader_PassesTimeoutToRunnerContext(t *testing.T) {
	t.Parallel()

	var observedDeadline bool
	stub := &stubRunner{
		responses: map[string][]byte{"show": []byte(`[{"id":"x","status":"open"}]`)},
	}
	wrap := func(ctx context.Context, args ...string) ([]byte, error) {
		_, hasDeadline := ctx.Deadline()
		observedDeadline = hasDeadline
		return stub.run(ctx, args...)
	}
	r := infrastructure.NewBeadsGraphReaderWithRunner(wrap)

	_, err := r.ReadParent(context.Background(), "x")

	require.NoError(t, err)
	assert.True(t, observedDeadline, "adapter must apply its own timeout deadline")
}
