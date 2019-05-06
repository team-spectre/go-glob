package guts

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

func Norm(in string) ExplodedString {
	var out ExplodedString

	tmp0 := takeByteSlice(uint(len(in)))
	defer giveByteSlice(tmp0)

	// tmp0 <- normalized UTF-8 bytes
	tmp0 = norm.NFKC.AppendString(tmp0, in)

	// out.String <- copy tmp0 UTF-8 bytes to new UTF-8 string
	out.String = string(tmp0)

	// numRunes <- count # of runes in UTF-8 string
	numRunes := uint(utf8.RuneCountInString(out.String))

	tmp1 := takeRuneSlice(numRunes)
	defer giveRuneSlice(tmp1)

	// tmp1 <- copy UTF-8 runes to new slice of UTF-32 runes
	// out.Map <- map from <UTF-32 offset> to <UTF-8 offset>
	out.Map = make([]uint, 0, numRunes+1)
	for bi, ch := range out.String {
		tmp1 = append(tmp1, ch)
		out.Map = append(out.Map, uint(bi))
	}
	out.Map = append(out.Map, uint(len(out.String)))

	// out.Runes <- permanent copy of tmp1
	out.Runes = copyRunes(tmp1)

	return out
}

func (x ExplodedString) Substring(i, j uint) string {
	bi := x.Map[i]
	bj := x.Map[j]
	return x.String[bi:bj]
}

func IsOct(ch rune) bool {
	return (ch >= '0' && ch <= '7')
}

func IsHex(ch rune) bool {
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

func IsPunct(ch rune) bool {
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

func SafeAppendRune(runes []rune, ch rune) []rune {
	if IsPunct(ch) {
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

func DenseBit(ch rune) uint64 {
	shift := uint32(ch) & 0x3f
	return uint64(1) << shift
}

func EqualRunes(a, b []rune) bool {
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
