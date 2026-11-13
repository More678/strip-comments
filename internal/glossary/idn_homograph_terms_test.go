package glossary

import (
	"strings"
	"testing"
)

func TestIDNHomographRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"IDN homograph attack", "IDN homograph domain spoofing risk review", "IDN_homograph_domain_spoofing_risk_review", 1},
		{"idn homograph attack", "IDN homograph domain spoofing risk review", "IDN_homograph_domain_spoofing_risk_review", 1},
		{"Idn homograph attack", "IDN homograph domain spoofing risk review", "IDN_homograph_domain_spoofing_risk_review", 1},
		{"iDn HoMoGrApH AtTaCk", "IDN homograph domain spoofing risk review", "IDN_homograph_domain_spoofing_risk_review", 1},
		{"IDN HOMOGRAPH ATTACK", "IDN HOMOGRAPH DOMAIN SPOOFING RISK REVIEW", "IDN_HOMOGRAPH_DOMAIN_SPOOFING_RISK_REVIEW", 1},
		{"IDN homograph attack; IDN homograph attack", "IDN homograph domain spoofing risk review; IDN homograph domain spoofing risk review", "IDN_homograph_domain_spoofing_risk_review; IDN_homograph_domain_spoofing_risk_review", 2},
		{"review:\nIDN homograph attack", "review:\nIDN homograph domain spoofing risk review", "review:\nIDN_homograph_domain_spoofing_risk_review", 1},
		{"(IDN homograph attack).", "(IDN homograph domain spoofing risk review).", "(IDN_homograph_domain_spoofing_risk_review).", 1},
		{"软件包拼写抢注攻击", "软件包名称仿冒风险核查", "软件包名称仿冒风险核查", 1},
		{"a package typosquatting attack", "a package name impersonation risk review", "a package_name_impersonation_risk_review", 1},
		{"reverse tabnabbing attack", "reverse tabnabbing risk review", "reverse_tabnabbing_risk_review", 1},
	}
	for _, in := range []string{
		"IDN IDNA Unicode Punycode UTS #39 homograph confusable skeleton IDN homograph",
		"IDN homograph prevention toASCII toUnicode encodePunycode normalizeDomain",
		"IDN homograph domain spoofing risk review",
		"IDN同形攻击 国际化域名同形攻击",
		"homograph attack IDN homoglyph attack",
		"IDN-homograph-attack IDN_homograph_attack",
		"IDN homograph attacks",
		"IDN\thomograph attack",
		"IDN  homograph attack",
		"IDN\nhomograph attack",
		"IDN homograph\tattack",
		"xIDN homograph attack",
		"IDN homograph attackx",
		"2IDN homograph attack",
		"IDN homograph attack2",
		"_IDN homograph attack",
		"IDN homograph attack_suffix",
		"前IDN homograph attack",
		"IDN homograph attack后",
		`import "IDN homograph attack"`,
		"https://example.invalid/IDN%20homograph%20attack",
		"https://example.invalid/例子",
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
