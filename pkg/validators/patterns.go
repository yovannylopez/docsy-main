package validators

import "regexp"

// Expressions compiled once (request validation; avoids MatchString that recompiles each time).
var (
	emailFormatRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	defaultPhoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

	passwordHasUpperRe  = regexp.MustCompile(`[A-Z]`)
	passwordHasLowerRe  = regexp.MustCompile(`[a-z]`)
	passwordHasNumberRe = regexp.MustCompile(`[0-9]`)
	passwordHasSymbolRe = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?]`)

	personNameCharsRe = regexp.MustCompile(`^[a-zA-ZáéíóúÁÉÍÓÚñÑ\s\-'\.]+$`)
)
