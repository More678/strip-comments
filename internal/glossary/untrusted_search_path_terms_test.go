package glossary

import (
	"strings"
	"testing"
)

func TestUntrustedSearchPathRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"不可信搜索路径漏洞", "不可信搜索路径资源选择风险核查", "不可信搜索路径资源选择风险核查", 1},
		{"untrusted search path vulnerability", "untrusted search path resource selection risk review", "untrusted_search_path_resource_selection_risk_review", 1},
		{"Untrusted search path vulnerability", "Untrusted search path resource selection risk review", "Untrusted_search_path_resource_selection_risk_review", 1},
		{"uNtRuStEd SeArCh PaTh VuLnErAbIlItY", "untrusted search path resource selection risk review", "untrusted_search_path_resource_selection_risk_review", 1},
		{"UNTRUSTED SEARCH PATH VULNERABILITY", "UNTRUSTED SEARCH PATH RESOURCE SELECTION RISK REVIEW", "UNTRUSTED_SEARCH_PATH_RESOURCE_SELECTION_RISK_REVIEW", 1},
		{"不可信搜索路径漏洞; untrusted search path vulnerability", "不可信搜索路径资源选择风险核查; untrusted search path resource selection risk review", "不可信搜索路径资源选择风险核查; untrusted_search_path_resource_selection_risk_review", 2},
		{"review:\nuntrusted search path vulnerability", "review:\nuntrusted search path resource selection risk review", "review:\nuntrusted_search_path_resource_selection_risk_review", 1},
		{"(untrusted search path vulnerability).", "(untrusted search path resource selection risk review).", "(untrusted_search_path_resource_selection_risk_review).", 1},
		{"操作系统命令参数注入攻击", "操作系统命令参数注入风险核查", "操作系统命令参数注入风险核查", 1},
		{"OS command argument injection attack", "OS command argument injection risk review", "OS_command_argument_injection_risk_review", 1},
		{"外部可控反射漏洞", "外部可控反射代码选择风险核查", "外部可控反射代码选择风险核查", 1},
		{"externally controlled reflection vulnerability", "externally controlled reflection code selection risk review", "externally_controlled_reflection_code_selection_risk_review", 1},
	}
	for _, in := range []string{
		"untrusted search path 不可信搜索路径 PATH LD_LIBRARY_PATH SearchPath",
		"LoadLibrary exec.LookPath CWE-426 CWE-427",
		"untrusted search path prevention validateSearchPath SetDefaultDllDirectories",
		"untrusted search path resource selection risk review 不可信搜索路径资源选择风险核查",
		"不可信搜索路径防护 不可信 搜索路径漏洞 不可信搜索路径 漏洞",
		"DLL hijacking attack search order hijacking attack path hijacking attack",
		"untrusted search-path vulnerability untrusted-search-path-vulnerability untrusted_search_path_vulnerability",
		"untrusted search path vulnerabilities",
		"untrusted\tsearch path vulnerability",
		"untrusted  search path vulnerability",
		"untrusted\nsearch path vulnerability",
		"untrusted search path\tvulnerability",
		"xuntrusted search path vulnerability",
		"untrusted search path vulnerabilityx",
		"2untrusted search path vulnerability",
		"untrusted search path vulnerability2",
		"_untrusted search path vulnerability",
		"untrusted search path vulnerability_suffix",
		"前untrusted search path vulnerability",
		"untrusted search path vulnerability后",
		`import "untrusted search path vulnerability"`,
		`import "不可信搜索路径漏洞"`,
		"https://example.invalid/untrusted%20search%20path%20vulnerability",
		"https://example.invalid/不可信搜索路径漏洞",
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
