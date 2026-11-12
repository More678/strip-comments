package glossary

import (
	"strings"
	"testing"
)

func TestWebSocketRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"跨站WebSocket劫持攻击", "跨站WebSocket会话劫持风险核查", "跨站WebSocket会话劫持风险核查", 1},
		{"cross-site WebSocket hijacking attack", "cross-site WebSocket session hijacking risk review", "cross_site_WebSocket_session_hijacking_risk_review", 1},
		{"cross-site websocket hijacking attack", "cross-site WebSocket session hijacking risk review", "cross_site_WebSocket_session_hijacking_risk_review", 1},
		{"Cross-site websocket hijacking attack", "Cross-site WebSocket session hijacking risk review", "Cross_site_WebSocket_session_hijacking_risk_review", 1},
		{"cRoSs-SiTe WeBsOcKeT hIjAcKiNg AtTaCk", "cross-site WebSocket session hijacking risk review", "cross_site_WebSocket_session_hijacking_risk_review", 1},
		{"CROSS-SITE WEBSOCKET HIJACKING ATTACK", "CROSS-SITE WEBSOCKET SESSION HIJACKING RISK REVIEW", "CROSS_SITE_WEBSOCKET_SESSION_HIJACKING_RISK_REVIEW", 1},
		{"跨站WebSocket劫持攻击; cross-site WebSocket hijacking attack", "跨站WebSocket会话劫持风险核查; cross-site WebSocket session hijacking risk review", "跨站WebSocket会话劫持风险核查; cross_site_WebSocket_session_hijacking_risk_review", 2},
		{"浏览器会话劫持", "浏览器会话完整性核查", "浏览器会话完整性核查", 1},
		{"browser session hijacking", "browser session integrity review", "browser_session_integrity_review", 1},
	}
	for _, in := range []string{
		"cross-site WebSocket hijacking 跨站WebSocket劫持 CSWSH",
		"WebSocket Origin Cookie SameSite Upgrade Sec-WebSocket-Key",
		"WebSocket.send WebSocket.close CheckOrigin handleUpgrade verifyClient",
		"cross-site WebSocket hijacking prevention CrossSiteWebSocketHijackingProtection",
		"跨站WebSocket劫持防护 跨站websocket劫持攻击 跨站 WebSocket劫持攻击",
		"cross site WebSocket hijacking attack cross-site web socket hijacking attack",
		"cross-site-WebSocket-hijacking-attack cross_site_WebSocket_hijacking_attack",
		"cross-site\tWebSocket hijacking attack",
		"xcross-site WebSocket hijacking attack",
		"cross-site WebSocket hijacking attacks",
		"2cross-site WebSocket hijacking attack",
		"cross-site WebSocket hijacking attack2",
		"_cross-site WebSocket hijacking attack",
		"cross-site WebSocket hijacking attack_suffix",
		`import "cross-site WebSocket hijacking attack"`,
		`import "跨站WebSocket劫持攻击"`,
		"https://example.invalid/跨站WebSocket劫持攻击",
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
