package scan

import (
	"bytes"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/More678/strip-comments/internal/lang"
)

// cmtKind tags an entry in the merged comment rule list.
type cmtKind uint8

const (
	kindLine cmtKind = iota
	kindBlock
)

// cmtRule points at one lang.LineComment or lang.BlockComment.
type cmtRule struct {
	kind cmtKind
	idx  int
	tok  string // Token for a line rule, Start for a block rule
	ci   bool
}

// recipe holds the lookup tables derived once from a *lang.Language: rules
// sorted longest token first, plus a first-byte map so the hot loop can skip
// uninteresting bytes without touching any rule list.
type recipe struct {
	strs      []int     // indexes into l.Strings, longest Start first
	cmts      []cmtRule // line and block rules, longest token first
	blockBase []string  // per block rule: shortest Start sharing its End
	cppRaw    bool
	rustRaw   bool
	luaLong   bool
	interest  [256]bool
}

var recipes sync.Map // *lang.Language -> *recipe

// recipeFor returns the cached recipe for l, building it on first use.
func recipeFor(l *lang.Language) *recipe {
	if v, ok := recipes.Load(l); ok {
		return v.(*recipe)
	}
	r := buildRecipe(l)
	recipes.Store(l, r)
	return r
}

func buildRecipe(l *lang.Language) *recipe {
	r := &recipe{}
	mark := func(tok string, ci bool) {
		if tok == "" {
			return
		}
		r.interest[tok[0]] = true
		if ci {
			r.interest[asciiLower(tok[0])] = true
			r.interest[asciiUpper(tok[0])] = true
		}
	}
	for i := range l.Strings {
		if l.Strings[i].Start == "" || l.Strings[i].End == "" {
			continue
		}
		r.strs = append(r.strs, i)
		mark(l.Strings[i].Start, false)
	}
	sort.SliceStable(r.strs, func(a, b int) bool {
		return len(l.Strings[r.strs[a]].Start) > len(l.Strings[r.strs[b]].Start)
	})
	for i := range l.Line {
		if l.Line[i].Token == "" {
			continue
		}
		r.cmts = append(r.cmts, cmtRule{kindLine, i, l.Line[i].Token, l.Line[i].CI})
		mark(l.Line[i].Token, l.Line[i].CI)
	}
	for i := range l.Block {
		if l.Block[i].Start == "" { // End == "" is allowed: closes right after Start
			continue
		}
		r.cmts = append(r.cmts, cmtRule{kindBlock, i, l.Block[i].Start, l.Block[i].CI})
		mark(l.Block[i].Start, l.Block[i].CI)
	}
	sort.SliceStable(r.cmts, func(a, b int) bool {
		return len(r.cmts[a].tok) > len(r.cmts[b].tok)
	})
	r.blockBase = make([]string, len(l.Block))
	for i := range l.Block {
		r.blockBase[i] = shortestBlockStart(l, i)
	}
	for _, name := range l.Raw {
		switch name {
		case "cpp_raw":
			r.cppRaw = true
			for _, c := range []byte{'R', 'u', 'U', 'L'} {
				r.interest[c] = true
			}
		case "rust_raw":
			r.rustRaw = true
			r.interest['r'], r.interest['b'] = true, true
		case "lua_long":
			r.luaLong = true
			r.interest['-'], r.interest['['] = true, true
		}
	}
	if l.Heredoc {
		r.interest['<'] = true
	}
	if l.Regex {
		r.interest['/'] = true
	}
	if l.CharLiteral {
		r.interest['\''] = true
	}
	if l.PHPRegions {
		r.interest['?'] = true
	}
	r.interest['\n'] = true
	return r
}

// pendingHeredoc is one `<<EOF` opener waiting for the end of its line.
type pendingHeredoc struct {
	id     string
	indent bool // `<<-` / `<<~`: the terminator may be indented
}

// scanner carries the mutable state of one Comments call.
type scanner struct {
	src     []byte
	l       *lang.Language
	r       *recipe
	spans   []Span
	bol     int  // offset of the first byte of the current line
	prevSig byte // last significant byte, ignoring whitespace and comments
	heredoc []pendingHeredoc
	hd      *heredocIndex // lazily built terminator index
	yamlOn  bool          // a block scalar opener was seen on the current line
	yamlInd int           // indentation of the line that opened it
	inPHP   bool
	stop    bool // EndMarker reached
}

// Comments returns every comment in src as an ascending list of half-open byte
// spans. unsafeReason is empty on success; when an unterminated string, raw
// literal or block comment is found it describes the problem and the spans
// discovered so far are returned.
func Comments(src []byte, l *lang.Language) (spans []Span, unsafeReason string) {
	if l == nil || len(src) == 0 {
		return nil, ""
	}
	s := &scanner{src: src, l: l, r: recipeFor(l), inPHP: !l.PHPRegions}
	reason := s.run()
	return s.spans, reason
}

// run is the main loop: at every byte the lexers are tried in priority order
// (regions, raw literals, heredocs, strings, chars, regex, comments) and the
// cursor otherwise advances one byte.
func (s *scanner) run() string {
	lastHook := -1
	for i := 0; i < len(s.src); {
		if i == s.bol && i != lastHook {
			lastHook = i
			ni := s.lineStart(i)
			if s.stop {
				return ""
			}
			if ni != i {
				i = ni
				continue
			}
		}
		if !s.inPHP {
			j := nextPHPOpen(s.src, i)
			if j < 0 {
				return ""
			}
			s.inPHP, s.prevSig = true, 0
			i = j
			s.resyncBOL(i)
			lastHook = -1
			continue
		}
		c := s.src[i]
		if c == '\n' {
			i = s.newline(i)
			continue
		}
		if !s.r.interest[c] {
			if !isSpace(c) {
				s.prevSig = c
			}
			i++
			continue
		}
		ni, reason := s.step(i)
		if reason != "" {
			return reason
		}
		i = ni
	}
	return ""
}

// step handles one interesting byte and returns the next cursor position.
func (s *scanner) step(i int) (int, string) {
	src, l, r := s.src, s.l, s.r
	if l.PHPRegions && src[i] == '?' && i+1 < len(src) && src[i+1] == '>' {
		s.inPHP, s.prevSig = false, 0
		return i + 2, ""
	}
	if r.cppRaw || r.rustRaw || r.luaLong {
		if end, isCmt, reason, ok := s.rawAt(i); ok {
			if reason != "" {
				if isCmt { // unterminated comment runs to EOF (used with Force)
					s.spans = append(s.spans, Span{Start: i, End: len(s.src)})
				}
				return 0, reason
			}
			if isCmt {
				s.spans = append(s.spans, Span{Start: i, End: end})
			}
			s.consumed(i, end, !isCmt)
			return end, ""
		}
	}
	if l.Heredoc {
		if end, ok := s.heredocAt(i); ok {
			return end, ""
		}
	}
	if end, reason, ok := s.stringAt(i); ok {
		if reason != "" {
			return 0, reason
		}
		s.consumed(i, end, true)
		return end, ""
	}
	if l.CharLiteral && src[i] == '\'' {
		if end, ok := charAt(src, i); ok {
			s.consumed(i, end, true)
			return end, ""
		}
	}
	if l.Regex && src[i] == '/' {
		if end, ok := s.regexAt(i); ok {
			s.consumed(i, end, true)
			return end, ""
		}
	}
	if start, end, doc, reason, ok := s.commentAt(i); ok {
		if reason != "" { // unterminated: the span runs to EOF (used with Force)
			s.spans = append(s.spans, Span{Start: start, End: len(src)})
			return 0, reason
		}
		s.spans = append(s.spans, Span{Start: start, End: end, Doc: doc})
		s.consumed(start, end, false)
		return end, ""
	}
	if !isSpace(src[i]) {
		s.prevSig = src[i]
	}
	return i + 1, ""
}

// consumed updates the line-start cache after a lexer swallowed [start, end),
// and records the last byte as significant when the range was not a comment.
func (s *scanner) consumed(start, end int, significant bool) {
	if bytes.IndexByte(s.src[start:end], '\n') >= 0 {
		s.resyncBOL(end)
	}
	if significant && end > start {
		if c := s.src[end-1]; !isSpace(c) {
			s.prevSig = c
		}
	}
}

// newline steps over the '\n' at i and then over any heredoc bodies or YAML
// block scalar queued by the line that just ended.
func (s *scanner) newline(i int) int {
	i++
	s.bol = i
	if len(s.heredoc) > 0 {
		for _, h := range s.heredoc {
			if end, ok := s.hdIndex().find(i, h.id, h.indent); ok {
				i = end
			}
		}
		s.heredoc = s.heredoc[:0]
		s.bol = i
	}
	if s.yamlOn {
		s.yamlOn = false
		i = skipYAMLBlock(s.src, i, s.yamlInd)
		s.bol = i
	}
	return i
}

// lineStart runs the line-level rules at the first byte of a line. It returns
// i unchanged when normal scanning should continue on this line.
func (s *scanner) lineStart(i int) int {
	src, l := s.src, s.l
	if l.MakeRecipe && src[i] == '\t' {
		return s.lineEnd(i)
	}
	if l.EndMarker != "" {
		line := bytes.TrimRight(src[i:s.lineEnd(i)], " \t\r")
		if string(line) == l.EndMarker {
			s.stop = true
			return i
		}
	}
	if l.YAMLBlock {
		if ind, ok := yamlOpener(src[i:s.lineEnd(i)]); ok {
			s.yamlOn, s.yamlInd = true, ind
		}
	}
	return i
}

// lineEnd returns the offset of the '\n' ending the line at i, or len(src).
func (s *scanner) lineEnd(i int) int {
	if j := bytes.IndexByte(s.src[i:], '\n'); j >= 0 {
		return i + j
	}
	return len(s.src)
}

// resyncBOL recomputes the start of the line containing off.
func (s *scanner) resyncBOL(off int) {
	if j := bytes.LastIndexByte(s.src[:off], '\n'); j >= 0 {
		s.bol = j + 1
	} else {
		s.bol = 0
	}
}

// stringAt scans a string literal starting at i. When the language also has
// strict char literals, a rule whose Start is `'` only applies if the text is
// not a valid char literal.
func (s *scanner) stringAt(i int) (int, string, bool) {
	src, l := s.src, s.l
	for _, ri := range s.r.strs {
		sr := &l.Strings[ri]
		if !hasPrefixAt(src, i, sr.Start, false) {
			continue
		}
		if sr.Start == "'" && l.CharLiteral {
			if _, ok := charAt(src, i); ok {
				return 0, "", false
			}
		}
		end, reason := s.stringBody(i, sr)
		return end, reason, true
	}
	return 0, "", false
}

// stringBody walks the inside of a literal and returns the offset just past
// its closing delimiter.
func (s *scanner) stringBody(i int, sr *lang.StringRule) (int, string) {
	src := s.src
	j := i + len(sr.Start)
	for j < len(src) {
		c := src[j]
		if sr.Escape != 0 && c == sr.Escape && j+1 < len(src) {
			j += 2
			continue
		}
		if hasPrefixAt(src, j, sr.End, false) {
			if sr.Doubled && hasPrefixAt(src, j+len(sr.End), sr.End, false) {
				j += 2 * len(sr.End)
				continue
			}
			return j + len(sr.End), ""
		}
		if c == '\n' && !sr.Multiline {
			return j, "" // unclosed single-line literal: ends at the newline
		}
		j++
	}
	if sr.Multiline {
		return 0, "unterminated string at line " + itoa(lineOf(src, i))
	}
	return len(src), ""
}

// commentAt tries every line and block rule, longest token first.
func (s *scanner) commentAt(i int) (start, end int, doc bool, reason string, ok bool) {
	for _, c := range s.r.cmts {
		if !hasPrefixAt(s.src, i, c.tok, c.ci) {
			continue
		}
		if c.kind == kindLine {
			if st, en, d, hit := s.lineCommentAt(i, &s.l.Line[c.idx]); hit {
				return st, en, d, "", true
			}
			continue
		}
		if st, en, d, rs, hit := s.blockCommentAt(i, c.idx); hit {
			return st, en, d, rs, true
		}
	}
	return 0, 0, false, "", false
}

// lineCommentAt checks the guards of a line rule whose token matched at i and
// returns the span, which runs to the end of the line, newline excluded.
func (s *scanner) lineCommentAt(i int, lc *lang.LineComment) (int, int, bool, bool) {
	src := s.src
	after := i + len(lc.Token)
	if lc.WordBoundary && i > 0 && strings.IndexByte(lang.WordBefore, src[i-1]) < 0 {
		return 0, 0, false, false
	}
	if lc.NotAfter != "" && i > 0 && strings.IndexByte(lc.NotAfter, src[i-1]) >= 0 {
		return 0, 0, false, false
	}
	if lc.NeedsSepAfter && after < len(src) &&
		strings.IndexByte(lang.SepAfter, src[after]) < 0 {
		return 0, 0, false, false
	}
	if lc.NotFollowedBy != "" && after < len(src) &&
		strings.IndexByte(lc.NotFollowedBy, src[after]) >= 0 {
		return 0, 0, false, false
	}
	start := i
	if lc.AtLineStart {
		j := i
		for j > s.bol && (isSpace(src[j-1]) ||
			strings.IndexByte(s.l.LineStartSkip, src[j-1]) >= 0) {
			j--
		}
		if j != s.bol {
			return 0, 0, false, false
		}
		// Swallow the LineStartSkip bytes too (batch `@REM`), so the whole
		// line disappears instead of leaving a bare `@` behind.
		for start > s.bol && strings.IndexByte(s.l.LineStartSkip, src[start-1]) >= 0 {
			start--
		}
	}
	end := after
	for {
		nl := bytes.IndexByte(src[end:], '\n')
		if nl < 0 {
			return start, len(src), lc.Doc, true
		}
		end += nl
		if s.l.PHPRegions {
			if k := bytes.Index(src[after:end], []byte("?>")); k >= 0 {
				return start, after + k, lc.Doc, true
			}
		}
		if !lc.BackslashCont || !endsWithBackslash(src[:end]) {
			return start, end, lc.Doc, true
		}
		end++
	}
}

// endsWithBackslash reports whether the last byte before the newline (ignoring
// a trailing '\r') is a '\'.
func endsWithBackslash(line []byte) bool {
	for len(line) > 0 && line[len(line)-1] == '\r' {
		line = line[:len(line)-1]
	}
	return len(line) > 0 && line[len(line)-1] == '\\'
}

// blockCommentAt scans a block comment whose Start matched at i.
func (s *scanner) blockCommentAt(i, ri int) (int, int, bool, string, bool) {
	src, bc := s.src, &s.l.Block[ri]
	if bc.AtLineStart && i != s.bol {
		return 0, 0, false, "", false
	}
	if bc.End == "" {
		// Self-contained token such as `/**/` or D's `/++/`: no closer search,
		// never unsafe, no nesting.
		return i, s.absorbBlanks(i, i+len(bc.Start)), bc.Doc, "", true
	}
	if bc.AtLineStart {
		return s.blockAtLineStart(i, bc)
	}
	base := s.r.blockBase[ri]
	j, depth := i+len(base), 1
	for j < len(src) {
		if hasPrefixAt(src, j, bc.End, bc.CI) {
			if depth--; depth == 0 {
				doc := bc.Doc && j >= i+len(bc.Start)
				return i, s.absorbBlanks(i, j+len(bc.End)), doc, "", true
			}
			j += len(bc.End)
			continue
		}
		if bc.Nested && hasPrefixAt(src, j, base, bc.CI) {
			depth++
			j += len(base)
			continue
		}
		j++
	}
	return i, len(src), false,
		"unterminated block comment at line " + itoa(lineOf(src, i)), true
}

// absorbBlanks extends the end of an inline block comment over the spaces
// and tabs that follow it when spaces or tabs also precede it, so
// `a; /* c */ }` becomes `a; }` rather than `a;  }`. Only the end moves:
// the comment text still starts the span, which keeps KeepRE and
// header detection in filter.go unaffected. Comments at BOL keep their
// right-hand indentation untouched.
func (s *scanner) absorbBlanks(start, end int) int {
	if start == 0 || (s.src[start-1] != ' ' && s.src[start-1] != '\t') {
		return end
	}
	for end < len(s.src) && (s.src[end] == ' ' || s.src[end] == '\t') {
		end++
	}
	return end
}

// blockAtLineStart handles rules like Ruby `=begin` / `=end`, where both
// delimiters must sit in column 0. The span covers whole lines.
func (s *scanner) blockAtLineStart(i int, bc *lang.BlockComment) (int, int, bool, string, bool) {
	src := s.src
	for j := i; j < len(src); {
		nl := bytes.IndexByte(src[j:], '\n')
		if nl < 0 {
			break
		}
		j += nl + 1
		if hasPrefixAt(src, j, bc.End, bc.CI) {
			end := j + nl2(src, j)
			return i, end, bc.Doc, "", true
		}
	}
	return i, len(src), false,
		"unterminated block comment at line " + itoa(lineOf(src, i)), true
}

// nl2 returns the distance from off to the end of its line.
func nl2(src []byte, off int) int {
	if j := bytes.IndexByte(src[off:], '\n'); j >= 0 {
		return j
	}
	return len(src) - off
}

// hasPrefixAt reports whether src[i:] starts with tok, optionally comparing
// ASCII letters case-insensitively.
func hasPrefixAt(src []byte, i int, tok string, ci bool) bool {
	if i < 0 || i+len(tok) > len(src) {
		return false
	}
	if !ci {
		return string(src[i:i+len(tok)]) == tok
	}
	for k := 0; k < len(tok); k++ {
		if asciiLower(src[i+k]) != asciiLower(tok[k]) {
			return false
		}
	}
	return true
}

func asciiLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 32
	}
	return c
}

func asciiUpper(c byte) byte {
	if c >= 'a' && c <= 'z' {
		return c - 32
	}
	return c
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '\v' || c == '\f'
}

// lineOf returns the 1-based line number of off.
func lineOf(src []byte, off int) int {
	if off > len(src) {
		off = len(src)
	}
	return 1 + bytes.Count(src[:off], []byte{'\n'})
}

func itoa(n int) string { return strconv.Itoa(n) }

// shortestBlockStart returns the base opener of block rule i: the shortest
// Start among the rules sharing its End that is a prefix of its own Start.
// The End search begins after the base, so `/**/` closes as an empty comment
// while `/*/` does not, and a plain `/*` nests inside a `/**` doc comment.
func shortestBlockStart(l *lang.Language, i int) string {
	tok := l.Block[i].Start
	for j := range l.Block {
		if l.Block[j].End == l.Block[i].End && len(l.Block[j].Start) < len(tok) &&
			strings.HasPrefix(tok, l.Block[j].Start) {
			tok = l.Block[j].Start
		}
	}
	return tok
}
