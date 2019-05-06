package glob

import (
	"fmt"
	"strconv"
)

type parser struct {
	inputString      string
	inputRunes       []rune
	inputIndex       []uint
	segments         []segment
	ranges           []runeMatchRange
	partialLiteral   []rune
	partialEscape    []rune
	lastSegment      *segment
	err              error
	inputP           uint
	inputQ           uint
	inputI           uint
	inputJ           uint
	minLength        uint
	maxLength        uint
	escapeIntroducer rune
	state            parseState
	escapeLen        byte
	negate           bool
	wantSet          bool
}

func (p *parser) setInput(str string) {
	p.inputString, p.inputRunes, p.inputIndex = normString(str)
	p.inputP = 0
	p.inputQ = 0
	p.inputI = 0
	p.inputJ = uint(len(p.inputRunes))
}

func (p *parser) makeError(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

func (p *parser) fail(format string, args ...interface{}) {
	if p.err == nil {
		p.err = p.makeError(format, args...)
	}
}

func (p *parser) emitSegment(t segmentType, patternP, patternQ uint) {
	n := uint(len(p.segments))
	p.segments = append(p.segments, segment{
		stype:    t,
		patternP: patternP,
		patternQ: patternQ,
	})
	p.lastSegment = &p.segments[n]
}

func (p *parser) emitLiteral(ch rune) {
	if p.partialLiteral == nil {
		p.inputP = p.inputQ
		p.partialLiteral = takeRuneSlice(0)
	}
	p.partialLiteral = append(p.partialLiteral, ch)
}

func (p *parser) emitSetLo(ch rune) {
	if p.ranges == nil {
		p.inputP = p.inputQ
		p.ranges = make([]runeMatchRange, 0, 8)
	}
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

func (p *parser) flushLiteral() {
	if p.partialLiteral == nil {
		return
	}

	str := string(p.partialLiteral)
	giveRuneSlice(p.partialLiteral)
	p.partialLiteral = nil

	p.emitSegment(literalSegment, p.inputP, p.inputQ)
	p.lastSegment.setLiteral(str)
}

func (p *parser) flushSet() {
	set := buildSet(p.ranges)
	if p.negate {
		set = set.Not()
	}
	p.ranges = nil
	p.negate = false

	p.emitSegment(runeMatchSegment, p.inputP, p.inputQ)
	p.lastSegment.matcher = set
}

func (p *parser) processEscape(ch rune, ifOct, ifHex, ifPunct parseState, emit func(rune)) {
	// NB: keep in sync with util.go isPunct
	switch ch {
	case 'o':
		p.state = ifOct
		p.escapeIntroducer = ch
		p.escapeLen = 3
		p.partialEscape = takeRuneSlice(0)

	case 'x':
		p.state = ifHex
		p.escapeIntroducer = ch
		p.escapeLen = 2
		p.partialEscape = takeRuneSlice(0)

	case 'u':
		p.state = ifHex
		p.escapeIntroducer = ch
		p.escapeLen = 4
		p.partialEscape = takeRuneSlice(0)

	case 'U':
		p.state = ifHex
		p.escapeIntroducer = ch
		p.escapeLen = 8
		p.partialEscape = takeRuneSlice(0)

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
	for p.inputI < p.inputJ {
		p.inputQ = p.inputI
		ch := p.inputRunes[p.inputI]
		p.inputI++

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
				if p.lastSegment != nil && p.lastSegment.stype == doubleStarSegment {
					p.fail("unexpected '***'")
					return
				}
				if p.lastSegment != nil && p.lastSegment.stype == starSegment {
					p.lastSegment.stype = doubleStarSegment
					continue
				}
				p.emitSegment(starSegment, p.inputQ, p.inputI)

			case '?':
				p.flushLiteral()
				p.emitSegment(questionSegment, p.inputQ, p.inputI)

			case '\\':
				p.state = rootEscState

			case '/':
				if p.lastSegment != nil && p.lastSegment.stype == doubleStarSegment {
					p.lastSegment.stype = doubleStarSlashSegment
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

	p.inputQ = p.inputJ
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
	segmentI := uint(0)
	segmentJ := uint(len(p.segments))
	for segmentJ > segmentI {
		segmentJ--
		seg := &p.segments[segmentJ]

		var segMin, segMax uint
		switch seg.stype {
		case literalSegment:
			segMin = uint(len(seg.literalRunes))
			segMax = uint(len(seg.literalRunes))

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
