package glossary

import (
	"strings"
	"testing"
)

func TestWebCacheRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"Web缓存投毒", "Web缓存响应完整性风险核查", "Web缓存响应完整性风险核查", 1},
		{"Web缓存欺骗", "Web私有响应错误缓存风险核查", "Web私有响应错误缓存风险核查", 1},
		{"web cache poisoning", "web cache response integrity risk review", "web_cache_response_integrity_risk_review", 1},
		{"web cache deception", "web private response caching risk review", "web_private_response_caching_risk_review", 1},
		{"Web cache poisoning", "Web cache response integrity risk review", "Web_cache_response_integrity_risk_review", 1},
		{"wEb CaChE DeCePtIoN", "web private response caching risk review", "web_private_response_caching_risk_review", 1},
		{"WEB CACHE DECEPTION", "WEB PRIVATE RESPONSE CACHING RISK REVIEW", "WEB_PRIVATE_RESPONSE_CACHING_RISK_REVIEW", 1},
		{"web cache poisoning; web cache deception", "web cache response integrity risk review; web private response caching risk review", "web_cache_response_integrity_risk_review; web_private_response_caching_risk_review", 2},
		{"DNS缓存投毒", "名称解析缓存完整性风险核查", "名称解析缓存完整性风险核查", 1},
		{"ARP缓存投毒", "地址映射完整性风险核查", "地址映射完整性风险核查", 1},
		{"DNS cache poisoning", "name resolution cache integrity risk review", "name_resolution_cache_integrity_risk_review", 1},
		{"ARP cache poisoning", "address mapping integrity risk review", "address_mapping_integrity_risk_review", 1},
		{"cache poisoning", "cache poisoning", "cache poisoning", 0},
		{"cache deception", "cache deception", "cache deception", 0},
		{"web cache", "web cache", "web cache", 0},
		{"缓存投毒 缓存欺骗", "缓存投毒 缓存欺骗", "缓存投毒 缓存欺骗", 0},
		{"Web 缓存投毒", "Web 缓存投毒", "Web 缓存投毒", 0},
		{"web缓存欺骗", "web缓存欺骗", "web缓存欺骗", 0},
		{"web\tcache deception", "web\tcache deception", "web\tcache deception", 0},
		{"xweb cache poisoning", "xweb cache poisoning", "xweb cache poisoning", 0},
		{"web cache deception2", "web cache deception2", "web cache deception2", 0},
		{"web cache poisoning_suffix", "web cache poisoning_suffix", "web cache poisoning_suffix", 0},
		{"my_web_cache_deception", "my_web_cache_deception", "my_web_cache_deception", 0},
		{`import "web cache poisoning"`, `import "web cache poisoning"`, `import "web cache poisoning"`, 0},
		{"https://example.invalid/Web缓存投毒", "https://example.invalid/Web缓存投毒", "https://example.invalid/Web缓存投毒", 0},
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
