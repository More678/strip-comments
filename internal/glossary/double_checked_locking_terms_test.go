package glossary

import (
	"strings"
	"testing"
)

func TestDoubleCheckedLockingRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"double-checked locking vulnerability", "double-checked locking initialization visibility risk review", "double_checked_locking_initialization_visibility_risk_review", 1},
		{"Double-checked locking vulnerability", "Double-checked locking initialization visibility risk review", "Double_checked_locking_initialization_visibility_risk_review", 1},
		{"dOuBlE-cHeCkEd LoCkInG VuLnErAbIlItY", "double-checked locking initialization visibility risk review", "double_checked_locking_initialization_visibility_risk_review", 1},
		{"DOUBLE-CHECKED LOCKING VULNERABILITY", "DOUBLE-CHECKED LOCKING INITIALIZATION VISIBILITY RISK REVIEW", "DOUBLE_CHECKED_LOCKING_INITIALIZATION_VISIBILITY_RISK_REVIEW", 1},
		{"double-checked locking vulnerability; double-checked locking vulnerability", "double-checked locking initialization visibility risk review; double-checked locking initialization visibility risk review", "double_checked_locking_initialization_visibility_risk_review; double_checked_locking_initialization_visibility_risk_review", 2},
		{"review:\ndouble-checked locking vulnerability", "review:\ndouble-checked locking initialization visibility risk review", "review:\ndouble_checked_locking_initialization_visibility_risk_review", 1},
		{"(double-checked locking vulnerability).", "(double-checked locking initialization visibility risk review).", "(double_checked_locking_initialization_visibility_risk_review).", 1},
		{"TOCTOU race condition vulnerability", "TOCTOU check-use race risk review", "TOCTOU_check_use_race_risk_review", 1},
		{"未初始化变量使用漏洞", "未初始化变量使用风险核查", "未初始化变量使用风险核查", 1},
		{"uninitialized variable use vulnerability", "uninitialized variable use risk review", "uninitialized_variable_use_risk_review", 1},
	}
	for _, in := range []string{
		"double-checked locking DCL volatile synchronized sync.Once call_once",
		"Lazy atomic happens-before CWE-609 lock initialization",
		"double-checked locking prevention safePublication validateInitialization",
		"double-checked locking initialization visibility risk review",
		"双重检查锁定漏洞 双重检查锁漏洞",
		"DCL vulnerability double checked locking vulnerability",
		"double_checked_locking_vulnerability double-checked-locking-vulnerability",
		"double-checked locking vulnerabilities",
		"double-checked\tlocking vulnerability",
		"double-checked  locking vulnerability",
		"double-checked\nlocking vulnerability",
		"double-checked locking\tvulnerability",
		"xdouble-checked locking vulnerability",
		"double-checked locking vulnerabilityx",
		"2double-checked locking vulnerability",
		"double-checked locking vulnerability2",
		"_double-checked locking vulnerability",
		"double-checked locking vulnerability_suffix",
		"前double-checked locking vulnerability",
		"double-checked locking vulnerability后",
		`import "double-checked locking vulnerability"`,
		"https://example.invalid/double-checked%20locking%20vulnerability",
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
