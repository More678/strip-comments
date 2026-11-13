package glossary

import (
	"strings"
	"testing"
)

func TestSessionPuzzlingRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"session puzzling attack", "session variable overloading risk review", "session_variable_overloading_risk_review", 1},
		{"Session puzzling attack", "Session variable overloading risk review", "Session_variable_overloading_risk_review", 1},
		{"sEsSiOn PuZzLiNg AtTaCk", "session variable overloading risk review", "session_variable_overloading_risk_review", 1},
		{"SESSION PUZZLING ATTACK", "SESSION VARIABLE OVERLOADING RISK REVIEW", "SESSION_VARIABLE_OVERLOADING_RISK_REVIEW", 1},
		{"session puzzling attack; session puzzling attack", "session variable overloading risk review; session variable overloading risk review", "session_variable_overloading_risk_review; session_variable_overloading_risk_review", 2},
		{"review:\nsession puzzling attack", "review:\nsession variable overloading risk review", "review:\nsession_variable_overloading_risk_review", 1},
		{"(session puzzling attack).", "(session variable overloading risk review).", "(session_variable_overloading_risk_review).", 1},
		{"会话固定攻击", "会话标识轮换风险核查", "会话标识轮换风险核查", 1},
		{"session fixation attack", "session identifier rotation risk review", "session_identifier_rotation_risk_review", 1},
		{"reverse tabnabbing attack", "reverse tabnabbing risk review", "reverse_tabnabbing_risk_review", 1},
	}
	for _, in := range []string{
		"session puzzling session variable overloading session 会话 session_id",
		"HttpSession SessionState WSTG-SESS-08 setAttribute getAttribute",
		"session puzzling prevention validateSessionState SessionFixationProtection",
		"session variable overloading risk review",
		"会话混淆攻击 会话变量重载攻击",
		"session variable overloading attack",
		"session-puzzling-attack session_puzzling_attack",
		"session puzzling attacks",
		"session\tpuzzling attack",
		"session  puzzling attack",
		"session\npuzzling attack",
		"session puzzling\tattack",
		"xsession puzzling attack",
		"session puzzling attackx",
		"2session puzzling attack",
		"session puzzling attack2",
		"_session puzzling attack",
		"session puzzling attack_suffix",
		"前session puzzling attack",
		"session puzzling attack后",
		`import "session puzzling attack"`,
		"https://example.invalid/session%20puzzling%20attack",
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
