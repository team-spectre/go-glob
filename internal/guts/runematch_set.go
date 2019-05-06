package guts

import (
	"sort"
	"unicode"
)

func (x SortedLoHi) Len() int {
	return len(x)
}

func (x SortedLoHi) Less(i, j int) bool {
	a, b := x[i], x[j]
	if a.Lo != b.Lo {
		return a.Lo < b.Lo
	}
	return a.Hi < b.Hi
}

func (x SortedLoHi) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func (m *SetMatch) ForEachRange(fn func(lo, hi rune)) {
	for _, r := range m.Ranges {
		fn(r.Lo, r.Hi)
	}
}

func (m *SetMatch) MatchRune(ch rune) bool {
	if ch < 0x40 {
		bit := DenseBit(ch)
		return (m.Dense0 & bit) == bit
	} else if ch < 0x80 {
		bit := DenseBit(ch)
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

func (m *SetMatch) Not() RuneMatcher {
	return &ExceptSetMatch{
		Dense0: m.Dense0,
		Dense1: m.Dense1,
		Ranges: m.Ranges,
	}
}

func (m *ExceptSetMatch) ForEachRange(fn func(lo, hi rune)) {
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

func (m *ExceptSetMatch) MatchRune(ch rune) bool {
	if ch < 0x40 {
		bit := DenseBit(ch)
		return (m.Dense0 & bit) == 0
	} else if ch < 0x80 {
		bit := DenseBit(ch)
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

func (m *ExceptSetMatch) Not() RuneMatcher {
	return &SetMatch{
		Dense0: m.Dense0,
		Dense1: m.Dense1,
		Ranges: m.Ranges,
	}
}

func BuildSet(ranges []LoHi) RuneMatcher {
	if len(ranges) == 0 {
		return NoneValue
	}
	if len(ranges) == 1 {
		r := ranges[0]
		return Range(r.Lo, r.Hi)
	}

	sorted := make(SortedLoHi, 0, len(ranges))
	for _, r := range ranges {
		if r.Lo <= r.Hi {
			sorted = append(sorted, r)
		}
	}
	sort.Sort(sorted)
	n := uint(len(sorted))

	out := make(SortedLoHi, 0, n)
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
				dense0 |= DenseBit(lo)
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
				dense1 |= DenseBit(lo)
				lo++
			}
		}
	}

	return &SetMatch{
		Dense0: dense0,
		Dense1: dense1,
		Ranges: out,
	}
}
