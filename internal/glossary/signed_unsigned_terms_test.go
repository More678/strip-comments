package glossary

import (
	"strings"
	"testing"
)

func TestSignedUnsignedRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"signed to unsigned conversion vulnerability", "signed to unsigned conversion risk review", "signed_to_unsigned_conversion_risk_review", 1},
		{"Signed to unsigned conversion vulnerability", "Signed to unsigned conversion risk review", "Signed_to_unsigned_conversion_risk_review", 1},
		{"sIgNeD To UnSiGnEd CoNvErSiOn VuLnErAbIlItY", "signed to unsigned conversion risk review", "signed_to_unsigned_conversion_risk_review", 1},
		{"SIGNED TO UNSIGNED CONVERSION VULNERABILITY", "SIGNED TO UNSIGNED CONVERSION RISK REVIEW", "SIGNED_TO_UNSIGNED_CONVERSION_RISK_REVIEW", 1},
		{"signed to unsigned conversion vulnerability; signed to unsigned conversion vulnerability", "signed to unsigned conversion risk review; signed to unsigned conversion risk review", "signed_to_unsigned_conversion_risk_review; signed_to_unsigned_conversion_risk_review", 2},
		{"review:\nsigned to unsigned conversion vulnerability", "review:\nsigned to unsigned conversion risk review", "review:\nsigned_to_unsigned_conversion_risk_review", 1},
		{"(signed to unsigned conversion vulnerability).", "(signed to unsigned conversion risk review).", "(signed_to_unsigned_conversion_risk_review).", 1},
		{"整数溢出漏洞", "整数溢出风险核查", "整数溢出风险核查", 1},
		{"整数下溢漏洞", "整数下溢风险核查", "整数下溢风险核查", 1},
		{"integer overflow vulnerability", "integer overflow risk review", "integer_overflow_risk_review", 1},
		{"integer underflow vulnerability", "integer underflow risk review", "integer_underflow_risk_review", 1},
	}
	for _, in := range []string{
		"signed to unsigned conversion signed unsigned cast conversion",
		"int32 uint32 size_t ssize_t checked TryFrom CWE-195",
		"signed to unsigned conversion prevention checkedCast validateRange",
		"signed to unsigned conversion risk review",
		"有符号转无符号转换漏洞 有符号转无符号转换",
		"unsigned to signed conversion vulnerability",
		"signed conversion vulnerability conversion vulnerability",
		"signed-to-unsigned conversion vulnerability signed_to_unsigned_conversion_vulnerability",
		"signed to unsigned conversion vulnerabilities",
		"signed\tto unsigned conversion vulnerability",
		"signed  to unsigned conversion vulnerability",
		"signed\nto unsigned conversion vulnerability",
		"signed to unsigned conversion\tvulnerability",
		"xsigned to unsigned conversion vulnerability",
		"signed to unsigned conversion vulnerabilityx",
		"2signed to unsigned conversion vulnerability",
		"signed to unsigned conversion vulnerability2",
		"_signed to unsigned conversion vulnerability",
		"signed to unsigned conversion vulnerability_suffix",
		"前signed to unsigned conversion vulnerability",
		"signed to unsigned conversion vulnerability后",
		`import "signed to unsigned conversion vulnerability"`,
		"https://example.invalid/signed%20to%20unsigned%20conversion%20vulnerability",
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
