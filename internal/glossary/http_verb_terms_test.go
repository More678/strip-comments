package glossary

import (
	"strings"
	"testing"
)

func TestHTTPVerbRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"HTTP verb tampering attack", "HTTP verb tampering access control risk review", "HTTP_verb_tampering_access_control_risk_review", 1},
		{"http verb tampering attack", "HTTP verb tampering access control risk review", "HTTP_verb_tampering_access_control_risk_review", 1},
		{"Http verb tampering attack", "HTTP verb tampering access control risk review", "HTTP_verb_tampering_access_control_risk_review", 1},
		{"hTtP vErB tAmPeRiNg AtTaCk", "HTTP verb tampering access control risk review", "HTTP_verb_tampering_access_control_risk_review", 1},
		{"HTTP VERB TAMPERING ATTACK", "HTTP VERB TAMPERING ACCESS CONTROL RISK REVIEW", "HTTP_VERB_TAMPERING_ACCESS_CONTROL_RISK_REVIEW", 1},
		{"HTTP verb tampering attack; HTTP verb tampering attack", "HTTP verb tampering access control risk review; HTTP verb tampering access control risk review", "HTTP_verb_tampering_access_control_risk_review; HTTP_verb_tampering_access_control_risk_review", 2},
		{"HTTP参数污染攻击", "HTTP参数污染与解释风险核查", "HTTP参数污染与解释风险核查", 1},
		{"HTTP主机头注入攻击", "HTTP主机头注入风险核查", "HTTP主机头注入风险核查", 1},
	}
	for _, in := range []string{
		"HTTP verb tampering HTTP verb method",
		"GET POST HEAD OPTIONS PUT DELETE PATCH Allow WSTG-INPV-03",
		"Request.Method http.MethodGet HttpRequest.getMethod",
		"HTTP verb tampering prevention HTTPVerbTamperingProtection",
		"HTTP方法篡改攻击 HTTP method tampering attack",
		"HTTP-verb-tampering-attack HTTP_verb_tampering_attack",
		"HTTP\tverb tampering attack",
		"xHTTP verb tampering attack",
		"HTTP verb tampering attacks",
		"2HTTP verb tampering attack",
		"HTTP verb tampering attack2",
		"_HTTP verb tampering attack",
		"HTTP verb tampering attack_suffix",
		`import "HTTP verb tampering attack"`,
		"https://example.invalid/HTTP%20verb%20tampering%20attack",
	} {
		cases = append(cases, struct {
			in, prose, code string
			hits            int
		}{in, in, in, 0})
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			for _, comment := range []bool{false, true} {
				opt := Options{}
				want := tc.code
				if comment {
					opt.Comments = []Region{{Start: 0, End: len(tc.in), Comment: true}}
					want = tc.prose
				}
				out, hits := tb.Apply([]byte(tc.in), opt)
				if string(out) != want || len(hits) != tc.hits {
					t.Fatalf("comment=%v: got %q (%d hits), want %q (%d hits)", comment, out, len(hits), want, tc.hits)
				}
				for _, h := range hits {
					if h.InComment != comment || h.Line != 1 || tc.in[h.Start:h.End] != h.From || !strings.Contains(want, h.To) {
						t.Errorf("incorrect hit metadata: %+v", h)
					}
				}
				if comment {
					opt.Comments[0].End = len(out)
				}
				again, secondHits := tb.Apply(out, opt)
				if string(again) != string(out) || len(secondHits) != 0 {
					t.Fatalf("second pass changed output: %q (%+v)", again, secondHits)
				}
			}
		})
	}
}
