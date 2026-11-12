package glossary

import (
	"strings"
	"testing"
)

func TestHostHeaderRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"HTTP主机头注入攻击", "HTTP主机头注入风险核查", "HTTP主机头注入风险核查", 1},
		{"HTTP host header injection attack", "HTTP host header injection risk review", "HTTP_host_header_injection_risk_review", 1},
		{"http host header injection attack", "HTTP host header injection risk review", "HTTP_host_header_injection_risk_review", 1},
		{"Http host header injection attack", "HTTP host header injection risk review", "HTTP_host_header_injection_risk_review", 1},
		{"hTtP hOsT hEaDeR iNjEcTiOn AtTaCk", "HTTP host header injection risk review", "HTTP_host_header_injection_risk_review", 1},
		{"HTTP HOST HEADER INJECTION ATTACK", "HTTP HOST HEADER INJECTION RISK REVIEW", "HTTP_HOST_HEADER_INJECTION_RISK_REVIEW", 1},
		{"HTTP主机头注入攻击; HTTP host header injection attack", "HTTP主机头注入风险核查; HTTP host header injection risk review", "HTTP主机头注入风险核查; HTTP_host_header_injection_risk_review", 2},
		{"HTTP响应拆分", "HTTP响应头换行处理风险核查", "HTTP响应头换行处理风险核查", 1},
		{"HTTP response splitting", "HTTP response header newline handling risk review", "HTTP_response_header_newline_handling_risk_review", 1},
		{"Web缓存投毒", "Web缓存响应完整性风险核查", "Web缓存响应完整性风险核查", 1},
		{"web cache poisoning", "web cache response integrity risk review", "web_cache_response_integrity_risk_review", 1},
	}
	for _, in := range []string{
		"host header injection HTTP host header injection HTTP主机头注入",
		"Host X-Forwarded-Host Forwarded :authority Request.Host WSTG-INPV-17",
		"HTTP host header 主机头 Host头 header 注入",
		"HTTP主机头注入防护 http主机头注入攻击 HTTP 主机头注入攻击",
		"host header injection attack 主机头注入攻击 Host头注入攻击",
		"HTTP-host-header-injection-attack HTTP_host_header_injection_attack",
		"HTTP\thost header injection attack",
		"xHTTP host header injection attack",
		"HTTP host header injection attacks",
		"2HTTP host header injection attack",
		"HTTP host header injection attack2",
		"_HTTP host header injection attack",
		"HTTP host header injection attack_suffix",
		`import "HTTP host header injection attack"`,
		`import "HTTP主机头注入攻击"`,
		"https://example.invalid/HTTP主机头注入攻击",
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
