package glossary

import (
	"strings"
	"testing"
)

func TestZipSlipRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"Zip Slip attack", "archive extraction path traversal risk review", "archive_extraction_path_traversal_risk_review", 1},
		{"zip slip attack", "archive extraction path traversal risk review", "archive_extraction_path_traversal_risk_review", 1},
		{"Zip slip attack", "Archive extraction path traversal risk review", "Archive_extraction_path_traversal_risk_review", 1},
		{"zIp SlIp AtTaCk", "archive extraction path traversal risk review", "archive_extraction_path_traversal_risk_review", 1},
		{"ZIP SLIP ATTACK", "ARCHIVE EXTRACTION PATH TRAVERSAL RISK REVIEW", "ARCHIVE_EXTRACTION_PATH_TRAVERSAL_RISK_REVIEW", 1},
		{"Zip Slip attack; zip slip attack", "archive extraction path traversal risk review; archive extraction path traversal risk review", "archive_extraction_path_traversal_risk_review; archive_extraction_path_traversal_risk_review", 2},
		{"路径穿越", "路径访问核查", "路径访问核查", 1},
		{"目录遍历漏洞", "路径访问核查", "路径访问核查", 1},
	}
	for _, in := range []string{
		"Zip Slip ZIP archive extraction 归档 解压",
		"递归目录遍历",
		"zip.OpenReader zip.NewReader ZipFile.ExtractToDirectory filepath.Join",
		"Zip Slip prevention ZipSlipProtection",
		"zip-slip attack ZipSlip attack zip_slip_attack",
		"Zip\tSlip attack",
		"xZip Slip attack",
		"Zip Slip attacks",
		"2Zip Slip attack",
		"Zip Slip attack2",
		"_Zip Slip attack",
		"Zip Slip attack_suffix",
		`import "Zip Slip attack"`,
		"https://example.invalid/Zip%20Slip%20attack",
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
