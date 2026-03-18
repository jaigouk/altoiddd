package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alto-cli/alto/internal/fitness/domain"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// Role enum
// ---------------------------------------------------------------------------

func TestRoleValues(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "execution", string(domain.RoleExecution))
	assert.Equal(t, "analysis", string(domain.RoleAnalysis))
	assert.Equal(t, "gateway", string(domain.RoleGateway))
	assert.Equal(t, "specification", string(domain.RoleSpecification))
	assert.Equal(t, "draft", string(domain.RoleDraft))
}

// ---------------------------------------------------------------------------
// StrategicClassification
// ---------------------------------------------------------------------------

func TestStrategicClassification(t *testing.T) {
	t.Parallel()
	sc := domain.NewStrategicClassification(vo.SubdomainCore, "Revenue", "Custom")
	assert.Equal(t, vo.SubdomainCore, sc.Domain())
	assert.Equal(t, "Revenue", sc.BusinessModel())
	assert.Equal(t, "Custom", sc.Evolution())
}

// ---------------------------------------------------------------------------
// CommunicationMessage
// ---------------------------------------------------------------------------

func TestCommunicationMessage(t *testing.T) {
	t.Parallel()
	msg := domain.NewCommunicationMessage("PlaceOrder", "Command", "API Gateway")
	assert.Equal(t, "PlaceOrder", msg.Message())
	assert.Equal(t, "Command", msg.MessageType())
	assert.Equal(t, "API Gateway", msg.Counterpart())
}

// ---------------------------------------------------------------------------
// BoundedContextCanvas
// ---------------------------------------------------------------------------

func TestBoundedContextCanvas(t *testing.T) {
	t.Parallel()

	t.Run("stores fields", func(t *testing.T) {
		t.Parallel()
		sc := domain.NewStrategicClassification(vo.SubdomainCore, "Revenue", "Custom")
		canvas := domain.NewBoundedContextCanvas(
			"Sales", "Manages orders", sc,
			[]domain.Role{domain.RoleExecution},
			[]domain.CommunicationMessage{domain.NewCommunicationMessage("PlaceOrder", "Command", "Gateway")},
			[]domain.CommunicationMessage{domain.NewCommunicationMessage("OrderPlaced", "Event", "Fulfillment")},
			[][2]string{{"Order", "A purchase request"}},
			[]string{"Order must have items"},
			nil, nil,
		)
		assert.Equal(t, "Sales", canvas.ContextName())
		assert.Equal(t, "Manages orders", canvas.Purpose())
		assert.Equal(t, vo.SubdomainCore, canvas.Classification().Domain())
		assert.Len(t, canvas.Roles(), 1)
		assert.Len(t, canvas.InboundCommunication(), 1)
		assert.Len(t, canvas.OutboundCommunication(), 1)
		assert.Len(t, canvas.UbiquitousLanguage(), 1)
		assert.Len(t, canvas.BusinessDecisions(), 1)
		assert.Empty(t, canvas.Assumptions())
		assert.Empty(t, canvas.OpenQuestions())
	})

	t.Run("defensive copy domain roles", func(t *testing.T) {
		t.Parallel()
		roles := []domain.Role{domain.RoleExecution}
		sc := domain.NewStrategicClassification(vo.SubdomainCore, "R", "C")
		canvas := domain.NewBoundedContextCanvas("X", "Y", sc, roles, nil, nil, nil, nil, nil, nil)
		roles[0] = domain.RoleDraft
		assert.Equal(t, domain.RoleExecution, canvas.Roles()[0])
	})

	t.Run("empty communications", func(t *testing.T) {
		t.Parallel()
		sc := domain.NewStrategicClassification(vo.SubdomainGeneric, "unclassified", "Commodity")
		canvas := domain.NewBoundedContextCanvas("Logging", "Records events", sc,
			[]domain.Role{domain.RoleGateway}, nil, nil, nil, nil, nil, nil)
		assert.Empty(t, canvas.InboundCommunication())
		assert.Empty(t, canvas.OutboundCommunication())
	})

	t.Run("special chars in name", func(t *testing.T) {
		t.Parallel()
		sc := domain.NewStrategicClassification(vo.SubdomainSupporting, "Compliance", "Product")
		canvas := domain.NewBoundedContextCanvas(`Auth & Identity "Service"`, "Handles auth", sc,
			[]domain.Role{domain.RoleSpecification}, nil, nil, nil, nil, nil, nil)
		assert.Equal(t, `Auth & Identity "Service"`, canvas.ContextName())
	})

	t.Run("very long purpose", func(t *testing.T) {
		t.Parallel()
		longPurpose := strings.Repeat("A", 600)
		sc := domain.NewStrategicClassification(vo.SubdomainCore, "Revenue", "Genesis")
		canvas := domain.NewBoundedContextCanvas("Verbose", longPurpose, sc,
			[]domain.Role{domain.RoleExecution}, nil, nil, nil, nil, nil, nil)
		assert.Len(t, canvas.Purpose(), 600)
	})
}

func TestStrategicClassificationEquality(t *testing.T) {
	t.Parallel()
	a := domain.NewStrategicClassification(vo.SubdomainCore, "Revenue", "Genesis")
	b := domain.NewStrategicClassification(vo.SubdomainCore, "Revenue", "Genesis")
	assert.Equal(t, a, b)
}

func TestAllRolesReturnsAllConstants(t *testing.T) {
	t.Parallel()
	all := domain.AllRoles()
	assert.Len(t, all, 5)
	assert.Contains(t, all, domain.RoleExecution)
	assert.Contains(t, all, domain.RoleAnalysis)
	assert.Contains(t, all, domain.RoleGateway)
	assert.Contains(t, all, domain.RoleSpecification)
	assert.Contains(t, all, domain.RoleDraft)
}
