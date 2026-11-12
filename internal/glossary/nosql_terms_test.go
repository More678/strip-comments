package glossary

import (
	"strings"
	"testing"
)

func TestNoSQLRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"NoSQL注入攻击", "NoSQL查询逻辑注入风险核查", "NoSQL查询逻辑注入风险核查", 1},
		{"NoSQL injection attack", "NoSQL query logic injection risk review", "NoSQL_query_logic_injection_risk_review", 1},
		{"nosql injection attack", "NoSQL query logic injection risk review", "NoSQL_query_logic_injection_risk_review", 1},
		{"Nosql injection attack", "NoSQL query logic injection risk review", "NoSQL_query_logic_injection_risk_review", 1},
		{"nOsQl InJeCtIoN aTtAcK", "NoSQL query logic injection risk review", "NoSQL_query_logic_injection_risk_review", 1},
		{"NOSQL INJECTION ATTACK", "NOSQL QUERY LOGIC INJECTION RISK REVIEW", "NOSQL_QUERY_LOGIC_INJECTION_RISK_REVIEW", 1},
		{"NoSQL注入攻击; NoSQL injection attack", "NoSQL查询逻辑注入风险核查; NoSQL query logic injection risk review", "NoSQL查询逻辑注入风险核查; NoSQL_query_logic_injection_risk_review", 2},
		{"LDAP注入攻击", "LDAP查询注入风险核查", "LDAP查询注入风险核查", 1},
		{"LDAP injection attack", "LDAP query injection risk review", "LDAP_query_injection_risk_review", 1},
		{"XPath注入攻击", "XPath查询表达式注入风险核查", "XPath查询表达式注入风险核查", 1},
		{"XPath injection attack", "XPath query expression injection risk review", "XPath_query_expression_injection_risk_review", 1},
	}
	for _, in := range []string{
		"NoSQL injection NoSQL注入 SQL injection SQL注入 NoSQL 查询 注入",
		"MongoDB CouchDB Redis DynamoDB CWE-943",
		"find findOne aggregate MongoClient collection.Find $eq $ne $where",
		"NoSQL注入防护 nosql注入攻击 NOSQL注入攻击 NoSQL 注入攻击",
		"No SQL injection attack No-SQL injection attack",
		"NoSQL-injection-attack NoSQL_injection_attack",
		"NoSQL\tinjection attack",
		"xNoSQL injection attack",
		"NoSQL injection attacks",
		"2NoSQL injection attack",
		"NoSQL injection attack2",
		"_NoSQL injection attack",
		"NoSQL injection attack_suffix",
		`import "NoSQL injection attack"`,
		`import "NoSQL注入攻击"`,
		"https://example.invalid/NoSQL注入攻击",
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
