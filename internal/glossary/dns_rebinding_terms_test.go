package glossary

import (
	"strings"
	"testing"
)

func TestDNSRebindingRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"DNS重绑定攻击", "DNS重绑定信任边界风险核查", "DNS重绑定信任边界风险核查", 1},
		{"DNS rebinding attack", "DNS rebinding trust boundary risk review", "DNS_rebinding_trust_boundary_risk_review", 1},
		{"dns rebinding attack", "DNS rebinding trust boundary risk review", "DNS_rebinding_trust_boundary_risk_review", 1},
		{"Dns rebinding attack", "DNS rebinding trust boundary risk review", "DNS_rebinding_trust_boundary_risk_review", 1},
		{"dNs ReBiNdInG aTtAcK", "DNS rebinding trust boundary risk review", "DNS_rebinding_trust_boundary_risk_review", 1},
		{"DNS REBINDING ATTACK", "DNS REBINDING TRUST BOUNDARY RISK REVIEW", "DNS_REBINDING_TRUST_BOUNDARY_RISK_REVIEW", 1},
		{"DNS重绑定攻击; DNS rebinding attack", "DNS重绑定信任边界风险核查; DNS rebinding trust boundary risk review", "DNS重绑定信任边界风险核查; DNS_rebinding_trust_boundary_risk_review", 2},
		{"DNS缓存投毒", "名称解析缓存完整性风险核查", "名称解析缓存完整性风险核查", 1},
		{"ARP缓存投毒", "地址映射完整性风险核查", "地址映射完整性风险核查", 1},
	}
	for _, in := range []string{
		"DNS rebinding DNS重绑定 DNS rebinding TTL A AAAA CAPEC-275",
		"net.LookupHost net.LookupIP Dns.GetHostAddresses dns.lookup",
		"DNS rebinding prevention DNSRebindingProtection",
		"DNS重绑定防护 dns重绑定攻击 DNS 重绑定攻击",
		"DNS rebind attack DNS-rebinding attack dns_rebinding_attack",
		"DNS\trebinding attack",
		"xDNS rebinding attack",
		"DNS rebinding attacks",
		"2DNS rebinding attack",
		"DNS rebinding attack2",
		"_DNS rebinding attack",
		"DNS rebinding attack_suffix",
		`import "DNS rebinding attack"`,
		`import "DNS重绑定攻击"`,
		"https://example.invalid/DNS重绑定攻击",
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
