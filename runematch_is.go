package glob

import (
	"fmt"
)

type runeMatchIs struct{ Rune rune }

func (m *runeMatchIs) ForEachRange(fn func(lo, hi rune)) {
	fn(m.Rune, m.Rune)
}

func (m *runeMatchIs) MatchRune(ch rune) bool {
	return ch == m.Rune
}

func (m *runeMatchIs) String() string {
	buf := make([]rune, 0, 10)
	buf = safeAppendRune(buf, m.Rune)
	return string(buf)
}

func (m *runeMatchIs) GoString() string {
	return fmt.Sprintf("glob.Is(%q)", m.Rune)
}

func (m *runeMatchIs) Not() RuneMatcher {
	return &runeMatchNotIs{Rune: m.Rune}
}

var _ RuneMatcher = (*runeMatchIs)(nil)

func Is(ch rune) RuneMatcher {
	return &runeMatchIs{Rune: ch}
}
