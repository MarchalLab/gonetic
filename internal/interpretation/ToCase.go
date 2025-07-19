package interpretation

import "strings"

// ToCamelCase converts a string to camel case
func ToCamelCase(s string) string {
	if s == "" {
		return ""
	}
	// replace separators with spaces
	for _, sep := range []string{"-", "_", "\t", "\n"} {
		s = strings.ReplaceAll(s, sep, " ")
	}
	// split words
	words := strings.Split(s, " ")
	// capitalize first letter of each word except the first one
	result := strings.ToLower(words[0])
	for _, word := range words[1:] {
		result += strings.ToUpper(word[0:1]) + strings.ToLower(word[1:])
	}
	return result
}

// ToPascalCase converts a string to pascal case
func ToPascalCase(s string) string {
	camelCase := ToCamelCase(s)
	// capitalize first letter of camel case
	return strings.ToUpper(camelCase[0:1]) + strings.ToLower(camelCase[1:])
}
