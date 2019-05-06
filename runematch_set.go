package glob

import (
	"fmt"
	"sort"
	"unicode"
)

type runeMatchSet struct {
	Dense0 uint64
	Dense1 uint64
	Ranges sortedRanges
}

func (m *runeMatchSet) ForEachRange(fn func(lo, hi rune)) {
	for _, r := range m.Ranges {
		fn(r.Lo, r.Hi)
	}
}

func (m *runeMatchSet) MatchRune(ch rune) bool {
	if ch < 0x40 {
		bit := denseBit(ch)
		return (m.Dense0 & bit) == bit
	} else if ch < 0x80 {
		bit := denseBit(ch)
		return (m.Dense1 & bit) == bit
	} else {
		for _, r := range m.Ranges {
			if ch >= r.Lo && ch <= r.Hi {
				return true
			}
		}
		return false
	}
}

func (m *runeMatchSet) String() string {
	buf := make([]rune, 0, 64)
	for _, r := range m.Ranges {
		buf = safeAppendRune(buf, r.Lo)
		if r.Lo < r.Hi {
			buf = append(buf, '-')
			buf = safeAppendRune(buf, r.Hi)
		}
	}
	return string(buf)
}

func (m *runeMatchSet) GoString() string {
	return fmt.Sprintf("glob.MustCompileRuneMatcher(%q)", m.String())
}

func (m *runeMatchSet) Not() RuneMatcher {
	return &runeMatchNotSet{
		Dense0: m.Dense0,
		Dense1: m.Dense1,
		Ranges: m.Ranges,
	}
}

var _ RuneMatcher = (*runeMatchSet)(nil)

func Set(matchers ...RuneMatcher) RuneMatcher {
	ranges := make([]runeMatchRange, 0, len(matchers))
	for _, m := range matchers {
		m.ForEachRange(func(lo, hi rune) {
			if lo <= hi {
				ranges = append(ranges, runeMatchRange{lo, hi})
			}
		})
	}
	return buildSet(ranges)
}

func buildSet(ranges []runeMatchRange) RuneMatcher {
	if len(ranges) == 0 {
		return None()
	}
	if len(ranges) == 1 {
		r := ranges[0]
		return Range(r.Lo, r.Hi)
	}

	sorted := make(sortedRanges, 0, len(ranges))
	for _, r := range ranges {
		if r.Lo <= r.Hi {
			sorted = append(sorted, r)
		}
	}
	sort.Sort(sorted)
	n := uint(len(sorted))

	out := make(sortedRanges, 0, n)
	out = append(out, sorted[0])
	prevIdx := uint(0)
	prevPtr := &out[prevIdx]
	for i := uint(0); i < n; i++ {
		r := sorted[i]
		if r.Lo > (prevPtr.Hi + 1) {
			out = append(out, r)
			prevIdx++
			prevPtr = &out[prevIdx]
		} else {
			prevPtr.Hi = r.Hi
		}
	}

	if len(out) == 0 {
		return None()
	}

	if len(out) == 1 {
		r := out[0]
		if r.Lo == 0 && r.Hi == unicode.MaxRune {
			return Any()
		}
		if r.Lo == r.Hi {
			return Is(r.Lo)
		}
		return Range(r.Lo, r.Hi)
	}

	var dense0, dense1 uint64
	for _, r := range out {
		if r.Lo < 0x40 {
			lo := r.Lo
			hi := r.Hi
			if hi > 0x3f {
				hi = 0x3f
			}
			for lo <= hi {
				dense0 |= denseBit(lo)
				lo++
			}
		}

		if r.Lo < 0x80 && r.Hi >= 0x40 {
			lo := r.Lo
			hi := r.Hi
			if lo < 0x40 {
				lo = 0x40
			}
			if hi > 0x7f {
				hi = 0x7f
			}
			for lo <= hi {
				dense1 |= denseBit(lo)
				lo++
			}
		}
	}

	return &runeMatchSet{
		Dense0: dense0,
		Dense1: dense1,
		Ranges: out,
	}
}

type sortedRanges []runeMatchRange

func (x sortedRanges) Len() int {
	return len(x)
}

func (x sortedRanges) Less(i, j int) bool {
	a, b := x[i], x[j]
	if a.Lo != b.Lo {
		return a.Lo < b.Lo
	}
	return a.Hi < b.Hi
}

func (x sortedRanges) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

var _ sort.Interface = sortedRanges(nil)
