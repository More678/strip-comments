package glossary

import (
	"strings"
	"testing"
)

func TestUnsafeReflectionRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"外部可控反射漏洞", "外部可控反射代码选择风险核查", "外部可控反射代码选择风险核查", 1},
		{"externally controlled reflection vulnerability", "externally controlled reflection code selection risk review", "externally_controlled_reflection_code_selection_risk_review", 1},
		{"Externally controlled reflection vulnerability", "Externally controlled reflection code selection risk review", "Externally_controlled_reflection_code_selection_risk_review", 1},
		{"eXtErNaLlY CoNtRoLlEd ReFlEcTiOn VuLnErAbIlItY", "externally controlled reflection code selection risk review", "externally_controlled_reflection_code_selection_risk_review", 1},
		{"EXTERNALLY CONTROLLED REFLECTION VULNERABILITY", "EXTERNALLY CONTROLLED REFLECTION CODE SELECTION RISK REVIEW", "EXTERNALLY_CONTROLLED_REFLECTION_CODE_SELECTION_RISK_REVIEW", 1},
		{"外部可控反射漏洞; externally controlled reflection vulnerability", "外部可控反射代码选择风险核查; externally controlled reflection code selection risk review", "外部可控反射代码选择风险核查; externally_controlled_reflection_code_selection_risk_review", 2},
		{"review:\nexternally controlled reflection vulnerability", "review:\nexternally controlled reflection code selection risk review", "review:\nexternally_controlled_reflection_code_selection_risk_review", 1},
		{"(externally controlled reflection vulnerability).", "(externally controlled reflection code selection risk review).", "(externally_controlled_reflection_code_selection_risk_review).", 1},
		{"PHP对象注入攻击", "PHP反序列化对象注入风险核查", "PHP反序列化对象注入风险核查", 1},
		{"PHP object injection attack", "PHP deserialization object injection risk review", "PHP_deserialization_object_injection_risk_review", 1},
		{"表达式语言注入攻击", "表达式语言求值注入风险核查", "表达式语言求值注入风险核查", 1},
		{"expression language injection attack", "expression language evaluation injection risk review", "expression_language_evaluation_injection_risk_review", 1},
	}
	for _, in := range []string{
		"unsafe reflection externally controlled reflection reflection 反射 class code 动态加载",
		"Class.forName Method.invoke Type.GetType Activator.CreateInstance CWE-470",
		"externally controlled reflection prevention validateClassName allowlist",
		"externally controlled reflection code selection risk review 外部可控反射代码选择风险核查",
		"外部可控反射防护 外部可控 反射漏洞 外部可控反射 漏洞",
		"unsafe reflection vulnerability reflection vulnerability 不安全反射漏洞",
		"externally-controlled reflection vulnerability externally_controlled_reflection_vulnerability",
		"externally controlled reflection vulnerabilities",
		"externally\tcontrolled reflection vulnerability",
		"externally  controlled reflection vulnerability",
		"externally\ncontrolled reflection vulnerability",
		"externally controlled reflection\tvulnerability",
		"xexternally controlled reflection vulnerability",
		"externally controlled reflection vulnerabilityx",
		"2externally controlled reflection vulnerability",
		"externally controlled reflection vulnerability2",
		"_externally controlled reflection vulnerability",
		"externally controlled reflection vulnerability_suffix",
		"前externally controlled reflection vulnerability",
		"externally controlled reflection vulnerability后",
		`import "externally controlled reflection vulnerability"`,
		`import "外部可控反射漏洞"`,
		"https://example.invalid/externally%20controlled%20reflection%20vulnerability",
		"https://example.invalid/外部可控反射漏洞",
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
