package glob

import (
	"fmt"
)

type RuneMatcher interface {
	MatchRune(rune) bool
	ForEachRange(func(lo, hi rune))
	Not() RuneMatcher
	fmt.Stringer
	fmt.GoStringer
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
	p.setInput(input)
	p.segments = make([]segment, 0, 1)
	p.ranges = make([]runeMatchRange, 0, 16)
	p.state = charsetInitialState
	p.wantSet = true

	p.run()

	if p.err != nil {
		return nil, fmt.Errorf("failed to parse character set: %q: %v", p.inputString, p.err)
	}
	if len(p.segments) != 1 {
		panic(fmt.Errorf("BUG! expected 1 runeMatchSegment, got %d segments", len(p.segments)))
	}
	if p.segments[0].stype != runeMatchSegment {
		panic(fmt.Errorf("BUG! expected 1 runeMatchSegment, got 1 %#v", p.segments[0].stype))
	}
	return p.segments[0].matcher, nil
}
