package optimization

// PathID encodes
// (1) path type index (8 bits),
// (2) sample index (24 bits), and
// (3) path index (32 bits)
type PathID uint64

func NewPathID(pathType, sampleIndex, pathIndex int) PathID {
	return PathID(uint64(pathType)<<56 | uint64(sampleIndex)<<32 | uint64(pathIndex))
}

func (id PathID) PathTypeIndex() int {
	return int(id >> 56)
}

func (id PathID) SampleIndex() int {
	return int(id >> 32 & 0xFFFFFF) // 24 bits
}

func (id PathID) PathIndex() int {
	return int(id & 0xFFFFFFFF) // 32 bits
}
