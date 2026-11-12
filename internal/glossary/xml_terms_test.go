package glossary

import (
	"strings"
	"testing"
)

func TestXMLRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"XML外部实体攻击", "XML外部实体解析风险核查", "XML外部实体解析风险核查", 1},
		{"XML递归实体展开攻击", "XML递归实体展开资源耗尽风险核查", "XML递归实体展开资源耗尽风险核查", 1},
		{"XML external entity attack", "XML external entity resolution risk review", "XML_external_entity_resolution_risk_review", 1},
		{"XML recursive entity expansion attack", "XML recursive entity expansion resource exhaustion risk review", "XML_recursive_entity_expansion_resource_exhaustion_risk_review", 1},
		{"xml external entity attack", "XML external entity resolution risk review", "XML_external_entity_resolution_risk_review", 1},
		{"xMl ExTeRnAl EnTiTy AtTaCk", "XML external entity resolution risk review", "XML_external_entity_resolution_risk_review", 1},
		{"XML RECURSIVE ENTITY EXPANSION ATTACK", "XML RECURSIVE ENTITY EXPANSION RESOURCE EXHAUSTION RISK REVIEW", "XML_RECURSIVE_ENTITY_EXPANSION_RESOURCE_EXHAUSTION_RISK_REVIEW", 1},
		{"XML外部实体攻击; XML递归实体展开攻击", "XML外部实体解析风险核查; XML递归实体展开资源耗尽风险核查", "XML外部实体解析风险核查; XML递归实体展开资源耗尽风险核查", 2},
		{"XML external entity; XML recursive entity expansion", "XML external entity; XML recursive entity expansion", "XML external entity; XML recursive entity expansion", 0},
		{"XML XXE DTD entity 实体 展开", "XML XXE DTD entity 实体 展开", "XML XXE DTD entity 实体 展开", 0},
		{"XML外部实体防护 XML递归实体展开限制", "XML外部实体防护 XML递归实体展开限制", "XML外部实体防护 XML递归实体展开限制", 0},
		{"CWE-611 CWE-776 DOCTYPE ENTITY", "CWE-611 CWE-776 DOCTYPE ENTITY", "CWE-611 CWE-776 DOCTYPE ENTITY", 0},
		{"disallow-doctype-decl external-general-entities external-parameter-entities", "disallow-doctype-decl external-general-entities external-parameter-entities", "disallow-doctype-decl external-general-entities external-parameter-entities", 0},
		{"XML bomb Billion Laughs XML entity expansion attack", "XML bomb Billion Laughs XML entity expansion attack", "XML bomb Billion Laughs XML entity expansion attack", 0},
		{"xml外部实体攻击 XML 外部实体攻击", "xml外部实体攻击 XML 外部实体攻击", "xml外部实体攻击 XML 外部实体攻击", 0},
		{"xml递归实体展开攻击 XML 递归实体展开攻击", "xml递归实体展开攻击 XML 递归实体展开攻击", "xml递归实体展开攻击 XML 递归实体展开攻击", 0},
		{"xXML external entity attack", "xXML external entity attack", "xXML external entity attack", 0},
		{"XML external entity attacks", "XML external entity attacks", "XML external entity attacks", 0},
		{"2XML recursive entity expansion attack", "2XML recursive entity expansion attack", "2XML recursive entity expansion attack", 0},
		{"XML recursive entity expansion attack2", "XML recursive entity expansion attack2", "XML recursive entity expansion attack2", 0},
		{"_XML external entity attack", "_XML external entity attack", "_XML external entity attack", 0},
		{"XML recursive entity expansion attack_suffix", "XML recursive entity expansion attack_suffix", "XML recursive entity expansion attack_suffix", 0},
		{"XML_external_entity_attack", "XML_external_entity_attack", "XML_external_entity_attack", 0},
		{"XML\texternal entity attack", "XML\texternal entity attack", "XML\texternal entity attack", 0},
		{`import "XML external entity attack"`, `import "XML external entity attack"`, `import "XML external entity attack"`, 0},
		{`import "XML recursive entity expansion attack"`, `import "XML recursive entity expansion attack"`, `import "XML recursive entity expansion attack"`, 0},
		{"https://example.invalid/XML外部实体攻击", "https://example.invalid/XML外部实体攻击", "https://example.invalid/XML外部实体攻击", 0},
		{"https://example.invalid/XML递归实体展开攻击", "https://example.invalid/XML递归实体展开攻击", "https://example.invalid/XML递归实体展开攻击", 0},
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
