package glossary

import (
	"strings"
	"testing"
)

func TestOAuthRefreshRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"OAuth刷新令牌重放攻击", "OAuth刷新令牌重放风险核查", "OAuth刷新令牌重放风险核查", 1},
		{"OAuth refresh token replay attack", "OAuth refresh token replay risk review", "OAuth_refresh_token_replay_risk_review", 1},
		{"oauth refresh token replay attack", "OAuth refresh token replay risk review", "OAuth_refresh_token_replay_risk_review", 1},
		{"Oauth refresh token replay attack", "OAuth refresh token replay risk review", "OAuth_refresh_token_replay_risk_review", 1},
		{"oAuTh ReFrEsH tOkEn RePlAy AtTaCk", "OAuth refresh token replay risk review", "OAuth_refresh_token_replay_risk_review", 1},
		{"OAUTH REFRESH TOKEN REPLAY ATTACK", "OAUTH REFRESH TOKEN REPLAY RISK REVIEW", "OAUTH_REFRESH_TOKEN_REPLAY_RISK_REVIEW", 1},
		{"OAuth刷新令牌重放攻击; OAuth refresh token replay attack", "OAuth刷新令牌重放风险核查; OAuth refresh token replay risk review", "OAuth刷新令牌重放风险核查; OAuth_refresh_token_replay_risk_review", 2},
		{"OAuth2.0授权码注入攻击", "OAuth2.0授权码注入风险核查", "OAuth2.0授权码注入风险核查", 1},
		{"OAuth mix-up attack", "OAuth authorization server mix-up risk review", "OAuth_authorization_server_mix_up_risk_review", 1},
	}
	for _, in := range []string{
		"OAuth refresh token replay OAuth刷新令牌重放 refresh token 刷新令牌 replay 轮换",
		"refresh_token grant_type client_id DPoP mTLS RFC 9700",
		"rotateRefreshToken validateRefreshToken OAuthRefreshTokenProtection",
		"OAuth刷新令牌重放防护 oauth刷新令牌重放攻击 OAuth 刷新令牌重放攻击",
		"refresh token replay attack 刷新令牌重放攻击",
		"OAuth-refresh-token-replay-attack OAuth_refresh_token_replay_attack",
		"OAuth\trefresh token replay attack",
		"xOAuth refresh token replay attack",
		"OAuth refresh token replay attacks",
		"2OAuth refresh token replay attack",
		"OAuth refresh token replay attack2",
		"_OAuth refresh token replay attack",
		"OAuth refresh token replay attack_suffix",
		`import "OAuth refresh token replay attack"`,
		`import "OAuth刷新令牌重放攻击"`,
		"https://example.invalid/OAuth刷新令牌重放攻击",
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
