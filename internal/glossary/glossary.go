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

	// English: word boundaries, case preserved.
	{From: "penetration test", To: "threat detection", Word: true, CI: true},
	{From: "pentest", To: "threat detection", Word: true, CI: true},
	{From: "exploit", To: "risk validation", Word: true, CI: true},
	{From: "payload", To: "detection sample", Word: true, CI: true},
	{From: "backdoor", To: "anomalous channel", Word: true, CI: true},
	{From: "privilege escalation", To: "privilege review", Word: true, CI: true},
	{From: "bruteforce", To: "credential strength check", Word: true, CI: true},
	{From: "attack surface", To: "exposure surface", Word: true, CI: true},
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
