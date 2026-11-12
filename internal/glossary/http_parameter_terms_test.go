package glossary

import (
	"strings"
	"testing"
)

func TestHTTPParameterRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"HTTP参数污染攻击", "HTTP参数污染与解释风险核查", "HTTP参数污染与解释风险核查", 1},
		{"HTTP parameter pollution attack", "HTTP parameter pollution interpretation risk review", "HTTP_parameter_pollution_interpretation_risk_review", 1},
		{"http parameter pollution attack", "HTTP parameter pollution interpretation risk review", "HTTP_parameter_pollution_interpretation_risk_review", 1},
		{"Http parameter pollution attack", "HTTP parameter pollution interpretation risk review", "HTTP_parameter_pollution_interpretation_risk_review", 1},
		{"hTtP pArAmEtEr PoLlUtIoN aTtAcK", "HTTP parameter pollution interpretation risk review", "HTTP_parameter_pollution_interpretation_risk_review", 1},
		{"HTTP PARAMETER POLLUTION ATTACK", "HTTP PARAMETER POLLUTION INTERPRETATION RISK REVIEW", "HTTP_PARAMETER_POLLUTION_INTERPRETATION_RISK_REVIEW", 1},
		{"HTTP参数污染攻击; HTTP parameter pollution attack", "HTTP参数污染与解释风险核查; HTTP parameter pollution interpretation risk review", "HTTP参数污染与解释风险核查; HTTP_parameter_pollution_interpretation_risk_review", 2},
		{"HTTP请求走私", "HTTP请求解析差异风险核查", "HTTP请求解析差异风险核查", 1},
		{"原型污染攻击", "对象原型篡改风险核查", "对象原型篡改风险核查", 1},
	}
	for _, in := range []string{
		"HTTP parameter pollution HTTP参数污染 HPP 重复参数",
		"query parameter 参数 HTTP request response",
		"CAPEC-460 WSTG-INPV-04 Content-Length Transfer-Encoding",
		"url.Values Query.Get Query.GetValues getParameterValues",
		"HTTP参数污染防护 http参数污染攻击 HTTP 参数污染攻击",
		"http-parameter-pollution-attack HTTP_parameter_pollution_attack",
		"HTTP\tparameter pollution attack",
		"xHTTP parameter pollution attack",
		"HTTP parameter pollution attacks",
		"2HTTP parameter pollution attack",
		"HTTP parameter pollution attack2",
		"_HTTP parameter pollution attack",
		"HTTP parameter pollution attack_suffix",
		`import "HTTP parameter pollution attack"`,
		`import "HTTP参数污染攻击"`,
		"https://example.invalid/HTTP参数污染攻击",
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
