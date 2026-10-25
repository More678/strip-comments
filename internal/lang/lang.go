// Package lang describes the comment and literal syntax of each supported
// language as plain data. The scanner in internal/scan interprets it.
package lang

import (
	"regexp"
	"strings"
)

// WordBefore: `#` opens a shell comment only after one of these (or at BOL).
const WordBefore = " \t\r\n;&|(){}<>"

// SepAfter: batch `REM` must be followed by one of these (or EOL/EOF).
const SepAfter = " \t\r\n.:=,;()[]<>&|"

// RegexBefore: a `/` after one of these (or at BOL) starts a regex literal.
const RegexBefore = "(,=:[!&|?{};+-*%~^<>"

// RegexKeywords: identifiers that may be directly followed by a regex literal.
var RegexKeywords = map[string]bool{}

func init() {
	for _, w := range strings.Fields("return typeof case in of new delete void " +
		"throw do else yield await instanceof when and or not unless until " +
		"while if elsif then split match gsub sub grep scan each") {
		RegexKeywords[w] = true
	}
}

// LineComment runs from Token to the end of the line (newline not included).
type LineComment struct {
	Token         string
	AtLineStart   bool   // only whitespace (and LineStartSkip) may precede it
	WordBoundary  bool   // previous byte must be BOL or in WordBefore
	NeedsSepAfter bool   // next byte must be EOF or in SepAfter
	NotFollowedBy string // cancel if the next byte is one of these
	NotAfter      string // cancel if the previous byte is one of these
	CI            bool   // case-insensitive token match
	BackslashCont bool   // a `\` right before the newline continues it (C)
	Doc           bool   // documentation comment (kept with KeepDoc)
}

// BlockComment is delimited by Start ... End and may span lines.
type BlockComment struct {
	Start, End  string
	Nested      bool // `/* /* */ */` nesting (Rust, Haskell, Nim, Julia)
	AtLineStart bool // Start and End must sit in column 0 (Ruby =begin/=end)
	CI          bool
	Doc         bool
}

// StringRule is a literal whose contents are never scanned for comments.
type StringRule struct {
	Start, End string
	Escape     byte // escape byte inside the literal, 0 = none
	Doubled    bool // End written twice inside means a literal End (SQL '')
	Multiline  bool // false: an unclosed literal silently stops at EOL
}

// Language bundles everything the scanner needs for one file type.
type Language struct {
	Name      string
	Exts      []string // lower-case, with dot: ".cpp"
	Filenames []string // exact base names: "Makefile", "Dockerfile"
	Line      []LineComment
	Block     []BlockComment
	Strings   []StringRule
	// LineStartSkip: bytes ignored when testing AtLineStart (batch `@REM`).
	LineStartSkip string
	CharLiteral   bool     // strict `'x'` / `'\n'` literals (C, Java, Rust)
	Raw           []string // special lexers: "cpp_raw", "rust_raw", "lua_long"
	Heredoc       bool     // `<<EOF`, `<<-EOF`, `<<'EOF'`, `<<~EOF` bodies
	Regex         bool     // `/.../flags` literals (JS, Ruby)
	YAMLBlock     bool     // YAML `|` / `>` block scalars are opaque
	MakeRecipe    bool     // Makefile: tab-indented lines are left untouched
	PHPRegions    bool     // only scan inside `<?php ... ?>` / `<?= ... ?>`
	EndMarker     string   // a line equal to this ends the code (`__END__`)
	Keep          []string // RE2 patterns matched against the full comment text
	Risky         bool     // heuristic support only; opt-in via --lang/--ext

	keepRE *regexp.Regexp
}

// KeepRE returns the compiled Keep patterns (nil when there are none).
// Patterns are case-insensitive and tested with FindString on the whole
// comment text including its delimiters, e.g. "//go:build linux".
func (l *Language) KeepRE() *regexp.Regexp { return l.keepRE }

var (
	byName = map[string]*Language{}
	byExt  = map[string]*Language{}
	byFile = map[string]*Language{}
	order  []*Language
)

func register(l *Language) {
	if len(l.Keep) > 0 {
		l.keepRE = regexp.MustCompile("(?i)(?:" + strings.Join(l.Keep, ")|(?:") + ")")
	}
	byName[l.Name] = l
	for _, e := range l.Exts {
		byExt[e] = l
	}
	for _, f := range l.Filenames {
		byFile[strings.ToLower(f)] = l
	}
	order = append(order, l)
}

// ByName looks a language up by its Name ("go", "cpp", "python", ...).
func ByName(name string) *Language { return byName[strings.ToLower(name)] }

// ByExt looks a language up by extension; ".m" resolves to Objective-C.
func ByExt(ext string) *Language { return byExt[strings.ToLower(ext)] }

// ByFilename looks a language up by exact base name, falling back to ByExt.
func ByFilename(base string) *Language {
	if l := byFile[strings.ToLower(base)]; l != nil {
		return l
	}
	if i := strings.LastIndexByte(base, '.'); i > 0 {
		return ByExt(base[i:])
	}
	return nil
}

// All returns every registered language in registration order.
func All() []*Language { return order }
