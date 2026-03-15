package infrastructure_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/docimport/infrastructure"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

const sampleDDD = `---
last_reviewed: 2026-02-22
owner: architecture
status: draft
---

# Domain-Driven Design Artifacts: sample-project

## 3. Subdomain Classification

### Summary

| Subdomain | Type | Rationale | Architecture Approach |
|-----------|------|-----------|----------------------|
| Orders | **Core** | Main differentiator | Hexagonal |
| Shipping | **Supporting** | Custom but not core | Layered |
| Payments | **Generic** | Commodity | ACL |

## 4. Bounded Contexts

### Context: Orders

**Responsibility:** Owns order lifecycle management

**Key domain objects:**
- ` + "`Order`" + ` (Aggregate) — the main order entity
- ` + "`OrderItem`" + ` (Value Object) — a line item
- ` + "`OrderPlaced`" + ` (Domain Event) — emitted when order is placed

**External dependencies:** None

### Context: Shipping

**Responsibility:** Owns shipment tracking and delivery

**Key domain objects:**
- ` + "`Shipment`" + ` (Aggregate) — tracks a shipment
- ` + "`ShippingLabel`" + ` (Value Object) — label for a package

**External dependencies:** Receives input from Orders

### Context: Payments

**Responsibility:** Owns payment processing

**Key domain objects:**
- ` + "`Payment`" + ` (Aggregate) — a payment transaction

**External dependencies:** Stripe API

### Context Map (Relationships)

| Upstream Context | Downstream Context | Integration Pattern |
|-----------------|-------------------|---------------------|
| Orders | Shipping | Domain Events (OrderPlaced) |
| Orders | Payments | Domain Events (OrderPlaced) |
`

func TestMarkdownDocParser_Import_WhenValidDDD_ExtractsBoundedContexts(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "DDD.md"), []byte(sampleDDD), 0o644))

	parser := infrastructure.NewMarkdownDocParser()
	result, err := parser.Import(context.Background(), dir)
	require.NoError(t, err)
	assert.Empty(t, result.Warnings())

	model := result.Model()
	contexts := model.BoundedContexts()
	assert.Len(t, contexts, 3)

	// Check names
	names := make([]string, len(contexts))
	for i, c := range contexts {
		names[i] = c.Name()
	}
	assert.Contains(t, names, "Orders")
	assert.Contains(t, names, "Shipping")
	assert.Contains(t, names, "Payments")
}

func TestMarkdownDocParser_Import_WhenValidDDD_ExtractsClassifications(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "DDD.md"), []byte(sampleDDD), 0o644))

	parser := infrastructure.NewMarkdownDocParser()
	result, err := parser.Import(context.Background(), dir)
	require.NoError(t, err)

	contexts := result.Model().BoundedContexts()
	classMap := make(map[string]*vo.SubdomainClassification)
	for _, c := range contexts {
		classMap[c.Name()] = c.Classification()
	}

	require.NotNil(t, classMap["Orders"])
	assert.Equal(t, vo.SubdomainCore, *classMap["Orders"])
	require.NotNil(t, classMap["Shipping"])
	assert.Equal(t, vo.SubdomainSupporting, *classMap["Shipping"])
	require.NotNil(t, classMap["Payments"])
	assert.Equal(t, vo.SubdomainGeneric, *classMap["Payments"])
}

func TestMarkdownDocParser_Import_WhenValidDDD_ExtractsResponsibilities(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "DDD.md"), []byte(sampleDDD), 0o644))

	parser := infrastructure.NewMarkdownDocParser()
	result, err := parser.Import(context.Background(), dir)
	require.NoError(t, err)

	contexts := result.Model().BoundedContexts()
	respMap := make(map[string]string)
	for _, c := range contexts {
		respMap[c.Name()] = c.Responsibility()
	}

	assert.Equal(t, "Owns order lifecycle management", respMap["Orders"])
	assert.Equal(t, "Owns shipment tracking and delivery", respMap["Shipping"])
	assert.Equal(t, "Owns payment processing", respMap["Payments"])
}

func TestMarkdownDocParser_Import_WhenValidDDD_ExtractsContextRelationships(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "DDD.md"), []byte(sampleDDD), 0o644))

	parser := infrastructure.NewMarkdownDocParser()
	result, err := parser.Import(context.Background(), dir)
	require.NoError(t, err)

	rels := result.Model().ContextRelationships()
	assert.Len(t, rels, 2)
	assert.Equal(t, "Orders", rels[0].Upstream())
	assert.Equal(t, "Shipping", rels[0].Downstream())
	assert.Equal(t, "Domain Events (OrderPlaced)", rels[0].IntegrationPattern())
}

func TestMarkdownDocParser_Import_WhenNoDDDFile_ReturnsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	parser := infrastructure.NewMarkdownDocParser()
	_, err := parser.Import(context.Background(), dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DDD.md")
}

func TestMarkdownDocParser_Import_WhenEmptyDDDFile_ReturnsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "DDD.md"), []byte(""), 0o644))

	parser := infrastructure.NewMarkdownDocParser()
	_, err := parser.Import(context.Background(), dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no bounded contexts")
}

func TestMarkdownDocParser_Import_WhenNoContextHeadings_ReturnsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	content := "# Some Doc\n\n## Section\n\nSome text\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "DDD.md"), []byte(content), 0o644))

	parser := infrastructure.NewMarkdownDocParser()
	_, err := parser.Import(context.Background(), dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no bounded contexts")
}

func TestMarkdownDocParser_Import_WhenNumberedContextHeadings_ExtractsContexts(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	content := `# DDD

## 4. Bounded Contexts

### 1. Orders (Core)

**Responsibility:** Manages orders

### 2. Shipping (Supporting)

**Responsibility:** Handles shipping
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "DDD.md"), []byte(content), 0o644))

	parser := infrastructure.NewMarkdownDocParser()
	result, err := parser.Import(context.Background(), dir)
	require.NoError(t, err)

	contexts := result.Model().BoundedContexts()
	assert.Len(t, contexts, 2)
	assert.Equal(t, "Orders", contexts[0].Name())
	assert.Equal(t, "Shipping", contexts[1].Name())

	require.NotNil(t, contexts[0].Classification())
	assert.Equal(t, vo.SubdomainCore, *contexts[0].Classification())
	require.NotNil(t, contexts[1].Classification())
	assert.Equal(t, vo.SubdomainSupporting, *contexts[1].Classification())
}

// Compile-time check that adapter satisfies port.
var _ = infrastructure.NewMarkdownDocParser
