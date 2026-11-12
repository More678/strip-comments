package glossary

import (
	"strings"
	"testing"
)

func TestRegexComplexityRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"正则表达式拒绝服务攻击", "正则表达式计算资源耗尽风险核查", "正则表达式计算资源耗尽风险核查", 1},
		{"regular expression denial of service attack", "regular expression computational resource exhaustion risk review", "regular_expression_computational_resource_exhaustion_risk_review", 1},
		{"Regular expression denial of service attack", "Regular expression computational resource exhaustion risk review", "Regular_expression_computational_resource_exhaustion_risk_review", 1},
		{"rEgUlAr ExPrEsSiOn DeNiAl Of SeRvIcE aTtAcK", "regular expression computational resource exhaustion risk review", "regular_expression_computational_resource_exhaustion_risk_review", 1},
		{"REGULAR EXPRESSION DENIAL OF SERVICE ATTACK", "REGULAR EXPRESSION COMPUTATIONAL RESOURCE EXHAUSTION RISK REVIEW", "REGULAR_EXPRESSION_COMPUTATIONAL_RESOURCE_EXHAUSTION_RISK_REVIEW", 1},
		{"正则表达式拒绝服务攻击; regular expression denial of service attack", "正则表达式计算资源耗尽风险核查; regular expression computational resource exhaustion risk review", "正则表达式计算资源耗尽风险核查; regular_expression_computational_resource_exhaustion_risk_review", 2},
		{"XML递归实体展开攻击", "XML递归实体展开资源耗尽风险核查", "XML递归实体展开资源耗尽风险核查", 1},
	}
	for _, in := range []string{
		"ReDoS regex regular expression catastrophic backtracking",
		"正则表达式 回溯 拒绝服务 DoS DDoS CWE-1333",
		"regular expression denial of service 正则表达式拒绝服务",
		"正则表达式拒绝服务防护 正则表达式 拒绝服务攻击",
		"regexp.Compile regexp.MatchString Regex.IsMatch re.compile",
		"regex denial of service attack",
		"regular expression denial-of-service attack",
		"regular_expression_denial_of_service_attack",
		"regular\texpression denial of service attack",
		"xregular expression denial of service attack",
		"regular expression denial of service attacks",
		"2regular expression denial of service attack",
		"regular expression denial of service attack2",
		"_regular expression denial of service attack",
		"regular expression denial of service attack_suffix",
		`import "regular expression denial of service attack"`,
		`import "正则表达式拒绝服务攻击"`,
		"https://example.invalid/正则表达式拒绝服务攻击",
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
