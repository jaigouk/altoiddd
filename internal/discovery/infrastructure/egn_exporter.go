package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	discoveryapp "github.com/alto-cli/alto/internal/discovery/application"
	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
)

// Compile-time interface check.
var _ discoveryapp.EgnRenderer = (*EgnExporter)(nil)

// EgnExporter renders a DomainStory as an Egon.io-compatible JSON string.
type EgnExporter struct{}

// egnDocument is the top-level Egon JSON structure.
type egnDocument struct {
	Domain egnDomain         `json:"domain"`
	DST    []json.RawMessage `json:"dst"`
}

type egnDomain struct {
	Name        string            `json:"name"`
	Actors      map[string]string `json:"actors"`
	WorkObjects map[string]string `json:"workObjects"`
}

type egnShape struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	ID          string         `json:"id"`
	PickedColor string         `json:"pickedColor"`
	X           int            `json:"x"`
	Y           int            `json:"y"`
	DType       string         `json:"$type"`
	DI          map[string]any `json:"di"`
	Descriptor  map[string]any `json:"$descriptor"`
}

type egnActivity struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	ID          string         `json:"id"`
	PickedColor string         `json:"pickedColor"`
	Number      *int           `json:"number"`
	Source      string         `json:"source"`
	Target      string         `json:"target"`
	Waypoints   []egnWaypoint  `json:"waypoints"`
	DType       string         `json:"$type"`
	DI          map[string]any `json:"di"`
	Descriptor  map[string]any `json:"$descriptor"`
}

type egnWaypoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type egnMetaInfo struct {
	Info string `json:"info"`
}

type egnMetaVersion struct {
	Version string `json:"version"`
}

// Actor type SVG icons.
var actorSVGs = map[discoverydomain.ActorType]string{
	discoverydomain.ActorTypePerson: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 26"><path d="M12 4C9.79 4 8 5.79 8 8s1.79 4 4 4 4-1.79 4-4-1.79-4-4-4zm0 9c-2.67 0-8 1.34-8 4v3h16v-3c0-2.66-5.33-4-8-4z"/></svg>`,
	discoverydomain.ActorTypeSystem: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 26"><path d="M20,18c1.1,0,2-0.9,2-2V6c0-1.1-0.9-2-2-2H4C2.9,4,2,4.9,2,6v10c0,1.1,0.9,2,2,2H0v2h24v-2H20z"/></svg>`,
	discoverydomain.ActorTypeGroup:  `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 26"><path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/></svg>`,
}

// Work object type SVG icons.
var workObjectSVGs = map[discoverydomain.WorkObjectType]string{
	discoverydomain.WorkObjectTypeDocument:     `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 26"><path d="M14 2H6c-1.1 0-2 .9-2 2v16c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V8l-6-6z"/></svg>`,
	discoverydomain.WorkObjectTypeFolder:       `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 26"><path d="M10 4H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"/></svg>`,
	discoverydomain.WorkObjectTypeCall:         `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 26"><path d="M20.01 15.38c-1.23 0-2.42-.2-3.53-.56-.35-.12-.74-.03-1.01.24l-1.57 1.97c-2.83-1.35-5.48-3.9-6.89-6.83l1.95-1.66c.27-.28.35-.67.24-1.02-.37-1.11-.56-2.3-.56-3.53 0-.54-.45-.99-.99-.99H4.19C3.65 3 3 3.24 3 3.99 3 13.28 10.73 21 20.01 21c.71 0 .99-.63.99-1.18v-3.45c0-.54-.45-.99-.99-.99z"/></svg>`,
	discoverydomain.WorkObjectTypeEmail:        `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 26"><path d="M20 4H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm0 4l-8 5-8-5V6l8 5 8-5v2z"/></svg>`,
	discoverydomain.WorkObjectTypeConversation: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 26"><path d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2z"/></svg>`,
	discoverydomain.WorkObjectTypeInfo:         `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 26"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z"/></svg>`,
	discoverydomain.WorkObjectTypeData:         `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 26"><ellipse cx="12" cy="5" rx="8" ry="3"/><path d="M4 5v6c0 1.66 3.58 3 8 3s8-1.34 8-3V5"/><path d="M4 11v6c0 1.66 3.58 3 8 3s8-1.34 8-3v-6"/></svg>`,
}

// Render converts a DomainStory into Egon.io JSON format.
func (e *EgnExporter) Render(_ context.Context, story *discoverydomain.DomainStory) (string, error) {
	if story == nil {
		return "", fmt.Errorf("story must not be nil")
	}

	actors := story.Actors()
	workObjects := story.WorkObjects()
	sentences := story.Sentences()

	// Build shape list and ID lookup map.
	var dst []json.RawMessage
	shapeIDMap := make(map[string]string) // name (lowercase) → shape ID
	shapeCounter := 0

	// Actors as shapes.
	for i, actor := range actors {
		shapeCounter++
		shapeID := fmt.Sprintf("shape_%04d", shapeCounter)
		shapeIDMap[strings.ToLower(actor.Name())] = shapeID

		actorType, err := egnActorType(actor.Type())
		if err != nil {
			return "", fmt.Errorf("mapping actor type: %w", err)
		}

		shape := egnShape{
			Type:        actorType,
			Name:        actor.Name(),
			ID:          shapeID,
			PickedColor: "",
			X:           100 + i*400,
			Y:           200,
			DType:       "Element",
			DI:          map[string]any{},
			Descriptor:  map[string]any{},
		}

		raw, err := json.Marshal(shape)
		if err != nil {
			return "", fmt.Errorf("marshaling actor shape: %w", err)
		}

		dst = append(dst, raw)
	}

	// Work objects as shapes.
	woStartX := len(actors)*400 + 200
	for i, wo := range workObjects {
		shapeCounter++
		shapeID := fmt.Sprintf("shape_%04d", shapeCounter)
		shapeIDMap[strings.ToLower(wo.Name())] = shapeID

		woType, err := egnWorkObjectType(wo.Type())
		if err != nil {
			return "", fmt.Errorf("mapping work object type: %w", err)
		}

		yPos := 50
		if i%2 == 1 {
			yPos = 350
		}

		shape := egnShape{
			Type:        woType,
			Name:        wo.Name(),
			ID:          shapeID,
			PickedColor: "",
			X:           woStartX + i*200,
			Y:           yPos,
			DType:       "Element",
			DI:          map[string]any{},
			Descriptor:  map[string]any{},
		}

		raw, err := json.Marshal(shape)
		if err != nil {
			return "", fmt.Errorf("marshaling work object shape: %w", err)
		}

		dst = append(dst, raw)
	}

	// Connections from sentences.
	for _, sentence := range sentences {
		sourceID := shapeIDMap[strings.ToLower(sentence.Subject())]
		targetID := shapeIDMap[strings.ToLower(sentence.Object())]

		connID := fmt.Sprintf("connection_%04d", sentence.Step())
		step := sentence.Step()

		activity := egnActivity{
			Type:        "domainStory:activity",
			Name:        sentence.Activity(),
			ID:          connID,
			PickedColor: "",
			Number:      &step,
			Source:      sourceID,
			Target:      targetID,
			Waypoints:   []egnWaypoint{{X: 0, Y: 0}, {X: 100, Y: 100}},
			DType:       "Element",
			DI:          map[string]any{},
			Descriptor:  map[string]any{},
		}

		raw, err := json.Marshal(activity)
		if err != nil {
			return "", fmt.Errorf("marshaling connection: %w", err)
		}

		dst = append(dst, raw)

		// Bridge connection for indirect objects.
		if sentence.HasIndirectObject() {
			bridgeTargetID := shapeIDMap[strings.ToLower(sentence.IndirectObject())]
			bridgeID := fmt.Sprintf("connection_%04db", sentence.Step())

			bridge := egnActivity{
				Type:        "domainStory:activity",
				Name:        sentence.Preposition(),
				ID:          bridgeID,
				PickedColor: "",
				Number:      nil,
				Source:      targetID,
				Target:      bridgeTargetID,
				Waypoints:   []egnWaypoint{{X: 0, Y: 0}, {X: 100, Y: 100}},
				DType:       "Element",
				DI:          map[string]any{},
				Descriptor:  map[string]any{},
			}

			raw, err := json.Marshal(bridge)
			if err != nil {
				return "", fmt.Errorf("marshaling bridge connection: %w", err)
			}

			dst = append(dst, raw)
		}
	}

	// Metadata entries.
	infoRaw, err := json.Marshal(egnMetaInfo{Info: story.Title()})
	if err != nil {
		return "", fmt.Errorf("marshaling info metadata: %w", err)
	}

	dst = append(dst, infoRaw)

	versionRaw, err := json.Marshal(egnMetaVersion{Version: "2.0.1"})
	if err != nil {
		return "", fmt.Errorf("marshaling version metadata: %w", err)
	}

	dst = append(dst, versionRaw)

	// Build domain block with only used types.
	usedActorTypes := make(map[discoverydomain.ActorType]struct{})
	for _, actor := range actors {
		usedActorTypes[actor.Type()] = struct{}{}
	}

	usedWOTypes := make(map[discoverydomain.WorkObjectType]struct{})
	for _, wo := range workObjects {
		usedWOTypes[wo.Type()] = struct{}{}
	}

	domainActors := make(map[string]string)
	for at := range usedActorTypes {
		key, keyErr := egnActorDomainKey(at)
		if keyErr != nil {
			return "", fmt.Errorf("mapping actor domain key: %w", keyErr)
		}

		domainActors[key] = actorSVGs[at]
	}

	domainWOs := make(map[string]string)
	for wot := range usedWOTypes {
		key, keyErr := egnWorkObjectDomainKey(wot)
		if keyErr != nil {
			return "", fmt.Errorf("mapping work object domain key: %w", keyErr)
		}

		domainWOs[key] = workObjectSVGs[wot]
	}

	doc := egnDocument{
		Domain: egnDomain{
			Name:        story.Title(),
			Actors:      domainActors,
			WorkObjects: domainWOs,
		},
		DST: dst,
	}

	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling egn document: %w", err)
	}

	return string(out), nil
}

func egnActorType(at discoverydomain.ActorType) (string, error) {
	switch at {
	case discoverydomain.ActorTypePerson:
		return "domainStory:actorPerson", nil
	case discoverydomain.ActorTypeSystem:
		return "domainStory:actorSystem", nil
	case discoverydomain.ActorTypeGroup:
		return "domainStory:actorGroup", nil
	default:
		return "", fmt.Errorf("unknown actor type %q", at)
	}
}

func egnActorDomainKey(at discoverydomain.ActorType) (string, error) {
	switch at {
	case discoverydomain.ActorTypePerson:
		return "Person-svg", nil
	case discoverydomain.ActorTypeSystem:
		return "System-svg", nil
	case discoverydomain.ActorTypeGroup:
		return "Group-svg", nil
	default:
		return "", fmt.Errorf("unknown actor type %q", at)
	}
}

func egnWorkObjectType(wot discoverydomain.WorkObjectType) (string, error) {
	switch wot {
	case discoverydomain.WorkObjectTypeDocument:
		return "domainStory:workObjectDocument", nil
	case discoverydomain.WorkObjectTypeFolder:
		return "domainStory:workObjectFolder", nil
	case discoverydomain.WorkObjectTypeCall:
		return "domainStory:workObjectCall", nil
	case discoverydomain.WorkObjectTypeEmail:
		return "domainStory:workObjectEmail", nil
	case discoverydomain.WorkObjectTypeConversation:
		return "domainStory:workObjectConversation", nil
	case discoverydomain.WorkObjectTypeInfo:
		return "domainStory:workObjectInfo", nil
	case discoverydomain.WorkObjectTypeData:
		return "domainStory:workObjectData", nil
	default:
		return "", fmt.Errorf("unknown work object type %q", wot)
	}
}

func egnWorkObjectDomainKey(wot discoverydomain.WorkObjectType) (string, error) {
	switch wot {
	case discoverydomain.WorkObjectTypeDocument:
		return "Document-svg", nil
	case discoverydomain.WorkObjectTypeFolder:
		return "Folder-svg", nil
	case discoverydomain.WorkObjectTypeCall:
		return "Call-svg", nil
	case discoverydomain.WorkObjectTypeEmail:
		return "Email-svg", nil
	case discoverydomain.WorkObjectTypeConversation:
		return "Conversation-svg", nil
	case discoverydomain.WorkObjectTypeInfo:
		return "Info-svg", nil
	case discoverydomain.WorkObjectTypeData:
		return "Data-svg", nil
	default:
		return "", fmt.Errorf("unknown work object type %q", wot)
	}
}
