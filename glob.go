package glob

import (
	"fmt"
)

type Glob struct {
	pattern   string
	segments  []segment
	minLength uint
	maxLength uint
}

type segment struct {
	stype     segmentType
	runes     []rune
	matcher   RuneMatcher
	minLength uint
	maxLength uint
}

func (g *Glob) Matcher(input string) *Matcher {
	return g.RunesMatcher(stringToRunes(input))
}

func (g *Glob) BytesMatcher(input []byte) *Matcher {
	return g.RunesMatcher(bytesToRunes(input))
}

func (g *Glob) Match(input string) bool {
	return g.Matcher(input).Matches()
}

func (g *Glob) ByteMatch(input []byte) bool {
	return g.BytesMatcher(input).Matches()
}

func (g *Glob) RuneMatch(input []rune) bool {
	return g.RunesMatcher(input).Matches()
}

func (g *Glob) String() string {
	if g == nil || len(g.segments) == 0 {
		return ""
	}
	return g.pattern
}

func (g *Glob) GoString() string {
	if g == nil || len(g.segments) == 0 {
		return "nil"
	}
	return fmt.Sprintf("glob.MustCompile(%q)", g.pattern)
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
	p.runes = stringToRunes(input)
	p.segments = make([]segment, 0, 16)
	p.text = input
	p.i = 0
	p.j = uint(len(p.runes))
	p.state = rootState
	p.lastSegmentType = literalSegment
	p.wantSet = false

	p.run()

	if p.err != nil {
		return nil, fmt.Errorf("failed to parse glob pattern: %q: %v", input, p.err)
	}

	g := &Glob{
		pattern:   input,
		segments:  p.segments,
		minLength: p.minLength,
		maxLength: p.maxLength,
	}
	return g, nil
}
