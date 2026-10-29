package glossary

import (
	"strings"
	"testing"
)

// TestDefaultReplacementsAreFixedPoints guards idempotence for the whole
// built-in table at once: no replacement may contain a term that some rule
// would rewrite again, or every run would drift the file a little further.
// It also covers the identifier form, since that is what code positions get.
func TestDefaultReplacementsAreFixedPoints(t *testing.T) {
	tb := NewTable(Default)
	for _, r := range Default {
		for _, to := range []string{r.To, toIdent(r.To), strings.ToUpper(r.To)} {
			// Pad with spaces so a word-boundary rule sees real edges.
			got, hits := tb.Apply([]byte(" "+to+" "), Options{})
			if len(hits) != 0 {
				t.Errorf("replacement %q (for %q) is rewritten again to %q",
					to, r.From, got)
			}
		}
	}
}

// TestDefaultHasNoDuplicates: two rules for the same term would make the
// result depend on table order, which nobody reading the table would expect.
func TestDefaultHasNoDuplicates(t *testing.T) {
	seen := map[string]bool{}
	for _, r := range Default {
		key := r.From
		if r.CI {
			key = strings.ToLower(key)
		}
		if seen[key] {
			t.Errorf("term %q appears twice", r.From)
		}
		seen[key] = true
	}
}

func TestNewTermsBoundaries(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct{ in, want string }{
		// C2 is the shortest rule; its word boundary carries all the weight.
		{"connect to the C2 server", "connect to the outbound anomaly server"},
		{"C2H5OH", "C2H5OH"},
		{"x := 0xC2", "x := 0xC2"},
		{"mC2 := 1", "mC2 := 1"},
		// Acronyms are case-sensitive: a lowercase c2 is usually a variable,
		// and carrying the case of "MITM" would shout the replacement.
		{"c2 := 1", "c2 := 1"},
		{"a MITM proxy", "a link hijack detection proxy"},
		// Ambiguous words were added only as compounds.
		{"钓鱼竿", "钓鱼竿"},
		{"收到钓鱼邮件", "收到可疑邮件"},
		{"中间人", "中间人"},
		{"防范中间人攻击", "防范链路劫持检测"},
		{"定向爆破", "定向爆破"},
		{"口令爆破", "口令强度核查"},
		// Mixed CJK/ASCII terms match inside Chinese prose.
		{"检测反弹shell行为", "检测异常外联检测行为"},
	}
	for _, tc := range cases {
		got, _ := tb.Apply([]byte(tc.in), Options{
			// Treat it all as a comment so prose keeps its spaces.
			Comments: []Region{{Start: 0, End: len(tc.in), Comment: true}},
		})
		if string(got) != tc.want {
			t.Errorf("%q: got %q, want %q", tc.in, got, tc.want)
		}
	}
}
