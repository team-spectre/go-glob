package glob

import (
	"fmt"
	"unicode"
)

type runeMatchNotRange struct{ Lo, Hi rune }

func (m *runeMatchNotRange) ForEachRange(fn func(lo, hi rune)) {
	if m.Lo > 0 {
		fn(0, m.Lo-1)
	}
	if m.Hi < unicode.MaxRune {
		fn(m.Hi+1, unicode.MaxRune)
	}
}

func (m *runeMatchNotRange) MatchRune(ch rune) bool {
	return (ch < m.Lo || ch > m.Hi)
}

func (m *runeMatchNotRange) String() string {
	buf := make([]rune, 0, 24)
	buf = append(buf, '^')
	buf = safeAppendRune(buf, m.Lo)
	if m.Lo < m.Hi {
		buf = append(buf, '-')
		buf = safeAppendRune(buf, m.Hi)
	}
	return runesToString(buf)
}

func (m *runeMatchNotRange) GoString() string {
	return fmt.Sprintf("glob.Range(%q, %q).Not()", m.Lo, m.Hi)
}

func (m *runeMatchNotRange) Not() RuneMatcher {
	return &runeMatchRange{Lo: m.Lo, Hi: m.Hi}
}

var _ RuneMatcher = (*runeMatchNotRange)(nil)
