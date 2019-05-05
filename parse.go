package glob

import (
	"fmt"
	"strconv"
)

type parser struct {
	runes            []rune
	segments         []segment
	ranges           []runeMatchRange
	partialLiteral   []rune
	partialEscape    []rune
	err              error
	text             string
	i, j             uint
	minLength        uint
	maxLength        uint
	escapeIntroducer rune
	state            parseState
	lastSegmentType  segmentType
	escapeLen        byte
	negate           bool
	wantSet          bool
}

func (p *parser) makeError(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

func (p *parser) fail(format string, args ...interface{}) {
	if p.err == nil {
		p.err = p.makeError(format, args...)
	}
}

func (p *parser) emitSegment(t segmentType, r []rune, m RuneMatcher) {
	p.segments = append(p.segments, segment{stype: t, runes: r, matcher: m})
	p.lastSegmentType = t
}

func (p *parser) emitLiteral(ch rune) {
	if p.partialLiteral == nil {
		p.partialLiteral = takeRuneSlice()
	}
	p.partialLiteral = append(p.partialLiteral, ch)
}

func (p *parser) emitSetLo(ch rune) {
	p.ranges = append(p.ranges, runeMatchRange{Lo: ch, Hi: ch})
}

func (p *parser) emitSetHi(ch rune) {
	index := uint(len(p.ranges)) - 1
	r := &p.ranges[index]
	if ch < r.Lo {
		p.fail("invalid range, lo %U > hi %U", r.Lo, ch)
	} else {
		r.Hi = ch
	}
}

func (p *parser) upgradeSegment(t segmentType) {
	index := uint(len(p.segments)) - 1
	p.segments[index].stype = t
	p.lastSegmentType = t
}

func (p *parser) flushLiteral() {
	if p.partialLiteral == nil {
		return
	}
	runes := copyRunes(p.partialLiteral)
	giveRuneSlice(p.partialLiteral)
	p.partialLiteral = nil
	p.emitSegment(literalSegment, runes, nil)
}

func (p *parser) flushSet() {
	set := buildSet(p.ranges)
	if p.negate {
		set = set.Not()
	}
	p.ranges = nil
	p.negate = false
	p.emitSegment(runeMatchSegment, nil, set)
}

func (p *parser) processEscape(ch rune, ifOct, ifHex, ifPunct parseState, emit func(rune)) {
	// NB: keep in sync with util.go isPunct
	switch ch {
	case 'o':
		p.state = ifOct
		p.escapeIntroducer = ch
		p.escapeLen = 3
		p.partialEscape = takeRuneSlice()

	case 'x':
		p.state = ifHex
		p.escapeIntroducer = ch
		p.escapeLen = 2
		p.partialEscape = takeRuneSlice()

	case 'u':
		p.state = ifHex
		p.escapeIntroducer = ch
		p.escapeLen = 4
		p.partialEscape = takeRuneSlice()

	case 'U':
		p.state = ifHex
		p.escapeIntroducer = ch
		p.escapeLen = 8
		p.partialEscape = takeRuneSlice()

	case '\\':
		fallthrough
	case '*':
		fallthrough
	case '?':
		fallthrough
	case '{':
		fallthrough
	case '}':
		fallthrough
	case '[':
		fallthrough
	case ']':
		fallthrough
	case '^':
		fallthrough
	case '-':
		p.state = ifPunct
		emit(ch)

	case '0':
		p.state = ifPunct
		emit(0)

	default:
		p.fail("invalid escape \\%c", ch)
	}
}

func (p *parser) processOct(ch rune, ifDone parseState, emit func(rune)) {
	if !isOct(ch) {
		str := string(p.partialEscape)
		giveRuneSlice(p.partialEscape)
		p.partialEscape = nil
		p.fail("invalid escape \\%c%s%c", p.escapeIntroducer, str, ch)
		return
	}

	p.partialEscape = append(p.partialEscape, ch)
	if uint(len(p.partialEscape)) < uint(p.escapeLen) {
		return
	}

	str := string(p.partialEscape)
	giveRuneSlice(p.partialEscape)
	p.partialEscape = nil
	u64, err := strconv.ParseUint(str, 8, 32)
	if err != nil {
		panic(err)
	}

	emit(rune(u64))
	p.state = ifDone
}

func (p *parser) processHex(ch rune, ifDone parseState, emit func(rune)) {
	if !isHex(ch) {
		str := string(p.partialEscape)
		giveRuneSlice(p.partialEscape)
		p.partialEscape = nil
		p.fail("invalid escape \\%c%s%c", p.escapeIntroducer, str, ch)
		return
	}

	p.partialEscape = append(p.partialEscape, ch)
	if uint(len(p.partialEscape)) < uint(p.escapeLen) {
		return
	}

	str := string(p.partialEscape)
	giveRuneSlice(p.partialEscape)
	p.partialEscape = nil
	u64, err := strconv.ParseUint(str, 16, 32)
	if err != nil {
		panic(err)
	}

	emit(rune(u64))
	p.state = ifDone
}

func (p *parser) run() {
	for p.i < p.j {
		ch := p.runes[p.i]
		p.i++

		switch p.state {
		case rootState:
			switch ch {
			case '[':
				p.flushLiteral()
				p.state = charsetInitialState

			case ']':
				p.fail("unexpected ']'")
				return

			case '{':
				p.fail("alternative matches are not yet implemented: '{'")
				return

			case '}':
				p.fail("unexpected '}'")
				return

			case '*':
				p.flushLiteral()
				if p.lastSegmentType == doubleStarSegment {
					p.fail("unexpected '***'")
					return
				}
				if p.lastSegmentType == starSegment {
					p.upgradeSegment(doubleStarSegment)
					continue
				}
				p.emitSegment(starSegment, nil, nil)

			case '?':
				p.flushLiteral()
				p.emitSegment(questionSegment, nil, nil)

			case '\\':
				p.state = rootEscState

			case '/':
				if p.lastSegmentType == doubleStarSegment {
					p.upgradeSegment(doubleStarSlashSegment)
					continue
				}
				fallthrough
			default:
				p.emitLiteral(ch)
			}

		case rootEscState:
			p.processEscape(ch, rootOctState, rootHexState, rootState, p.emitLiteral)

		case rootOctState:
			p.processOct(ch, rootState, p.emitLiteral)

		case rootHexState:
			p.processHex(ch, rootState, p.emitLiteral)

		case charsetInitialState:
			switch ch {
			case '[':
				p.fail("unexpected '['")
				return

			case ']':
				if p.wantSet {
					p.fail("unexpected ']'")
					return
				}
				p.flushSet()
				p.state = rootState

			case '\\':
				p.state = charsetHeadEscState

			case '^':
				p.negate = true
				p.state = charsetHeadState

			default:
				p.emitSetLo(ch)
				p.state = charsetMidState
			}

		case charsetHeadState:
			switch ch {
			case '[':
				p.fail("unexpected '['")
				return

			case ']':
				if p.wantSet {
					p.fail("unexpected ']'")
					return
				}
				p.flushSet()
				p.state = rootState

			case '\\':
				p.state = charsetHeadEscState

			default:
				p.emitSetLo(ch)
				p.state = charsetMidState
			}

		case charsetHeadEscState:
			p.processEscape(ch, charsetHeadOctState, charsetHeadHexState, charsetMidState, p.emitSetLo)

		case charsetHeadOctState:
			p.processOct(ch, charsetMidState, p.emitSetLo)

		case charsetHeadHexState:
			p.processHex(ch, charsetMidState, p.emitSetLo)

		case charsetMidState:
			switch ch {
			case '[':
				p.fail("unexpected '['")
				return

			case ']':
				if p.wantSet {
					p.fail("unexpected ']'")
					return
				}
				p.flushSet()
				p.state = rootState

			case '\\':
				p.state = charsetHeadEscState

			case '-':
				p.state = charsetTailState

			default:
				p.emitSetLo(ch)
				p.state = charsetMidState
			}

		case charsetTailState:
			switch ch {
			case '[':
				p.fail("unexpected '['")
				return

			case ']':
				if p.wantSet {
					p.fail("unexpected ']'")
					return
				}
				p.emitSetLo('-')
				p.flushSet()
				p.state = rootState

			case '\\':
				p.state = charsetTailEscState

			default:
				p.emitSetHi(ch)
				p.state = charsetHeadState
			}

		case charsetTailEscState:
			p.processEscape(ch, charsetTailOctState, charsetTailHexState, charsetHeadState, p.emitSetHi)

		case charsetTailOctState:
			p.processOct(ch, charsetHeadState, p.emitSetHi)

		case charsetTailHexState:
			p.processHex(ch, charsetHeadState, p.emitSetHi)
		}
	}
	if p.err != nil {
		return
	}

	switch p.state {
	case rootState:
		if p.wantSet {
			panic(p.makeError("BUG! parseState is %#v but wantSet is true", p.state))
		}
		p.flushLiteral()

	case charsetInitialState:
		fallthrough
	case charsetHeadState:
		fallthrough
	case charsetMidState:
		if !p.wantSet {
			p.fail("unterminated character set")
			return
		}
		p.flushSet()

	case charsetTailState:
		if !p.wantSet {
			p.fail("unterminated character set")
			return
		}
		p.emitSetLo('-')
		p.flushSet()

	default:
		p.fail("unterminated backslash escape")
		return
	}

	min := uint(0)
	max := uint(0)
	si := uint(0)
	sj := uint(len(p.segments))
	for sj > si {
		sj--
		seg := &p.segments[sj]

		var segMin, segMax uint
		switch seg.stype {
		case literalSegment:
			segMin = uint(len(seg.runes))
			segMax = uint(len(seg.runes))

		case runeMatchSegment:
			fallthrough
		case questionSegment:
			segMin = 1
			segMax = 1

		case starSegment:
			fallthrough
		case doubleStarSegment:
			fallthrough
		case doubleStarSlashSegment:
			segMin = 0
			segMax = uintMax

		default:
			panic(fmt.Errorf("BUG! unknown segmentType %#v", seg.stype))
		}

		min += segMin
		if max == uintMax || segMax == uintMax {
			max = uintMax
		} else {
			max += segMax
		}

		seg.minLength = min
		seg.maxLength = max
	}
	p.minLength = min
	p.maxLength = max
}
