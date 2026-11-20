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
var Default = append([]Rule{
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
	{From: "XML外部实体攻击", To: "XML外部实体解析风险核查"},
	{From: "XML递归实体展开攻击", To: "XML递归实体展开资源耗尽风险核查"},
	{From: "XML external entity attack", To: "XML external entity resolution risk review", Word: true, CI: true},
	{From: "XML recursive entity expansion attack", To: "XML recursive entity expansion resource exhaustion risk review", Word: true, CI: true},
	{From: "服务端模板注入攻击", To: "服务端模板表达式注入风险核查"},
	{From: "XPath注入攻击", To: "XPath查询表达式注入风险核查"},
	{From: "server-side template injection attack", To: "server-side template expression injection risk review", Word: true, CI: true},
	{From: "XPath injection attack", To: "XPath query expression injection risk review", Word: true, CI: true},
	{From: "LDAP注入攻击", To: "LDAP查询注入风险核查"},
	{From: "日志伪造攻击", To: "日志记录伪造风险核查"},
	{From: "LDAP injection attack", To: "LDAP query injection risk review", Word: true, CI: true},
	{From: "log forging attack", To: "log record forgery risk review", Word: true, CI: true},
	{From: "正则表达式拒绝服务攻击", To: "正则表达式计算资源耗尽风险核查"},
	{From: "regular expression denial of service attack", To: "regular expression computational resource exhaustion risk review", Word: true, CI: true},
	{From: "HTTP参数污染攻击", To: "HTTP参数污染与解释风险核查"},
	{From: "HTTP parameter pollution attack", To: "HTTP parameter pollution interpretation risk review", Word: true, CI: true},
	{From: "CSV公式注入攻击", To: "CSV单元格公式注入风险核查"},
	{From: "CSV formula injection attack", To: "CSV cell formula injection risk review", Word: true, CI: true},
	{From: "开放重定向攻击", To: "非可信目标重定向风险核查"},
	{From: "open redirect attack", To: "untrusted destination redirect risk review", Word: true, CI: true},
	{From: "DOM clobbering attack", To: "DOM named property clobbering risk review", Word: true, CI: true},
	{From: "HTTP主机头注入攻击", To: "HTTP主机头注入风险核查"},
	{From: "HTTP host header injection attack", To: "HTTP host header injection risk review", Word: true, CI: true},
	{From: "NoSQL注入攻击", To: "NoSQL查询逻辑注入风险核查"},
	{From: "NoSQL injection attack", To: "NoSQL query logic injection risk review", Word: true, CI: true},
	{From: "跨站WebSocket劫持攻击", To: "跨站WebSocket会话劫持风险核查"},
	{From: "cross-site WebSocket hijacking attack", To: "cross-site WebSocket session hijacking risk review", Word: true, CI: true},
	{From: "JWT算法混淆攻击", To: "JWT验签算法混淆风险核查"},
	{From: "JWT algorithm confusion attack", To: "JWT verification algorithm confusion risk review", Word: true, CI: true},
	{From: "XML signature wrapping attack", To: "XML signature wrapping risk review", Word: true, CI: true},
	{From: "Zip Slip attack", To: "archive extraction path traversal risk review", Word: true, CI: true},
	{From: "OAuth2.0授权码注入攻击", To: "OAuth2.0授权码注入风险核查"},
	{From: "OAuth 2.0 authorization code injection attack", To: "OAuth 2.0 authorization code injection risk review", Word: true, CI: true},
	{From: "OAuth mix-up attack", To: "OAuth authorization server mix-up risk review", Word: true, CI: true},
	{From: "外部可控格式字符串漏洞", To: "外部可控格式字符串风险核查"},
	{From: "externally controlled format string vulnerability", To: "externally controlled format string risk review", Word: true, CI: true},
	{From: "DNS重绑定攻击", To: "DNS重绑定信任边界风险核查"},
	{From: "DNS rebinding attack", To: "DNS rebinding trust boundary risk review", Word: true, CI: true},
	{From: "HTTP verb tampering attack", To: "HTTP verb tampering access control risk review", Word: true, CI: true},
	{From: "TLS协议降级攻击", To: "TLS协议降级风险核查"},
	{From: "TLS protocol downgrade attack", To: "TLS protocol downgrade risk review", Word: true, CI: true},
	{From: "HTML注入攻击", To: "HTML标记注入风险核查"},
	{From: "HTML injection attack", To: "HTML markup injection risk review", Word: true, CI: true},
	{From: "内存释放后使用漏洞", To: "内存释放后使用风险核查"},
	{From: "内存重复释放漏洞", To: "内存重复释放风险核查"},
	{From: "use-after-free vulnerability", To: "use-after-free risk review", Word: true, CI: true},
	{From: "double free vulnerability", To: "double free risk review", Word: true, CI: true},
	{From: "内存越界读取漏洞", To: "内存越界读取风险核查"},
	{From: "内存越界写入漏洞", To: "内存越界写入风险核查"},
	{From: "out-of-bounds read vulnerability", To: "out-of-bounds read risk review", Word: true, CI: true},
	{From: "out-of-bounds write vulnerability", To: "out-of-bounds write risk review", Word: true, CI: true},
	{From: "TOCTOU race condition vulnerability", To: "TOCTOU check-use race risk review", Word: true, CI: true},
	{From: "整数溢出漏洞", To: "整数溢出风险核查"},
	{From: "整数下溢漏洞", To: "整数下溢风险核查"},
	{From: "integer overflow vulnerability", To: "integer overflow risk review", Word: true, CI: true},
	{From: "integer underflow vulnerability", To: "integer underflow risk review", Word: true, CI: true},
	{From: "类型混淆漏洞", To: "类型混淆风险核查"},
	{From: "type confusion vulnerability", To: "type confusion risk review", Word: true, CI: true},
	{From: "空指针解引用漏洞", To: "空指针解引用风险核查"},
	{From: "null pointer dereference vulnerability", To: "null pointer dereference risk review", Word: true, CI: true},
	{From: "不受控递归漏洞", To: "不受控递归资源耗尽风险核查"},
	{From: "uncontrolled recursion vulnerability", To: "uncontrolled recursion resource exhaustion risk review", Word: true, CI: true},
	{From: "未初始化变量使用漏洞", To: "未初始化变量使用风险核查"},
	{From: "uninitialized variable use vulnerability", To: "uninitialized variable use risk review", Word: true, CI: true},
	{From: "OAuth刷新令牌重放攻击", To: "OAuth刷新令牌重放风险核查"},
	{From: "OAuth refresh token replay attack", To: "OAuth refresh token replay risk review", Word: true, CI: true},
	{From: "padding oracle attack", To: "padding oracle risk review", Word: true, CI: true},
	{From: "reverse tabnabbing attack", To: "reverse tabnabbing risk review", Word: true, CI: true},
	{From: "跨站脚本包含攻击", To: "跨站脚本包含数据泄露风险核查"},
	{From: "cross site script inclusion attack", To: "cross site script inclusion data exposure risk review", Word: true, CI: true},
	{From: "CSS注入攻击", To: "CSS样式注入风险核查"},
	{From: "CSS injection attack", To: "CSS style injection risk review", Word: true, CI: true},
	{From: "SSI注入攻击", To: "SSI服务端包含指令注入风险核查"},
	{From: "SSI injection attack", To: "SSI server-side include directive injection risk review", Word: true, CI: true},
	{From: "IDN homograph attack", To: "IDN homograph domain spoofing risk review", Word: true, CI: true},
	{From: "操作系统命令参数注入攻击", To: "操作系统命令参数注入风险核查"},
	{From: "OS command argument injection attack", To: "OS command argument injection risk review", Word: true, CI: true},
	{From: "session puzzling attack", To: "session variable overloading risk review", Word: true, CI: true},
	{From: "表达式语言注入攻击", To: "表达式语言求值注入风险核查"},
	{From: "expression language injection attack", To: "expression language evaluation injection risk review", Word: true, CI: true},
	{From: "reflected file download attack", To: "reflected file download risk review", Word: true, CI: true},
	{From: "CORS错误配置漏洞", To: "CORS错误配置风险核查"},
	{From: "CORS misconfiguration vulnerability", To: "CORS misconfiguration risk review", Word: true, CI: true},
	{From: "XQuery注入攻击", To: "XQuery查询结构注入风险核查"},
	{From: "XQuery injection attack", To: "XQuery query structure injection risk review", Word: true, CI: true},
	{From: "PHP对象注入攻击", To: "PHP反序列化对象注入风险核查"},
	{From: "PHP object injection attack", To: "PHP deserialization object injection risk review", Word: true, CI: true},
	{From: "TLS证书主机名验证漏洞", To: "TLS证书主机名验证风险核查"},
	{From: "TLS certificate hostname validation vulnerability", To: "TLS certificate hostname validation risk review", Word: true, CI: true},
	{From: "外部可控反射漏洞", To: "外部可控反射代码选择风险核查"},
	{From: "externally controlled reflection vulnerability", To: "externally controlled reflection code selection risk review", Word: true, CI: true},
	{From: "不可信搜索路径漏洞", To: "不可信搜索路径资源选择风险核查"},
	{From: "untrusted search path vulnerability", To: "untrusted search path resource selection risk review", Word: true, CI: true},
	{From: "cryptographically weak PRNG vulnerability", To: "cryptographically weak PRNG risk review", Word: true, CI: true},
	{From: "postMessage来源验证漏洞", To: "postMessage来源验证风险核查"},
	{From: "postMessage origin validation vulnerability", To: "postMessage origin validation risk review", Word: true, CI: true},
	{From: "不受限文件上传漏洞", To: "不受限文件上传风险核查"},
	{From: "unrestricted file upload vulnerability", To: "unrestricted file upload risk review", Word: true, CI: true},
	{From: "signed to unsigned conversion vulnerability", To: "signed to unsigned conversion risk review", Word: true, CI: true},
	{From: "double-checked locking vulnerability", To: "double-checked locking initialization visibility risk review", Word: true, CI: true},
	{From: "未检查返回值漏洞", To: "未检查返回值风险核查"},
	{From: "unchecked return value vulnerability", To: "unchecked return value risk review", Word: true, CI: true},
	{From: "主机型入侵检测系统", To: "基于主机的入侵检测系统"},
	{From: "网络型入侵检测系统", To: "基于网络的入侵检测系统"},
	// Keep IDS case-sensitive: IDs can mean identifiers instead.
	{From: "host-based IDS", To: "host-based intrusion detection system", Word: true},
	{From: "network-based IDS", To: "network intrusion detection system", Word: true},
}, auditRules...)

// NewAuditTable prepares only the meaning-preserving audit rules for the browser preview.
func NewAuditTable() *Table { return NewTable(auditRules) }

var auditRules = []Rule{
	{From: "audit record retention period", To: "retention period for audit records", Word: true, CI: true},
	{From: "timestamps for audit records", To: "time stamps for audit records", Word: true, CI: true},
	{From: "audit event result", To: "outcome of the audited event", Word: true, CI: true},
	{From: "allocate audit log storage capacity", To: "allocate storage capacity for audit logs", Word: true, CI: true},
	{From: "the disk with the audit log is almost full", To: "the disk holding the audit log is nearly full", Word: true, CI: true},
	{From: "rotation of audit log files", To: "audit log file rotation", Word: true, CI: true},
	{From: "compression of rotated audit log files", To: "compression of audit log files that have been rotated", Word: true, CI: true},
	{From: "verify audit log file integrity", To: "verify the integrity of the audit log file", Word: true, CI: true},
	{From: "provides read-only access to audit records", To: "allows read-only access to audit records", Word: true, CI: true},
	{From: "original time ordering of audit records", To: "original chronological order of audit records", Word: true, CI: true},
	{From: "the original content of audit records is preserved", To: "the original content of audit records is kept intact", Word: true, CI: true},
	{From: "generate on-demand audit reports", To: "generate audit reports on demand", Word: true, CI: true},
	{From: "search and filter audit records by event type", To: "search and filter audit log entries by event type", Word: true, CI: true},
	{From: "find all audit records within a time range", To: "find all audit records in a specified time range", Word: true, CI: true},
	{From: "search audit records by actor", To: "query audit records by actor", Word: true, CI: true},
	{From: "get audit logs containing a given target resource", To: "retrieve audit logs containing a given target resource", Word: true, CI: true},
	{From: "Get audit logs by correlation ID", To: "Retrieve audit logs by correlation ID", Word: true},
	{From: "get the next page of audit logs", To: "retrieve the next page of audit logs", Word: true, CI: true},
	{From: "maximum number of audit logs to return per page", To: "maximum number of audit logs to return on any one page", Word: true, CI: true},
	{From: "export the results of an audit log search to a CSV file", To: "export audit log search results to a CSV file", Word: true},
	{From: "create a saved audit log filter", To: "save a new audit log filter", Word: true, CI: true},
	{From: "clear all audit log filters and reload", To: "clear every audit log filter and reload", Word: true, CI: true},
	{From: "when you expand an audit log entry, you see exactly what changed", To: "expanding an audit log entry shows you exactly what changed", Word: true, CI: true},
	{From: "the audit log shows the old and new values", To: "the audit log shows the before and after values", Word: true, CI: true},
	{From: "Copy Audit Event ID", To: "Copy Audit Event Identifier", Word: true},
	{From: "timestamps in the audit logs table and detail view are displayed in your local time zone", To: "timestamps in the audit logs table and detail view are shown in your local time zone", Word: true, CI: true},
	{From: "The time the event described by the log entry occurred", To: "When the event described by the log entry occurred", Word: true},
	{From: "The time the log entry was received by Logging", To: "When the log entry was received by Logging", Word: true},
	{From: "no audit logs match your filters", To: "no audit log entries match your filters", Word: true, CI: true},
	{From: "Loading audit log entries", To: "Loading audit log records", Word: true},
	{From: "Retry loading audit logs", To: "Try loading audit logs again", Word: true},
	{From: "audit log query timed out", To: "audit log query did not return in time", Word: true, CI: true},
	{From: "Cancel a running CloudTrail Lake query", To: "Request cancellation of a running CloudTrail Lake query", Word: true},
	{From: "Get the status of a CloudTrail Lake query", To: "Check the status of a CloudTrail Lake query", Word: true},
	{From: "If the query results were delivered to an S3 bucket, the response also provides the S3 URI and the delivery status", To: "If the query results were delivered to an S3 bucket, the response also includes the S3 URI and the delivery status", Word: true},
	{From: "The number of events that the query scanned in the event data store", To: "How many events the query scanned in the event data store", Word: true},
	{From: "The total bytes that the query scanned in the event data store", To: "The total number of bytes that the query scanned in the event data store", Word: true},
	{From: "The query's run time, in milliseconds", To: "The duration of query execution in milliseconds", Word: true},
	{From: "Lists matching query results, and shows query ID, status, and creation time of each query", To: "Lists matching query results and shows the ID, status, and creation time for each query", Word: true},
	{From: "The status of queries that you want to return in results", To: "The status to match when listing queries", Word: true},
	{From: "to bound a ListQueries request, and limit its results to only those queries run within a specified time period", To: "to limit ListQueries results to queries run within the specified time range", Word: true},
	{From: "The ARN (or the ID suffix of the ARN) of an event data store on which queries were run", To: "The ARN or its ID suffix for the event data store where the queries ran", Word: true},
	{From: "NextPage retrieves the next ListQueries page", To: "NextPage retrieves the next page from ListQueries", Word: true},
	{From: "The maximum number of queries to show on a page", To: "The maximum number of queries to list per page", Word: true},
	{From: "The error message returned if a query failed", To: "The error message returned if the query itself failed", Word: true},
	{From: "invalid audit log page token", To: "invalid audit log pagination token", Word: true, CI: true},
	{From: "Information about whether the event is a write event or a read event", To: "Whether the event is classified as a write or a read", Word: true},
	{From: "failed to lookup CloudTrail events by event name", To: "failed to look up CloudTrail events by event name", Word: true},
	{From: "Retrieves a specific audit event by its ID", To: "Retrieves a specific audit event by its identifier", Word: true},
	{From: "Parameter for filtering audit logs by resource type", To: "Parameter to filter audit logs by resource type", Word: true},
	{From: "Get audit logs by source IP", To: "Retrieve audit logs by source IP", Word: true},
	{From: "Filter OpenShift API server audit logs by user agent", To: "Filter OpenShift API server audit log entries by user agent", Word: true},
	{From: "Find audit logs by HTTP status code", To: "Find audit log entries by HTTP status code", Word: true},
	{From: "Find audit logs by HTTP method", To: "Find audit log entries by HTTP method", Word: true},
	{From: "Find audit logs by request path", To: "Find audit log entries by request path", Word: true},
	{From: "auth_decision: Filter by authorization decision", To: "auth_decision: Authorization decision to match", Word: true},
	{From: "Filter by minimum request duration in milliseconds", To: "Filter by minimum request duration in ms", Word: true},
	{From: "Filter by maximum request duration in milliseconds", To: "Filter by maximum request duration in ms", Word: true},
	{From: "This filters all TLS connections with 1.1 and below", To: "This filters all TLS connections using version 1.1 or lower", Word: true},
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
