package glossary

import (
	"strings"
	"testing"
)

func TestRecursionRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"不受控递归漏洞", "不受控递归资源耗尽风险核查", "不受控递归资源耗尽风险核查", 1},
		{"uncontrolled recursion vulnerability", "uncontrolled recursion resource exhaustion risk review", "uncontrolled_recursion_resource_exhaustion_risk_review", 1},
		{"Uncontrolled recursion vulnerability", "Uncontrolled recursion resource exhaustion risk review", "Uncontrolled_recursion_resource_exhaustion_risk_review", 1},
		{"uNcOnTrOlLeD rEcUrSiOn VuLnErAbIlItY", "uncontrolled recursion resource exhaustion risk review", "uncontrolled_recursion_resource_exhaustion_risk_review", 1},
		{"UNCONTROLLED RECURSION VULNERABILITY", "UNCONTROLLED RECURSION RESOURCE EXHAUSTION RISK REVIEW", "UNCONTROLLED_RECURSION_RESOURCE_EXHAUSTION_RISK_REVIEW", 1},
		{"不受控递归漏洞; uncontrolled recursion vulnerability", "不受控递归资源耗尽风险核查; uncontrolled recursion resource exhaustion risk review", "不受控递归资源耗尽风险核查; uncontrolled_recursion_resource_exhaustion_risk_review", 2},
		{"XML递归实体展开攻击", "XML递归实体展开资源耗尽风险核查", "XML递归实体展开资源耗尽风险核查", 1},
		{"正则表达式拒绝服务攻击", "正则表达式计算资源耗尽风险核查", "正则表达式计算资源耗尽风险核查", 1},
	}
	for _, in := range []string{
		"uncontrolled recursion 不受控递归 recursion 递归 depth stack CWE-674",
		"maxDepth RecursionError setrecursionlimit",
		"不受控递归防护 UncontrolledRecursionProtection",
		"不受控 递归漏洞 无限递归漏洞 无界递归漏洞",
		"uncontrolled-recursion vulnerability uncontrolled_recursion_vulnerability",
		"uncontrolled\trecursion vulnerability",
		"xuncontrolled recursion vulnerability",
		"uncontrolled recursion vulnerabilityx",
		"2uncontrolled recursion vulnerability",
		"uncontrolled recursion vulnerability2",
		"_uncontrolled recursion vulnerability",
		"uncontrolled recursion vulnerability_suffix",
		`import "uncontrolled recursion vulnerability"`,
		`import "不受控递归漏洞"`,
		"https://example.invalid/不受控递归漏洞",
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
