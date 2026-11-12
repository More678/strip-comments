package glossary

import (
	"strings"
	"testing"
)

func TestTOCTOURiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"TOCTOU race condition vulnerability", "TOCTOU check-use race risk review", "TOCTOU_check_use_race_risk_review", 1},
		{"toctou race condition vulnerability", "TOCTOU check-use race risk review", "TOCTOU_check_use_race_risk_review", 1},
		{"Toctou race condition vulnerability", "TOCTOU check-use race risk review", "TOCTOU_check_use_race_risk_review", 1},
		{"tOcToU rAcE cOnDiTiOn VuLnErAbIlItY", "TOCTOU check-use race risk review", "TOCTOU_check_use_race_risk_review", 1},
		{"TOCTOU RACE CONDITION VULNERABILITY", "TOCTOU CHECK-USE RACE RISK REVIEW", "TOCTOU_CHECK_USE_RACE_RISK_REVIEW", 1},
		{"TOCTOU race condition vulnerability; TOCTOU race condition vulnerability", "TOCTOU check-use race risk review; TOCTOU check-use race risk review", "TOCTOU_check_use_race_risk_review; TOCTOU_check_use_race_risk_review", 2},
		{"内存释放后使用漏洞", "内存释放后使用风险核查", "内存释放后使用风险核查", 1},
		{"double free vulnerability", "double free risk review", "double_free_risk_review", 1},
	}
	for _, in := range []string{
		"TOCTOU TOCTOU race condition race condition check use 竞态 竞争条件 CWE-367",
		"os.Stat os.Open sync.Mutex Lock Unlock atomic.LoadInt32",
		"TOCTOU race condition prevention TOCTOUProtection",
		"TOCTOU竞争条件漏洞 time-of-check time-of-use vulnerability",
		"TOCTOU-race-condition-vulnerability TOCTOU_race_condition_vulnerability",
		"TOCTOU\trace condition vulnerability",
		"xTOCTOU race condition vulnerability",
		"TOCTOU race condition vulnerabilityx",
		"2TOCTOU race condition vulnerability",
		"TOCTOU race condition vulnerability2",
		"_TOCTOU race condition vulnerability",
		"TOCTOU race condition vulnerability_suffix",
		`import "TOCTOU race condition vulnerability"`,
		"https://example.invalid/TOCTOU%20race%20condition%20vulnerability",
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
