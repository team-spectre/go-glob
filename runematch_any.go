package glob

import (
	"unicode"
)

type runeMatchAny struct{}

func (*runeMatchAny) ForEachRange(fn func(lo, hi rune)) {
	fn(0, unicode.MaxRune)
}

func (*runeMatchAny) MatchRune(ch rune) bool {
	return true
}

func (*runeMatchAny) String() string {
	return "^"
}

func (*runeMatchAny) GoString() string {
	return "glob.Any()"
}

func (*runeMatchAny) Not() RuneMatcher {
	return None()
}

var anyMatcher RuneMatcher = &runeMatchAny{}

func Any() RuneMatcher {
	return anyMatcher
}
