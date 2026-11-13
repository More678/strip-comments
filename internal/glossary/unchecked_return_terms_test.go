package glossary

import (
	"strings"
	"testing"
)

func TestUncheckedReturnRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"未检查返回值漏洞", "未检查返回值风险核查", "未检查返回值风险核查", 1},
		{"unchecked return value vulnerability", "unchecked return value risk review", "unchecked_return_value_risk_review", 1},
		{"Unchecked return value vulnerability", "Unchecked return value risk review", "Unchecked_return_value_risk_review", 1},
		{"uNcHeCkEd ReTuRn VaLuE VuLnErAbIlItY", "unchecked return value risk review", "unchecked_return_value_risk_review", 1},
		{"UNCHECKED RETURN VALUE VULNERABILITY", "UNCHECKED RETURN VALUE RISK REVIEW", "UNCHECKED_RETURN_VALUE_RISK_REVIEW", 1},
		{"未检查返回值漏洞; unchecked return value vulnerability", "未检查返回值风险核查; unchecked return value risk review", "未检查返回值风险核查; unchecked_return_value_risk_review", 2},
		{"review:\nunchecked return value vulnerability", "review:\nunchecked return value risk review", "review:\nunchecked_return_value_risk_review", 1},
		{"(unchecked return value vulnerability).", "(unchecked return value risk review).", "(unchecked_return_value_risk_review).", 1},
		{"空指针解引用漏洞", "空指针解引用风险核查", "空指针解引用风险核查", 1},
		{"null pointer dereference vulnerability", "null pointer dereference risk review", "null_pointer_dereference_risk_review", 1},
		{"未初始化变量使用漏洞", "未初始化变量使用风险核查", "未初始化变量使用风险核查", 1},
		{"uninitialized variable use vulnerability", "uninitialized variable use risk review", "uninitialized_variable_use_risk_review", 1},
	}
	for _, in := range []string{
		"unchecked return value 未检查返回值 return result errno Result",
		"nodiscard warn_unused_result errors.Is CWE-252 CWE-253",
		"unchecked return value prevention validateResult handleError",
		"unchecked return value risk review 未检查返回值风险核查",
		"未检查返回值防护 未检查 返回值漏洞 未检查返回值 漏洞",
		"ignored return value vulnerability 返回值未检查漏洞",
		"return value vulnerability incorrect return value check",
		"unchecked-return-value-vulnerability unchecked_return_value_vulnerability",
		"unchecked return value vulnerabilities",
		"unchecked\treturn value vulnerability",
		"unchecked  return value vulnerability",
		"unchecked\nreturn value vulnerability",
		"unchecked return value\tvulnerability",
		"xunchecked return value vulnerability",
		"unchecked return value vulnerabilityx",
		"2unchecked return value vulnerability",
		"unchecked return value vulnerability2",
		"_unchecked return value vulnerability",
		"unchecked return value vulnerability_suffix",
		"前unchecked return value vulnerability",
		"unchecked return value vulnerability后",
		`import "unchecked return value vulnerability"`,
		`import "未检查返回值漏洞"`,
		"https://example.invalid/unchecked%20return%20value%20vulnerability",
		"https://example.invalid/未检查返回值漏洞",
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
