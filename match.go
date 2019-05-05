package glob

import (
	"fmt"
)

type Matcher struct {
	memo   memoMap
	runes  []rune
	g      *Glob
	c      Capture
	rk, rn uint
	sk, sn uint
	ok     bool
}

type indexSet map[uint]bool

type memoMap map[memoKey]*memoValue

type memoKey struct{ rk, sk uint }

type memoValue struct {
	checked  bool
	rejected bool
	index    uint
}

func (g *Glob) RunesMatcher(input []rune) *Matcher {
	var rn, sn, minLength, maxLength uint
	rn = uint(len(input))
	if g != nil {
		sn = uint(len(g.segments))
		minLength = g.minLength
		maxLength = g.maxLength
	}
	if rn < minLength || rn > maxLength {
		// fast reject if length of input is so (short/long) that it will never match
		return nil
	}
	return &Matcher{
		g:     g,
		runes: input,
		memo:  make(memoMap),
		rk:    0,
		rn:    rn,
		sk:    0,
		sn:    sn,
		ok:    true,
	}
}

func (m *Matcher) HasNext() bool {
	if m == nil {
		return false
	}

	// Clear previous capture
	m.c = Capture{}

	if m.sk >= m.sn {
		m.ok = m.ok && (m.rk >= m.rn)
		return false
	}

	key := memoKey{m.rk, m.sk}
	seg := m.g.segments[m.sk]
	m.sk++

	remain := m.rn - m.rk
	if remain < seg.minLength || remain > seg.maxLength {
		m.ok = false
		return false
	}

	memo := m.memo[key]
	if memo != nil {
		if memo.rejected {
			m.ok = false
			return false
		}
		if memo.checked {
			m.c = Capture{
				m:  m,
				qi: 0,  // FIXME
				qj: 0,  // FIXME
				ri: m.rk,
				rj: memo.index,
				si: m.sk - 1,
			}
			m.rk = memo.index
			return true
		}
		panic(fmt.Errorf("BUG! infinite recursion"))
	}

	memo = &memoValue{index: uintMax}
	m.memo[key] = memo
	index, ok := m.tick(seg, m.sk < m.sn)
	memo.checked = true
	memo.rejected = false
	memo.index = index
	if !ok {
		memo.rejected = true
		m.ok = false
		return false
	}
	m.c = Capture{
		m:  m,
		qi: 0,  // FIXME
		qj: 0,  // FIXME
		ri: m.rk,
		rj: index,
		si: m.sk - 1,
	}
	m.rk = index
	return true
}

func (m *Matcher) Capture() *Capture {
	if m == nil || m.c.m == nil {
		return nil
	}
	return &m.c
}

func (m *Matcher) OK() bool {
	if m != nil && m.ok {
		return true
	}
	return false
}

func (m *Matcher) Matches() bool {
	for m.HasNext() {
	}
	return m.OK()
}

type Capture struct {
	m *Matcher
	qi, qj uint
	ri, rj uint
	si uint
}

func (c *Capture) Pattern() (uint, uint) {
	return c.qi, c.qj
}

func (c *Capture) PatternStart() uint {
	return c.qi
}

func (c *Capture) PatternEnd() uint {
	return c.qj
}

func (c *Capture) PatternRunes() []rune {
	return c.m.g.runes[c.qi:c.qj]
}

func (c *Capture) Input() (uint, uint) {
	return c.ri, c.rj
}

func (c *Capture) InputStart() uint {
	return c.ri
}

func (c *Capture) InputEnd() uint {
	return c.rj
}

func (c *Capture) InputRunes() []rune {
	return c.m.runes[c.ri:c.rj]
}

func (c *Capture) InputString() string {
	return runesToString(c.InputRunes())
}

func (m *Matcher) shallowClone() *Matcher {
	dupe := new(Matcher)
	*dupe = *m
	return dupe
}

func (m *Matcher) wouldAccept(rk uint) bool {
	dupe := m.shallowClone()
	dupe.rk = rk
	return dupe.Matches()
}

func (m *Matcher) tick(seg segment, moreSegments bool) (uint, bool) {
	i := m.rk
	j := i
	n := m.rn

	switch seg.stype {
	case literalSegment:
		j += uint(len(seg.runes))
		if j > n {
			return 0, false
		}
		runes := m.runes[i:j]
		if !equalRunes(seg.runes, runes) {
			return 0, false
		}
		return j, true

	case runeMatchSegment:
		j++
		if j > n {
			return 0, false
		}
		ch := m.runes[i]
		if !seg.matcher.MatchRune(ch) {
			return 0, false
		}
		return j, true

	case questionSegment:
		j++
		if j > n {
			return 0, false
		}
		ch := m.runes[i]
		if ch == '/' {
			return 0, false
		}
		return j, true

	case starSegment:
		// find the next '/'
		for j < n && m.runes[j] != '/' {
			j++
		}

		// no segments after this?
		// -> either there is a '/', or there isn't
		// -> -> no '/': accept the rest of the string, no calculations needed
		// -> -> yes '/': "accept" up to just before the slash, then reject on next tick
		if !moreSegments {
			return j, true
		}

		// accept string of any length ∈ [0..n] where n := (j - i), longer is better
		ub := j
		if m.wouldAccept(j) {
			return j, true
		}
		for j > i {
			j--
			if m.wouldAccept(j) {
				return j, true
			}
		}

		// did not find any length which would lead to a match
		// -> blindly accept the maximum permissible length, then reject on some future tick
		return ub, true

	case doubleStarSegment:
		j = n

		// accept empty string
		if i >= j {
			return j, true
		}

		// no segments after this?
		// -> accept rest of string, no further calculations needed
		if !moreSegments {
			return j, true
		}

		// accept string of any length ∈ [0..n] where n := (j - i), longer is better
		if m.wouldAccept(j) {
			return j, true
		}
		for j > i {
			j--
			if m.wouldAccept(j) {
				return j, true
			}
		}
		return j, true

	case doubleStarSlashSegment:
		// find the last '/'
		slashes := make(indexSet, n-i)
		j = i
		slashes[j] = true
		for k := i; k < n; k++ {
			if m.runes[k] == '/' {
				j = k + 1
				slashes[j] = true
			}
		}

		// no segments after this?
		// -> either the final rune is '/', or it isn't
		// -> -> is '/': accept the rest of the string, no calculations needed
		// -> -> not '/': "accept" the longest permissible string, then reject on next tick
		if !moreSegments {
			return j, true
		}

		// accept string [(length ∈ [0..n]) ∧ (j ∈ slashes)] where n := (j - i), longer is better
		ub := j
		if slashes[j] && m.wouldAccept(j) {
			return j, true
		}
		for j > i {
			j--
			if slashes[j] && m.wouldAccept(j) {
				return j, true
			}
		}

		// did not find any length which would lead to a match
		// -> blindly accept the maximum permissible length, then reject on some future tick
		return ub, true

	default:
		panic(fmt.Errorf("BUG! unknown segmentType %#v", seg.stype))
	}
}
