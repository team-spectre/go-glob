package guts

import (
	"testing"
	"unicode"
)

func TestMatchers(t *testing.T) {
	type testrow struct {
		Name         string
		Value        RuneMatcher
		ExpectRanges []RangeMatch
		ExpectAccept []rune
		ExpectReject []rune
	}
	testdata := []testrow{
		{
			Name:         "None",
			Value:        NoneValue,
			ExpectRanges: nil,
			ExpectAccept: nil,
			ExpectReject: []rune{0, '0', 'A', 'a', unicode.MaxRune},
		},
		{
			Name:         "Any",
			Value:        AnyValue,
			ExpectRanges: []RangeMatch{{0, unicode.MaxRune}},
			ExpectAccept: []rune{0, '0', 'A', 'a', unicode.MaxRune},
		},
		{
			Name:         "Range [0-9]",
			Value:        &RangeMatch{'0', '9'},
			ExpectRanges: []RangeMatch{{'0', '9'}},
			ExpectAccept: []rune{'0', '9'},
			ExpectReject: []rune{0, '/', ':', 'A', 'a', unicode.MaxRune},
		},
		{
			Name:         "Range [^0-9]",
			Value:        &ExceptRangeMatch{'0', '9'},
			ExpectRanges: []RangeMatch{{0, '/'}, {':', unicode.MaxRune}},
			ExpectAccept: []rune{0, '/', ':', 'A', 'a', unicode.MaxRune},
			ExpectReject: []rune{'0', '9'},
		},
	}
	for _, row := range testdata {
		t.Run(row.Name, func(t *testing.T) {
			m := row.Value

			actualRanges := make([]RangeMatch, 0, len(row.ExpectRanges))
			m.ForEachRange(func(lo, hi rune) {
				actualRanges = append(actualRanges, RangeMatch{lo, hi})
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

func testEqualRanges(t *testing.T, actual, expect []RangeMatch) {
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
	m := AnyValue.Not()
	_, ok := m.(*NoneMatch)
	if !ok {
		t.Errorf("Any Not: expected %T, got %T", (*NoneMatch)(nil), m)
	}

	m = NoneValue.Not()
	_, ok = m.(*AnyMatch)
	if !ok {
		t.Errorf("None Not: expected %T, got %T", (*AnyMatch)(nil), m)
	}

	m = (&IsMatch{'@'}).Not()
	_, ok = m.(*IsNotMatch)
	if !ok {
		t.Errorf("Is '@' Not: expected %T, got %T", (*IsNotMatch)(nil), m)
	}

	m = (&IsNotMatch{'@'}).Not()
	_, ok = m.(*IsMatch)
	if !ok {
		t.Errorf("Is '@' Not Not: expected %T, got %T", (*IsMatch)(nil), m)
	}
}
