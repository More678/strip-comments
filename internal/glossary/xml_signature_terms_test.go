package glossary

import (
	"strings"
	"testing"
)

func TestXMLSignatureRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"XML signature wrapping attack", "XML signature wrapping risk review", "XML_signature_wrapping_risk_review", 1},
		{"xml signature wrapping attack", "XML signature wrapping risk review", "XML_signature_wrapping_risk_review", 1},
		{"Xml signature wrapping attack", "XML signature wrapping risk review", "XML_signature_wrapping_risk_review", 1},
		{"xMl SiGnAtUrE wRaPpInG aTtAcK", "XML signature wrapping risk review", "XML_signature_wrapping_risk_review", 1},
		{"XML SIGNATURE WRAPPING ATTACK", "XML SIGNATURE WRAPPING RISK REVIEW", "XML_SIGNATURE_WRAPPING_RISK_REVIEW", 1},
		{"XML signature wrapping attack; XML signature wrapping attack", "XML signature wrapping risk review; XML signature wrapping risk review", "XML_signature_wrapping_risk_review; XML_signature_wrapping_risk_review", 2},
		{"JWT算法混淆攻击", "JWT验签算法混淆风险核查", "JWT验签算法混淆风险核查", 1},
		{"XML外部实体攻击", "XML外部实体解析风险核查", "XML外部实体解析风险核查", 1},
	}
	for _, in := range []string{
		"XML signature wrapping XSW SAML XML signature wrapping 签名",
		"Signature Reference Assertion Response ID SignedXml.CheckSignature",
		"urn:oasis:names:tc:SAML:2.0:assertion",
		"XML signature wrapping prevention XMLSignatureWrappingProtection",
		"XML签名包装攻击 signature wrapping attack",
		"XML-signature-wrapping-attack XML_signature_wrapping_attack",
		"XML\tsignature wrapping attack",
		"xXML signature wrapping attack",
		"XML signature wrapping attacks",
		"2XML signature wrapping attack",
		"XML signature wrapping attack2",
		"_XML signature wrapping attack",
		"XML signature wrapping attack_suffix",
		`import "XML signature wrapping attack"`,
		"https://example.invalid/XML%20signature%20wrapping%20attack",
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
