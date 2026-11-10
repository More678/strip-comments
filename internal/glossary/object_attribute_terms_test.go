package glossary

import (
	"strings"
	"testing"
)

func TestObjectAttributeRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"原型污染攻击", "对象原型篡改风险核查", "对象原型篡改风险核查", 1},
		{"批量赋值漏洞", "非预期字段写入风险核查", "非预期字段写入风险核查", 1},
		{"prototype pollution attack", "object prototype tampering risk review", "object_prototype_tampering_risk_review", 1},
		{"mass assignment vulnerability", "unintended field modification risk review", "unintended_field_modification_risk_review", 1},
		{"Prototype pollution attack", "Object prototype tampering risk review", "Object_prototype_tampering_risk_review", 1},
		{"mAsS aSsIgNmEnT vUlNeRaBiLiTy", "unintended field modification risk review", "unintended_field_modification_risk_review", 1},
		{"PROTOTYPE POLLUTION ATTACK", "OBJECT PROTOTYPE TAMPERING RISK REVIEW", "OBJECT_PROTOTYPE_TAMPERING_RISK_REVIEW", 1},
		{"原型污染攻击; 批量赋值漏洞", "对象原型篡改风险核查; 非预期字段写入风险核查", "对象原型篡改风险核查; 非预期字段写入风险核查", 2},
		{"prototype pollution", "prototype pollution", "prototype pollution", 0},
		{"mass assignment", "mass assignment", "mass assignment", 0},
		{"批量赋值 原型污染防护", "批量赋值 原型污染防护", "批量赋值 原型污染防护", 0},
		{"PrototypePollutionProtection", "PrototypePollutionProtection", "PrototypePollutionProtection", 0},
		{"__proto__ constructor prototype CWE-1321 CWE-915", "__proto__ constructor prototype CWE-1321 CWE-915", "__proto__ constructor prototype CWE-1321 CWE-915", 0},
		{"xprototype pollution attack", "xprototype pollution attack", "xprototype pollution attack", 0},
		{"mass assignment vulnerability2", "mass assignment vulnerability2", "mass assignment vulnerability2", 0},
		{"prototype pollution attack_suffix", "prototype pollution attack_suffix", "prototype pollution attack_suffix", 0},
		{"my_mass_assignment_vulnerability", "my_mass_assignment_vulnerability", "my_mass_assignment_vulnerability", 0},
		{"prototype\tpollution attack", "prototype\tpollution attack", "prototype\tpollution attack", 0},
		{`import "prototype pollution attack"`, `import "prototype pollution attack"`, `import "prototype pollution attack"`, 0},
		{"https://example.invalid/原型污染攻击", "https://example.invalid/原型污染攻击", "https://example.invalid/原型污染攻击", 0},
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
