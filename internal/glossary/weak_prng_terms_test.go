package glossary

import (
	"strings"
	"testing"
)

func TestWeakPRNGRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"cryptographically weak PRNG vulnerability", "cryptographically weak PRNG risk review", "cryptographically_weak_PRNG_risk_review", 1},
		{"cryptographically weak prng vulnerability", "cryptographically weak PRNG risk review", "cryptographically_weak_PRNG_risk_review", 1},
		{"Cryptographically weak prng vulnerability", "Cryptographically weak PRNG risk review", "Cryptographically_weak_PRNG_risk_review", 1},
		{"cRyPtOgRaPhIcAlLy WeAk PrNg VuLnErAbIlItY", "cryptographically weak PRNG risk review", "cryptographically_weak_PRNG_risk_review", 1},
		{"CRYPTOGRAPHICALLY WEAK PRNG VULNERABILITY", "CRYPTOGRAPHICALLY WEAK PRNG RISK REVIEW", "CRYPTOGRAPHICALLY_WEAK_PRNG_RISK_REVIEW", 1},
		{"cryptographically weak PRNG vulnerability; cryptographically weak PRNG vulnerability", "cryptographically weak PRNG risk review; cryptographically weak PRNG risk review", "cryptographically_weak_PRNG_risk_review; cryptographically_weak_PRNG_risk_review", 2},
		{"review:\ncryptographically weak PRNG vulnerability", "review:\ncryptographically weak PRNG risk review", "review:\ncryptographically_weak_PRNG_risk_review", 1},
		{"(cryptographically weak PRNG vulnerability).", "(cryptographically weak PRNG risk review).", "(cryptographically_weak_PRNG_risk_review).", 1},
		{"padding oracle attack", "padding oracle risk review", "padding_oracle_risk_review", 1},
		{"TLS协议降级攻击", "TLS协议降级风险核查", "TLS协议降级风险核查", 1},
		{"TLS protocol downgrade attack", "TLS protocol downgrade risk review", "TLS_protocol_downgrade_risk_review", 1},
	}
	for _, in := range []string{
		"PRNG CSPRNG cryptographically weak PRNG random seed entropy",
		"math/rand crypto/rand SecureRandom RandomNumberGenerator CWE-338",
		"cryptographically weak PRNG prevention reseedGenerator validateEntropy",
		"cryptographically weak PRNG risk review",
		"密码学弱伪随机生成器漏洞 弱随机数漏洞",
		"weak PRNG vulnerability weak random vulnerability",
		"cryptographically weak pseudo-random number generator vulnerability",
		"cryptographically-weak-PRNG-vulnerability cryptographically_weak_PRNG_vulnerability",
		"cryptographically weak PRNG vulnerabilities",
		"cryptographically\tweak PRNG vulnerability",
		"cryptographically  weak PRNG vulnerability",
		"cryptographically\nweak PRNG vulnerability",
		"cryptographically weak PRNG\tvulnerability",
		"xcryptographically weak PRNG vulnerability",
		"cryptographically weak PRNG vulnerabilityx",
		"2cryptographically weak PRNG vulnerability",
		"cryptographically weak PRNG vulnerability2",
		"_cryptographically weak PRNG vulnerability",
		"cryptographically weak PRNG vulnerability_suffix",
		"前cryptographically weak PRNG vulnerability",
		"cryptographically weak PRNG vulnerability后",
		`import "cryptographically weak PRNG vulnerability"`,
		"https://example.invalid/cryptographically%20weak%20PRNG%20vulnerability",
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
