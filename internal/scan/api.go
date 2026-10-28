// Package scan removes comments from source text using the declarative
// syntax tables in internal/lang.
//
// Input is raw bytes with "\n" line endings (internal/textio normalises CRLF
// and strips the BOM first). All delimiters are ASCII, so UTF-8 / GBK /
// Shift-JIS content passes through byte-for-byte without decoding.
package scan

import (
	"github.com/More678/strip-comments/internal/glossary"
	"github.com/More678/strip-comments/internal/lang"
)

// Options tweaks what counts as removable.
type Options struct {
	KeepDoc   bool // keep comments whose rule has Doc=true (///, /** */)
	KeepBlank bool // keep lines that become empty instead of deleting them
	// KeepLicense keeps the leading comment header: the first comment (only
	// whitespace or a shebang before it) plus every following comment that is
	// separated from the previous one by whitespace without a blank line.
	KeepLicense bool
	Force       bool // strip even when an unterminated literal/comment is found

	// Glossary rewrites terminology in code and comments. Nil disables it,
	// which leaves the comment-only behaviour exactly as it was.
	Glossary *glossary.Table
	// TermsInComments restricts the glossary to comments, leaving code alone.
	TermsInComments bool
	// KeepComments runs the glossary without removing any comments, for when
	// the point is to restate the vocabulary rather than strip the file.
	KeepComments bool
}

// Result of stripping one buffer.
type Result struct {
	Output       []byte         // equals the input when Unsafe && !Force
	Comments     int            // comments removed
	RemovedLines int            // lines deleted entirely
	Unsafe       bool           // an unterminated string/block comment was hit
	Reason       string         // human readable cause when Unsafe
	Terms        []glossary.Hit // terminology replacements, in file order
}

// Span is a half-open byte range [Start, End) of one comment.
type Span struct {
	Start, End int
	Doc        bool
}

// Strip removes comments from src according to l.
//
// Rebuild rules (see rebuild.go):
//   - Comment bytes are deleted, but newlines inside a block comment are kept
//     so the surrounding line structure is preserved.
//   - A line that contained a removed comment and is now whitespace-only is
//     deleted together with its newline (unless KeepBlank). Lines that were
//     blank in the original are never touched.
//   - A line that still has code gets its trailing whitespace trimmed, so
//     `x = 1  // note` becomes `x = 1`.
//   - An inline block comment with blanks on both sides also absorbs the
//     blanks after it, so `a /* x */ b` becomes `a b`, while `f(/* x */ 1)`
//     (no blank before) becomes `f( 1)`.
//   - Comments matching l.KeepRE() and a `#!` shebang on line 1 are kept.
func Strip(src []byte, l *lang.Language, opt Options) Result {
	spans, reason := Comments(src, l)
	if reason != "" && !opt.Force {
		return Result{Output: src, Unsafe: true, Reason: reason}
	}

	// Terms go first: removing comments first would delete the terms inside
	// them before the report could show them.
	orig := src
	var terms []glossary.Hit
	if opt.Glossary != nil {
		var rewritten []byte
		rewritten, terms = opt.Glossary.Apply(src, glossary.Options{
			Comments:     regions(spans),
			CommentsOnly: opt.TermsInComments,
		})
		if len(terms) > 0 {
			// Replacements change lengths, so every span after the first hit
			// is now off. Re-lex rather than shift offsets by hand.
			src = rewritten
			spans, reason = Comments(src, l)
			if reason != "" && !opt.Force {
				// A replacement opened a literal it did not close. Refuse
				// rather than strip a file whose structure just changed.
				// Output stays the original, per the Result contract.
				return Result{Output: orig, Unsafe: true,
					Reason: "after term replacement: " + reason, Terms: terms}
			}
		}
	}

	if opt.KeepComments {
		return Result{Output: src, Unsafe: reason != "", Reason: reason,
			Terms: terms}
	}
	spans = filter(src, spans, l, opt)
	out, removed := rebuild(src, spans, opt.KeepBlank)
	return Result{Output: out, Comments: len(spans), RemovedLines: removed,
		Unsafe: reason != "", Reason: reason, Terms: terms}
}

// regions converts comment spans to the shape the glossary expects.
func regions(spans []Span) []glossary.Region {
	out := make([]glossary.Region, len(spans))
	for i, s := range spans {
		out[i] = glossary.Region{Start: s.Start, End: s.End, Comment: true}
	}
	return out
}
