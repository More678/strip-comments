package glossary

import (
	"strings"
	"testing"
)

func TestSSIInjectionRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"SSI注入攻击", "SSI服务端包含指令注入风险核查", "SSI服务端包含指令注入风险核查", 1},
		{"SSI injection attack", "SSI server-side include directive injection risk review", "SSI_server_side_include_directive_injection_risk_review", 1},
		{"ssi injection attack", "SSI server-side include directive injection risk review", "SSI_server_side_include_directive_injection_risk_review", 1},
		{"Ssi injection attack", "SSI server-side include directive injection risk review", "SSI_server_side_include_directive_injection_risk_review", 1},
		{"sSi InJeCtIoN AtTaCk", "SSI server-side include directive injection risk review", "SSI_server_side_include_directive_injection_risk_review", 1},
		{"SSI INJECTION ATTACK", "SSI SERVER-SIDE INCLUDE DIRECTIVE INJECTION RISK REVIEW", "SSI_SERVER_SIDE_INCLUDE_DIRECTIVE_INJECTION_RISK_REVIEW", 1},
		{"SSI注入攻击; SSI injection attack", "SSI服务端包含指令注入风险核查; SSI server-side include directive injection risk review", "SSI服务端包含指令注入风险核查; SSI_server_side_include_directive_injection_risk_review", 2},
		{"review:\nSSI injection attack", "review:\nSSI server-side include directive injection risk review", "review:\nSSI_server_side_include_directive_injection_risk_review", 1},
		{"(SSI injection attack).", "(SSI server-side include directive injection risk review).", "(SSI_server_side_include_directive_injection_risk_review).", 1},
		{"服务端模板注入攻击", "服务端模板表达式注入风险核查", "服务端模板表达式注入风险核查", 1},
		{"server-side template injection attack", "server-side template expression injection risk review", "server_side_template_expression_injection_risk_review", 1},
		{"跨站脚本包含攻击", "跨站脚本包含数据泄露风险核查", "跨站脚本包含数据泄露风险核查", 1},
		{"cross site script inclusion attack", "cross site script inclusion data exposure risk review", "cross_site_script_inclusion_data_exposure_risk_review", 1},
	}
	for _, in := range []string{
		"SSI SSI injection Server-Side Includes SSI注入 shtml include",
		"Options IncludesNOEXEC AddOutputFilter INCLUDES mod_include WSTG-INPV-08",
		"SSI injection prevention disableSSI validateSSI",
		"SSI server-side include directive injection risk review SSI服务端包含指令注入风险核查",
		"SSI注入防护 ssi注入攻击 Ssi注入攻击 SSI 注入攻击 SSI注入 攻击",
		"server-side includes injection attack 服务器端包含注入攻击",
		"SSI-injection-attack SSI_injection_attack",
		"SSI injection attacks",
		"SSI\tinjection attack",
		"SSI  injection attack",
		"SSI\ninjection attack",
		"SSI injection\tattack",
		"xSSI injection attack",
		"SSI injection attackx",
		"2SSI injection attack",
		"SSI injection attack2",
		"_SSI injection attack",
		"SSI injection attack_suffix",
		"前SSI injection attack",
		"SSI injection attack后",
		"XSSI injection attack CSSI injection attack",
		`import "SSI injection attack"`,
		`import "SSI注入攻击"`,
		"https://example.invalid/SSI%20injection%20attack",
		"https://example.invalid/SSI注入攻击",
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
