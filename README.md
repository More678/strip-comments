# strip-comments

Remove every comment from the source files in a folder: bat, Go, Rust, C/C++,
Java, Python, shell, JavaScript and 50-odd other languages.

It lexes each file instead of matching patterns, so comment markers inside
strings, raw strings, heredocs and regex literals are left alone:
`"http://x"`, Go backticks, C++ `R"x( // )x"`, Rust `r#"//"#`, Java text
blocks, shell heredocs and `/\/\//g` all survive.

## Install

Download a binary from [Releases](../../releases). They are static and have no
runtime dependencies.

| file | platform |
| --- | --- |
| `strip-comments-linux-amd64` | Linux x86-64 |
| `strip-comments-linux-arm64` | Linux ARM64 |
| `strip-comments-windows-amd64.exe` | Windows x86-64 |
| `strip-comments-windows-arm64.exe` | Windows ARM64 |

Or build from source with Go 1.22+:

```sh
go install github.com/More678/strip-comments@latest
```

## Usage

```sh
strip-comments [flags] [path ...]    # path defaults to .
```

**The default is a dry run.** It reports what would change and writes nothing.
Pass `-w` to modify files.

```sh
strip-comments src            # preview
strip-comments -w src         # apply
strip-comments -w -backup .   # apply, keeping <file>.bak copies
strip-comments -stdout a.go   # print one stripped file
```

| flag | meaning |
| --- | --- |
| `-w` | write changes (default: dry run) |
| `-backup` | save `<file>.bak` before writing |
| `-stdout` | print the stripped content of a single file |
| `-ext .go,.rs` | only these extensions |
| `-lang NAME` | force a language (see `-list`) |
| `-exclude a,b` | extra dir/file names or globs to skip |
| `-keep-doc` | keep doc comments (`///`, `/** */`) |
| `-keep-license` | keep the comment header at the top of each file |
| `-keep-blank` | keep lines that become empty instead of deleting them |
| `-force` | strip even files with an unterminated string or comment |
| `-risky` | include languages with heuristic support (perl, vim, matlab, asm) |
| `-j N` | worker count (default: CPU count) |
| `-list` | list supported languages and extensions |
| `-replace` | also replace terminology (see below; off by default) |
| `-terms FILE` | use this term table instead of the built-in one |
| `-terms-add FILE` | add a term table that takes precedence over the built-in one |
| `-terms-in-comments` | replace terms only inside comments |
| `-keep-comments` | replace terms without removing any comment |
| `-list-terms` | print the term table in effect and exit |

## What it keeps on purpose

Some comments change how code builds or behaves, so they are never removed:

- Go `//go:build`, `//go:generate`, `//go:embed`, and the cgo preamble above
  `import "C"`
- Python `# -*- coding: -*-`, `# type:`, `# noqa`, `# pragma`
- a `#!` shebang on line 1
- SQL optimizer hints such as `/*+ INDEX(t) */`
- linter and formatter directives: `// NOLINT`, `// NOSONAR`, `@ts-ignore`,
  `eslint-disable`, `clang-format off`, `shellcheck`, `# rubocop:` and similar

A file with an unterminated string or block comment is skipped and reported as
`SKIP`, because its real extent can't be known. Pass `-force` to strip it
anyway.

Line endings, a UTF-8 BOM and the file mode are preserved. A file with mixed
CRLF and LF endings that does lose a comment comes back in its majority style.
Files with no comments to remove are never rewritten.

## Terminology replacement

`-replace` rewrites terms in **both code and comments**, so a codebase can be
restated in the vocabulary of threat detection and compliance assessment. It is
off unless you ask for it.

```sh
strip-comments -replace .                     # dry run: list every hit
strip-comments -replace -w -backup .          # apply, keeping .bak copies
strip-comments -replace -keep-comments -w .   # only rewrite terms, keep comments
strip-comments -replace -terms-in-comments .  # leave code alone
strip-comments -list-terms                    # print the table in effect
```

A dry run lists what would change, which is how you review a table before
applying it:

```
  a.go  go  -2 comments, -1 lines, 3 terms
      line 3  渗透测试 -> 威胁检测  (comment)
      line 4  pentest -> threat_detection  (code)
      line 4  漏洞利用 -> 风险验证  (comment)
```

### Term tables

TSV, one rule per line, with an optional third column of flags:

```
# comment lines and blank lines are skipped
渗透	威胁检测
pentest	threat detection
Exploit	risk validation	nocase
backdoor	anomalous channel	noword,ci
```

Flags are `word` / `noword` (require ASCII word boundaries) and `ci` / `nocase`
(match case-insensitively and carry the original casing to the replacement).
Omit them and they are inferred the way the built-in table is written: a term
with ASCII edges gets both, a CJK term gets neither, because Chinese text has
no word characters to delimit.

A `.json` file is read as `[{"from":"渗透","to":"威胁检测","word":false,"ci":false}]`.

The built-in table, which `-list-terms` prints in match order:

| from | to | | from | to |
| --- | --- | --- | --- | --- |
| 渗透测试 / 渗透攻击 / 渗透 | 威胁检测 | | `penetration test` / `pentest` | `threat detection` |
| 漏洞利用 | 风险验证 | | `exploit` | `risk validation` |
| 攻击载荷 | 检测样本 | | `payload` | `detection sample` |
| 攻击面 | 暴露面 | | `attack surface` | `exposure surface` |
| 提权 | 权限核查 | | `privilege escalation` | `privilege review` |
| 后门 | 异常通道 | | `backdoor` | `anomalous channel` |
| 木马 | 可疑程序 | | `bruteforce` | `credential strength check` |
| 暴力破解 | 口令强度核查 | | | |
| 绕过 | 合规偏差 | | | |
| 等级保护 / 等保测评 | 等保分析 | | | |

`-terms` replaces the built-in table outright. `-terms-add` keeps it and puts
your rules first, so your version of a term wins. Longer terms always match
before shorter ones, so `渗透测试` beats `渗透` and you never get
`威胁检测测试`.

### Three things to know before using it

**It is not a refactoring tool.** Renaming an identifier is consistent only
across the files this run touched. A reference from a config file, another
language, or an outside caller does not follow. Run it on a whole project, and
check what else names the symbols you are about to rename.

**Multi-word replacements become identifiers in code.** `pentest := 1` would
not parse as `threat detection := 1`, so spaces fold to underscores outside
comments: `threat_detection := 1`. Casing carries over, so Go's exported
`Pentest` stays exported as `Threat_detection`. Inside comments the spaces are
kept, because a comment is prose.

**String literals get the identifier form too.** The replacer is given comment
ranges and nothing else, so it cannot tell `"pentest"` from an identifier and
writes `"threat_detection"`. The file still compiles; the string just reads
less naturally.

Import paths, `#include` lines and URLs are never rewritten, since those are
resolved by a toolchain rather than read by a person.

If a replacement would leave an unterminated string or comment behind — a
replacement text containing a quote, say — the file is reported as a `SKIP`
and left untouched even under `-w`:

```
SKIP b.go: after term replacement: unterminated string at line 3
```

## Deliberately unsupported

- `.s` / `.S`: GAS uses `;` as a statement separator, so stripping it would
  delete code (`rep; movsb`). NASM/MASM `.asm` is supported.
- `.pp`: shared by Pascal and Puppet, and Pascal's `{ }` comments would delete
  Puppet resource bodies.
- `.m` is treated as Objective-C. Use `-lang matlab -risky` for MATLAB.

## Build

```sh
bash build.sh       # or: powershell ./build.ps1
```

Writes the four binaries plus `SHA256SUMS` to `dist/`. Pushing a `v*` tag runs
the tests, builds, smoke-tests the Linux binary and publishes a GitHub release.
