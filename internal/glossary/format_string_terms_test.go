package glossary

import (
	"strings"
	"testing"
)

func TestFormatStringRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"外部可控格式字符串漏洞", "外部可控格式字符串风险核查", "外部可控格式字符串风险核查", 1},
		{"externally controlled format string vulnerability", "externally controlled format string risk review", "externally_controlled_format_string_risk_review", 1},
		{"Externally controlled format string vulnerability", "Externally controlled format string risk review", "Externally_controlled_format_string_risk_review", 1},
		{"eXtErNaLlY cOnTrOlLeD fOrMaT sTrInG vUlNeRaBiLiTy", "externally controlled format string risk review", "externally_controlled_format_string_risk_review", 1},
		{"EXTERNALLY CONTROLLED FORMAT STRING VULNERABILITY", "EXTERNALLY CONTROLLED FORMAT STRING RISK REVIEW", "EXTERNALLY_CONTROLLED_FORMAT_STRING_RISK_REVIEW", 1},
		{"外部可控格式字符串漏洞; externally controlled format string vulnerability", "外部可控格式字符串风险核查; externally controlled format string risk review", "外部可控格式字符串风险核查; externally_controlled_format_string_risk_review", 2},
		{"服务端模板注入攻击", "服务端模板表达式注入风险核查", "服务端模板表达式注入风险核查", 1},
		{"日志伪造攻击", "日志记录伪造风险核查", "日志记录伪造风险核查", 1},
	}
	for _, in := range []string{
		"externally controlled format string 外部可控格式字符串",
		"format string 格式字符串 格式化字符串 CWE-134",
		"printf sprintf fmt.Printf String.Format",
		"外部可控格式字符串防护 FormatStringProtection",
		"外部可控 格式字符串漏洞 外部可控格式化字符串漏洞",
		"format string vulnerability externally-controlled format string vulnerability",
		"externally_controlled_format_string_vulnerability",
		"externally\tcontrolled format string vulnerability",
		"xexternally controlled format string vulnerability",
		"externally controlled format string vulnerabilityx",
		"2externally controlled format string vulnerability",
		"externally controlled format string vulnerability2",
		"_externally controlled format string vulnerability",
		"externally controlled format string vulnerability_suffix",
		`import "externally controlled format string vulnerability"`,
		`import "外部可控格式字符串漏洞"`,
		"https://example.invalid/外部可控格式字符串漏洞",
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
