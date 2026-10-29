package udiff

import (
	"fmt"
	"strings"
)

// op is one line of the edit script.
type op struct {
	kind byte // ' ' keep, '-' delete, '+' insert
	text string
	ai   int // 1-based line number in a, 0 for an insert
	bi   int // 1-based line number in b, 0 for a delete
}

// script returns the edit script turning a into b, via a line-level LCS.
func script(a, b []string) []op {
	// n[i][j] is the LCS length of a[i:] and b[j:].
	n := make([][]int, len(a)+1)
	for i := range n {
		n[i] = make([]int, len(b)+1)
	}
	for i := len(a) - 1; i >= 0; i-- {
		for j := len(b) - 1; j >= 0; j-- {
			if a[i] == b[j] {
				n[i][j] = n[i+1][j+1] + 1
				continue
			}
			if n[i+1][j] >= n[i][j+1] {
				n[i][j] = n[i+1][j]
			} else {
				n[i][j] = n[i][j+1]
			}
		}
	}

	var out []op
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] == b[j]:
			out = append(out, op{' ', a[i], i + 1, j + 1})
			i++
			j++
		case n[i+1][j] >= n[i][j+1]:
			out = append(out, op{'-', a[i], i + 1, 0})
			i++
		default:
			out = append(out, op{'+', b[j], 0, j + 1})
			j++
		}
	}
	for ; i < len(a); i++ {
		out = append(out, op{'-', a[i], i + 1, 0})
	}
	for ; j < len(b); j++ {
		out = append(out, op{'+', b[j], 0, j + 1})
	}
	return out
}

// hunks groups the edit script into unified-diff hunks, each carrying up to
// ctx unchanged lines on either side of the changes.
func hunks(a, b []string) []string {
	ops := script(a, b)

	// interesting reports whether ops[i] is an insert or a delete.
	interesting := func(i int) bool { return ops[i].kind != ' ' }

	var out []string
	for i := 0; i < len(ops); {
		if !interesting(i) {
			i++
			continue
		}
		// Walk back for leading context, then forward to the end of the run,
		// absorbing gaps of at most 2*ctx keeps so neighbouring edits share
		// one hunk instead of producing two with overlapping context.
		start := max(0, i-ctx)
		end := i
		for end < len(ops) {
			if interesting(end) {
				end++
				continue
			}
			gap := end
			for gap < len(ops) && !interesting(gap) {
				gap++
			}
			if gap < len(ops) && gap-end <= 2*ctx {
				end = gap
				continue
			}
			break
		}
		end = min(len(ops), end+ctx)
		out = append(out, render(ops[start:end]))
		i = end
	}
	return out
}

// render writes one hunk, including its @@ header.
func render(ops []op) string {
	var aStart, bStart, aLen, bLen int
	for _, o := range ops {
		if o.ai > 0 {
			if aStart == 0 {
				aStart = o.ai
			}
			aLen++
		}
		if o.bi > 0 {
			if bStart == 0 {
				bStart = o.bi
			}
			bLen++
		}
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "@@ -%d,%d +%d,%d @@\n", aStart, aLen, bStart, bLen)
	for _, o := range ops {
		sb.WriteByte(o.kind)
		sb.WriteString(o.text)
		sb.WriteByte('\n')
	}
	return sb.String()
}
