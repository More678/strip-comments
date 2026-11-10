# strip-comments

Remove every comment from the source files in a folder: bat, Go, Rust, C/C++,
Java, Python, shell, JavaScript and 50-odd other languages.

It lexes each file instead of matching patterns, so comment markers inside
strings, raw strings, heredocs and regex literals are left alone:
`"http://x"`, Go backticks, C++ `R"x( // )x"`, Rust `r#"//"#`, Java text
blocks, shell heredocs and `/\/\//g` all survive.

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
| `-diff` | print a unified diff of every change, write nothing |
| `-check` | exit non-zero if any file would change, write nothing (for CI) |
| `-no-verify` | skip the structural check that runs before writing |
| `-allow-secrets` | process files that look like credentials |

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
