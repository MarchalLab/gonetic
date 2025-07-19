package types

type InteractionIDSet map[InteractionID]struct{}

func NewInteractionIDSet() InteractionIDSet {
	return make(map[InteractionID]struct{})
}

func (set InteractionIDSet) Size() int {
	return len(set)
}

func (set InteractionIDSet) Empty() bool {
	return len(set) == 0
}

func (set InteractionIDSet) Delete(interactionID InteractionID) {
	if !set.Has(interactionID) {
		panic("unset interaction")
	}
	delete(set, interactionID)
}

func (set InteractionIDSet) Set(interactionID InteractionID) InteractionID {
	set[interactionID] = struct{}{}
	return interactionID
}

func (set InteractionIDSet) Has(interactionID InteractionID) bool {
	_, ok := set[interactionID]
	return ok
}

func (set InteractionIDSet) Copy() InteractionIDSet {
	entries := make(map[InteractionID]struct{}, len(set))
	for id := range set {
		entries[id] = struct{}{}
	}
	return entries
}

func (set InteractionIDSet) Equals(other InteractionIDSet) bool {
	if len(set) != len(other) {
		return false
	}
	for interaction := range set {
		if _, ok := other[interaction]; !ok {
			return false
		}
	}
	return true
}

// Count the number of nodes in the subnetwork
func (set InteractionIDSet) NetworkNodesSize() int {
	nodes := make(GeneSet)
	for id := range set {
		from, to := id.FromTo()
		nodes[from] = struct{}{}
		nodes[to] = struct{}{}
	}
	return len(nodes)
}

func (set InteractionIDSet) Add(interactionIDs []InteractionID) {
	for _, interactionID := range interactionIDs {
		set.Set(interactionID)
	}
}
