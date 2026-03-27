package domain

import (
	"strings"

	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// WithTrust returns a new StorySentence with upgraded trust level.
// Returns the receiver unchanged if newTrust is not higher than current trust.
// Source is cleared when upgrading away from AIResearched (source is only
// required for AIResearched per the domain invariant).
func (s StorySentence) WithTrust(newTrust vo.TrustLevel) StorySentence {
	if !newTrust.IsHigherTrust(s.trust) {
		return s
	}

	source := s.source
	if newTrust != vo.AIResearched {
		source = ""
	}

	return StorySentence{
		step:           s.step,
		subject:        s.subject,
		activity:       s.activity,
		object:         s.object,
		preposition:    s.preposition,
		indirectObject: s.indirectObject,
		trust:          newTrust,
		source:         source,
	}
}

// WithTrust returns a new StoryActor with upgraded trust level.
// Returns the receiver unchanged if newTrust is not higher than current trust.
func (a StoryActor) WithTrust(newTrust vo.TrustLevel) StoryActor {
	if !newTrust.IsHigherTrust(a.trust) {
		return a
	}

	source := a.source
	if newTrust != vo.AIResearched {
		source = ""
	}

	return StoryActor{
		name:      a.name,
		actorType: a.actorType,
		trust:     newTrust,
		source:    source,
	}
}

// WithTrust returns a new WorkObject with upgraded trust level.
// Returns the receiver unchanged if newTrust is not higher than current trust.
func (w WorkObject) WithTrust(newTrust vo.TrustLevel) WorkObject {
	if !newTrust.IsHigherTrust(w.trust) {
		return w
	}

	source := w.source
	if newTrust != vo.AIResearched {
		source = ""
	}

	return WorkObject{
		name:       w.name,
		objectType: w.objectType,
		trust:      newTrust,
		source:     source,
	}
}

// PropagateConfirmation upgrades trust on a confirmed story sentence and its
// referenced actor and work object within the story. When the user accepts a
// sentence unedited, trust becomes UserConfirmed. When the user edits any of
// Subject/Activity/Object (case-sensitive comparison), trust becomes UserStated.
// Rejected sentences return proposed unchanged with no story mutations.
func PropagateConfirmation(
	proposed, confirmed StorySentence,
	accepted bool,
	story *DomainStory,
) StorySentence {
	if !accepted {
		return proposed
	}

	// Determine trust level based on edit detection (case-sensitive).
	newTrust := vo.UserConfirmed
	if proposed.subject != confirmed.subject ||
		proposed.activity != confirmed.activity ||
		proposed.object != confirmed.object {
		newTrust = vo.UserStated
	}

	// Upgrade sentence trust.
	result := confirmed.WithTrust(newTrust)

	// Upgrade actor trust (subject).
	story.UpgradeActorTrust(confirmed.subject, newTrust)

	// Upgrade work object trust (object).
	story.UpgradeWorkObjectTrust(confirmed.object, newTrust)

	return result
}

// UpgradeActorTrust upgrades the trust level of the actor matching the given
// name (case-insensitive). No-op if the actor is not found or if the new trust
// is not higher than the current trust (WithTrust handles the guard).
func (d *DomainStory) UpgradeActorTrust(name string, newTrust vo.TrustLevel) {
	for i, actor := range d.actors {
		if strings.EqualFold(actor.Name(), name) {
			d.actors[i] = actor.WithTrust(newTrust)

			return
		}
	}
}

// UpgradeWorkObjectTrust upgrades the trust level of the work object matching
// the given name (case-insensitive). No-op if not found or if trust is not higher.
func (d *DomainStory) UpgradeWorkObjectTrust(name string, newTrust vo.TrustLevel) {
	for i, wo := range d.workObjects {
		if strings.EqualFold(wo.Name(), name) {
			d.workObjects[i] = wo.WithTrust(newTrust)

			return
		}
	}
}
