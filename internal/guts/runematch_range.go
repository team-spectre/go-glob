package guts

import (
	"unicode"
)

func (m *RangeMatch) ForEachRange(fn func(lo, hi rune)) {
	fn(m.Lo, m.Hi)
}

func (m *RangeMatch) MatchRune(ch rune) bool {
	return (ch >= m.Lo && ch <= m.Hi)
}

func (m *RangeMatch) Not() RuneMatcher {
	return &ExceptRangeMatch{Lo: m.Lo, Hi: m.Hi}
}

func (m *ExceptRangeMatch) ForEachRange(fn func(lo, hi rune)) {
	if m.Lo > 0 {
		fn(0, m.Lo-1)
	}
	if m.Hi < unicode.MaxRune {
		fn(m.Hi+1, unicode.MaxRune)
	}
}

func (m *ExceptRangeMatch) MatchRune(ch rune) bool {
	return (ch < m.Lo || ch > m.Hi)
}

func (m *ExceptRangeMatch) Not() RuneMatcher {
	return &RangeMatch{Lo: m.Lo, Hi: m.Hi}
}
