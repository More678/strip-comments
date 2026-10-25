package scan

import (
	"bytes"

	"github.com/More678/strip-comments/internal/lang"
)

// filter removes the spans that have to survive from the candidate list and
// returns the spans that may be deleted.
//
// Protected are comments matching l.KeepRE(), a `#!` shebang on the first
// line, documentation comments under KeepDoc, the leading file comment under
// KeepLicense and the cgo preamble sitting right above `import "C"`.
func filter(src []byte, spans []Span, l *lang.Language, opt Options) []Span {
	if len(spans) == 0 {
		return spans
	}
	keep := make([]bool, len(spans))
	re := l.KeepRE()
	for i, s := range spans {
		text := src[s.Start:s.End]
		switch {
		case re != nil && re.FindIndex(text) != nil:
			keep[i] = true
		case opt.KeepDoc && s.Doc:
			keep[i] = true
		case s.Start == 0 && hasShebang(src):
			keep[i] = true
		}
	}
	if opt.KeepLicense {
		for _, i := range licenseRun(src, spans) {
			keep[i] = true
		}
	}
	if l.Name == "go" {
		keepCgoPreamble(src, spans, keep)
	}

	kept := 0
	for _, k := range keep {
		if k {
			kept++
		}
	}
	if kept == 0 {
		return spans
	}
	out := make([]Span, 0, len(spans)-kept)
	for i, s := range spans {
		if !keep[i] {
			out = append(out, s)
		}
	}
	return out
}

// licenseRun returns the indexes of the comments that KeepLicense protects:
// the first span that is not the shebang itself and has nothing but
// whitespace and an optional shebang line in front of it, plus every span
// after it that is separated from its predecessor by whitespace without a
// blank line. It returns nil when the file does not open with a comment.
func licenseRun(src []byte, spans []Span) []int {
	first := -1
	for i, s := range spans {
		if s.Start == 0 && hasShebang(src) {
			continue // kept by the shebang rule already
		}
		if headerOnly(src[:s.Start]) {
			first = i
		}
		break
	}
	if first < 0 {
		return nil
	}
	run := []int{first}
	for i := first + 1; i < len(spans); i++ {
		if !adjacent(src[spans[i-1].End:spans[i].Start]) {
			break
		}
		run = append(run, i)
	}
	return run
}

// adjacent reports whether gap is whitespace holding at most one newline,
// i.e. the two comments around it sit on consecutive lines (or one line).
func adjacent(gap []byte) bool {
	newlines := 0
	for _, c := range gap {
		switch c {
		case '\n':
			if newlines++; newlines > 1 {
				return false
			}
		case ' ', '\t', '\r', '\f', '\v':
		default:
			return false
		}
	}
	return true
}

// headerOnly reports whether prefix holds nothing but whitespace, optionally
// preceded by a shebang line.
func headerOnly(prefix []byte) bool {
	if hasShebang(prefix) {
		if i := bytes.IndexByte(prefix, '\n'); i >= 0 {
			prefix = prefix[i+1:]
		} else {
			prefix = nil
		}
	}
	return len(bytes.TrimSpace(prefix)) == 0
}

// keepCgoPreamble protects the cgo preamble: the run of comments directly
// above an `import "C"` clause, separated from it by whitespace only.
func keepCgoPreamble(src []byte, spans []Span, keep []bool) {
	for i, s := range spans {
		if !cgoImportAt(src, s.End) {
			continue
		}
		for j := i; j >= 0; j-- {
			keep[j] = true
			if j == 0 || !onlySpaceBetween(src, spans[j-1].End, spans[j].Start) {
				break
			}
		}
	}
}

// cgoImportAt reports whether `import "C"` starts at i, ignoring whitespace.
func cgoImportAt(src []byte, i int) bool {
	j := firstNonSpace(src, i)
	if !bytes.HasPrefix(src[j:], []byte("import")) {
		return false
	}
	j = firstNonSpace(src, j+len("import"))
	return bytes.HasPrefix(src[j:], []byte(`"C"`))
}

func hasShebang(b []byte) bool {
	return len(b) >= 2 && b[0] == '#' && b[1] == '!'
}

// firstNonSpace returns the index of the first byte at or after i that is not
// whitespace, or len(src).
func firstNonSpace(src []byte, i int) int {
	for ; i < len(src); i++ {
		switch src[i] {
		case ' ', '\t', '\r', '\n', '\v', '\f':
		default:
			return i
		}
	}
	return len(src)
}

// onlySpaceBetween reports whether src[a:b] is whitespace only.
func onlySpaceBetween(src []byte, a, b int) bool {
	if a < 0 || b > len(src) || a > b {
		return false
	}
	return firstNonSpace(src, a) >= b
}
