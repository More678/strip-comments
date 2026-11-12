package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultBackupAndRestore(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		body string
	}{
		{"comments", nil, "\xef\xbb\xbfpackage p\r\n// gone\nvar x = 1\r\n"},
		{"terms", []string{"-replace", "-keep-comments"}, "\xef\xbb\xbfpackage p\r\n// pentest\nvar pentest = 1\r\n"},
		{"both", []string{"-replace"}, "package p\r\n// pentest\r\nvar pentest = 1\r\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := fixture(t, "a.go", tc.body)
			args := append(append([]string{"-w"}, tc.args...), f.Path)
			if code, out := runCLI(t, args...); code != exitOK {
				t.Fatalf("write exit=%d: %s", code, out)
			}
			changed := readBack(t, f.Path)
			if changed == tc.body || readBack(t, f.Path+".bak") != tc.body {
				t.Fatal("write must change source and back up exact original bytes")
			}
			if code, _ := runCLI(t, args...); code != exitOK {
				t.Fatal("unchanged rerun must succeed despite existing backup")
			}
			if readBack(t, f.Path+".bak") != tc.body {
				t.Fatal("rerun changed backup")
			}
			if code, out := runCLI(t, "-restore", f.Path); code != exitOK || !strings.Contains(out, "would restore") {
				t.Fatalf("restore preview exit=%d: %s", code, out)
			}
			if readBack(t, f.Path) != changed {
				t.Fatal("restore preview wrote source")
			}
			assertAbsent(t, f.Path+".restore.bak")
			if code, _ := runCLI(t, "-restore", "-w", f.Path); code != exitOK {
				t.Fatal("restore failed")
			}
			if readBack(t, f.Path) != tc.body || readBack(t, f.Path+".bak") != tc.body || readBack(t, f.Path+".restore.bak") != changed {
				t.Fatal("restore did not preserve all three byte sequences")
			}
			if code, _ := runCLI(t, "-restore", "-w", f.Path); code != exitOK {
				t.Fatal("already-restored file must be a successful no-op")
			}
		})
	}
}

func TestBackupConflictKeepsSource(t *testing.T) {
	for _, directory := range []bool{false, true} {
		t.Run(fmt.Sprint(directory), func(t *testing.T) {
			f := fixture(t, "a.go", "package p\n// new comment\n")
			if directory {
				if err := os.Mkdir(f.Path+".bak", 0o700); err != nil {
					t.Fatal(err)
				}
			} else {
				writeTestFile(t, f.Path+".bak", "older original")
			}
			if code, _ := runCLI(t, "-w", f.Path); code != exitError {
				t.Fatalf("existing backup should reject write, got %d", code)
			}
			if readBack(t, f.Path) != "package p\n// new comment\n" {
				t.Fatal("source changed after backup failure")
			}
			if !directory && readBack(t, f.Path+".bak") != "older original" {
				t.Fatal("existing backup overwritten")
			}
		})
	}
}

func TestRestorePreservesLaterEditsAndRejectsConflict(t *testing.T) {
	f := fixture(t, "a.go", "package p\n// original\n")
	if code, _ := runCLI(t, "-w", f.Path); code != exitOK {
		t.Fatal("initial write failed")
	}
	later := "\xef\xbb\xbfpackage p\r\n// user edited\nvar y = 7\r\n"
	writeTestFile(t, f.Path, later)
	if code, _ := runCLI(t, "-restore", "-w", f.Path); code != exitOK {
		t.Fatal("restore with later edits failed")
	}
	if readBack(t, f.Path+".restore.bak") != later {
		t.Fatal("later edits were lost")
	}
	writeTestFile(t, f.Path, "another edit")
	for _, args := range [][]string{{"-restore", f.Path}, {"-restore", "-w", f.Path}} {
		if code, _ := runCLI(t, args...); code != exitError {
			t.Fatalf("restore conflict should fail: %v", args)
		}
	}
	if readBack(t, f.Path) != "another edit" || readBack(t, f.Path+".restore.bak") != later {
		t.Fatal("restore conflict changed a file")
	}
}

func TestBackupReadOnlyModesAndOptOut(t *testing.T) {
	for _, args := range [][]string{
		{}, {"-backup"}, {"-diff"}, {"-check"}, {"-stdout"},
		{"-replace", "-keep-comments"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			body := "package p\n// pentest\n"
			f := fixture(t, "a.go", body)
			runCLI(t, append(args, f.Path)...)
			if readBack(t, f.Path) != body {
				t.Fatal("read-only mode wrote source")
			}
			assertAbsent(t, f.Path+".bak")
			assertAbsent(t, f.Path+".restore.bak")
		})
	}
	f := fixture(t, "a.go", "package p\n// gone\n")
	if code, _ := runCLI(t, "-w", "-backup=false", f.Path); code != exitOK {
		t.Fatal("explicit opt-out failed")
	}
	assertAbsent(t, f.Path+".bak")
	if readBack(t, f.Path) != "package p\n" {
		t.Fatal("opt-out did not apply change")
	}
	if code, _ := runCLI(t, "-w", "-replace", f.Path); code != exitOK {
		t.Fatal("no-op failed")
	}
	assertAbsent(t, f.Path+".bak")
}

func TestRestoreFlagConflicts(t *testing.T) {
	for _, flags := range [][]string{
		{"-replace"}, {"-backup"}, {"-backup=false"}, {"-stdout"}, {"-diff"}, {"-check"},
		{"-keep-comments"}, {"-terms", "missing.tsv"}, {"-lang", "go"}, {"-ext", ".go"},
		{"-exclude", "x"}, {"-force"}, {"-no-verify"}, {"-list"}, {"-version"},
	} {
		t.Run(strings.Join(flags, " "), func(t *testing.T) {
			f := fixture(t, "a.go", "package p\n")
			writeTestFile(t, f.Path+".bak", "package p\n// old\n")
			args := append(append([]string{"-restore", "-w"}, flags...), f.Path)
			if code, _ := runCLI(t, args...); code != exitUsage {
				t.Fatalf("conflicting flags exit=%d", code)
			}
			if readBack(t, f.Path) != "package p\n" {
				t.Fatal("conflicting flags wrote source")
			}
			assertAbsent(t, f.Path+".restore.bak")
		})
	}
	if code, _ := runCLI(t, "-restore", "-w"); code != exitUsage {
		t.Fatal("restore must require explicit paths")
	}
}

func TestRestoreErrors(t *testing.T) {
	for _, name := range []string{"missing backup", "backup directory", "source directory", "backup path", "missing source", "credential", "pre-restore directory"} {
		t.Run(name, func(t *testing.T) {
			f := fixture(t, "a.go", "package p\n")
			path := f.Path
			switch name {
			case "missing backup":
			case "backup directory":
				if err := os.Mkdir(path+".bak", 0o700); err != nil {
					t.Fatal(err)
				}
			case "source directory":
				path = filepath.Dir(path)
			case "backup path":
				path += ".bak"
				writeTestFile(t, path, "original")
			case "missing source":
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				writeTestFile(t, path+".bak", "original")
			case "credential":
				path = filepath.Join(filepath.Dir(path), "credentials.json")
				writeTestFile(t, path, "current")
				writeTestFile(t, path+".bak", "original")
			case "pre-restore directory":
				writeTestFile(t, path+".bak", "original")
				if err := os.Mkdir(path+".restore.bak", 0o700); err != nil {
					t.Fatal(err)
				}
			}
			if code, _ := runCLI(t, "-restore", "-w", path); code != exitError {
				t.Fatalf("restore should fail, exit=%d", code)
			}
			if name != "missing source" && readBack(t, f.Path) != "package p\n" {
				t.Fatal("failed restore changed source")
			}
		})
	}
}

func TestRestoreRawBytesAndCredentialOptIn(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	writeTestFile(t, path, "current")
	original := "\xef\xbb\xbfraw\x00bytes\r\nwith\nmixed\r\n"
	writeTestFile(t, path+".bak", original)
	if code, out := runCLI(t, "-restore", "-write", "-q", "-allow-secrets", path); code != exitOK || out != "" {
		t.Fatalf("quiet raw restore exit=%d: %s", code, out)
	}
	if readBack(t, path) != original || readBack(t, path+".restore.bak") != "current" {
		t.Fatal("raw restore changed bytes or lost current contents")
	}
}

func TestRestoreRejectsSymlinks(t *testing.T) {
	for _, sourceLink := range []bool{false, true} {
		t.Run(fmt.Sprint(sourceLink), func(t *testing.T) {
			f := fixture(t, "a.go", "package p\n")
			target := filepath.Join(filepath.Dir(f.Path), "target")
			writeTestFile(t, target, "keep me")
			link := f.Path + ".bak"
			if sourceLink {
				link = f.Path
				if err := os.Remove(link); err != nil {
					t.Fatal(err)
				}
				writeTestFile(t, f.Path+".bak", "original")
			}
			if err := os.Symlink(target, link); err != nil {
				t.Skipf("symlink unavailable: %v", err)
			}
			if code, _ := runCLI(t, "-restore", "-w", f.Path); code != exitError {
				t.Fatal("restore accepted a symlink")
			}
			if readBack(t, target) != "keep me" {
				t.Fatal("symlink target changed")
			}
			assertAbsent(t, f.Path+".restore.bak")
		})
	}
}

func writeTestFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertAbsent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("%s should not exist: %v", path, err)
	}
}
