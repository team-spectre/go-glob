package glob

import (
	"fmt"
	"unicode"
)

type runeMatchNotSet struct {
	Dense0 uint64
	Dense1 uint64
	Ranges sortedRanges
}

func (m *runeMatchNotSet) ForEachRange(fn func(lo, hi rune)) {
	sweeper := rune(0)
	for _, r := range m.Ranges {
		if r.Lo > 0 {
			loSub1 := r.Lo - 1
			if sweeper <= loSub1 {
				fn(sweeper, loSub1)
			}
		}
		sweeper = r.Hi + 1
	}
	if sweeper <= unicode.MaxRune {
		fn(sweeper, unicode.MaxRune)
	}
}

func (m *runeMatchNotSet) MatchRune(ch rune) bool {
	if ch < 0x40 {
		bit := denseBit(ch)
		return (m.Dense0 & bit) == 0
	} else if ch < 0x80 {
		bit := denseBit(ch)
		return (m.Dense1 & bit) == 0
	} else {
		for _, r := range m.Ranges {
			if ch >= r.Lo && ch <= r.Hi {
				return false
			}
		}
		return true
	}
}

func (m *runeMatchNotSet) String() string {
	buf := make([]rune, 0, 64)
	buf = append(buf, '^')
	for _, r := range m.Ranges {
		buf = safeAppendRune(buf, r.Lo)
		if r.Lo < r.Hi {
			buf = append(buf, '-')
			buf = safeAppendRune(buf, r.Hi)
		}
	}
	return runesToString(buf)
}

func (m *runeMatchNotSet) GoString() string {
	return fmt.Sprintf("glob.MustCompileRuneMatcher(%q)", m.String())
}

func (m *runeMatchNotSet) Not() RuneMatcher {
	return &runeMatchSet{
		Dense0: m.Dense0,
		Dense1: m.Dense1,
		Ranges: m.Ranges,
	}
}

var _ RuneMatcher = (*runeMatchNotSet)(nil)
