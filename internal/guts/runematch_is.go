package guts

import (
	"unicode"
)

func (m *IsMatch) ForEachRange(fn func(lo, hi rune)) {
	fn(m.Rune, m.Rune)
}

func (m *IsMatch) MatchRune(ch rune) bool {
	return ch == m.Rune
}

func (m *IsMatch) Not() RuneMatcher {
	return &IsNotMatch{Rune: m.Rune}
}

func (m *IsNotMatch) ForEachRange(fn func(lo, hi rune)) {
	if m.Rune > 0 {
		fn(0, m.Rune-1)
	}
	if m.Rune < unicode.MaxRune {
		fn(m.Rune+1, unicode.MaxRune)
	}
}

func (m *IsNotMatch) MatchRune(ch rune) bool {
	return ch != m.Rune
}

func (m *IsNotMatch) Not() RuneMatcher {
	return &IsMatch{Rune: m.Rune}
}
