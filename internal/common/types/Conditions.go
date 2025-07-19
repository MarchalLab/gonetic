package types

type Conditions []Condition

func (p Conditions) Len() int           { return len(p) }
func (p Conditions) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p Conditions) Less(i, j int) bool { return p[i] < p[j] }
