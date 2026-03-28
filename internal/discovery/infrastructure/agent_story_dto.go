package infrastructure

import (
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
)

// StoryOutput is the JSONL envelope data for a story.
type StoryOutput struct {
	SessionID   string             `json:"session_id"`
	StoryIndex  int                `json:"story_index"`
	Title       string             `json:"title"`
	StoryType   string             `json:"story_type"`
	TimeType    string             `json:"time_type"`
	PurityType  string             `json:"purity_type"`
	Trigger     string             `json:"trigger"`
	Actors      []ActorOutput      `json:"actors"`
	WorkObjects []WorkObjectOutput `json:"work_objects"`
	Sentences   []SentenceOutput   `json:"sentences"`
	Annotations []AnnotationOutput `json:"annotations,omitempty"`
	Variations  []string           `json:"variations,omitempty"`
}

// ActorOutput is the JSON representation of an actor in a story.
type ActorOutput struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Trust  string `json:"trust"`
	Source string `json:"source,omitempty"`
}

// WorkObjectOutput is the JSON representation of a work object in a story.
type WorkObjectOutput struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Trust  string `json:"trust"`
	Source string `json:"source,omitempty"`
}

// SentenceOutput is the JSON representation of a sentence in a story.
type SentenceOutput struct {
	Step           int    `json:"step"`
	Subject        string `json:"subject"`
	Activity       string `json:"activity"`
	Object         string `json:"object"`
	Preposition    string `json:"preposition,omitempty"`
	IndirectObject string `json:"indirect_object,omitempty"`
	Trust          string `json:"trust"`
	Source         string `json:"source,omitempty"`
}

// AnnotationOutput is the JSON representation of an annotation in a story.
type AnnotationOutput struct {
	Text        string `json:"text"`
	Type        string `json:"type"`
	SentenceRef *int   `json:"sentence_ref,omitempty"`
	Trust       string `json:"trust"`
	Source      string `json:"source,omitempty"`
}

// BoundaryProposalOutput is the JSON representation of a boundary proposal.
type BoundaryProposalOutput struct {
	Name           string   `json:"name"`
	Classification string   `json:"classification"`
	Confidence     float64  `json:"confidence"`
	Actors         []string `json:"actors"`
	WorkObjects    []string `json:"work_objects"`
	Stories        []string `json:"stories"`
}

// DiscoveryCompleteOutput is the JSON representation of a discovery completion event.
type DiscoveryCompleteOutput struct {
	SessionID   string `json:"session_id"`
	StoryCount  int    `json:"story_count"`
	SketchCount int    `json:"sketch_count"`
	Mode        string `json:"mode"`
}

// NewStoryOutput maps a DomainStory to a StoryOutput DTO.
func NewStoryOutput(sessionID string, storyIndex int, story *discoverydomain.DomainStory) StoryOutput {
	actors := story.Actors()
	actorOutputs := make([]ActorOutput, len(actors))

	for i, a := range actors {
		actorOutputs[i] = ActorOutput{
			Name:   a.Name(),
			Type:   a.Type().String(),
			Trust:  a.Trust().String(),
			Source: a.Source(),
		}
	}

	workObjects := story.WorkObjects()
	woOutputs := make([]WorkObjectOutput, len(workObjects))

	for i, wo := range workObjects {
		woOutputs[i] = WorkObjectOutput{
			Name:   wo.Name(),
			Type:   wo.Type().String(),
			Trust:  wo.Trust().String(),
			Source: wo.Source(),
		}
	}

	sentences := story.Sentences()
	sentenceOutputs := make([]SentenceOutput, len(sentences))

	for i, s := range sentences {
		sentenceOutputs[i] = SentenceOutput{
			Step:           s.Step(),
			Subject:        s.Subject(),
			Activity:       s.Activity(),
			Object:         s.Object(),
			Preposition:    s.Preposition(),
			IndirectObject: s.IndirectObject(),
			Trust:          s.Trust().String(),
			Source:         s.Source(),
		}
	}

	annotations := story.Annotations()
	var annotationOutputs []AnnotationOutput

	if len(annotations) > 0 {
		annotationOutputs = make([]AnnotationOutput, len(annotations))

		for i, a := range annotations {
			annotationOutputs[i] = AnnotationOutput{
				Text:        a.Text(),
				Type:        a.Type().String(),
				SentenceRef: a.SentenceRef(),
				Trust:       a.Trust().String(),
				Source:      a.Source(),
			}
		}
	}

	variations := story.Variations()
	var variationOutputs []string

	if len(variations) > 0 {
		variationOutputs = make([]string, len(variations))
		copy(variationOutputs, variations)
	}

	return StoryOutput{
		SessionID:   sessionID,
		StoryIndex:  storyIndex,
		Title:       story.Title(),
		StoryType:   story.Type().String(),
		TimeType:    story.Time().String(),
		PurityType:  story.Purity().String(),
		Trigger:     story.Trigger(),
		Actors:      actorOutputs,
		WorkObjects: woOutputs,
		Sentences:   sentenceOutputs,
		Annotations: annotationOutputs,
		Variations:  variationOutputs,
	}
}
