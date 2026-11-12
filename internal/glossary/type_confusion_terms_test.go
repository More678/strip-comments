package glossary

import (
	"strings"
	"testing"
)

func TestTypeConfusionRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"类型混淆漏洞", "类型混淆风险核查", "类型混淆风险核查", 1},
		{"type confusion vulnerability", "type confusion risk review", "type_confusion_risk_review", 1},
		{"Type confusion vulnerability", "Type confusion risk review", "Type_confusion_risk_review", 1},
		{"tYpE cOnFuSiOn VuLnErAbIlItY", "type confusion risk review", "type_confusion_risk_review", 1},
		{"TYPE CONFUSION VULNERABILITY", "TYPE CONFUSION RISK REVIEW", "TYPE_CONFUSION_RISK_REVIEW", 1},
		{"类型混淆漏洞; type confusion vulnerability", "类型混淆风险核查; type confusion risk review", "类型混淆风险核查; type_confusion_risk_review", 2},
		{"整数溢出漏洞", "整数溢出风险核查", "整数溢出风险核查", 1},
		{"integer underflow vulnerability", "integer underflow risk review", "integer_underflow_risk_review", 1},
		{"内存释放后使用漏洞", "内存释放后使用风险核查", "内存释放后使用风险核查", 1},
		{"double free vulnerability", "double free risk review", "double_free_risk_review", 1},
	}
	for _, in := range []string{
		"type confusion 类型混淆 type 类型 cast 泛型 CWE-843",
		"typeof instanceof dynamic_cast Type.GetType",
		"类型混淆防护 TypeConfusionProtection",
		"类型 混淆漏洞 类型转换错误",
		"type-confusion vulnerability type_confusion_vulnerability",
		"type\tconfusion vulnerability",
		"xtype confusion vulnerability",
		"type confusion vulnerabilityx",
		"2type confusion vulnerability",
		"type confusion vulnerability2",
		"_type confusion vulnerability",
		"type confusion vulnerability_suffix",
		`import "type confusion vulnerability"`,
		`import "类型混淆漏洞"`,
		"https://example.invalid/类型混淆漏洞",
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
