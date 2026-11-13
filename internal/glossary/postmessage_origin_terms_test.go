package glossary

import (
	"strings"
	"testing"
)

func TestPostMessageOriginRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"postMessage来源验证漏洞", "postMessage来源验证风险核查", "postMessage来源验证风险核查", 1},
		{"postMessage origin validation vulnerability", "postMessage origin validation risk review", "postMessage_origin_validation_risk_review", 1},
		{"postmessage origin validation vulnerability", "postMessage origin validation risk review", "postMessage_origin_validation_risk_review", 1},
		{"Postmessage origin validation vulnerability", "PostMessage origin validation risk review", "PostMessage_origin_validation_risk_review", 1},
		{"pOsTmEsSaGe OrIgIn VaLiDaTiOn VuLnErAbIlItY", "postMessage origin validation risk review", "postMessage_origin_validation_risk_review", 1},
		{"POSTMESSAGE ORIGIN VALIDATION VULNERABILITY", "POSTMESSAGE ORIGIN VALIDATION RISK REVIEW", "POSTMESSAGE_ORIGIN_VALIDATION_RISK_REVIEW", 1},
		{"postMessage来源验证漏洞; postMessage origin validation vulnerability", "postMessage来源验证风险核查; postMessage origin validation risk review", "postMessage来源验证风险核查; postMessage_origin_validation_risk_review", 2},
		{"review:\npostMessage origin validation vulnerability", "review:\npostMessage origin validation risk review", "review:\npostMessage_origin_validation_risk_review", 1},
		{"(postMessage origin validation vulnerability).", "(postMessage origin validation risk review).", "(postMessage_origin_validation_risk_review).", 1},
		{"CORS错误配置漏洞", "CORS错误配置风险核查", "CORS错误配置风险核查", 1},
		{"CORS misconfiguration vulnerability", "CORS misconfiguration risk review", "CORS_misconfiguration_risk_review", 1},
		{"reverse tabnabbing attack", "reverse tabnabbing risk review", "reverse_tabnabbing_risk_review", 1},
	}
	for _, in := range []string{
		"postMessage postMessage origin validation postMessage来源验证 MessageEvent",
		"event.origin event.source event.data targetOrigin addEventListener WSTG-CLNT-11",
		"postMessage origin validation prevention validateMessageOrigin allowlist",
		"postMessage origin validation risk review postMessage来源验证风险核查",
		"postMessage来源验证防护 postmessage来源验证漏洞 PostMessage来源验证漏洞 postMessage 来源验证漏洞 postMessage来源验证 漏洞",
		"origin validation vulnerability 来源验证漏洞",
		"postMessage origin verification vulnerability",
		"postMessage-origin-validation-vulnerability postMessage_origin_validation_vulnerability",
		"postMessage origin validation vulnerabilities",
		"postMessage\torigin validation vulnerability",
		"postMessage  origin validation vulnerability",
		"postMessage\norigin validation vulnerability",
		"postMessage origin validation\tvulnerability",
		"xpostMessage origin validation vulnerability",
		"postMessage origin validation vulnerabilityx",
		"2postMessage origin validation vulnerability",
		"postMessage origin validation vulnerability2",
		"_postMessage origin validation vulnerability",
		"postMessage origin validation vulnerability_suffix",
		"前postMessage origin validation vulnerability",
		"postMessage origin validation vulnerability后",
		`import "postMessage origin validation vulnerability"`,
		`import "postMessage来源验证漏洞"`,
		"https://example.invalid/postMessage%20origin%20validation%20vulnerability",
		"https://example.invalid/postMessage来源验证漏洞",
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
