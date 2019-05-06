package glob

import (
	"reflect"
	"testing"
)

type globCompileExpectation func(*testing.T, *Glob)
type runeMatcherCast func(RuneMatcher) (interface{}, bool)

const (
	emptyString     = ""
	emptyGoString   = "nil"
	simpleString    = "foo/bar/[0-9][0-9]-?"
	simpleGoString  = "glob.MustCompile(\"foo/bar/[0-9][0-9]-?\")"
	complexString   = "foo/bar/**/[0-9][0-9]-*.[ch]"
	complexGoString = "glob.MustCompile(\"foo/bar/**/[0-9][0-9]-*.[ch]\")"
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
	emptyRunes   = []rune(emptyString)
	simpleRunes  = []rune(simpleString)
	complexRunes = []rune(complexString)
)

var (
	digitRange = runeMatchRange{'0', '9'}
	upperRange = runeMatchRange{'A', 'Z'}
	lowerRange = runeMatchRange{'a', 'z'}
	cRange     = runeMatchRange{'c', 'c'}
	hRange     = runeMatchRange{'h', 'h'}
)

var (
	digitRanges        = []runeMatchRange{digitRange}
	alphanumericRanges = []runeMatchRange{digitRange, upperRange, lowerRange}
	chRanges           = []runeMatchRange{cRange, hRange}
)

var (
	digitSet = runeMatchSet{
		Dense0: digitDense0,
		Dense1: digitDense1,
		Ranges: digitRanges,
	}
	alphanumericSet = runeMatchSet{
		Dense0: alphanumericDense0,
		Dense1: alphanumericDense1,
		Ranges: alphanumericRanges,
	}
	chSet = runeMatchSet{
		Dense0: chDense0,
		Dense1: chDense1,
		Ranges: chRanges,
	}
)

func TestCompile(t *testing.T) {
	type testrow struct {
		Name              string
		Pattern           string
		ExpectNil         bool
		ExpectNumSegments uint
		Expectations      []globCompileExpectation
	}
	testdata := []testrow{
		{
			Name:      "Empty",
			Pattern:   emptyString,
			ExpectNil: true,
		},
		{
			Name:              "Simple",
			Pattern:           simpleString,
			ExpectNumSegments: 5,
			Expectations: []globCompileExpectation{
				expectLiteralSegment(0, "foo/bar/"),
				expectRuneMatchSegment(1, (*runeMatchRange)(nil), digitRange, asRange),
				expectRuneMatchSegment(2, (*runeMatchRange)(nil), digitRange, asRange),
				expectLiteralSegment(3, "-"),
				expectSpecialSegment(4, questionSegment),
			},
		},
		{
			Name:              "Complex",
			Pattern:           complexString,
			ExpectNumSegments: 8,
			Expectations: []globCompileExpectation{
				expectLiteralSegment(0, "foo/bar/"),
				expectSpecialSegment(1, doubleStarSlashSegment),
				expectRuneMatchSegment(2, (*runeMatchRange)(nil), digitRange, asRange),
				expectRuneMatchSegment(3, (*runeMatchRange)(nil), digitRange, asRange),
				expectLiteralSegment(4, "-"),
				expectSpecialSegment(5, starSegment),
				expectLiteralSegment(6, "."),
				expectRuneMatchSegment(7, (*runeMatchSet)(nil), chSet, asSet),
			},
		},
	}
	for _, row := range testdata {
		t.Run(row.Name, func(t *testing.T) {
			g, err := Compile(row.Pattern)
			if err != nil {
				t.Errorf("expected success, got %v", err)
				return
			}
			if g == nil {
				if !row.ExpectNil {
					t.Error("expected non-nil, got nil")
				}
				return
			}
			if row.ExpectNil {
				t.Error("expected nil, got non-nil")
				return
			}
			if n := uint(len(g.segments)); n != row.ExpectNumSegments {
				t.Errorf("expected %d segments, got %d segments", row.ExpectNumSegments, n)
			}
			for _, expect := range row.Expectations {
				expect(t, g)
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
			g, err := Compile(row.Pattern)
			if err == nil && g != nil {
				t.Errorf("unexpected success: %#v", *g)
				return
			}
			if err == nil {
				t.Error("unexpected success: nil")
				return
			}
			if g != nil {
				t.Errorf("unexpected value: %#v", *g)
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
			Expect:  &runeMatchNotAny{},
		},
		{
			Name:    "Caret",
			Pattern: "^",
			Expect:  &runeMatchAny{},
		},
		{
			Name:    "JustA",
			Pattern: "A",
			Expect:  &runeMatchIs{Rune: 'A'},
		},
		{
			Name:    "NotA",
			Pattern: "^A",
			Expect:  &runeMatchNotIs{Rune: 'A'},
		},
		{
			Name:    "AZ",
			Pattern: "A-Z",
			Expect:  &upperRange,
		},
		{
			Name:    "NotAZ",
			Pattern: "^A-Z",
			Expect:  &runeMatchNotRange{Lo: 'A', Hi: 'Z'},
		},
		{
			Name:    "Alphanumeric",
			Pattern: "0-9A-Za-z",
			Expect:  &alphanumericSet,
		},
		{
			Name:    "NotAlphanumeric",
			Pattern: "^0-9A-Za-z",
			Expect: &runeMatchNotSet{
				Dense0: alphanumericDense0,
				Dense1: alphanumericDense1,
				Ranges: alphanumericRanges,
			},
		},
		{
			Name:    "ADash",
			Pattern: "a-",
			Expect: &runeMatchSet{
				Dense0: 0x0000200000000000,
				Dense1: 0x0000000200000000,
				Ranges: []runeMatchRange{
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

func TestMustCompile(t *testing.T) {
	_ = MustCompile(emptyString)
}

func TestMustCompileRuneMatcher(t *testing.T) {
	_ = MustCompileRuneMatcher("")
}

func expectLiteralSegment(index uint, expect string) globCompileExpectation {
	return func(t *testing.T, g *Glob) {
		if index < uint(len(g.segments)) {
			seg := g.segments[index]
			if seg.stype != literalSegment {
				t.Errorf("Glob.segments[%d].type: expected %#v, got %#v", index, literalSegment, seg.stype)
				return
			}
			actual := seg.literalString
			if actual != expect {
				t.Errorf("Glob.segments[%d].runes: expected %q, got %q", index, expect, actual)
			}
			if seg.matcher != nil {
				t.Errorf("Glob.segments[%d].matcher: expected nil, got %T", index, seg.matcher)
			}
		}
	}
}

func expectRuneMatchSegment(index uint, prototype RuneMatcher, expect interface{}, cast runeMatcherCast) globCompileExpectation {
	return func(t *testing.T, g *Glob) {
		if index < uint(len(g.segments)) {
			seg := g.segments[index]
			if seg.stype != runeMatchSegment {
				t.Errorf("Glob.segments[%d].type: expected %#v, got %#v", index, runeMatchSegment, seg.stype)
				return
			}
			if seg.literalRunes != nil {
				str := seg.literalString
				t.Errorf("Glob.segments[%d].runes: expected nil, got %q => len %d", index, str, len(seg.literalRunes))
			}
			if seg.matcher == nil {
				t.Errorf("Glob.segments[%d].matcher: expected %T, got nil", index, prototype)
				return
			}
			actual, ok := cast(seg.matcher)
			if !ok {
				t.Errorf("Glob.segments[%d].matcher: expected %T, got %T", index, prototype, seg.matcher)
				return
			}
			if !reflect.DeepEqual(actual, expect) {
				t.Errorf("Glob.segments[%d].matcher: expected %#v, got %#v", index, expect, actual)
			}
		}
	}
}

func expectSpecialSegment(index uint, expect segmentType) globCompileExpectation {
	return func(t *testing.T, g *Glob) {
		if index < uint(len(g.segments)) {
			seg := g.segments[index]
			if seg.stype != expect {
				t.Errorf("Glob.segments[%d].type: expected %#v, got %#v", index, expect, seg.stype)
				return
			}
			if seg.literalRunes != nil {
				str := seg.literalString
				t.Errorf("Glob.segments[%d].runes: expected nil, got %q => len %d", index, str, len(seg.literalRunes))
			}
			if seg.matcher != nil {
				t.Errorf("Glob.segments[%d].matcher: expected nil, got %T", index, seg.matcher)
			}
		}
	}
}

func asRange(matcher RuneMatcher) (interface{}, bool) {
	v, ok := matcher.(*runeMatchRange)
	return *v, ok
}

func asSet(matcher RuneMatcher) (interface{}, bool) {
	v, ok := matcher.(*runeMatchSet)
	return *v, ok
}
