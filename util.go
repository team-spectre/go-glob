package glob

import (
	"fmt"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

func isOct(ch rune) bool {
	return (ch >= '0' && ch <= '7')
}

func isHex(ch rune) bool {
	if ch >= '0' && ch <= '9' {
		return true
	}
	if ch >= 'A' && ch <= 'F' {
		return true
	}
	if ch >= 'a' && ch <= 'f' {
		return true
	}
	return false
}

func isPunct(ch rune) bool {
	// NB: keep in sync with parse.go processEscape
	switch ch {
	case '\\':
		fallthrough
	case '*':
		fallthrough
	case '?':
		fallthrough
	case '{':
		fallthrough
	case '}':
		fallthrough
	case '[':
		fallthrough
	case ']':
		fallthrough
	case '^':
		fallthrough
	case '-':
		return true
	}
	return false
}

func safeAppendRune(runes []rune, ch rune) []rune {
	if isPunct(ch) {
		str := fmt.Sprintf("\\%c", ch)
		return append(runes, []rune(str)...)
	} else if unicode.IsGraphic(ch) {
		return append(runes, ch)
	} else if ch == 0x00 {
		return append(runes, '\\', '0')
	} else if ch < 0x80 {
		str := fmt.Sprintf("\\x%02x", ch)
		return append(runes, []rune(str)...)
	} else if ch < 0x10000 {
		str := fmt.Sprintf("\\u%04x", ch)
		return append(runes, []rune(str)...)
	} else {
		str := fmt.Sprintf("\\U%08x", ch)
		return append(runes, []rune(str)...)
	}
}

func denseBit(ch rune) uint64 {
	shift := uint32(ch) & 0x3f
	return uint64(1) << shift
}

func stringToRunes(input string) []rune {
	return []rune(norm.NFD.String(input))
}

func bytesToRunes(input []byte) []rune {
	return stringToRunes(string(input))
}

func runesToString(input []rune) string {
	return string(input)
}

func runesToBytes(input []rune) []byte {
	return []byte(runesToString(input))
}

func equalRunes(a, b []rune) bool {
	if a == nil || b == nil {
		return (a == nil && b == nil)
	}
	if len(a) != len(b) {
		return false
	}
	n := uint(len(a))
	for i := uint(0); i < n; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
