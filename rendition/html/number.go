package html

import (
	"strings"
	"unicode"
)

// A Text that is one number ("$850,000,000,000.00", a grid's cell) may
// break after its group separators: a zero-width space follows each, which
// Blitz breaks at (it ignores <wbr>). A narrow column then breaks the
// number between groups, not between any two digits, and the number is no
// wider at its narrowest than a group. Prose keeps its numbers whole.

// groupBreaks is a Text's HTML h, its Markdown plain, with break chances
// after its group separators when plain is one number.
func groupBreaks(plain, h string) string {
	if !oneNumber(plain) {
		return h
	}
	return mapText(h, func(s string) string {
		var b strings.Builder
		for i, r := range s {
			b.WriteRune(r)
			if groupSep(s, i, r) {
				b.WriteRune('\u200b')
			}
		}
		return b.String()
	})
}

// oneNumber reports whether s is one number: no spaces, a digit, and
// only digits, separators, signs, percent and currency symbols.
func oneNumber(s string) bool {
	digit := false
	for _, r := range s {
		switch {
		case unicode.IsDigit(r):
			digit = true
		case strings.ContainsRune(",.'\u00a0\u202f+-−%", r), unicode.Is(unicode.Sc, r):
		default:
			return false
		}
	}
	return digit
}

// groupSep reports whether r, at byte i of s, separates groups of digits:
// it is a separator between digits, and three digits follow it, then no
// digit. A decimal point is not one ("500.25"), but in "1.000,00" a point
// is.
func groupSep(s string, i int, r rune) bool {
	if !strings.ContainsRune(",.'\u00a0\u202f", r) || i == 0 || !isDigit(s[i-1]) {
		return false
	}
	rest := s[i+len(string(r)):]
	return len(rest) >= 3 && isDigit(rest[0]) && isDigit(rest[1]) && isDigit(rest[2]) && (len(rest) == 3 || !isDigit(rest[3]))
}

func isDigit(c byte) bool { return '0' <= c && c <= '9' }
