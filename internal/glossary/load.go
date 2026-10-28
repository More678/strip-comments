package glossary

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// jsonRule is the on-disk shape of a rule in a .json term file.
type jsonRule struct {
	From string `json:"from"`
	To   string `json:"to"`
	Word *bool  `json:"word,omitempty"`
	CI   *bool  `json:"ci,omitempty"`
}

// Load reads a term table from path. A ".json" extension is parsed as JSON,
// anything else as TSV.
//
// TSV is one rule per line, `from<TAB>to`, with an optional third column of
// comma-separated flags (word, noword, ci, nocase). Blank lines and lines
// starting with `#` are skipped.
//
// When the flags are not given they are inferred the way the default table is
// written: a term with ASCII edges gets word-boundary and case-insensitive
// matching, a CJK term gets neither.
func Load(path string) ([]Rule, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(filepath.Ext(path), ".json") {
		return parseJSON(raw, path)
	}
	return parseTSV(raw, path)
}

func parseJSON(raw []byte, path string) ([]Rule, error) {
	var jr []jsonRule
	if err := json.Unmarshal(raw, &jr); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	out := make([]Rule, 0, len(jr))
	for i, r := range jr {
		if r.From == "" {
			return nil, fmt.Errorf("%s: entry %d has an empty \"from\"", path, i)
		}
		rule := inferred(r.From, r.To)
		if r.Word != nil {
			rule.Word = *r.Word
		}
		if r.CI != nil {
			rule.CI = *r.CI
		}
		out = append(out, rule)
	}
	return out, nil
}

func parseTSV(raw []byte, path string) ([]Rule, error) {
	var out []Rule
	sc := bufio.NewScanner(bytes.NewReader(raw))
	sc.Buffer(make([]byte, 0, 64<<10), 1<<20)
	for line := 1; sc.Scan(); line++ {
		text := strings.TrimRight(sc.Text(), " \t\r")
		if text == "" || strings.HasPrefix(strings.TrimSpace(text), "#") {
			continue
		}
		cols := strings.Split(text, "\t")
		if len(cols) < 2 {
			return nil, fmt.Errorf("%s:%d: want from<TAB>to, got %q",
				path, line, text)
		}
		from := strings.TrimSpace(cols[0])
		if from == "" {
			return nil, fmt.Errorf("%s:%d: empty term", path, line)
		}
		rule := inferred(from, strings.TrimSpace(cols[1]))
		if len(cols) > 2 {
			if err := applyFlags(&rule, cols[2]); err != nil {
				return nil, fmt.Errorf("%s:%d: %w", path, line, err)
			}
		}
		out = append(out, rule)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return out, nil
}

// inferred builds a rule with the flags the default table would use.
func inferred(from, to string) Rule {
	ascii := asciiEdges(from)
	return Rule{From: from, To: to, Word: ascii, CI: ascii}
}

// applyFlags reads a comma-separated flag column.
func applyFlags(r *Rule, col string) error {
	for _, f := range strings.Split(col, ",") {
		switch strings.ToLower(strings.TrimSpace(f)) {
		case "":
		case "word":
			r.Word = true
		case "noword":
			r.Word = false
		case "ci":
			r.CI = true
		case "nocase", "cs":
			r.CI = false
		default:
			return fmt.Errorf("unknown flag %q, want word/noword/ci/nocase", f)
		}
	}
	return nil
}
