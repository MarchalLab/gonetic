package interpretation

import "encoding/base64"

// BoolsToBase64 converts a slice of booleans to a base64 encoded string
func BoolsToBase64(bools []bool) string {
	return bytesToBase64(boolsToBytes(bools))
}

// boolsToBytes converts a slice of booleans to a byte slice
func boolsToBytes(bools []bool) []byte {
	bytes := make([]byte, (len(bools)+7)/8) // +7 to round up
	for i, b := range bools {
		if b {
			bytes[i/8] |= 1 << (7 - (i % 8)) // MSB-first
		}
	}
	return bytes
}

// bytesToBase64 converts a byte slice to a base64 encoded string
func bytesToBase64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}
