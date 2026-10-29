package scan

import "bytes"

// rebuild deletes the bytes covered by spans and tidies up the lines that
// were touched. See the rules documented on Strip.
//
// Newlines inside a span are kept, so a block comment spanning three lines
// still yields three lines. Lines that end up whitespace-only because a
// comment was removed disappear together with their newline, unless keepBlank
// asks for an empty line instead. Lines that keep code lose their trailing
// blanks. Lines without a removed comment are copied byte for byte, so
// original blank lines and a missing final newline survive untouched.
func rebuild(src []byte, spans []Span, keepBlank, textContent bool) (out []byte, removedLines int) {
	if len(spans) == 0 {
		return src, 0
	}
	out = make([]byte, 0, len(src))
	line := make([]byte, 0, 128)
	touched := false

	// flush emits the line that has been collected so far; nl reports whether
	// it was terminated by a newline in the source.
	flush := func(nl bool) {
		switch {
		case touched && lineIsBlank(line):
			if keepBlank {
				if nl {
					out = append(out, '\n')
				}
			} else {
				removedLines++
			}
		default:
			if touched {
				line = trimLineEnd(line)
			}
			out = append(out, line...)
			if nl {
				out = append(out, '\n')
			}
		}
		line = line[:0]
		touched = false
	}

	// sep is set after an inline comment is dropped. A comment separates
	// tokens, so `int/**/x` must not become `intx` nor `1 -/**/-1` become
	// `1 --1`: the next kept byte gets a space if it would glue on.
	sep := false

	// keepText copies text that stays, breaking it into lines.
	keepText := func(b []byte) {
		if sep && !textContent && len(b) > 0 && len(line) > 0 &&
			glues(line[len(line)-1], b[0]) {
			line = append(line, ' ')
		}
		if len(b) > 0 {
			sep = false
		}
		for {
			i := bytes.IndexByte(b, '\n')
			if i < 0 {
				line = append(line, b...)
				return
			}
			line = append(line, b[:i]...)
			flush(true)
			b = b[i+1:]
		}
	}

	// dropText discards comment bytes but preserves the newlines inside them.
	dropText := func(b []byte) {
		if len(b) == 0 {
			return
		}
		sep = true
		for {
			touched = true
			i := bytes.IndexByte(b, '\n')
			if i < 0 {
				return
			}
			// A kept newline already separates whatever comes next.
			sep = false
			flush(true)
			b = b[i+1:]
		}
	}

	pos := 0
	for _, s := range spans {
		if s.Start > len(src) {
			break
		}
		if s.End > len(src) {
			s.End = len(src)
		}
		if s.Start < pos { // tolerate unsorted or overlapping input
			if s.End <= pos {
				continue
			}
			s.Start = pos
		}
		keepText(src[pos:s.Start])
		dropText(src[s.Start:s.End])
		pos = s.End
	}
	keepText(src[pos:])
	if len(line) > 0 || touched {
		flush(false)
	}
	return out, removedLines
}

// lineIsBlank reports whether b holds whitespace only.
// glues reports whether a and b would read as one token once the comment
// between them is gone: two identifier bytes (`int` + `x`) or two operator
// bytes (`-` + `-`, `/` + `/`). Brackets, quotes and separators never merge,
// so `f(/* x */1)` needs no space.
func glues(a, b byte) bool {
	return (isWord(a) && isWord(b)) || (isOp(a) && isOp(b))
}

func isWord(c byte) bool {
	return c == '_' || c == '$' || c >= 0x80 ||
		(c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isOp(c byte) bool {
	switch c {
	case '+', '-', '*', '/', '%', '=', '<', '>', '!', '&', '|', '^', '~',
		'.', ':', '?', '#', '@':
		return true
	}
	return false
}

func lineIsBlank(b []byte) bool {
	for _, c := range b {
		switch c {
		case ' ', '\t', '\r', '\v', '\f':
		default:
			return false
		}
	}
	return true
}

// trimLineEnd drops trailing spaces and tabs from a line that keeps code.
func trimLineEnd(b []byte) []byte {
	i := len(b)
	for i > 0 && (b[i-1] == ' ' || b[i-1] == '\t') {
		i--
	}
	return b[:i]
}
