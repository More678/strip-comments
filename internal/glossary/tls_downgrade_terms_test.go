package glossary

import (
	"strings"
	"testing"
)

func TestTLSDowngradeRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"TLS协议降级攻击", "TLS协议降级风险核查", "TLS协议降级风险核查", 1},
		{"TLS protocol downgrade attack", "TLS protocol downgrade risk review", "TLS_protocol_downgrade_risk_review", 1},
		{"tls protocol downgrade attack", "TLS protocol downgrade risk review", "TLS_protocol_downgrade_risk_review", 1},
		{"Tls protocol downgrade attack", "TLS protocol downgrade risk review", "TLS_protocol_downgrade_risk_review", 1},
		{"tLs PrOtOcOl DoWnGrAdE aTtAcK", "TLS protocol downgrade risk review", "TLS_protocol_downgrade_risk_review", 1},
		{"TLS PROTOCOL DOWNGRADE ATTACK", "TLS PROTOCOL DOWNGRADE RISK REVIEW", "TLS_PROTOCOL_DOWNGRADE_RISK_REVIEW", 1},
		{"TLS协议降级攻击; TLS protocol downgrade attack", "TLS协议降级风险核查; TLS protocol downgrade risk review", "TLS协议降级风险核查; TLS_protocol_downgrade_risk_review", 2},
		{"JWT算法混淆攻击", "JWT验签算法混淆风险核查", "JWT验签算法混淆风险核查", 1},
		{"JWT algorithm confusion attack", "JWT verification algorithm confusion risk review", "JWT_verification_algorithm_confusion_risk_review", 1},
	}
	for _, in := range []string{
		"TLS protocol downgrade TLS协议降级 TLS SSL downgrade 降级",
		"ClientHello ServerHello supported_versions MinVersion MaxVersion RFC 8446 CWE-757",
		"tls.Config tls.Client SslStream.AuthenticateAsClient",
		"TLS protocol downgrade prevention TLSDowngradeProtection",
		"TLS协议降级防护 tls协议降级攻击 TLS 协议降级攻击",
		"TLS downgrade attack protocol downgrade attack",
		"TLS-protocol-downgrade-attack TLS_protocol_downgrade_attack",
		"TLS\tprotocol downgrade attack",
		"xTLS protocol downgrade attack",
		"TLS protocol downgrade attacks",
		"2TLS protocol downgrade attack",
		"TLS protocol downgrade attack2",
		"_TLS protocol downgrade attack",
		"TLS protocol downgrade attack_suffix",
		`import "TLS protocol downgrade attack"`,
		`import "TLS协议降级攻击"`,
		"https://example.invalid/TLS协议降级攻击",
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
