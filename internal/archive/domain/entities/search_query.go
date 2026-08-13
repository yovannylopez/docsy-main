package entities

import (
	"strings"
	"unicode"
)

const (
	searchQueryMaxTokens = 8
	searchQueryMinRunes  = 2
	searchYearTokenLen   = 4
)

// TokenizeSearchQuery splits a free-text query into AND tokens for document search.
// Empty/whitespace-only input yields nil. Single-rune tokens are dropped unless they are digits.
// At most searchQueryMaxTokens tokens are returned.
func TokenizeSearchQuery(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Fields(raw)
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		token := strings.TrimSpace(part)
		if token == "" {
			continue
		}
		runes := []rune(token)
		if len(runes) < searchQueryMinRunes && !isAllDigits(runes) {
			continue
		}
		out = append(out, token)
		if len(out) >= searchQueryMaxTokens {
			break
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// IsYearToken reports whether token is exactly four ASCII digits (e.g. "1998").
func IsYearToken(token string) bool {
	if len(token) != searchYearTokenLen {
		return false
	}
	for i := 0; i < searchYearTokenLen; i++ {
		if token[i] < '0' || token[i] > '9' {
			return false
		}
	}
	return true
}

func isAllDigits(runes []rune) bool {
	if len(runes) == 0 {
		return false
	}
	for _, r := range runes {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
