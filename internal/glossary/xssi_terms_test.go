package glossary

import (
	"strings"
	"testing"
)

func TestXSSIRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"跨站脚本包含攻击", "跨站脚本包含数据泄露风险核查", "跨站脚本包含数据泄露风险核查", 1},
		{"cross site script inclusion attack", "cross site script inclusion data exposure risk review", "cross_site_script_inclusion_data_exposure_risk_review", 1},
		{"Cross site script inclusion attack", "Cross site script inclusion data exposure risk review", "Cross_site_script_inclusion_data_exposure_risk_review", 1},
		{"cRoSs SiTe ScRiPt InClUsIoN AtTaCk", "cross site script inclusion data exposure risk review", "cross_site_script_inclusion_data_exposure_risk_review", 1},
		{"CROSS SITE SCRIPT INCLUSION ATTACK", "CROSS SITE SCRIPT INCLUSION DATA EXPOSURE RISK REVIEW", "CROSS_SITE_SCRIPT_INCLUSION_DATA_EXPOSURE_RISK_REVIEW", 1},
		{"跨站脚本包含攻击; cross site script inclusion attack", "跨站脚本包含数据泄露风险核查; cross site script inclusion data exposure risk review", "跨站脚本包含数据泄露风险核查; cross_site_script_inclusion_data_exposure_risk_review", 2},
		{"review:\ncross site script inclusion attack", "review:\ncross site script inclusion data exposure risk review", "review:\ncross_site_script_inclusion_data_exposure_risk_review", 1},
		{"(cross site script inclusion attack).", "(cross site script inclusion data exposure risk review).", "(cross_site_script_inclusion_data_exposure_risk_review).", 1},
		{"跨站WebSocket劫持攻击", "跨站WebSocket会话劫持风险核查", "跨站WebSocket会话劫持风险核查", 1},
		{"cross-site WebSocket hijacking attack", "cross-site WebSocket session hijacking risk review", "cross_site_WebSocket_session_hijacking_risk_review", 1},
		{"reverse tabnabbing attack", "reverse tabnabbing risk review", "reverse_tabnabbing_risk_review", 1},
	}
	for _, in := range []string{
		"XSSI cross site script inclusion 跨站脚本包含 script include 跨站 包含",
		"JSONP JSON CSV Content-Type SameSite WSTG-CLNT-13",
		"XSSI prevention preventXSSI loadScript fetch importScripts createElement",
		"cross site script inclusion data exposure risk review 跨站脚本包含数据泄露风险核查",
		"跨站脚本包含防护 跨站脚本 包含攻击 跨站脚本包含 攻击",
		"cross-site script inclusion attack cross-site-script-inclusion-attack cross_site_script_inclusion_attack",
		"XSSI attack script inclusion attack",
		"cross site script inclusion attacks",
		"cross\tsite script inclusion attack",
		"cross  site script inclusion attack",
		"cross\nsite script inclusion attack",
		"cross site script inclusion\tattack",
		"xcross site script inclusion attack",
		"cross site script inclusion attackx",
		"2cross site script inclusion attack",
		"cross site script inclusion attack2",
		"_cross site script inclusion attack",
		"cross site script inclusion attack_suffix",
		"前cross site script inclusion attack",
		"cross site script inclusion attack后",
		`import "cross site script inclusion attack"`,
		`import "跨站脚本包含攻击"`,
		"https://example.invalid/cross%20site%20script%20inclusion%20attack",
		"https://example.invalid/跨站脚本包含攻击",
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
