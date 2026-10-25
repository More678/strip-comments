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
func rebuild(src []byte, spans []Span, keepBlank bool) (out []byte, removedLines int) {
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

	// keepText copies text that stays, breaking it into lines.
	keepText := func(b []byte) {
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
		for {
			touched = true
			i := bytes.IndexByte(b, '\n')
			if i < 0 {
				return
			}
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
