// Package verify checks that a rewritten file is still structurally sound
// before it is written. It compares the file before and after the rewrite
// rather than judging the result on its own, so a file that was already
// unbalanced is not blamed on the tool.
package verify

import (
	"fmt"
	"go/parser"
	"go/token"

	"github.com/More678/strip-comments/internal/lang"
	"github.com/More678/strip-comments/internal/scan"
)

// pairs are the brackets whose balance must survive a rewrite.
var pairs = [...]struct{ open, close byte }{{'(', ')'}, {'[', ']'}, {'{', '}'}}

// Check returns an error when after is broken in a way before was not:
//   - the rewrite introduced an unterminated string or block comment;
//   - the net balance of (), [] or {} outside comments changed;
//   - for Go, before parses and after does not.
//
// Comments are excluded from the bracket count on both sides, so deleting a
// comment that happens to hold a `(` is never reported.
func Check(before, after []byte, l *lang.Language) error {
	bSpans, bReason := scan.Comments(before, l)
	aSpans, aReason := scan.Comments(after, l)
	if bReason == "" && aReason != "" {
		return fmt.Errorf("rewrite introduced an %s", aReason)
	}
	bc, ac := balance(before, bSpans), balance(after, aSpans)
	for i, p := range pairs {
		if bc[i] != ac[i] {
			return fmt.Errorf("balance of %c%c changed from %d to %d",
				p.open, p.close, bc[i], ac[i])
		}
	}
	if l.Name == "go" && parses(before) && !parses(after) {
		_, err := parser.ParseFile(token.NewFileSet(), "", after, parser.SkipObjectResolution)
		return fmt.Errorf("no longer parses as Go: %v", err)
	}
	return nil
}

// balance counts opens minus closes for each pair, skipping comment spans.
func balance(src []byte, spans []scan.Span) [len(pairs)]int {
	var n [len(pairs)]int
	k := 0
	for i := 0; i < len(src); i++ {
		for k < len(spans) && spans[k].End <= i {
			k++
		}
		if k < len(spans) && i >= spans[k].Start {
			i = spans[k].End - 1
			continue
		}
		for j, p := range pairs {
			switch src[i] {
			case p.open:
				n[j]++
			case p.close:
				n[j]--
			}
		}
	}
	return n
}

func parses(src []byte) bool {
	_, err := parser.ParseFile(token.NewFileSet(), "", src, parser.SkipObjectResolution)
	return err == nil
}
