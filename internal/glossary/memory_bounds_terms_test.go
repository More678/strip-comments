package glossary

import (
	"strings"
	"testing"
)

func TestMemoryBoundsRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"内存越界读取漏洞", "内存越界读取风险核查", "内存越界读取风险核查", 1},
		{"内存越界写入漏洞", "内存越界写入风险核查", "内存越界写入风险核查", 1},
		{"out-of-bounds read vulnerability", "out-of-bounds read risk review", "out_of_bounds_read_risk_review", 1},
		{"out-of-bounds write vulnerability", "out-of-bounds write risk review", "out_of_bounds_write_risk_review", 1},
		{"Out-of-bounds read vulnerability", "Out-of-bounds read risk review", "Out_of_bounds_read_risk_review", 1},
		{"Out-of-bounds write vulnerability", "Out-of-bounds write risk review", "Out_of_bounds_write_risk_review", 1},
		{"oUt-Of-BoUnDs ReAd VuLnErAbIlItY", "out-of-bounds read risk review", "out_of_bounds_read_risk_review", 1},
		{"oUt-Of-BoUnDs WrItE vUlNeRaBiLiTy", "out-of-bounds write risk review", "out_of_bounds_write_risk_review", 1},
		{"OUT-OF-BOUNDS READ VULNERABILITY", "OUT-OF-BOUNDS READ RISK REVIEW", "OUT_OF_BOUNDS_READ_RISK_REVIEW", 1},
		{"OUT-OF-BOUNDS WRITE VULNERABILITY", "OUT-OF-BOUNDS WRITE RISK REVIEW", "OUT_OF_BOUNDS_WRITE_RISK_REVIEW", 1},
		{"内存越界读取漏洞; 内存越界写入漏洞", "内存越界读取风险核查; 内存越界写入风险核查", "内存越界读取风险核查; 内存越界写入风险核查", 2},
		{"out-of-bounds read vulnerability; out-of-bounds write vulnerability", "out-of-bounds read risk review; out-of-bounds write risk review", "out_of_bounds_read_risk_review; out_of_bounds_write_risk_review", 2},
		{"内存释放后使用漏洞", "内存释放后使用风险核查", "内存释放后使用风险核查", 1},
		{"内存重复释放漏洞", "内存重复释放风险核查", "内存重复释放风险核查", 1},
		{"use-after-free vulnerability", "use-after-free risk review", "use_after_free_risk_review", 1},
		{"double free vulnerability", "double free risk review", "double_free_risk_review", 1},
	}
	for _, in := range []string{
		"out-of-bounds read out-of-bounds write 越界读取 越界写入",
		"buffer index read write OOB 内存 CWE-125 CWE-787",
		"memcpy memmove fread fwrite ReadFile WriteFile",
		"内存越界读取防护 内存越界写入防护 MemoryBoundsProtection",
		"内存 越界读取漏洞 内存 越界写入漏洞 越界读取漏洞 越界写入漏洞",
		"out of bounds read vulnerability out of bounds write vulnerability",
		"out_of_bounds_read_vulnerability out_of_bounds_write_vulnerability",
		"out-of-bounds\tread vulnerability out-of-bounds\twrite vulnerability",
		"xout-of-bounds read vulnerability",
		"out-of-bounds read vulnerabilityx",
		"2out-of-bounds read vulnerability",
		"out-of-bounds read vulnerability2",
		"_out-of-bounds read vulnerability",
		"out-of-bounds read vulnerability_suffix",
		"xout-of-bounds write vulnerability",
		"out-of-bounds write vulnerabilityx",
		"2out-of-bounds write vulnerability",
		"out-of-bounds write vulnerability2",
		"_out-of-bounds write vulnerability",
		"out-of-bounds write vulnerability_suffix",
		`import "out-of-bounds read vulnerability"`,
		`import "out-of-bounds write vulnerability"`,
		`import "内存越界读取漏洞"`,
		`import "内存越界写入漏洞"`,
		"https://example.invalid/内存越界读取漏洞",
		"https://example.invalid/内存越界写入漏洞",
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
