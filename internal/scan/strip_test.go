package scan

import (
	"testing"

	"github.com/More678/strip-comments/internal/lang"
)

// cnt wraps a literal so a table case can distinguish "expect 0" from
// "don't check".
func cnt(n int) *int { return &n }

// required languages: a missing one is a bug in the language table, not a
// reason to skip. Everything else is skipped when not registered yet.
var required = []string{"c", "cpp", "java", "go", "rust", "python", "bat",
	"shell", "javascript"}

type stripCase struct {
	name         string
	lang         string
	opts         Options
	in           string
	want         string
	wantUnsafe   bool
	wantComments *int // nil = not checked
	wantRemoved  *int // nil = not checked
}

func TestStrip(t *testing.T) {
	for _, name := range required {
		if lang.ByName(name) == nil {
			t.Errorf("required language %q is not registered", name)
		}
	}
	for _, tc := range stripCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			l := lang.ByName(tc.lang)
			if l == nil {
				t.Skipf("language %q not registered", tc.lang)
			}
			got := Strip([]byte(tc.in), l, tc.opts)
			if string(got.Output) != tc.want {
				t.Errorf("Strip() output mismatch\n in:   %q\n got:  %q\n want: %q",
					tc.in, got.Output, tc.want)
			}
			if got.Unsafe != tc.wantUnsafe {
				t.Errorf("Unsafe = %v, want %v (reason %q)",
					got.Unsafe, tc.wantUnsafe, got.Reason)
			}
			if tc.wantUnsafe && !tc.opts.Force && string(got.Output) != tc.in {
				t.Errorf("unsafe without Force must return the input verbatim, got %q",
					got.Output)
			}
			if tc.wantComments != nil && got.Comments != *tc.wantComments {
				t.Errorf("Comments = %d, want %d", got.Comments, *tc.wantComments)
			}
			if tc.wantRemoved != nil && got.RemovedLines != *tc.wantRemoved {
				t.Errorf("RemovedLines = %d, want %d", got.RemovedLines, *tc.wantRemoved)
			}
		})
	}
}

var stripCases = []stripCase{
	// ---------------------------------------------------------------- C
	{
		name: "c/trailing comment is trimmed",
		lang: "c",
		in:   "int x = 1; // note\nint y = 2;\n",
		want: "int x = 1;\nint y = 2;\n",

		wantComments: cnt(1),
		wantRemoved:  cnt(0),
	},
	{
		name: "c/whole line comment is deleted with its newline",
		lang: "c",
		in:   "int a;\n    // just a comment\nint b;\n",
		want: "int a;\nint b;\n",

		wantComments: cnt(1),
		wantRemoved:  cnt(1),
	},
	{
		name: "c/multi line block comment keeps no blank lines",
		lang: "c",
		in:   "int a;\n/* one\n   two */\nint b;\n",
		want: "int a;\nint b;\n",

		wantComments: cnt(1),
		wantRemoved:  cnt(2),
	},
	{
		name: "c/inline block comment leaves surrounding code",
		lang: "c",
		in:   "f(/* x */ 1);\n",
		want: "f( 1);\n",

		wantComments: cnt(1),
	},
	{
		name: "c/comment markers inside strings are literal",
		lang: "c",
		in:   "char *u = \"http://x\"; // real\nchar *h = \"#\";\nchar *d = \"--\";\n",
		want: "char *u = \"http://x\";\nchar *h = \"#\";\nchar *d = \"--\";\n",

		wantComments: cnt(1),
	},
	{
		name: "c/char literals quote and double quote",
		lang: "c",
		in:   "char q = '\\''; // one\nchar d = '\"'; // two\n",
		want: "char q = '\\'';\nchar d = '\"';\n",

		wantComments: cnt(2),
	},
	{
		// The apostrophe in can't must not open a char literal, so the
		// following // c is still a comment.
		name: "c/apostrophe in error directive",
		lang: "c",
		in:   "#error can't happen\nint x; // c\n",
		want: "#error can't happen\nint x;\n",

		wantComments: cnt(1),
	},
	{
		// LineComment.BackslashCont: the continued line is part of the
		// same comment, so both lines go and it counts once.
		name: "c/backslash continues a line comment",
		lang: "c",
		in:   "int a;\n// first \\\nstill comment\nint b;\n",
		want: "int a;\nint b;\n",

		wantComments: cnt(1),
		wantRemoved:  cnt(2),
	},
	{
		name: "c/empty block comment followed by code",
		lang: "c",
		in:   "/**/int x = 1;\n",
		want: "int x = 1;\n",

		wantComments: cnt(1),
	},
	{
		name:       "c/unterminated block comment is unsafe",
		lang:       "c",
		in:         "int a;\n/* never closed\nint b;\n",
		want:       "int a;\n/* never closed\nint b;\n",
		wantUnsafe: true,
	},
	{
		// Force accepts the damage: the unterminated comment runs to EOF,
		// so everything after `int a;` disappears.
		name:       "c/unterminated block comment with Force",
		lang:       "c",
		opts:       Options{Force: true},
		in:         "int a;\n/* never closed\nint b;\n",
		want:       "int a;\n",
		wantUnsafe: true,
	},
	{
		// StringRule.Multiline is false for C, so an unclosed "..." stops
		// silently at EOL instead of poisoning the rest of the file.
		name: "c/unterminated single line string stops at EOL",
		lang: "c",
		in:   "char *s = \"open;\nint b; // c\n",
		want: "char *s = \"open;\nint b;\n",

		wantComments: cnt(1),
	},
	{
		name: "c/file without final newline",
		lang: "c",
		in:   "int x = 1; // note",
		want: "int x = 1;",

		wantComments: cnt(1),
	},
	{
		name: "c/only comments yields empty output",
		lang: "c",
		in:   "// a\n/* b */\n// c\n",
		want: "",

		wantComments: cnt(3),
		wantRemoved:  cnt(3),
	},
	{
		name: "c/original blank lines are preserved",
		lang: "c",
		in:   "int a;\n\n// gone\n\nint b;\n",
		want: "int a;\n\n\nint b;\n",

		wantComments: cnt(1),
		wantRemoved:  cnt(1),
	},
	{
		// Only lines that lost a comment are right-trimmed; a line the
		// scanner never touched is emitted verbatim, trailing blanks and all.
		name: "c/untouched line keeps its trailing whitespace",
		lang: "c",
		in:   "int a;   \nint b; // c\n",
		want: "int a;   \nint b;\n",

		wantComments: cnt(1),
	},
	{
		name: "c/KeepBlank leaves an empty line behind",
		lang: "c",
		opts: Options{KeepBlank: true},
		in:   "int a;\n// gone\nint b;\n",
		want: "int a;\n\nint b;\n",

		wantComments: cnt(1),
		wantRemoved:  cnt(0),
	},
	{
		name: "c/KeepLicense keeps the leading banner only",
		lang: "c",
		opts: Options{KeepLicense: true},
		in:   "/* Copyright 2024 */\nint a; // gone\n",
		want: "/* Copyright 2024 */\nint a;\n",

		wantComments: cnt(1),
	},

	{
		// KeepLicense keeps the whole leading run after a shebang; a blank
		// line ends the run, so the comment after it is removed.
		name: "shell/KeepLicense run stops at a blank line",
		lang: "shell",
		opts: Options{KeepLicense: true},
		in:   "#!/bin/sh\n# Copyright 2024\n# MIT License\n\n# gone\necho hi\n",
		want: "#!/bin/sh\n# Copyright 2024\n# MIT License\n\necho hi\n",

		wantComments: cnt(1),
		wantRemoved:  cnt(1),
	},
	{
		name: "c/KeepDoc keeps a block doc comment",
		lang: "c",
		opts: Options{KeepDoc: true},
		in:   "/** doc */\nint a; /* plain */\n",
		want: "/** doc */\nint a;\n",

		wantComments: cnt(1),
	},

	// -------------------------------------------------------------- C++
	{
		name: "cpp/raw string hides comment markers",
		lang: "cpp",
		in:   "auto s = R\"x( // not /* a */ comment )x\"; // c\n",
		want: "auto s = R\"x( // not /* a */ comment )x\";\n",

		wantComments: cnt(1),
	},
	{
		// The `)"` inside does not close a delimited raw string.
		name: "cpp/raw string with fake terminator",
		lang: "cpp",
		in:   "auto s = R\"d(a)\" // b)d\"; // c\n",
		want: "auto s = R\"d(a)\" // b)d\";\n",

		wantComments: cnt(1),
	},
	{
		name: "cpp/multi line raw string",
		lang: "cpp",
		in:   "auto s = R\"(\n// line\n)\";\n// gone\n",
		want: "auto s = R\"(\n// line\n)\";\n",

		wantComments: cnt(1),
	},
	{
		name: "cpp/digit separator is not a char literal",
		lang: "cpp",
		in:   "int n = 1'000'000; // c\n",
		want: "int n = 1'000'000;\n",

		wantComments: cnt(1),
	},

	// --------------------------------------------------------------- Go
	{
		name: "go/backtick raw string with slashes",
		lang: "go",
		in:   "var s = `// not\n/* a */ comment` // c\n",
		want: "var s = `// not\n/* a */ comment`\n",

		wantComments: cnt(1),
	},
	{
		name: "go/build and generate directives are kept",
		lang: "go",
		in: "//go:build linux\n\n// Package p is gone.\npackage p\n\n" +
			"//go:generate stringer -type=T\ntype T int // c\n",
		want: "//go:build linux\n\npackage p\n\n" +
			"//go:generate stringer -type=T\ntype T int\n",

		wantComments: cnt(2),
		wantRemoved:  cnt(1),
	},
	{
		// The comment right before import "C" is C source (cgo preamble).
		name: "go/cgo preamble is kept",
		lang: "go",
		in: "package p\n\n// #include <stdio.h>\n// #define N 1\nimport \"C\"\n\n" +
			"var x = 1 // c\n",
		want: "package p\n\n// #include <stdio.h>\n// #define N 1\nimport \"C\"\n\n" +
			"var x = 1\n",

		wantComments: cnt(1),
	},
	{
		name: "go/cgo block preamble is kept",
		lang: "go",
		in:   "package p\n\n/*\n#include <stdio.h>\n*/\nimport \"C\"\n",
		want: "package p\n\n/*\n#include <stdio.h>\n*/\nimport \"C\"\n",

		wantComments: cnt(0),
	},
	{
		name: "go/rune literal with quote",
		lang: "go",
		in:   "r := '\"' // c\nq := '\\'' // d\n",
		want: "r := '\"'\nq := '\\''\n",

		wantComments: cnt(2),
	},

	// ------------------------------------------------------------- Rust
	{
		name: "rust/nested block comment",
		lang: "rust",
		in:   "let a = 1; /* outer /* inner */ still outer */\nlet b = 2;\n",
		want: "let a = 1;\nlet b = 2;\n",

		wantComments: cnt(1),
	},
	{
		name: "rust/raw string with hashes",
		lang: "rust",
		in:   "let s = r#\"// not \"quoted\" /* x */\"#; // c\n",
		want: "let s = r#\"// not \"quoted\" /* x */\"#;\n",

		wantComments: cnt(1),
	},
	{
		// 'a must not open a char literal that swallows the comment.
		name: "rust/lifetimes are not char literals",
		lang: "rust",
		in:   "fn f<'a>(x: &'a str) -> &'a str { x } // c\n",
		want: "fn f<'a>(x: &'a str) -> &'a str { x }\n",

		wantComments: cnt(1),
	},
	{
		name: "rust/char literals",
		lang: "rust",
		in:   "let c = '\"'; let d = '\\''; // c\n",
		want: "let c = '\"'; let d = '\\'';\n",

		wantComments: cnt(1),
	},
	{
		name: "rust/doc comments removed by default",
		lang: "rust",
		in:   "/// Adds.\n//! Crate doc.\nfn add() {} // c\n",
		want: "fn add() {}\n",

		wantComments: cnt(3),
		wantRemoved:  cnt(2),
	},
	{
		name: "rust/doc comments kept with KeepDoc",
		lang: "rust",
		opts: Options{KeepDoc: true},
		in:   "/// Adds.\nfn add() {} // c\n",
		want: "/// Adds.\nfn add() {}\n",

		wantComments: cnt(1),
	},

	// ------------------------------------------------------------- Java
	{
		name: "java/text block containing slashes",
		lang: "java",
		in: "String s = \"\"\"\n    // not a comment\n    /* nor this */\n    \"\"\"; // c\n" +
			"int x = 1;\n",
		want: "String s = \"\"\"\n    // not a comment\n    /* nor this */\n    \"\"\";\n" +
			"int x = 1;\n",

		wantComments: cnt(1),
	},
	{
		name: "java/javadoc removed by default",
		lang: "java",
		in:   "/** Doc. */\nclass A {} /* plain */\n",
		want: "class A {}\n",

		wantComments: cnt(2),
	},
	{
		name: "java/javadoc kept with KeepDoc",
		lang: "java",
		opts: Options{KeepDoc: true},
		in:   "/** Doc. */\nclass A {} /* plain */\n",
		want: "/** Doc. */\nclass A {}\n",

		wantComments: cnt(1),
	},

	// ----------------------------------------------------------- Python
	{
		name: "python/shebang coding and noqa are kept",
		lang: "python",
		in: "#!/usr/bin/env python3\n# -*- coding: utf-8 -*-\n# gone\n" +
			"import os # noqa\nx = 1 # c\n",
		want: "#!/usr/bin/env python3\n# -*- coding: utf-8 -*-\n" +
			"import os # noqa\nx = 1\n",

		wantComments: cnt(2),
		wantRemoved:  cnt(1),
	},
	{
		name: "python/hash inside a string",
		lang: "python",
		in:   "x = \"#\" # c\ny = '# also' # d\n",
		want: "x = \"#\"\ny = '# also'\n",

		wantComments: cnt(2),
	},
	{
		name: "python/triple quoted string containing hash",
		lang: "python",
		in:   "s = \"\"\"\n# not a comment\n\"\"\"\nt = '''# nor'''  # c\n",
		want: "s = \"\"\"\n# not a comment\n\"\"\"\nt = '''# nor'''\n",

		wantComments: cnt(1),
	},
	{
		name:       "python/unterminated triple quote is unsafe",
		lang:       "python",
		in:         "x = 1 # c\ns = \"\"\"\n# inside\n",
		want:       "x = 1 # c\ns = \"\"\"\n# inside\n",
		wantUnsafe: true,
	},
	{
		// Force strips what was found before the open literal; the rest
		// is still treated as string content.
		name:       "python/unterminated triple quote with Force",
		lang:       "python",
		opts:       Options{Force: true},
		in:         "x = 1 # c\ns = \"\"\"\n# inside\n",
		want:       "x = 1\ns = \"\"\"\n# inside\n",
		wantUnsafe: true,
	},

	// ------------------------------------------------------------ Batch
	{
		name: "bat/REM variants and labels comment",
		lang: "bat",
		in: "@echo off\nREM x\n@REM x\nrem x\n:: x\nREM\n" +
			"echo hi\n",
		want: "@echo off\necho hi\n",

		wantComments: cnt(5),
		wantRemoved:  cnt(5),
	},
	{
		name: "bat/REM lookalikes are untouched",
		lang: "bat",
		in:   "REMOVE.exe /q\necho a::b\necho \"REM\"\nset REMOTE=1\n",
		want: "REMOVE.exe /q\necho a::b\necho \"REM\"\nset REMOTE=1\n",

		wantComments: cnt(0),
	},
	{
		name: "bat/REM followed by separator",
		lang: "bat",
		in:   "REM.\nrem:note\necho ok\n",
		want: "echo ok\n",

		wantComments: cnt(2),
	},

	// ------------------------------------------------------------ Shell
	{
		name: "shell/hash inside a word and trailing comment",
		lang: "shell",
		in:   "#!/bin/sh\n# gone\necho a#b # c\n",
		want: "#!/bin/sh\necho a#b\n",

		wantComments: cnt(2),
		wantRemoved:  cnt(1),
	},
	{
		name: "shell/parameter expansions with hash",
		lang: "shell",
		in:   "n=${#arr[@]}\ns=${v#pre} # c\necho $#\n",
		want: "n=${#arr[@]}\ns=${v#pre}\necho $#\n",

		wantComments: cnt(1),
	},
	{
		name: "shell/quoted hash",
		lang: "shell",
		in:   "a=$'#'\nb=\"# x\" # c\nc='#'\n",
		want: "a=$'#'\nb=\"# x\"\nc='#'\n",

		wantComments: cnt(1),
	},
	{
		name: "shell/heredoc body is untouched",
		lang: "shell",
		in: "cat <<EOF # c\n# not a comment\nEOF\n" +
			"cat <<-'END'\n\t# nor this\n\tEND\necho done # d\n",
		want: "cat <<EOF\n# not a comment\nEOF\n" +
			"cat <<-'END'\n\t# nor this\n\tEND\necho done\n",

		wantComments: cnt(2),
	},

	// ------------------------------------------------------- PowerShell
	{
		name: "powershell/line and block comments",
		lang: "powershell",
		in:   "<# block\n   comment #>\n$x = 1 # c\n$s = \"# not\"\n",
		want: "$x = 1\n$s = \"# not\"\n",

		wantComments: cnt(2),
		wantRemoved:  cnt(2),
	},

	// ------------------------------------------------------- JavaScript
	{
		name: "javascript/regex literal with escaped slashes",
		lang: "javascript",
		in:   "a = /\\/\\//g; // c\n",
		want: "a = /\\/\\//g;\n",

		wantComments: cnt(1),
	},
	{
		name: "javascript/division is not a regex",
		lang: "javascript",
		in:   "z = x / y // c\nw = (a) / 2 / b; // d\n",
		want: "z = x / y\nw = (a) / 2 / b;\n",

		wantComments: cnt(2),
	},
	{
		name: "javascript/regex after keyword",
		lang: "javascript",
		in:   "function f() { return /\\/\\*/.test(s); } // c\n",
		want: "function f() { return /\\/\\*/.test(s); }\n",

		wantComments: cnt(1),
	},
	{
		name: "javascript/template literal with slashes",
		lang: "javascript",
		in:   "const u = `http://x/${a /* in expr */}\n// not`; // c\n",
		want: "const u = `http://x/${a /* in expr */}\n// not`;\n",

		wantComments: cnt(1),
	},
	{
		name: "javascript/strings with comment markers",
		lang: "javascript",
		in:   "s = \"http://x\" + '/* no */'; // c\n",
		want: "s = \"http://x\" + '/* no */';\n",

		wantComments: cnt(1),
	},

	// ------------------------------------------------------- TypeScript
	{
		name: "typescript/line and block comments",
		lang: "typescript",
		in:   "/* header */\nlet x: number = 1; // c\n",
		want: "let x: number = 1;\n",

		wantComments: cnt(2),
	},

	// ----------------------------------------------------------- C#
	{
		name: "csharp/verbatim string with slashes",
		lang: "csharp",
		in:   "var p = @\"C:\\// \"\"q\"\"\"; // c\n",
		want: "var p = @\"C:\\// \"\"q\"\"\";\n",

		wantComments: cnt(1),
	},

	// ----------------------------------------------------------- Kotlin
	{
		name: "kotlin/raw string with slashes",
		lang: "kotlin",
		in:   "val s = \"\"\"// not\n/* a */ x\"\"\" // c\n",
		want: "val s = \"\"\"// not\n/* a */ x\"\"\"\n",

		wantComments: cnt(1),
	},

	// ------------------------------------------------------------ Swift
	{
		name: "swift/nested block comment",
		lang: "swift",
		in:   "let a = 1 /* x /* y */ z */\nlet s = \"//\" // c\n",
		want: "let a = 1\nlet s = \"//\"\n",

		wantComments: cnt(2),
	},

	// ------------------------------------------------------------- Ruby
	{
		name: "ruby/begin end block and END data",
		lang: "ruby",
		in: "=begin\ndoc\n=end\nx = 1 # c\ns = \"#{x} # not\"\n" +
			"__END__\n# data, not a comment\n",
		want: "x = 1\ns = \"#{x} # not\"\n" +
			"__END__\n# data, not a comment\n",

		wantComments: cnt(2),
		wantRemoved:  cnt(3),
	},

	// -------------------------------------------------------------- PHP
	{
		// Only <?php ... ?> regions are scanned; surrounding HTML is opaque.
		name: "php/only php regions are scanned",
		lang: "php",
		in: "<p># not php</p>\n<?php\n$a = 1; // c\n$b = \"# not\"; # d\n" +
			"/* e */\n?>\n<p>// still html</p>\n",
		want: "<p># not php</p>\n<?php\n$a = 1;\n$b = \"# not\";\n" +
			"?>\n<p>// still html</p>\n",

		wantComments: cnt(3),
	},

	// -------------------------------------------------------------- Lua
	{
		name: "lua/long bracket comment and dash string",
		lang: "lua",
		in: "local s = \"--\" -- c\n--[==[ long\n]] still ]=] comment\n]==]\n" +
			"local t = [[ -- not ]]\n",
		want: "local s = \"--\"\n" +
			"local t = [[ -- not ]]\n",

		wantComments: cnt(2),
		wantRemoved:  cnt(3),
	},

	// -------------------------------------------------------------- SQL
	{
		// /*+ ... */ is an optimizer hint, not a comment: Keep must match it.
		name: "sql/optimizer hint kept and dashes removed",
		lang: "sql",
		in:   "SELECT /*+ INDEX(t idx) */ a -- c\nFROM t; /* gone */\n",
		want: "SELECT /*+ INDEX(t idx) */ a\nFROM t;\n",

		wantComments: cnt(2),
	},
	{
		// StringRule.Doubled: '' inside '...' is one literal quote.
		name: "sql/doubled quotes inside a string",
		lang: "sql",
		in:   "SELECT 'it''s -- not' AS s; -- c\n",
		want: "SELECT 'it''s -- not' AS s;\n",

		wantComments: cnt(1),
	},

	// -------------------------------------------------------------- CSS
	{
		name: "css/block comment and url with slashes",
		lang: "css",
		// Only the line end is trimmed, so the comment sits after `}`.
		in:   "/* header */\na { background: url(http://x//y.png); } /* c */\n",
		want: "a { background: url(http://x//y.png); }\n",

		wantComments: cnt(2),
		wantRemoved:  cnt(1),
	},

	// ------------------------------------------------------------- HTML
	{
		name: "html/comment removed",
		lang: "html",
		in:   "<div>\n  <!-- gone -->\n  <p>a<!-- b -->c</p>\n</div>\n",
		// Comment bytes are deleted outright; no space is inserted.
		want: "<div>\n  <p>ac</p>\n</div>\n",

		wantComments: cnt(2),
		wantRemoved:  cnt(1),
	},

	// ------------------------------------------------------------- YAML
	{
		name: "yaml/block scalar body is opaque",
		lang: "yaml",
		in: "# gone\nsteps:\n  - run: |\n      echo a # keep\n      # keep too\n" +
			"  - name: b # c\nkey: \"# not\"\n",
		want: "steps:\n  - run: |\n      echo a # keep\n      # keep too\n" +
			"  - name: b\nkey: \"# not\"\n",

		wantComments: cnt(2),
		wantRemoved:  cnt(1),
	},

	// ------------------------------------------------------------- TOML
	{
		name: "toml/hash inside a string",
		lang: "toml",
		in:   "# gone\nkey = \"a # b\" # c\n",
		want: "key = \"a # b\"\n",

		wantComments: cnt(2),
		wantRemoved:  cnt(1),
	},

	// --------------------------------------------------------- Makefile
	{
		// MakeRecipe: a tab-indented recipe line is shell, left verbatim.
		name: "makefile/recipe line is untouched",
		lang: "makefile",
		in:   "# gone\nall:\n\techo # keep\n\t@echo b # keep too\nVAR = 1 # c\n",
		want: "all:\n\techo # keep\n\t@echo b # keep too\nVAR = 1\n",

		wantComments: cnt(2),
		wantRemoved:  cnt(1),
	},

	// ---------------------------------------------------------- Haskell
	{
		// `-->` is an operator: NotFollowedBy must cancel the comment.
		name: "haskell/arrow operator is not a comment",
		lang: "haskell",
		in:   "x --> y = x\nz = 1 -- c\n",
		want: "x --> y = x\nz = 1\n",

		wantComments: cnt(1),
	},
	{
		name: "haskell/nested block comment",
		lang: "haskell",
		in:   "{- {- inner -} outer -}\nmain = return () -- c\n",
		want: "main = return ()\n",

		wantComments: cnt(2),
		wantRemoved:  cnt(1),
	},

	// ------------------------------------------------------------- Lisp
	{
		name: "lisp/semicolon and block comments",
		lang: "lisp",
		in:   "; gone\n#| block\n   comment |#\n(print \"; not\") ; c\n",
		want: "(print \"; not\")\n",

		wantComments: cnt(3),
		wantRemoved:  cnt(3),
	},
}
