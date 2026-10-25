package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/More678/strip-comments/internal/lang"
	"github.com/More678/strip-comments/internal/scan"
	"github.com/More678/strip-comments/internal/walk"
)

// fixture writes body to a fresh .go file and returns it as a walk.File.
func fixture(t *testing.T, name, body string) walk.File {
	t.Helper()
	l := lang.ByName("go")
	if l == nil {
		t.Fatal("go is not registered")
	}
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return walk.File{Path: path, Lang: l, Size: st.Size(), Mode: st.Mode()}
}

func readBack(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestStripOne(t *testing.T) {
	cases := []struct {
		name        string
		cfg         config
		in          string
		wantFile    string // file content afterwards
		wantChanged bool
		wantSkip    bool
	}{
		{
			// The guard: nothing to strip, so mixed endings stay as they are.
			// Re-encoding would have normalised every line to CRLF.
			name:     "mixed endings without comments are untouched",
			cfg:      config{write: true},
			in:       "package p\r\nvar x = 1\nvar y = 2\r\n",
			wantFile: "package p\r\nvar x = 1\nvar y = 2\r\n",
		},
		{
			name:     "LF only without comments is untouched",
			cfg:      config{write: true},
			in:       "package p\nvar x = 1\n",
			wantFile: "package p\nvar x = 1\n",
		},
		{
			// A real strip re-encodes, so the majority ending (CRLF, 2 of 3)
			// is applied to every line. That is the documented trade-off.
			name:        "mixed endings with a comment are stripped",
			cfg:         config{write: true},
			in:          "package p\r\nvar x = 1 // c\nvar y = 2\r\n",
			wantFile:    "package p\r\nvar x = 1\r\nvar y = 2\r\n",
			wantChanged: true,
		},
		{
			name:        "CRLF with a comment keeps CRLF",
			cfg:         config{write: true},
			in:          "package p\r\n// gone\r\nvar x = 1\r\n",
			wantFile:    "package p\r\nvar x = 1\r\n",
			wantChanged: true,
		},
		{
			// Dry run reports the change but must not write anything.
			name:        "dry run does not write",
			cfg:         config{},
			in:          "package p\nvar x = 1 // c\n",
			wantFile:    "package p\nvar x = 1 // c\n",
			wantChanged: true,
		},
		{
			name:     "binary file is skipped",
			cfg:      config{write: true},
			in:       "package p\x00// c\n",
			wantFile: "package p\x00// c\n",
			wantSkip: true,
		},
		{
			name:     "unsafe input is skipped without force",
			cfg:      config{write: true},
			in:       "package p\n/* never closed\nvar x = 1\n",
			wantFile: "package p\n/* never closed\nvar x = 1\n",
			wantSkip: true,
		},
		{
			name:        "unsafe input is stripped with force",
			cfg:         config{write: true, force: true},
			in:          "package p\n/* never closed\nvar x = 1\n",
			wantFile:    "package p\n",
			wantChanged: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := fixture(t, "a.go", tc.in)
			res := stripOne(f, tc.cfg, scan.Options{Force: tc.cfg.force})
			if res.err != nil {
				t.Fatalf("err = %v", res.err)
			}
			if res.changed != tc.wantChanged {
				t.Errorf("changed = %v, want %v", res.changed, tc.wantChanged)
			}
			if gotSkip := res.skip != ""; gotSkip != tc.wantSkip {
				t.Errorf("skip = %q, want skipped=%v", res.skip, tc.wantSkip)
			}
			if got := readBack(t, f.Path); got != tc.wantFile {
				t.Errorf("file = %q, want %q", got, tc.wantFile)
			}
		})
	}
}

func TestStripOneBackup(t *testing.T) {
	in := "package p\nvar x = 1 // c\n"
	f := fixture(t, "a.go", in)
	res := stripOne(f, config{write: true, backup: true}, scan.Options{})
	if res.err != nil {
		t.Fatalf("err = %v", res.err)
	}
	if !res.changed {
		t.Fatal("changed = false, want true")
	}
	if got := readBack(t, f.Path); got != "package p\nvar x = 1\n" {
		t.Errorf("stripped file = %q", got)
	}
	if got := readBack(t, f.Path+".bak"); got != in {
		t.Errorf("backup = %q, want the original %q", got, in)
	}
}

func TestStripOneNoBackupWhenUnchanged(t *testing.T) {
	f := fixture(t, "a.go", "package p\nvar x = 1\n")
	if res := stripOne(f, config{write: true, backup: true}, scan.Options{}); res.err != nil {
		t.Fatalf("err = %v", res.err)
	}
	if _, err := os.Stat(f.Path + ".bak"); !os.IsNotExist(err) {
		t.Errorf("a .bak was written for an unchanged file (err = %v)", err)
	}
}

func TestStripOneKeepsBOM(t *testing.T) {
	f := fixture(t, "a.go", "\xEF\xBB\xBFpackage p\nvar x = 1 // c\n")
	res := stripOne(f, config{write: true}, scan.Options{})
	if res.err != nil {
		t.Fatalf("err = %v", res.err)
	}
	want := "\xEF\xBB\xBFpackage p\nvar x = 1\n"
	if got := readBack(t, f.Path); got != want {
		t.Errorf("file = %q, want %q", got, want)
	}
}

func TestStripOneCounts(t *testing.T) {
	f := fixture(t, "a.go", "package p\n// one\nvar x = 1 // two\n")
	res := stripOne(f, config{}, scan.Options{})
	if res.comments != 2 {
		t.Errorf("comments = %d, want 2", res.comments)
	}
	if res.lines != 1 {
		t.Errorf("lines = %d, want 1", res.lines)
	}
}

func TestStripOneMissingFile(t *testing.T) {
	f := fixture(t, "a.go", "package p\n")
	if err := os.Remove(f.Path); err != nil {
		t.Fatal(err)
	}
	if res := stripOne(f, config{write: true}, scan.Options{}); res.err == nil {
		t.Error("err = nil, want a read error")
	}
}
