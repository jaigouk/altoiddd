package domain

// Scoring constants from spike Section 10.
// Base confidence per signal type.
const (
	BaseConfidenceSameObjectDiffContext float64 = 0.40
	BaseConfidenceOneWayFlow            float64 = 0.25
	BaseConfidenceOrgBoundary           float64 = 0.20
	BaseConfidenceDifferentTrigger      float64 = 0.15
	BaseConfidenceLanguageDifference    float64 = 0.15
	BaseConfidenceWorkObjectCluster     float64 = 0.15

	TypeBonusCoefficient  float64 = 0.15
	StoryBonusCoefficient float64 = 0.10
	StoryBonusDivisor     float64 = 3.0

	// False positive mitigation (spike Section 11).
	NotificationVerbDiscount     float64 = 0.30
	SequentialFlowDiscount       float64 = 0.50
	OrgBoundaryStoryCeilingCount int     = 3
	OrgBoundaryStoryCeilingScore float64 = 0.40
)

// ComputeBoundaryScore calculates the boundary confidence score using the
// spike-validated formula: signalAvg + 0.15*distinctTypeCount + 0.10*(storyCount/3.0).
func ComputeBoundaryScore(signalAvg float64, distinctTypeCount, storyCount int) float64 {
	if distinctTypeCount == 0 && storyCount == 0 {
		return 0.0
	}

	return signalAvg +
		TypeBonusCoefficient*float64(distinctTypeCount) +
		StoryBonusCoefficient*(float64(storyCount)/StoryBonusDivisor)
}
