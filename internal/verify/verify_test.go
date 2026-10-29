package verify

import (
	"strings"
	"testing"

	"github.com/More678/strip-comments/internal/lang"
)

func TestCheck(t *testing.T) {
	cases := []struct {
		name, lang, before, after string
		wantErr                   string // "" means Check must pass
	}{
		{
			// Brackets inside the deleted comment must not count.
			name: "deleting a comment holding brackets is fine", lang: "c",
			before: "int x = 1; /* see f(a[0]) { */\n",
			after:  "int x = 1;\n",
		},
		{
			name: "a lost paren is caught", lang: "c",
			before:  "int x = f(1);\n",
			after:   "int x = f(1;\n",
			wantErr: "balance of ()",
		},
		{
			name: "a lost brace is caught", lang: "python",
			before:  "d = {'a': 1}\n",
			after:   "d = {'a': 1\n",
			wantErr: "balance of {}",
		},
		{
			// Before was already unbalanced; the rewrite kept it that way.
			// That is not the tool's doing and must not block the write.
			name: "pre-existing imbalance is not blamed", lang: "c",
			before: "int x = f(1; // c\n",
			after:  "int x = f(1;\n",
		},
		{
			name: "an unterminated string introduced is caught", lang: "go",
			before:  "package x\n\nvar s = 1\n",
			after:   "package x\n\nvar s = `1\n",
			wantErr: "unterminated",
		},
		{
			// Balanced brackets, but no longer Go: only the parser sees it.
			name: "go syntax broken is caught", lang: "go",
			before:  "package x\n\nvar n = 1\n",
			after:   "package x\n\nvar n = = 1\n",
			wantErr: "no longer parses",
		},
		{
			name: "go file that never parsed is not blamed", lang: "go",
			before: "package x\n\nvar = 1 // c\n",
			after:  "package x\n\nvar = 1\n",
		},
		{
			name: "brackets inside strings are counted consistently", lang: "c",
			before: "char *s = \"(\"; // c\n",
			after:  "char *s = \"(\";\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := lang.ByName(tc.lang)
			if l == nil {
				t.Fatalf("language %q not registered", tc.lang)
			}
			err := Check([]byte(tc.before), []byte(tc.after), l)
			switch {
			case tc.wantErr == "" && err != nil:
				t.Errorf("Check failed, want pass: %v", err)
			case tc.wantErr != "" && err == nil:
				t.Errorf("Check passed, want an error mentioning %q", tc.wantErr)
			case tc.wantErr != "" && !strings.Contains(err.Error(), tc.wantErr):
				t.Errorf("error %q does not mention %q", err, tc.wantErr)
			}
		})
	}
}
