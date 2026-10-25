package scan

import (
	"bytes"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/More678/Remove-All-Commet/internal/lang"
)

// charAt matches a strict char literal at i: `'` + one UTF-8 codepoint (not
// `'`, `\`, newline) or one escape sequence + `'`. Anything else, such as a
// Rust lifetime `'a`, a C++14 digit separator `1'000` or `#error can't`,
// leaves the quote as a plain byte.
func charAt(src []byte, i int) (int, bool) {
	j := i + 1
	if j >= len(src) {
		return 0, false
	}
	switch c := src[j]; {
	case c == '\'' || c == '\n':
		return 0, false
	case c == '\\':
		j = escapeEnd(src, j)
		if j < 0 {
			return 0, false
		}
	default:
		r, n := utf8.DecodeRune(src[j:])
		if r == utf8.RuneError && n <= 1 {
			n = 1 // undecodable byte (GBK etc.): treat as one unit
		}
		j += n
	}
	if j < len(src) && src[j] == '\'' {
		return j + 1, true
	}
	return 0, false
}

// escapeEnd returns the offset just past the escape sequence starting with the
// backslash at j, or -1 when it is malformed.
func escapeEnd(src []byte, j int) int {
	j++
	if j >= len(src) || src[j] == '\n' {
		return -1
	}
	switch c := src[j]; {
	case c == 'x':
		k := j + 1
		for k < len(src) && isHex(src[k]) {
			k++
		}
		if k == j+1 {
			return -1
		}
		return k
	case c == 'u' && j+1 < len(src) && src[j+1] == '{':
		hi := min(j+2+9, len(src)) // at most 8 hex digits plus the brace
		k := bytes.IndexByte(src[j+2:hi], '}')
		if k < 1 || k > 8 {
			return -1
		}
		for _, h := range src[j+2 : j+2+k] {
			if !isHex(h) && h != '_' {
				return -1
			}
		}
		return j + 3 + k
	case c == 'u' || c == 'U':
		n := 4
		if c == 'U' {
			n = 8
		}
		if j+1+n > len(src) {
			return -1
		}
		for _, h := range src[j+1 : j+1+n] {
			if !isHex(h) {
				return -1
			}
		}
		return j + 1 + n
	case c >= '0' && c <= '7':
		k := j + 1
		for k < len(src) && k < j+3 && src[k] >= '0' && src[k] <= '7' {
			k++
		}
		return k
	}
	_, n := utf8.DecodeRune(src[j:])
	return j + n
}

func isHex(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

// regexAt matches a `/.../flags` literal at i. The closing slash must sit on
// the same line and outside a `[...]` class, otherwise the slash is division.
func (s *scanner) regexAt(i int) (int, bool) {
	src := s.src
	if i+1 < len(src) && (src[i+1] == '/' || src[i+1] == '*') {
		return 0, false
	}
	if !s.regexAllowed(i) {
		return 0, false
	}
	j, class := i+1, false
	for j < len(src) {
		switch src[j] {
		case '\n':
			return 0, false
		case '\\':
			j += 2
			continue
		case '[':
			class = true
		case ']':
			class = false
		case '/':
			if !class {
				j++
				for j < len(src) && src[j] >= 'a' && src[j] <= 'z' {
					j++
				}
				return j, true
			}
		}
		j++
	}
	return 0, false
}

// regexAllowed decides whether a `/` at i can open a regex literal, based on
// the previous significant byte or the keyword in front of it.
func (s *scanner) regexAllowed(i int) bool {
	src := s.src
	if s.prevSig == 0 || strings.IndexByte(lang.RegexBefore, s.prevSig) >= 0 {
		return true
	}
	j := i
	for j > s.bol && isSpace(src[j-1]) {
		j--
	}
	if j == s.bol {
		return true
	}
	// Only trust the backward scan when it agrees with prevSig; otherwise a
	// comment sits in between and the word in front of it is irrelevant.
	if !isWordByte(src[j-1]) || src[j-1] != s.prevSig {
		return false
	}
	k := j
	for k > 0 && isWordByte(src[k-1]) {
		k--
	}
	return lang.RegexKeywords[string(src[k:j])]
}

// heredocAt matches a `<<EOF` / `<<-'EOF'` / `<<~EOF` opener at i and queues
// the body, which is skipped once the opener line ends. The opener counts only
// when a later line really is the terminator.
func (s *scanner) heredocAt(i int) (int, bool) {
	src := s.src
	if !hasPrefixAt(src, i, "<<", false) {
		return 0, false
	}
	j := i + 2
	if j < len(src) && src[j] == '<' {
		return 0, false
	}
	indent := false
	if j < len(src) && (src[j] == '-' || src[j] == '~') {
		indent, j = true, j+1
	}
	var quote byte
	if j < len(src) && (src[j] == '\'' || src[j] == '"') {
		quote, j = src[j], j+1
	}
	start := j
	if j >= len(src) || !(src[j] == '_' || isAlpha(src[j])) {
		return 0, false
	}
	for j < len(src) && isWordByte(src[j]) {
		j++
	}
	id := string(src[start:j])
	if quote != 0 {
		if j >= len(src) || src[j] != quote {
			return 0, false
		}
		j++
	}
	if _, ok := s.hdIndex().find(min(s.lineEnd(i)+1, len(src)), id, indent); !ok {
		return 0, false
	}
	s.heredoc = append(s.heredoc, pendingHeredoc{id: id, indent: indent})
	return j, true
}

// heredocIndex lists, per candidate terminator text, the ascending offsets
// just past the lines that could close a heredoc. Built in one pass on the
// first opener so that `a<<EOF` on every line of a big file stays linear
// instead of rescanning to EOF each time.
type heredocIndex struct {
	exact map[string][]int // line without leading blanks, for plain `<<EOF`
	trim  map[string][]int // leading blanks trimmed, for `<<-EOF` / `<<~EOF`
}

// hdIndex returns the index for this buffer, building it on first use.
func (s *scanner) hdIndex() *heredocIndex {
	if s.hd == nil {
		s.hd = buildHeredocIndex(s.src)
	}
	return s.hd
}

func buildHeredocIndex(src []byte) *heredocIndex {
	ix := &heredocIndex{exact: map[string][]int{}, trim: map[string][]int{}}
	for p := 0; p < len(src); {
		e := p + nl2(src, p)
		past := e
		if e < len(src) {
			past++
		}
		line := bytes.TrimRight(src[p:e], " \t\r")
		if isIdentWord(line) {
			ix.exact[string(line)] = append(ix.exact[string(line)], past)
		}
		if t := bytes.TrimLeft(line, " \t"); len(t) != len(line) && isIdentWord(t) {
			ix.trim[string(t)] = append(ix.trim[string(t)], past)
		} else if len(t) == len(line) && isIdentWord(line) {
			ix.trim[string(line)] = append(ix.trim[string(line)], past)
		}
		if e >= len(src) {
			break
		}
		p = e + 1
	}
	return ix
}

// find returns the offset just past the first terminator line starting at or
// after from. from is a line start, so "line start >= from" is equivalent to
// the stored "past end > from".
func (ix *heredocIndex) find(from int, id string, indent bool) (int, bool) {
	m := ix.exact
	if indent {
		m = ix.trim
	}
	offs := m[id]
	if i := sort.SearchInts(offs, from+1); i < len(offs) {
		return offs[i], true
	}
	return 0, false
}

// isIdentWord reports whether b is a `[A-Za-z_][A-Za-z0-9_]*` identifier.
func isIdentWord(b []byte) bool {
	if len(b) == 0 || !(b[0] == '_' || isAlpha(b[0])) {
		return false
	}
	for _, c := range b {
		if !isWordByte(c) {
			return false
		}
	}
	return true
}

// yamlOpener reports whether line ends with a `|` / `>` block scalar header
// and returns that line's indentation.
func yamlOpener(line []byte) (int, bool) {
	ind := 0
	for ind < len(line) && (line[ind] == ' ' || line[ind] == '\t') {
		ind++
	}
	body := line[ind:]
	if len(body) == 0 || body[0] == '#' {
		return 0, false
	}
	for k := 0; k+1 < len(body); k++ {
		if body[k+1] == '#' && (body[k] == ' ' || body[k] == '\t') {
			body = body[:k+1]
			break
		}
	}
	body = bytes.TrimRight(body, " \t\r")
	for len(body) > 0 {
		if c := body[len(body)-1]; c == '-' || c == '+' || c >= '0' && c <= '9' {
			body = body[:len(body)-1]
			continue
		}
		break
	}
	if len(body) == 0 {
		return 0, false
	}
	if c := body[len(body)-1]; c != '|' && c != '>' {
		return 0, false
	}
	if len(body) >= 2 {
		if p := body[len(body)-2]; p != ' ' && p != '\t' {
			return 0, false
		}
	}
	return ind, true
}

// skipYAMLBlock swallows the blank or deeper-indented lines that make up a
// block scalar opened by a line indented by ind.
func skipYAMLBlock(src []byte, from, ind int) int {
	for p := from; p < len(src); {
		e := p + nl2(src, p)
		k := p
		for k < e && (src[k] == ' ' || src[k] == '\t') {
			k++
		}
		if blank := k == e || src[k] == '\r'; !blank && k-p <= ind {
			return p
		}
		if e >= len(src) {
			return len(src)
		}
		p = e + 1
	}
	return len(src)
}

// nextPHPOpen returns the offset just past the next `<?php` / `<?=`, or -1.
func nextPHPOpen(src []byte, from int) int {
	for p := from; p < len(src); {
		k := bytes.IndexByte(src[p:], '<')
		if k < 0 {
			return -1
		}
		p += k
		if hasPrefixAt(src, p, "<?php", true) {
			return p + 5
		}
		if hasPrefixAt(src, p, "<?=", false) {
			return p + 3
		}
		p++
	}
	return -1
}

func isAlpha(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }

func isWordByte(c byte) bool { return isAlpha(c) || c >= '0' && c <= '9' || c == '_' }

// rawAt tries the special lexers listed in l.Raw. ok reports a match; isCmt
// marks a Lua long comment, everything else is a literal to skip.
func (s *scanner) rawAt(i int) (end int, isCmt bool, reason string, ok bool) {
	src, r := s.src, s.r
	if i > 0 && isWordByte(src[i-1]) {
		return 0, false, "", false // raw prefixes must start a token
	}
	c := src[i]
	if r.luaLong && (c == '-' || c == '[') {
		return luaLongAt(src, i)
	}
	if r.cppRaw && (c == 'R' || c == 'u' || c == 'U' || c == 'L') {
		if end, reason, ok := cppRawAt(src, i); ok {
			return end, false, reason, true
		}
	}
	if r.rustRaw && (c == 'r' || c == 'b') {
		if end, reason, ok := rustRawAt(src, i); ok {
			return end, false, reason, true
		}
	}
	return 0, false, "", false
}

// cppRawAt matches `[u8|u|U|L]R"delim( ... )delim"`.
func cppRawAt(src []byte, i int) (int, string, bool) {
	j := i
	switch {
	case hasPrefixAt(src, j, "u8R", false):
		j += 2
	case src[j] == 'u' || src[j] == 'U' || src[j] == 'L':
		j++
	}
	if !hasPrefixAt(src, j, `R"`, false) {
		return 0, "", false
	}
	j += 2
	d := j
	for j < len(src) && j-d <= 16 {
		c := src[j]
		if c == '(' {
			break
		}
		if c == ' ' || c == ')' || c == '\\' || c == '\t' || c == '\n' || c == '"' {
			return 0, "", false
		}
		j++
	}
	if j >= len(src) || src[j] != '(' || j-d > 16 {
		return 0, "", false
	}
	closer := ")" + string(src[d:j]) + `"`
	k := bytes.Index(src[j+1:], []byte(closer))
	if k < 0 {
		return 0, "unterminated raw string at line " + itoa(lineOf(src, i)), true
	}
	return j + 1 + k + len(closer), "", true
}

// rustRawAt matches `r#"..."#` / `br##"..."##` with any number of hashes.
func rustRawAt(src []byte, i int) (int, string, bool) {
	j := i
	if src[j] == 'b' {
		j++
	}
	if j >= len(src) || src[j] != 'r' {
		return 0, "", false
	}
	j++
	h := j
	for j < len(src) && src[j] == '#' {
		j++
	}
	if j >= len(src) || src[j] != '"' {
		return 0, "", false
	}
	closer := `"` + strings.Repeat("#", j-h)
	k := bytes.Index(src[j+1:], []byte(closer))
	if k < 0 {
		return 0, "unterminated raw string at line " + itoa(lineOf(src, i)), true
	}
	return j + 1 + k + len(closer), "", true
}

// luaLongAt matches a `--[==[ ... ]==]` long comment or a `[==[ ... ]==]`
// long string.
func luaLongAt(src []byte, i int) (int, bool, string, bool) {
	j, isCmt := i, false
	if src[j] == '-' {
		if !hasPrefixAt(src, j, "--[", false) {
			return 0, false, "", false
		}
		j, isCmt = j+2, true
	}
	j++ // the first '['
	eq := j
	for j < len(src) && src[j] == '=' {
		j++
	}
	if j >= len(src) || src[j] != '[' {
		return 0, false, "", false
	}
	closer := "]" + strings.Repeat("=", j-eq) + "]"
	k := bytes.Index(src[j+1:], []byte(closer))
	if k < 0 {
		what := "long string"
		if isCmt {
			what = "block comment"
		}
		return len(src), isCmt, "unterminated " + what + " at line " + itoa(lineOf(src, i)), true
	}
	return j + 1 + k + len(closer), isCmt, "", true
}
