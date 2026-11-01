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

// TestListTermsRoundTripFlagless covers the rule the built-in table happens
// not to contain: an ASCII term with neither flag. An empty flag column is
// read back with the inferred word,ci, so -list-terms has to print
// noword,nocase for it, or the table changes on every round trip.
func TestListTermsRoundTripFlagless(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.tsv")
	body := "foo\tbar\tnoword,nocase\n渗透\t威胁检测\n"
	if err := os.WriteFile(src, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	_, first := runCLI(t, "-list-terms", "-terms", src)
	if !strings.Contains(first, "foo\tbar\tnoword,nocase") {
		t.Errorf("flagless ASCII rule printed without its flags:\n%s", first)
	}

	again := filepath.Join(dir, "again.tsv")
	if err := os.WriteFile(again, []byte(first), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, second := runCLI(t, "-list-terms", "-terms", again); second != first {
		t.Errorf("round trip changed the table:\n got:\n%s\n want:\n%s", second, first)
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

// TestStdoutRefusesSecrets covers the one path that does not go through walk.
// Printing a key file is not a write, but it still puts the contents
// somewhere the user did not ask for, so it needs the same guard.
func TestStdoutRefusesSecrets(t *testing.T) {
	dir := srcDir(t, map[string]string{
		"secrets.yml": "db:\n  pass: hunter2 # c\n",
		"app.yml":     "x: 1 # c\n",
	})

	code, out := runCLI(t, "-stdout", filepath.Join(dir, "secrets.yml"))
	if code == exitOK {
		t.Errorf("exit = %d, want non-zero for a credential file", code)
	}
	if strings.Contains(out, "hunter2") {
		t.Errorf("the credential reached stdout:\n%s", out)
	}

	// An ordinary file still works, and -allow-secrets lifts the guard.
	if code, out := runCLI(t, "-stdout", filepath.Join(dir, "app.yml")); code != exitOK {
		t.Errorf("ordinary file: exit = %d, out %q", code, out)
	}
	code, out = runCLI(t, "-stdout", "-allow-secrets", filepath.Join(dir, "secrets.yml"))
	if code != exitOK {
		t.Errorf("-allow-secrets: exit = %d", code)
	}
	if !strings.Contains(out, "hunter2") {
		t.Errorf("-allow-secrets did not print the file:\n%s", out)
	}
}

// TestVerifyBlocksBrokenWrite drives the check end to end. A term table whose
// replacement drops a stray bracket into code would leave a file that no
// longer parses, so the write must be refused, the file left byte-identical,
// and the reason shown. -no-verify lets it through, the escape hatch for a
// user who knows better.
func TestVerifyBlocksBrokenWrite(t *testing.T) {
	const body = "package x\n\nfunc F() int { return HOLE }\n"
	dir := srcDir(t, map[string]string{"a.go": body})
	table := filepath.Join(t.TempDir(), "t.tsv")
	if err := os.WriteFile(table, []byte("HOLE\t1)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "a.go")

	code, out := runCLI(t, "-replace", "-keep-comments", "-terms", table, "-w", dir)
	if code != exitOK {
		t.Fatalf("exit = %d:\n%s", code, out)
	}
	if !strings.Contains(out, "verification failed") {
		t.Errorf("report does not name the failed check:\n%s", out)
	}
	if got, _ := os.ReadFile(path); string(got) != body {
		t.Errorf("broken rewrite was written:\n%s", got)
	}

	runCLI(t, "-replace", "-keep-comments", "-terms", table, "-no-verify", "-w", dir)
	if got, _ := os.ReadFile(path); string(got) == body {
		t.Error("-no-verify did not let the write through")
	}
}

// TestCheckExitCode is the CI contract: non-zero when anything would change,
// zero on a clean tree, and the files are never touched either way.
func TestCheckExitCode(t *testing.T) {
	const dirty = "package x\n\nvar n = 1 // c\n"
	dir := srcDir(t, map[string]string{"a.go": dirty})

	if code, _ := runCLI(t, "-check", "-q", dir); code != exitError {
		t.Errorf("dirty tree: exit = %d, want %d", code, exitError)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "a.go")); string(got) != dirty {
		t.Errorf("-check wrote the file:\n%s", got)
	}

	clean := srcDir(t, map[string]string{"a.go": "package x\n\nvar n = 1\n"})
	if code, _ := runCLI(t, "-check", "-q", clean); code != exitOK {
		t.Errorf("clean tree: exit = %d, want %d", code, exitOK)
	}
}

func TestDiffPrintsAndDoesNotWrite(t *testing.T) {
	const body = "package x\n\n// note\nvar n = 1\n"
	dir := srcDir(t, map[string]string{"a.go": body})

	code, out := runCLI(t, "-diff", dir)
	if code != exitOK {
		t.Fatalf("exit = %d:\n%s", code, out)
	}
	for _, want := range []string{"@@ -", "-// note", " var n = 1"} {
		if !strings.Contains(out, want) {
			t.Errorf("diff output lacks %q:\n%s", want, out)
		}
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "a.go")); string(got) != body {
		t.Errorf("-diff wrote the file:\n%s", got)
	}
}

func TestReportOnlyFlagsRejectWrite(t *testing.T) {
	dir := srcDir(t, map[string]string{"a.go": "package x\n"})
	for _, flag := range []string{"-check", "-diff"} {
		if code, _ := runCLI(t, flag, "-w", dir); code != exitUsage {
			t.Errorf("%s -w: exit = %d, want %d", flag, code, exitUsage)
		}
	}
}
