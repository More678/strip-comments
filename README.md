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
Replacement phrases are project-specific display wording, not standard
synonyms. Renaming does not remediate vulnerabilities or guarantee unchanged
program behavior, and is not a way to bypass safety review.

Java `.java` files already support line comments, block comments and Javadoc.
`-keep-doc` preserves Javadoc; comment markers in strings, character literals
and text blocks remain intact. CLI regression tests compile the stripped
fixture with Java 17 when `javac` is available.

Standard vulnerability names such as SQL注入, XSS and CSRF are deliberately not
in the table: they are the precise term, and replacing them would make the code
harder to read. Defensive vocabulary (蜜罐, 取证, 应急响应, 零信任, 蓝队) is
left alone too, since it is already the wording the table rewrites towards.

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
