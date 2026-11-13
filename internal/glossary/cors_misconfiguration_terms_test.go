package glossary

import (
	"strings"
	"testing"
)

func TestCORSMisconfigurationRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"CORS错误配置漏洞", "CORS错误配置风险核查", "CORS错误配置风险核查", 1},
		{"CORS misconfiguration vulnerability", "CORS misconfiguration risk review", "CORS_misconfiguration_risk_review", 1},
		{"cors misconfiguration vulnerability", "CORS misconfiguration risk review", "CORS_misconfiguration_risk_review", 1},
		{"Cors misconfiguration vulnerability", "CORS misconfiguration risk review", "CORS_misconfiguration_risk_review", 1},
		{"cOrS MiScOnFiGuRaTiOn VuLnErAbIlItY", "CORS misconfiguration risk review", "CORS_misconfiguration_risk_review", 1},
		{"CORS MISCONFIGURATION VULNERABILITY", "CORS MISCONFIGURATION RISK REVIEW", "CORS_MISCONFIGURATION_RISK_REVIEW", 1},
		{"CORS错误配置漏洞; CORS misconfiguration vulnerability", "CORS错误配置风险核查; CORS misconfiguration risk review", "CORS错误配置风险核查; CORS_misconfiguration_risk_review", 2},
		{"review:\nCORS misconfiguration vulnerability", "review:\nCORS misconfiguration risk review", "review:\nCORS_misconfiguration_risk_review", 1},
		{"(CORS misconfiguration vulnerability).", "(CORS misconfiguration risk review).", "(CORS_misconfiguration_risk_review).", 1},
		{"跨站WebSocket劫持攻击", "跨站WebSocket会话劫持风险核查", "跨站WebSocket会话劫持风险核查", 1},
		{"cross-site WebSocket hijacking attack", "cross-site WebSocket session hijacking risk review", "cross_site_WebSocket_session_hijacking_risk_review", 1},
		{"跨站脚本包含攻击", "跨站脚本包含数据泄露风险核查", "跨站脚本包含数据泄露风险核查", 1},
		{"cross site script inclusion attack", "cross site script inclusion data exposure risk review", "cross_site_script_inclusion_data_exposure_risk_review", 1},
	}
	for _, in := range []string{
		"CORS CORS misconfiguration CORS错误配置 Origin",
		"Access-Control-Allow-Origin Access-Control-Allow-Credentials Vary OPTIONS",
		"fetch XMLHttpRequest WSTG-CLNT-07 credentials withCredentials",
		"CORS misconfiguration prevention validateOrigin CorsPolicy allowlist",
		"CORS misconfiguration risk review CORS错误配置风险核查",
		"CORS错误配置防护 cors错误配置漏洞 Cors错误配置漏洞 CORS 错误配置漏洞 CORS错误配置 漏洞",
		"跨源资源共享错误配置漏洞 CORS attack misconfiguration vulnerability",
		"CORS-misconfiguration-vulnerability CORS_misconfiguration_vulnerability",
		"CORS misconfiguration vulnerabilities",
		"CORS\tmisconfiguration vulnerability",
		"CORS  misconfiguration vulnerability",
		"CORS\nmisconfiguration vulnerability",
		"CORS misconfiguration\tvulnerability",
		"xCORS misconfiguration vulnerability",
		"CORS misconfiguration vulnerabilityx",
		"2CORS misconfiguration vulnerability",
		"CORS misconfiguration vulnerability2",
		"_CORS misconfiguration vulnerability",
		"CORS misconfiguration vulnerability_suffix",
		"前CORS misconfiguration vulnerability",
		"CORS misconfiguration vulnerability后",
		`import "CORS misconfiguration vulnerability"`,
		`import "CORS错误配置漏洞"`,
		"https://example.invalid/CORS%20misconfiguration%20vulnerability",
		"https://example.invalid/CORS错误配置漏洞",
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
