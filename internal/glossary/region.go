package glossary

import (
	"bytes"
	"strings"
)

// Region marks a byte range that Apply treats specially.
type Region struct {
	Start, End int
	Comment    bool // the range is a comment
}

// Options controls Apply.
type Options struct {
	// Comments are the comment ranges in the buffer, from the lexer. Anything
	// outside them counts as code.
	Comments []Region
	// CommentsOnly restricts replacement to comments, leaving code alone.
	CommentsOnly bool
}

// noTouch lists line prefixes whose contents must never be rewritten: an
// import path or an include is resolved by the toolchain, so renaming a term
// inside one turns a working build into a missing-package error.
var noTouch = []string{
	"import", "from", "#include", "#import", "require", "require_once",
	"include", "include_once", "using", "package", "module", "go:generate",
	"use", "extern crate", "source", "load",
}

// skipLine reports whether the line holding src[i] is one whose contents are
// resolved by a toolchain rather than read by a person.
func skipLine(src []byte, i int) bool {
	start := bytes.LastIndexByte(src[:i], '\n') + 1
	end := bytes.IndexByte(src[i:], '\n')
	if end < 0 {
		end = len(src)
	} else {
		end += i
	}
	line := strings.TrimSpace(strings.ToLower(string(src[start:end])))
	line = strings.TrimLeft(line, "/*#-;% \t")
	for _, p := range noTouch {
		if line == p || strings.HasPrefix(line, p+" ") ||
			strings.HasPrefix(line, p+"\t") || strings.HasPrefix(line, p+"(") ||
			strings.HasPrefix(line, p+"\"") || strings.HasPrefix(line, p+"<") {
			return true
		}
	}
	return false
}

// urlSchemes introduce a token that is resolved by something other than the
// reader, so the term inside it has to stay exactly as written.
var urlSchemes = []string{
	"http://", "https://", "ftp://", "git://", "ssh://", "file://",
	"mailto:", "git@", "www.",
}

// inURL reports whether i sits inside a URL or a URL-like token on its line.
// Rewriting a term inside one would silently change where a request goes.
func inURL(src []byte, i int) bool {
	start := bytes.LastIndexByte(src[:i], '\n') + 1
	lower := strings.ToLower(string(src[start:i]))
	for _, s := range urlSchemes {
		at := strings.LastIndex(lower, s)
		if at < 0 {
			continue
		}
		// The URL runs until whitespace or a quote ends the token.
		if !strings.ContainsAny(lower[at:], " \t\"'`<>()") {
			return true
		}
	}
	return false
}
