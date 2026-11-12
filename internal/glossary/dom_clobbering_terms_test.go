package glossary

import (
	"strings"
	"testing"
)

func TestDOMClobberingRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"DOM clobbering attack", "DOM named property clobbering risk review", "DOM_named_property_clobbering_risk_review", 1},
		{"dom clobbering attack", "DOM named property clobbering risk review", "DOM_named_property_clobbering_risk_review", 1},
		{"Dom clobbering attack", "DOM named property clobbering risk review", "DOM_named_property_clobbering_risk_review", 1},
		{"dOm ClObBeRiNg AtTaCk", "DOM named property clobbering risk review", "DOM_named_property_clobbering_risk_review", 1},
		{"DOM CLOBBERING ATTACK", "DOM NAMED PROPERTY CLOBBERING RISK REVIEW", "DOM_NAMED_PROPERTY_CLOBBERING_RISK_REVIEW", 1},
		{"DOM clobbering attack; DOM clobbering attack", "DOM named property clobbering risk review; DOM named property clobbering risk review", "DOM_named_property_clobbering_risk_review; DOM_named_property_clobbering_risk_review", 2},
		{"原型污染攻击", "对象原型篡改风险核查", "对象原型篡改风险核查", 1},
		{"prototype pollution attack", "object prototype tampering risk review", "object_prototype_tampering_risk_review", 1},
	}
	for _, in := range []string{
		"DOM clobbering DOM clobbering prevention DOM clobbering defense",
		"DOM clobbering window document id name",
		"document.getElementById window.location SANITIZE_NAMED_PROPS DOMPurify",
		"DOMClobberingProtection dom_clobbering_attack",
		"DOM命名属性覆盖攻击 DOM属性覆盖攻击",
		"DOM-clobbering-attack",
		"DOM\tclobbering attack",
		"xDOM clobbering attack",
		"DOM clobbering attacks",
		"2DOM clobbering attack",
		"DOM clobbering attack2",
		"_DOM clobbering attack",
		"DOM clobbering attack_suffix",
		`import "DOM clobbering attack"`,
		"https://example.invalid/DOM-clobbering-attack",
		"https://example.invalid/DOM%20clobbering%20attack",
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
