package glossary

import "testing"

func TestApplyDefault(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		name, in, want string
	}{
		{"longest wins", "渗透测试报告", "威胁检测报告"},
		{"bare term", "开始渗透", "开始威胁检测"},
		// Outside a comment the replacement has to stay a legal identifier,
		// so a multi-word term is joined with underscores.
		{"english word", "run pentest now", "run threat_detection now"},
		{"case carried up", "PENTEST", "THREAT_DETECTION"},
		{"case carried title", "Pentest", "Threat_detection"},
		{"no partial word", "pentester", "pentester"},
		{"import untouched", "import \"exploit/pkg\"\n", "import \"exploit/pkg\"\n"},
		{"url untouched", "see https://x.io/exploit\n", "see https://x.io/exploit\n"},
		{"url then code", "// https://x.io/exploit\nvar exploit = 1\n",
			"// https://x.io/exploit\nvar risk_validation = 1\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := tb.Apply([]byte(tc.in), Options{})
			if string(got) != tc.want {
				t.Errorf("got  %q\nwant %q", got, tc.want)
			}
		})
	}
}

func TestApplyNoHitsReturnsInput(t *testing.T) {
	tb := NewTable(Default)
	in := []byte("package main\n")
	got, hits := tb.Apply(in, Options{})
	if len(hits) != 0 {
		t.Errorf("hits = %v, want none", hits)
	}
	if string(got) != string(in) {
		t.Errorf("got %q, want %q", got, in)
	}
}

func TestCommentsOnly(t *testing.T) {
	tb := NewTable(Default)
	src := []byte("// 渗透\nvar 渗透 = 1\n")
	// The comment runs from 0 to the newline at index 8.
	got, hits := tb.Apply(src, Options{
		Comments:     []Region{{Start: 0, End: 8, Comment: true}},
		CommentsOnly: true,
	})
	if want := "// 威胁检测\nvar 渗透 = 1\n"; string(got) != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
	if len(hits) != 1 || !hits[0].InComment {
		t.Errorf("hits = %+v, want one in-comment hit", hits)
	}
}

// TestStringLiteralGetsIdentForm pins a known cost rather than a bug: the
// glossary sees comment ranges and nothing else, so it cannot tell a string
// literal from an identifier and applies toIdent to both. The result still
// compiles; the string just reads less naturally.
//
// Fixing it means teaching the scanner to hand over string spans too. Until
// then this test exists so the behaviour is not "fixed" by accident.
func TestStringLiteralGetsIdentForm(t *testing.T) {
	tb := NewTable(Default)
	got, _ := tb.Apply([]byte(`msg := "pentest failed"`), Options{})
	if want := `msg := "threat_detection failed"`; string(got) != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}
