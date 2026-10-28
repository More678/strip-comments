package scan

import (
	"testing"

	"github.com/More678/strip-comments/internal/glossary"
	"github.com/More678/strip-comments/internal/lang"
)

func TestStripWithTerms(t *testing.T) {
	tb := glossary.NewTable(glossary.Default)
	cases := []struct {
		name      string
		opt       Options
		in, want  string
		wantTerms int
		wantInCom int // how many of the hits were inside a comment
	}{
		{
			name: "code and comment both hit, comment then stripped",
			opt:  Options{Glossary: tb},
			in:   "var 渗透 = 1 // 渗透\n",
			want: "var 威胁检测 = 1\n", wantTerms: 2, wantInCom: 1,
		},
		{
			// "pentest" -> "threat_detection" grows the buffer by 9 bytes
			// before both comments. Stale spans would cut into the code.
			// The underscore is the known cost of toIdent: the glossary
			// cannot tell a string literal from an identifier.
			name: "length change before comments re-lexes spans",
			opt:  Options{Glossary: tb},
			in:   "s := \"pentest\" // a\nt := 1 // b\n",
			want: "s := \"threat_detection\"\nt := 1\n", wantTerms: 1,
		},
		{
			name: "comments only leaves code alone",
			opt:  Options{Glossary: tb, TermsInComments: true, KeepComments: true},
			in:   "var 渗透 = 1 // 渗透\n",
			want: "var 渗透 = 1 // 威胁检测\n", wantTerms: 1, wantInCom: 1,
		},
		{
			name: "keep comments rewrites without stripping",
			opt:  Options{Glossary: tb, KeepComments: true},
			in:   "// 漏洞利用\nvar x = 1\n",
			want: "// 风险验证\nvar x = 1\n", wantTerms: 1, wantInCom: 1,
		},
		{
			name: "nil glossary is the old behaviour",
			opt:  Options{},
			in:   "var 渗透 = 1 // 渗透\n",
			want: "var 渗透 = 1\n",
		},
	}
	goLang := lang.ByName("go")
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := Strip([]byte(tc.in), goLang, tc.opt)
			if string(res.Output) != tc.want {
				t.Errorf("output\n got  %q\n want %q", res.Output, tc.want)
			}
			if len(res.Terms) != tc.wantTerms {
				t.Errorf("terms = %d, want %d: %+v", len(res.Terms), tc.wantTerms, res.Terms)
			}
			in := 0
			for _, h := range res.Terms {
				if h.InComment {
					in++
				}
			}
			if in != tc.wantInCom {
				t.Errorf("in-comment hits = %d, want %d", in, tc.wantInCom)
			}
		})
	}
}

// TestTermsUnsafeKeepsOriginal covers a replacement that opens a literal it
// does not close. The file must come back untouched: Output equals the input
// whenever Unsafe is set without Force, so a caller writing it back cannot
// save a file whose structure the replacement just broke.
func TestTermsUnsafeKeepsOriginal(t *testing.T) {
	// The replacement injects a backtick in code position, which opens a Go
	// raw string. Those are multiline, so it runs to EOF unterminated.
	tb := glossary.NewTable([]glossary.Rule{{From: "QUOTE", To: "`"}})
	in := "var s = QUOTE\n// c\n"
	res := Strip([]byte(in), lang.ByName("go"), Options{Glossary: tb})

	if !res.Unsafe {
		t.Errorf("Unsafe = false, want true: %q", res.Output)
	}
	if string(res.Output) != in {
		t.Errorf("output = %q, want the input %q", res.Output, in)
	}
	if len(res.Terms) != 1 {
		t.Errorf("terms = %+v, want the hit to still be reported", res.Terms)
	}
	if res.Reason == "" {
		t.Error("Reason is empty, want it to name the cause")
	}
}

// TestTermsForceStillStrips is the counterpart: with Force the caller has
// said it accepts the risk, so the rewrite goes through.
func TestTermsForceStillStrips(t *testing.T) {
	tb := glossary.NewTable([]glossary.Rule{{From: "QUOTE", To: "`"}})
	in := "var s = QUOTE\n// c\n"
	res := Strip([]byte(in), lang.ByName("go"), Options{Glossary: tb, Force: true})
	if string(res.Output) == in {
		t.Errorf("output unchanged with Force: %q", res.Output)
	}
}
