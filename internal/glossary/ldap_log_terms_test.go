package glossary

import (
	"strings"
	"testing"
)

func TestLDAPLogRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"LDAP注入攻击", "LDAP查询注入风险核查", "LDAP查询注入风险核查", 1},
		{"日志伪造攻击", "日志记录伪造风险核查", "日志记录伪造风险核查", 1},
		{"LDAP injection attack", "LDAP query injection risk review", "LDAP_query_injection_risk_review", 1},
		{"log forging attack", "log record forgery risk review", "log_record_forgery_risk_review", 1},
		{"ldap injection attack", "LDAP query injection risk review", "LDAP_query_injection_risk_review", 1},
		{"lDaP iNjEcTiOn AtTaCk", "LDAP query injection risk review", "LDAP_query_injection_risk_review", 1},
		{"LDAP INJECTION ATTACK", "LDAP QUERY INJECTION RISK REVIEW", "LDAP_QUERY_INJECTION_RISK_REVIEW", 1},
		{"Log forging attack", "Log record forgery risk review", "Log_record_forgery_risk_review", 1},
		{"LOG FORGING ATTACK", "LOG RECORD FORGERY RISK REVIEW", "LOG_RECORD_FORGERY_RISK_REVIEW", 1},
		{"LDAP注入攻击; 日志伪造攻击", "LDAP查询注入风险核查; 日志记录伪造风险核查", "LDAP查询注入风险核查; 日志记录伪造风险核查", 2},
		{"XPath注入攻击", "XPath查询表达式注入风险核查", "XPath查询表达式注入风险核查", 1},
		{"XPath injection attack", "XPath query expression injection risk review", "XPath_query_expression_injection_risk_review", 1},
	}
	for _, in := range []string{
		"LDAP injection; log forging; LDAP注入; 日志伪造",
		"LDAP log logging 日志 注入 伪造 CWE-90 CWE-117",
		"cn uid objectClass distinguishedName INFO WARN ERROR DEBUG",
		"ldap.Search SearchRequest logger.Printf logging.getLogger",
		"LDAP注入防护 日志伪造防护",
		"ldap注入攻击 LDAP 注入攻击 日志 伪造攻击",
		"log injection attack",
		"LDAP_injection_attack log_forging_attack",
		"LDAP\tinjection attack log\tforging attack",
		"xLDAP injection attack",
		"LDAP injection attack2",
		"_LDAP injection attack",
		"log forging attacks",
		"2log forging attack",
		"log forging attack_suffix",
		`import "LDAP injection attack"`,
		`import "log forging attack"`,
		"https://example.invalid/LDAP注入攻击",
		"https://example.invalid/日志伪造攻击",
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
