package glossary

import (
	"strings"
	"testing"
)

func TestCSVFormulaRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"CSV公式注入攻击", "CSV单元格公式注入风险核查", "CSV单元格公式注入风险核查", 1},
		{"CSV formula injection attack", "CSV cell formula injection risk review", "CSV_cell_formula_injection_risk_review", 1},
		{"csv formula injection attack", "CSV cell formula injection risk review", "CSV_cell_formula_injection_risk_review", 1},
		{"Csv formula injection attack", "CSV cell formula injection risk review", "CSV_cell_formula_injection_risk_review", 1},
		{"cSv FoRmUlA iNjEcTiOn AtTaCk", "CSV cell formula injection risk review", "CSV_cell_formula_injection_risk_review", 1},
		{"CSV FORMULA INJECTION ATTACK", "CSV CELL FORMULA INJECTION RISK REVIEW", "CSV_CELL_FORMULA_INJECTION_RISK_REVIEW", 1},
		{"CSV公式注入攻击; CSV formula injection attack", "CSV单元格公式注入风险核查; CSV cell formula injection risk review", "CSV单元格公式注入风险核查; CSV_cell_formula_injection_risk_review", 2},
		{"服务端模板注入攻击", "服务端模板表达式注入风险核查", "服务端模板表达式注入风险核查", 1},
		{"server-side template injection attack", "server-side template expression injection risk review", "server_side_template_expression_injection_risk_review", 1},
	}
	for _, in := range []string{
		"CSV injection formula injection CSV formula injection CSV公式注入",
		"CSV formula spreadsheet 公式 单元格 CWE-1236",
		"csv.NewWriter csv.Writer csv.writer write_csv setCellFormula",
		"CSV公式注入防护 csv公式注入攻击 CSV 公式注入攻击",
		"CSV injection attack formula injection attack",
		"CSV-formula-injection-attack CSV_formula_injection_attack",
		"CSV\tformula injection attack",
		"xCSV formula injection attack",
		"CSV formula injection attacks",
		"2CSV formula injection attack",
		"CSV formula injection attack2",
		"_CSV formula injection attack",
		"CSV formula injection attack_suffix",
		`import "CSV formula injection attack"`,
		`import "CSV公式注入攻击"`,
		"https://example.invalid/CSV公式注入攻击",
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
					if h.InComment != comment || h.Line != 1 || tc.in[h.Start:h.End] != h.From || !strings.Contains(want, h.To) {
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
