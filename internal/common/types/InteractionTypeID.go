package types

type InteractionTypeID uint64

type InteractionTypeIDMap = IDMap[InteractionTypeID, string]

func NewInteractionTypeIDMap() *InteractionTypeIDMap {
	return NewIDMap[InteractionTypeID, string]()
}

type IsRegulatory = map[InteractionTypeID]bool

func NewIsRegulatory() IsRegulatory {
	return make(map[InteractionTypeID]bool)
}
