package main

import (
	"path/filepath"
	"testing"
)

func TestStdoutDetectsExtensionlessScripts(t *testing.T) {
	const script = "#!/usr/bin/env python3\n# remove this\nprint('ok')\n"
	dir := srcDir(t, map[string]string{"helper": script, "unknown": "x = 1\n"})
	path := filepath.Join(dir, "helper")
	if code, out := runCLI(t, "-stdout", path); code != exitOK || out != "#!/usr/bin/env python3\nprint('ok')\n" {
		t.Fatalf("detected preview: code %d, output %q", code, out)
	}
	if got := readBack(t, path); got != script {
		t.Fatalf("preview changed the original: %q", got)
	}
	if code, _ := runCLI(t, "-stdout", filepath.Join(dir, "unknown")); code != exitError {
		t.Fatalf("ambiguous content: exit %d, want %d", code, exitError)
	}
}

func TestStdoutGuardsHeuristicDetection(t *testing.T) {
	const input = "#!/usr/bin/env perl\n# note\nprint 1;\n"
	file := filepath.Join(srcDir(t, map[string]string{"tool": input}), "tool")
	if code, _ := runCLI(t, "-stdout", file); code != exitError {
		t.Fatalf("automatic heuristic lexer: exit %d, want %d", code, exitError)
	}
	if code, out := runCLI(t, "-lang", "perl", "-stdout", file); code != exitOK || out != "#!/usr/bin/env perl\nprint 1;\n" {
		t.Fatalf("explicit language: exit %d, output %q", code, out)
	}
}

func TestWebRejectsPathsAndOtherModes(t *testing.T) {
	for _, args := range [][]string{{"-web", "."}, {"-web", "-w"}, {"-web", "-restore"}, {"-web", "-force"}, {"-web", "-terms", "rules.tsv"}} {
		if code, _ := runCLI(t, args...); code != exitUsage {
			t.Errorf("run(%v) = %d, want usage error", args, code)
		}
	}
}
