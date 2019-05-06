package guts

import (
	"fmt"
)

func (g *Glob) Compile(input string) error {
	*g = Glob{}

	var p Parser
	p.Input = Norm(input)
	p.InputJ = uint(len(p.Input.Runes))
	p.Segments = make([]Segment, 0, 16)
	p.State = RootState
	p.WantSet = false
	p.Run()

	if p.Err != nil {
		return fmt.Errorf("failed to parse glob pattern: %q: %v", p.Input.String, p.Err)
	}

	g.Pattern = p.Input
	g.Segments = p.Segments
	g.MinLength = p.MinLength
	g.MaxLength = p.MaxLength
	return nil
}

func (g *Glob) Matcher(out *Matcher, input string) {
	*out = Matcher{}
	out.Memo = make(MemoMap)
	out.Input = Norm(input)
	out.InputJ = uint(len(out.Input.Runes))
	out.SegmentJ = uint(len(g.Segments))
	minLength := g.MinLength
	maxLength := g.MaxLength

	// Fast reject the input is too short or too long to ever match;
	// (*Matcher)(nil) is a valid matcher that will never match any string.
	out.Valid = (out.InputJ >= minLength && out.InputJ <= maxLength)
}
