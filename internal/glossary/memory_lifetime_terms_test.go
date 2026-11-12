package glossary

import (
	"strings"
	"testing"
)

func TestMemoryLifetimeRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"内存释放后使用漏洞", "内存释放后使用风险核查", "内存释放后使用风险核查", 1},
		{"内存重复释放漏洞", "内存重复释放风险核查", "内存重复释放风险核查", 1},
		{"use-after-free vulnerability", "use-after-free risk review", "use_after_free_risk_review", 1},
		{"double free vulnerability", "double free risk review", "double_free_risk_review", 1},
		{"Use-after-free vulnerability", "Use-after-free risk review", "Use_after_free_risk_review", 1},
		{"Double free vulnerability", "Double free risk review", "Double_free_risk_review", 1},
		{"uSe-AfTeR-fReE vUlNeRaBiLiTy", "use-after-free risk review", "use_after_free_risk_review", 1},
		{"dOuBlE fReE vUlNeRaBiLiTy", "double free risk review", "double_free_risk_review", 1},
		{"USE-AFTER-FREE VULNERABILITY", "USE-AFTER-FREE RISK REVIEW", "USE_AFTER_FREE_RISK_REVIEW", 1},
		{"DOUBLE FREE VULNERABILITY", "DOUBLE FREE RISK REVIEW", "DOUBLE_FREE_RISK_REVIEW", 1},
		{"内存释放后使用漏洞; 内存重复释放漏洞", "内存释放后使用风险核查; 内存重复释放风险核查", "内存释放后使用风险核查; 内存重复释放风险核查", 2},
		{"use-after-free vulnerability; double free vulnerability", "use-after-free risk review; double free risk review", "use_after_free_risk_review; double_free_risk_review", 2},
	}
	for _, in := range []string{
		"use-after-free double free UAF free malloc delete 内存 释放",
		"CWE-415 CWE-416 calloc realloc std::free unique_ptr drop",
		"内存释放后使用防护 内存重复释放防护 MemoryLifetimeProtection",
		"内存 释放后使用漏洞 内存 重复释放漏洞 释放后使用漏洞 重复释放漏洞",
		"use after free vulnerability double-free vulnerability",
		"use_after_free_vulnerability double_free_vulnerability",
		"use-after-free\tvulnerability double\tfree vulnerability",
		"xuse-after-free vulnerability",
		"use-after-free vulnerabilityx",
		"2use-after-free vulnerability",
		"use-after-free vulnerability2",
		"_use-after-free vulnerability",
		"use-after-free vulnerability_suffix",
		"xdouble free vulnerability",
		"double free vulnerabilityx",
		"2double free vulnerability",
		"double free vulnerability2",
		"_double free vulnerability",
		"double free vulnerability_suffix",
		`import "use-after-free vulnerability"`,
		`import "double free vulnerability"`,
		`import "内存释放后使用漏洞"`,
		`import "内存重复释放漏洞"`,
		"https://example.invalid/内存释放后使用漏洞",
		"https://example.invalid/内存重复释放漏洞",
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
