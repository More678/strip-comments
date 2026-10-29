package udiff

import (
	"fmt"
	"strings"
	"testing"
)

// apply replays an edit script onto a, which is the strongest check that the
// script is right: whatever the hunks look like, keeps and inserts in order
// must rebuild b exactly.
func apply(ops []op) []string {
	var out []string
	for _, o := range ops {
		if o.kind != '-' {
			out = append(out, o.text)
		}
	}
	return out
}

func TestScriptRebuildsTarget(t *testing.T) {
	cases := [][2]string{
		{"a\nb\nc\n", "a\nc\n"},
		{"a\nb\nc\n", "a\nb\nc\nd\n"},
		{"", "x\n"},
		{"x\n", ""},
		{"a\nb\nc\nd\ne\n", "a\nX\nc\nY\ne\n"},
		{"// c\nint x;\n// d\nint y;\n", "int x;\nint y;\n"},
	}
	for _, c := range cases {
		a, b := splitLines([]byte(c[0])), splitLines([]byte(c[1]))
		got := apply(script(a, b))
		if strings.Join(got, "\n") != strings.Join(b, "\n") {
			t.Errorf("script(%q -> %q) rebuilt %q", c[0], c[1], got)
		}
	}
}

func TestUnifiedIdentical(t *testing.T) {
	if d := Unified("f", []byte("a\nb\n"), []byte("a\nb\n")); d != "" {
		t.Errorf("identical input gave a diff:\n%s", d)
	}
}

func TestUnifiedShape(t *testing.T) {
	before := "package x\n\n// note\nvar n = 1\n"
	after := "package x\n\nvar n = 1\n"
	d := Unified("a.go", []byte(before), []byte(after))

	for _, want := range []string{
		"--- a.go\n", "+++ a.go\n",
		"@@ -1,4 +1,3 @@\n",
		"-// note\n", " var n = 1\n",
	} {
		if !strings.Contains(d, want) {
			t.Errorf("diff lacks %q:\n%s", want, d)
		}
	}
}

// TestUnifiedSplitsDistantHunks: edits far apart get their own hunk, close
// ones share one, so the output stays readable on a long file.
func TestUnifiedSplitsDistantHunks(t *testing.T) {
	var a, b []string
	for i := 1; i <= 40; i++ {
		line := fmt.Sprintf("line %d", i)
		a = append(a, line)
		if i != 5 && i != 35 {
			b = append(b, line)
		}
	}
	d := Unified("f", []byte(strings.Join(a, "\n")+"\n"), []byte(strings.Join(b, "\n")+"\n"))
	if n := strings.Count(d, "@@ -"); n != 2 {
		t.Errorf("got %d hunks, want 2 for edits 30 lines apart:\n%s", n, d)
	}
}

func TestUnifiedLargeFileSummary(t *testing.T) {
	big := strings.Repeat("x\n", maxLCS+1)
	d := Unified("big", []byte(big), []byte(big+"y\n"))
	if !strings.Contains(d, "lines ->") {
		t.Errorf("large file should get a summary, got:\n%.200s", d)
	}
}
