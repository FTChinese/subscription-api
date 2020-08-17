package validator

import (
	"strings"
	"unicode/utf8"
)

// Required checks if a string is empty.
// Leading and trailing spaces are trimmed.
func Required(str string) bool {
	return strings.TrimSpace(str) != ""
}

// StringInLength checks if a string's length, including multi bytes string,
// is within a range, inclusive.
func StringInLength(str string, min, max int) bool {
	if min > max {
		min, max = max, min
	}
	strLength := utf8.RuneCountInString(str)
	return strLength >= min && strLength <= max
}

// MinStringLength checks if a string's length is longer than min
func MinStringLength(str string, min int) bool {
	strLength := utf8.RuneCountInString(str)
	return strLength >= min
}

// MaxStringLength checks if a string's length is under max
// Return true if the length of str is under or equal to max; false otherwise
func MaxStringLength(str string, max int) bool {
	strLength := utf8.RuneCountInString(str)
	return strLength <= max
}
