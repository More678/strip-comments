package scan

import (
	"testing"

	"github.com/More678/strip-comments/internal/lang"
)

// TestNoTokenGlue covers a bug present since v1.0: a comment separates the
// tokens around it, so deleting one must not weld them together. `int/**/x`
// compiled before the strip and read as `intx` after it, and `1 -/**/-1`
// turned into `1 --1`, which is a different expression.
func TestNoTokenGlue(t *testing.T) {
	cases := []struct {
		name, lang, in, want string
	}{
		{
			name: "c identifiers stay apart",
			lang: "c", in: "int/**/x = 1;\n", want: "int x = 1;\n",
		},
		{
			name: "c operators stay apart",
			lang: "c", in: "int y = 1 -/**/-1;\n", want: "int y = 1 - -1;\n",
		},
		{
			name: "go keyword and name stay apart",
			lang: "go", in: "var/**/n = 1\n", want: "var n = 1\n",
		},
		{
			// A bracket cannot merge with anything, so no space is added and
			// the long-standing `f( 1)` shape is unchanged.
			name: "brackets need no space",
			lang: "c", in: "f(/* x */1);\n", want: "f(1);\n",
		},
		{
			name: "existing space is not doubled",
			lang: "c", in: "int /* x */ y;\n", want: "int y;\n",
		},
		{
			// Markup is text, not tokens: a space here would change what the
			// page renders, so the halves are joined.
			name: "html joins text",
			lang: "html", in: "<p>a<!-- b -->c</p>\n", want: "<p>ac</p>\n",
		},
		{
			name: "multi-line comment needs no space",
			lang: "c", in: "int x;/* a\nb */int y;\n", want: "int x;\nint y;\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := lang.ByName(tc.lang)
			if l == nil {
				t.Fatalf("language %q not registered", tc.lang)
			}
			got := Strip([]byte(tc.in), l, Options{})
			if string(got.Output) != tc.want {
				t.Errorf("got  %q\nwant %q", got.Output, tc.want)
			}
		})
	}
}
