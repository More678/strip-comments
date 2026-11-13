package glossary

import (
	"strings"
	"testing"
)

func TestXQueryRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"XQuery注入攻击", "XQuery查询结构注入风险核查", "XQuery查询结构注入风险核查", 1},
		{"XQuery injection attack", "XQuery query structure injection risk review", "XQuery_query_structure_injection_risk_review", 1},
		{"xquery injection attack", "XQuery query structure injection risk review", "XQuery_query_structure_injection_risk_review", 1},
		{"Xquery injection attack", "XQuery query structure injection risk review", "XQuery_query_structure_injection_risk_review", 1},
		{"xQuErY InJeCtIoN AtTaCk", "XQuery query structure injection risk review", "XQuery_query_structure_injection_risk_review", 1},
		{"XQUERY INJECTION ATTACK", "XQUERY QUERY STRUCTURE INJECTION RISK REVIEW", "XQUERY_QUERY_STRUCTURE_INJECTION_RISK_REVIEW", 1},
		{"XQuery注入攻击; XQuery injection attack", "XQuery查询结构注入风险核查; XQuery query structure injection risk review", "XQuery查询结构注入风险核查; XQuery_query_structure_injection_risk_review", 2},
		{"review:\nXQuery injection attack", "review:\nXQuery query structure injection risk review", "review:\nXQuery_query_structure_injection_risk_review", 1},
		{"(XQuery injection attack).", "(XQuery query structure injection risk review).", "(XQuery_query_structure_injection_risk_review).", 1},
		{"XPath注入攻击", "XPath查询表达式注入风险核查", "XPath查询表达式注入风险核查", 1},
		{"XPath injection attack", "XPath query expression injection risk review", "XPath_query_expression_injection_risk_review", 1},
		{"表达式语言注入攻击", "表达式语言求值注入风险核查", "表达式语言求值注入风险核查", 1},
		{"expression language injection attack", "expression language evaluation injection risk review", "expression_language_evaluation_injection_risk_review", 1},
	}
	for _, in := range []string{
		"XQuery XQuery injection XQuery注入 XPath XML FLWOR query",
		"XQConnection XQPreparedExpression bindString CWE-652",
		"XQuery injection prevention validateXQuery bindVariable",
		"XQuery query structure injection risk review XQuery查询结构注入风险核查",
		"XQuery注入防护 xquery注入攻击 Xquery注入攻击 XQUERY注入攻击 XQuery 注入攻击 XQuery注入 攻击",
		"XQuery-injection-attack XQuery_injection_attack",
		"XQuery injection attacks",
		"XQuery\tinjection attack",
		"XQuery  injection attack",
		"XQuery\ninjection attack",
		"XQuery injection\tattack",
		"xXQuery injection attack",
		"XQuery injection attackx",
		"2XQuery injection attack",
		"XQuery injection attack2",
		"_XQuery injection attack",
		"XQuery injection attack_suffix",
		"前XQuery injection attack",
		"XQuery injection attack后",
		`import "XQuery injection attack"`,
		`import "XQuery注入攻击"`,
		"https://example.invalid/XQuery%20injection%20attack",
		"https://example.invalid/XQuery注入攻击",
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
