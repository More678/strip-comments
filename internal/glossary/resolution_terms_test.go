package glossary

import "testing"

func TestResolutionRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct{ in, want string }{
		{"DNS缓存投毒", "名称解析缓存完整性风险核查"},
		{"ARP缓存投毒", "地址映射完整性风险核查"},
		{"DNS cache poisoning", "name resolution cache integrity risk review"},
		{"arp cache poisoning", "address mapping integrity risk review"},
		{"DNS CACHE POISONING", "NAME RESOLUTION CACHE INTEGRITY RISK REVIEW"},
		{"DNS cache", "DNS cache"},
		{"ARP table", "ARP table"},
		{"缓存管理", "缓存管理"},
		{"my_DNS_cache_poisoning", "my_DNS_cache_poisoning"},
		{"DNS cache poisoning2", "DNS cache poisoning2"},
		{"xARP cache poisoning", "xARP cache poisoning"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			for _, comment := range []bool{false, true} {
				opt := Options{}
				want := tc.want
				if comment {
					opt.Comments = []Region{{Start: 0, End: len(tc.in), Comment: true}}
				} else if tc.want != tc.in {
					want = toIdent(want)
				}
				out, _ := tb.Apply([]byte(tc.in), opt)
				if string(out) != want {
					t.Fatalf("comment=%v: got %q, want %q", comment, out, want)
				}
				if comment {
					opt.Comments[0].End = len(out)
				}
				again, hits := tb.Apply(out, opt)
				if string(again) != string(out) || len(hits) != 0 {
					t.Fatalf("second pass changed output: %q (%+v)", again, hits)
				}
			}
		})
	}
}
