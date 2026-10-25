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
go install github.com/More678/Remove-All-Commet@latest
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
