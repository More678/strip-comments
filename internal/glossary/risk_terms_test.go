package glossary

import "testing"

func TestTicketAndCredentialRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct{ in, want string }{
		{"黄金票据攻击", "Kerberos票据伪造风险核查"},
		{"票据传递攻击", "Kerberos票据复用风险核查"},
		{"凭据转储攻击", "认证材料暴露风险核查"},
		{"golden ticket attack", "Kerberos ticket forgery risk review"},
		{"pass the ticket", "Kerberos ticket reuse risk review"},
		{"OS credential dumping", "credential exposure risk assessment"},
		{"PASS THE TICKET", "KERBEROS TICKET REUSE RISK REVIEW"},
		{"黄金票据", "黄金票据"},
		{"会计票据传递", "会计票据传递"},
		{"core dumping", "core dumping"},
		{"my_pass_the_ticket", "my_pass_the_ticket"},
		{"pass the ticket2", "pass the ticket2"},
		{"golden ticket attacks", "golden ticket attacks"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			opt := Options{Comments: []Region{{Start: 0, End: len(tc.in), Comment: true}}}
			out, _ := tb.Apply([]byte(tc.in), opt)
			if string(out) != tc.want {
				t.Fatalf("got %q, want %q", out, tc.want)
			}
			opt.Comments[0].End = len(out)
			again, hits := tb.Apply(out, opt)
			if string(again) != string(out) || len(hits) != 0 {
				t.Fatalf("second pass changed output: %q (%+v)", again, hits)
			}
		})
	}
}
