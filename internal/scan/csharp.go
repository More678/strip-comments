package scan

import "strings"

// csharpStringAt keeps nested interpolation literals separate from expression comments.
func (s *scanner) csharpStringAt(i, depth int) (int, string, bool) {
	src := s.src
	j, dollars, verbatim := i, 0, false
	if src[j] == '@' {
		verbatim = true
		j++
	}
	for j < len(src) && src[j] == '$' {
		dollars++
		j++
	}
	if !verbatim && j < len(src) && src[j] == '@' {
		verbatim = true
		j++
	}
	if j >= len(src) || src[j] != '"' {
		return 0, "", false
	}
	if depth >= 128 {
		return 0, "C# string nesting limit at line " + itoa(lineOf(src, i)), true
	}
	quotes := 0
	for j+quotes < len(src) && src[j+quotes] == '"' {
		quotes++
	}
	raw := !verbatim && quotes >= 3
	if !raw {
		quotes = 1
		if dollars > 1 {
			return 0, "invalid C# string prefix at line " + itoa(lineOf(src, i)), true
		}
	}
	end, reason := s.csharpStringBody(j+quotes, quotes, dollars, verbatim, raw, depth+1)
	return end, reason, true
}

func (s *scanner) csharpStringBody(j, quotes, dollars int, verbatim, raw bool, depth int) (int, string) {
	src := s.src
	for j < len(src) {
		c := src[j]
		if !raw && !verbatim && c == '\\' && j+1 < len(src) {
			j += 2
			continue
		}
		if c == '"' {
			if verbatim && j+1 < len(src) && src[j+1] == '"' {
				j += 2
				continue
			}
			n := 1
			if raw {
				for j+n < len(src) && src[j+n] == '"' {
					n++
				}
			}
			if n >= quotes {
				return j + quotes, ""
			}
			j += n
			continue
		}
		if dollars > 0 && c == '{' {
			n := 1
			for j+n < len(src) && src[j+n] == '{' {
				n++
			}
			j += n
			if (!raw && n%2 == 1) || (raw && n >= dollars) {
				end, reason := s.csharpExpression(j, dollars, quotes, depth, verbatim, raw)
				if reason != "" {
					return 0, reason
				}
				j = end
			}
			continue
		}
		if c == '\n' && !raw && !verbatim {
			return 0, "unterminated C# string at line " + itoa(lineOf(src, j))
		}
		j++
	}
	return 0, "unterminated C# string"
}

func (s *scanner) csharpExpression(j, braces, quotes, depth int, verbatim, raw bool) (int, string) {
	src := s.src
	var stack []byte
	for j < len(src) {
		if end, reason, ok := s.csharpStringAt(j, depth); ok {
			if reason != "" {
				return 0, reason
			}
			j = end
			continue
		}
		if src[j] == '\'' {
			if end, ok := charAt(src, j); ok {
				j = end
				continue
			}
		}
		if start, end, doc, reason, ok := s.commentAt(j); ok {
			if reason != "" {
				return 0, reason
			}
			s.spans = append(s.spans, Span{Start: start, End: end, Doc: doc})
			j = end
			continue
		}
		c := src[j]
		switch c {
		case '(', '[', '{':
			stack = append(stack, c)
		case ')', ']', '}':
			if len(stack) == 0 {
				if c == '}' && hasPrefixAt(src, j, strings.Repeat("}", braces), false) {
					return j + braces, ""
				}
				return 0, "unbalanced C# interpolation at line " + itoa(lineOf(src, j))
			}
			opener := stack[len(stack)-1]
			if !(opener == '(' && c == ')' || opener == '[' && c == ']' || opener == '{' && c == '}') {
				return 0, "unbalanced C# interpolation at line " + itoa(lineOf(src, j))
			}
			stack = stack[:len(stack)-1]
		case ':':
			if j+1 < len(src) && src[j+1] == ':' {
				j += 2 // global:: and other alias-qualified names are not format clauses.
				continue
			}
			if len(stack) == 0 {
				return s.csharpFormat(j+1, braces, quotes, verbatim, raw)
			}
		}
		j++
	}
	return 0, "unterminated C# interpolation"
}

func (s *scanner) csharpFormat(j, braces, quotes int, verbatim, raw bool) (int, string) {
	src := s.src
	closer := strings.Repeat("}", braces)
	quoteEnd := strings.Repeat(`"`, quotes)
	for j < len(src) {
		if !raw && !verbatim && src[j] == '\\' && j+1 < len(src) {
			j += 2
			continue
		}
		if verbatim && hasPrefixAt(src, j, `""`, false) {
			j += 2
			continue
		}
		if hasPrefixAt(src, j, quoteEnd, false) {
			break // A format clause cannot cross its enclosing string delimiter.
		}
		if hasPrefixAt(src, j, closer, false) {
			return j + braces, ""
		}
		if src[j] == '\n' && !raw && !verbatim {
			break
		}
		j++
	}
	return 0, "unterminated C# interpolation format"
}
