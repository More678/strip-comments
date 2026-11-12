package glossary

import (
	"strings"
	"testing"
)

func TestNullPointerRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"空指针解引用漏洞", "空指针解引用风险核查", "空指针解引用风险核查", 1},
		{"null pointer dereference vulnerability", "null pointer dereference risk review", "null_pointer_dereference_risk_review", 1},
		{"Null pointer dereference vulnerability", "Null pointer dereference risk review", "Null_pointer_dereference_risk_review", 1},
		{"nUlL pOiNtEr DeReFeReNcE vUlNeRaBiLiTy", "null pointer dereference risk review", "null_pointer_dereference_risk_review", 1},
		{"NULL POINTER DEREFERENCE VULNERABILITY", "NULL POINTER DEREFERENCE RISK REVIEW", "NULL_POINTER_DEREFERENCE_RISK_REVIEW", 1},
		{"空指针解引用漏洞; null pointer dereference vulnerability", "空指针解引用风险核查; null pointer dereference risk review", "空指针解引用风险核查; null_pointer_dereference_risk_review", 2},
		{"类型混淆漏洞", "类型混淆风险核查", "类型混淆风险核查", 1},
		{"type confusion vulnerability", "type confusion risk review", "type_confusion_risk_review", 1},
		{"内存释放后使用漏洞", "内存释放后使用风险核查", "内存释放后使用风险核查", 1},
		{"use-after-free vulnerability", "use-after-free risk review", "use_after_free_risk_review", 1},
	}
	for _, in := range []string{
		"null pointer dereference 空指针解引用 NULL null nil nullptr 指针 CWE-476",
		"NullPointerException ArgumentNullException is_null",
		"空指针解引用防护 NullPointerProtection",
		"空指针 解引用漏洞 空引用漏洞",
		"null-pointer-dereference vulnerability null_pointer_dereference_vulnerability",
		"null\tpointer dereference vulnerability",
		"xnull pointer dereference vulnerability",
		"null pointer dereference vulnerabilityx",
		"2null pointer dereference vulnerability",
		"null pointer dereference vulnerability2",
		"_null pointer dereference vulnerability",
		"null pointer dereference vulnerability_suffix",
		`import "null pointer dereference vulnerability"`,
		`import "空指针解引用漏洞"`,
		"https://example.invalid/空指针解引用漏洞",
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
