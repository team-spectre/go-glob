package glob

type runeMatchNotAny struct{}

func (*runeMatchNotAny) ForEachRange(func(lo, hi rune)) {
	// no op
}

func (*runeMatchNotAny) MatchRune(ch rune) bool {
	return false
}

func (*runeMatchNotAny) String() string {
	return ""
}

func (*runeMatchNotAny) GoString() string {
	return "glob.None()"
}

func (*runeMatchNotAny) Not() RuneMatcher {
	return Any()
}

var noneMatcher RuneMatcher = &runeMatchNotAny{}

func None() RuneMatcher {
	return noneMatcher
}
