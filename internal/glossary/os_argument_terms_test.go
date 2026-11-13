package glossary

import (
	"strings"
	"testing"
)

func TestOSArgumentRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"操作系统命令参数注入攻击", "操作系统命令参数注入风险核查", "操作系统命令参数注入风险核查", 1},
		{"OS command argument injection attack", "OS command argument injection risk review", "OS_command_argument_injection_risk_review", 1},
		{"os command argument injection attack", "OS command argument injection risk review", "OS_command_argument_injection_risk_review", 1},
		{"Os command argument injection attack", "OS command argument injection risk review", "OS_command_argument_injection_risk_review", 1},
		{"oS CoMmAnD ArGuMeNt InJeCtIoN AtTaCk", "OS command argument injection risk review", "OS_command_argument_injection_risk_review", 1},
		{"OS COMMAND ARGUMENT INJECTION ATTACK", "OS COMMAND ARGUMENT INJECTION RISK REVIEW", "OS_COMMAND_ARGUMENT_INJECTION_RISK_REVIEW", 1},
		{"操作系统命令参数注入攻击; OS command argument injection attack", "操作系统命令参数注入风险核查; OS command argument injection risk review", "操作系统命令参数注入风险核查; OS_command_argument_injection_risk_review", 2},
		{"review:\nOS command argument injection attack", "review:\nOS command argument injection risk review", "review:\nOS_command_argument_injection_risk_review", 1},
		{"(OS command argument injection attack).", "(OS command argument injection risk review).", "(OS_command_argument_injection_risk_review).", 1},
		{"命令注入", "命令执行风险核查", "命令执行风险核查", 1},
		{"HTTP参数污染攻击", "HTTP参数污染与解释风险核查", "HTTP参数污染与解释风险核查", 1},
		{"HTTP parameter pollution attack", "HTTP parameter pollution interpretation risk review", "HTTP_parameter_pollution_interpretation_risk_review", 1},
	}
	for _, in := range []string{
		"argument injection 参数注入 OS command argument injection argv execve",
		"exec.Command ProcessStartInfo.ArgumentList CWE-88 shell escaping allowlist",
		"OS command argument injection prevention validateArguments escapeArgument",
		"OS command argument injection risk review 操作系统命令参数注入风险核查",
		"操作系统命令参数注入防护 操作系统 命令参数注入攻击 操作系统命令参数注入 攻击",
		"argument injection attack command argument injection attack OS argument injection attack",
		"命令参数注入攻击 参数注入攻击",
		"OS-command-argument-injection-attack OS_command_argument_injection_attack",
		"OS command argument injection attacks",
		"OS\tcommand argument injection attack",
		"OS  command argument injection attack",
		"OS\ncommand argument injection attack",
		"OS command argument injection\tattack",
		"xOS command argument injection attack",
		"OS command argument injection attackx",
		"2OS command argument injection attack",
		"OS command argument injection attack2",
		"_OS command argument injection attack",
		"OS command argument injection attack_suffix",
		"前OS command argument injection attack",
		"OS command argument injection attack后",
		`import "OS command argument injection attack"`,
		`import "操作系统命令参数注入攻击"`,
		"https://example.invalid/OS%20command%20argument%20injection%20attack",
		"https://example.invalid/操作系统命令参数注入攻击",
	} {
		cases = append(cases, struct {
			in, prose, code string
			hits            int
		}{in, in, in, 0})
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			for _, comment := range []bool{false, true} {
				opt := Options{}
				want := tc.code
				if comment {
					opt.Comments = []Region{{Start: 0, End: len(tc.in), Comment: true}}
					want = tc.prose
				}
				out, hits := tb.Apply([]byte(tc.in), opt)
				if string(out) != want || len(hits) != tc.hits {
					t.Fatalf("comment=%v: got %q (%d hits), want %q (%d hits)", comment, out, len(hits), want, tc.hits)
				}
				for _, h := range hits {
					if h.Start < 0 || h.End > len(tc.in) || h.Start >= h.End {
						t.Fatalf("invalid hit range: %+v", h)
					}
					if h.InComment != comment || h.Line != strings.Count(tc.in[:h.Start], "\n")+1 || tc.in[h.Start:h.End] != h.From || !strings.Contains(want, h.To) {
						t.Errorf("incorrect hit metadata: %+v", h)
					}
				}
				if comment {
					opt.Comments[0].End = len(out)
				}
				again, secondHits := tb.Apply(out, opt)
				if string(again) != string(out) || len(secondHits) != 0 {
					t.Fatalf("second pass changed output: %q (%+v)", again, secondHits)
				}
			}
		})
	}
}
