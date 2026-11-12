package glossary

import (
	"strings"
	"testing"
)

func TestTemplateXPathRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"服务端模板注入攻击", "服务端模板表达式注入风险核查", "服务端模板表达式注入风险核查", 1},
		{"XPath注入攻击", "XPath查询表达式注入风险核查", "XPath查询表达式注入风险核查", 1},
		{"server-side template injection attack", "server-side template expression injection risk review", "server_side_template_expression_injection_risk_review", 1},
		{"XPath injection attack", "XPath query expression injection risk review", "XPath_query_expression_injection_risk_review", 1},
		{"Server-side template injection attack", "Server-side template expression injection risk review", "Server_side_template_expression_injection_risk_review", 1},
		{"sErVeR-sIdE tEmPlAtE iNjEcTiOn AtTaCk", "server-side template expression injection risk review", "server_side_template_expression_injection_risk_review", 1},
		{"SERVER-SIDE TEMPLATE INJECTION ATTACK", "SERVER-SIDE TEMPLATE EXPRESSION INJECTION RISK REVIEW", "SERVER_SIDE_TEMPLATE_EXPRESSION_INJECTION_RISK_REVIEW", 1},
		{"xpath injection attack", "XPath query expression injection risk review", "XPath_query_expression_injection_risk_review", 1},
		{"XPATH INJECTION ATTACK", "XPATH QUERY EXPRESSION INJECTION RISK REVIEW", "XPATH_QUERY_EXPRESSION_INJECTION_RISK_REVIEW", 1},
		{"服务端模板注入攻击; XPath注入攻击", "服务端模板表达式注入风险核查; XPath查询表达式注入风险核查", "服务端模板表达式注入风险核查; XPath查询表达式注入风险核查", 2},
		{"XML外部实体攻击", "XML外部实体解析风险核查", "XML外部实体解析风险核查", 1},
		{"XML递归实体展开攻击", "XML递归实体展开资源耗尽风险核查", "XML递归实体展开资源耗尽风险核查", 1},
	}
	for _, in := range []string{
		"server-side template injection; XPath injection; 服务端模板注入; XPath注入",
		"template 模板 XPath SSTI injection 注入 CWE-1336 CWE-643",
		"template.Execute ExecuteTemplate XPathExpression XPathFactory evaluate",
		"服务端模板注入防护 XPath注入防护",
		"服务器端模板注入攻击 客户端模板注入攻击",
		"client-side template injection attack",
		"server side template injection attack",
		"server_side_template_injection_attack",
		"server-side\ttemplate injection attack",
		"xpath注入攻击 XPath 注入攻击 服务端 模板注入攻击",
		"xserver-side template injection attack",
		"server-side template injection attacks",
		"2XPath injection attack",
		"XPath injection attack2",
		"_XPath injection attack",
		"XPath injection attack_suffix",
		`import "server-side template injection attack"`,
		`import "XPath injection attack"`,
		"https://example.invalid/服务端模板注入攻击",
		"https://example.invalid/XPath注入攻击",
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
