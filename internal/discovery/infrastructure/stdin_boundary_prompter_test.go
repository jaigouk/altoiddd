package infrastructure_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/discovery/application"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
	"github.com/alto-cli/alto/internal/discovery/infrastructure"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// Compile-time interface satisfaction check.
var _ application.BoundaryPrompter = (*infrastructure.StdinBoundaryPrompter)(nil)

// testBoundarySketch creates a valid BoundedContextSketch for testing.
func testBoundarySketch(t *testing.T, name string, classification vo.SubdomainClassification, confidence float64) discoverydomain.BoundedContextSketch {
	t.Helper()

	sketch, err := discoverydomain.NewBoundedContextSketch(
		name, classification, confidence,
		[]string{"Actor"}, []string{"Object"}, []string{"Story"},
		nil, vo.UserStated,
	)
	require.NoError(t, err)

	return sketch
}

func TestStdinBoundaryPrompter_DisplayBoundaryProposals_AcceptAll(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("all\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinBoundaryPrompter(reader, writer)

	sketches := []discoverydomain.BoundedContextSketch{
		testBoundarySketch(t, "OrderMgmt", vo.SubdomainCore, 0.85),
		testBoundarySketch(t, "Shipping", vo.SubdomainSupporting, 0.70),
		testBoundarySketch(t, "Billing", vo.SubdomainCore, 0.90),
	}

	accepted, err := p.DisplayBoundaryProposals(context.Background(), sketches)
	require.NoError(t, err)
	assert.Len(t, accepted, 3)
	assert.Equal(t, []string{"OrderMgmt", "Shipping", "Billing"}, accepted)
}

func TestStdinBoundaryPrompter_DisplayBoundaryProposals_AcceptSome(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("OrderMgmt Shipping\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinBoundaryPrompter(reader, writer)

	sketches := []discoverydomain.BoundedContextSketch{
		testBoundarySketch(t, "OrderMgmt", vo.SubdomainCore, 0.85),
		testBoundarySketch(t, "Shipping", vo.SubdomainSupporting, 0.70),
		testBoundarySketch(t, "Billing", vo.SubdomainCore, 0.90),
	}

	accepted, err := p.DisplayBoundaryProposals(context.Background(), sketches)
	require.NoError(t, err)
	assert.Equal(t, []string{"OrderMgmt", "Shipping"}, accepted)
}

func TestStdinBoundaryPrompter_DisplayBoundaryProposals_AcceptNone(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("none\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinBoundaryPrompter(reader, writer)

	sketches := []discoverydomain.BoundedContextSketch{
		testBoundarySketch(t, "OrderMgmt", vo.SubdomainCore, 0.85),
		testBoundarySketch(t, "Shipping", vo.SubdomainSupporting, 0.70),
		testBoundarySketch(t, "Billing", vo.SubdomainCore, 0.90),
	}

	accepted, err := p.DisplayBoundaryProposals(context.Background(), sketches)
	require.NoError(t, err)
	assert.Empty(t, accepted)
}

func TestStdinBoundaryPrompter_DisplayBoundaryProposals_DisplaysSketchInfo(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("all\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinBoundaryPrompter(reader, writer)

	sketches := []discoverydomain.BoundedContextSketch{
		testBoundarySketch(t, "OrderMgmt", vo.SubdomainCore, 0.85),
		testBoundarySketch(t, "Shipping", vo.SubdomainSupporting, 0.70),
		testBoundarySketch(t, "Billing", vo.SubdomainCore, 0.90),
	}

	_, err := p.DisplayBoundaryProposals(context.Background(), sketches)
	require.NoError(t, err)

	output := writer.String()
	// Each sketch name must appear in the output.
	assert.Contains(t, output, "OrderMgmt")
	assert.Contains(t, output, "Shipping")
	assert.Contains(t, output, "Billing")
	// Classification strings must appear.
	assert.Contains(t, strings.ToLower(output), "core")
	assert.Contains(t, strings.ToLower(output), "supporting")
	// Confidence values must appear.
	assert.Contains(t, output, fmt.Sprintf("%.2f", 0.85))
	assert.Contains(t, output, fmt.Sprintf("%.2f", 0.70))
	assert.Contains(t, output, fmt.Sprintf("%.2f", 0.90))
}

func TestStdinBoundaryPrompter_DisplayBoundaryProposals_EOF(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinBoundaryPrompter(reader, writer)

	sketches := []discoverydomain.BoundedContextSketch{
		testBoundarySketch(t, "OrderMgmt", vo.SubdomainCore, 0.85),
	}

	_, err := p.DisplayBoundaryProposals(context.Background(), sketches)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestStdinBoundaryPrompter_AskMissingContext_ReturnsText(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("PaymentProcessing\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinBoundaryPrompter(reader, writer)

	answer, err := p.AskMissingContext(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "PaymentProcessing", answer)
}

func TestStdinBoundaryPrompter_AskMissingContext_EmptyReturnsEmpty(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("\n")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinBoundaryPrompter(reader, writer)

	answer, err := p.AskMissingContext(context.Background())
	require.NoError(t, err)
	assert.Empty(t, answer)
}

func TestStdinBoundaryPrompter_AskMissingContext_EOF(t *testing.T) {
	t.Parallel()

	reader := bytes.NewBufferString("")
	writer := &bytes.Buffer{}
	p := infrastructure.NewStdinBoundaryPrompter(reader, writer)

	_, err := p.AskMissingContext(context.Background())
	assert.ErrorIs(t, err, context.Canceled)
}
