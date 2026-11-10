package glossary

import (
	"strings"
	"testing"
)

func TestDependencyRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"依赖混淆攻击", "依赖来源解析风险核查", "依赖来源解析风险核查", 1},
		{"软件包拼写抢注攻击", "软件包名称仿冒风险核查", "软件包名称仿冒风险核查", 1},
		{"dependency confusion attack", "dependency source resolution risk review", "dependency_source_resolution_risk_review", 1},
		{"a package typosquatting attack", "a package name impersonation risk review", "a package_name_impersonation_risk_review", 1},
		{"Dependency confusion attack", "Dependency source resolution risk review", "Dependency_source_resolution_risk_review", 1},
		{"dEpEnDeNcY cOnFuSiOn AtTaCk", "dependency source resolution risk review", "dependency_source_resolution_risk_review", 1},
		{"a PACKAGE TYPOSQUATTING ATTACK", "a PACKAGE NAME IMPERSONATION RISK REVIEW", "a PACKAGE_NAME_IMPERSONATION_RISK_REVIEW", 1},
		{"依赖混淆攻击; 软件包拼写抢注攻击", "依赖来源解析风险核查; 软件包名称仿冒风险核查", "依赖来源解析风险核查; 软件包名称仿冒风险核查", 2},
		{"供应链攻击", "供应链风险评估", "供应链风险评估", 1},
		{"dependency confusion", "dependency confusion", "dependency confusion", 0},
		{"typosquatting", "typosquatting", "typosquatting", 0},
		{"依赖混淆防护 软件包拼写抢注防护", "依赖混淆防护 软件包拼写抢注防护", "依赖混淆防护 软件包拼写抢注防护", 0},
		{"requests lodash dependency-confusion package-typosquatting", "requests lodash dependency-confusion package-typosquatting", "requests lodash dependency-confusion package-typosquatting", 0},
		{"xdependency confusion attack", "xdependency confusion attack", "xdependency confusion attack", 0},
		{"dependency confusion attack2", "dependency confusion attack2", "dependency confusion attack2", 0},
		{"a package typosquatting attack_suffix", "a package typosquatting attack_suffix", "a package typosquatting attack_suffix", 0},
		{"my_dependency_confusion_attack", "my_dependency_confusion_attack", "my_dependency_confusion_attack", 0},
		{"dependency\tconfusion attack", "dependency\tconfusion attack", "dependency\tconfusion attack", 0},
		{`import "dependency confusion attack"`, `import "dependency confusion attack"`, `import "dependency confusion attack"`, 0},
		{"https://example.invalid/依赖混淆攻击", "https://example.invalid/依赖混淆攻击", "https://example.invalid/依赖混淆攻击", 0},
		// The existing package-line guard wins, even in prose at line start.
		{"package typosquatting attack", "package typosquatting attack", "package typosquatting attack", 0},
		{"// package typosquatting attack", "// package typosquatting attack", "// package typosquatting attack", 0},
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
