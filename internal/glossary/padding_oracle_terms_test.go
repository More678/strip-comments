package glossary

import (
	"strings"
	"testing"
)

func TestPaddingOracleRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"padding oracle attack", "padding oracle risk review", "padding_oracle_risk_review", 1},
		{"Padding oracle attack", "Padding oracle risk review", "Padding_oracle_risk_review", 1},
		{"pAdDiNg OrAcLe AtTaCk", "padding oracle risk review", "padding_oracle_risk_review", 1},
		{"PADDING ORACLE ATTACK", "PADDING ORACLE RISK REVIEW", "PADDING_ORACLE_RISK_REVIEW", 1},
		{"padding oracle attack; padding oracle attack", "padding oracle risk review; padding oracle risk review", "padding_oracle_risk_review; padding_oracle_risk_review", 2},
		{"TLS协议降级攻击", "TLS协议降级风险核查", "TLS协议降级风险核查", 1},
		{"TLS protocol downgrade attack", "TLS protocol downgrade risk review", "TLS_protocol_downgrade_risk_review", 1},
		{"JWT算法混淆攻击", "JWT验签算法混淆风险核查", "JWT验签算法混淆风险核查", 1},
		{"JWT algorithm confusion attack", "JWT verification algorithm confusion risk review", "JWT_verification_algorithm_confusion_risk_review", 1},
	}
	for _, in := range []string{
		"padding oracle padding oracle risk review padding oracle CBC PKCS7 WSTG-CRYP-02",
		"padding oracle prevention BadPaddingException InvalidPaddingException Decrypt validatePadding",
		"PKCS7Padding AES/CBC/PKCS5Padding authenticated encryption AEAD Oracle database",
		"填充预言攻击 填充谕示攻击",
		"padding-oracle attack padding-oracle-attack padding_oracle_attack",
		"padding oracle attacks",
		"padding\toracle attack",
		"padding  oracle attack",
		"padding\noracle attack",
		"xpadding oracle attack",
		"padding oracle attackx",
		"2padding oracle attack",
		"padding oracle attack2",
		"_padding oracle attack",
		"padding oracle attack_suffix",
		"前padding oracle attack",
		"padding oracle attack后",
		`import "padding oracle attack"`,
		"https://example.invalid/padding%20oracle%20attack",
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
