package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runCLI calls run with args and returns its exit code plus whatever it wrote
// to stdout. stderr is left alone so a failing test still shows the reason.
func runCLI(t *testing.T, args ...string) (int, string) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	saved := os.Stdout
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()

	code := run(args)
	w.Close()
	os.Stdout = saved
	return code, <-done
}

// srcDir writes name/body pairs into a fresh temp dir and returns it.
func srcDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// TestReplaceOffByDefault is the safety property: without -replace not one
// term may change, however many the file contains.
func TestReplaceOffByDefault(t *testing.T) {
	const body = "package x\n\nvar pentest = 1\nvar 渗透 = 2\n"
	dir := srcDir(t, map[string]string{"a.go": body})

	code, out := runCLI(t, "-w", dir)
	if code != exitOK {
		t.Fatalf("exit = %d, out:\n%s", code, out)
	}
	if strings.Contains(out, "terms") {
		t.Errorf("report mentions terms without -replace:\n%s", out)
	}
	got, err := os.ReadFile(filepath.Join(dir, "a.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != body {
		t.Errorf("file changed without -replace:\n got  %q\n want %q", got, body)
	}
}

// TestTermsFlagsNeedReplace covers the guard: a term table the run would
// silently ignore is a usage error, not a quiet no-op.
func TestTermsFlagsNeedReplace(t *testing.T) {
	dir := srcDir(t, map[string]string{"a.go": "package x\n"})
	for _, flag := range []string{
		"-terms", "-terms-add",
	} {
		t.Run(flag, func(t *testing.T) {
			if code, _ := runCLI(t, flag, "x.tsv", dir); code != exitUsage {
				t.Errorf("exit = %d, want %d", code, exitUsage)
			}
		})
	}
	for _, flag := range []string{"-terms-in-comments", "-keep-comments"} {
		t.Run(flag, func(t *testing.T) {
			if code, _ := runCLI(t, flag, dir); code != exitUsage {
				t.Errorf("exit = %d, want %d", code, exitUsage)
			}
		})
	}
}

// TestMissingTermsFile must fail loudly rather than fall back to the built-in
// table, which would replace terms the user never asked for.
func TestMissingTermsFile(t *testing.T) {
	dir := srcDir(t, map[string]string{"a.go": "package x\n"})
	missing := filepath.Join(dir, "nope.tsv")
	code, _ := runCLI(t, "-replace", "-terms", missing, dir)
	if code != exitUsage {
		t.Errorf("exit = %d, want %d for a missing term file", code, exitUsage)
	}
}

// TestListTermsRoundTrip: -list-terms prints the table in the TSV shape that
// -terms reads, so the output can be edited and fed straight back in.
func TestListTermsRoundTrip(t *testing.T) {
	_, out := runCLI(t, "-list-terms")
	if out == "" {
		t.Fatal("-list-terms printed nothing")
	}
	dir := t.TempDir()
	table := filepath.Join(dir, "terms.tsv")
	if err := os.WriteFile(table, []byte(out), 0o644); err != nil {
		t.Fatal(err)
	}
	_, again := runCLI(t, "-list-terms", "-terms", table)
	if again != out {
		t.Errorf("round trip changed the table:\n got:\n%s\n want:\n%s", again, out)
	}
}

// TestReplaceIdempotent: a second run must find nothing. If a replacement
// could itself be matched by another rule, the file would keep drifting on
// every run, which is the worst failure mode for a tool invoked from CI.
func TestReplaceIdempotent(t *testing.T) {
	dir := srcDir(t, map[string]string{
		"a.go": "package x\n\n// 渗透测试与漏洞利用\nvar pentest = 1 // 提权\n",
		"b.py": "# 渗透\nx = 1  # 后门\n",
		"c.c":  "/* 暴力破解 */\nint exploit = 1; // 绕过\n",
	})

	code, first := runCLI(t, "-replace", "-keep-comments", "-w", dir)
	if code != exitOK {
		t.Fatalf("first run exit = %d:\n%s", code, first)
	}
	if !strings.Contains(first, "terms replaced") {
		t.Fatalf("first run replaced nothing:\n%s", first)
	}

	// Snapshot every file, then run again.
	before := map[string]string{}
	for _, name := range []string{"a.go", "b.py", "c.c"} {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		before[name] = string(b)
	}

	code, second := runCLI(t, "-replace", "-keep-comments", "-w", dir)
	if code != exitOK {
		t.Fatalf("second run exit = %d:\n%s", code, second)
	}
	if !strings.Contains(second, "0 terms replaced") {
		t.Errorf("second run was not a no-op:\n%s", second)
	}
	for name, want := range before {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != want {
			t.Errorf("%s drifted on the second run:\n got  %q\n want %q",
				name, b, want)
		}
	}
}
