package optimization

// haveSameKeys checks if two maps have the same keys
func haveSameKeys[K comparable, V any](map1, map2 map[K]V) bool {
	if len(map1) != len(map2) {
		return false
	}
	for key := range map1 {
		if _, exists := map2[key]; !exists {
			return false
		}
	}
	return true
}
