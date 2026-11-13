package glossary

import (
	"strings"
	"testing"
)

func TestPHPObjectRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"PHP对象注入攻击", "PHP反序列化对象注入风险核查", "PHP反序列化对象注入风险核查", 1},
		{"PHP object injection attack", "PHP deserialization object injection risk review", "PHP_deserialization_object_injection_risk_review", 1},
		{"php object injection attack", "PHP deserialization object injection risk review", "PHP_deserialization_object_injection_risk_review", 1},
		{"Php object injection attack", "PHP deserialization object injection risk review", "PHP_deserialization_object_injection_risk_review", 1},
		{"pHp ObJeCt InJeCtIoN AtTaCk", "PHP deserialization object injection risk review", "PHP_deserialization_object_injection_risk_review", 1},
		{"PHP OBJECT INJECTION ATTACK", "PHP DESERIALIZATION OBJECT INJECTION RISK REVIEW", "PHP_DESERIALIZATION_OBJECT_INJECTION_RISK_REVIEW", 1},
		{"PHP对象注入攻击; PHP object injection attack", "PHP反序列化对象注入风险核查; PHP deserialization object injection risk review", "PHP反序列化对象注入风险核查; PHP_deserialization_object_injection_risk_review", 2},
		{"review:\nPHP object injection attack", "review:\nPHP deserialization object injection risk review", "review:\nPHP_deserialization_object_injection_risk_review", 1},
		{"(PHP object injection attack).", "(PHP deserialization object injection risk review).", "(PHP_deserialization_object_injection_risk_review).", 1},
		{"反序列化漏洞", "反序列化风险", "反序列化风险", 1},
		{"原型污染攻击", "对象原型篡改风险核查", "对象原型篡改风险核查", 1},
		{"prototype pollution attack", "object prototype tampering risk review", "object_prototype_tampering_risk_review", 1},
	}
	for _, in := range []string{
		"PHP PHP object injection PHP对象注入 unserialize serialize",
		"__wakeup __destruct __unserialize allowed_classes autoload POP chain",
		"PHP object injection prevention validateSerializedInput JsonSerializable",
		"PHP deserialization object injection risk review PHP反序列化对象注入风险核查",
		"PHP对象注入防护 php对象注入攻击 Php对象注入攻击 PHP 对象注入攻击 PHP对象注入 攻击",
		"object injection attack 对象注入攻击",
		"PHP-object-injection-attack PHP_object_injection_attack",
		"PHP object injection attacks",
		"PHP\tobject injection attack",
		"PHP  object injection attack",
		"PHP\nobject injection attack",
		"PHP object injection\tattack",
		"xPHP object injection attack",
		"PHP object injection attackx",
		"2PHP object injection attack",
		"PHP object injection attack2",
		"_PHP object injection attack",
		"PHP object injection attack_suffix",
		"前PHP object injection attack",
		"PHP object injection attack后",
		`import "PHP object injection attack"`,
		`import "PHP对象注入攻击"`,
		"https://example.invalid/PHP%20object%20injection%20attack",
		"https://example.invalid/PHP对象注入攻击",
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
