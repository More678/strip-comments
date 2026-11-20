package lang

import (
	"bytes"
	"path"
	"regexp"
	"strings"
	"unicode/utf8"
)

const detectionLimit = 16 << 10

// Detect uses a registered filename first. For pasted text and unknown filenames it
// only trusts a recognized interpreter or a distinctive opening declaration.
// An ambiguous sample returns (nil, "") so the caller can ask for a language.
func Detect(filename string, src []byte) (*Language, string) {
	if l := ByFilename(path.Base(strings.ReplaceAll(filename, `\`, "/"))); l != nil {
		return l, "filename"
	}
	if len(src) > detectionLimit {
		// Don't take a partial UTF-8 rune or a partial declaration as evidence.
		end := bytes.LastIndexByte(src[:detectionLimit], '\n')
		if end < 0 {
			return nil, ""
		}
		src = src[:end+1]
	}
	if !utf8.Valid(src) || bytes.IndexByte(src, 0) != -1 {
		return nil, ""
	}
	s := string(bytes.TrimPrefix(src, []byte{0xef, 0xbb, 0xbf}))
	if rest, ok := strings.CutPrefix(s, "#!"); ok && !strings.HasPrefix(rest, "[") {
		line, _, _ := strings.Cut(rest, "\n")
		if l := interpreter(line); l != nil {
			return l, "shebang"
		}
		return nil, "" // code beneath an unknown interpreter may be embedded data
	}
	s = skipHeader(s)
	var match *Language
	for _, sig := range signatures {
		if sig.start.MatchString(s) {
			if match != nil {
				return nil, ""
			}
			match = ByName(sig.name)
		}
	}
	if match != nil {
		return match, "content"
	}
	return nil, ""
}

var pythonCommand = regexp.MustCompile(`^python([23](\.[0-9]+)*)?$`)

func interpreter(line string) *Language {
	args := strings.Fields(line)
	if len(args) == 0 {
		return nil
	}
	cmd := path.Base(args[0])
	if cmd == "env" {
		args = args[1:]
		for len(args) > 0 && (args[0] == "-S" || args[0] == "--split-string" || args[0] == "--" || args[0] == "-i") {
			args = args[1:]
		}
		if len(args) == 0 {
			return nil
		}
		cmd = path.Base(args[0])
	}
	if pythonCommand.MatchString(cmd) {
		return ByName("python")
	}
	switch cmd {
	case "sh", "bash", "dash", "ash", "ksh", "zsh":
		return ByName("shell")
	case "node", "nodejs":
		return ByName("javascript")
	case "tsx", "ts-node":
		return ByName("typescript")
	case "pwsh", "powershell":
		return ByName("powershell")
	case "Rscript":
		return ByName("r")
	case "ruby", "perl", "php", "lua", "julia", "swift", "elixir", "groovy":
		return ByName(cmd)
	}
	return nil
}

var directive = regexp.MustCompile(`^#[ \t]*(?:include|import|define|pragma|if|ifdef|ifndef)\b|^#!?\[`)

func skipHeader(s string) string {
	for {
		s = strings.TrimSpace(s)
		var end string
		switch {
		case strings.HasPrefix(s, "/*"):
			end = "*/"
		case strings.HasPrefix(s, "<!--"):
			end = "-->"
		case strings.HasPrefix(s, "//"), strings.HasPrefix(s, "-- "), strings.HasPrefix(s, "#") && !directive.MatchString(s):
			end = "\n"
		default:
			return s
		}
		_, rest, ok := strings.Cut(s, end)
		if !ok {
			return ""
		}
		s = rest
	}
}

// Anchor at the start: a comment, quoted example or embedded language farther
// down the document must not win merely because it contains a familiar keyword.
var signatures = []struct {
	name  string
	start *regexp.Regexp
}{
	{"go", regexp.MustCompile(`^package[ \t]+[\w]+[ \t]*;?[ \t]*(?://[^\n]*)?\r?\n\s*(?:(?://[^\n]*\n|/\*(?s:.*?)\*/)\s*)*(?:func\s+|type\s+\w+\s+(?:struct|interface)\b|import\s*[("\x60]|(?:var|const)\s+\w+\s*=)`)},
	{"python", regexp.MustCompile(`^(?:(?:async\s+)?def\s+\w+\s*\([^\n)]*\)\s*(?:->[^:\n]+)?:|from\s+[.\w]+\s+import\s+|import\s+(?:sys|os|asyncio|pathlib|typing)\b|if\s+__name__\s*==|class\s+\w+(?:\([^\n)]*\))?\s*:)`)},
	{"rust", regexp.MustCompile(`^(?:(?:pub(?:\(crate\))?\s+)?(?:async\s+)?fn\s+\w+\s*\([^\n)]*\)\s*(?:->|\{)|use\s+(?:std|crate|super|self)::|#!?\[(?:derive|cfg|allow)\b)`)},
	{"cpp", regexp.MustCompile(`^(?:#\s*include\s*<(?:iostream|vector|string|algorithm|memory|cstdio|cstdint|bits/stdc\+\+\.h)>|(?:using\s+namespace\s+std\s*;|std::\w+|template\s*<))`)},
	{"c", regexp.MustCompile(`^(?:_Static_assert\s*\(|_Generic\s*\(|_Atomic\s*\()`)},
	{"csharp", regexp.MustCompile(`^(?:(?:global\s+)?using\s+(?:System|Microsoft)(?:\.\w+)*\s*;|(?:System\.)?Console\.Write(?:Line)?\s*\()`)},
	{"java", regexp.MustCompile(`^(?:package\s+[\w.]+\s*;\s*(?:import\s+|public\s+(?:final\s+)?(?:class|interface|enum)\s+)|import\s+(?:static\s+)?(?:java|javax)\.|(?:public\s+)?(?:final\s+)?class\s+\w+\s*\{\s*public\s+static\s+void\s+main\s*\(|System\.out\.print(?:ln)?\s*\()`)},
	{"javascript", regexp.MustCompile(`^(?:["']use strict["']\s*;|(?:module\.exports\b|exports\.\w+\s*=)|(?:const|let|var)\s+[\w$]+\s*=\s*require\s*\(|console\.log\s*\(|(?:export\s+(?:default\s+)?)?(?:async\s+)?function\s+\w+\s*\([\w,$ \t]*\)\s*\{)`)},
	{"typescript", regexp.MustCompile(`^(?:export\s+(?:default\s+)?)?(?:(?:interface\s+\w+(?:\s+extends\s+\w+)?|type\s+\w+\s*=)\s*\{\s*(?:readonly\s+)?\w+\??\s*:|(?:let|var|const)\s+\w+\s*:\s*[\w{]|(?:async\s+)?function\s+\w+\s*\((?:[^)\n]*\w+\??\s*:\s*\w+|[^)\n]*\)\s*:\s*\w+))`)},
	{"html", regexp.MustCompile(`(?i)^<(?:!doctype\s+html\s*>|html(?:\s|>))`)},
	{"php", regexp.MustCompile(`^<\?php(?:\s|$)`)},
	{"sql", regexp.MustCompile(`(?i)^(?:create\s+table\s+(?:if\s+not\s+exists\s+)?[\w.]+\s*\(|select\s+[\w.*, \t]+\s+from\s+[\w.]+\s*(?:where\s+|;|$))`)},
	{"css", regexp.MustCompile(`^(?:@charset\s+["'][\w-]+["']\s*;|(?:body|html|:root|[.#][\w-]+)\s*\{\s*(?:--[\w-]+|[a-z][\w-]*)\s*:\s*[^;{}\n]+;)`)},
	{"ruby", regexp.MustCompile(`^(?:require_relative\s+["']|(?:class|module)\s+[A-Z][\w:]*(?:\s*<\s*[A-Z][\w:]*)?[ \t]*\r?\n\s*(?:def\s+|attr_(?:reader|writer|accessor)\b))`)},
	{"lua", regexp.MustCompile(`^local\s+function\s+\w+\s*\(`)},
	{"powershell", regexp.MustCompile(`(?i)^(?:\[CmdletBinding\(\)\]|Write-(?:Host|Output|Verbose)\s+)`)},
	{"dockerfile", regexp.MustCompile(`(?i)^FROM[ \t]+[^\s]+(?:[ \t]+as[ \t]+[\w-]+)?[ \t]*\r?\n\s*(?:RUN|COPY|WORKDIR|ENV|ENTRYPOINT|CMD)\s+`)},
	{"shell", regexp.MustCompile(`^set[ \t]+-[eux]*o[ \t]+pipefail\b`)},
	{"swift", regexp.MustCompile(`^import\s+(?:Foundation|SwiftUI|UIKit)\b`)},
	{"kotlin", regexp.MustCompile(`^(?:(?:package\s+[\w.]+\s+)?fun\s+main\s*\(|import\s+kotlin\.)`)},
}
