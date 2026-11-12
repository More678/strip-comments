package glossary

import (
	"strings"
	"testing"
)

func TestOAuthMixupRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"OAuth mix-up attack", "OAuth authorization server mix-up risk review", "OAuth_authorization_server_mix_up_risk_review", 1},
		{"oauth mix-up attack", "OAuth authorization server mix-up risk review", "OAuth_authorization_server_mix_up_risk_review", 1},
		{"Oauth mix-up attack", "OAuth authorization server mix-up risk review", "OAuth_authorization_server_mix_up_risk_review", 1},
		{"oAuTh MiX-uP aTtAcK", "OAuth authorization server mix-up risk review", "OAuth_authorization_server_mix_up_risk_review", 1},
		{"OAUTH MIX-UP ATTACK", "OAUTH AUTHORIZATION SERVER MIX-UP RISK REVIEW", "OAUTH_AUTHORIZATION_SERVER_MIX_UP_RISK_REVIEW", 1},
		{"OAuth mix-up attack; OAuth mix-up attack", "OAuth authorization server mix-up risk review; OAuth authorization server mix-up risk review", "OAuth_authorization_server_mix_up_risk_review; OAuth_authorization_server_mix_up_risk_review", 2},
		{"OAuth2.0授权码注入攻击", "OAuth2.0授权码注入风险核查", "OAuth2.0授权码注入风险核查", 1},
		{"JWT算法混淆攻击", "JWT验签算法混淆风险核查", "JWT验签算法混淆风险核查", 1},
	}
	for _, in := range []string{
		"OAuth mix-up mix-up OAuth issuer authorization server",
		"iss authorization_endpoint token_endpoint redirect_uri RFC 9700",
		"OAuth mix-up prevention OAuthMixupProtection validateIssuer",
		"OAuth授权服务器混淆攻击",
		"OAuth mix up attack OAuth mixup attack OAuth_mix_up_attack",
		"OAuth\tmix-up attack",
		"xOAuth mix-up attack",
		"OAuth mix-up attacks",
		"2OAuth mix-up attack",
		"OAuth mix-up attack2",
		"_OAuth mix-up attack",
		"OAuth mix-up attack_suffix",
		`import "OAuth mix-up attack"`,
		"https://example.invalid/OAuth%20mix-up%20attack",
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
