// Package udiff renders a unified diff between two versions of a file.
// The standard library has no diff, and the shapes this tool produces are
// small and mostly deletions, so a line-level LCS is enough.
package udiff

import (
	"fmt"
	"strings"
)

// maxLCS caps the quadratic table. Above it, Unified falls back to a summary
// rather than spending seconds on a file nobody will read line by line.
const maxLCS = 5000

// ctx is how many unchanged lines surround each hunk.
const ctx = 3

// Unified returns a unified diff of before and after, labelled with name.
// It returns "" when the two are identical.
func Unified(name string, before, after []byte) string {
	a, b := splitLines(before), splitLines(after)
	if len(a) == len(b) && string(before) == string(after) {
		return ""
	}
	if len(a) > maxLCS || len(b) > maxLCS {
		return fmt.Sprintf("--- %s\n+++ %s\n@@ %d lines -> %d lines @@\n",
			name, name, len(a), len(b))
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "--- %s\n+++ %s\n", name, name)
	for _, h := range hunks(a, b) {
		sb.WriteString(h)
	}
	return sb.String()
}

// splitLines keeps the content of each line without its newline. A trailing
// newline does not produce a final empty line.
func splitLines(b []byte) []string {
	if len(b) == 0 {
		return nil
	}
	s := string(b)
	s = strings.TrimSuffix(s, "\n")
	return strings.Split(s, "\n")
}
