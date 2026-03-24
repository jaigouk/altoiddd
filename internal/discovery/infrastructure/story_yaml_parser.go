package infrastructure

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/alto-cli/alto/internal/discovery/application"
	"github.com/alto-cli/alto/internal/discovery/domain"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// StoryYAMLParser reads and writes domain story YAML files.
type StoryYAMLParser struct{}

// Compile-time interface compliance checks.
var (
	_ application.StoryReader = (*StoryYAMLParser)(nil)
	_ application.StoryWriter = (*StoryYAMLParser)(nil)
)

// storyYAML is the top-level YAML structure for a domain story file.
type storyYAML struct {
	Version     int              `yaml:"version,omitempty"`
	Title       string           `yaml:"title"`
	Type        string           `yaml:"type"`
	Time        string           `yaml:"time"`
	Purity      string           `yaml:"purity"`
	Trigger     string           `yaml:"trigger"`
	Actors      []actorYAML      `yaml:"actors"`
	WorkObjects []workObjectYAML `yaml:"work_objects"`
	Sentences   []sentenceYAML   `yaml:"sentences"`
	Annotations []annotationYAML `yaml:"annotations,omitempty"`
	Variations  []string         `yaml:"variations,omitempty"`
}

// actorYAML is the YAML structure for a single actor.
type actorYAML struct {
	Name   string `yaml:"name"`
	Type   string `yaml:"type"`
	Trust  string `yaml:"trust"`
	Source string `yaml:"source,omitempty"`
}

// workObjectYAML is the YAML structure for a single work object.
type workObjectYAML struct {
	Name   string `yaml:"name"`
	Type   string `yaml:"type"`
	Trust  string `yaml:"trust"`
	Source string `yaml:"source,omitempty"`
}

// sentenceYAML is the YAML structure for a single story sentence.
type sentenceYAML struct {
	Step           int    `yaml:"step"`
	Subject        string `yaml:"subject"`
	Activity       string `yaml:"activity"`
	Object         string `yaml:"object"`
	Preposition    string `yaml:"preposition,omitempty"`
	IndirectObject string `yaml:"indirect_object,omitempty"`
	Trust          string `yaml:"trust"`
	Source         string `yaml:"source,omitempty"`
}

// annotationYAML is the YAML structure for a single annotation.
type annotationYAML struct {
	Text     string `yaml:"text"`
	Type     string `yaml:"type"`
	Sentence *int   `yaml:"sentence,omitempty"`
	Trust    string `yaml:"trust,omitempty"`
	Source   string `yaml:"source,omitempty"`
}

// Read reads a DomainStory from a YAML file at path.
func (p *StoryYAMLParser) Read(ctx context.Context, path string) (*domain.DomainStory, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("reading story: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading story file %q: %w", path, err)
	}

	return p.Parse(data)
}

// Write writes a DomainStory to a YAML file at path.
func (p *StoryYAMLParser) Write(ctx context.Context, path string, story *domain.DomainStory) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("writing story: %w", err)
	}

	data, err := p.Serialize(story)
	if err != nil {
		return fmt.Errorf("serializing story: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing story file %q: %w", path, err)
	}

	return nil
}

// Parse parses YAML bytes into a DomainStory aggregate.
func (p *StoryYAMLParser) Parse(data []byte) (*domain.DomainStory, error) {
	var doc storyYAML
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing story YAML: %w", err)
	}

	if err := validateRequiredFields(doc); err != nil {
		return nil, err
	}

	storyType, err := domain.NewStoryType(doc.Type)
	if err != nil {
		return nil, fmt.Errorf("type: %w", err)
	}

	timeType, err := domain.NewTimeType(doc.Time)
	if err != nil {
		return nil, fmt.Errorf("time: %w", err)
	}

	purityType, err := domain.NewPurityType(doc.Purity)
	if err != nil {
		return nil, fmt.Errorf("purity: %w", err)
	}

	story, err := domain.NewDomainStory(doc.Title, storyType, timeType, purityType, doc.Trigger)
	if err != nil {
		return nil, fmt.Errorf("creating domain story: %w", err)
	}

	if err := parseActors(story, doc.Actors); err != nil {
		return nil, err
	}

	if err := parseWorkObjects(story, doc.WorkObjects); err != nil {
		return nil, err
	}

	if err := parseSentences(story, doc.Sentences); err != nil {
		return nil, err
	}

	if err := parseAnnotations(story, doc.Annotations); err != nil {
		return nil, err
	}

	for i, v := range doc.Variations {
		if err := story.AddVariation(v); err != nil {
			return nil, fmt.Errorf("variations[%d]: %w", i, err)
		}
	}

	if err := story.Validate(); err != nil {
		return nil, fmt.Errorf("validating story: %w", err)
	}

	return story, nil
}

// Serialize converts a DomainStory to YAML bytes.
func (p *StoryYAMLParser) Serialize(story *domain.DomainStory) ([]byte, error) {
	doc := storyYAML{
		Version: 1,
		Title:   story.Title(),
		Type:    story.Type().String(),
		Time:    story.Time().String(),
		Purity:  story.Purity().String(),
		Trigger: story.Trigger(),
	}

	actors := story.Actors()
	doc.Actors = make([]actorYAML, len(actors))

	for i, a := range actors {
		doc.Actors[i] = actorYAML{
			Name:   a.Name(),
			Type:   a.Type().String(),
			Trust:  a.Trust().String(),
			Source: a.Source(),
		}
	}

	workObjects := story.WorkObjects()
	doc.WorkObjects = make([]workObjectYAML, len(workObjects))

	for i, wo := range workObjects {
		doc.WorkObjects[i] = workObjectYAML{
			Name:   wo.Name(),
			Type:   wo.Type().String(),
			Trust:  wo.Trust().String(),
			Source: wo.Source(),
		}
	}

	sentences := story.Sentences()
	doc.Sentences = make([]sentenceYAML, len(sentences))

	for i, s := range sentences {
		doc.Sentences[i] = sentenceYAML{
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
	if len(annotations) > 0 {
		doc.Annotations = make([]annotationYAML, len(annotations))

		for i, a := range annotations {
			doc.Annotations[i] = annotationYAML{
				Text:     a.Text(),
				Type:     a.Type().String(),
				Sentence: a.SentenceRef(),
				Trust:    a.Trust().String(),
				Source:   a.Source(),
			}
		}
	}

	doc.Variations = story.Variations()

	data, err := yaml.Marshal(&doc)
	if err != nil {
		return nil, fmt.Errorf("serializing story YAML: %w", err)
	}

	return data, nil
}

func validateRequiredFields(doc storyYAML) error {
	if doc.Title == "" {
		return fmt.Errorf("title: must not be empty")
	}

	if doc.Type == "" {
		return fmt.Errorf("type: must not be empty")
	}

	if doc.Time == "" {
		return fmt.Errorf("time: must not be empty")
	}

	if doc.Purity == "" {
		return fmt.Errorf("purity: must not be empty")
	}

	if doc.Trigger == "" {
		return fmt.Errorf("trigger: must not be empty")
	}

	if len(doc.Actors) == 0 {
		return fmt.Errorf("actors: must have at least one actor")
	}

	if len(doc.WorkObjects) == 0 {
		return fmt.Errorf("work_objects: must have at least one work object")
	}

	if len(doc.Sentences) == 0 {
		return fmt.Errorf("sentences: must have at least one sentence")
	}

	return nil
}

func parseActors(story *domain.DomainStory, actors []actorYAML) error {
	for i, a := range actors {
		actorType, err := domain.NewActorType(a.Type)
		if err != nil {
			return fmt.Errorf("actors[%d].type: %w", i, err)
		}

		trust, err := vo.ParseTrustLevel(a.Trust)
		if err != nil {
			return fmt.Errorf("actors[%d].trust: %w", i, err)
		}

		actor, err := domain.NewStoryActor(a.Name, actorType, trust, a.Source)
		if err != nil {
			return fmt.Errorf("actors[%d]: %w", i, err)
		}

		if err := story.AddActor(actor); err != nil {
			return fmt.Errorf("actors[%d]: %w", i, err)
		}
	}

	return nil
}

func parseWorkObjects(story *domain.DomainStory, workObjects []workObjectYAML) error {
	for i, wo := range workObjects {
		woType, err := domain.NewWorkObjectType(wo.Type)
		if err != nil {
			return fmt.Errorf("work_objects[%d].type: %w", i, err)
		}

		trust, err := vo.ParseTrustLevel(wo.Trust)
		if err != nil {
			return fmt.Errorf("work_objects[%d].trust: %w", i, err)
		}

		workObj, err := domain.NewWorkObject(wo.Name, woType, trust, wo.Source)
		if err != nil {
			return fmt.Errorf("work_objects[%d]: %w", i, err)
		}

		if err := story.AddWorkObject(workObj); err != nil {
			return fmt.Errorf("work_objects[%d]: %w", i, err)
		}
	}

	return nil
}

func parseSentences(story *domain.DomainStory, sentences []sentenceYAML) error {
	for i, s := range sentences {
		trust, err := vo.ParseTrustLevel(s.Trust)
		if err != nil {
			return fmt.Errorf("sentences[%d].trust: %w", i, err)
		}

		sentence, err := domain.NewStorySentence(s.Step, s.Subject, s.Activity, s.Object, trust, s.Source)
		if err != nil {
			return fmt.Errorf("sentences[%d]: %w", i, err)
		}

		if s.Preposition != "" {
			sentence, err = sentence.WithPreposition(s.Preposition, s.IndirectObject)
			if err != nil {
				return fmt.Errorf("sentences[%d]: %w", i, err)
			}
		}

		if err := story.AddSentence(sentence); err != nil {
			return fmt.Errorf("sentences[%d]: %w", i, err)
		}
	}

	return nil
}

func parseAnnotations(story *domain.DomainStory, annotations []annotationYAML) error {
	for i, a := range annotations {
		annType, err := domain.NewAnnotationType(a.Type)
		if err != nil {
			return fmt.Errorf("annotations[%d].type: %w", i, err)
		}

		// Default trust to user_stated when omitted in YAML.
		trust := vo.UserStated
		if a.Trust != "" {
			trust, err = vo.ParseTrustLevel(a.Trust)
			if err != nil {
				return fmt.Errorf("annotations[%d].trust: %w", i, err)
			}
		}

		ann, err := domain.NewAnnotation(a.Text, annType, a.Sentence, trust, a.Source)
		if err != nil {
			return fmt.Errorf("annotations[%d]: %w", i, err)
		}

		if err := story.AddAnnotation(ann); err != nil {
			return fmt.Errorf("annotations[%d]: %w", i, err)
		}
	}

	return nil
}
