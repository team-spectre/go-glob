package guts

import (
	"fmt"
)

func (m *Matcher) HasNext(g *Glob) bool {
	// Clear previous capture.
	m.C = Capture{}

	// No more segments? Success iff all input was consumed.
	if m.SegmentI >= m.SegmentJ {
		m.Valid = m.Valid && (m.InputI >= m.InputJ)
		return false
	}

	// Grab next segment, and prepare memoization key while we're here.
	key := MemoKey{InputI: m.InputI, SegmentI: m.SegmentI}
	seg := g.Segments[m.SegmentI]
	m.SegmentI++
	moreSegments := (m.SegmentI < m.SegmentJ)

	// Fast reject if the remaining input is too short or too long to ever match.
	remain := m.InputJ - m.InputI
	if remain < seg.MinLength || remain > seg.MaxLength {
		m.Valid = false
		return false
	}

	// Check for memoized results.
	memo := m.Memo[key]
	if memo != nil {
		if memo.Rejected {
			m.Valid = false
			return false
		}
		if memo.Checked {
			m.C = Capture{
				InputP:   key.InputI,
				InputQ:   memo.Index,
				SegmentP: key.SegmentI,
				PatternP: seg.PatternP,
				PatternQ: seg.PatternQ,
			}
			m.InputI = memo.Index
			return true
		}
		panic(fmt.Errorf("BUG! infinite recursion"))
	}

	// Match some input, then memoize the outcome.
	memo = &MemoValue{Index: UintMax}
	m.Memo[key] = memo
	index, ok := m.Tick(g, seg, moreSegments)
	memo.Checked = true
	memo.Rejected = false
	memo.Index = index
	if !ok {
		memo.Rejected = true
		m.Valid = false
		return false
	}
	m.C = Capture{
		InputP:   key.InputI,
		InputQ:   index,
		SegmentP: key.SegmentI,
		PatternP: seg.PatternP,
		PatternQ: seg.PatternQ,
	}
	m.InputI = index
	return true
}

func (m *Matcher) Capture() *Capture {
	if !m.Valid {
		panic(fmt.Errorf("call to Capture() after HasNext() return false"))
	}
	return &m.C
}

func (m *Matcher) OK() bool {
	return m.Valid
}

func (m *Matcher) Matches(g *Glob) bool {
	for m.HasNext(g) {
	}
	return m.OK()
}

func (m *Matcher) WouldAccept(g *Glob, i uint) bool {
	var dupe Matcher
	dupe = *m
	dupe.InputI = i
	return dupe.Matches(g)
}

func (m *Matcher) Tick(g *Glob, seg Segment, moreSegments bool) (uint, bool) {
	inputI := m.InputI
	inputJ := inputI
	inputL := m.InputJ

	switch seg.Type {
	case LiteralSegment:
		inputJ += uint(len(seg.Literal.Runes))
		if inputJ > inputL {
			return 0, false
		}
		runes := m.Input.Runes[inputI:inputJ]
		if !EqualRunes(seg.Literal.Runes, runes) {
			return 0, false
		}
		return inputJ, true

	case RuneMatchSegment:
		inputJ++
		if inputJ > inputL {
			return 0, false
		}
		ch := m.Input.Runes[inputI]
		if !seg.Matcher.MatchRune(ch) {
			return 0, false
		}
		return inputJ, true

	case QuestionSegment:
		inputJ++
		if inputJ > inputL {
			return 0, false
		}
		ch := m.Input.Runes[inputI]
		if ch == '/' {
			return 0, false
		}
		return inputJ, true

	case StarSegment:
		// find the next '/'
		for inputJ < inputL && m.Input.Runes[inputJ] != '/' {
			inputJ++
		}

		// no segments after this?
		// -> either there is a '/', or there isn't
		// -> -> no '/': accept the rest of the string, no calculations needed
		// -> -> yes '/': "accept" up to just before the slash, then reject on next tick
		if !moreSegments {
			return inputJ, true
		}

		// accept string where (length ∈ [0..n]) given n := (inputJ - inputI), longer is better
		inputUB := inputJ
		if m.WouldAccept(g, inputJ) {
			return inputJ, true
		}
		for inputJ > inputI {
			inputJ--
			if m.WouldAccept(g, inputJ) {
				return inputJ, true
			}
		}

		// did not find any length which would lead to a match
		// -> blindly accept the maximum permissible length, then reject on some future tick
		return inputUB, true

	case DoubleStarSegment:
		inputJ = inputL

		// accept empty string
		if inputI >= inputJ {
			return inputJ, true
		}

		// no segments after this?
		// -> accept rest of string, no further calculations needed
		if !moreSegments {
			return inputJ, true
		}

		// accept string where (length ∈ [0..n]) given n := (inputJ - inputI), longer is better
		if m.WouldAccept(g, inputJ) {
			return inputJ, true
		}
		for inputJ > inputI {
			inputJ--
			if m.WouldAccept(g, inputJ) {
				return inputJ, true
			}
		}
		return inputJ, true

	case DoubleStarSlashSegment:
		// find the last '/'
		slashes := make(IndexSet, inputL-inputI)
		inputJ = inputI
		slashes[inputJ] = true
		for inputK := inputI; inputK < inputL; inputK++ {
			if m.Input.Runes[inputK] == '/' {
				inputJ = inputK + 1
				slashes[inputJ] = true
			}
		}

		// no segments after this?
		// -> either the final rune is '/', or it isn't
		// -> -> is '/': accept the rest of the string, no calculations needed
		// -> -> not '/': "accept" the longest permissible string, then reject on next tick
		if !moreSegments {
			return inputJ, true
		}

		// accept string where [(length ∈ [0..n]) ∧ (inputJ ∈ slashes)] given n := (inputJ - inputI), longer is better
		inputUB := inputJ
		if slashes[inputJ] && m.WouldAccept(g, inputJ) {
			return inputJ, true
		}
		for inputJ > inputI {
			inputJ--
			if slashes[inputJ] && m.WouldAccept(g, inputJ) {
				return inputJ, true
			}
		}

		// did not find any length which would lead to a match
		// -> blindly accept the maximum permissible length, then reject on some future tick
		return inputUB, true

	default:
		panic(fmt.Errorf("BUG! unknown SegmentType %#v", seg.Type))
	}
}
