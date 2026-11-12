package glossary

import (
	"strings"
	"testing"
)

func TestJWTAlgorithmRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"JWT算法混淆攻击", "JWT验签算法混淆风险核查", "JWT验签算法混淆风险核查", 1},
		{"JWT algorithm confusion attack", "JWT verification algorithm confusion risk review", "JWT_verification_algorithm_confusion_risk_review", 1},
		{"jwt algorithm confusion attack", "JWT verification algorithm confusion risk review", "JWT_verification_algorithm_confusion_risk_review", 1},
		{"Jwt algorithm confusion attack", "JWT verification algorithm confusion risk review", "JWT_verification_algorithm_confusion_risk_review", 1},
		{"jWt AlGoRiThM cOnFuSiOn AtTaCk", "JWT verification algorithm confusion risk review", "JWT_verification_algorithm_confusion_risk_review", 1},
		{"JWT ALGORITHM CONFUSION ATTACK", "JWT VERIFICATION ALGORITHM CONFUSION RISK REVIEW", "JWT_VERIFICATION_ALGORITHM_CONFUSION_RISK_REVIEW", 1},
		{"JWT算法混淆攻击; JWT algorithm confusion attack", "JWT验签算法混淆风险核查; JWT verification algorithm confusion risk review", "JWT验签算法混淆风险核查; JWT_verification_algorithm_confusion_risk_review", 2},
		{"依赖混淆攻击", "依赖来源解析风险核查", "依赖来源解析风险核查", 1},
		{"dependency confusion attack", "dependency source resolution risk review", "dependency_source_resolution_risk_review", 1},
	}
	for _, in := range []string{
		"JWT algorithm confusion JWT算法混淆 algorithm confusion JWT",
		"alg RS256 HS256 none kid iss aud exp RFC 8725",
		"jwt.Parse jwt.ParseWithClaims jwt.Verify ValidateToken",
		"JWT算法混淆防护 JWTAlgorithmConfusionProtection",
		"jwt算法混淆攻击 JWT 算法混淆攻击",
		"JWT key confusion attack algorithm confusion attack",
		"JWT-algorithm-confusion-attack JWT_algorithm_confusion_attack",
		"JWT\talgorithm confusion attack",
		"xJWT algorithm confusion attack",
		"JWT algorithm confusion attacks",
		"2JWT algorithm confusion attack",
		"JWT algorithm confusion attack2",
		"_JWT algorithm confusion attack",
		"JWT algorithm confusion attack_suffix",
		`import "JWT algorithm confusion attack"`,
		`import "JWT算法混淆攻击"`,
		"https://example.invalid/JWT算法混淆攻击",
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
