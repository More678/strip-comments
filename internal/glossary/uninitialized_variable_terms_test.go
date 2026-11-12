package glossary

import (
	"strings"
	"testing"
)

func TestUninitializedVariableRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"未初始化变量使用漏洞", "未初始化变量使用风险核查", "未初始化变量使用风险核查", 1},
		{"uninitialized variable use vulnerability", "uninitialized variable use risk review", "uninitialized_variable_use_risk_review", 1},
		{"Uninitialized variable use vulnerability", "Uninitialized variable use risk review", "Uninitialized_variable_use_risk_review", 1},
		{"uNiNiTiAlIzEd VaRiAbLe UsE vUlNeRaBiLiTy", "uninitialized variable use risk review", "uninitialized_variable_use_risk_review", 1},
		{"UNINITIALIZED VARIABLE USE VULNERABILITY", "UNINITIALIZED VARIABLE USE RISK REVIEW", "UNINITIALIZED_VARIABLE_USE_RISK_REVIEW", 1},
		{"未初始化变量使用漏洞; uninitialized variable use vulnerability", "未初始化变量使用风险核查; uninitialized variable use risk review", "未初始化变量使用风险核查; uninitialized_variable_use_risk_review", 2},
		{"空指针解引用漏洞", "空指针解引用风险核查", "空指针解引用风险核查", 1},
		{"null pointer dereference vulnerability", "null pointer dereference risk review", "null_pointer_dereference_risk_review", 1},
		{"内存释放后使用漏洞", "内存释放后使用风险核查", "内存释放后使用风险核查", 1},
		{"use-after-free vulnerability", "use-after-free risk review", "use_after_free_risk_review", 1},
	}
	for _, in := range []string{
		"Use of Uninitialized Variable uninitialized variable 未初始化变量 初始化 variable default zero value",
		"CWE-457 CWE-908 MaybeUninit Default init",
		"未初始化变量使用防护 UninitializedVariableProtection",
		"未初始化 变量使用漏洞 未初始化资源使用漏洞",
		"uninitialized variable vulnerability use of uninitialized variable vulnerability",
		"uninitialized-variable-use vulnerability uninitialized_variable_use_vulnerability",
		"uninitialized\tvariable use vulnerability",
		"xuninitialized variable use vulnerability",
		"uninitialized variable use vulnerabilityx",
		"2uninitialized variable use vulnerability",
		"uninitialized variable use vulnerability2",
		"_uninitialized variable use vulnerability",
		"uninitialized variable use vulnerability_suffix",
		`import "uninitialized variable use vulnerability"`,
		`import "未初始化变量使用漏洞"`,
		"https://example.invalid/未初始化变量使用漏洞",
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
