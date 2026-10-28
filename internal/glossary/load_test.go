package glossary

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTemp drops body into a temp file named name and returns its path.
func writeTemp(t *testing.T, name, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoadTSV(t *testing.T) {
	p := writeTemp(t, "terms.tsv", strings.Join([]string{
		"# a comment",
		"",
		"渗透\t威胁检测",
		"pentest\tthreat detection",
		"Exploit\trisk validation\tnocase",
		"backdoor\tanomalous channel\tnoword,ci",
		"  spaced  \tvalue  ",
	}, "\n")+"\n")

	rules, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 5 {
		t.Fatalf("got %d rules, want 5: %+v", len(rules), rules)
	}
	// A CJK term gets neither flag; an ASCII term gets both.
	if rules[0] != (Rule{From: "渗透", To: "威胁检测"}) {
		t.Errorf("rules[0] = %+v", rules[0])
	}
	if rules[1] != (Rule{From: "pentest", To: "threat detection", Word: true, CI: true}) {
		t.Errorf("rules[1] = %+v", rules[1])
	}
	// The flag column overrides the inference.
	if rules[2].CI || !rules[2].Word {
		t.Errorf("rules[2] = %+v, want Word && !CI", rules[2])
	}
	if rules[3].Word || !rules[3].CI {
		t.Errorf("rules[3] = %+v, want CI && !Word", rules[3])
	}
	if rules[4].From != "spaced" || rules[4].To != "value" {
		t.Errorf("rules[4] = %+v, want trimmed", rules[4])
	}
}

func TestLoadJSON(t *testing.T) {
	p := writeTemp(t, "terms.json",
		`[{"from":"渗透","to":"威胁检测"},
		  {"from":"pentest","to":"threat detection"},
		  {"from":"exploit","to":"risk validation","word":false,"ci":false}]`)

	rules, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 3 {
		t.Fatalf("got %d rules, want 3", len(rules))
	}
	if rules[0].Word || rules[0].CI {
		t.Errorf("rules[0] = %+v, want no flags for a CJK term", rules[0])
	}
	if !rules[1].Word || !rules[1].CI {
		t.Errorf("rules[1] = %+v, want both flags inferred", rules[1])
	}
	if rules[2].Word || rules[2].CI {
		t.Errorf("rules[2] = %+v, want the explicit false to win", rules[2])
	}
}

func TestLoadErrors(t *testing.T) {
	cases := []struct {
		name, file, body, want string
	}{
		{"missing tab", "t.tsv", "渗透 威胁检测\n", "want from<TAB>to"},
		{"empty term", "t.tsv", "\t威胁检测\n", "empty term"},
		{"unknown flag", "t.tsv", "a\tb\tnope\n", "unknown flag"},
		{"bad json", "t.json", "{not json", "t.json"},
		{"empty json from", "t.json", `[{"from":"","to":"x"}]`, "empty \"from\""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Load(writeTemp(t, tc.file, tc.body))
			if err == nil {
				t.Fatal("got nil error, want a failure")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not mention %q", err, tc.want)
			}
		})
	}
}

func TestLoadErrorCarriesLineNumber(t *testing.T) {
	p := writeTemp(t, "terms.tsv", "a\tb\n\n# c\nbroken line\n")
	_, err := Load(p)
	if err == nil {
		t.Fatal("got nil error")
	}
	// The bad line is line 4, counting the blank and the comment.
	if !strings.Contains(err.Error(), ":4:") {
		t.Errorf("error %q should point at line 4", err)
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "nope.tsv")); err == nil {
		t.Fatal("got nil error for a missing file")
	}
}
