package guts

import (
	"fmt"
	"unicode"
)

func CompileRuneMatcher(input string) (RuneMatcher, error) {
	if input == "" {
		return NoneValue, nil
	}
	if input == "^" {
		return AnyValue, nil
	}

	var p Parser
	p.Input = Norm(input)
	p.InputJ = uint(len(p.Input.Runes))
	p.Segments = make([]Segment, 0, 1)
	p.Ranges = make([]LoHi, 0, 16)
	p.State = CharsetInitialState
	p.WantSet = true

	p.Run()

	if p.Err != nil {
		return nil, fmt.Errorf("failed to parse character set: %q: %v", p.Input.String, p.Err)
	}
	if len(p.Segments) != 1 {
		panic(fmt.Errorf("BUG! expected 1 RuneMatchSegment, got %d segments", len(p.Segments)))
	}
	if p.Segments[0].Type != RuneMatchSegment {
		panic(fmt.Errorf("BUG! expected 1 RuneMatchSegment, got 1 %#v", p.Segments[0].Type))
	}
	return p.Segments[0].Matcher, nil
}

func Any() RuneMatcher {
	return AnyValue
}

func None() RuneMatcher {
	return NoneValue
}

func Is(ch rune) RuneMatcher {
	return &IsMatch{Rune: ch}
}

func Range(lo, hi rune) RuneMatcher {
	if lo == 0 && hi == unicode.MaxRune {
		return Any()
	} else if lo == hi {
		return Is(lo)
	} else if lo < hi {
		return &RangeMatch{Lo: lo, Hi: hi}
	} else {
		panic(fmt.Errorf("lo %q is greater than hi %q", lo, hi))
	}
}

func Set(matchers ...RuneMatcher) RuneMatcher {
	ranges := make([]LoHi, 0, len(matchers))
	for _, m := range matchers {
		m.ForEachRange(func(lo, hi rune) {
			if lo <= hi {
				ranges = append(ranges, LoHi{lo, hi})
			}
		})
	}
	return BuildSet(ranges)
}
