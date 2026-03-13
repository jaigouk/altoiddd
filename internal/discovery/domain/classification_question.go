package domain

import (
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// ClassificationResult is the result of classifying a bounded context.
type ClassificationResult struct {
	classification vo.SubdomainClassification
	rationale      string
}

// NewClassificationResult creates a ClassificationResult value object.
func NewClassificationResult(classification vo.SubdomainClassification, rationale string) ClassificationResult {
	return ClassificationResult{
		classification: classification,
		rationale:      rationale,
	}
}

// Classification returns the subdomain classification.
func (r ClassificationResult) Classification() vo.SubdomainClassification {
	return r.classification
}

// Rationale returns the human-readable rationale for the classification.
func (r ClassificationResult) Rationale() string {
	return r.rationale
}

// Equal returns true if two ClassificationResults have the same values.
func (r ClassificationResult) Equal(other ClassificationResult) bool {
	return r.classification == other.classification && r.rationale == other.rationale
}

// ClassificationDecisionTree implements Khononov's decision tree for subdomain classification.
// Questions:
//  1. BUY: "Could you use an off-the-shelf solution without losing competitive advantage?"
//  2. COMPLEXITY: "Is it mostly CRUD or complex rules?"
//  3. COMPETITOR: "If a competitor copied this exactly, would it threaten your business?"
type ClassificationDecisionTree struct{}

// NewClassificationDecisionTree creates a new decision tree.
func NewClassificationDecisionTree() *ClassificationDecisionTree {
	return &ClassificationDecisionTree{}
}

// Classify determines the subdomain classification based on the Khononov decision tree.
// buyYes: true if user answered "yes" to the BUY question (could use off-the-shelf)
// complexRules: true if the domain has complex rules (not simple CRUD)
// competitorThreat: true if a competitor copying this would be a threat.
func (t *ClassificationDecisionTree) Classify(buyYes, complexRules, competitorThreat bool) ClassificationResult {
	// BUY question: Can you use off-the-shelf?
	if buyYes {
		return NewClassificationResult(
			vo.SubdomainGeneric,
			"Can use off-the-shelf solution without losing competitive advantage",
		)
	}

	// COMPLEXITY question: Simple CRUD or complex rules?
	if !complexRules {
		return NewClassificationResult(
			vo.SubdomainSupporting,
			"Necessary for business but mostly simple data operations",
		)
	}

	// COMPETITOR question: Would copying threaten business?
	if competitorThreat {
		return NewClassificationResult(
			vo.SubdomainCore,
			"Complex rules that provide competitive advantage",
		)
	}

	// Complex but not differentiating
	return NewClassificationResult(
		vo.SubdomainSupporting,
		"Has complex but not differentiating rules",
	)
}

// ClassificationQuestion represents a question for classifying a bounded context.
type ClassificationQuestion struct {
	contextName      string
	questionType     ClassificationQuestionType
	technicalText    string
	nonTechnicalText string
}

// ClassificationQuestionType indicates which step in the decision tree.
type ClassificationQuestionType string

// Classification question type constants.
const (
	QuestionTypeBuy        ClassificationQuestionType = "buy"
	QuestionTypeComplexity ClassificationQuestionType = "complexity"
	QuestionTypeCompetitor ClassificationQuestionType = "competitor"
)

// NewClassificationQuestion creates a classification question for a bounded context.
func NewClassificationQuestion(contextName string, qType ClassificationQuestionType) ClassificationQuestion {
	var technical, nonTechnical string

	switch qType {
	case QuestionTypeBuy:
		technical = "Could you use an existing product for " + contextName + " without losing competitive advantage?"
		nonTechnical = "For " + contextName + ", could you use an off-the-shelf solution without losing what makes your business special?"
	case QuestionTypeComplexity:
		technical = "Is " + contextName + " mostly CRUD operations or does it have complex business rules?"
		nonTechnical = "When your team talks about " + contextName + ", is it mostly storing/retrieving data, or are there complex rules and special cases?"
	case QuestionTypeCompetitor:
		technical = "If a competitor implemented " + contextName + " exactly as you have, would that threaten your business?"
		nonTechnical = "If a competitor copied " + contextName + " exactly, would that threaten your business?"
	}

	return ClassificationQuestion{
		contextName:      contextName,
		questionType:     qType,
		technicalText:    technical,
		nonTechnicalText: nonTechnical,
	}
}

// ContextName returns the bounded context name this question is about.
func (q ClassificationQuestion) ContextName() string { return q.contextName }

// QuestionType returns the type of classification question.
func (q ClassificationQuestion) QuestionType() ClassificationQuestionType { return q.questionType }

// TechnicalText returns the DDD/engineering phrasing.
func (q ClassificationQuestion) TechnicalText() string { return q.technicalText }

// NonTechnicalText returns the plain-language phrasing.
func (q ClassificationQuestion) NonTechnicalText() string { return q.nonTechnicalText }

// Text returns the appropriate text based on register.
func (q ClassificationQuestion) Text(register DiscoveryRegister) string {
	if register == RegisterTechnical {
		return q.technicalText
	}
	return q.nonTechnicalText
}
