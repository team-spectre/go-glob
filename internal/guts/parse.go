package guts

import (
	"fmt"
	"strconv"
)

func (p *Parser) MakeError(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

func (p *Parser) Fail(format string, args ...interface{}) {
	if p.Err == nil {
		p.Err = p.MakeError(format, args...)
	}
}

func (p *Parser) EmitSegment(t SegmentType, PatternP, PatternQ uint) {
	n := uint(len(p.Segments))
	p.Segments = append(p.Segments, Segment{
		Type:     t,
		PatternP: PatternP,
		PatternQ: PatternQ,
	})
	p.LastSegment = &p.Segments[n]
}

func (p *Parser) EmitLiteral(ch rune) {
	if p.PartialLiteral == nil {
		p.InputP = p.InputQ
		p.PartialLiteral = takeRuneSlice(0)
	}
	p.PartialLiteral = append(p.PartialLiteral, ch)
}

func (p *Parser) EmitSetLo(ch rune) {
	if p.Ranges == nil {
		p.InputP = p.InputQ
		p.Ranges = make([]LoHi, 0, 8)
	}
	p.Ranges = append(p.Ranges, LoHi{Lo: ch, Hi: ch})
}

func (p *Parser) EmitSetHi(ch rune) {
	index := uint(len(p.Ranges)) - 1
	r := &p.Ranges[index]
	if ch < r.Lo {
		p.Fail("invalid range, lo %U > hi %U", r.Lo, ch)
	} else {
		r.Hi = ch
	}
}

func (p *Parser) FlushLiteral() {
	if p.PartialLiteral == nil {
		return
	}

	str := string(p.PartialLiteral)
	giveRuneSlice(p.PartialLiteral)
	p.PartialLiteral = nil

	p.EmitSegment(LiteralSegment, p.InputP, p.InputQ)
	p.LastSegment.Literal = Norm(str)
}

func (p *Parser) FlushSet() {
	set := BuildSet(p.Ranges)
	if p.Negate {
		set = set.Not()
	}
	p.Ranges = nil
	p.Negate = false

	p.EmitSegment(RuneMatchSegment, p.InputP, p.InputQ)
	p.LastSegment.Matcher = set
}

func (p *Parser) ProcessEscape(ch rune, ifOct, ifHex, ifPunct ParseState, emit func(rune)) {
	// NB: keep in sync with util.go IsPunct
	switch ch {
	case 'o':
		p.State = ifOct
		p.EscapeIntroducer = ch
		p.EscapeLen = 3
		p.PartialEscape = takeRuneSlice(0)

	case 'x':
		p.State = ifHex
		p.EscapeIntroducer = ch
		p.EscapeLen = 2
		p.PartialEscape = takeRuneSlice(0)

	case 'u':
		p.State = ifHex
		p.EscapeIntroducer = ch
		p.EscapeLen = 4
		p.PartialEscape = takeRuneSlice(0)

	case 'U':
		p.State = ifHex
		p.EscapeIntroducer = ch
		p.EscapeLen = 8
		p.PartialEscape = takeRuneSlice(0)

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
		p.State = ifPunct
		emit(ch)

	case '0':
		p.State = ifPunct
		emit(0)

	default:
		p.Fail("invalid escape \\%c", ch)
	}
}

func (p *Parser) ProcessOct(ch rune, ifDone ParseState, emit func(rune)) {
	if !IsOct(ch) {
		str := string(p.PartialEscape)
		giveRuneSlice(p.PartialEscape)
		p.PartialEscape = nil
		p.Fail("invalid escape \\%c%s%c", p.EscapeIntroducer, str, ch)
		return
	}

	p.PartialEscape = append(p.PartialEscape, ch)
	if uint(len(p.PartialEscape)) < uint(p.EscapeLen) {
		return
	}

	str := string(p.PartialEscape)
	giveRuneSlice(p.PartialEscape)
	p.PartialEscape = nil
	u64, Err := strconv.ParseUint(str, 8, 32)
	if Err != nil {
		panic(Err)
	}

	emit(rune(u64))
	p.State = ifDone
}

func (p *Parser) ProcessHex(ch rune, ifDone ParseState, emit func(rune)) {
	if !IsHex(ch) {
		str := string(p.PartialEscape)
		giveRuneSlice(p.PartialEscape)
		p.PartialEscape = nil
		p.Fail("invalid escape \\%c%s%c", p.EscapeIntroducer, str, ch)
		return
	}

	p.PartialEscape = append(p.PartialEscape, ch)
	if uint(len(p.PartialEscape)) < uint(p.EscapeLen) {
		return
	}

	str := string(p.PartialEscape)
	giveRuneSlice(p.PartialEscape)
	p.PartialEscape = nil
	u64, Err := strconv.ParseUint(str, 16, 32)
	if Err != nil {
		panic(Err)
	}

	emit(rune(u64))
	p.State = ifDone
}

func (p *Parser) Run() {
	for p.InputI < p.InputJ {
		p.InputQ = p.InputI
		ch := p.Input.Runes[p.InputI]
		p.InputI++

		switch p.State {
		case RootState:
			switch ch {
			case '[':
				p.FlushLiteral()
				p.State = CharsetInitialState

			case ']':
				p.Fail("unexpected ']'")
				return

			case '{':
				p.Fail("alternative matches are not yet implemented: '{'")
				return

			case '}':
				p.Fail("unexpected '}'")
				return

			case '*':
				p.FlushLiteral()
				if p.LastSegment != nil && p.LastSegment.Type == DoubleStarSegment {
					p.Fail("unexpected '***'")
					return
				}
				if p.LastSegment != nil && p.LastSegment.Type == StarSegment {
					p.LastSegment.Type = DoubleStarSegment
					continue
				}
				p.EmitSegment(StarSegment, p.InputQ, p.InputI)

			case '?':
				p.FlushLiteral()
				p.EmitSegment(QuestionSegment, p.InputQ, p.InputI)

			case '\\':
				p.State = RootEscState

			case '/':
				if p.LastSegment != nil && p.LastSegment.Type == DoubleStarSegment {
					p.LastSegment.Type = DoubleStarSlashSegment
					continue
				}
				fallthrough
			default:
				p.EmitLiteral(ch)
			}

		case RootEscState:
			p.ProcessEscape(ch, RootOctState, RootHexState, RootState, p.EmitLiteral)

		case RootOctState:
			p.ProcessOct(ch, RootState, p.EmitLiteral)

		case RootHexState:
			p.ProcessHex(ch, RootState, p.EmitLiteral)

		case CharsetInitialState:
			switch ch {
			case '[':
				p.Fail("unexpected '['")
				return

			case ']':
				if p.WantSet {
					p.Fail("unexpected ']'")
					return
				}
				p.FlushSet()
				p.State = RootState

			case '\\':
				p.State = CharsetHeadEscState

			case '^':
				p.Negate = true
				p.State = CharsetHeadState

			default:
				p.EmitSetLo(ch)
				p.State = CharsetMidState
			}

		case CharsetHeadState:
			switch ch {
			case '[':
				p.Fail("unexpected '['")
				return

			case ']':
				if p.WantSet {
					p.Fail("unexpected ']'")
					return
				}
				p.FlushSet()
				p.State = RootState

			case '\\':
				p.State = CharsetHeadEscState

			default:
				p.EmitSetLo(ch)
				p.State = CharsetMidState
			}

		case CharsetHeadEscState:
			p.ProcessEscape(ch, CharsetHeadOctState, CharsetHeadHexState, CharsetMidState, p.EmitSetLo)

		case CharsetHeadOctState:
			p.ProcessOct(ch, CharsetMidState, p.EmitSetLo)

		case CharsetHeadHexState:
			p.ProcessHex(ch, CharsetMidState, p.EmitSetLo)

		case CharsetMidState:
			switch ch {
			case '[':
				p.Fail("unexpected '['")
				return

			case ']':
				if p.WantSet {
					p.Fail("unexpected ']'")
					return
				}
				p.FlushSet()
				p.State = RootState

			case '\\':
				p.State = CharsetHeadEscState

			case '-':
				p.State = CharsetTailState

			default:
				p.EmitSetLo(ch)
				p.State = CharsetMidState
			}

		case CharsetTailState:
			switch ch {
			case '[':
				p.Fail("unexpected '['")
				return

			case ']':
				if p.WantSet {
					p.Fail("unexpected ']'")
					return
				}
				p.EmitSetLo('-')
				p.FlushSet()
				p.State = RootState

			case '\\':
				p.State = CharsetTailEscState

			default:
				p.EmitSetHi(ch)
				p.State = CharsetHeadState
			}

		case CharsetTailEscState:
			p.ProcessEscape(ch, CharsetTailOctState, CharsetTailHexState, CharsetHeadState, p.EmitSetHi)

		case CharsetTailOctState:
			p.ProcessOct(ch, CharsetHeadState, p.EmitSetHi)

		case CharsetTailHexState:
			p.ProcessHex(ch, CharsetHeadState, p.EmitSetHi)
		}
	}
	if p.Err != nil {
		return
	}

	p.InputQ = p.InputJ
	switch p.State {
	case RootState:
		if p.WantSet {
			panic(p.MakeError("BUG! ParseState is %#v but WantSet is true", p.State))
		}
		p.FlushLiteral()

	case CharsetInitialState:
		fallthrough
	case CharsetHeadState:
		fallthrough
	case CharsetMidState:
		if !p.WantSet {
			p.Fail("unterminated character set")
			return
		}
		p.FlushSet()

	case CharsetTailState:
		if !p.WantSet {
			p.Fail("unterminated character set")
			return
		}
		p.EmitSetLo('-')
		p.FlushSet()

	default:
		p.Fail("unterminated backslash escape")
		return
	}

	min := uint(0)
	max := uint(0)
	segmentI := uint(0)
	segmentJ := uint(len(p.Segments))
	for segmentJ > segmentI {
		segmentJ--
		seg := &p.Segments[segmentJ]

		var segMin, segMax uint
		switch seg.Type {
		case LiteralSegment:
			segMin = uint(len(seg.Literal.Runes))
			segMax = uint(len(seg.Literal.Runes))

		case RuneMatchSegment:
			fallthrough
		case QuestionSegment:
			segMin = 1
			segMax = 1

		case StarSegment:
			fallthrough
		case DoubleStarSegment:
			fallthrough
		case DoubleStarSlashSegment:
			segMin = 0
			segMax = UintMax

		default:
			panic(fmt.Errorf("BUG! unknown SegmentType %#v", seg.Type))
		}

		min += segMin
		if max == UintMax || segMax == UintMax {
			max = UintMax
		} else {
			max += segMax
		}

		seg.MinLength = min
		seg.MaxLength = max
	}
	p.MinLength = min
	p.MaxLength = max
}
