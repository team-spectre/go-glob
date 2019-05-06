package glob

import (
	"fmt"
)

type runeMatchRange struct{ Lo, Hi rune }

func (m *runeMatchRange) ForEachRange(fn func(lo, hi rune)) {
	fn(m.Lo, m.Hi)
}

func (m *runeMatchRange) MatchRune(ch rune) bool {
	return (ch >= m.Lo && ch <= m.Hi)
}

func (m *runeMatchRange) String() string {
	buf := make([]rune, 0, 24)
	buf = safeAppendRune(buf, m.Lo)
	if m.Lo < m.Hi {
		buf = append(buf, '-')
		buf = safeAppendRune(buf, m.Hi)
	}
	return string(buf)
}

func (m *runeMatchRange) GoString() string {
	return fmt.Sprintf("glob.Range(%q, %q)", m.Lo, m.Hi)
}

func (m *runeMatchRange) Not() RuneMatcher {
	return &runeMatchNotRange{Lo: m.Lo, Hi: m.Hi}
}

var _ RuneMatcher = (*runeMatchRange)(nil)

func Range(lo, hi rune) RuneMatcher {
	if lo < hi {
		return &runeMatchRange{Lo: lo, Hi: hi}
	} else if lo == hi {
		return Is(lo)
	} else {
		panic(fmt.Errorf("lo %q is greater than hi %q", lo, hi))
	}
}
