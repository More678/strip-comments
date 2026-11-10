// Package glossary rewrites terminology in source code and comments: each
// rule maps a term to its replacement, so a codebase can be restated in the
// vocabulary of threat detection and compliance assessment.
//
// Replacement is off unless the caller asks for it. Renaming an identifier can
// break references in files this run never looked at, so the default is to
// report hits and change nothing.
package glossary

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Rule maps From to To.
type Rule struct {
	From string
	To   string
	// Word requires ASCII word boundaries around From, so "pentest" does not
	// match inside "pentester". It is ignored when From starts or ends with a
	// non-ASCII rune, because CJK text has no word characters to delimit.
	Word bool
	// CI matches From case-insensitively and carries the original casing over
	// to To (SCAN -> DETECT, Scan -> Detect, scan -> detect).
	CI bool
}

// Hit is one replacement found in a buffer.
type Hit struct {
	Start, End int    // byte range of the matched term
	Line       int    // 1-based line number
	From, To   string // the text that matched and what replaces it
	InComment  bool   // the match sits inside a comment
}

// Default restates offensive-security vocabulary as threat detection and
// compliance assessment. Longer terms come first so "渗透测试" wins over
// "渗透"; Rules() enforces that regardless of the order written here.
//
// Override the whole table with a user file rather than editing this list.
var Default = []Rule{
	// Chinese: no word boundary, CJK has no word characters.
	{From: "渗透测试", To: "威胁检测"},
	{From: "渗透攻击", To: "威胁检测"},
	{From: "渗透", To: "威胁检测"},
	{From: "攻击载荷", To: "检测样本"},
	{From: "攻击面", To: "暴露面"},
	{From: "漏洞利用", To: "风险验证"},
	{From: "提权", To: "权限核查"},
	{From: "后门", To: "异常通道"},
	{From: "木马", To: "可疑程序"},
	{From: "暴力破解", To: "口令强度核查"},
	{From: "绕过", To: "合规偏差"},
	{From: "等级保护", To: "等保分析"},
	{From: "等保测评", To: "等保分析"},

	// Attack stages and techniques. Ambiguous everyday words are only added
	// as compounds: 爆破 alone also means demolition, 社工 a social worker,
	// 中间人 a broker, 钓鱼 fishing.
	{From: "内网穿透", To: "内网连通性检测"},
	{From: "横向移动", To: "横向流量分析"},
	{From: "权限维持", To: "持久化检测"},
	{From: "数据窃取", To: "数据外泄检测"},
	{From: "反弹shell", To: "异常外联检测"},
	{From: "反弹Shell", To: "异常外联检测"},
	{From: "命令与控制", To: "异常外联检测"},
	{From: "免杀", To: "检测规避分析"},
	{From: "嗅探", To: "流量监测"},
	{From: "中间人攻击", To: "链路劫持检测"},
	{From: "社会工程学", To: "安全意识评估"},
	{From: "钓鱼邮件", To: "可疑邮件"},
	{From: "钓鱼网站", To: "仿冒网站"},
	{From: "钓鱼攻击", To: "仿冒攻击检测"},
	{From: "撞库", To: "凭据复用检测"},
	{From: "口令爆破", To: "口令强度核查"},
	{From: "密码爆破", To: "口令强度核查"},

	// Vulnerability classes and recon. Generic phrases are only added as
	// compounds: 目录遍历 alone is also an ordinary directory walk, 代码执行
	// an ordinary "code runs", 红队 a sports team.
	{From: "命令注入", To: "命令执行风险核查"},
	{From: "远程代码执行", To: "远程执行风险核查"},
	{From: "目录遍历漏洞", To: "路径访问核查"},
	{From: "路径穿越", To: "路径访问核查"},
	{From: "反序列化漏洞", To: "反序列化风险"},
	{From: "越权访问", To: "权限越界核查"},
	{From: "未授权访问", To: "访问控制核查"},
	{From: "弱口令", To: "口令强度核查"},
	{From: "信息泄露", To: "信息暴露核查"},
	{From: "提权漏洞", To: "权限风险"},
	{From: "供应链攻击", To: "供应链风险评估"},
	{From: "端口扫描", To: "端口开放核查"},
	{From: "漏洞扫描", To: "风险扫描"},
	{From: "资产探测", To: "资产梳理"},
	{From: "零日漏洞", To: "未公开风险"},
	{From: "漏洞挖掘", To: "风险发现"},
	{From: "攻防演练", To: "安全演练"},
	{From: "红队测试", To: "对抗性评估"},
	{From: "密码喷洒", To: "批量认证风险核查"},
	{From: "口令喷洒", To: "批量认证风险核查"},
	{From: "浏览器会话劫持", To: "浏览器会话完整性核查"},
	{From: "远程会话劫持", To: "远程会话完整性核查"},

	// English: word boundaries, case preserved.
	{From: "penetration test", To: "threat detection", Word: true, CI: true},
	{From: "pentest", To: "threat detection", Word: true, CI: true},
	{From: "exploit", To: "risk validation", Word: true, CI: true},
	{From: "payload", To: "detection sample", Word: true, CI: true},
	{From: "backdoor", To: "anomalous channel", Word: true, CI: true},
	{From: "privilege escalation", To: "privilege review", Word: true, CI: true},
	{From: "bruteforce", To: "credential strength check", Word: true, CI: true},
	{From: "attack surface", To: "exposure surface", Word: true, CI: true},
	{From: "lateral movement", To: "lateral traffic analysis", Word: true, CI: true},
	{From: "data exfiltration", To: "data egress detection", Word: true, CI: true},
	{From: "webshell", To: "suspicious script", Word: true, CI: true},
	{From: "reverse shell", To: "outbound anomaly", Word: true, CI: true},
	// C2 is two bytes, so the word boundary is what keeps it off C2H5OH,
	// 0xC2 and mC2. Acronyms are matched case-sensitively: carrying the case
	// of "C2" would shout "OUTBOUND ANOMALY", and a lowercase c2 in code is
	// usually a variable, not a command-and-control server.
	{From: "C2", To: "outbound anomaly", Word: true},
	{From: "phishing", To: "impersonation detection", Word: true, CI: true},
	{From: "sniffer", To: "traffic monitor", Word: true, CI: true},
	{From: "MITM", To: "link hijack detection", Word: true},
	{From: "credential stuffing", To: "credential reuse detection", Word: true, CI: true},

	// v1.3.0: vulnerability classes, recon and tooling. Acronyms are
	// case-sensitive for the same reason as C2; the word boundary keeps RCE
	// off FORCE and SOURCE. "red team" alone also names a sports side, so
	// only the activity form is added.
	{From: "RCE", To: "remote execution risk", Word: true},
	{From: "SSRF", To: "server request review", Word: true},
	{From: "LFI", To: "file inclusion review", Word: true},
	{From: "privilege bypass", To: "access control review", Word: true, CI: true},
	{From: "port scan", To: "open port review", Word: true, CI: true},
	{From: "port scanning", To: "open port review", Word: true, CI: true},
	{From: "port scanner", To: "open port auditor", Word: true, CI: true},
	{From: "vulnerability scan", To: "risk scan", Word: true, CI: true},
	{From: "vulnerability scanner", To: "risk scanner", Word: true, CI: true},
	{From: "zero-day", To: "undisclosed risk", Word: true, CI: true},
	{From: "0day", To: "undisclosed risk", Word: true, CI: true},
	{From: "red teaming", To: "adversarial assessment", Word: true, CI: true},
	{From: "exploit kit", To: "risk validation toolkit", Word: true, CI: true},
	{From: "rootkit", To: "kernel-level anomaly", Word: true, CI: true},
	{From: "keylogger", To: "input monitor", Word: true, CI: true},
	{From: "password spraying", To: "bulk authentication review", Word: true, CI: true},
	{From: "password spray", To: "bulk authentication review", Word: true, CI: true},
	{From: "browser session hijacking", To: "browser session integrity review", Word: true, CI: true},
	{From: "remote service session hijacking", To: "remote session integrity review", Word: true, CI: true},
	{From: "黄金票据攻击", To: "Kerberos票据伪造风险核查"},
	{From: "票据传递攻击", To: "Kerberos票据复用风险核查"},
	{From: "凭据转储攻击", To: "认证材料暴露风险核查"},
	{From: "golden ticket attack", To: "Kerberos ticket forgery risk review", Word: true, CI: true},
	{From: "pass the ticket", To: "Kerberos ticket reuse risk review", Word: true, CI: true},
	{From: "OS credential dumping", To: "credential exposure risk assessment", Word: true, CI: true},
	{From: "DNS缓存投毒", To: "名称解析缓存完整性风险核查"},
	{From: "ARP缓存投毒", To: "地址映射完整性风险核查"},
	{From: "DNS cache poisoning", To: "name resolution cache integrity risk review", Word: true, CI: true},
	{From: "ARP cache poisoning", To: "address mapping integrity risk review", Word: true, CI: true},
	{From: "会话固定攻击", To: "会话标识轮换风险核查"},
	{From: "点击劫持攻击", To: "界面交互劫持风险核查"},
	{From: "session fixation attack", To: "session identifier rotation risk review", Word: true, CI: true},
	{From: "clickjacking attack", To: "interface interaction hijack risk review", Word: true, CI: true},
	{From: "HTTP请求走私", To: "HTTP请求解析差异风险核查"},
	{From: "HTTP响应拆分", To: "HTTP响应头换行处理风险核查"},
	{From: "HTTP request smuggling", To: "HTTP request parsing inconsistency risk review", Word: true, CI: true},
	{From: "HTTP response splitting", To: "HTTP response header newline handling risk review", Word: true, CI: true},
	{From: "Web缓存投毒", To: "Web缓存响应完整性风险核查"},
	{From: "Web缓存欺骗", To: "Web私有响应错误缓存风险核查"},
	{From: "web cache poisoning", To: "web cache response integrity risk review", Word: true, CI: true},
	{From: "web cache deception", To: "web private response caching risk review", Word: true, CI: true},
	{From: "原型污染攻击", To: "对象原型篡改风险核查"},
	{From: "批量赋值漏洞", To: "非预期字段写入风险核查"},
	{From: "prototype pollution attack", To: "object prototype tampering risk review", Word: true, CI: true},
	{From: "mass assignment vulnerability", To: "unintended field modification risk review", Word: true, CI: true},
	{From: "依赖混淆攻击", To: "依赖来源解析风险核查"},
	{From: "软件包拼写抢注攻击", To: "软件包名称仿冒风险核查"},
	{From: "dependency confusion attack", To: "dependency source resolution risk review", Word: true, CI: true},
	{From: "package typosquatting attack", To: "package name impersonation risk review", Word: true, CI: true},
}

// asciiEdges reports whether term begins and ends with an ASCII letter or
// digit. Only then does a word boundary mean anything: "渗透" is flanked by
// whatever the surrounding prose is, and demanding a non-word byte there
// would stop it matching inside a Chinese sentence.
func asciiEdges(term string) bool {
	if term == "" {
		return false
	}
	first, _ := utf8.DecodeRuneInString(term)
	last, _ := utf8.DecodeLastRuneInString(term)
	return isASCIIWord(first) && isASCIIWord(last)
}

// HasASCIIEdges reports whether a term file would infer word,ci for term when
// its flag column is empty. A writer of term files needs this to know when an
// explicit noword,nocase is required to keep a rule's flags off.
func HasASCIIEdges(term string) bool { return asciiEdges(term) }

func isASCIIWord(r rune) bool {
	return r < utf8.RuneSelf && (unicode.IsLetter(r) || unicode.IsDigit(r))
}

// carryCase maps the casing of the matched text onto the replacement, so an
// all-caps PENTEST becomes THREAT DETECTION and a capitalised Pentest becomes
// Threat detection. Mixed casing is left to the replacement as written.
func carryCase(matched, to string) string {
	switch {
	case matched == strings.ToUpper(matched) && hasLetter(matched):
		return strings.ToUpper(to)
	case isCapitalised(matched):
		r, n := utf8.DecodeRuneInString(to)
		if n == 0 {
			return to
		}
		return string(unicode.ToUpper(r)) + to[n:]
	default:
		return to
	}
}

func hasLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

// isCapitalised reports whether s starts upper-case and the rest is lower.
func isCapitalised(s string) bool {
	first, n := utf8.DecodeRuneInString(s)
	if n == 0 || !unicode.IsUpper(first) {
		return false
	}
	return s[n:] == strings.ToLower(s[n:])
}
