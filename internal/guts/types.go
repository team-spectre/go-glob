package guts

import (
	"fmt"
	"sort"
)

type ExplodedString struct {
	String string
	Runes  []rune
	Map    []uint
}

type LoHi struct{ Lo, Hi rune }
type SortedLoHi []LoHi

type RuneMatcher interface {
	MatchRune(rune) bool
	ForEachRange(func(lo, hi rune))
	Not() RuneMatcher
}

type AnyMatch struct{}
type NoneMatch struct{}
type IsMatch struct{ Rune rune }
type IsNotMatch struct{ Rune rune }
type RangeMatch struct{ Lo, Hi rune }
type ExceptRangeMatch struct{ Lo, Hi rune }
type SetMatch struct {
	Dense0 uint64
	Dense1 uint64
	Ranges SortedLoHi
}
type ExceptSetMatch struct {
	Dense0 uint64
	Dense1 uint64
	Ranges SortedLoHi
}

type MemoMap map[MemoKey]*MemoValue
type MemoKey struct{ InputI, SegmentI uint }
type MemoValue struct {
	Checked  bool
	Rejected bool
	Index    uint
}

type IndexSet map[uint]bool

type SegmentType byte

type Segment struct {
	Type      SegmentType
	Literal   ExplodedString
	Matcher   RuneMatcher
	PatternP  uint
	PatternQ  uint
	MinLength uint
	MaxLength uint
}

type Glob struct {
	Pattern   ExplodedString
	Segments  []Segment
	MinLength uint
	MaxLength uint
}

type Matcher struct {
	Memo     MemoMap
	Input    ExplodedString
	C        Capture
	InputI   uint
	InputJ   uint
	SegmentI uint
	SegmentJ uint
	Valid    bool
}

type Capture struct {
	InputP   uint
	InputQ   uint
	SegmentP uint
	PatternP uint
	PatternQ uint
}

type ParseState byte
type Parser struct {
	Input            ExplodedString
	Segments         []Segment
	Ranges           []LoHi
	PartialLiteral   []rune
	PartialEscape    []rune
	LastSegment      *Segment
	Err              error
	InputP           uint
	InputQ           uint
	InputI           uint
	InputJ           uint
	MinLength        uint
	MaxLength        uint
	EscapeIntroducer rune
	State            ParseState
	EscapeLen        byte
	Negate           bool
	WantSet          bool
}

var _ RuneMatcher = (*AnyMatch)(nil)
var _ RuneMatcher = (*NoneMatch)(nil)
var _ RuneMatcher = (*IsMatch)(nil)
var _ RuneMatcher = (*IsNotMatch)(nil)
var _ RuneMatcher = (*RangeMatch)(nil)
var _ RuneMatcher = (*ExceptRangeMatch)(nil)
var _ RuneMatcher = (*SetMatch)(nil)
var _ RuneMatcher = (*ExceptSetMatch)(nil)
var _ sort.Interface = SortedLoHi(nil)
var _ fmt.Stringer = SegmentType(0)
var _ fmt.Stringer = ParseState(0)
var _ fmt.GoStringer = SegmentType(0)
var _ fmt.GoStringer = ParseState(0)
