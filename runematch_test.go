package glob

import (
	"testing"
	"unicode"
)

func TestMatchers(t *testing.T) {
	type testrow struct {
		Name           string
		New            func() RuneMatcher
		ExpectString   string
		ExpectGoString string
		ExpectRanges   []runeMatchRange
		ExpectAccept   []rune
		ExpectReject   []rune
	}
	testdata := []testrow{
		{
			Name:           "None",
			New:            None,
			ExpectString:   "",
			ExpectGoString: "glob.None()",
			ExpectRanges:   nil,
			ExpectAccept:   nil,
			ExpectReject:   []rune{0, '0', 'A', 'a', unicode.MaxRune},
		},
		{
			Name:           "Any",
			New:            Any,
			ExpectString:   "^",
			ExpectGoString: "glob.Any()",
			ExpectRanges:   []runeMatchRange{{0, unicode.MaxRune}},
			ExpectAccept:   []rune{0, '0', 'A', 'a', unicode.MaxRune},
		},
		{
			Name:           "Range [0-9]",
			New:            func() RuneMatcher { return Range('0', '9') },
			ExpectString:   "0-9",
			ExpectGoString: "glob.Range('0', '9')",
			ExpectRanges:   []runeMatchRange{{'0', '9'}},
			ExpectAccept:   []rune{'0', '9'},
			ExpectReject:   []rune{0, '/', ':', 'A', 'a', unicode.MaxRune},
		},
		{
			Name:           "Range [^0-9]",
			New:            func() RuneMatcher { return Range('0', '9').Not() },
			ExpectString:   "^0-9",
			ExpectGoString: "glob.Range('0', '9').Not()",
			ExpectRanges:   []runeMatchRange{{0, '/'}, {':', unicode.MaxRune}},
			ExpectAccept:   []rune{0, '/', ':', 'A', 'a', unicode.MaxRune},
			ExpectReject:   []rune{'0', '9'},
		},
	}
	for _, row := range testdata {
		t.Run(row.Name, func(t *testing.T) {
			m := row.New()

			actualString := m.String()
			if actualString != row.ExpectString {
				t.Errorf("String: expected %q, got %q", row.ExpectString, actualString)
			}

			actualGoString := m.GoString()
			if actualGoString != row.ExpectGoString {
				t.Errorf("GoString: expected %q, got %q", row.ExpectGoString, actualGoString)
			}

			actualRanges := make([]runeMatchRange, 0, len(row.ExpectRanges))
			m.ForEachRange(func(lo, hi rune) {
				actualRanges = append(actualRanges, runeMatchRange{lo, hi})
			})
			testEqualRanges(t, actualRanges, row.ExpectRanges)

			for _, ch := range row.ExpectAccept {
				if !m.MatchRune(ch) {
					t.Errorf("MatchRune: expected %v, got %v", true, false)
				}
			}
			for _, ch := range row.ExpectReject {
				if m.MatchRune(ch) {
					t.Errorf("MatchRune: expected %v, got %v", false, true)
				}
			}
		})
	}
}

func testEqualRanges(t *testing.T, actual, expect []runeMatchRange) {
	actualLen := uint(len(actual))
	expectLen := uint(len(expect))
	if actualLen < expectLen {
		t.Errorf("ForEachRange: expected len %d, got len %d (missing %d items)", expectLen, actualLen, expectLen-actualLen)
	} else if actualLen > expectLen {
		t.Errorf("ForEachRange: expected len %d, got len %d (excess %d items)", expectLen, actualLen, actualLen-expectLen)
	}

	var i, j uint
	for i < actualLen && j < expectLen {
		actualItem := actual[i]
		expectItem := expect[j]
		if actualItem.Lo < expectItem.Lo {
			t.Errorf("ForEachRange: unexpected item %v", actualItem)
			i++
			continue
		}
		if expectItem.Lo < actualItem.Lo {
			t.Errorf("ForEachRange: missing item %v", expectItem)
			j++
			continue
		}
		if actualItem.Hi != expectItem.Hi {
			t.Errorf("ForEachRange: wrong item: expected %v, got %v", expectItem, actualItem)
		}
		i++
		j++
	}
	for i < actualLen {
		actualItem := actual[i]
		i++
		t.Errorf("ForEachRange: unexpected item %v", actualItem)
	}
	for j < expectLen {
		expectItem := expect[j]
		j++
		t.Errorf("ForEachRange: missing item %v", expectItem)
	}
}

func TestNotSimplifications(t *testing.T) {
	m := Any().Not()
	_, ok := m.(*runeMatchNotAny)
	if !ok {
		t.Errorf("Any Not: expected %T, got %T", (*runeMatchNotAny)(nil), m)
	}

	m = None().Not()
	_, ok = m.(*runeMatchAny)
	if !ok {
		t.Errorf("None Not: expected %T, got %T", (*runeMatchAny)(nil), m)
	}

	m = Is('@').Not()
	_, ok = m.(*runeMatchNotIs)
	if !ok {
		t.Errorf("Is '@' Not: expected %T, got %T", (*runeMatchNotIs)(nil), m)
	}

	m = Is('@').Not().Not()
	_, ok = m.(*runeMatchIs)
	if !ok {
		t.Errorf("Is '@' Not Not: expected %T, got %T", (*runeMatchIs)(nil), m)
	}
}
