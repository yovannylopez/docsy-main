package validators

import "strings"

// isBlank reports whether s is empty or contains only whitespace (trimmed).
func isBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}
