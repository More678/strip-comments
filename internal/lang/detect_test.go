package lang

import (
	"strings"
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	bom := string(rune(0xfeff))
	for _, tc := range []struct{ name, filename, input, language, by string }{
		{"filename wins", "code.PY", "#!/usr/bin/env node\npackage main\nfunc main() {}", "python", "filename"},
		{"registered special name", "Dockerfile", "", "dockerfile", "filename"},
		{"no content needed for known file", "src/unit.rs", "", "rust", "filename"},
		{"BOM and CRLF in script", "task", bom + "#!/usr/bin/env -S python3.12 -I\r\n", "python", "shebang"},
		{"shell interpreter", "", "#!/bin/bash -eu\nprintf hello", "shell", "shebang"},
		{"node interpreter", "", "#!/usr/bin/env node\n", "javascript", "shebang"},
		{"powershell interpreter", "", "#!/usr/bin/pwsh\n", "powershell", "shebang"},
		{"Go paste", "", "package main\nfunc main() {}\n", "go", "content"},
		{"Go after header", "helper.unknown", bom + "/* license */\r\npackage tools\r\n// package docs\r\nimport (\"fmt\")\r\n", "go", "content"},
		{"Python function", "", "# license\ndef total(n: int) -> int:\n    return n + 1", "python", "content"},
		{"Rust main", "", "fn main() { println!(\"hello\"); }", "rust", "content"},
		{"C generic", "", "_Static_assert(sizeof(char) == 1, \"size\");", "c", "content"},
		{"C++ header", "", "#include <iostream>\nint main() {}", "cpp", "content"},
		{"Java main", "", "public class Hello { public static void main(String[] args) {} }", "java", "content"},
		{"C sharp", "", "using System;\nclass App {}", "csharp", "content"},
		{"JavaScript module", "", "module.exports = {run};", "javascript", "content"},
		{"typed function", "", "export function run(x: number): number {return x;}", "typescript", "content"},
		{"typed interface", "", "interface Result { readonly count: number; }", "typescript", "content"},
		{"markup", "", "<!-- note -->\n<!doctype html><html></html>", "html", "content"},
		{"php", "", "<?php print 1;", "php", "content"},
		{"SQL", "", "-- summary\nSELECT count FROM totals;", "sql", "content"},
		{"stylesheet", "", ":root { --accent: #203540; }", "css", "content"},
		{"Ruby", "", "class Greeter\n  def greet\n    puts 'Hello'\n  end\nend", "ruby", "content"},
		{"Lua", "", "local function run()\nend", "lua", "content"},
		{"PowerShell", "", "Write-Host 'done'", "powershell", "content"},
		{"container recipe", "", "FROM scratch\nCOPY entry /entry", "dockerfile", "content"},
		{"crate attribute not a shebang", "", "#![allow(dead_code)]\nfn main() {}", "rust", "content"},
		{"shell options", "", "set -euo pipefail\nprintf ok", "shell", "content"},
		{"Kotlin", "", "package sample\n\nfun main() {}", "kotlin", "content"},
		{"Swift", "", "import Foundation\nstruct Item {}", "swift", "content"},
		{"unknown interpreter", "", "#!/usr/bin/python-backup\nfrom os import environ", "", ""},
		{"unknown env option", "", "#!/usr/bin/env -u python3\n", "", ""},
		{"comment only", "", "/*\npackage main\nfunc main() {}\n*/\n", "", ""},
		{"quoted sample", "", "\"\"\"\npackage main\nfunc main() {}\n\"\"\"", "", ""},
		{"ambiguous assignment", "", "x = 1\n", "", ""},
		{"ambiguous JS or TS", "", "const count = 1;", "", ""},
		{"ambiguous C or C++", "", "#include <stdio.h>\nint main(void) {return 0;}", "", ""},
		{"ambiguous package", "", "package example\n", "", ""},
		{"plain text", "", "Language tips: use package main in Go.", "", ""},
		{"binary", "", "package main\nfunc main() {}\x00", "", ""},
		{"bounded sample", "", strings.Repeat("// padding\n", detectionLimit/8) + "package main\nfunc main() {}", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			l, by := Detect(tc.filename, []byte(tc.input))
			name := ""
			if l != nil {
				name = l.Name
			}
			if name != tc.language || by != tc.by {
				t.Fatalf("got (%q, %q), want (%q, %q)", name, by, tc.language, tc.by)
			}
		})
	}
}
