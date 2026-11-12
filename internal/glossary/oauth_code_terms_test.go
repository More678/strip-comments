package glossary

import (
	"strings"
	"testing"
)

func TestOAuthCodeRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"OAuth2.0授权码注入攻击", "OAuth2.0授权码注入风险核查", "OAuth2.0授权码注入风险核查", 1},
		{"OAuth 2.0 authorization code injection attack", "OAuth 2.0 authorization code injection risk review", "OAuth_2.0_authorization_code_injection_risk_review", 1},
		{"oauth 2.0 authorization code injection attack", "OAuth 2.0 authorization code injection risk review", "OAuth_2.0_authorization_code_injection_risk_review", 1},
		{"Oauth 2.0 authorization code injection attack", "OAuth 2.0 authorization code injection risk review", "OAuth_2.0_authorization_code_injection_risk_review", 1},
		{"oAuTh 2.0 AuThOrIzAtIoN cOdE iNjEcTiOn AtTaCk", "OAuth 2.0 authorization code injection risk review", "OAuth_2.0_authorization_code_injection_risk_review", 1},
		{"OAUTH 2.0 AUTHORIZATION CODE INJECTION ATTACK", "OAUTH 2.0 AUTHORIZATION CODE INJECTION RISK REVIEW", "OAUTH_2.0_AUTHORIZATION_CODE_INJECTION_RISK_REVIEW", 1},
		{"OAuth2.0授权码注入攻击; OAuth 2.0 authorization code injection attack", "OAuth2.0授权码注入风险核查; OAuth 2.0 authorization code injection risk review", "OAuth2.0授权码注入风险核查; OAuth_2.0_authorization_code_injection_risk_review", 2},
		{"JWT算法混淆攻击", "JWT验签算法混淆风险核查", "JWT验签算法混淆风险核查", 1},
		{"JWT algorithm confusion attack", "JWT verification algorithm confusion risk review", "JWT_verification_algorithm_confusion_risk_review", 1},
	}
	for _, in := range []string{
		"OAuth 2.0 authorization code injection OAuth2.0授权码注入",
		"OAuth authorization code 授权码 client_id redirect_uri code_challenge code_verifier",
		"PKCE state nonce RFC 9700 exchangeCode validateAuthorizationResponse",
		"OAuth2.0授权码注入防护 OAuthCodeInjectionProtection",
		"oauth2.0授权码注入攻击 OAuth 2.0授权码注入攻击",
		"OAuth authorization code injection attack OAuth2 authorization code injection attack",
		"OAuth-2.0-authorization-code-injection-attack OAuth_2.0_authorization_code_injection_attack",
		"OAuth\t2.0 authorization code injection attack",
		"xOAuth 2.0 authorization code injection attack",
		"OAuth 2.0 authorization code injection attacks",
		"2OAuth 2.0 authorization code injection attack",
		"OAuth 2.0 authorization code injection attack2",
		"_OAuth 2.0 authorization code injection attack",
		"OAuth 2.0 authorization code injection attack_suffix",
		`import "OAuth 2.0 authorization code injection attack"`,
		`import "OAuth2.0授权码注入攻击"`,
		"https://example.invalid/OAuth2.0授权码注入攻击",
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
