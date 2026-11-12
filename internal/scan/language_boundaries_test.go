package scan

import (
	"testing"

	"github.com/More678/strip-comments/internal/lang"
)

func TestCSharpFormatBoundaries(t *testing.T) {
	for _, src := range []string{
		`var s = $"{1:F"; var other = "}"; // keep` + "\n",
		`var s = $@"{1:F"; var other = "}"; // keep` + "\n",
		`var s = $"""{1:F"""; var other = "}"""; // keep` + "\n",
		`var s = $$""""{{1:F""""; var other = "}}""""; // keep` + "\n",
	} {
		t.Run(src, func(t *testing.T) {
			res := Strip([]byte(src), lang.ByName("csharp"), Options{})
			if !res.Unsafe || string(res.Output) != src {
				t.Fatalf("expected unchanged malformed format, got %+v", res)
			}
		})
	}
	for _, literal := range []string{
		`$"{1:\"/* keep */\"}"`,
		`$@"{1:""/* keep */""}"`,
		`$"""{1:"/* keep */"}"""`,
		`$$""""{{1:"""/* keep */"""}}""""`,
	} {
		t.Run(literal, func(t *testing.T) {
			want := "var s = " + literal + ";\n"
			src := "var s = " + literal + "; // gone\n"
			res := Strip([]byte(src), lang.ByName("csharp"), Options{})
			if res.Unsafe || string(res.Output) != want || res.Comments != 1 {
				t.Fatalf("got %+v, want %q", res, want)
			}
			again := Strip(res.Output, lang.ByName("csharp"), Options{})
			if again.Unsafe || string(again.Output) != want || again.Comments != 0 {
				t.Fatalf("second pass changed output: %+v", again)
			}
		})
	}
}

func TestRustDocCommentBoundaries(t *testing.T) {
	cases := []struct {
		comment string
		doc     bool
	}{
		{"/// outer docs", true},
		{"//! inner docs", true},
		{"//// ordinary", false},
		{"/** outer docs */", true},
		{"/*! inner docs */", true},
		{"/*** ordinary */", false},
		{"/**** ordinary */", false},
		{"/***/", false},
		{"/**/", false},
		{"/*** outer /* nested */ tail */", false},
	}
	for _, tc := range cases {
		t.Run(tc.comment, func(t *testing.T) {
			for _, keep := range []bool{false, true} {
				src := tc.comment + "\nfn main() {}\n"
				want, comments := "fn main() {}\n", 1
				if keep && tc.doc {
					want, comments = src, 0
				}
				opt := Options{KeepDoc: keep}
				res := Strip([]byte(src), lang.ByName("rust"), opt)
				if res.Unsafe || string(res.Output) != want || res.Comments != comments {
					t.Fatalf("keep-doc=%v: got %+v, want %q (%d removed)", keep, res, want, comments)
				}
				again := Strip(res.Output, lang.ByName("rust"), opt)
				if again.Unsafe || string(again.Output) != want || again.Comments != 0 {
					t.Fatalf("second pass changed output: %+v", again)
				}
			}
		})
	}
}
