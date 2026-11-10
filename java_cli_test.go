package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestJavaCommentRemovalCLI(t *testing.T) {
	const source = `/** documentation marker */
public class Example {
    // line marker
    static String url = "https://example.invalid/a//b";
    static String block = "/* literal marker */";
    static char slash = '/';
    static String text = """
        // text block marker
        /* text block marker */
        """;
    static int/**/value = 1; /* ordinary marker */
}
`
	for _, keepDoc := range []bool{false, true} {
		name := "remove-doc"
		if keepDoc {
			name = "keep-doc"
		}
		t.Run(name, func(t *testing.T) {
			dir := srcDir(t, map[string]string{"Example.java": source})
			args := []string{"-w"}
			if keepDoc {
				args = append(args, "-keep-doc")
			}
			code, report := runCLI(t, append(args, dir)...)
			if code != exitOK {
				t.Fatalf("exit %d: %s", code, report)
			}
			path := filepath.Join(dir, "Example.java")
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			out := string(raw)
			for _, marker := range []string{"line marker", "ordinary marker", "int/**/value"} {
				if strings.Contains(out, marker) {
					t.Errorf("comment survived: %s", marker)
				}
			}
			for _, literal := range []string{`"https://example.invalid/a//b"`, `"/* literal marker */"`, "// text block marker", "/* text block marker */", "int value"} {
				if !strings.Contains(out, literal) {
					t.Errorf("literal or token separator lost: %s", literal)
				}
			}
			if strings.Contains(out, "documentation marker") != keepDoc {
				t.Errorf("Javadoc preservation differs from KeepDoc=%v", keepDoc)
			}
			code, report = runCLI(t, append(args, dir)...)
			again, err := os.ReadFile(path)
			if err != nil || code != exitOK || string(again) != out {
				t.Fatalf("second pass changed output: %v; %s", err, report)
			}
			t.Run("javac17", func(t *testing.T) {
				javac, err := exec.LookPath("javac")
				if err != nil {
					t.Skip("javac unavailable; output assertions already ran")
				}
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, javac, "-proc:none", "-encoding", "UTF-8", "--release", "17", "-d", dir, path)
				output, err := cmd.CombinedOutput()
				if strings.Contains(string(output), "release version 17 not supported") || strings.Contains(string(output), "invalid flag: --release") {
					t.Skip("Java 17 compiler required for text blocks")
				}
				if err != nil {
					t.Fatalf("javac: %v\n%s", err, output)
				}
			})
		})
	}
}
