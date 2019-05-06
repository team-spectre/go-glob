package glob

import (
	"fmt"
	"unicode"
	"unicode/utf8"

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

func normString(in string) (string, []rune, []uint) {
	tmp0 := takeByteSlice(uint(len(in)))
	defer giveByteSlice(tmp0)

	// tmp0 <- normalized UTF-8 bytes
	tmp0 = norm.NFKC.AppendString(tmp0, in)

	// outString <- copy tmp0 UTF-8 bytes to new UTF-8 string
	outString := string(tmp0)

	// numRunes <- count # of runes in UTF-8 string
	numRunes := uint(utf8.RuneCountInString(outString))

	tmp1 := takeRuneSlice(numRunes)
	defer giveRuneSlice(tmp1)

	// tmp1 <- copy UTF-8 runes to new slice of UTF-32 runes
	// offsetMap <- map from <UTF-32 offset> to <UTF-8 offset>
	offsetMap := make([]uint, 0, numRunes+1)
	for bi, ch := range outString {
		tmp1 = append(tmp1, ch)
		offsetMap = append(offsetMap, uint(bi))
	}
	offsetMap = append(offsetMap, uint(len(outString)))

	// outRunes <- permanent copy of tmp1
	outRunes := copyRunes(tmp1)

	return outString, outRunes, offsetMap
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
