package glossary

import (
	"bytes"
	"strings"
)

// Apply finds every term from t in src and returns the rewritten buffer along
// with the hits. The returned buffer is a fresh slice; src is never modified,
// so a caller doing a dry run can discard it and keep only the hits.
func (t *Table) Apply(src []byte, opt Options) (out []byte, hits []Hit) {
	if len(t.rules) == 0 {
		return src, nil
	}
	var buf bytes.Buffer
	buf.Grow(len(src))

	line := 1
	last := 0 // start of the pending verbatim stretch
	for i := 0; i < len(src); {
		if src[i] == '\n' {
			line++
			i++
			continue
		}
		inComment := inRanges(opt.Comments, i)
		if opt.CommentsOnly && !inComment {
			i++
			continue
		}
		r := t.matchAt(src, i)
		if r == nil {
			i++
			continue
		}
		if skipLine(src, i) || inURL(src, i) {
			i += len(r.From)
			continue
		}
		matched := string(src[i : i+len(r.From)])
		to := r.To
		if r.CI {
			to = carryCase(matched, to)
		}
		if !inComment {
			// "pentest := 1" must not become "threat detection := 1", which
			// no language parses. Casing was carried above, so PENTEST ends
			// up THREAT_DETECTION and Go's exported Pentest stays exported.
			to = toIdent(to)
		}
		buf.Write(src[last:i])
		buf.WriteString(to)
		hits = append(hits, Hit{
			Start: i, End: i + len(r.From), Line: line,
			From: matched, To: to, InComment: inComment,
		})
		i += len(r.From)
		last = i
	}
	if len(hits) == 0 {
		return src, nil
	}
	buf.Write(src[last:])
	return buf.Bytes(), hits
}

// toIdent turns a replacement into something that can stand where an
// identifier stood: runs of spaces and hyphens become one underscore. Letters,
// digits and CJK are already legal identifier characters in every language the
// tool handles, so they pass through.
//
// This also applies to string literals in code, which the glossary cannot tell
// apart from identifiers: "pentest" becomes "threat_detection". The program
// still builds; the string just reads less naturally.
func toIdent(s string) string {
	if !strings.ContainsAny(s, " -\t") {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	sep := false
	for _, r := range s {
		if r == ' ' || r == '-' || r == '\t' {
			sep = true
			continue
		}
		if sep && b.Len() > 0 {
			b.WriteByte('_')
		}
		sep = false
		b.WriteRune(r)
	}
	return b.String()
}

// inRanges reports whether i falls in one of the ranges.
func inRanges(rs []Region, i int) bool {
	for _, r := range rs {
		if i >= r.Start && i < r.End {
			return true
		}
	}
	return false
}
