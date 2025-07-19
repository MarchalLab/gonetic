package normalform

import (
	"github.com/MarchalLab/gonetic/internal/common/types"
)

// SubNetwork interface
// This class keeps track of which interaction are selected and which not.
// It also includes methods to select/deselect interactions

type SubNetwork interface {
	// getters
	CommittedSelectionSize() int
	CommittedSelection() types.InteractionIDSet
	UncommittedSelectionSize() int
	LastActionWasCommit() bool
	// mutators
	Commit()
	Expand(types.InteractionIDSet)
	Reduce(types.InteractionID)
	SoftReset()
	Clear()
	// determine uncommitted interactions that are present in the nnf
	Intersect(nnf NNF) types.InteractionIDSet
}

// Generic SubNetwork implementation
type GenericSubNetwork struct {
	committedSelection   types.InteractionIDSet
	uncommittedSelection types.InteractionIDSet
	dirtyRemove          types.InteractionIDSet
	dirtyAdd             types.InteractionIDSet
	lastActionWasCommit  bool
}

func NewGenericSubNetwork() *GenericSubNetwork {
	return &GenericSubNetwork{}
}

func (builder *GenericSubNetwork) CommittedSelectionSize() int {
	return builder.committedSelection.Size()
}

func (builder *GenericSubNetwork) CommittedSelection() types.InteractionIDSet {
	return builder.committedSelection
}

func (builder *GenericSubNetwork) UncommittedSelectionSize() int {
	return builder.uncommittedSelection.Size()
}

func (builder *GenericSubNetwork) UncommittedSelection() types.InteractionIDSet {
	return builder.uncommittedSelection
}

func (builder *GenericSubNetwork) LastActionWasCommit() bool {
	return builder.lastActionWasCommit
}

func (builder *GenericSubNetwork) Commit() {
	for interactionID := range builder.dirtyAdd {
		builder.committedSelection.Set(interactionID)
	}
	for interactionID := range builder.dirtyRemove {
		builder.committedSelection.Delete(interactionID)
	}
	builder.dirtyAdd = types.NewInteractionIDSet()
	builder.dirtyRemove = types.NewInteractionIDSet()
	builder.lastActionWasCommit = true
}

func (builder *GenericSubNetwork) Expand(expansion types.InteractionIDSet) {
	for interactionID := range expansion {
		if builder.committedSelection.Has(interactionID) {
			// interaction is already selected
			continue
		}
		// propose to add interaction
		builder.dirtyAdd.Set(interactionID)
		builder.uncommittedSelection.Set(interactionID)
	}
}

func (builder *GenericSubNetwork) Reduce(reduction types.InteractionID) {
	if builder.committedSelection.Has(reduction) {
		// interaction is selected and can be removed
		builder.dirtyRemove.Set(reduction)
		builder.uncommittedSelection.Delete(reduction)
	}
}

// Method to reset a sub network. It basically deselects previously selected interactions and selects previously pruned interactions (if no commit has been executed for those interactions)
func (builder *GenericSubNetwork) SoftReset() {
	for interactionID := range builder.dirtyRemove {
		builder.uncommittedSelection.Set(interactionID)
	}
	for interactionID := range builder.dirtyAdd {
		builder.uncommittedSelection.Delete(interactionID)
	}
	builder.dirtyAdd = types.NewInteractionIDSet()
	builder.dirtyRemove = types.NewInteractionIDSet()
	builder.lastActionWasCommit = false
}

func (builder *GenericSubNetwork) Clear() {
	builder.uncommittedSelection = types.NewInteractionIDSet()
	builder.committedSelection = types.NewInteractionIDSet()
	builder.dirtyAdd = types.NewInteractionIDSet()
	builder.dirtyRemove = types.NewInteractionIDSet()
	builder.lastActionWasCommit = false
}

func (builder *GenericSubNetwork) Intersect(nnf NNF) types.InteractionIDSet {
	intersection := types.NewInteractionIDSet()
	for interactionID := range nnf.values {
		if builder.uncommittedSelection.Has(interactionID) {
			// interaction is selected
			intersection.Set(interactionID)
		}
	}
	return intersection
}
