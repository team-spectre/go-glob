package guts

import (
	"unicode"
)

var (
	AnyValue  RuneMatcher = &AnyMatch{}
	NoneValue RuneMatcher = &NoneMatch{}
)

func (*AnyMatch) ForEachRange(fn func(lo, hi rune)) {
	fn(0, unicode.MaxRune)
}

func (*AnyMatch) MatchRune(ch rune) bool {
	return true
}

func (*AnyMatch) Not() RuneMatcher {
	return NoneValue
}

func (*NoneMatch) ForEachRange(func(lo, hi rune)) {
	// no op
}

func (*NoneMatch) MatchRune(ch rune) bool {
	return false
}

func (*NoneMatch) Not() RuneMatcher {
	return AnyValue
}
