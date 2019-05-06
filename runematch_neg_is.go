package glob

import (
	"fmt"
	"unicode"
)

type runeMatchNotIs struct{ Rune rune }

func (m *runeMatchNotIs) ForEachRange(fn func(lo, hi rune)) {
	if m.Rune > 0 {
		fn(0, m.Rune-1)
	}
	if m.Rune < unicode.MaxRune {
		fn(m.Rune+1, unicode.MaxRune)
	}
}

func (m *runeMatchNotIs) MatchRune(ch rune) bool {
	return ch != m.Rune
}

func (m *runeMatchNotIs) String() string {
	buf := make([]rune, 0, 12)
	buf = append(buf, '^')
	buf = safeAppendRune(buf, m.Rune)
	return string(buf)
}

func (m *runeMatchNotIs) GoString() string {
	return fmt.Sprintf("glob.Is(%q).Not()", m.Rune)
}

func (m *runeMatchNotIs) Not() RuneMatcher {
	return Is(m.Rune)
}

var _ RuneMatcher = (*runeMatchNotIs)(nil)
