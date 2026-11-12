package glossary

import (
	"strings"
	"testing"
)

func TestIntegerRangeRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"整数溢出漏洞", "整数溢出风险核查", "整数溢出风险核查", 1},
		{"整数下溢漏洞", "整数下溢风险核查", "整数下溢风险核查", 1},
		{"integer overflow vulnerability", "integer overflow risk review", "integer_overflow_risk_review", 1},
		{"integer underflow vulnerability", "integer underflow risk review", "integer_underflow_risk_review", 1},
		{"Integer overflow vulnerability", "Integer overflow risk review", "Integer_overflow_risk_review", 1},
		{"Integer underflow vulnerability", "Integer underflow risk review", "Integer_underflow_risk_review", 1},
		{"iNtEgEr OvErFlOw VuLnErAbIlItY", "integer overflow risk review", "integer_overflow_risk_review", 1},
		{"iNtEgEr UnDeRfLoW vUlNeRaBiLiTy", "integer underflow risk review", "integer_underflow_risk_review", 1},
		{"INTEGER OVERFLOW VULNERABILITY", "INTEGER OVERFLOW RISK REVIEW", "INTEGER_OVERFLOW_RISK_REVIEW", 1},
		{"INTEGER UNDERFLOW VULNERABILITY", "INTEGER UNDERFLOW RISK REVIEW", "INTEGER_UNDERFLOW_RISK_REVIEW", 1},
		{"整数溢出漏洞; 整数下溢漏洞", "整数溢出风险核查; 整数下溢风险核查", "整数溢出风险核查; 整数下溢风险核查", 2},
		{"integer overflow vulnerability; integer underflow vulnerability", "integer overflow risk review; integer underflow risk review", "integer_overflow_risk_review; integer_underflow_risk_review", 2},
		{"内存越界读取漏洞", "内存越界读取风险核查", "内存越界读取风险核查", 1},
		{"内存越界写入漏洞", "内存越界写入风险核查", "内存越界写入风险核查", 1},
		{"out-of-bounds read vulnerability", "out-of-bounds read risk review", "out_of_bounds_read_risk_review", 1},
		{"out-of-bounds write vulnerability", "out-of-bounds write risk review", "out_of_bounds_write_risk_review", 1},
	}
	for _, in := range []string{
		"integer overflow integer underflow 整数溢出 整数下溢 wraparound",
		"overflow underflow 整数 算术 CWE-190 CWE-191",
		"int32 uint64 checked unchecked wrapping_add checked_sub Math.addExact",
		"整数溢出防护 整数下溢防护 IntegerRangeProtection",
		"整数 溢出漏洞 整数 下溢漏洞",
		"integer-overflow vulnerability integer-underflow vulnerability",
		"integer_overflow_vulnerability integer_underflow_vulnerability",
		"integer\toverflow vulnerability integer\tunderflow vulnerability",
		"xinteger overflow vulnerability",
		"integer overflow vulnerabilityx",
		"2integer overflow vulnerability",
		"integer overflow vulnerability2",
		"_integer overflow vulnerability",
		"integer overflow vulnerability_suffix",
		"xinteger underflow vulnerability",
		"integer underflow vulnerabilityx",
		"2integer underflow vulnerability",
		"integer underflow vulnerability2",
		"_integer underflow vulnerability",
		"integer underflow vulnerability_suffix",
		`import "integer overflow vulnerability"`,
		`import "integer underflow vulnerability"`,
		`import "整数溢出漏洞"`,
		`import "整数下溢漏洞"`,
		"https://example.invalid/整数溢出漏洞",
		"https://example.invalid/整数下溢漏洞",
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
