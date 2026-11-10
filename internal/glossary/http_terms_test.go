package glossary

import (
	"strings"
	"testing"
)

func TestHTTPRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"HTTP请求走私", "HTTP请求解析差异风险核查", "HTTP请求解析差异风险核查", 1},
		{"HTTP响应拆分", "HTTP响应头换行处理风险核查", "HTTP响应头换行处理风险核查", 1},
		{"HTTP request smuggling", "HTTP request parsing inconsistency risk review", "HTTP_request_parsing_inconsistency_risk_review", 1},
		{"http response splitting", "HTTP response header newline handling risk review", "HTTP_response_header_newline_handling_risk_review", 1},
		{"hTtP ReQuEsT SmUgGlInG", "HTTP request parsing inconsistency risk review", "HTTP_request_parsing_inconsistency_risk_review", 1},
		{"HTTP RESPONSE SPLITTING", "HTTP RESPONSE HEADER NEWLINE HANDLING RISK REVIEW", "HTTP_RESPONSE_HEADER_NEWLINE_HANDLING_RISK_REVIEW", 1},
		{"HTTP request smuggling; HTTP response splitting", "HTTP request parsing inconsistency risk review; HTTP response header newline handling risk review", "HTTP_request_parsing_inconsistency_risk_review; HTTP_response_header_newline_handling_risk_review", 2},
		{"HTTP request", "HTTP request", "HTTP request", 0},
		{"request smuggling", "request smuggling", "request smuggling", 0},
		{"HTTP 请求走私", "HTTP 请求走私", "HTTP 请求走私", 0},
		{"HTTP\tresponse splitting", "HTTP\tresponse splitting", "HTTP\tresponse splitting", 0},
		{"xHTTP request smuggling", "xHTTP request smuggling", "xHTTP request smuggling", 0},
		{"HTTP response splitting2", "HTTP response splitting2", "HTTP response splitting2", 0},
		{"HTTP request smuggling_suffix", "HTTP request smuggling_suffix", "HTTP request smuggling_suffix", 0},
		{"my_HTTP_request_smuggling", "my_HTTP_request_smuggling", "my_HTTP_request_smuggling", 0},
		{"CWE-444 CWE-113 Content-Length Transfer-Encoding", "CWE-444 CWE-113 Content-Length Transfer-Encoding", "CWE-444 CWE-113 Content-Length Transfer-Encoding", 0},
		{`import "HTTP request smuggling"`, `import "HTTP request smuggling"`, `import "HTTP request smuggling"`, 0},
		{"https://example.invalid/HTTP请求走私", "https://example.invalid/HTTP请求走私", "https://example.invalid/HTTP请求走私", 0},
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
