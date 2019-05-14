package glob

import (
	"fmt"

	"github.com/team-spectre/go-glob/internal/guts"
)

// Glob represents a compiled glob pattern, ready to match path names.
type Glob struct {
	impl guts.Glob
}

func Compile(input string) (*Glob, error) {
	g := new(Glob)
	if err := g.impl.Compile(input); err != nil {
		return nil, fmt.Errorf("failed to parse glob pattern: %q: %v", input, err)
	}
	return g, nil
}

func MustCompile(input string) *Glob {
	compiled, err := Compile(input)
	if err != nil {
		panic(err)
	}
	return compiled
}

func (g *Glob) Matcher(input string) *Matcher {
	m := new(Matcher)
	m.g = &g.impl
	g.impl.Matcher(&m.impl, input)
	return m
}

func (g *Glob) Pattern() string {
	return g.impl.Pattern.String
}

func (g *Glob) PatternSubstring(i, j uint) string {
	return g.impl.Pattern.Substring(i, j)
}

func (g *Glob) String() string {
	return g.Pattern()
}

func (g *Glob) GoString() string {
	return fmt.Sprintf("glob.MustCompile(%q)", g.Pattern())
}

var _ fmt.Stringer = (*Glob)(nil)
var _ fmt.GoStringer = (*Glob)(nil)

type Matcher struct {
	impl guts.Matcher
	g    *guts.Glob
}

func (m *Matcher) Input() string {
	return m.impl.Input.String
}

func (m *Matcher) InputSubstring(i, j uint) string {
	return m.impl.Input.Substring(i, j)
}

func (m *Matcher) HasNext() bool {
	return m.impl.HasNext(m.g)
}

func (m *Matcher) Capture() *Capture {
	c := new(Capture)
	c.impl = m.impl.Capture()
	c.m = &m.impl
	c.g = m.g
	return c
}

func (m *Matcher) OK() bool {
	return m.impl.Valid
}

func (m *Matcher) Matches() bool {
	for m.HasNext() {
	}
	return m.OK()
}

type Capture struct {
	impl *guts.Capture
	m    *guts.Matcher
	g    *guts.Glob
}

func (c *Capture) PatternLocation() (uint, uint) {
	return c.impl.PatternP, c.impl.PatternQ
}

func (c *Capture) PatternStart() uint {
	return c.impl.PatternP
}

func (c *Capture) PatternEnd() uint {
	return c.impl.PatternQ
}

func (c *Capture) Pattern() string {
	return c.g.Pattern.Substring(c.impl.PatternP, c.impl.PatternQ)
}

func (c *Capture) InputLocation() (uint, uint) {
	return c.impl.InputP, c.impl.InputQ
}

func (c *Capture) InputStart() uint {
	return c.impl.InputP
}

func (c *Capture) InputEnd() uint {
	return c.impl.InputQ
}

func (c *Capture) Input() string {
	return c.m.Input.Substring(c.impl.InputP, c.impl.InputQ)
}
