# strip-comments

Remove every comment from the source files in a folder: bat, Go, Rust, C/C++,
Java, Python, shell, JavaScript and 50-odd other languages.

It lexes each file instead of matching patterns, so comment markers inside
strings, raw strings, heredocs and regex literals are left alone:
`"http://x"`, Go backticks, C++ `R"x( // )x"`, Rust `r#"//"#`, Java text
blocks, shell heredocs and `/\/\//g` all survive.

## Go, Rust and C#

These languages are automatically recognised; no runtime or compiler is needed
for comment removal. Use `-lang csharp` (not `c#`) when selecting C# explicitly.

| `-lang` | Extension | Comments / protected literals |
| --- | --- | --- |
| `go` | `.go` | `//`, `/* */`; quoted strings, backticks and runes; existing Go directives and cgo preambles are preserved |
| `rust` | `.rs` | `//`, nested `/* */`; strings, characters, lifetimes, `r`, `br` and `cr` raw strings with hash delimiters |
| `csharp` | `.cs` | `//`, `/* */`; ordinary/verbatim strings, variable-length raw quote delimiters and interpolated strings |

C# interpolation expressions are scanned separately: real expression comments
are removed, while nested strings and format text remain intact. Rust/C# doc
comments can be retained with `-keep-doc`. In Rust, `////`, `/*** ... */` and
`/**/` are ordinary comments and are still removed with that option. Go tool
directives remain protected independently of it. C# format clauses cannot cross
the enclosing string's closing quote delimiter; malformed clauses are skipped
rather than consuming subsequent code. Escaped quotes, doubled verbatim quotes
and shorter quote runs inside raw format text remain protected.

```sh
strip-comments -diff -ext .go,.rs,.cs src
strip-comments -w -backup -ext .go,.rs,.cs src
strip-comments -lang csharp -keep-doc -stdout Example.cs
```

These are lexical rules, not full Rust or C# compiler validation. C# string
nesting is limited to 128 levels; unterminated or rejected string structures
leave the file unchanged unless `-force` is used. Review the diff before writing;
the shared write-time verifier cannot guarantee semantic equivalence.

Literal references: [Rust tokens](https://doc.rust-lang.org/reference/tokens.html),
[Rust comments](https://doc.rust-lang.org/reference/comments.html),
[C# raw strings](https://learn.microsoft.com/en-us/dotnet/csharp/language-reference/tokens/raw-string)
and [C# interpolation](https://learn.microsoft.com/en-us/dotnet/csharp/language-reference/tokens/interpolated).

## Additional language support

These languages use the existing lexer; no extra runtime or compiler is needed:

| `-lang` | Automatically recognised files | Comments / protected literals |
| --- | --- | --- |
| `gdscript` | `.gd` | `#`, `##` documentation; quoted and triple-quoted strings |
| `gdshader` | `.gdshader`, `.gdshaderinc` | `//`, `/* */`, `/** */` documentation; quoted include paths |
| `wgsl` | `.wgsl` | `//` and nested `/* */` |
| `meson` | `meson.build`, `meson.options`, `meson_options.txt` | `#`; single-quoted strings and raw triple-single-quoted strings |
| `starlark` | `.bzl`, `.bazel`, `BUILD`, `WORKSPACE` | `#`; single-, double- and triple-quoted strings |

`BUILD.bazel`, `WORKSPACE.bazel` and `MODULE.bazel` are recognised by extension.
Filename matching follows the existing case-insensitive lookup. `-keep-doc`
preserves GDScript `##` and Godot Shader `/** */` comments. Triple-quoted text
remains a string, not a removable block comment. An unterminated multiline
literal or block comment is skipped unless `-force` is passed.

```sh
strip-comments -diff -ext .gd,.gdshader,.gdshaderinc,.wgsl src
strip-comments -w -backup meson.build BUILD rules.bzl
```

These are lexical rules, not full language parsers. Third-party WGSL
preprocessors and embedded languages are outside this support. Ada is not
registered: its apostrophe syntax needs language-specific handling rather than
the existing C-style character lexer. Always review a dry run before writing.

Rule references: [GDScript](https://docs.godotengine.org/en/stable/tutorials/scripting/gdscript/gdscript_basics.html),
[Godot shaders](https://docs.godotengine.org/en/stable/tutorials/shaders/shader_reference/shading_language.html),
[WGSL](https://www.w3.org/TR/WGSL/), [Meson](https://mesonbuild.com/Syntax.html),
and [Starlark](https://github.com/bazelbuild/starlark/blob/master/spec.md).

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

## Local WebUI

From a checkout with Go 1.22+, run `go run . -web` (or `strip-comments -web` with a binary built from this version). Open the `http://127.0.0.1:<port>/` URL printed in the terminal. The port is assigned automatically; **Ctrl+C** stops the server. The interface is embedded in the binary and works offline without Node.js or a separate frontend build.

Paste source text or select one UTF-8 file (up to 256 KiB). Automatic detection checks the filename, then the interpreter in a shebang (for example `#!/usr/bin/env python3`), then distinctive opening code in the first 16 KiB. It recognizes common signatures such as a Go package with a function, a Rust `fn`, a typed TypeScript declaration, or an HTML doctype. The result shows which signal matched. Short or ambiguous snippets such as `x = 1` need a manual language selection; you can always override detection. The browser sends source to the local service only when you request a preview.

Review the output, counts and unified diff. For large inputs the line diff may be omitted; you can still copy the verified result or download a separate file. The server works in memory: it does not accept filesystem paths, traverse folders or overwrite originals. Known credential filenames are rejected; do not paste secrets.

The single-file `-stdout` mode also detects an unrecognized filename from its shebang or source. For example, `strip-comments -stdout deploy-helper` can preview an extensionless shell script. Recursive directory processing still uses the existing filename rules and does not start reading previously unknown files.

The optional browser glossary is **off by default** and contains only the meaning-preserving audit rules, not the CLI's historical mappings. You can restrict replacements to comments and keep comments intact. Renaming terms does not fix vulnerabilities or establish compliance. For in-place writes and restorable backups, use the CLI flags below.

## Usage

```sh
strip-comments [flags] [path ...]    # path defaults to .
```

**The default is a dry run.** It reports what would change and writes nothing.
Pass `-w` to modify files; each changed file gets a same-directory `<file>.bak`
containing its original bytes before either comment removal or term replacement.

```sh
strip-comments src                # preview
strip-comments -w src             # apply, keeping <file>.bak copies by default
strip-comments -w -backup .       # same behavior; legacy flag still works
strip-comments -restore a.go      # preview restoring a.go from a.go.bak
strip-comments -restore -w a.go    # restore; keep current bytes in a.go.restore.bak
strip-comments -stdout a.go       # print one stripped file
```

| flag | meaning |
| --- | --- |
| `-w` | write changes with backups (default: dry run) |
| `-backup` | enabled by default; `-backup=false` explicitly opts out for ordinary rewrites |
| `-restore` | restore named source files from `<file>.bak`; preview unless `-w` |
| `-web` | serve the local, single-file preview UI; use without other flags or paths |
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
| `-diff` | print a unified diff of every change, write nothing |
| `-check` | exit non-zero if any file would change, write nothing (for CI) |
| `-no-verify` | skip the structural check that runs before writing |
| `-allow-secrets` | process files that look like credentials |

## Backups and restoration

`-w` creates `<file>.bak` beside each file it actually changes. The backup is a
byte-for-byte copy, including BOM, CRLF or mixed line endings, taken before
comment removal and/or terminology replacement. Dry runs, `-stdout`, `-diff`,
`-check`, skipped files and unchanged files create no backup.

**Existing backups are never replaced.** If `<file>.bak` already exists, a
further rewrite of that file fails without changing it. A no-op rerun still
succeeds. Move the old backup somewhere safe before beginning another round;
`-backup=false` explicitly disables backup creation for ordinary writes, leaves
any old backup alone, and makes the new change unrecoverable through that backup.
Backup creation failures also leave the source untouched.

```sh
strip-comments -restore src/a.go src/b.rs     # preview only
strip-comments -restore -w src/a.go src/b.rs  # restore original bytes
```

Pass **source file paths**, not directories or `.bak` paths. Restoration does
not decode, strip or replace anything: it copies `<file>.bak` back verbatim.
Before doing so it saves the current file, including any later manual edits,
to `<file>.restore.bak`. Both backups remain afterward; neither is deleted or
overwritten. An existing `.restore.bak` blocks a changed restore, including its
preview. Move that copy somewhere safe first. A file already equal to `.bak`
is a successful no-op. Missing sources/backups, directories and symlinks are
rejected. `-restore` accepts only `-w`/`-write`, `-q` and `-allow-secrets` alongside
it; transformation, backup and traversal flags are usage errors.

Backups use a synced temporary file and an exclusive hard link in the same
directory. Filesystems without hard-link support (for example FAT/exFAT) fail
closed; use a supporting filesystem for reversible writes. This is a one-round,
per-file backup, not version control or an all-files transaction. A failed
source write may leave a valid backup to inspect before retrying. Stop editors
and other writers while rewriting/restoring: files are not locked against
concurrent edits. File permission bits are retained, but timestamps, ownership,
ACLs and extended attributes are not restored. Keep sensitive backups private,
especially when opting into `-allow-secrets`.

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
| 暴力破解 / 口令爆破 | 口令强度核查 | | `credential stuffing` | `credential reuse detection` |
| 绕过 | 合规偏差 | | `lateral movement` | `lateral traffic analysis` |
| 等级保护 / 等保测评 | 等保分析 | | `data exfiltration` | `data egress detection` |
| 内网穿透 | 内网连通性检测 | | `webshell` | `suspicious script` |
| 横向移动 | 横向流量分析 | | `reverse shell` / `C2` | `outbound anomaly` |
| 权限维持 | 持久化检测 | | `phishing` | `impersonation detection` |
| 数据窃取 | 数据外泄检测 | | `sniffer` | `traffic monitor` |
| 反弹shell / 命令与控制 | 异常外联检测 | | `MITM` | `link hijack detection` |
| 免杀 | 检测规避分析 | | | |
| 嗅探 | 流量监测 | | | |
| 中间人攻击 | 链路劫持检测 | | | |
| 社会工程学 | 安全意识评估 | | | |
| 钓鱼邮件 / 钓鱼网站 | 可疑邮件 / 仿冒网站 | | | |
| 撞库 | 凭据复用检测 | | | |

Vulnerability classes, recon and tooling:

| from | to | | from | to |
| --- | --- | --- | --- | --- |
| 命令注入 | 命令执行风险核查 | | `RCE` | `remote execution risk` |
| 远程代码执行 | 远程执行风险核查 | | `SSRF` | `server request review` |
| 目录遍历漏洞 / 路径穿越 | 路径访问核查 | | `LFI` | `file inclusion review` |
| 反序列化漏洞 | 反序列化风险 | | `privilege bypass` | `access control review` |
| 越权访问 | 权限越界核查 | | `port scan` / `port scanning` | `open port review` |
| 未授权访问 | 访问控制核查 | | `port scanner` | `open port auditor` |
| 弱口令 | 口令强度核查 | | `vulnerability scan` | `risk scan` |
| 信息泄露 | 信息暴露核查 | | `vulnerability scanner` | `risk scanner` |
| 提权漏洞 | 权限风险 | | `zero-day` / `0day` | `undisclosed risk` |
| 供应链攻击 | 供应链风险评估 | | `red teaming` | `adversarial assessment` |
| 端口扫描 | 端口开放核查 | | `exploit kit` | `risk validation toolkit` |
| 漏洞扫描 | 风险扫描 | | `rootkit` | `kernel-level anomaly` |
| 资产探测 | 资产梳理 | | `keylogger` | `input monitor` |
| 零日漏洞 | 未公开风险 | | | |
| 漏洞挖掘 | 风险发现 | | | |
| 攻防演练 | 安全演练 | | | |
| 红队测试 | 对抗性评估 | | | |

Ambiguous everyday words are only matched as compounds, since 爆破 also means
demolition, 社工 a social worker, 中间人 a broker, 钓鱼 fishing, 目录遍历 an
ordinary directory walk and 红队 a sports side. Acronyms (`C2`, `MITM`, `RCE`,
`SSRF`, `LFI`) are matched case-sensitively and as whole words, so `0xC2`,
`C2H5OH`, `FORCE`, `SOURCE` and a `c2` or `rce` variable are left alone.

Additional authentication and session review terms:

| from | to |
| --- | --- |
| 密码喷洒 / 口令喷洒 | 批量认证风险核查 |
| 浏览器会话劫持 | 浏览器会话完整性核查 |
| 远程会话劫持 | 远程会话完整性核查 |
| `password spraying` / `password spray` | `bulk authentication review` |
| `browser session hijacking` | `browser session integrity review` |
| `remote service session hijacking` | `remote session integrity review` |
| 黄金票据攻击 | Kerberos票据伪造风险核查 |
| 票据传递攻击 | Kerberos票据复用风险核查 |
| 凭据转储攻击 | 认证材料暴露风险核查 |
| `golden ticket attack` | `Kerberos ticket forgery risk review` |
| `pass the ticket` | `Kerberos ticket reuse risk review` |
| `OS credential dumping` | `credential exposure risk assessment` |
| DNS缓存投毒 | 名称解析缓存完整性风险核查 |
| ARP缓存投毒 | 地址映射完整性风险核查 |
| `DNS cache poisoning` | `name resolution cache integrity risk review` |
| `ARP cache poisoning` | `address mapping integrity risk review` |
| 会话固定攻击 | 会话标识轮换风险核查 |
| 点击劫持攻击 | 界面交互劫持风险核查 |
| `session fixation attack` | `session identifier rotation risk review` |
| `clickjacking attack` | `interface interaction hijack risk review` |
| HTTP请求走私 | HTTP请求解析差异风险核查 |
| HTTP响应拆分 | HTTP响应头换行处理风险核查 |
| `HTTP request smuggling` | `HTTP request parsing inconsistency risk review` |
| `HTTP response splitting` | `HTTP response header newline handling risk review` |
| Web缓存投毒 | Web缓存响应完整性风险核查 |
| Web缓存欺骗 | Web私有响应错误缓存风险核查 |
| `web cache poisoning` | `web cache response integrity risk review` |
| `web cache deception` | `web private response caching risk review` |
| 原型污染攻击 | 对象原型篡改风险核查 |
| 批量赋值漏洞 | 非预期字段写入风险核查 |
| `prototype pollution attack` | `object prototype tampering risk review` |
| `mass assignment vulnerability` | `unintended field modification risk review` |
| 依赖混淆攻击 | 依赖来源解析风险核查 |
| 软件包拼写抢注攻击 | 软件包名称仿冒风险核查 |
| `dependency confusion attack` | `dependency source resolution risk review` |
| `package typosquatting attack` | `package name impersonation risk review` |
| XML外部实体攻击 | XML外部实体解析风险核查 |
| XML递归实体展开攻击 | XML递归实体展开资源耗尽风险核查 |
| `XML external entity attack` | `XML external entity resolution risk review` |
| `XML recursive entity expansion attack` | `XML recursive entity expansion resource exhaustion risk review` |
| 服务端模板注入攻击 | 服务端模板表达式注入风险核查 |
| XPath注入攻击 | XPath查询表达式注入风险核查 |
| `server-side template injection attack` | `server-side template expression injection risk review` |
| `XPath injection attack` | `XPath query expression injection risk review` |
| LDAP注入攻击 | LDAP查询注入风险核查 |
| 日志伪造攻击 | 日志记录伪造风险核查 |
| `LDAP injection attack` | `LDAP query injection risk review` |
| `log forging attack` | `log record forgery risk review` |
| 正则表达式拒绝服务攻击 | 正则表达式计算资源耗尽风险核查 |
| `regular expression denial of service attack` | `regular expression computational resource exhaustion risk review` |
| HTTP参数污染攻击 | HTTP参数污染与解释风险核查 |
| `HTTP parameter pollution attack` | `HTTP parameter pollution interpretation risk review` |
| CSV公式注入攻击 | CSV单元格公式注入风险核查 |
| `CSV formula injection attack` | `CSV cell formula injection risk review` |
| 开放重定向攻击 | 非可信目标重定向风险核查 |
| `open redirect attack` | `untrusted destination redirect risk review` |
| `DOM clobbering attack` | `DOM named property clobbering risk review` |
| HTTP主机头注入攻击 | HTTP主机头注入风险核查 |
| `HTTP host header injection attack` | `HTTP host header injection risk review` |
| NoSQL注入攻击 | NoSQL查询逻辑注入风险核查 |
| `NoSQL injection attack` | `NoSQL query logic injection risk review` |
| 跨站WebSocket劫持攻击 | 跨站WebSocket会话劫持风险核查 |
| `cross-site WebSocket hijacking attack` | `cross-site WebSocket session hijacking risk review` |
| JWT算法混淆攻击 | JWT验签算法混淆风险核查 |
| `JWT algorithm confusion attack` | `JWT verification algorithm confusion risk review` |
| `XML signature wrapping attack` | `XML signature wrapping risk review` |
| `Zip Slip attack` | `archive extraction path traversal risk review` |
| OAuth2.0授权码注入攻击 | OAuth2.0授权码注入风险核查 |
| `OAuth 2.0 authorization code injection attack` | `OAuth 2.0 authorization code injection risk review` |
| `OAuth mix-up attack` | `OAuth authorization server mix-up risk review` |
| 外部可控格式字符串漏洞 | 外部可控格式字符串风险核查 |
| `externally controlled format string vulnerability` | `externally controlled format string risk review` |
| DNS重绑定攻击 | DNS重绑定信任边界风险核查 |
| `DNS rebinding attack` | `DNS rebinding trust boundary risk review` |
| `HTTP verb tampering attack` | `HTTP verb tampering access control risk review` |
| TLS协议降级攻击 | TLS协议降级风险核查 |
| `TLS protocol downgrade attack` | `TLS protocol downgrade risk review` |
| HTML注入攻击 | HTML标记注入风险核查 |
| `HTML injection attack` | `HTML markup injection risk review` |
| 内存释放后使用漏洞 | 内存释放后使用风险核查 |
| 内存重复释放漏洞 | 内存重复释放风险核查 |
| `use-after-free vulnerability` | `use-after-free risk review` |
| `double free vulnerability` | `double free risk review` |
| 内存越界读取漏洞 | 内存越界读取风险核查 |
| 内存越界写入漏洞 | 内存越界写入风险核查 |
| `out-of-bounds read vulnerability` | `out-of-bounds read risk review` |
| `out-of-bounds write vulnerability` | `out-of-bounds write risk review` |
| `TOCTOU race condition vulnerability` | `TOCTOU check-use race risk review` |
| 整数溢出漏洞 | 整数溢出风险核查 |
| 整数下溢漏洞 | 整数下溢风险核查 |
| `integer overflow vulnerability` | `integer overflow risk review` |
| `integer underflow vulnerability` | `integer underflow risk review` |
| 类型混淆漏洞 | 类型混淆风险核查 |
| `type confusion vulnerability` | `type confusion risk review` |
| 空指针解引用漏洞 | 空指针解引用风险核查 |
| `null pointer dereference vulnerability` | `null pointer dereference risk review` |
| 不受控递归漏洞 | 不受控递归资源耗尽风险核查 |
| `uncontrolled recursion vulnerability` | `uncontrolled recursion resource exhaustion risk review` |
| 未初始化变量使用漏洞 | 未初始化变量使用风险核查 |
| `uninitialized variable use vulnerability` | `uninitialized variable use risk review` |
| OAuth刷新令牌重放攻击 | OAuth刷新令牌重放风险核查 |
| `OAuth refresh token replay attack` | `OAuth refresh token replay risk review` |
| `padding oracle attack` | `padding oracle risk review` |
| `reverse tabnabbing attack` | `reverse tabnabbing risk review` |
| 跨站脚本包含攻击 | 跨站脚本包含数据泄露风险核查 |
| `cross site script inclusion attack` | `cross site script inclusion data exposure risk review` |
| CSS注入攻击 | CSS样式注入风险核查 |
| `CSS injection attack` | `CSS style injection risk review` |
| SSI注入攻击 | SSI服务端包含指令注入风险核查 |
| `SSI injection attack` | `SSI server-side include directive injection risk review` |
| `IDN homograph attack` | `IDN homograph domain spoofing risk review` |
| 操作系统命令参数注入攻击 | 操作系统命令参数注入风险核查 |
| `OS command argument injection attack` | `OS command argument injection risk review` |
| `session puzzling attack` | `session variable overloading risk review` |
| 表达式语言注入攻击 | 表达式语言求值注入风险核查 |
| `expression language injection attack` | `expression language evaluation injection risk review` |
| `reflected file download attack` | `reflected file download risk review` |
| CORS错误配置漏洞 | CORS错误配置风险核查 |
| `CORS misconfiguration vulnerability` | `CORS misconfiguration risk review` |
| XQuery注入攻击 | XQuery查询结构注入风险核查 |
| `XQuery injection attack` | `XQuery query structure injection risk review` |
| PHP对象注入攻击 | PHP反序列化对象注入风险核查 |
| `PHP object injection attack` | `PHP deserialization object injection risk review` |
| TLS证书主机名验证漏洞 | TLS证书主机名验证风险核查 |
| `TLS certificate hostname validation vulnerability` | `TLS certificate hostname validation risk review` |
| 外部可控反射漏洞 | 外部可控反射代码选择风险核查 |
| `externally controlled reflection vulnerability` | `externally controlled reflection code selection risk review` |
| 不可信搜索路径漏洞 | 不可信搜索路径资源选择风险核查 |
| `untrusted search path vulnerability` | `untrusted search path resource selection risk review` |
| `cryptographically weak PRNG vulnerability` | `cryptographically weak PRNG risk review` |
| postMessage来源验证漏洞 | postMessage来源验证风险核查 |
| `postMessage origin validation vulnerability` | `postMessage origin validation risk review` |
| 不受限文件上传漏洞 | 不受限文件上传风险核查 |
| `unrestricted file upload vulnerability` | `unrestricted file upload risk review` |
| `signed to unsigned conversion vulnerability` | `signed to unsigned conversion risk review` |
| `double-checked locking vulnerability` | `double-checked locking initialization visibility risk review` |
| 未检查返回值漏洞 | 未检查返回值风险核查 |
| `unchecked return value vulnerability` | `unchecked return value risk review` |
| 主机型入侵检测系统 | 基于主机的入侵检测系统 |
| 网络型入侵检测系统 | 基于网络的入侵检测系统 |
| `host-based IDS` | `host-based intrusion detection system` |
| `network-based IDS` | `network intrusion detection system` |
| `audit record retention period` | `retention period for audit records` |
| `timestamps for audit records` | `time stamps for audit records` |
| `audit event result` | `outcome of the audited event` |
| `allocate audit log storage capacity` | `allocate storage capacity for audit logs` |
| `the disk with the audit log is almost full` | `the disk holding the audit log is nearly full` |
| `rotation of audit log files` | `audit log file rotation` |
| `compression of rotated audit log files` | `compression of audit log files that have been rotated` |
| `verify audit log file integrity` | `verify the integrity of the audit log file` |
| `provides read-only access to audit records` | `allows read-only access to audit records` |
| `original time ordering of audit records` | `original chronological order of audit records` |
| `the original content of audit records is preserved` | `the original content of audit records is kept intact` |
| `generate on-demand audit reports` | `generate audit reports on demand` |
| `search and filter audit records by event type` | `search and filter audit log entries by event type` |
| `find all audit records within a time range` | `find all audit records in a specified time range` |
| `search audit records by actor` | `query audit records by actor` |
| `get audit logs containing a given target resource` | `retrieve audit logs containing a given target resource` |
| `Get audit logs by correlation ID` | `Retrieve audit logs by correlation ID` |
| `get the next page of audit logs` | `retrieve the next page of audit logs` |
| `maximum number of audit logs to return per page` | `maximum number of audit logs to return on any one page` |
| `export the results of an audit log search to a CSV file` | `export audit log search results to a CSV file` |
| `create a saved audit log filter` | `save a new audit log filter` |
| `clear all audit log filters and reload` | `clear every audit log filter and reload` |
| `when you expand an audit log entry, you see exactly what changed` | `expanding an audit log entry shows you exactly what changed` |
| `the audit log shows the old and new values` | `the audit log shows the before and after values` |
| `Copy Audit Event ID` | `Copy Audit Event Identifier` |
| `timestamps in the audit logs table and detail view are displayed in your local time zone` | `timestamps in the audit logs table and detail view are shown in your local time zone` |
| `The time the event described by the log entry occurred` | `When the event described by the log entry occurred` |
| `The time the log entry was received by Logging` | `When the log entry was received by Logging` |
| `no audit logs match your filters` | `no audit log entries match your filters` |
| `Loading audit log entries` | `Loading audit log records` |
| `Retry loading audit logs` | `Try loading audit logs again` |
| `audit log query timed out` | `audit log query did not return in time` |
| `Cancel a running CloudTrail Lake query` | `Request cancellation of a running CloudTrail Lake query` |
| `Get the status of a CloudTrail Lake query` | `Check the status of a CloudTrail Lake query` |
| `If the query results were delivered to an S3 bucket, the response also provides the S3 URI and the delivery status` | `If the query results were delivered to an S3 bucket, the response also includes the S3 URI and the delivery status` |
| `The number of events that the query scanned in the event data store` | `How many events the query scanned in the event data store` |
| `The total bytes that the query scanned in the event data store` | `The total number of bytes that the query scanned in the event data store` |
| `The query's run time, in milliseconds` | `The duration of query execution in milliseconds` |
| `Lists matching query results, and shows query ID, status, and creation time of each query` | `Lists matching query results and shows the ID, status, and creation time for each query` |
| `The status of queries that you want to return in results` | `The status to match when listing queries` |
| `to bound a ListQueries request, and limit its results to only those queries run within a specified time period` | `to limit ListQueries results to queries run within the specified time range` |
| `The ARN (or the ID suffix of the ARN) of an event data store on which queries were run` | `The ARN or its ID suffix for the event data store where the queries ran` |
| `NextPage retrieves the next ListQueries page` | `NextPage retrieves the next page from ListQueries` |
| `The maximum number of queries to show on a page` | `The maximum number of queries to list per page` |
| `The error message returned if a query failed` | `The error message returned if the query itself failed` |
| `invalid audit log page token` | `invalid audit log pagination token` |
| `Information about whether the event is a write event or a read event` | `Whether the event is classified as a write or a read` |
| `failed to lookup CloudTrail events by event name` | `failed to look up CloudTrail events by event name` |
| `Retrieves a specific audit event by its ID` | `Retrieves a specific audit event by its identifier` |
| `Parameter for filtering audit logs by resource type` | `Parameter to filter audit logs by resource type` |
| `Get audit logs by source IP` | `Retrieve audit logs by source IP` |
| `Filter OpenShift API server audit logs by user agent` | `Filter OpenShift API server audit log entries by user agent` |
| `Find audit logs by HTTP status code` | `Find audit log entries by HTTP status code` |
| `Find audit logs by HTTP method` | `Find audit log entries by HTTP method` |
| `Find audit logs by request path` | `Find audit log entries by request path` |
| `auth_decision: Filter by authorization decision` | `auth_decision: Authorization decision to match` |
| `Filter by minimum request duration in milliseconds` | `Filter by minimum request duration in ms` |
| `Filter by maximum request duration in milliseconds` | `Filter by maximum request duration in ms` |
| `This filters all TLS connections with 1.1 and below` | `This filters all TLS connections using version 1.1 or lower` |

The [AWS CloudTrail Lake sample](https://github.com/aws-samples/cloud-trail-lake-query-samples/blob/f1befa2825a249af0e5cba019697f0d08e6e0ee5/tls-unsupported-version.sql#L3-L29) filters on the recorded `tlsDetails.tlsVersion` before grouping and counting events. Both wordings include version 1.1. Neither describes the versions a client supports or the minimum a server accepts. The rule changes the description, not the query, and leaves missing versions missing.

The [audit API](https://github.com/ganjianping/gjpb-api-spring-web-public/blob/db2680196d5df62da7717f763189560f63c252c7/resources/docs/dashboard-api.md#L227-L228) documents `minDurationMs` and `maxDurationMs` as filters on the recorded request duration. Its [query](https://github.com/ganjianping/gjpb-api-spring-web-public/blob/db2680196d5df62da7717f763189560f63c252c7/src/main/java/org/ganjp/blog/common/audit/repository/AuditLogRepository.java#L127-L128) includes both endpoints when both bounds are set. The two rules abbreviate the unit only in the phrases shown; they do not convert values, measure the log query’s runtime, or classify a slow request as timed out.

The [registry audit API](https://github.com/agentic-community/mcp-gateway-registry/blob/45f537a85309fc924d03952dcf0efc31763d452a/registry/audit/routes.py#L547-L552) uses this parameter to [match the decision stored in an audit entry](https://github.com/agentic-community/mcp-gateway-registry/blob/45f537a85309fc924d03952dcf0efc31763d452a/registry/audit/routes.py#L627-L636). The rule matches the full parameter description, not a bare field or value. An `ALLOW` decision does not prove the request succeeded or the reader may view the logs; a missing decision is not `DENY`.

The [Trading API audit repository](https://github.com/APLionsAdvancedSWE/Advanced-Programming-Group-Project/blob/6060c2326eb820db83b532fb9435d48962c9af37/src/main/java/com/dev/tradingapi/repository/AuditLogRepository.java#L25-L31) uses this description for `findByPath`. It searches the recorded path (for example, `/books`), not the request target with its query string (`/books?limit=2`) or the endpoint used to read the logs. The rule changes only the description, not any path or query parameter.

The [audit repository](https://github.com/ganjianping/gjpb-api-spring-web-public/blob/db2680196d5df62da7717f763189560f63c252c7/src/main/java/org/ganjp/blog/common/audit/repository/AuditLogRepository.java#L37-L40) uses this description for a lookup on the recorded `httpMethod`. It filters entries by the method of the logged request, not the method used to fetch the logs or an application operation name. The rule changes only the label, not a method value or a request.

The [Books Authors API repository](https://github.com/desobsesor/books-authors-api/blob/6bd87358a0e1bf757b29e307f1852f7ab5fa5891/domain/src/main/java/com/books/domain/repository/ApiAuditLogRepository.java#L45-L51) uses this phrase for `findByStatusCode`. The filter applies to the HTTP status code stored in each entry, not to the status of the log query itself. The rule matches the capitalization shown; it does not change a code, imply a match, or infer a business outcome.

The [OpenShift example](https://github.com/openshift/openshift-docs/blob/48f05989ccf7749c9822482b98e01e446faa0a7a/modules/security-audit-log-filtering.adoc#L40-L55) filters entries by the recorded `userAgent` value; the next example only displays that field after filtering by API version. The rule matches the capitalization shown and keeps the filter as an action, not a claim that any entry matched. The [Kubernetes audit schema](https://github.com/kubernetes/apiserver/blob/v0.31.0/pkg/apis/audit/v1/types.go#L103-L106) identifies `userAgent` as client-reported, not proof of client or operator identity.

The [SIGRAP audit API](https://github.com/Nick220505/sigrap-backend/blob/acbb9b5d475454523dff668443b58f84247bba31/src/main/java/com/sigrap/audit/AuditLogController.java#L229-L239) uses this label for a query on the recorded source IP address. The rule changes only the verb in the phrase shown. A source IP is not a destination address or proof of the operator’s identity. Requesting the logs does not imply that any matched.

The [Red Hat RBAC client](https://github.com/RedHatInsights/javascript-clients/blob/89a1679201c66f6df38cdf4eb76711ec9ab1c588/packages/rbac/src/GetAuditlogs/index.ts#L33-L38) uses this phrase for its `resourceType` query parameter. The rule matches the capitalization shown and leaves the parameter name and values alone. Resource type is the filter here, not a resource name or an event type; the description does not say any records matched.

The [Vonage SDK](https://github.com/Vonage/vonage-node-sdk/blob/ec5a075258a5ca11839eb3fde895fa438835e952/packages/audit/lib/audit.ts#L91-L107) uses this sentence to describe `getEvent(eventId)`. The rule expands `ID` only in the phrase shown, with the same capitalization. It refers to the event’s own identifier, not an operation name or a correlation ID. It describes what the method does, not whether a particular lookup found a record.

[cfnrc](https://github.com/megaproaktiv/cfnrc/blob/e367701a9e8b72328a01643a1ece4e7c26188c37/cloudtrail/cloudtrail.go) uses the message on the left when a lookup filtered by `EventName` fails. The rule matches the capitalization shown. For an API event, the filter names the operation, not the individual event (`EventId`) or the caller (`Username`). A failed lookup says nothing about whether the recorded operation succeeded.

CloudTrail’s [`LookupEvents`](https://github.com/aws/aws-sdk-go-v2/blob/3c9cb8bc1bb7235972be82b37b8af6bad3c2187f/service/cloudtrail/api_op_LookupEvents.go#L25-L37) can filter event history by [`ReadOnly`](https://github.com/aws/aws-sdk-go-v2/blob/3c9cb8bc1bb7235972be82b37b8af6bad3c2187f/service/cloudtrail/types/types.go#L547-L548). The rule matches the field description as capitalized above, not the field name or its value. The flag classifies the recorded operation; it says neither who can view the log nor whether the operation succeeded.

[Airbyte’s audit log reader](https://github.com/airbytehq/airbyte-platform/blob/63d665f6442d2b821fe3ed2a33b209c1b4132da3/airbyte-audit-logging/src/main/kotlin/io/airbyte/audit/logging/read/AuditLogQuery.kt) uses this message when it rejects a page token. The error alone does not mean the query ID is missing, the results expired, or the last page was reached.

[CloudTrail’s `DescribeQuery` response](https://github.com/aws/aws-sdk-go-v2/blob/3c9cb8bc1bb7235972be82b37b8af6bad3c2187f/service/cloudtrail/api_op_DescribeQuery.go#L99-L100) has an optional `ErrorMessage` separate from `QueryStatus`. The rule matches only the field description as capitalized above; it does not supply a missing reason. A failed status request does not tell the caller that the query itself failed.

[CloudTrail `ListQueries`](https://github.com/aws/aws-sdk-go-v2/blob/3c9cb8bc1bb7235972be82b37b8af6bad3c2187f/service/cloudtrail/api_op_ListQueries.go) uses this description for `MaxResults`. It limits query executions on one page, not the total number of queries. [`GetQueryResults.MaxQueryResults`](https://github.com/aws/aws-sdk-go-v2/blob/3c9cb8bc1bb7235972be82b37b8af6bad3c2187f/service/cloudtrail/api_op_GetQueryResults.go) limits result rows for one query instead. The rule matches the capitalization shown and sets no page size.

[CloudTrail’s `ListQueries` paginator](https://github.com/aws/aws-sdk-go-v2/blob/3c9cb8bc1bb7235972be82b37b8af6bad3c2187f/service/cloudtrail/api_op_ListQueries.go#L227-L252) uses its `NextToken` to continue the list of query executions. [`GetQueryResults`](https://github.com/aws/aws-sdk-go-v2/blob/3c9cb8bc1bb7235972be82b37b8af6bad3c2187f/service/cloudtrail/api_op_GetQueryResults.go) takes a `QueryId` for one execution and uses its own `NextToken` to page through result rows. The rule matches only the description as capitalized above, not either token or the query ID.

[CloudTrail `ListQueries`](https://github.com/aws/aws-sdk-go-v2/blob/3c9cb8bc1bb7235972be82b37b8af6bad3c2187f/service/cloudtrail/api_op_ListQueries.go) accepts either the store’s full ARN or its ID suffix in `EventDataStore`. Both identify the store whose query executions are listed. Neither is a `QueryId` for one execution or a `DeliveryS3Uri` for its results. The rule matches only the field description as written, not identifiers or their values.

[CloudTrail’s `StartTime` and `EndTime` descriptions](https://github.com/aws/aws-sdk-go-v2/blob/3c9cb8bc1bb7235972be82b37b8af6bad3c2187f/service/cloudtrail/api_op_ListQueries.go) share the clause on the left. It filters the list by when the queries ran, not by when the audit events occurred. The rule matches the capitalization shown and leaves both timestamps and any retention policy unchanged.

In [CloudTrail’s `ListQueries` request](https://github.com/aws/aws-sdk-go-v2/blob/3c9cb8bc1bb7235972be82b37b8af6bad3c2187f/service/cloudtrail/api_op_ListQueries.go), `QueryStatus` filters the list of query executions, not the audit events or result rows. Requesting `FINISHED` does not establish that any query finished. Only the field description is rewritten.

[CloudTrail’s `ListQueries` response](https://github.com/aws/aws-sdk-go-v2/blob/3c9cb8bc1bb7235972be82b37b8af6bad3c2187f/service/cloudtrail/api_op_ListQueries.go) lists query executions, not audit events. The case-sensitive rule matches the full description on one line. `CreationTime` belongs to the query even when `QueryStatus` is `QUEUED`; it is neither an execution start time nor the timestamp of an event being searched.

[CloudTrail’s `ExecutionTimeInMillis` definition](https://github.com/aws/aws-sdk-go-v2/blob/3c9cb8bc1bb7235972be82b37b8af6bad3c2187f/service/cloudtrail/types/types.go#L1520-L1521) describes how long a single query ran, not how long a client waited or its timeout limit. The rule matches the wording and capitalization shown; it leaves the unit, value and query status unchanged.

[CloudTrail’s `BytesScanned` definition](https://github.com/aws/aws-sdk-go-v2/blob/3c9cb8bc1bb7235972be82b37b8af6bad3c2187f/service/cloudtrail/types/types.go#L1504-L1518) measures bytes scanned by one query, not events scanned or bytes written to a result file. The rule matches the wording and capitalization shown. It does not change any value or estimate a charge.

[CloudTrail’s query statistics](https://github.com/aws/aws-sdk-go-v2/blob/3c9cb8bc1bb7235972be82b37b8af6bad3c2187f/service/cloudtrail/types/types.go#L1514-L1518) use the phrase on the left for `EventsScanned`. It describes one query, not the number of events that matched (`EventsMatched`) or the rows returned on a page (`QueryResultRows`). Only the wording shown matches; field names and counts stay unchanged.

[CloudTrail’s `DescribeQuery` documentation](https://github.com/aws/aws-sdk-go-v2/blob/3c9cb8bc1bb7235972be82b37b8af6bad3c2187f/service/cloudtrail/api_op_DescribeQuery.go) distinguishes query execution (`QueryStatus`) from result delivery (`DeliveryStatus`). The rule matches the full, case-sensitive sentence and preserves the S3 condition. A finished query does not establish delivery, and delivery does not establish that a caller has read the results.

[AWS Labs’ CloudTrail integration](https://github.com/awslabs/mcp/blob/0386d53aad258e58163e6bba2bc4ca78780dd526/src/cloudtrail-mcp-server/awslabs/cloudtrail_mcp_server/tools.py) checks one query execution using its `QueryId`, not an audit event ID. The rule matches the capitalization shown. Checking the status does not fetch result rows; `get_query_results` is a separate action.

[Sim’s CloudTrail integration](https://github.com/simstudioai/sim/blob/e28e35bb961eb6d83bf3796ac70b366fa41fe0ae/apps/sim/tools/cloudtrail/cancel_query.ts) uses the case-sensitive label on the left. It names an action, not a timeout or confirmation that processing stopped. [AWS documents](https://github.com/aws/aws-sdk-go-v2/blob/main/service/cloudtrail/api_op_CancelQuery.go) that even `CANCELLED` can appear before the operation finishes.

[Omnilogs](https://github.com/MathPro29/Omnilogs/blob/f502388510d243063d75ff36729f96b5213c0aa6/backend/api/internal/dashboard/handler/handler.go) uses the message on the left for an expired deadline. [Triage](https://github.com/nurejev/triage/blob/154a5ab3596298763928e94bb6e22a1fc6fc4f48/js/graph.js) distinguishes a wait timeout from a cancelled query. Neither wording says that the server stopped processing or that there were no matches.

[PropelIQ’s audit viewer](https://github.com/BhalaGanesh30/PropelIQ-POC-Appointment/blob/f4f4d895a55d1a43bcd9d8bf84ca49b0ae6b5e0c/app/src/app/features/admin/pages/audit-log/audit-log-viewer.component.html) uses this phrase as the accessible label for a button shown after a load error. The rule matches the capitalization shown and leaves the generic `Retry` button text alone. The label offers another attempt; it does not say one has started or succeeded.

[Opfield’s audit viewer](https://github.com/opfield/opfield/blob/3a97f714d8b5ef0f89a516949dd814f5c4c65d74/packages/frontend/src/pages/AuditLog.tsx) displays this label while a request is pending. The rule matches the capitalization shown. It leaves `No audit log entries found` and `Error loading audit log entries` unchanged; the loading label gives no record count.

[FeatBit’s audit viewer](https://github.com/featbit/featbit/blob/de394e5a4f0623fdd565bd279746137a0101536c/modules/front-end/design/auditlogs-page-design.md) uses `No audit logs match your filters` when no entries match the current selection. It treats `No audit logs yet`, permission errors, and load failures as separate states. This rule rewrites only the filtered message; it says nothing about how many records are stored.

The [Cloud Logging `LogEntry` schema](https://github.com/googleapis/googleapis/blob/9415ba048aa587b1b2df2b96fc00aa009c831597/google/logging/v2/log_entry.proto) distinguishes `timestamp` (when the event occurred) from `receive_timestamp` (when Logging received the entry). The two rules match the labels as capitalized above. Neither renames a field, changes a time value, or treats receipt as a completed disk write.

The [DecisionRules audit guide](https://github.com/decisionrules/documentation/blob/72eb044084ee95ef75376fb24ed3bde6fb305aae/intelligence/audit-logs.md) uses the sentence on the left. Only `displayed` changes to `shown`: the rule neither selects a time zone nor changes stored timestamps or the system clock.

In [RakshakGIS’s audit detail view](https://github.com/OjaswiJoshi13/rakshakgis/blob/bccc35212b390b764b6ee4279ac7e35a2fb1307d/frontend/src/components/operations/audit/AuditDetailModal.tsx), `Copy Audit Event ID` copies `record.id`, not the event object or its hash. The rule only expands `ID` in that case-sensitive label; it does not copy data or export logs.

[Quotacore](https://github.com/haneefojay/Quotacore/blob/fb8d6572dfaf5b08c6ad3f0d7ad3934443b57b37/docs/product/use-cases.md) uses the wording on the left. [Bower](https://github.com/bower-labs/docs/blob/f1cb94a14f68a706b0b68dc46abf032895c3561c/trust-center/audit-logs.mdx) describes the values before and after a change for each field. The rule only changes how that comparison is described: it does not fill in missing values, rename `oldValue` or `newValue`, or restore a previous value.

[Bower’s audit log guide](https://github.com/bower-labs/docs/blob/f1cb94a14f68a706b0b68dc46abf032895c3561c/trust-center/audit-logs.mdx) uses the wording on the left to describe viewing one entry’s changes. The rule matches the full clause, not the button label `Expand`. Editing a record or exporting it is a different action.

[Kaiku’s audit viewer](https://github.com/Detair/kaiku/blob/28a5402266f56b9bd73500f3841e37c69c4243e2/client/src/stores/admin.ts) reloads the list with no active filters. The rule retains both actions; it does not match `Clear filters`, deleting a saved preset, or deleting audit records.

[NetSentinel’s client](https://github.com/TahirAlauddin/NetSentinel/blob/cc868a4cb5a356d6693c21155e471bf9c0bae1f7/data-plane/src/frontend/lib/api-client/ipam.ts) uses the phrase on the left. Its [filter model](https://github.com/TahirAlauddin/NetSentinel/blob/cc868a4cb5a356d6693c21155e471bf9c0bae1f7/data-plane/src/backend/ipam/models/ip_audit_log.py) stores a combination of search criteria for reuse. Saving a new filter does not mean that matching records have been saved or that the query is scheduled to run.

[Microsoft Learn](https://github.com/MicrosoftDocs/learn/blob/4d04386c689abe81f68f29905a946ffb0cff2c1d/learn-pr/wwl-sci/investigate-threats-using-audit-in-microsoft-365-defender-microsoft-purview-standard/includes/6-export-configure-view-audit-log-records.md) uses the phrase on the left for exporting search results, not the entire audit log. The rule matches the case shown. It does not set CSV as the default or say that an export succeeded.

The [Cortex XDR integration](https://github.com/demisto/content/blob/09dffd301cebb8809f12020773f362ed68787d6a/Packs/CortexXDR/Integrations/CortexXDRIR/README.md) uses the phrase on the left for a per-page upper bound. A page may contain fewer entries, and the total number of matches may be larger ([Google AIP-158](https://google.aip.dev/158)). This rule does not set a default page size.

[ProKnow’s audit SDK](https://github.com/proknow/proknow-sdk-dotnet/blob/52c86ca2d2e1e5ab523b7d3e0210437aaa181812/proknow-sdk/Audit/AuditPage.cs) uses `get` for a single next page, not all remaining records. In [CloudTrail’s event lookup](https://github.com/aws/aws-sdk-go-v2/blob/main/service/cloudtrail/api_op_LookupEvents.go), `NextToken` continues the query; `EventId` identifies an event. Neither field name is rewritten.

The [Artifact Keeper query](https://github.com/artifact-keeper/artifact-keeper/blob/4e7de166f036079af67bbf3d558986bd4cf78b7b/backend/src/services/audit_service.rs) uses the sentence-case wording shown above; only that spelling matches. A correlation ID groups related activities. It is not the event ID, the affected resource ID, or proof of a shared actor (see [Microsoft Graph’s separate fields](https://github.com/microsoftgraph/microsoft-graph-docs-contrib/blob/main/api-reference/v1.0/resources/directoryaudit.md)).

[Azure’s directory audit example](https://github.com/Azure/azure-docs-powershell-azuread/blob/693b89e07c621ddc4d4202e3bfcd12d88fc43932/azureadps-2.0-preview/AzureADPreview/Get-AzureADAuditDirectoryLogs.md#example-3-get-audit-logs-containing-a-given-target-resource) uses `get`. The rule changes only that verb to `retrieve`. The filter still selects entries containing the affected resource, not entries selected by the actor or the resource owner.

SensitiveFlow’s [samples](https://github.com/MateusjsSilva/SensitiveFlow/blob/dd0b9f8f22d0102def78290391061917db61d642/samples/README.md) and [changelog](https://github.com/MateusjsSilva/SensitiveFlow/blob/dd0b9f8f22d0102def78290391061917db61d642/CHANGELOG.md) use these two verbs for the same search. The filter still selects the actor who performed the action, not the affected account or the service that reported it.

[Open Foundry’s audit query](https://github.com/syzygyhack/open-foundry/blob/9736b099fa08499e58c6fd22489312a8e0907d70/packages/security/src/audit/audit-query.ts) uses the wording on the left. As in [Microsoft’s audit log search](https://github.com/MicrosoftDocs/office-docs-powershell/blob/d00ca99f53eaad31d801eb68dfbd7032bf57ef91/exchange/exchange-ps/ExchangePowerShell/Search-UnifiedAuditLog.md), the range limits which stored records to look for, not when to collect them or how long to keep them. The rule leaves the range and its endpoints unchanged.

[SentryDrive’s activity view](https://github.com/christliebdela/SentryDrive/blob/b2459b8f9b7fcd8e5aade709b71ea67fa619aea1/src/features/activity/ActivityView.tsx) uses the phrase on the left. [Mixpanel’s audit log](https://github.com/mixpanel/docs/blob/b0d38dbc21ca4c9825f9bcfa22b1e53b91c16bc1/docs/access-security/audit-log.mdx) likewise filters displayed entries by event type. The rule keeps both `search` and `filter`: it describes narrowing query results, not changing which events are collected or deleting stored entries.

[Domino](https://github.com/dominodatalab/audit-report/blob/148732140f5b9d778a1c61c72d7e3e48e50e25f7/ReadMe.md) and [OneUptime](https://github.com/OneUptime/blog/blob/acd0a2f22356a3090d1c7390757e23a83058f7cf/posts/2026-02-21-how-to-use-ansible-for-compliance-as-code/README.md) use these phrases for generating a report when requested. The rule does not match scheduled reporting, a report already generated, or the collection of audit records.

The [content-preservation objective](https://github.com/zerobias-org/framework/blob/fe8e9e34860b9dbb8ed4abaa2326c821d4a6c324/package/nist/800_171a/r3/elements/a_03_03_06_b_01_.yml) uses `is preserved`. Here `is kept intact` means the original content is unchanged during reporting, not that a verifier checked it or that nobody can change it. The separate phrase `is not preserved` does not match.

[NIST AU-7](https://doi.org/10.6028/NIST.SP.800-53r5) distinguishes leaving the time ordering of audit records intact from changing it. The [assessment objective](https://github.com/zerobias-org/framework/blob/fe8e9e34860b9dbb8ed4abaa2326c821d4a6c324/package/nist/800_171a/r3/elements/a_03_03_06_b_02_.yml) uses the wording on the left. The rule only renames the existing order; it neither sorts timestamps nor asserts that the order was preserved or the clocks were synchronized.

[PagerDuty](https://github.com/PagerDuty/api-schema/blob/c928d7cf3c7c8b50030316149e04318b6b9300e1/docs/CONCEPTS.md) uses `provides` to describe its audit API; the rule changes only that verb. Read-only describes what this access path allows, not whether anyone read a record or whether another writer can change it.

[encry-outcry](https://github.com/Odiambo/encry-outcry/blob/327784110fe8aeb5df4e32cb96b05ac563bdb5f5/audit/audit_log.py) uses the compact phrase for a check that returns a validity flag and any errors; [bank-account-rust](https://github.com/ProdByBuddha/bank-account-rust/blob/63b3173ec2522fd0c13828d91a720e156ee111c2/src/audit/mod.rs) uses the expanded wording. Both name the check itself. Rewriting the label does not run the check or report that the file passed.

[OxiDB](https://github.com/parisxmas/OxiDB/blob/1cfd96f15a58f87c41372571d697ce95e210cab9/oxidb-server/tests/security_audit_gzip.rs) uses the compact phrase. The expansion makes clear that only files already rotated are meant, not the active log.
The [Kubernetes audit flag](https://github.com/kubernetes/website/blob/be58364903ad78d60314019b81759f05c33b4af5/content/en/docs/reference/command-line-tools-reference/kube-apiserver.md) specifies gzip for its implementation; the glossary neither selects a codec nor changes retention, archival or encryption settings.

The [collective.fingerpointing changelog](https://github.com/collective/collective.fingerpointing/blob/6ddcd2029300dcebc309c5d0396e0f32f25aa32f/CHANGES.rst) and [OpenShift documentation](https://github.com/openshift/openshift-docs/blob/e380ba9a51df95e69697254c673600f2fb801d0e/microshift_configuring/microshift_auth_security/microshift-audit-logs-config.adoc) use these two phrases for file rotation. The rule renames the operation; it does not change retention settings or match backup, archival or deletion of audit logs.

The [auditing lesson](https://github.com/tokitaari/ansible-cis-lab/blob/c7d9c235841ddd275b69e83e56bbfa25266d671f/docs/lessons/L09-auditing-and-integrity.md) and the [Credenza manual](https://github.com/dantoac/credenza/blob/3a1d91aec583effabc48eb2bf0bfa4df1413e5f5/docs/user-manual.md) both describe low free space on the disk that holds the log.
This rule does not set a warning threshold or match a full disk, a failed write, or a warning about the log file alone.

[NIST AU-4](https://doi.org/10.6028/NIST.SP.800-53r5) and [MITRE’s audit storage check](https://github.com/mitre/mongodb-enterprise-advanced-4-stig-baseline/blob/ec1a61b231f1a77ddd7e094f1266fe8edcac2c43/controls/SV-252172.rb) use these two wordings for the same allocation action.
Only the phrase beginning with `allocate` is rewritten. A capacity limit, remaining free space and a retention period are distinct; this rule does not change any of them.

[Eclipse ECSP](https://github.com/eclipse-ecsp/uidam-authorization-server/blob/aff1df5f8040256bf25e7902c1d232fbbf302979/src/main/java/org/eclipse/ecsp/audit/enums/AuditEventResult.java) defines the label as the outcome of the event being audited, not whether the audit record was written.
[NIST AU-3](https://doi.org/10.6028/NIST.SP.800-53r5) notes that an outcome can include event-specific details beyond success or failure. This rule changes the label, not values such as `success` or `failure`.

The [Google Cloud PBMM design document](https://github.com/GoogleCloudPlatform/pbmm-on-gcp-onboarding/blob/6b87cee68d33f08ece6c399e8d16b33722cba913/docs/technical-design-document.md) uses the one-word spelling `timestamps` in this phrase;
[NIST SP 800-53 Rev. 5, AU-8](https://doi.org/10.6028/NIST.SP.800-53r5) uses `time stamps`. Only the full plural phrase is normalized.
This neither changes timestamp values nor verifies clock synchronization.

[Bluelink](https://github.com/newstack-cloud/bluelink/blob/647c0439d2fbda7c6bf3153ffd4c77ec53a3d1f3/apps/deploy-engine/README.md#audit-record-retention-period) uses the compact term;
[Microsoft](https://github.com/MicrosoftDocs/win32/blob/bc82a65f74ef2c4f2f44a3d7db5f7cfebfdc3583/desktop-src/SecMgmt/policy-object-access-rights.md) uses the expanded wording for the same duration.
The rule does not change a retention policy or set a number of days. [NIST SP 800-53 Rev. 5, AU-11](https://doi.org/10.6028/NIST.SP.800-53r5) leaves the time period to the organization.

The IDS entries distinguish a system monitoring a host ([OSSEC](https://github.com/ossec/ossec-docs/blob/5b1d9c56f72847619c0aeae7c431276196bb0bef/about.rst))
from one monitoring network traffic ([Security Onion](https://github.com/Security-Onion-Solutions/securityonion-docs/blob/2.4/nids.rst)).
The English phrases match only as spelled above, at word boundaries: `host-based IDs` can refer to identifiers.
Bare `IDS`, `HIDS`, `NIDS` and `IPS` remain unchanged. The Chinese expansions are project wording, not a compliance assessment.

Term names were checked against MITRE ATT&CK's
[Password Spraying](https://attack.mitre.org/techniques/T1110/003/) and
[Browser Session Hijacking](https://attack.mitre.org/techniques/T1185/)
entries and its [technique directory](https://attack.mitre.org/techniques/).
Ticket and credential terminology also follows
[Pass the Ticket](https://attack.mitre.org/techniques/T1550/003/) and
[OS Credential Dumping](https://attack.mitre.org/techniques/T1003/).
Generic words such as `ticket`, `dumping` and 票据 are not added as rules.
Name-resolution terminology was checked against
[MITRE Adversary-in-the-Middle](https://attack.mitre.org/techniques/T1557/) and
[OWASP DNS Cache Poisoning](https://cornucopia.owasp.org/taxonomy/capec-3.9/142).
The standalone names `DNS`, `ARP` and `cache` remain unchanged.
Session and interaction terminology was checked against
[OWASP Session Fixation](https://community.owasp.org/attacks/Session_fixation) and
[Clickjacking Defense](https://cheatsheetseries.owasp.org/cheatsheets/Clickjacking_Defense_Cheat_Sheet.html).
Only the attack phrases are added; `session fixation`, `clickjacking defense`
and `SessionFixationProtection` stay unchanged.
HTTP terminology follows [CWE-444](https://cwe.mitre.org/data/definitions/444.html)
for inconsistent parsing between HTTP components and
[CWE-113](https://cwe.mitre.org/data/definitions/113.html) for CRLF handling in
HTTP headers. These remain distinct risks. Only the complete phrases in the
table match; Chinese forms require the displayed spelling without spaces.
English forms are case-insensitive with word boundaries, and all-uppercase
input carries its casing to the output. `HTTP`, `request`, `response`,
`CWE-444`, `CWE-113`, `Content-Length` and `Transfer-Encoding` are not rules.
Web-cache terminology follows
[OWASP Web Cache Security](https://cheatsheetseries.owasp.org/cheatsheets/Web_Cache_Security_Cheat_Sheet.html)
and [PortSwigger Web Cache Deception](https://portswigger.net/web-security/web-cache-deception).
Poisoning concerns the integrity of cached responses; deception concerns private
responses being cached incorrectly and exposed. Their replacement phrases
remain distinct, and the existing DNS/ARP mappings are unchanged. Only the
complete phrases in the table match: Chinese forms require the displayed
`Web` casing without spaces; English forms use case-insensitive word boundaries.
Standalone `cache poisoning`, `cache deception`, 缓存投毒 and 缓存欺骗 are not rules.
Object-attribute terminology follows [CWE-1321](https://cwe.mitre.org/data/definitions/1321.html)
and [OWASP Mass Assignment](https://cheatsheetseries.owasp.org/cheatsheets/Mass_Assignment_Cheat_Sheet.html).
Prototype tampering and unintended field modification remain distinct risks;
[CWE-915](https://cwe.mitre.org/data/definitions/915.html) is a broader attribute
modification weakness, not an exact synonym for mass assignment. Only the full
attack/vulnerability phrases are rules. Standard terms `prototype pollution`,
`mass assignment`, ordinary 批量赋值, protection names and `__proto__`,
`constructor`, `prototype`, `CWE-1321`, `CWE-915` stay unchanged.
Dependency terminology follows
[OWASP Dependency Chain Abuse](https://owasp.github.io/www-project-top-10-ci-cd-security-risks/CICD-SEC-03-Dependency-Chain-Abuse)
and [CI/CD Security](https://cheatsheetseries.owasp.org/cheatsheets/CI_CD_Security_Cheat_Sheet.html).
Non-intended dependency source resolution differs from package-name
impersonation; the original 供应链攻击 mapping remains unchanged. Only the
attack phrases in the table are rules, not `dependency confusion`,
`typosquatting` or generic package/dependency words. Existing import, package
and URL guards take precedence: `package typosquatting attack` at the start
of a line (including a comment after its marker) is left untouched. In normal
prose such as `a package typosquatting attack`, the complete phrase matches.
Real package names and import paths are not renamed by these rules.
XML terminology follows [CWE-611](https://cwe.mitre.org/data/definitions/611.html),
[CWE-776](https://cwe.mitre.org/data/definitions/776.html) and
[OWASP XML External Entity Prevention](https://cheatsheetseries.owasp.org/cheatsheets/XML_External_Entity_Prevention_Cheat_Sheet.html).
External entity resolution can access unintended resources; recursive entity
expansion can exhaust resources without external access. These are distinct
risks, and disabling external entities alone is not complete protection against
recursive expansion. Only the complete attack phrases above match. Chinese
forms require uppercase `XML` and no spaces; English forms use case-insensitive
word boundaries. Standard terms without `attack`, `XML`, `XXE`, `DTD`,
`DOCTYPE`, `ENTITY`, parser configuration names and CWE identifiers stay
unchanged. No `XML bomb` or `Billion Laughs` aliases are added.
Template and XPath terminology follows
[CWE-1336](https://cwe.mitre.org/data/definitions/1336.html) and
[CWE-643](https://cwe.mitre.org/data/definitions/643.html). Template injection
interprets untrusted input as template expressions or directives; XPath injection
changes query-expression structure. CWE-1336 also covers client-side cases,
whereas these rules intentionally cover only the server-side attack phrase.
Neither risk is equivalent to external entity resolution or recursive entity
expansion. Replacements retain `injection` rather than imply protection.
Only the four complete attack phrases above are added: Chinese forms have no
spaces and use the displayed `XPath` casing; English forms use case-insensitive
word boundaries and require the literal `server-side` hyphen where shown.
Standard terms without the attack qualifier, generic words, `SSTI`, CWE
identifiers, template/query API names and client-side phrases stay unchanged.
LDAP and log-forging terminology follows
[CWE-90](https://cwe.mitre.org/data/definitions/90.html) and
[CWE-117](https://cwe.mitre.org/data/definitions/117.html). LDAP injection alters
directory-query structure; log forging compromises log-record integrity. Log
forging is one manifestation of CWE-117, not its entire scope or every form of
log-file tampering. Replacements preserve the injection/forgery risk distinction.
Only the four complete attack phrases above are added. Chinese forms require
uppercase `LDAP` where shown and no spaces; English forms use case-insensitive
word boundaries. Standard terms without the attack qualifier, LDAP attribute
names, log levels, API names and CWE identifiers stay unchanged. Generic `LDAP`,
`log`, `logging`, 日志 and the `log injection attack` alias are not added.
Regular-expression complexity terminology follows
[CWE-1333](https://cwe.mitre.org/data/definitions/1333.html). Inefficient worst-case
matching can consume excessive CPU and reduce availability; this is distinct
from XML entity expansion and does not make all regexes or backtracking unsafe.
Only the two complete attack phrases above are rules. Chinese requires the
shown spelling without spaces; English uses case-insensitive word boundaries
and literal spaces. `ReDoS`, `DoS`, `DDoS`, `regex`, `regular expression`,
`catastrophic backtracking`, API names and `CWE-1333` remain unchanged, as do
standard terms without the attack qualifier. No hyphenated aliases are added.
HTTP parameter-pollution terminology follows
[CAPEC-460](https://capec.mitre.org/data/definitions/460.html) and
[OWASP WSTG](https://owasp.github.io/www-project-web-security-testing-guide/stable/4-Web_Application_Security_Testing/07-Input_Validation_Testing/04-Testing_for_HTTP_Parameter_Pollution).
Duplicate parameters alone do not establish a vulnerability. Unexpected value
interpretation, inconsistent handling or unsafe concatenation can affect
application logic; this does not always require multiple components. It is
distinct from HTTP request smuggling and object prototype pollution. Only the
two complete attack phrases above are rules: Chinese requires uppercase `HTTP`
and no spaces; English uses case-insensitive word boundaries and literal spaces.
Standard terms without the attack qualifier, `HPP`, generic parameter words,
protocol fields, API names, `CAPEC-460` and `WSTG-INPV-04` remain unchanged.
CSV formula-injection terminology follows
[CWE-1236](https://cwe.mitre.org/data/definitions/1236.html) and
[OWASP CSV Injection](https://community.owasp.org/attacks/CSV_Injection).
Untrusted CSV export values may be interpreted as formulas by a spreadsheet
client; impact depends on the client and its configuration. Normal formulas
are not automatically attacks, and command execution is not guaranteed.
These rules are limited to the two complete CSV attack phrases above, not all
spreadsheet or template injection. Chinese requires uppercase `CSV` and no
spaces; English uses case-insensitive word boundaries and literal spaces.
Standard terms without the attack qualifier, generic words, API names and
`CWE-1236` remain unchanged; no `CSV injection attack` or `formula injection attack`
alias is added. Renaming does not escape or sanitize exported CSV contents.
Redirect terminology follows
[CWE-601](https://cwe.mitre.org/data/definitions/601.html): externally controlled
redirects to untrusted destinations cross a target trust boundary. Normal
redirects, HTTP 3xx responses and redirect APIs alone do not establish a
vulnerability. This is distinct from HTTP response splitting and does not
cover every navigation or internal forward. Only the two complete attack
phrases above are rules. Chinese requires the shown spelling without spaces;
English uses case-insensitive word boundaries and literal spaces. `open redirect`,
开放重定向, generic redirect words, `URL`, `Location`, status codes, API names
and `CWE-601` remain unchanged. No extra aliases or phishing rules are added.
DOM-clobbering terminology follows
[OWASP DOM Clobbering Prevention](https://cheatsheetseries.owasp.org/cheatsheets/DOM_Clobbering_Prevention_Cheat_Sheet.html).
Untrusted HTML can introduce named-property collisions affecting property
resolution. This differs from prototype pollution, which changes object
prototypes; normal DOM operations or `id`/`name` attributes alone are not attacks,
and script execution is not inevitable. Only `DOM clobbering attack` is added,
with case-insensitive word boundaries and literal spaces. `DOM clobbering`,
protection names, DOM APIs and properties remain unchanged. No Chinese alias,
hyphenated spelling or underscore spelling is added. Renaming does not modify
the actual DOM, sanitize HTML or repair the underlying trust boundary.
HTTP host-header terminology follows
[OWASP WSTG Host Header Injection](https://wstg.owasp.org/stable/4-Web_Application_Security_Testing/07-Input_Validation_Testing/17-Testing_for_Host_Header_Injection/).
Improper trust in external host-header values can affect generated links,
destinations or caching, depending on how the application uses them. Normal
host fields, virtual hosts and proxy configuration alone do not establish a
vulnerability. This is distinct from response splitting; cache poisoning or
redirects are possible consequences, not equivalent risks or guaranteed results.
Only the two complete HTTP attack phrases above are rules. Chinese requires
uppercase `HTTP` and no spaces; English uses case-insensitive word boundaries
and literal spaces. Standard terms without the attack qualifier, `Host`,
`X-Forwarded-Host`, `Forwarded`, `:authority`, `Request.Host` and `WSTG-INPV-17`
remain unchanged. No shorter attack aliases or generic header words are added.
NoSQL terminology follows
[OWASP NoSQL Security](https://cheatsheetseries.owasp.org/cheatsheets/NoSQL_Security_Cheat_Sheet.html)
and [CWE-943](https://cwe.mitre.org/data/definitions/943.html). Untrusted input
that changes query structure or operators can change intended query logic;
structured query objects alone do not guarantee safety. Normal queries,
operators and database names do not establish a vulnerability. CWE-943 covers
query logic more broadly, not only NoSQL injection. Only the two complete
NoSQL attack phrases above are rules. Chinese requires the displayed `NoSQL`
casing without spaces; English uses case-insensitive word boundaries and
literal spaces. Standard terms without the attack qualifier, SQL terminology,
database/API/operator names and `CWE-943` remain unchanged. No `No SQL`, `No-SQL`
or other aliases are added.
Cross-site WebSocket terminology follows
[OWASP WebSocket Security](https://cheatsheetseries.owasp.org/cheatsheets/WebSocket_Security_Cheat_Sheet.html).
A browser handshake may include applicable session cookies; accepting untrusted
origins while relying on those credentials can permit session hijacking.
Impact depends on origin validation, authentication and authorization. Normal
cross-site connections or the `Origin` field alone do not establish a
vulnerability, and `Origin` is not a standalone authentication credential.
Only the two complete attack phrases above are rules. Chinese requires the
displayed `WebSocket` casing without spaces; English uses case-insensitive word
boundaries, the literal `cross-site` hyphen and literal spaces. Standard terms
without the attack qualifier, `CSWSH`, protocol fields, API names and protection
names remain unchanged. No `cross site` or `web socket` aliases are added.
Renaming does not configure handshake validation or other session defenses.
JWT algorithm-confusion terminology follows
[RFC 8725](https://www.rfc-editor.org/rfc/rfc8725.html) and
[PortSwigger Algorithm Confusion](https://portswigger.net/web-security/jwt/algorithm-confusion).
RFC section 2.1 describes signature-validation weaknesses; section 3.1 specifies
algorithm verification. Accepting an unintended algorithm or misinterpreting key
usage can break the intended verification boundary. This is not every JWT risk,
and normal token validation or algorithm names alone do not establish a
vulnerability. Only the two complete JWT attack phrases above are rules.
Chinese requires uppercase `JWT` and no spaces; English uses case-insensitive
word boundaries and literal spaces. Standard terms without the attack qualifier,
JWT fields, algorithm/API names and `RFC 8725` remain unchanged. No key-confusion
aliases are added. Renaming neither validates signatures nor changes algorithm
configuration, keys or actual tokens.
XML signature-wrapping terminology follows
[OWASP SAML Security](https://cheatsheetseries.owasp.org/cheatsheets/SAML_Security_Cheat_Sheet.html).
The XML element actually consumed must be bound to the element protected by the
validated signature; a valid signature elsewhere does not authenticate unrelated
elements. Normal XML signatures, SAML assertions or element lookups alone do
not establish a vulnerability. This differs from JWT algorithm confusion,
external entity resolution and recursive entity expansion. Only the complete
`XML signature wrapping attack` phrase is added, with case-insensitive word
boundaries and literal spaces. Standard terms without `attack`, `XSW`, SAML
identifiers, XML element names, API names and protection names remain unchanged.
No Chinese alias or alternative attack spelling is added. Renaming does not
validate signatures or enforce signed-element binding.
Zip Slip terminology follows
[OWASP WSTG Archive Directory Traversal](https://wstg.owasp.org/latest/4-Web_Application_Security_Testing/10-Business_Logic/09-Upload_of_Malicious_Files/).
Unconstrained archive-entry paths can cause extraction outside the intended
destination directory. Ordinary archives, extraction APIs or directory entries
alone do not establish a vulnerability; this is not every path-access problem
or archive resource-exhaustion risk. Only `Zip Slip attack` is added, using
case-insensitive word boundaries and literal spaces. `Zip Slip`, generic archive
words, API names and protection names remain unchanged. Existing Chinese
路径穿越 and 目录遍历漏洞 mappings are retained; no Chinese rule or spelling
alias is added. Renaming does not constrain actual extraction paths.
OAuth authorization-code terminology follows
[RFC 9700 section 4.5](https://www.rfc-editor.org/rfc/rfc9700.html#section-4.5).
This concerns unexpected authorization codes entering a client flow and their
binding to the intended request/session, not executable-code injection or every
OAuth issue. Normal codes, redirects and protocol fields do not establish a
vulnerability. Only the two complete OAuth 2.0 attack phrases above are rules.
Chinese requires the displayed `OAuth2.0` spelling without spaces; English uses
case-insensitive word boundaries and literal spaces. Standard terms without the
attack qualifier, protocol fields, API/protection names and `RFC 9700` remain
unchanged; no versionless or alternative spelling is added. Code-mode spacing
becomes underscores but the dot in `2.0` remains, so this does not guarantee a
valid identifier. Renaming does not configure PKCE or validate session binding.
OAuth mix-up terminology follows
[RFC 9700 section 4.4](https://www.rfc-editor.org/rfc/rfc9700.html#section-4.4).
Confusing authorization-server identity or response origin can send codes or
tokens to an unintended server; normal multi-server configuration alone is not
a vulnerability. This differs from section 4.5 authorization-code injection
and JWT algorithm confusion. Only `OAuth mix-up attack` is added, using
case-insensitive word boundaries, a literal hyphen and literal spaces.
Standard terms without `attack`, issuer/protocol fields and protection names
remain unchanged. No Chinese, `mix up`, `mixup` or underscore aliases are added.
Renaming does not configure issuer verification or bind authorization requests.
Format-string terminology follows
[CWE-134](https://cwe.mitre.org/data/definitions/134.html). The boundary is external
input being interpreted as format control rather than supplied as ordinary data.
Consequences depend on the language, formatter and call site; memory modification
or code execution is not inevitable. Normal formatting, format specifiers and
APIs alone are not vulnerabilities, and this differs from template injection
and log forging. Only the two complete externally-controlled vulnerability
phrases above are rules. Chinese requires the displayed spelling without spaces;
English uses case-insensitive word boundaries and literal spaces. Standard terms
without the vulnerability qualifier, API/protection names and `CWE-134` remain
unchanged. No hyphenated or alternative wording is added. Renaming does not
repair formatting calls or validate their inputs.
DNS-rebinding terminology follows
[CAPEC-275](https://capec.mitre.org/data/definitions/275.html). Changes in a domain's
resolved address can cross the intended browser/resource trust boundary; normal
DNS updates, multiple addresses or cache expiration alone are not attacks.
This differs from DNS cache poisoning and is not every server-request risk.
Only the two complete attack phrases above are rules. Chinese requires uppercase
`DNS` without spaces; English uses case-insensitive word boundaries and literal
spaces. Standard terms without the attack qualifier, DNS fields, API/protection
names and `CAPEC-275` remain unchanged. No rebind, hyphenated or underscore aliases
are added. Renaming does not validate actual resolution results or enforce
network-access boundaries.
HTTP verb-tampering terminology follows
[OWASP WSTG v4.1](https://wstg.owasp.org/v4.1/4-Web_Application_Security_Testing/07-Input_Validation_Testing/03-Testing_for_HTTP_Verb_Tampering/).
It concerns consistent authentication/authorization constraints across HTTP
methods, not every use of multiple methods or support for HEAD/OPTIONS.
This differs from request smuggling, parameter pollution and host-header
injection. Only `HTTP verb tampering attack` is added, using case-insensitive
word boundaries and literal spaces. Standard terms without `attack`, method
names, API/protocol fields, protection names and `WSTG-INPV-03` remain unchanged.
No method-tampering, Chinese or alternative spelling aliases are added.
Renaming does not configure routes or enforce permission checks.
TLS protocol-downgrade terminology follows
[RFC 8446 section 4.1.3](https://www.rfc-editor.org/rfc/rfc8446.html#section-4.1.3)
and [CWE-757](https://cwe.mitre.org/data/definitions/757.html). Interference with
negotiation or fallback can select an unintended weaker protocol version;
normal negotiation, compatibility settings or connection failures alone do not
prove an attack. CWE-757 covers algorithm negotiation more broadly, and the
RFC mechanism has applicability limits rather than covering every downgrade.
Only the two complete TLS protocol attack phrases above are rules. Chinese
requires uppercase `TLS` without spaces; English uses case-insensitive word
boundaries and literal spaces. Standard terms without the attack qualifier,
protocol fields, API/protection names and specification identifiers remain
unchanged. No shorter or alternative aliases are added. Renaming neither
configures TLS versions nor validates handshakes.
HTML-injection terminology follows
[OWASP WSTG HTML Injection](https://wstg.owasp.org/stable/4-Web_Application_Security_Testing/11-Client-side_Testing/03-Testing_for_HTML_Injection/).
Untrusted input interpreted as HTML markup rather than plain text can change
page content or structure. Normal rendering, tags and DOM APIs alone do not
establish a vulnerability. This differs from template-expression execution and
DOM named-property clobbering; script execution is not inevitable, and it is not
an exact synonym for every XSS scenario. Only the two complete attack phrases
above are rules. Chinese requires uppercase `HTML` without spaces; English uses
case-insensitive word boundaries and literal spaces. Standard terms without
`attack`, API/protection names, `XSS` and `WSTG-CLNT-03` remain unchanged.
No alternative spellings are added. Renaming does not encode or sanitize HTML.
Memory-lifetime terminology follows
[CWE-416](https://cwe.mitre.org/data/definitions/416.html) and
[CWE-415](https://cwe.mitre.org/data/definitions/415.html). Using memory after its
release differs from releasing the same memory twice. Normal deallocation,
allocation APIs and variable names alone do not establish a vulnerability;
consequences depend on the implementation and runtime conditions, and code
execution is not inevitable. Only the four complete vulnerability phrases
above are rules. Chinese includes the displayed memory qualifier without
spaces; English uses case-insensitive word boundaries, literal spaces and
the shown `use-after-free` hyphens. Standard terms without the vulnerability
qualifier, `UAF`, allocation/release APIs, protection names and CWE identifiers
remain unchanged. No spelling aliases are added. Renaming does not change
ownership or establish pointer validity.
Memory-boundary terminology follows
[CWE-125](https://cwe.mitre.org/data/definitions/125.html) and
[CWE-787](https://cwe.mitre.org/data/definitions/787.html). Reading outside an
intended buffer differs from writing outside it; both can occur before the
beginning or after the end. These spatial-boundary risks differ from memory
lifetime errors and are not limited to stacks, heaps or one language. Normal
indexing and memory APIs alone are not vulnerabilities; disclosure, crashes or
code execution depend on implementation and runtime conditions. Only the four
complete vulnerability phrases above are added. Chinese includes the memory
qualifier without spaces; English uses case-insensitive word boundaries,
literal spaces and the shown `out-of-bounds` hyphens. Standard terms without
`vulnerability`, API/protection names and CWE identifiers remain unchanged.
No overflow, OOB or alternative spelling aliases are added. Renaming does not
introduce bounds checks or repair memory accesses.
TOCTOU terminology follows
[CWE-367](https://cwe.mitre.org/data/definitions/367.html). A resource's state can
change between checking it and using it, invalidating the earlier check.
This is not every concurrency, data-race or memory-lifetime issue; normal checks,
file access or threading alone do not establish a vulnerability. Shortening the
interval does not eliminate the underlying problem. Only the complete
`TOCTOU race condition vulnerability` phrase is added, using case-insensitive
word boundaries and literal spaces. Standard terms without `vulnerability`,
file/synchronization APIs, protection names and `CWE-367` remain unchanged.
No Chinese or expanded-name aliases are added. Renaming does not acquire locks
or make operations atomic.
Integer-range terminology follows
[CWE-190](https://cwe.mitre.org/data/definitions/190.html) and
[CWE-191](https://cwe.mitre.org/data/definitions/191.html). Exceeding an integer's
upper bound differs from falling below its representable minimum. Wrapping,
exceptions or other results depend on language, type and runtime conditions;
normal arithmetic or intentional controlled wrapping alone is not a
vulnerability. These are not memory-buffer boundaries or floating-point
underflow. Only the four complete vulnerability phrases above are added.
Chinese requires the displayed spelling without spaces; English uses
case-insensitive word boundaries and literal spaces. Standard terms without
`vulnerability`, numeric types, API/protection names and CWE identifiers remain
unchanged. No alternative spellings are added. Renaming does not add arithmetic
checks or change numeric types.
Type-confusion terminology follows
[CWE-843](https://cwe.mitre.org/data/definitions/843.html). A resource created or
initialized with one type is accessed using an incompatible type, causing
unintended interpretation or operations. Normal casts, generics or dynamic
typing alone are not vulnerabilities. This is not limited to C/C++, does not
guarantee code execution, and is distinct from integer-range, memory-lifetime
or bounds errors even when those occur as consequences. Only the two complete
vulnerability phrases above are added. Chinese requires the displayed spelling
without spaces; English uses case-insensitive word boundaries and literal spaces.
Standard terms without `vulnerability`, type APIs, protection names and `CWE-843`
remain unchanged. No spelling aliases are added. Renaming does not insert type
checks or establish type safety.
Null-pointer terminology follows
[CWE-476](https://cwe.mitre.org/data/definitions/476.html). A pointer expected to
be valid is dereferenced when it is actually NULL, potentially causing an
exception, crash or availability impact. Outcomes depend on the implementation;
normal null checks, null representations and API names alone are not
vulnerabilities. This differs from use-after-free, type confusion and bounds
errors, and does not cover every null/nil or exception-handling scenario.
Only the two complete vulnerability phrases above are added. Chinese requires
the displayed spelling without spaces; English uses case-insensitive word
boundaries and literal spaces. Standard terms without `vulnerability`, null
representations, API/protection names and `CWE-476` remain unchanged. No spelling
aliases are added. Renaming does not insert null checks or validate pointers.
Uncontrolled-recursion terminology follows
[CWE-674](https://cwe.mitre.org/data/definitions/674.html). Recursion without
appropriate limits or termination can consume excessive stack, memory or CPU;
normal recursion, bounded traversal and recursion APIs alone are not
vulnerabilities. Outcomes depend on implementation and input conditions. This
is not every infinite loop, regex-complexity or XML entity-expansion issue.
Only the two complete vulnerability phrases above are added. Chinese requires
the displayed spelling without spaces; English uses case-insensitive word
boundaries and literal spaces. Standard terms without `vulnerability`, depth
settings, API/protection names and `CWE-674` remain unchanged. No infinite/unbounded
recursion aliases are added. Renaming does not add termination or depth limits.
Uninitialized-variable terminology follows
[CWE-457](https://cwe.mitre.org/data/definitions/457.html). Using a variable before
initialization can produce unintended results, depending on language, type and
execution conditions. A declaration without an explicit initializer is not
necessarily vulnerable: the language may provide deterministic default
initialization. This is not every uninitialized resource, null-pointer or
use-after-free issue. Only the two complete vulnerability phrases above are
added; the English phrase is project-selected wording, not the verbatim CWE
title. Chinese requires the displayed spelling without spaces; English uses
case-insensitive word boundaries and literal spaces. Standard terms, alternate
word orders, API/type/protection names and CWE identifiers remain unchanged.
No spelling aliases are added. Renaming does not initialize variables or insert
data-flow checks.
OAuth refresh-token replay terminology follows
[RFC 9700 section 4.14](https://www.rfc-editor.org/rfc/rfc9700.html#section-4.14),
particularly section 4.14.2. It concerns unintended use and replay of exposed
refresh tokens, not every repeated refresh request. Normal token refresh,
rotation APIs and protocol fields alone are not vulnerabilities. The RFC's
specific public-client requirements are not a single implementation prescription
for all clients. This differs from authorization-code injection, server mix-up
and JWT algorithm confusion. Only the two complete OAuth attack phrases above
are added. Chinese requires the displayed `OAuth` casing without spaces;
English uses case-insensitive word boundaries and literal spaces. Standard
terms, protocol/API/protection names and `RFC 9700` remain unchanged. No shorter
or spelling aliases are added. Renaming does not validate tokens, configure
rotation or apply sender constraints.
Padding-oracle terminology follows OWASP WSTG's
[Padding Oracle](https://wstg.owasp.org/latest/4-Web_Application_Security_Testing/09-Weak_Cryptography/02-Padding_Oracle/)
(WSTG-CRYP-02). Observable errors, response differences or timing may expose
whether ciphertext padding is valid; normal padding, a decryption failure or an
exception name alone does not establish a vulnerability. This is distinct from
TLS downgrade, JWT algorithm confusion and Oracle databases. Only the complete
English attack phrase above is added, with case-insensitive word boundaries and
literal spaces. Standard terms such as `padding oracle`, `CBC`, `PKCS7` and
`BadPaddingException` remain unchanged; no Chinese, hyphenated, underscored or
plural attack aliases are added. Renaming does not unify decryption errors or
implement authenticated encryption.
Reverse-tabnabbing terminology follows OWASP Community's
[Reverse Tabnabbing](https://community.owasp.org/attacks/Reverse_Tabnabbing),
using its search-index summaries; the full page was not retrievable during
research. A newly opened page may navigate its opener to an untrusted page when
an opener reference and the necessary navigation capability remain available.
Modern browsers give `target="_blank"` links implicit `noopener` behavior; older
browsers and other opening methods need separate consideration. This is not a
claim that every new-tab link is vulnerable or that `window.open` always has the
same defaults. It differs from open redirects and DOM clobbering and does not
imply arbitrary cross-origin reading of the opener. Only the complete English
attack phrase above is added with case-insensitive word boundaries and literal
spaces. `reverse tabnabbing`, API/property names and protection terms remain
unchanged; no Chinese, spelling or plural aliases are added. Renaming does not
isolate the opener, add `rel` attributes or implement phishing protection.
XSSI terminology follows OWASP WSTG's
[Testing for Cross Site Script Inclusion](https://wstg.owasp.org/stable/4-Web_Application_Security_Testing/11-Client-side_Testing/13-Testing_for_Cross_Site_Script_Inclusion/)
(WSTG-CLNT-13), using its search-index summaries; the full page was not
retrievable during research. Cross-origin script inclusion may expose sensitive
data under particular conditions; ordinary public script loading, JSONP names,
JSON/CSV files or a logged-in session alone do not establish a vulnerability.
This concerns unintended data exposure, not injecting untrusted code into a
page as in XSS, cross-site WebSocket hijacking or reverse tabnabbing. It does
not imply that every cross-origin script can read private data. Only the two
complete attack phrases above are added: Chinese uses the fixed spelling and
English uses case-insensitive word boundaries and literal spaces. `XSSI`,
`cross site script inclusion`, 跨站脚本包含, protocol fields, API/protection names
and `WSTG-CLNT-13` remain unchanged. No hyphenated, underscored, acronym-only or
plural attack aliases are added. Renaming does not configure authentication,
response policies or resource isolation.
CSS-injection terminology follows OWASP WSTG's
[CSS Injection](https://wstg.owasp.org/latest/4-Web_Application_Security_Testing/11-Client-side/05-CSS_Injection/)
(WSTG-CLNT-05), using its search-index summaries; the full page was not retrieved
during research. This concerns untrusted CSS interpreted in a trusted site's
style context, not ordinary style settings or the presence of CSS APIs. Style
manipulation or data exposure depends on the injection point, page content,
policies and browser. Legacy-browser script-execution examples are not a claim
that CSS always executes JavaScript in modern browsers. This is distinct from
HTML injection, XSSI and arbitrary script execution. Only the two complete
attack phrases above are added. Chinese requires `CSS` casing and no spaces;
English uses case-insensitive word boundaries and literal spaces. `CSS`,
`CSS injection`, `CSS注入`, API/field/protection names and `WSTG-CLNT-05` remain
unchanged; no stylesheet, translated, hyphenated, underscored or plural attack
aliases are added. Renaming does not sanitize CSS or configure CSP.
SSI-injection terminology follows OWASP WSTG's
[Testing for SSI Injection](https://wstg.owasp.org/v4.2/4-Web_Application_Security_Testing/07-Input_Validation_Testing/08-Testing_for_SSI_Injection/)
(WSTG-INPV-08), using its search-index summaries; the full page was not
retrievable during research. This concerns untrusted input interpreted as
server-side include directives in an SSI-processing context. Normal SSI use,
trusted directives, file extensions or configuration names alone do not prove a
vulnerability. Information exposure, file inclusion or command execution depends
on server configuration, enabled directives, permissions and context; command
execution is not inevitable. This is distinct from server-side template
injection, browser XSSI and ordinary module imports. Only the two complete attack
phrases above are added. Chinese requires `SSI` casing and no spaces; English
uses case-insensitive word boundaries and literal spaces. `SSI`, `SSI injection`,
`Server-Side Includes`, `SSI注入`, configuration/API/protection names and
`WSTG-INPV-08` remain unchanged; no full-name, translated, hyphenated, underscored
or plural attack aliases are added. Renaming does not disable SSI, validate input
or change server permissions.
IDN-homograph terminology follows ICANN's historical
[IDN Homograph Attacks statement](https://www.icann.org/en/announcements/details/icann-statement-on-idn-homograph-attacks-and-request-for-public-comment-23-2-2005-en)
(2005-02-23) and Unicode's
[UTS #39, Revision 29](https://www.unicode.org/reports/tr39/tr39-29.html), using
search-index summaries rather than retrieved full texts. These support the
concept of visual domain-name impersonation, not current browser behavior or
complete compliance with the latest standard. Normal IDNs, Unicode characters,
Punycode or mixed scripts alone do not establish abuse. Confusability depends
on display context and is not limited to mixed-script strings. This phrase is
restricted to IDN domain impersonation, not all lookalike identifiers or package
typosquatting. Only the complete English attack phrase above is added with
case-insensitive word boundaries and literal spaces. `IDN`, `IDNA`, `Unicode`,
`Punycode`, `UTS #39`, `IDN homograph` and API/protection names remain unchanged;
no Chinese, unqualified, homoglyph, hyphenated, underscored or plural attack
aliases are added. Renaming does not detect confusables, normalize domains,
perform IDNA/Punycode conversion or change browser display policies.
OS-command-argument terminology follows the Argument Injection section of
OWASP's [OS Command Injection Defense Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/OS_Command_Injection_Defense_Cheat_Sheet.html),
using its search-index summaries; the full page was not retrievable during
research. This concerns unintended interpretation of untrusted OS command
arguments or options, not ordinary argument passing, HTTP parameter pollution,
database parameters or general function arguments. Outcomes depend on the
invoked command, supported options, privileges and input context; arbitrary
command execution is not inevitable. Shell-metacharacter escaping alone does
not resolve every argument-injection risk. Only the two complete OS-qualified
attack phrases above are added, with fixed Chinese spelling and case-insensitive
English word boundaries and literal spaces. Standard terms, API names such as
`exec.Command` and `ProcessStartInfo.ArgumentList`, and `CWE-88` remain unchanged.
No shorter, hyphenated, underscored or plural attack aliases are added. Renaming
does not validate arguments, enforce allow-lists or isolate arguments or privileges.
Session-puzzling terminology follows OWASP WSTG's
[Testing for Session Puzzling](https://owasp.org/www-project-web-security-testing-guide/latest/4-Web_Application_Security_Testing/06-Session_Management_Testing/08-Testing_for_Session_Puzzling)
(WSTG-SESS-08), using search-index summaries; the full page was not retrievable
during research. This concerns session variables used for inconsistent purposes,
where state set in one context is incorrectly trusted in another. Ordinary
session storage, variable reuse or state shared across requests alone does not
establish a vulnerability. Impact depends on control flow and authorization
checks, not inevitable authentication bypass. This differs from session fixation,
token theft or cryptographic confusion. Only the complete English attack phrase
above is added with case-insensitive word boundaries and literal spaces.
`session puzzling`, `session variable overloading`, API/protection names and
`WSTG-SESS-08` remain unchanged; no Chinese, synonymous, hyphenated, underscored
or plural attack aliases are added. Renaming does not separate session variables,
validate authorization or rotate session identifiers.
Expression-language terminology follows OWASP Community's
[Expression Language Injection](https://community.owasp.org/vulnerabilities/Expression_Language_Injection),
using search-index summaries rather than retrieved full text. This concerns
untrusted data interpreted as expression language, not ordinary evaluation of
trusted expressions or the presence of an API or framework. Data exposure or
code execution depends on interpreter capabilities, accessible objects,
permissions and evaluation context; arbitrary code execution is not inevitable.
This may overlap with template injection but does not equate every template,
XPath or SQL injection with EL injection. Only the two complete attack phrases
above are added: fixed Chinese spelling, and case-insensitive English word
boundaries with literal spaces. `expression language injection`, 表达式语言注入,
`EL`, `SpEL`, `OGNL`, `UEL`, API/protection names and `CWE-917` remain unchanged.
No acronym, shorter, hyphenated, underscored or plural attack aliases are added.
Renaming does not validate expressions, configure a sandbox or restrict interpreter
permissions, and makes no vulnerability claim about a particular framework version.
Reflected-file-download terminology follows the OWASP-hosted
[Less Known Web Application Vulnerabilities](https://wiki.owasp.org/images/4/4e/OWASP-Ionut-Popescu-Less-Known-Web-Application-Vulnerabilities-Stripped.pdf)
presentation and Oren Hafif's historical SpiderLabs research,
[Reflected File Download](https://www.levelblue.com/blogs/spiderlabs-blog/reflected-file-download-a-new-web-attack-vector)
(2014-10-30), using search-index summaries rather than retrieved full texts.
Reflected input may be returned as downloadable content with an attacker-influenced
filename or extension; normal downloads, reflected text or a response header alone
do not establish a vulnerability. Risk depends on content control, download/name
handling and use context. Downloading does not imply automatic execution; the
historical execution scenario also requires the user to run the file. This is
not XSSI, HTML injection or an open redirect and makes no current-browser or
complete-header-protection claim. Only the complete English attack phrase above
is added with case-insensitive word boundaries and literal spaces. Standard terms,
API/protection names and response fields remain unchanged; no Chinese, acronym,
hyphenated, underscored or plural attack aliases are added. Renaming does not set
headers, validate filenames, sanitize download contents or prevent execution.
CORS-misconfiguration terminology follows OWASP WSTG's
[Cross Origin Resource Sharing](https://wstg.owasp.org/latest/4-Web_Application_Security_Testing/11-Client-side/07-Cross_Origin_Resource_Sharing/)
(WSTG-CLNT-07), using search-index summaries; the full page was not retrievable
during research. This concerns configurations permitting unintended origins to
read responses, with impact depending on resource sensitivity, allowed origins,
authentication/credential modes and browser checks. Normal CORS use, header names
or wildcard origins for public resources do not alone establish a vulnerability;
wildcards combined with credentials do not imply inevitable data exposure.
CORS is not authentication, authorization, CSRF protection or WebSocket origin
validation. A browser refusing response access does not mean no request was sent.
Only the two complete vulnerability phrases above are added. Chinese requires
`CORS` casing and no spaces; English uses case-insensitive word boundaries and
literal spaces. Standard terms, header/API/protection names and `WSTG-CLNT-07`
remain unchanged; no translated, shorter, hyphenated, underscored or plural aliases
are added. Renaming does not configure response headers, origin allow-lists,
credentials or authorization.
XQuery-injection terminology follows MITRE's
[CWE-652](https://cwe.mitre.org/data/definitions/652.html), cross-referenced by
[OWASP A03:2021 Injection](https://top10.owasp.org/A03_2021-Injection/), using
search-index summaries rather than retrieved full text. This concerns external
input altering XQuery expression structure, not ordinary queries, trusted
expressions, XML data or query APIs. Impact depends on query context, database
permissions and application logic; arbitrary code execution or authentication
bypass is not inevitable. XQuery and XPath are related but not interchangeable,
and XML external entities or general EL evaluation are separate concepts.
Only the two complete attack phrases above are added: Chinese requires `XQuery`
casing and no spaces; English uses case-insensitive word boundaries and literal
spaces. Standard terms, `XQuery`, `XPath`, `XML`, `FLWOR`, query API/protection
names and `CWE-652` remain unchanged. No spelling, hyphenated, underscored or
plural attack aliases are added. Renaming does not bind query parameters,
validate input or restrict database permissions.
PHP-object-injection terminology follows OWASP Community's
[PHP Object Injection](https://community.owasp.org/vulnerabilities/PHP_Object_Injection),
using search-index summaries rather than retrieved full text or PHP-version
validation. This concerns untrusted deserialization input and resulting object
behavior; exploitation depends on available classes, dangerous magic methods or
object call chains, loading conditions, permissions and application context.
Normal object creation, trusted deserialization or an API/magic-method name alone
does not prove a vulnerability. Arbitrary code execution is not inevitable, and
neither generic sanitization nor a single option is claimed as complete protection.
This is distinct from prototype pollution, mass assignment and EL injection.
Only the two complete PHP-qualified attack phrases above are added: fixed `PHP`
casing and no spaces in Chinese, case-insensitive word boundaries and literal
spaces in English. Standard terms, `unserialize`, magic-method names,
`allowed_classes` and API/protection names remain unchanged. No unqualified,
translated, hyphenated, underscored or plural attack aliases are added. Renaming
does not validate serialized input, disable magic methods or configure class allow-lists.
TLS-certificate-hostname terminology follows MITRE's
[CWE-297](https://cwe.mitre.org/data/definitions/297.html), using search-index
summaries rather than retrieved full text or complete TLS-specification validation.
This concerns failure to verify that a certificate identifies the intended host;
a trusted signature or certificate chain alone does not establish that match.
Normal validation, rejection of a mismatched connection or an API/configuration
name alone does not prove a vulnerability. Impact depends on client settings,
connection paths and attack conditions; a mismatch does not prove a hijack.
This is distinct from expired certificates, untrusted CAs, TLS downgrade and IDN
character confusion. Only the two complete TLS-qualified vulnerability phrases
above are added. Chinese requires `TLS` casing and no spaces; English uses
case-insensitive word boundaries and literal spaces. Standard terms, `VerifyHostname`,
`ServerName`, `HostnameVerifier`, `InsecureSkipVerify`, other API/protection names
and `CWE-297` remain unchanged; no unqualified, verification-spelling, hyphenated,
underscored or plural aliases are added. Renaming does not validate certificates,
configure hostnames or trust stores, or change TLS connections.
Externally-controlled-reflection terminology follows MITRE's
[CWE-470](https://cwe.mitre.org/data/definitions/470.html), using search-index
summaries rather than retrieved full text or language/framework-version validation.
This concerns insufficiently restricted external input selecting classes or code
through reflection. Ordinary reflection, trusted class names, allow-listed plugin
selection or API names alone do not prove a vulnerability. Code execution,
authorization bypass or crashes depend on available classes/methods, permissions
and call context; none is inevitable. This is not synonymous with all dynamic
loading, deserialization, expression evaluation or reflected file downloads.
Only the two complete vulnerability phrases above are added: fixed Chinese
spelling and case-insensitive English word boundaries with literal spaces.
`unsafe reflection`, `externally controlled reflection`, API/protection names
and `CWE-470` remain unchanged; no synonymous, hyphenated, underscored or plural
aliases are added. Renaming does not constrain class/method selection or configure
allow-lists or permissions.
Untrusted-search-path terminology follows MITRE's
[CWE-426](https://cwe.mitre.org/data/definitions/426.html), using search-index
summaries rather than retrieved full text or platform-loader validation.
This concerns externally controlled search paths selecting critical resources
outside the program's control. Control of the search path itself differs from
CWE-427's writable location within an otherwise fixed path. Ordinary PATH
settings, trusted plugin directories, relative paths or resource-lookup APIs
alone do not establish a vulnerability. Impact depends on attacker control,
lookup/loading order, available resources and privileges; code execution or
privilege escalation is not inevitable. This does not equate all DLL hijacking,
directory traversal or externally controlled reflection with this weakness.
Only the two complete vulnerability phrases above are added: fixed Chinese
spelling and case-insensitive English word boundaries with literal spaces.
Standard terms, environment variables, API/protection names and `CWE-426`/`CWE-427`
remain unchanged; no synonymous, hyphenated, underscored or plural aliases are
added. Renaming does not change environment variables, validate paths or configure
directory permissions or loading order.
Cryptographically-weak-PRNG terminology follows MITRE's
[CWE-338](https://cwe.mitre.org/data/definitions/338.html), using search-index
summaries rather than retrieved full text or algorithm-implementation validation.
This concerns a generator lacking cryptographic strength used in a security
context. Ordinary simulations, tests, sampling or other non-security PRNG uses
are not automatically vulnerabilities. API names, determinism or fixed seeds
alone do not establish application risk; purpose and implementation matter.
This is not synonymous with all entropy, seed-management or repeated-value
problems, and predictable output does not guarantee key recovery or account
compromise. Only the complete English vulnerability phrase above is added with
case-insensitive word boundaries and literal spaces. `PRNG`, `CSPRNG`, standard
terms, API/protection names and `CWE-338` remain unchanged; no Chinese, expanded,
shorter, hyphenated, underscored or plural aliases are added. Renaming does not
replace generators, add entropy or manage seeds.
postMessage-origin-validation terminology follows OWASP WSTG's
[Web Messaging](https://wstg.owasp.org/latest/4-Web_Application_Security_Testing/11-Client-side/11-Web_Messaging/)
(WSTG-CLNT-11), using search-index summaries rather than retrieved full text or
browser-implementation validation. This concerns missing or incorrect receiver
origin checks causing unintended messages to be trusted, not ordinary cross-window
communication or an API name alone. Impact depends on message purpose, receiver
authorization/data handling and reachable contexts; XSS or authentication bypass
is not inevitable. Receiver origin validation differs from sender `targetOrigin`
restrictions and message-content validation, and replaces neither content checks
nor business authorization. It is not the same mechanism as CORS or WebSocket
origin validation. Only the two complete vulnerability phrases above are added:
fixed `postMessage` casing and no spaces in Chinese, case-insensitive word
boundaries and literal spaces in English. Standard terms, event properties,
API/protection names and `WSTG-CLNT-11` remain unchanged; no unqualified,
verification-spelling, hyphenated, underscored or plural aliases are added.
Renaming does not validate origins, configure `targetOrigin`, check messages
or enforce authorization.
Unrestricted-file-upload terminology follows OWASP's
[Unrestricted File Upload](https://community.owasp.org/vulnerabilities/Unrestricted_File_Upload),
[File Upload Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/File_Upload_Cheat_Sheet.html)
and MITRE's [CWE-434](https://cwe.mitre.org/data/definitions/434.html), using search-index
summaries rather than retrieved full text or server-implementation validation.
This concerns inadequate upload restrictions allowing dangerous content into
storage or processing; CWE-434 specifically concerns dangerous file types and
is not synonymous with every upload issue. Impact depends on file type, storage,
processing components and permissions. Accepting an upload does not prove code
execution. Normal uploads, extensions, MIME types or API names alone do not prove
a vulnerability. MIME/signature checks alone are not claimed as complete protection,
nor does storage outside the webroot eliminate all processing risks. Only the two
complete vulnerability phrases above are added: fixed Chinese spelling and
case-insensitive English word boundaries with literal spaces. Standard terms,
API/protection names and `CWE-434` remain unchanged; no synonymous, shorter,
hyphenated, underscored or plural aliases are added. Renaming does not validate
file types or configure storage permissions, upload limits or scanners.
Signed-to-unsigned-conversion terminology follows MITRE's
[CWE-195](https://cwe.mitre.org/data/definitions/195.html), using search-index
summaries rather than retrieved full text or complete language/platform validation.
This concerns unexpected results when converting signed values to unsigned types;
negative values used as unsigned lengths may affect subsequent allocation or memory
operations. Results depend on language, type width, checks and later use. Normal
controlled conversions, type names, explicit casts or negative test values alone
do not establish a vulnerability. Not every conversion is undefined behavior or
necessarily causes an out-of-bounds access. This differs from arithmetic overflow,
underflow, unsigned-to-signed conversion, truncation and general type confusion.
Only the complete English vulnerability phrase above is added with case-insensitive
word boundaries and literal spaces. Standard terms, type/API/protection names and
`CWE-195` remain unchanged; no Chinese, reverse-direction, shorter, hyphenated,
underscored or plural aliases are added. Renaming does not check ranges, change
numeric types or add conversion guards.
Double-checked-locking terminology follows MITRE's
[CWE-609](https://cwe.mitre.org/data/definitions/609.html), using search-index
summaries rather than retrieved full text or complete memory-model validation.
Insufficient synchronization may let readers bypassing a lock observe an
incompletely initialized object. Preventing duplicate initialization does not
alone guarantee safe publication or visibility. Correctness depends on language,
version, memory model, synchronization and access paths; not every implementation
of the pattern is unsafe. The source's Java 5+ `volatile` discussion is not a claim
that adding `volatile` fixes this in every language. Normal lazy initialization,
locks, atomic variables or API names alone do not establish a vulnerability.
Only the complete English vulnerability phrase above is added with case-insensitive
word boundaries and literal spaces. Standard terms, `DCL`, `volatile`, `sync.Once`,
`call_once`, synchronization/API/protection names and `CWE-609` remain unchanged;
no Chinese, unhyphenated, acronym-only, underscored or plural aliases are added.
Renaming does not add locks, insert memory barriers or establish happens-before.
Unchecked-return-value terminology follows MITRE's
[CWE-252](https://cwe.mitre.org/data/definitions/252.html), using search-index
summaries rather than retrieved full text or validation of every API contract.
Ignoring a function or method result may leave a failure or unexpected state
undetected. Risk depends on result semantics, caller assumptions and error
handling; deliberately discarding an irrelevant result, a void function or an
API name alone does not establish a vulnerability. This differs from incorrectly
checking a return value (CWE-253), general exception handling or all null-pointer
problems. Crashes, privilege violations or code execution are not inevitable.
Only the two complete vulnerability phrases above are added: fixed Chinese
spelling and case-insensitive English word boundaries with literal spaces.
Standard terms, `nodiscard`, `warn_unused_result`, API/protection names and
`CWE-252`/`CWE-253` remain unchanged; no synonymous, shorter, hyphenated,
underscored or plural aliases are added. Renaming does not check results,
add error handling or implement rollback.
Replacement phrases are project-specific display wording, not standard
synonyms. Renaming does not remediate vulnerabilities or guarantee unchanged
program behavior, and is not a way to bypass safety review.

Java `.java` files already support line comments, block comments and Javadoc.
`-keep-doc` preserves Javadoc; comment markers in strings, character literals
and text blocks remain intact. CLI regression tests compile the stripped
fixture with Java 17 when `javac` is available.

Standard vulnerability names such as SQL注入, XSS and CSRF are deliberately not
in the table: they are the precise term, and replacing them would make the code
harder to read. The standalone defensive terms 蜜罐, 取证, 应急响应, 零信任 and
蓝队 are also left alone.

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

## Guards on writing

Three checks stand between a rewrite and your files.

**Credential files are skipped.** `.env`, `*.pem`, `*.key`, `id_rsa`,
`secrets.yml`, `credentials.json`, `terraform.tfvars`, anything under `.ssh/`,
`.aws/`, `.kube/` and the like. The reason is named in the report:

```
SKIP secrets.yml: credential file (secrets.yml), use -allow-secrets
```

Neither `-lang` nor naming the file on the command line lifts this; only
`-allow-secrets` does. `-stdout` is guarded too, since printing a key still
puts it somewhere you did not ask for.

**Every rewrite is verified before it lands.** The file is compared against
its original: if the bracket balance outside comments changed, a string or
comment was left unterminated, or a Go file stopped parsing, the write is
refused and the file reported as a `SKIP`. Only the difference is judged, so a
file that was already unbalanced is not blamed on the tool. `-no-verify`
turns this off.

**Nothing is written without `-w`.** `-diff` shows the exact change, `-check`
makes CI fail when anything would change, and both refuse to run with `-w`:

```sh
strip-comments -diff .            # review the change
strip-comments -check .           # exit 1 if anything would change
strip-comments -w -backup .       # apply, keeping .bak copies
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
