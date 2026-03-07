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

// GateResult captures the outcome of running a single quality gate.
type GateResult struct {
	gate       QualityGate
	output     string
	durationMS int
	passed     bool
}

// NewGateResult creates a GateResult value object.
func NewGateResult(gate QualityGate, passed bool, output string, durationMS int) GateResult {
	return GateResult{gate: gate, passed: passed, output: output, durationMS: durationMS}
}

// Gate returns the quality gate.
func (r GateResult) Gate() QualityGate { return r.gate }

// Passed returns whether the gate passed.
func (r GateResult) Passed() bool { return r.passed }

// Output returns the gate execution output.
func (r GateResult) Output() string { return r.output }

// DurationMS returns the gate execution duration in milliseconds.
func (r GateResult) DurationMS() int { return r.durationMS }

// QualityReport aggregates all gate results into a pass/fail verdict.
type QualityReport struct {
	results []GateResult
}

// NewQualityReport creates a QualityReport value object.
func NewQualityReport(results []GateResult) QualityReport {
	r := make([]GateResult, len(results))
	copy(r, results)
	return QualityReport{results: r}
}

// Results returns a defensive copy.
func (qr QualityReport) Results() []GateResult {
	out := make([]GateResult, len(qr.results))
	copy(out, qr.results)
	return out
}

// Passed returns true when every gate in the report passed.
func (qr QualityReport) Passed() bool {
	for _, r := range qr.results {
		if !r.passed {
			return false
		}
	}
	return true
}
