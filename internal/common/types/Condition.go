package types

// ConditionID is an integer that represents a condition ID
type ConditionID int

type Condition string

func (c Condition) String() string {
	return string(c)
}

type ConditionSet map[Condition]struct{}
