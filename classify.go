package stragts

import (
	"unicode"
)

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}

// isNumeric reports whenever r is the beginning of a numeric simpleValue.
func isNumeric(r rune) bool {
	return r == '+' || r == '-' || ('0' <= r && r <= '9')
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
