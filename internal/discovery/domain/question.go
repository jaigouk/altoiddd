package domain

// Question is a single discovery question with dual-register phrasing.
type Question struct {
	id               string
	phase            QuestionPhase
	technicalText    string
	nonTechnicalText string
	produces         []string
}

// NewQuestion creates a Question value object.
func NewQuestion(id string, phase QuestionPhase, technicalText, nonTechnicalText string, produces []string) Question {
	p := make([]string, len(produces))
	copy(p, produces)
	return Question{
		id:               id,
		phase:            phase,
		technicalText:    technicalText,
		nonTechnicalText: nonTechnicalText,
		produces:         p,
	}
}

// ID returns the question identifier.
func (q Question) ID() string { return q.id }

// Phase returns which discovery phase this question belongs to.
func (q Question) Phase() QuestionPhase { return q.phase }

// TechnicalText returns the DDD/engineering phrasing.
func (q Question) TechnicalText() string { return q.technicalText }

// NonTechnicalText returns the plain-language phrasing.
func (q Question) NonTechnicalText() string { return q.nonTechnicalText }

// Produces returns a defensive copy of the domain artifacts this question helps discover.
func (q Question) Produces() []string {
	out := make([]string, len(q.produces))
	copy(out, q.produces)
	return out
}

// questionCatalog is the 10-question DDD discovery catalog (singleton).
var questionCatalog = []Question{
	NewQuestion("Q1", PhaseActors,
		"Who are the actors (users, external systems) that interact with your system?",
		"Who will use this product, and what other systems does it talk to?",
		[]string{"actors", "external_systems"}),
	NewQuestion("Q2", PhaseActors,
		"What are the core entities (nouns) in your domain?",
		"What are the main things or concepts your product deals with?",
		[]string{"entities", "value_objects"}),
	NewQuestion("Q3", PhaseStory,
		"Describe the primary use case as a domain story: actor -> command -> event -> outcome.",
		"Walk me through the most important thing a user does, step by step.",
		[]string{"commands", "events", "domain_story"}),
	NewQuestion("Q4", PhaseStory,
		"What is the most critical failure mode? What invariants must hold?",
		"What could go wrong that would be a serious problem? What rules must never be broken?",
		[]string{"invariants", "failure_modes"}),
	NewQuestion("Q5", PhaseStory,
		"What other workflows or use cases exist beyond the primary one?",
		"What else can users do with the product besides the main thing?",
		[]string{"secondary_stories", "commands"}),
	NewQuestion("Q6", PhaseEvents,
		"What domain events are published when state changes occur?",
		"What important things happen in the system that other parts need to know about?",
		[]string{"domain_events"}),
	NewQuestion("Q7", PhaseEvents,
		"What policies (event -> command reactions) exist in the system?",
		"When something happens, what automatic actions should follow?",
		[]string{"policies", "reactions"}),
	NewQuestion("Q8", PhaseEvents,
		"What read models or projections does the system need?",
		"What views or reports do users need to see?",
		[]string{"read_models", "projections"}),
	NewQuestion("Q9", PhaseBoundaries,
		"How would you partition the domain into bounded contexts?",
		"If you split the product into independent teams, what would each team own?",
		[]string{"bounded_contexts"}),
	NewQuestion("Q10", PhaseBoundaries,
		"Classify each context: core (competitive advantage), supporting (necessary but not differentiating), or generic (commodity).",
		"Which parts are your secret sauce, which are necessary plumbing, and which are off-the-shelf?",
		[]string{"subdomain_classification"}),
}

// QuestionCatalog returns a copy of the 10-question catalog.
func QuestionCatalog() []Question {
	out := make([]Question, len(questionCatalog))
	copy(out, questionCatalog)
	return out
}

// QuestionByID returns a map of question ID to Question for fast lookup.
func QuestionByID() map[string]Question {
	m := make(map[string]Question, len(questionCatalog))
	for _, q := range questionCatalog {
		m[q.id] = q
	}
	return m
}

// mvpQuestionIDs are the minimum viable subset of questions.
var mvpQuestionIDs = map[string]bool{
	"Q1": true, "Q3": true, "Q4": true, "Q9": true, "Q10": true,
}

// MVPQuestionIDs returns a copy of the MVP question ID set.
func MVPQuestionIDs() map[string]bool {
	out := make(map[string]bool, len(mvpQuestionIDs))
	for k, v := range mvpQuestionIDs {
		out[k] = v
	}
	return out
}
