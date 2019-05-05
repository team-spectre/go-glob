package glob

import (
	"fmt"
)

type RuneMatcher interface {
	MatchRune(rune) bool
	ForEachRange(func(lo, hi rune))
	fmt.Stringer
	fmt.GoStringer
	Not() RuneMatcher
}

func MustCompileRuneMatcher(input string) RuneMatcher {
	compiled, err := CompileRuneMatcher(input)
	if err != nil {
		panic(err)
	}
	return compiled
}

func CompileRuneMatcher(input string) (RuneMatcher, error) {
	if input == "" {
		return None(), nil
	}
	if input == "^" {
		return Any(), nil
	}

	var p parser
	p.runes = stringToRunes(input)
	p.segments = make([]segment, 0, 1)
	p.ranges = make([]runeMatchRange, 0, 16)
	p.text = input
	p.i = 0
	p.j = uint(len(p.runes))
	p.state = charsetInitialState
	p.wantSet = true

	p.run()

	if p.err != nil {
		return nil, fmt.Errorf("failed to parse character set: %q: %v", input, p.err)
	}
	if len(p.segments) != 1 {
		panic(fmt.Errorf("BUG! expected 1 runeMatchSegment, got %d segments", len(p.segments)))
	}
	if p.segments[0].stype != runeMatchSegment {
		panic(fmt.Errorf("BUG! expected 1 runeMatchSegment, got 1 %#v", p.segments[0].stype))
	}
	return p.segments[0].matcher, nil
}
