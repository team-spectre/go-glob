package glob

import (
	"reflect"
	"testing"
)

const (
	emptyPattern    = ""
	emptyGoString   = "nil"
	simplePattern   = "foo/bar/[0-9][0-9]-?"
	simpleGoString  = "glob.MustCompile(\"foo/bar/[0-9][0-9]-?\")"
	complexPattern  = "foo/bar/**/[0-9][0-9]-*.[ch]"
	complexGoString = "glob.MustCompile(\"foo/bar/**/[0-9][0-9]-*.[ch]\")"
)

const (
	digitDense0 = 0x03ff000000000000
	digitDense1 = 0x0000000000000000
	chDense0    = 0x0000000000000000
	chDense1    = 0x0000010800000000
)

var (
	digitRange = runeMatchRange{'0', '9'}
	cRange     = runeMatchRange{'c', 'c'}
	hRange     = runeMatchRange{'h', 'h'}
)

var (
	digitSet = runeMatchSet{
		Dense0: digitDense0,
		Dense1: digitDense1,
		Ranges: []runeMatchRange{digitRange},
	}
	chSet = runeMatchSet{
		Dense0: chDense0,
		Dense1: chDense1,
		Ranges: []runeMatchRange{cRange, hRange},
	}
)

func TestGlob_Compile_Empty(t *testing.T) {
	g, err := Compile(emptyPattern)
	if err != nil {
		t.Errorf("failed to parse %q: %v", emptyPattern, err)
	}
	if err == nil && g != nil {
		t.Errorf("Compile: unexpected non-nil value %#v", *g)
	}
	g = MustCompile(emptyPattern)
	if g != nil {
		t.Errorf("Compile: unexpected non-nil value %#v", *g)
	}
}

func TestGlob_Compile_Simple(t *testing.T) {
	g, err := Compile(simplePattern)
	if err != nil {
		t.Errorf("failed to parse %q: %v", simplePattern, err)
		return
	}
	if g.pattern != simplePattern {
		t.Errorf("Glob.pattern: expected %q, got %q", simplePattern, g.pattern)
	}
	if len(g.segments) != 5 {
		t.Errorf("Glob.segments: expected len %d, got len %d", 5, len(g.segments))
	}
	expectLiteralSegment(t, g, 0, "foo/bar/")
	expectRuneMatchSegment(t, g, 1, (*runeMatchRange)(nil), digitRange, asRange)
	expectRuneMatchSegment(t, g, 2, (*runeMatchRange)(nil), digitRange, asRange)
	expectLiteralSegment(t, g, 3, "-")
	expectSpecialSegment(t, g, 4, questionSegment)
}

func TestGlob_Compile_Complex(t *testing.T) {
	g, err := Compile(complexPattern)
	if err != nil {
		t.Errorf("failed to parse %q: %v", complexPattern, err)
		return
	}
	if g.pattern != complexPattern {
		t.Errorf("Glob.pattern: expected %q, got %q", complexPattern, g.pattern)
	}
	if len(g.segments) != 8 {
		t.Errorf("Glob.segments: expected len %d, got len %d", 8, len(g.segments))
	}
	expectLiteralSegment(t, g, 0, "foo/bar/")
	expectSpecialSegment(t, g, 1, doubleStarSlashSegment)
	expectRuneMatchSegment(t, g, 2, (*runeMatchRange)(nil), digitRange, asRange)
	expectRuneMatchSegment(t, g, 3, (*runeMatchRange)(nil), digitRange, asRange)
	expectLiteralSegment(t, g, 4, "-")
	expectSpecialSegment(t, g, 5, starSegment)
	expectLiteralSegment(t, g, 6, ".")
	expectRuneMatchSegment(t, g, 7, (*runeMatchSet)(nil), chSet, asSet)
}

func TestGlob_Compile_Failure(t *testing.T) {
	type testrow struct {
		Name        string
		Pattern     string
		ErrorString string
	}
	testdata := []testrow{
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
			Name:        "DoubleOpenBracket2",
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

func expectLiteralSegment(t *testing.T, g *Glob, index uint, expect string) {
	if index < uint(len(g.segments)) {
		seg := g.segments[index]
		if seg.stype != literalSegment {
			t.Errorf("Glob.segments[%d].type: expected %#v, got %#v", index, literalSegment, seg.stype)
			return
		}
		actual := runesToString(seg.runes)
		if actual != expect {
			t.Errorf("Glob.segments[%d].runes: expected %q, got %q", index, expect, actual)
		}
		if seg.matcher != nil {
			t.Errorf("Glob.segments[%d].matcher: expected nil, got %T", index, seg.matcher)
		}
	}
}

func expectRuneMatchSegment(t *testing.T, g *Glob, index uint, prototype RuneMatcher, expect interface{}, cast func(RuneMatcher) (interface{}, bool)) {
	if index < uint(len(g.segments)) {
		seg := g.segments[index]
		if seg.stype != runeMatchSegment {
			t.Errorf("Glob.segments[%d].type: expected %#v, got %#v", index, runeMatchSegment, seg.stype)
			return
		}
		if seg.runes != nil {
			str := runesToString(seg.runes)
			t.Errorf("Glob.segments[%d].runes: expected nil, got %q => len %d", index, str, len(seg.runes))
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

func expectSpecialSegment(t *testing.T, g *Glob, index uint, expect segmentType) {
	if index < uint(len(g.segments)) {
		seg := g.segments[index]
		if seg.stype != expect {
			t.Errorf("Glob.segments[%d].type: expected %#v, got %#v", index, expect, seg.stype)
			return
		}
		if seg.runes != nil {
			str := runesToString(seg.runes)
			t.Errorf("Glob.segments[%d].runes: expected nil, got %q => len %d", index, str, len(seg.runes))
		}
		if seg.matcher != nil {
			t.Errorf("Glob.segments[%d].matcher: expected nil, got %T", index, seg.matcher)
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
