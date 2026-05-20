package ui

import "strings"

// sanitizeODataSearch strips dangerous characters from a search string
// before embedding it in an OData filter expression. Only alphanumeric
// characters, spaces, hyphens, and underscores are allowed.
func sanitizeODataSearch(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ' || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
