package valueobjects

// QualityGate enumerates the quality gates that can be executed against a project.
type QualityGate string

// Quality gate constants.
const (
	QualityGateLint    QualityGate = "lint"
	QualityGateTypes   QualityGate = "types"
	QualityGateTests   QualityGate = "tests"
	QualityGateFitness QualityGate = "fitness"
)

// AllQualityGates returns all quality gate values.
func AllQualityGates() []QualityGate {
	return []QualityGate{QualityGateLint, QualityGateTypes, QualityGateTests, QualityGateFitness}
}
