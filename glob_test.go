package glob

import (
	"testing"
)

func TestGlob_Match(t *testing.T) {
	emptyGlob := Glob{
		segments:  nil,
		minLength: 0,
		maxLength: 0,
	}
	emptyGlob.setPattern(emptyString)

	simpleGlob := Glob{
		segments: []segment{
			{
				stype:     literalSegment,
				minLength: 12,
				maxLength: 12,
			},
			{
				stype:     runeMatchSegment,
				matcher:   &digitSet,
				minLength: 4,
				maxLength: 4,
			},
			{
				stype:     runeMatchSegment,
				matcher:   &digitSet,
				minLength: 3,
				maxLength: 3,
			},
			{
				stype:     literalSegment,
				minLength: 2,
				maxLength: 2,
			},
			{
				stype:     questionSegment,
				minLength: 1,
				maxLength: 1,
			},
		},
		minLength: 12,
		maxLength: 12,
	}
	simpleGlob.setPattern(simpleString)
	simpleGlob.segments[0].setLiteral("foo/bar/")
	simpleGlob.segments[3].setLiteral("-")

	complexGlob := Glob{
		segments: []segment{
			{
				stype:     literalSegment,
				minLength: 13,
				maxLength: uintMax,
			},
			{
				stype:     doubleStarSlashSegment,
				minLength: 5,
				maxLength: uintMax,
			},
			{
				stype:     runeMatchSegment,
				matcher:   &digitSet,
				minLength: 5,
				maxLength: uintMax,
			},
			{
				stype:     runeMatchSegment,
				matcher:   &digitSet,
				minLength: 4,
				maxLength: uintMax,
			},
			{
				stype:     literalSegment,
				minLength: 3,
				maxLength: uintMax,
			},
			{
				stype:     starSegment,
				minLength: 2,
				maxLength: uintMax,
			},
			{
				stype:     literalSegment,
				minLength: 2,
				maxLength: 2,
			},
			{
				stype:     runeMatchSegment,
				matcher:   &chSet,
				minLength: 1,
				maxLength: 1,
			},
		},
		minLength: 13,
		maxLength: uintMax,
	}
	complexGlob.setPattern(complexString)
	complexGlob.segments[0].setLiteral("foo/bar/")
	complexGlob.segments[4].setLiteral("-")
	complexGlob.segments[6].setLiteral(".")

	type testrow struct {
		Name           string
		G              *Glob
		ExpectString   string
		ExpectGoString string
		ExpectAccept   []string
		ExpectReject   []string
	}

	testdata := []testrow{
		{
			Name:           "Nil",
			G:              nil,
			ExpectString:   emptyString,
			ExpectGoString: emptyGoString,
			ExpectAccept:   []string{""},
			ExpectReject:   []string{"a"},
		},
		{
			Name:           "Empty",
			G:              &emptyGlob,
			ExpectString:   emptyString,
			ExpectGoString: emptyGoString,
			ExpectAccept:   []string{""},
			ExpectReject:   []string{"a"},
		},
		{
			Name:           "Simple",
			G:              &simpleGlob,
			ExpectString:   simpleString,
			ExpectGoString: simpleGoString,
			ExpectAccept: []string{
				"foo/bar/00-x",
				"foo/bar/99-y",
			},
			ExpectReject: []string{
				"",
				"foo",
				"foo/bar/",
				"foo/bar/xx-x",
				"foo/bar/00-/",
			},
		},
		{
			Name:           "Complex",
			G:              &complexGlob,
			ExpectString:   complexString,
			ExpectGoString: complexGoString,
			ExpectAccept: []string{
				"foo/bar/baz/00-x.c",
				"foo/bar/baz/99-y.c",
				"foo/bar/baz/55-z.h",
				"foo/bar/42-A.h",
				"foo/bar/baz/42-A.h",
				"foo/bar/baz/quux/42-A.h",
			},
			ExpectReject: []string{
				"",
				"foo",
				"foo/bar/",
				"foo/bar/baz/xx-x.c",
				"foo/bar/baz/55-/.h",
			},
		},
	}

	for _, row := range testdata {
		t.Run(row.Name, func(t *testing.T) {
			if row.G.String() != row.ExpectString {
				t.Errorf("String: expected %q, got %q", row.ExpectString, row.G.String())
			}
			if row.G.GoString() != row.ExpectGoString {
				t.Errorf("GoString: expected %q, got %q", row.ExpectGoString, row.G.GoString())
			}
			for _, input := range row.ExpectReject {
				if row.G.Matcher(input).Matches() {
					t.Errorf("Match %q: unexpected acceptance", input)
				}
			}
			for _, input := range row.ExpectAccept {
				if !row.G.Matcher(input).Matches() {
					t.Errorf("Match %q: unexpected rejection", input)
				}
			}
		})
	}
}
