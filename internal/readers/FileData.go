package readers

type FileData struct {
	ID      string
	Headers map[string]int
	Entries [][]string
}
