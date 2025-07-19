package types

// PathDirection enum contains the possible directions a path can have at any time
type PathDirection int

const (
	UpstreamPath PathDirection = iota
	DownstreamPath
	UndirectedPath
	UpDownstreamPath
	DownUpstreamPath
	InvalidPathDirection
)

// NewPathDirection creates a new path direction from the given string
func NewPathDirection(id string) PathDirection {
	switch id {
	case "Upstream":
		return UpstreamPath
	case "Downstream":
		return DownstreamPath
	case "Undirected":
		return UndirectedPath
	case "UpDownstream":
		return UpDownstreamPath
	case "DownUpstream":
		return DownUpstreamPath
	default:
		return InvalidPathDirection
	}
}

// String returns the string representation of the path direction
func (d PathDirection) String() string {
	switch d {
	case UpstreamPath:
		return "Upstream"
	case DownstreamPath:
		return "Downstream"
	case UndirectedPath:
		return "Undirected"
	case UpDownstreamPath:
		return "UpDownstream"
	case DownUpstreamPath:
		return "DownUpstream"
	default:
		return "Invalid"
	}
}

// Matches returns true if the given direction matches the reference direction
func (d PathDirection) Matches(referenceDirection PathDirection) bool {
	if d == referenceDirection {
		return true
	}
	// every direction matches unknown direction, unknown means all edges are undirected
	if d == UndirectedPath || referenceDirection == UndirectedPath {
		return true
	}
	// down/up stream is a valid updown/downup stream
	if d == DownstreamPath || d == UpstreamPath {
		return referenceDirection == UpDownstreamPath || referenceDirection == DownUpstreamPath
	}
	return false
}
