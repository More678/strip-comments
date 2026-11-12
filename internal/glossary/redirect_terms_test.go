package glossary

import (
	"strings"
	"testing"
)

func TestRedirectRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"开放重定向攻击", "非可信目标重定向风险核查", "非可信目标重定向风险核查", 1},
		{"open redirect attack", "untrusted destination redirect risk review", "untrusted_destination_redirect_risk_review", 1},
		{"Open redirect attack", "Untrusted destination redirect risk review", "Untrusted_destination_redirect_risk_review", 1},
		{"oPeN rEdIrEcT aTtAcK", "untrusted destination redirect risk review", "untrusted_destination_redirect_risk_review", 1},
		{"OPEN REDIRECT ATTACK", "UNTRUSTED DESTINATION REDIRECT RISK REVIEW", "UNTRUSTED_DESTINATION_REDIRECT_RISK_REVIEW", 1},
		{"开放重定向攻击; open redirect attack", "非可信目标重定向风险核查; untrusted destination redirect risk review", "非可信目标重定向风险核查; untrusted_destination_redirect_risk_review", 2},
		{"HTTP响应拆分", "HTTP响应头换行处理风险核查", "HTTP响应头换行处理风险核查", 1},
		{"HTTP response splitting", "HTTP response header newline handling risk review", "HTTP_response_header_newline_handling_risk_review", 1},
	}
	for _, in := range []string{
		"open redirect 开放重定向 redirect 重定向 URL Location",
		"301 302 303 307 308 CWE-601",
		"http.Redirect RedirectResponse Response.Redirect sendRedirect",
		"开放重定向防护 开放 重定向攻击",
		"unvalidated redirect attack",
		"open-redirect-attack open_redirect_attack",
		"open\tredirect attack",
		"xopen redirect attack",
		"open redirect attacks",
		"2open redirect attack",
		"open redirect attack2",
		"_open redirect attack",
		"open redirect attack_suffix",
		`import "open redirect attack"`,
		`import "开放重定向攻击"`,
		"https://example.invalid/开放重定向攻击",
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
