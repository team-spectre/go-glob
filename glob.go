package glob

import (
	"fmt"
)

// Glob represents a compiled glob pattern, ready to match path names.
type Glob struct {
	patternString string
	patternRunes  []rune
	patternIndex  []uint
	segments      []segment
	minLength     uint
	maxLength     uint
}

type segment struct {
	stype         segmentType
	literalString string
	literalRunes  []rune
	literalIndex  []uint
	matcher       RuneMatcher
	patternP      uint
	patternQ      uint
	minLength     uint
	maxLength     uint
}

func (g *Glob) Pattern() string {
	if g == nil || len(g.segments) <= 0 {
		return ""
	}
	return g.patternString
}

func (g *Glob) PatternSubstring(i, j uint) string {
	bi := g.patternIndex[i]
	bj := g.patternIndex[j]
	return g.patternString[bi:bj]
}

func (g *Glob) String() string {
	return g.Pattern()
}

func (g *Glob) GoString() string {
	if g == nil || len(g.segments) <= 0 {
		return "nil"
	}
	return fmt.Sprintf("glob.MustCompile(%q)", g.patternString)
}

var _ fmt.Stringer = (*Glob)(nil)
var _ fmt.GoStringer = (*Glob)(nil)

func MustCompile(input string) *Glob {
	compiled, err := Compile(input)
	if err != nil {
		panic(err)
	}
	return compiled
}

func Compile(input string) (*Glob, error) {
	if input == "" {
		return nil, nil
	}

	var p parser
	p.setInput(input)
	p.segments = make([]segment, 0, 16)
	p.state = rootState
	p.wantSet = false

	p.run()

	if p.err != nil {
		return nil, fmt.Errorf("failed to parse glob pattern: %q: %v", p.inputString, p.err)
	}

	g := &Glob{
		patternString: p.inputString,
		patternRunes:  p.inputRunes,
		patternIndex:  p.inputIndex,
		segments:      p.segments,
		minLength:     p.minLength,
		maxLength:     p.maxLength,
	}
	return g, nil
}

func (g *Glob) setPattern(str string) {
	g.patternString, g.patternRunes, g.patternIndex = normString(str)
}

func (seg *segment) setLiteral(str string) {
	seg.literalString, seg.literalRunes, seg.literalIndex = normString(str)
}
