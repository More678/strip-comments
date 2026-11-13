package glossary

import (
	"strings"
	"testing"
)

func TestReverseTabnabbingRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"reverse tabnabbing attack", "reverse tabnabbing risk review", "reverse_tabnabbing_risk_review", 1},
		{"Reverse tabnabbing attack", "Reverse tabnabbing risk review", "Reverse_tabnabbing_risk_review", 1},
		{"rEvErSe TaBnAbBiNg AtTaCk", "reverse tabnabbing risk review", "reverse_tabnabbing_risk_review", 1},
		{"REVERSE TABNABBING ATTACK", "REVERSE TABNABBING RISK REVIEW", "REVERSE_TABNABBING_RISK_REVIEW", 1},
		{"reverse tabnabbing attack; reverse tabnabbing attack", "reverse tabnabbing risk review; reverse tabnabbing risk review", "reverse_tabnabbing_risk_review; reverse_tabnabbing_risk_review", 2},
		{"review:\nreverse tabnabbing attack", "review:\nreverse tabnabbing risk review", "review:\nreverse_tabnabbing_risk_review", 1},
		{"(reverse tabnabbing attack).", "(reverse tabnabbing risk review).", "(reverse_tabnabbing_risk_review).", 1},
		{"open redirect attack", "untrusted destination redirect risk review", "untrusted_destination_redirect_risk_review", 1},
		{"DOM clobbering attack", "DOM named property clobbering risk review", "DOM_named_property_clobbering_risk_review", 1},
	}
	for _, in := range []string{
		"reverse tabnabbing tabnabbing tab reverse tabnabbing risk review",
		"window.opener window.open target _blank noopener noreferrer rel",
		`target="_blank" rel="noopener noreferrer"`,
		"reverse tabnabbing prevention isolateOpener preventReverseTabnabbing",
		"反向标签页劫持攻击 反向标签劫持攻击",
		"reverse-tabnabbing attack reverse-tabnabbing-attack reverse_tabnabbing_attack",
		"reverse tab nabbing attack reverse tab-nabbing attack",
		"reverse tabnabbing attacks",
		"tabnabbing attack",
		"reverse\ttabnabbing attack",
		"reverse  tabnabbing attack",
		"reverse\ntabnabbing attack",
		"reverse tabnabbing\tattack",
		"xreverse tabnabbing attack",
		"reverse tabnabbing attackx",
		"2reverse tabnabbing attack",
		"reverse tabnabbing attack2",
		"_reverse tabnabbing attack",
		"reverse tabnabbing attack_suffix",
		"前reverse tabnabbing attack",
		"reverse tabnabbing attack后",
		`import "reverse tabnabbing attack"`,
		"https://example.invalid/reverse%20tabnabbing%20attack",
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
