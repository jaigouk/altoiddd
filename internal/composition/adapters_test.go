package composition

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	fitnessinfra "github.com/alto-cli/alto/internal/fitness/infrastructure"
)

// ---------------------------------------------------------------------------
// portScannerBridge — multi-directory scanning (1wu.2)
// ---------------------------------------------------------------------------

func TestPortScannerBridge_ScanPorts_WhenProjectRootHasMultipleContexts_MergesResults(t *testing.T) {
	t.Parallel()

	// Given: a temp project root with two contexts, each with an application/ dir
	projectRoot := t.TempDir()
	ordersAppDir := filepath.Join(projectRoot, "internal", "orders", "application")
	shippingAppDir := filepath.Join(projectRoot, "internal", "shipping", "application")
	require.NoError(t, os.MkdirAll(ordersAppDir, 0o755))
	require.NoError(t, os.MkdirAll(shippingAppDir, 0o755))

	ordersPort := `package application

// OrderRepository manages order persistence.
type OrderRepository interface {
	Save(ctx string, order string) error
	FindByID(ctx string, id string) (string, error)
}
`
	shippingPort := `package application

// ShipmentTracker tracks shipments.
type ShipmentTracker interface {
	Track(shipmentID string) (string, error)
}
`
	require.NoError(t, os.WriteFile(filepath.Join(ordersAppDir, "ports.go"), []byte(ordersPort), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(shippingAppDir, "ports.go"), []byte(shippingPort), 0o644))

	// When: bridge scans the project root
	bridge := &portScannerBridge{scanner: fitnessinfra.CodebasePortScanner{}}
	result := bridge.ScanPorts(projectRoot)

	// Then: both contexts' interfaces are found
	assert.Contains(t, result, "OrderRepository")
	assert.Contains(t, result, "ShipmentTracker")
	assert.Len(t, result, 2)

	// Verify methods are mapped correctly
	orderRepo := result["OrderRepository"]
	assert.Len(t, orderRepo.Methods(), 2)

	shipTracker := result["ShipmentTracker"]
	assert.Len(t, shipTracker.Methods(), 1)
}

func TestPortScannerBridge_ScanPorts_WhenDuplicateInterfaceNames_KeepsFirst(t *testing.T) {
	t.Parallel()

	// Given: two contexts that both define an interface named "Repository"
	projectRoot := t.TempDir()
	alphaAppDir := filepath.Join(projectRoot, "internal", "alpha", "application")
	betaAppDir := filepath.Join(projectRoot, "internal", "beta", "application")
	require.NoError(t, os.MkdirAll(alphaAppDir, 0o755))
	require.NoError(t, os.MkdirAll(betaAppDir, 0o755))

	alphaPort := `package application

type Repository interface {
	Save(id string) error
}
`
	betaPort := `package application

type Repository interface {
	Save(id string) error
	Delete(id string) error
}
`
	require.NoError(t, os.WriteFile(filepath.Join(alphaAppDir, "ports.go"), []byte(alphaPort), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(betaAppDir, "ports.go"), []byte(betaPort), 0o644))

	// When: bridge scans the project root
	bridge := &portScannerBridge{scanner: fitnessinfra.CodebasePortScanner{}}
	result := bridge.ScanPorts(projectRoot)

	// Then: only one Repository entry exists (first-wins, not overwritten)
	assert.Contains(t, result, "Repository")
	assert.Len(t, result, 1)

	// The first-wins entry has 1 method (alpha's), not 2 (beta's)
	repo := result["Repository"]
	assert.Len(t, repo.Methods(), 1)
}

func TestPortScannerBridge_ScanPorts_WhenEmptyProjectRoot_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	bridge := &portScannerBridge{scanner: fitnessinfra.CodebasePortScanner{}}
	result := bridge.ScanPorts("")

	assert.Empty(t, result)
}

func TestPortScannerBridge_ScanPorts_WhenNoApplicationDirs_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(projectRoot, "internal", "orders", "domain"), 0o755))

	bridge := &portScannerBridge{scanner: fitnessinfra.CodebasePortScanner{}}
	result := bridge.ScanPorts(projectRoot)

	assert.Empty(t, result)
}
