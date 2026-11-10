package glossary

import "testing"

func TestAuthenticationAndSessionTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct{ in, want string }{
		{"密码喷洒", "批量认证风险核查"},
		{"口令喷洒", "批量认证风险核查"},
		{"浏览器会话劫持", "浏览器会话完整性核查"},
		{"远程会话劫持", "远程会话完整性核查"},
		{"password spraying", "bulk authentication review"},
		{"password spray", "bulk authentication review"},
		{"Password spraying", "Bulk authentication review"},
		{"browser session hijacking", "browser session integrity review"},
		{"remote service session hijacking", "remote session integrity review"},
		{"password sprayer", "password sprayer"},
		{"my_password_spraying", "my_password_spraying"},
		{"password spraying2", "password spraying2"},
		{"browser session", "browser session"},
		{"喷洒系统", "喷洒系统"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			opt := Options{Comments: []Region{{Start: 0, End: len(tc.in), Comment: true}}}
			out, _ := tb.Apply([]byte(tc.in), opt)
			if string(out) != tc.want {
				t.Fatalf("got %q, want %q", out, tc.want)
			}
			opt.Comments[0].End = len(out)
			again, hits := tb.Apply(out, opt)
			if string(again) != string(out) || len(hits) != 0 {
				t.Fatalf("second pass changed %q to %q: %+v", out, again, hits)
			}
		})
	}
}
