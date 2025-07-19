package types

type GeneConditionMap[T any] map[GeneID]map[Condition]T

func (m GeneConditionMap[T]) Add(gene GeneID, condition Condition, score T) {
	if _, ok := m[gene]; !ok {
		m[gene] = make(map[Condition]T)
	}
	m[gene][condition] = score
}
func (m GeneConditionMap[T]) Get(gene GeneID, condition Condition) (T, bool) {
	if _, ok := m[gene]; !ok {
		return *new(T), false
	}
	val, ok := m[gene][condition]
	return val, ok
}
