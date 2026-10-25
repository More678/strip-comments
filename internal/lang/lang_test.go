package lang

import (
	"strings"
	"testing"
)

func TestLookups(t *testing.T) {
	cases := []struct {
		how, key, want string
	}{
		{"ext", ".go", "go"},
		{"ext", ".h", "c"},
		{"ext", ".c", "c"},
		{"ext", ".hpp", "cpp"},
		{"ext", ".cpp", "cpp"},
		{"ext", ".py", "python"},
		{"ext", ".bat", "bat"},
		{"ext", ".cmd", "bat"},
		{"ext", ".sh", "shell"},
		{"ext", ".js", "javascript"},
		{"ext", ".rs", "rust"},
		{"ext", ".java", "java"},
		{"ext", ".GO", "go"}, // lookups are case-insensitive
		{"file", "Makefile", "makefile"},
		{"file", "makefile", "makefile"},
		{"file", "main.go", "go"}, // falls back to the extension
		{"file", "x.tar.py", "python"},
		{"name", "Go", "go"},
		{"name", "PYTHON", "python"},
	}
	for _, tc := range cases {
		var l *Language
		switch tc.how {
		case "ext":
			l = ByExt(tc.key)
		case "file":
			l = ByFilename(tc.key)
		case "name":
			l = ByName(tc.key)
		}
		if l == nil {
			t.Errorf("%s(%q) = nil, want %q", tc.how, tc.key, tc.want)
			continue
		}
		if l.Name != tc.want {
			t.Errorf("%s(%q).Name = %q, want %q", tc.how, tc.key, l.Name, tc.want)
		}
	}
}

func TestLookupMisses(t *testing.T) {
	if l := ByExt(".no-such-ext"); l != nil {
		t.Errorf("ByExt(unknown) = %q, want nil", l.Name)
	}
	if l := ByName("no-such-lang"); l != nil {
		t.Errorf("ByName(unknown) = %q, want nil", l.Name)
	}
	// No dot past position 0: no extension fallback.
	for _, base := range []string{"README", ".nosuchrc"} {
		if l := ByFilename(base); l != nil {
			t.Errorf("ByFilename(%q) = %q, want nil", base, l.Name)
		}
	}
	// Deliberately unregistered, because the extension is shared with a
	// language whose comment syntax would destroy code: `.pp` is both Pascal
	// and Puppet, and Pascal's `{ }` block comment would delete the body of
	// every Puppet resource. Same reasoning for `.lhs` and `.bib`.
	for _, ext := range []string{".pp", ".lhs", ".bib", ".inc"} {
		if l := ByExt(ext); l != nil {
			t.Errorf("ByExt(%q) = %q, want nil: see .reports/lang-table.md", ext, l.Name)
		}
	}
}

// TestCaseInsensitiveExt checks the lower-casing in ByExt: an upper-case
// extension resolves to the language that registered the lower-case form.
func TestCaseInsensitiveExt(t *testing.T) {
	for ext, want := range map[string]string{
		".C": "c", ".PY": "python", ".Go": "go", ".BAT": "bat",
	} {
		l := ByExt(ext)
		if l == nil {
			t.Errorf("ByExt(%q) = nil, want %q", ext, want)
			continue
		}
		if l.Name != want {
			t.Errorf("ByExt(%q).Name = %q, want %q", ext, l.Name, want)
		}
	}
	// `.s`/`.S` stay unregistered on purpose: GAS uses `;` as a statement
	// separator, so stripping it would delete code (`rep; movsb`).
	for _, ext := range []string{".s", ".S"} {
		if l := ByExt(ext); l != nil {
			t.Errorf("ByExt(%q) = %q, want nil", ext, l.Name)
		}
	}
}

// TestExtlessLanguage pins matlab down: `.m` belongs to objc, so matlab
// carries no extension and is reachable only by name, opt-in via --lang.
func TestExtlessLanguage(t *testing.T) {
	l := ByName("matlab")
	if l == nil {
		t.Fatal(`ByName("matlab") = nil`)
	}
	if len(l.Exts) != 0 {
		t.Errorf("matlab.Exts = %v, want none (.m is objc)", l.Exts)
	}
	if !l.Risky {
		t.Error("matlab.Risky = false, want true")
	}
	if objc := ByExt(".m"); objc == nil {
		t.Error(`ByExt(".m") = nil, want objc`)
	} else if objc.Name != "objc" {
		t.Errorf(`ByExt(".m").Name = %q, want "objc"`, objc.Name)
	}
}

// TestBlockEnds: an empty End is only legal for the exact-match empty
// comments. They exist because `/**` is a doc opener: without a dedicated
// rule, `/**/` would match `/**` and then search for the next `*/` in the
// file, deleting all the code in between. Any other empty End is a typo.
func TestBlockEnds(t *testing.T) {
	emptyOK := map[string]bool{"/**/": true, "/++/": true}
	for _, l := range All() {
		for _, b := range l.Block {
			if b.Start == "" {
				t.Errorf("%s: block rule with empty Start", l.Name)
			}
			if b.End == "" && !emptyOK[b.Start] {
				t.Errorf("%s: block rule %q has an empty End", l.Name, b.Start)
			}
		}
		for _, lc := range l.Line {
			if lc.Token == "" {
				t.Errorf("%s: line rule with empty Token", l.Name)
			}
		}
	}
}

func TestRegistryInvariants(t *testing.T) {
	all := All()
	if len(all) == 0 {
		t.Fatal("All() is empty")
	}
	seenName := map[string]bool{}
	for _, l := range all {
		if l.Name == "" {
			t.Error("a language has an empty Name")
			continue
		}
		if l.Name != strings.ToLower(l.Name) {
			t.Errorf("%q: Name must be lower case", l.Name)
		}
		if seenName[l.Name] {
			t.Errorf("%q: duplicate language name", l.Name)
		}
		seenName[l.Name] = true

		if ByName(l.Name) != l {
			t.Errorf("%q: ByName does not round-trip", l.Name)
		}
		for _, e := range l.Exts {
			switch {
			case !strings.HasPrefix(e, "."):
				t.Errorf("%s: ext %q must start with a dot", l.Name, e)
			case e != strings.ToLower(e):
				t.Errorf("%s: ext %q must be lower case", l.Name, e)
			case len(e) < 2:
				t.Errorf("%s: ext %q is just a dot", l.Name, e)
			}
			// register() lets a later language overwrite an extension, so a
			// failure here means two languages claim the same one.
			if got := ByExt(e); got != l {
				name := "nil"
				if got != nil {
					name = got.Name
				}
				t.Errorf("%s: ByExt(%q) = %s, want %s", l.Name, e, name, l.Name)
			}
		}
		for _, f := range l.Filenames {
			if got := ByFilename(f); got != l {
				name := "nil"
				if got != nil {
					name = got.Name
				}
				t.Errorf("%s: ByFilename(%q) = %s, want %s", l.Name, f, name, l.Name)
			}
		}
		if len(l.Line) == 0 && len(l.Block) == 0 {
			t.Errorf("%s: no comment syntax at all", l.Name)
		}
	}
}

func TestKeepRE(t *testing.T) {
	goL := ByName("go")
	if goL == nil {
		t.Skip("go is not registered")
	}
	re := goL.KeepRE()
	if re == nil {
		t.Fatal("go has Keep patterns but KeepRE() is nil")
	}
	for _, text := range []string{"//go:build linux", "//go:generate stringer",
		"// #include <stdio.h>"} {
		if re.FindString(text) == "" {
			t.Errorf("go KeepRE does not match %q", text)
		}
	}
	if s := re.FindString("// a normal comment"); s != "" {
		t.Errorf("go KeepRE matched %q in a normal comment", s)
	}
	// A language without Keep patterns must report nil, not an empty regexp
	// that matches everything.
	for _, l := range All() {
		if len(l.Keep) == 0 && l.KeepRE() != nil {
			t.Errorf("%s: KeepRE() is non-nil without Keep patterns", l.Name)
		}
		if len(l.Keep) > 0 && l.KeepRE() == nil {
			t.Errorf("%s: KeepRE() is nil with %d Keep patterns", l.Name, len(l.Keep))
		}
	}
}
