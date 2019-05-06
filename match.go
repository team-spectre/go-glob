package glob

import (
	"fmt"
)

type Matcher struct {
	memo        memoMap
	inputString string
	inputRunes  []rune
	inputIndex  []uint
	g           *Glob
	c           Capture
	inputI      uint
	inputJ      uint
	segmentI    uint
	segmentJ    uint
	ok          bool
}

type Capture struct {
	m        *Matcher
	inputP   uint
	inputQ   uint
	segmentP uint
	patternP uint
	patternQ uint
}

type indexSet map[uint]bool

type memoMap map[memoKey]*memoValue

type memoKey struct{ inputI, segmentI uint }

type memoValue struct {
	checked  bool
	rejected bool
	index    uint
}

func (g *Glob) Matcher(input string) *Matcher {
	inputString, inputRunes, inputIndex := normString(input)

	// Grab some data about the glob, taking into account that (*Glob)(nil)
	// is a valid glob that matches the empty string and nothing else.
	var inputJ, segmentJ, minLength, maxLength uint
	inputJ = uint(len(inputRunes))
	if g != nil {
		segmentJ = uint(len(g.segments))
		minLength = g.minLength
		maxLength = g.maxLength
	}

	// Fast reject the input is too short or too long to ever match;
	// (*Matcher)(nil) is a valid matcher that will never match any string.
	if inputJ < minLength || inputJ > maxLength {
		return nil
	}

	// Construct and return a matcher for this glob and input.
	return &Matcher{
		memo:        make(memoMap),
		inputString: inputString,
		inputRunes:  inputRunes,
		inputIndex:  inputIndex,
		g:           g,
		inputI:      0,
		inputJ:      inputJ,
		segmentI:    0,
		segmentJ:    segmentJ,
		ok:          true,
	}
}

func (m *Matcher) Input() string {
	if m == nil {
		return ""
	}
	return m.inputString
}

func (m *Matcher) InputSubstring(i, j uint) string {
	if m == nil {
		return ""
	}
	bi := m.inputIndex[i]
	bj := m.inputIndex[j]
	return m.inputString[bi:bj]
}

func (m *Matcher) HasNext() bool {
	// (*Matcher)(nil) never has more data.
	if m == nil {
		return false
	}

	// Clear previous capture.
	m.c = Capture{}

	// No more segments? Success iff all input was consumed.
	if m.segmentI >= m.segmentJ {
		m.ok = m.ok && (m.inputI >= m.inputJ)
		return false
	}

	// Grab next segment, and prepare memoization key while we're here.
	key := memoKey{inputI: m.inputI, segmentI: m.segmentI}
	seg := m.g.segments[m.segmentI]
	m.segmentI++
	moreSegments := (m.segmentI < m.segmentJ)

	// Fast reject if the remaining input is too short or too long to ever match.
	remain := m.inputJ - m.inputI
	if remain < seg.minLength || remain > seg.maxLength {
		m.ok = false
		return false
	}

	// Check for memoized results.
	memo := m.memo[key]
	if memo != nil {
		if memo.rejected {
			m.ok = false
			return false
		}
		if memo.checked {
			m.c = Capture{
				m:        m,
				inputP:   key.inputI,
				inputQ:   memo.index,
				segmentP: key.segmentI,
				patternP: seg.patternP,
				patternQ: seg.patternQ,
			}
			m.inputI = memo.index
			return true
		}
		panic(fmt.Errorf("BUG! infinite recursion"))
	}

	// Match some input, then memoize the outcome.
	memo = &memoValue{index: uintMax}
	m.memo[key] = memo
	index, ok := m.tick(seg, moreSegments)
	memo.checked = true
	memo.rejected = false
	memo.index = index
	if !ok {
		memo.rejected = true
		m.ok = false
		return false
	}
	m.c = Capture{
		m:        m,
		inputP:   key.inputI,
		inputQ:   index,
		segmentP: key.segmentI,
		patternP: seg.patternP,
		patternQ: seg.patternQ,
	}
	m.inputI = index
	return true
}

func (m *Matcher) Capture() *Capture {
	// (*Matcher)(nil) never has captures.
	if m == nil || m.c.m == nil {
		return nil
	}
	return &m.c
}

func (m *Matcher) OK() bool {
	// (*Matcher)(nil) is never okay.
	if m == nil || !m.ok {
		return false
	}
	return true
}

func (m *Matcher) Matches() bool {
	for m.HasNext() {
	}
	return m.OK()
}

func (c *Capture) PatternLocation() (uint, uint) {
	return c.patternP, c.patternQ
}

func (c *Capture) PatternStart() uint {
	return c.patternP
}

func (c *Capture) PatternEnd() uint {
	return c.patternQ
}

func (c *Capture) Pattern() string {
	return c.m.g.PatternSubstring(c.patternP, c.patternQ)
}

func (c *Capture) InputLocation() (uint, uint) {
	return c.inputP, c.inputQ
}

func (c *Capture) InputStart() uint {
	return c.inputP
}

func (c *Capture) InputEnd() uint {
	return c.inputQ
}

func (c *Capture) Input() string {
	return c.m.InputSubstring(c.inputP, c.inputQ)
}

func (m *Matcher) wouldAccept(inputI uint) bool {
	var dupe Matcher
	dupe = *m
	dupe.inputI = inputI
	return dupe.Matches()
}

func (m *Matcher) tick(seg segment, moreSegments bool) (uint, bool) {
	inputI := m.inputI
	inputJ := inputI
	inputL := m.inputJ

	switch seg.stype {
	case literalSegment:
		inputJ += uint(len(seg.literalRunes))
		if inputJ > inputL {
			return 0, false
		}
		runes := m.inputRunes[inputI:inputJ]
		if !equalRunes(seg.literalRunes, runes) {
			return 0, false
		}
		return inputJ, true

	case runeMatchSegment:
		inputJ++
		if inputJ > inputL {
			return 0, false
		}
		ch := m.inputRunes[inputI]
		if !seg.matcher.MatchRune(ch) {
			return 0, false
		}
		return inputJ, true

	case questionSegment:
		inputJ++
		if inputJ > inputL {
			return 0, false
		}
		ch := m.inputRunes[inputI]
		if ch == '/' {
			return 0, false
		}
		return inputJ, true

	case starSegment:
		// find the next '/'
		for inputJ < inputL && m.inputRunes[inputJ] != '/' {
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
		if m.wouldAccept(inputJ) {
			return inputJ, true
		}
		for inputJ > inputI {
			inputJ--
			if m.wouldAccept(inputJ) {
				return inputJ, true
			}
		}

		// did not find any length which would lead to a match
		// -> blindly accept the maximum permissible length, then reject on some future tick
		return inputUB, true

	case doubleStarSegment:
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
		if m.wouldAccept(inputJ) {
			return inputJ, true
		}
		for inputJ > inputI {
			inputJ--
			if m.wouldAccept(inputJ) {
				return inputJ, true
			}
		}
		return inputJ, true

	case doubleStarSlashSegment:
		// find the last '/'
		slashes := make(indexSet, inputL-inputI)
		inputJ = inputI
		slashes[inputJ] = true
		for inputK := inputI; inputK < inputL; inputK++ {
			if m.inputRunes[inputK] == '/' {
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
		if slashes[inputJ] && m.wouldAccept(inputJ) {
			return inputJ, true
		}
		for inputJ > inputI {
			inputJ--
			if slashes[inputJ] && m.wouldAccept(inputJ) {
				return inputJ, true
			}
		}

		// did not find any length which would lead to a match
		// -> blindly accept the maximum permissible length, then reject on some future tick
		return inputUB, true

	default:
		panic(fmt.Errorf("BUG! unknown segmentType %#v", seg.stype))
	}
}
