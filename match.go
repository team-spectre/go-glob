package glob

import (
	"fmt"
)

type Matcher struct {
	g      *Glob
	runes  []rune
	memo   memoMap
	ri, rj uint
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
		ri:    uintMax,
		rj:    uintMax,
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

	if m.sk >= m.sn {
		m.ri = m.rk
		m.rj = m.rn
		m.ok = m.ok && (m.rk >= m.rn)
		return false
	}

	key := memoKey{m.rk, m.sk}
	seg := m.g.segments[m.sk]
	m.sk++

	remain := m.rn - m.rk
	if remain < seg.minLength || remain > seg.maxLength {
		m.ri = uintMax
		m.rj = uintMax
		m.ok = false
		return false
	}

	memo := m.memo[key]
	if memo != nil {
		if memo.rejected {
			m.ri = uintMax
			m.rj = uintMax
			m.ok = false
			return false
		}
		if memo.checked {
			m.ri = m.rk
			m.rj = memo.index
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
		m.ri = uintMax
		m.rj = uintMax
		m.ok = false
		return false
	}
	m.ri = m.rk
	m.rj = index
	m.rk = index
	return true
}

func (m *Matcher) Start() int {
	if m == nil || m.ri == uintMax {
		panic(fmt.Errorf("call to Start() after HasNext() returned false"))
	}
	return int(m.ri)
}

func (m *Matcher) End() int {
	if m == nil || m.ri == uintMax {
		panic(fmt.Errorf("call to End() after HasNext() returned false"))
	}
	return int(m.rj)
}

func (m *Matcher) Location() (int, int) {
	if m == nil || m.ri == uintMax {
		panic(fmt.Errorf("call to Location() after HasNext() returned false"))
	}
	return int(m.ri), int(m.rj)
}

func (m *Matcher) Runes() []rune {
	if m == nil || m.ri == uintMax {
		panic(fmt.Errorf("call to Runes() after HasNext() returned false"))
	}
	return m.runes[m.ri:m.rj]
}

func (m *Matcher) Text() string {
	return string(m.Runes())
}

func (m *Matcher) OK() bool {
	if m == nil {
		return false
	}
	return m.ok
}

func (m *Matcher) Matches() bool {
	for m.HasNext() {
	}
	return m.OK()
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
