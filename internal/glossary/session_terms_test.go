package glossary

import "testing"

func TestSessionAndInteractionRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct{ in, want string }{
		{"会话固定攻击", "会话标识轮换风险核查"},
		{"点击劫持攻击", "界面交互劫持风险核查"},
		{"session fixation attack", "session identifier rotation risk review"},
		{"clickjacking attack", "interface interaction hijack risk review"},
		{"Session fixation attack", "Session identifier rotation risk review"},
		{"CLICKJACKING ATTACK", "INTERFACE INTERACTION HIJACK RISK REVIEW"},
		{"会话固定防护", "会话固定防护"},
		{"点击劫持防护", "点击劫持防护"},
		{"session fixation", "session fixation"},
		{"SessionFixationProtection", "SessionFixationProtection"},
		{"clickjacking defense", "clickjacking defense"},
		{"my_session_fixation_attack", "my_session_fixation_attack"},
		{"session fixation attack2", "session fixation attack2"},
		{"preclickjacking attack", "preclickjacking attack"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			for _, comment := range []bool{false, true} {
				opt := Options{}
				want := tc.want
				if comment {
					opt.Comments = []Region{{Start: 0, End: len(tc.in), Comment: true}}
				} else if tc.want != tc.in {
					want = toIdent(want)
				}
				out, _ := tb.Apply([]byte(tc.in), opt)
				if string(out) != want {
					t.Fatalf("comment=%v: got %q, want %q", comment, out, want)
				}
				if comment {
					opt.Comments[0].End = len(out)
				}
				again, hits := tb.Apply(out, opt)
				if string(again) != string(out) || len(hits) != 0 {
					t.Fatalf("second pass changed output: %q (%+v)", again, hits)
				}
			}
		})
	}
}
