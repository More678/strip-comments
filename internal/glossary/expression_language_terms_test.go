package glossary

import (
	"strings"
	"testing"
)

func TestExpressionLanguageRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"表达式语言注入攻击", "表达式语言求值注入风险核查", "表达式语言求值注入风险核查", 1},
		{"expression language injection attack", "expression language evaluation injection risk review", "expression_language_evaluation_injection_risk_review", 1},
		{"Expression language injection attack", "Expression language evaluation injection risk review", "Expression_language_evaluation_injection_risk_review", 1},
		{"eXpReSsIoN LaNgUaGe InJeCtIoN AtTaCk", "expression language evaluation injection risk review", "expression_language_evaluation_injection_risk_review", 1},
		{"EXPRESSION LANGUAGE INJECTION ATTACK", "EXPRESSION LANGUAGE EVALUATION INJECTION RISK REVIEW", "EXPRESSION_LANGUAGE_EVALUATION_INJECTION_RISK_REVIEW", 1},
		{"表达式语言注入攻击; expression language injection attack", "表达式语言求值注入风险核查; expression language evaluation injection risk review", "表达式语言求值注入风险核查; expression_language_evaluation_injection_risk_review", 2},
		{"review:\nexpression language injection attack", "review:\nexpression language evaluation injection risk review", "review:\nexpression_language_evaluation_injection_risk_review", 1},
		{"(expression language injection attack).", "(expression language evaluation injection risk review).", "(expression_language_evaluation_injection_risk_review).", 1},
		{"服务端模板注入攻击", "服务端模板表达式注入风险核查", "服务端模板表达式注入风险核查", 1},
		{"server-side template injection attack", "server-side template expression injection risk review", "server_side_template_expression_injection_risk_review", 1},
		{"XPath注入攻击", "XPath查询表达式注入风险核查", "XPath查询表达式注入风险核查", 1},
		{"XPath injection attack", "XPath query expression injection risk review", "XPath_query_expression_injection_risk_review", 1},
	}
	for _, in := range []string{
		"expression language injection 表达式语言注入 EL SpEL OGNL UEL",
		"ExpressionFactory ValueExpression ELContext CWE-917 expression evaluate",
		"expression language injection prevention validateExpression sandbox",
		"expression language evaluation injection risk review 表达式语言求值注入风险核查",
		"表达式语言注入防护 表达式 语言注入攻击 表达式语言注入 攻击",
		"EL injection attack SpEL injection attack OGNL injection attack UEL injection attack",
		"expression injection attack language injection attack",
		"expression-language-injection-attack expression_language_injection_attack",
		"expression language injection attacks",
		"expression\tlanguage injection attack",
		"expression  language injection attack",
		"expression\nlanguage injection attack",
		"expression language injection\tattack",
		"xexpression language injection attack",
		"expression language injection attackx",
		"2expression language injection attack",
		"expression language injection attack2",
		"_expression language injection attack",
		"expression language injection attack_suffix",
		"前expression language injection attack",
		"expression language injection attack后",
		`import "expression language injection attack"`,
		`import "表达式语言注入攻击"`,
		"https://example.invalid/expression%20language%20injection%20attack",
		"https://example.invalid/表达式语言注入攻击",
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
