package glossary

import (
	"strings"
	"testing"
)

func TestCSSInjectionRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"CSS注入攻击", "CSS样式注入风险核查", "CSS样式注入风险核查", 1},
		{"CSS injection attack", "CSS style injection risk review", "CSS_style_injection_risk_review", 1},
		{"css injection attack", "CSS style injection risk review", "CSS_style_injection_risk_review", 1},
		{"Css injection attack", "CSS style injection risk review", "CSS_style_injection_risk_review", 1},
		{"cSs InJeCtIoN AtTaCk", "CSS style injection risk review", "CSS_style_injection_risk_review", 1},
		{"CSS INJECTION ATTACK", "CSS STYLE INJECTION RISK REVIEW", "CSS_STYLE_INJECTION_RISK_REVIEW", 1},
		{"CSS注入攻击; CSS injection attack", "CSS样式注入风险核查; CSS style injection risk review", "CSS样式注入风险核查; CSS_style_injection_risk_review", 2},
		{"review:\nCSS injection attack", "review:\nCSS style injection risk review", "review:\nCSS_style_injection_risk_review", 1},
		{"(CSS injection attack).", "(CSS style injection risk review).", "(CSS_style_injection_risk_review).", 1},
		{"HTML注入攻击", "HTML标记注入风险核查", "HTML标记注入风险核查", 1},
		{"HTML injection attack", "HTML markup injection risk review", "HTML_markup_injection_risk_review", 1},
		{"跨站脚本包含攻击", "跨站脚本包含数据泄露风险核查", "跨站脚本包含数据泄露风险核查", 1},
		{"cross site script inclusion attack", "cross site script inclusion data exposure risk review", "cross_site_script_inclusion_data_exposure_risk_review", 1},
	}
	for _, in := range []string{
		"CSS CSS injection CSS注入 style stylesheet",
		"cssText CSSStyleSheet setProperty insertRule Content-Security-Policy WSTG-CLNT-05",
		"CSS injection prevention sanitizeCSS CSP style-src",
		"CSS style injection risk review CSS样式注入风险核查",
		"CSS注入防护 css注入攻击 Css注入攻击 CSS 注入攻击 CSS注入 攻击",
		"stylesheet injection attack 样式表注入攻击",
		"CSS-injection-attack CSS_injection_attack",
		"CSS injection attacks",
		"CSS\tinjection attack",
		"CSS  injection attack",
		"CSS\ninjection attack",
		"CSS injection\tattack",
		"xCSS injection attack",
		"CSS injection attackx",
		"2CSS injection attack",
		"CSS injection attack2",
		"_CSS injection attack",
		"CSS injection attack_suffix",
		"前CSS injection attack",
		"CSS injection attack后",
		`import "CSS injection attack"`,
		`import "CSS注入攻击"`,
		"https://example.invalid/CSS%20injection%20attack",
		"https://example.invalid/CSS注入攻击",
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
					if h.Start < 0 || h.End > len(tc.in) || h.Start >= h.End {
						t.Fatalf("invalid hit range: %+v", h)
					}
					if h.InComment != comment || h.Line != strings.Count(tc.in[:h.Start], "\n")+1 || tc.in[h.Start:h.End] != h.From || !strings.Contains(want, h.To) {
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
