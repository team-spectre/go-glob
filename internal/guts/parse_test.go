package guts

import (
	"reflect"
	"testing"
)

type globCompileExpectation func(*testing.T, *Glob)
type RuneMatcherCast func(RuneMatcher) (interface{}, bool)

const (
	emptyString     = ""
	simpleString    = "foo/bar/[0-9][0-9]-?"
	complexString   = "foo/bar/**/[0-9][0-9]-*.[ch]"
)

const (
	digitDense0        = 0x03ff000000000000
	digitDense1        = 0x0000000000000000
	alphanumericDense0 = 0x03ff000000000000
	alphanumericDense1 = 0x07fffffe07fffffe
	chDense0           = 0x0000000000000000
	chDense1           = 0x0000010800000000
)

var (
	digitRange = LoHi{'0', '9'}
	upperRange = LoHi{'A', 'Z'}
	lowerRange = LoHi{'a', 'z'}
	cRange     = LoHi{'c', 'c'}
	hRange     = LoHi{'h', 'h'}
)

var (
	digitRanges        = SortedLoHi{digitRange}
	alphanumericRanges = SortedLoHi{digitRange, upperRange, lowerRange}
	chRanges           = SortedLoHi{cRange, hRange}
)

var (
	digitSet = SetMatch{
		Dense0: digitDense0,
		Dense1: digitDense1,
		Ranges: digitRanges,
	}
	alphanumericSet = SetMatch{
		Dense0: alphanumericDense0,
		Dense1: alphanumericDense1,
		Ranges: alphanumericRanges,
	}
	chSet = SetMatch{
		Dense0: chDense0,
		Dense1: chDense1,
		Ranges: chRanges,
	}
)

func TestCompile(t *testing.T) {
	type testrow struct {
		Name              string
		Pattern           string
		ExpectNumSegments uint
		Expectations      []globCompileExpectation
	}

	testdata := []testrow{
		{
			Name:    "Empty",
			Pattern: emptyString,
		},
		{
			Name:              "Simple",
			Pattern:           simpleString,
			ExpectNumSegments: 5,
			Expectations: []globCompileExpectation{
				expectLiteralSegment(0, "foo/bar/"),
				expectRuneMatchSegment(1, (*RangeMatch)(nil), digitRange, asRange),
				expectRuneMatchSegment(2, (*RangeMatch)(nil), digitRange, asRange),
				expectLiteralSegment(3, "-"),
				expectSpecialSegment(4, QuestionSegment),
			},
		},
		{
			Name:              "Complex",
			Pattern:           complexString,
			ExpectNumSegments: 8,
			Expectations: []globCompileExpectation{
				expectLiteralSegment(0, "foo/bar/"),
				expectSpecialSegment(1, DoubleStarSlashSegment),
				expectRuneMatchSegment(2, (*RangeMatch)(nil), digitRange, asRange),
				expectRuneMatchSegment(3, (*RangeMatch)(nil), digitRange, asRange),
				expectLiteralSegment(4, "-"),
				expectSpecialSegment(5, StarSegment),
				expectLiteralSegment(6, "."),
				expectRuneMatchSegment(7, (*SetMatch)(nil), chSet, asSet),
			},
		},
	}

	for _, row := range testdata {
		t.Run(row.Name, func(t *testing.T) {
			var g Glob
			if err := g.Compile(row.Pattern); err != nil {
				t.Errorf("expected success, got %v", err)
				return
			}
			if n := uint(len(g.Segments)); n != row.ExpectNumSegments {
				t.Errorf("expected %d segments, got %d segments", row.ExpectNumSegments, n)
			}
			for _, expect := range row.Expectations {
				expect(t, &g)
			}
		})
	}
}

func TestCompile_Failure(t *testing.T) {
	type testrow struct {
		Name        string
		Pattern     string
		ErrorString string
	}
	testdata := []testrow{
		{
			Name:        "TripleStar",
			Pattern:     "***",
			ErrorString: "failed to parse glob pattern: \"***\": unexpected '***'",
		},
		{
			Name:        "UnmatchedCloseBracket",
			Pattern:     "]",
			ErrorString: "failed to parse glob pattern: \"]\": unexpected ']'",
		},
		{
			Name:        "UnmatchedOpenBrace",
			Pattern:     "{",
			ErrorString: "failed to parse glob pattern: \"{\": alternative matches are not yet implemented: '{'",
		},
		{
			Name:        "UnmatchedCloseBrace",
			Pattern:     "}",
			ErrorString: "failed to parse glob pattern: \"}\": unexpected '}'",
		},
		{
			Name:        "DoubleOpenBracket1",
			Pattern:     "[[",
			ErrorString: "failed to parse glob pattern: \"[[\": unexpected '['",
		},
		{
			Name:        "DoubleOpenBracket2",
			Pattern:     "[^[",
			ErrorString: "failed to parse glob pattern: \"[^[\": unexpected '['",
		},
		{
			Name:        "DoubleOpenBracket3",
			Pattern:     "[a[",
			ErrorString: "failed to parse glob pattern: \"[a[\": unexpected '['",
		},
		{
			Name:        "DoubleOpenBracket4",
			Pattern:     "[a-[",
			ErrorString: "failed to parse glob pattern: \"[a-[\": unexpected '['",
		},
		{
			Name:        "DoubleOpenBracket5",
			Pattern:     "[a-z[",
			ErrorString: "failed to parse glob pattern: \"[a-z[\": unexpected '['",
		},
		{
			Name:        "UnterminatedCharSet1",
			Pattern:     "[",
			ErrorString: "failed to parse glob pattern: \"[\": unterminated character set",
		},
		{
			Name:        "UnterminatedCharSet2",
			Pattern:     "[a",
			ErrorString: "failed to parse glob pattern: \"[a\": unterminated character set",
		},
		{
			Name:        "UnterminatedCharSet3",
			Pattern:     "[a-",
			ErrorString: "failed to parse glob pattern: \"[a-\": unterminated character set",
		},
		{
			Name:        "UnterminatedCharSet4",
			Pattern:     "[a-z",
			ErrorString: "failed to parse glob pattern: \"[a-z\": unterminated character set",
		},
		{
			Name:        "ReversedCharRange",
			Pattern:     "[z-a]",
			ErrorString: "failed to parse glob pattern: \"[z-a]\": invalid range, lo U+007A > hi U+0061",
		},
		{
			Name:        "UnterminatedBackslashEscape1",
			Pattern:     "\\",
			ErrorString: "failed to parse glob pattern: \"\\\\\": unterminated backslash escape",
		},
		{
			Name:        "UnterminatedBackslashEscape2",
			Pattern:     "\\x",
			ErrorString: "failed to parse glob pattern: \"\\\\x\": unterminated backslash escape",
		},
		{
			Name:        "UnterminatedBackslashEscape3",
			Pattern:     "[\\",
			ErrorString: "failed to parse glob pattern: \"[\\\\\": unterminated backslash escape",
		},
		{
			Name:        "UnterminatedBackslashEscape4",
			Pattern:     "[\\x",
			ErrorString: "failed to parse glob pattern: \"[\\\\x\": unterminated backslash escape",
		},
	}
	for _, row := range testdata {
		t.Run(row.Name, func(t *testing.T) {
			var g Glob
			err := g.Compile(row.Pattern)
			if err == nil {
				t.Errorf("unexpected success: %#v", g)
				return
			}
			if actual := err.Error(); actual != row.ErrorString {
				t.Errorf("unexpected error:\n\texpect: %q\n\tactual: %q", row.ErrorString, actual)
			}
		})
	}
}

func TestCompileRuneMatcher(t *testing.T) {
	type testrow struct {
		Name    string
		Pattern string
		Expect  interface{}
	}
	testdata := []testrow{
		{
			Name:    "Empty",
			Pattern: "",
			Expect:  &NoneMatch{},
		},
		{
			Name:    "Caret",
			Pattern: "^",
			Expect:  &AnyMatch{},
		},
		{
			Name:    "JustA",
			Pattern: "A",
			Expect:  &IsMatch{Rune: 'A'},
		},
		{
			Name:    "NotA",
			Pattern: "^A",
			Expect:  &IsNotMatch{Rune: 'A'},
		},
		{
			Name:    "AZ",
			Pattern: "A-Z",
			Expect:  &RangeMatch{Lo: 'A', Hi: 'Z'},
		},
		{
			Name:    "NotAZ",
			Pattern: "^A-Z",
			Expect:  &ExceptRangeMatch{Lo: 'A', Hi: 'Z'},
		},
		{
			Name:    "Alphanumeric",
			Pattern: "0-9A-Za-z",
			Expect:  &alphanumericSet,
		},
		{
			Name:    "NotAlphanumeric",
			Pattern: "^0-9A-Za-z",
			Expect: &ExceptSetMatch{
				Dense0: alphanumericDense0,
				Dense1: alphanumericDense1,
				Ranges: alphanumericRanges,
			},
		},
		{
			Name:    "ADash",
			Pattern: "a-",
			Expect: &SetMatch{
				Dense0: 0x0000200000000000,
				Dense1: 0x0000000200000000,
				Ranges: []LoHi{
					{'-', '-'},
					{'a', 'a'},
				},
			},
		},
	}
	for _, row := range testdata {
		t.Run(row.Name, func(t *testing.T) {
			m, err := CompileRuneMatcher(row.Pattern)
			if err != nil {
				t.Errorf("expected success, got %v", err)
				return
			}
			if m == nil {
				t.Error("expected non-nil, got nil")
				return
			}
			if !reflect.DeepEqual(m, row.Expect) {
				t.Errorf("expected %#v, got %#v", row.Expect, m)
			}
		})
	}
}

func TestCompileRuneMatcher_Failure(t *testing.T) {
	type testrow struct {
		Name        string
		Pattern     string
		ErrorString string
	}
	testdata := []testrow{
		{
			Name:        "UnexpectedOpenBracket",
			Pattern:     "[",
			ErrorString: "failed to parse character set: \"[\": unexpected '['",
		},
		{
			Name:        "UnexpectedCloseBracket",
			Pattern:     "]",
			ErrorString: "failed to parse character set: \"]\": unexpected ']'",
		},
	}
	for _, row := range testdata {
		t.Run(row.Name, func(t *testing.T) {
			m, err := CompileRuneMatcher(row.Pattern)
			if err == nil {
				t.Errorf("unexpected success: %#v", m)
				return
			}
			if m != nil {
				t.Errorf("unexpected value: %#v", m)
			}
			if actual := err.Error(); actual != row.ErrorString {
				t.Errorf("unexpected error:\n\texpect: %q\n\tactual: %q", row.ErrorString, actual)
			}
		})
	}
}

func expectLiteralSegment(index uint, expect string) globCompileExpectation {
	return func(t *testing.T, g *Glob) {
		if index < uint(len(g.Segments)) {
			seg := g.Segments[index]
			if seg.Type != LiteralSegment {
				t.Errorf("Glob.Segments[%d].type: expected %#v, got %#v", index, LiteralSegment, seg.Type)
				return
			}
			actual := seg.Literal.String
			if actual != expect {
				t.Errorf("Glob.Segments[%d].runes: expected %q, got %q", index, expect, actual)
			}
			if seg.Matcher != nil {
				t.Errorf("Glob.Segments[%d].Matcher: expected nil, got %T", index, seg.Matcher)
			}
		}
	}
}

func expectRuneMatchSegment(index uint, prototype RuneMatcher, expect interface{}, cast RuneMatcherCast) globCompileExpectation {
	return func(t *testing.T, g *Glob) {
		if index < uint(len(g.Segments)) {
			seg := g.Segments[index]
			if seg.Type != RuneMatchSegment {
				t.Errorf("Glob.Segments[%d].type: expected %#v, got %#v", index, RuneMatchSegment, seg.Type)
				return
			}
			if seg.Literal.Runes != nil {
				str := seg.Literal.String
				t.Errorf("Glob.Segments[%d].runes: expected nil, got %q => len %d", index, str, len(seg.Literal.Runes))
			}
			if seg.Matcher == nil {
				t.Errorf("Glob.Segments[%d].Matcher: expected %T, got nil", index, prototype)
				return
			}
			actual, ok := cast(seg.Matcher)
			if !ok {
				t.Errorf("Glob.Segments[%d].Matcher: expected %T, got %T", index, prototype, seg.Matcher)
				return
			}
			if !reflect.DeepEqual(actual, expect) {
				t.Errorf("Glob.Segments[%d].Matcher: expected %#v, got %#v", index, expect, actual)
			}
		}
	}
}

func expectSpecialSegment(index uint, expect SegmentType) globCompileExpectation {
	return func(t *testing.T, g *Glob) {
		if index < uint(len(g.Segments)) {
			seg := g.Segments[index]
			if seg.Type != expect {
				t.Errorf("Glob.Segments[%d].type: expected %#v, got %#v", index, expect, seg.Type)
				return
			}
			if seg.Literal.Runes != nil {
				str := seg.Literal.String
				t.Errorf("Glob.Segments[%d].runes: expected nil, got %q => len %d", index, str, len(seg.Literal.Runes))
			}
			if seg.Matcher != nil {
				t.Errorf("Glob.Segments[%d].Matcher: expected nil, got %T", index, seg.Matcher)
			}
		}
	}
}

func asRange(matcher RuneMatcher) (interface{}, bool) {
	v, ok := matcher.(*RangeMatch)
	var w LoHi
	if ok {
		w = LoHi{v.Lo, v.Hi}
	}
	return w, ok
}

func asSet(matcher RuneMatcher) (interface{}, bool) {
	v, ok := matcher.(*SetMatch)
	return *v, ok
}
