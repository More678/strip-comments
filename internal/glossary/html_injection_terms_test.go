package glossary

import (
	"strings"
	"testing"
)

func TestHTMLInjectionRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"HTML注入攻击", "HTML标记注入风险核查", "HTML标记注入风险核查", 1},
		{"HTML injection attack", "HTML markup injection risk review", "HTML_markup_injection_risk_review", 1},
		{"html injection attack", "HTML markup injection risk review", "HTML_markup_injection_risk_review", 1},
		{"Html injection attack", "HTML markup injection risk review", "HTML_markup_injection_risk_review", 1},
		{"hTmL iNjEcTiOn AtTaCk", "HTML markup injection risk review", "HTML_markup_injection_risk_review", 1},
		{"HTML INJECTION ATTACK", "HTML MARKUP INJECTION RISK REVIEW", "HTML_MARKUP_INJECTION_RISK_REVIEW", 1},
		{"HTML注入攻击; HTML injection attack", "HTML标记注入风险核查; HTML markup injection risk review", "HTML标记注入风险核查; HTML_markup_injection_risk_review", 2},
		{"DOM clobbering attack", "DOM named property clobbering risk review", "DOM_named_property_clobbering_risk_review", 1},
		{"服务端模板注入攻击", "服务端模板表达式注入风险核查", "服务端模板表达式注入风险核查", 1},
	}
	for _, in := range []string{
		"HTML injection HTML注入 HTML markup XSS WSTG-CLNT-03",
		"innerHTML textContent document.write DOMPurify.sanitize",
		"HTML injection prevention HTMLInjectionProtection",
		"HTML注入防护 html注入攻击 HTML 注入攻击",
		"HTML标记注入攻击 markup injection attack",
		"HTML-injection-attack HTML_injection_attack",
		"HTML\tinjection attack",
		"xHTML injection attack",
		"HTML injection attacks",
		"2HTML injection attack",
		"HTML injection attack2",
		"_HTML injection attack",
		"HTML injection attack_suffix",
		`import "HTML injection attack"`,
		`import "HTML注入攻击"`,
		"https://example.invalid/HTML注入攻击",
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
