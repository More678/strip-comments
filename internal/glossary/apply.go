package glossary

import (
	"sort"
	"strings"
)

// Table is a prepared rule set: longest From first, so a longer term always
// wins over a prefix of itself.
type Table struct {
	rules []Rule
}

// NewTable prepares rules for matching. Empty From values are dropped, and a
// Word rule whose From has no ASCII edge is demoted to a plain match, since
// there is no boundary to test there.
func NewTable(rules []Rule) *Table {
	out := make([]Rule, 0, len(rules))
	for _, r := range rules {
		if r.From == "" {
			continue
		}
		if r.Word && !asciiEdges(r.From) {
			r.Word = false
		}
		out = append(out, r)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return len(out[i].From) > len(out[j].From)
	})
	return &Table{rules: out}
}

// Rules returns the prepared rules in match order.
func (t *Table) Rules() []Rule { return t.rules }

// matchAt reports the rule matching src at i, or nil.
func (t *Table) matchAt(src []byte, i int) *Rule {
	for k := range t.rules {
		r := &t.rules[k]
		if !hasAt(src, i, r.From, r.CI) {
			continue
		}
		if r.Word && !wordBounded(src, i, len(r.From)) {
			continue
		}
		return r
	}
	return nil
}

// hasAt reports whether term starts at src[i], optionally ignoring case.
func hasAt(src []byte, i int, term string, ci bool) bool {
	if i+len(term) > len(src) {
		return false
	}
	got := src[i : i+len(term)]
	if ci {
		return strings.EqualFold(string(got), term)
	}
	return string(got) == term
}

// wordBounded reports whether the n bytes at i are flanked by non-word bytes,
// so "pentest" does not match inside "pentester" or "my_pentest_run".
func wordBounded(src []byte, i, n int) bool {
	if i > 0 && isWordByte(src[i-1]) {
		return false
	}
	if j := i + n; j < len(src) && isWordByte(src[j]) {
		return false
	}
	return true
}

func isWordByte(c byte) bool {
	return c == '_' || c >= 0x80 ||
		(c >= '0' && c <= '9') ||
		(c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z')
}
