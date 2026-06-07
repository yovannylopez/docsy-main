package test_utils

// StringPtr returns a pointer to s for test fixtures (avoids shared literals in tests).
func StringPtr(s string) *string {
	return &s
}
