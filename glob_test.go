package glob

import (
	"testing"
)

func TestGlob_Match(t *testing.T) {
	const (
		emptyString   = ""
		simpleString  = "foo/bar/[0-9][0-9]-?"
		complexString = "foo/bar/**/[0-9][0-9]-?*.[ch]"
	)

	const (
		emptyGoString   = `glob.MustCompile("")`
		simpleGoString  = `glob.MustCompile("foo/bar/[0-9][0-9]-?")`
		complexGoString = `glob.MustCompile("foo/bar/**/[0-9][0-9]-?*.[ch]")`
	)

	emptyGlob := MustCompile(emptyString)
	simpleGlob := MustCompile(simpleString)
	complexGlob := MustCompile(complexString)

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
			Name:           "Empty",
			G:              emptyGlob,
			ExpectString:   emptyString,
			ExpectGoString: emptyGoString,
			ExpectAccept:   []string{""},
			ExpectReject:   []string{"a"},
		},
		{
			Name:           "Simple",
			G:              simpleGlob,
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
			G:              complexGlob,
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
