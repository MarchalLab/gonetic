package types

import "fmt"

func NewInteractionType(name string, regulatory bool) InteractionType {
	return InteractionType{
		Name:       name,
		Regulatory: regulatory,
	}
}

type InteractionType struct {
	Name       string
	Regulatory bool
}

func (typ InteractionType) String() string {
	var regulatory string
	if typ.Regulatory {
		regulatory = "regulatory"
	} else {
		regulatory = "non-regulatory"
	}
	return fmt.Sprintf("%% %s %s", typ.Name, regulatory)
}

type InteractionTypeSet map[string]InteractionType

type InteractionDirection bool

const (
	DirectedInteraction   InteractionDirection = true
	UndirectedInteraction InteractionDirection = false
)

func (d InteractionDirection) String() string {
	if d {
		return "directed"
	}
	return "undirected"
}
