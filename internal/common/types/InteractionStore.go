package types

import "fmt"

type InteractionStore struct {
	outgoing         map[GeneID]InteractionIDSet // outgoing interactions
	incoming         map[GeneID]InteractionIDSet // incoming interactions
	interactionCount int                         // Number of interactions
	interactionTypes *InteractionTypeIDMap       // Interaction types
	isRegulatory     IsRegulatory                // Regulatory interactions
}

// NewInteractionStore creates a new InteractionStore
func NewInteractionStore() *InteractionStore {
	return &InteractionStore{
		outgoing:         make(map[GeneID]InteractionIDSet),
		incoming:         make(map[GeneID]InteractionIDSet),
		interactionCount: 0,
		interactionTypes: NewInteractionTypeIDMap(),
		isRegulatory:     NewIsRegulatory(),
	}
}

func (s *InteractionStore) Outgoing() map[GeneID]InteractionIDSet {
	return s.outgoing
}

func (s *InteractionStore) Incoming() map[GeneID]InteractionIDSet {
	return s.incoming
}

func (s *InteractionStore) InteractionCount() int {
	return s.interactionCount
}

func (s *InteractionStore) InteractionTypes() *InteractionTypeIDMap {
	return s.interactionTypes
}

func (s *InteractionStore) SetInteractionTypes(interactionTypes *InteractionTypeIDMap) {
	s.interactionTypes = interactionTypes
}

func (s *InteractionStore) InteractionType(interactionID InteractionID) string {
	return s.interactionTypes.GetNameFromID(interactionID.Type())
}

func (s *InteractionStore) IsRegulatoryInteraction(interactionID InteractionID) bool {
	return s.isRegulatory[interactionID.Type()]
}

func (s *InteractionStore) InteractionTypeStringList() []string {
	lines := make([]string, 0, len(s.interactionTypes.IdToName()))
	for interactionTypeID, interactionTypeName := range s.interactionTypes.IdToName() {
		regulatory := "non-regulatory"
		if s.isRegulatory[interactionTypeID] {
			regulatory = "regulatory"
		}
		lines = append(lines, fmt.Sprintf("%% %s %s", interactionTypeName, regulatory))
	}
	return lines
}

// AddNode ensures a node exists in the graph
func (s *InteractionStore) AddNode(id GeneID) {
	if (s.outgoing[id] == nil && s.incoming[id] != nil) || (s.outgoing[id] != nil && s.incoming[id] == nil) {
		panic("inconsistent graph: node has incoming interactions but no outgoing interactions")
	}
	if s.outgoing[id] == nil || s.incoming[id] == nil {
		s.outgoing[id] = make(InteractionIDSet)
		s.incoming[id] = make(InteractionIDSet)
	}
}

func (s *InteractionStore) HasNode(id GeneID) bool {
	_, outgoingExists := s.outgoing[id]
	_, incomingExists := s.incoming[id]
	return outgoingExists && incomingExists
}

// AddInteraction adds an edge between two nodes
func (s *InteractionStore) AddInteraction(interaction InteractionID) {
	from, to := interaction.FromTo()
	s.AddNode(from)
	s.AddNode(to)

	// Add the nodes to the graph
	if !s.HasNode(from) {
		s.AddNode(from)
	}
	if !s.HasNode(to) {
		s.AddNode(to)
	}

	// Do not add the interaction if it already exists
	if _, exists := s.outgoing[from][interaction]; exists {
		return
	}

	// Add the interaction to the graph
	s.outgoing[from][interaction] = struct{}{}
	s.incoming[to][interaction] = struct{}{}
	s.interactionCount++
}

// RemoveInteraction deletes an edge
func (s *InteractionStore) RemoveInteraction(interaction InteractionID) {
	from, to := interaction.FromTo()
	if _, exists := s.outgoing[from]; exists {
		delete(s.outgoing[from], interaction)
	}
	if _, exists := s.incoming[to]; exists {
		delete(s.incoming[to], interaction)
	}
	s.interactionCount--
}

// GetOutgoing returns all interactions from a node
func (s *InteractionStore) GetOutgoing(id GeneID) []InteractionID {
	var interactions []InteractionID
	for interaction := range s.outgoing[id] {
		interactions = append(interactions, interaction)
	}
	return interactions
}

// GetIncoming returns all interactions to a node
func (s *InteractionStore) GetIncoming(id GeneID) []InteractionID {
	var interactions []InteractionID
	for interaction := range s.incoming[id] {
		interactions = append(interactions, interaction)
	}
	return interactions
}

// RemoveNode deletes a node and all its interactions
func (s *InteractionStore) RemoveNode(id GeneID) {
	delete(s.outgoing, id)
	delete(s.incoming, id)

	for interaction := range s.outgoing[id] {
		delete(s.outgoing[id], interaction)
	}
	for interaction := range s.incoming[id] {
		delete(s.incoming[id], interaction)
	}
}

// Has returns true if the graph contains the given interaction
func (s *InteractionStore) Has(interaction InteractionID) bool {
	from := interaction.From()
	if _, exists := s.outgoing[from]; exists {
		if _, exists := s.outgoing[from][interaction]; exists {
			return true
		}
	}
	return false
}

// SanityCheck checks if the graph is consistent
func (s *InteractionStore) SanityCheck() {
	incomingCount, outgoingCount := 0, 0
	for from, outgoing := range s.outgoing {
		for interaction := range outgoing {
			outgoingCount++
			ifrom, ito := interaction.FromTo()
			if ifrom != from {
				panic("outgoing interaction does not match source node")
			}
			if _, exists := s.incoming[ito][interaction]; !exists {
				panic("outgoing and incoming interactions do not match")
			}
		}
	}
	for _, incoming := range s.incoming {
		for range incoming {
			incomingCount++
		}
	}
	if incomingCount != outgoingCount {
		panic("outgoing and incoming interaction counts do not match")
	}
}

func (s *InteractionStore) AddInteractionType(interactionTypeString string, isRegulatory bool) {
	id := s.interactionTypes.SetName(interactionTypeString)
	s.isRegulatory[id] = isRegulatory
}

func (s *InteractionStore) GetInteractionTypeID(interactionTypeString string) InteractionTypeID {
	return s.interactionTypes.GetIDFromName(interactionTypeString)
}
